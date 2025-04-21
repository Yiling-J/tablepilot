package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Yiling-J/tablepilot/config"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/tidwall/gjson"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go.uber.org/zap"
)

type oaiChatCompletionService interface {
	New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (res *openai.ChatCompletion, err error)
}

func NewOpenAIChatCompletionService(config *config.OpenAI) *openai.ChatCompletionService {
	options := []option.RequestOption{
		option.WithAPIKey(config.Key),
		option.WithBaseURL(config.BaseURL),
	}
	// recording request/response to json snapshot file
	if name, _ := os.LookupEnv("TABLEPILOT_SNAPSHOT_RECORD"); len(name) > 0 {
		options = append(options, option.WithMiddleware(func(r *http.Request, mn option.MiddlewareNext) (*http.Response, error) {
			var reqBody []byte
			if r.Body != nil {
				reqBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(reqBody))
			}

			resp, err := mn(r)
			if err != nil {
				return resp, err
			}

			var respBody []byte
			if resp.Body != nil {
				respBody, _ = io.ReadAll(resp.Body)
				resp.Body = io.NopCloser(bytes.NewReader(respBody))
			}

			snapshot := map[string]string{
				"request":  string(reqBody),
				"response": string(respBody),
			}

			var snapshots []map[string]string
			filename := fmt.Sprintf("tests/snapshots/%s.json", name)
			fileData, err := os.ReadFile(filename)
			if err == nil {
				_ = json.Unmarshal(fileData, &snapshots)
			}
			snapshots = append(snapshots, snapshot)
			file, err := os.Create(filename)
			if err == nil {
				defer file.Close()
				_ = json.NewEncoder(file).Encode(snapshots)
			}

			return resp, nil
		}))
	}
	srv := openai.NewClient(
		options...,
	).Chat.Completions
	return &srv
}

type OpenAIClient struct {
	completionService oaiChatCompletionService
	logger            *zap.SugaredLogger
}

func NewOpenAIClient(completionService oaiChatCompletionService, logger *zap.SugaredLogger) *OpenAIClient {
	return &OpenAIClient{logger: logger, completionService: completionService}
}

func (c *OpenAIClient) Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	messages := []openai.ChatCompletionMessageParamUnion{}
	for _, m := range request.Messages {
		switch m.Role {
		case "user":
			for _, c := range m.Content {
				switch c.Type {
				case ContentTypeText:
					messages = append(messages, openai.UserMessage(c.Data))
				case ContentTypeImage:
					messages = append(messages, openai.UserMessage(
						[]openai.ChatCompletionContentPartUnionParam{
							openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
								URL: c.Data,
							}),
						},
					))
				}
			}
		}
	}
	chatParams := openai.ChatCompletionNewParams{
		Messages:            messages,
		Model:               request.Model,
		Temperature:         openai.Float(request.Temperature),
		PresencePenalty:     openai.Float(request.PresencePenalty),
		MaxCompletionTokens: openai.Int(request.MaxOutputTokens),
	}
	array := false
	if request.Schema != nil {
		// root object can't be array
		// https://community.openai.com/t/support-top-level-array-in-json-schema/896048
		if request.Schema.Type == "array" {
			rom := orderedmap.New[string, *jsonschema.Schema]()
			rom.Set("data", request.Schema)
			request.Schema = &jsonschema.Schema{
				Type:       "object",
				Properties: rom,
			}
			array = true
		}
		var schema any = request.Schema
		schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:        "schema",
			Description: openai.String("schema for table"),
			Schema:      schema,
		}
		chatParams.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		}
	}
	chatCompletion, err := c.completionService.New(ctx, chatParams)
	if err != nil {
		return nil, err
	}
	if len(chatCompletion.Choices) < 1 {
		return nil, errors.New("chat choices empty")
	}
	content := chatCompletion.Choices[0].Message.Content
	if array {
		content = gjson.Get(content, "data").String()
	}
	return &ChatResponse{
		Content: content,
		Tokens:  chatCompletion.Usage.TotalTokens,
	}, nil
}

func (c *OpenAIClient) FunctionCall(ctx context.Context, request *ChatRequest) (*FunctionCallResponse, error) {
	messages := []openai.ChatCompletionMessageParamUnion{}
	for _, m := range request.Messages {
		switch m.Role {
		case "user":
			for _, c := range m.Content {
				switch c.Type {
				case ContentTypeText:
					messages = append(messages, openai.UserMessage(c.Data))
				case ContentTypeImage:
					messages = append(messages, openai.UserMessage(
						[]openai.ChatCompletionContentPartUnionParam{
							openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
								URL: c.Data,
							}),
						},
					))
				}
			}
		}
	}
	tools := []openai.ChatCompletionToolParam{}
	for _, tool := range request.Tools {
		tools = append(tools, tool.OpenAITool())
	}
	chatParams := openai.ChatCompletionNewParams{
		Messages:            messages,
		Model:               request.Model,
		Temperature:         openai.Float(request.Temperature),
		PresencePenalty:     openai.Float(request.PresencePenalty),
		MaxCompletionTokens: openai.Int(request.MaxOutputTokens),
		Tools:               tools,
	}
	chatCompletion, err := c.completionService.New(ctx, chatParams)
	if err != nil {
		return nil, err
	}
	if len(chatCompletion.Choices) < 1 {
		return nil, errors.New("chat choices empty")
	}

	calls := []FunctionCall{}
	for _, tc := range chatCompletion.Choices[0].Message.ToolCalls {
		var m map[string]any
		err = json.Unmarshal([]byte(tc.Function.Arguments), &m)
		if err != nil {
			return nil, err
		}
		calls = append(calls, FunctionCall{
			Name:      tc.Function.Name,
			Arguments: m,
		})
	}
	return &FunctionCallResponse{
		Text:          chatCompletion.Choices[0].Message.Content,
		FunctionCalls: calls,
		Tokens:        chatCompletion.Usage.TotalTokens,
	}, nil
}

func (c *OpenAIClient) ImageGen(ctx context.Context, request *ChatRequest) (*ImageGenResponse, error) {
	return nil, errors.New("Not implemented")
}
