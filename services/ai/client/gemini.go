package client

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"go.uber.org/zap"
	"google.golang.org/genai"
)

//go:generate moq -rm -out gemini_moq.go . GenaiModelService
type GenaiModelService interface {
	GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

type GeminiClient struct {
	modelService GenaiModelService
	oaiClient    *OpenAIClient
	logger       *zap.SugaredLogger
}

func NewGeminiClient(cfg *config.Gemini, logger *zap.SugaredLogger) (*GeminiClient, error) {
	c, err := genai.NewClient(context.TODO(), &genai.ClientConfig{
		APIKey:  cfg.Key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	completion := NewOpenAIChatCompletionService(&config.OpenAI{
		Name:    cfg.Name,
		Type:    cfg.Type,
		Key:     cfg.Key,
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai/",
	})
	oai := NewOpenAIClient(completion, logger)
	return &GeminiClient{
		modelService: c.Models,
		oaiClient:    oai,
		logger:       logger,
	}, nil
}

func (c *GeminiClient) Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	return c.oaiClient.Chat(ctx, request)
}

func (c *GeminiClient) FunctionCall(ctx context.Context, request *ChatRequest) (*FunctionCallResponse, error) {
	return c.oaiClient.FunctionCall(ctx, request)
}

var imageIdRE = regexp.MustCompile(`<info\s+row_id="([0-9a-zA-Z]+)"\s+column_id="([0-9a-zA-Z]+)"\s*\/>`)

func (c *GeminiClient) ImageGen(ctx context.Context, request *ChatRequest) (*ImageGenResponse, error) {
	tp := float32(request.Temperature)
	maxTokens := int32(8192)

	messages := []*genai.Part{}
	for _, m := range request.Messages {
		switch m.Role {
		case "user":
			for _, c := range m.Content {
				switch c.Type {
				case ContentTypeText:
					messages = append(messages, genai.NewPartFromText(c.Data))
				case ContentTypeImage:
					// c.Data should be data url format: data:{content type};base64,{b64 image}
					tmp := strings.Split(c.Data, ";base64,")
					if len(tmp) == 2 {
						b, err := base64.StdEncoding.DecodeString(tmp[1])
						if err != nil {
							return nil, err
						}
						messages = append(messages, genai.NewPartFromBytes(b, strings.TrimPrefix(tmp[0], "data:")))
					}
				}
			}
		}
	}

	var resp *genai.GenerateContentResponse
	var err error
	retry := 0
	for retry < 3 {
		resp, err = c.modelService.GenerateContent(ctx, request.ImageModel, []*genai.Content{{
			Parts: messages,
			Role:  "user",
		}}, &genai.GenerateContentConfig{
			Temperature:        &tp,
			MaxOutputTokens:    &maxTokens,
			ResponseModalities: []string{"IMAGE", "TEXT"},
		})
		if err != nil {
			c.logger.Errorw("gemini image generation failed", "error", err, "retry", retry)
			// TODO: reduce this when gemini api is stable
			time.Sleep(5 * time.Second)
			retry += 1
			continue
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	images := map[string][]byte{}
	// image id text can appear before or after the generated image
	var imageID string
	var fileData []byte
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, errors.New("gemini return zero candidates")
	}
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			c.logger.Debugw("Gemini text response", "text", part.Text)
			match := imageIdRE.FindStringSubmatch(part.Text)
			if len(match) == 3 {
				rowID := match[1]
				columnID := match[2]
				imageID = fmt.Sprintf("%s-%s", rowID, columnID)
				// sometimes gemini return id text twice
				if _, ok := images[imageID]; ok {
					imageID = ""
					continue
				}
				// image id after image generated
				if fileData != nil {
					images[imageID] = fileData
					fileData = nil
					imageID = ""
				}
			}
		}
		if part.InlineData != nil {
			c.logger.Debug("Gemini image response")
			// image id before image generated
			if imageID != "" {
				images[imageID] = part.InlineData.Data
				imageID = ""
			} else {
				fileData = part.InlineData.Data
			}
		}
	}

	var tokens int32 = 0
	if resp.UsageMetadata != nil {
		tokens = resp.UsageMetadata.TotalTokenCount
	}

	return &ImageGenResponse{
		Images: images,
		Tokens: int64(tokens),
	}, nil
}
