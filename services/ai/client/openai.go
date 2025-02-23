package client

import (
	"context"
	"errors"
	"tablepilot/config"

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
	return openai.NewClient(
		option.WithAPIKey(config.Key),
		option.WithBaseURL(config.BaseURL),
	).Chat.Completions
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
			messages = append(messages, openai.UserMessage(m.Content))
		}
	}
	chatParams := openai.ChatCompletionNewParams{
		Messages:        openai.F(messages),
		Model:           openai.F(request.Model),
		Temperature:     openai.Float(request.Temperature),
		PresencePenalty: openai.Float(request.PresencePenalty),
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
			Name:        openai.F("schema"),
			Description: openai.F("schema for table"),
			Schema:      openai.F(schema),
		}
		chatParams.ResponseFormat = openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		)
	}
	chatCompletion, err := c.completionService.New(ctx, chatParams)
	if err != nil {
		return nil, err
	}
	if len(chatCompletion.Choices) < 1 {
		return nil, errors.New("")
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
