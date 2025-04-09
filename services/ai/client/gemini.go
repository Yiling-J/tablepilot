package client

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/Yiling-J/tablepilot/config"
	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(config *config.Gemini) (*GeminiClient, error) {
	c, err := genai.NewClient(context.TODO(), &genai.ClientConfig{
		APIKey:  config.Key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	return &GeminiClient{
		client: c,
	}, nil
}

func (c *GeminiClient) Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error) {
	return nil, errors.New("Not implemented")
}

var imageIdRE = regexp.MustCompile(`<gen\s+row_id="([0-9a-zA-Z]+)"\s+column_id="([0-9a-zA-Z]+)"\s*\/>`)

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
					messages = append(messages, genai.NewPartFromBytes([]byte(c.Data), string(c.Type)))
				}
			}
		}
	}

	resp, err := c.client.Models.GenerateContent(ctx, request.ImageModel, []*genai.Content{{
		Parts: messages,
		Role:  "user",
	}}, &genai.GenerateContentConfig{
		Temperature:        &tp,
		MaxOutputTokens:    &maxTokens,
		ResponseModalities: []string{"IMAGE", "TEXT"},
	})
	if err != nil {
		return nil, err
	}

	images := map[string][]byte{}
	// image id text can appear before or after the generated image
	var imageID string
	var fileData []byte
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			match := imageIdRE.FindStringSubmatch(part.Text)
			if len(match) == 3 {
				rowID := match[1]
				columnID := match[2]
				imageID = fmt.Sprintf("%s-%s", rowID, columnID)
				// sometimes gemini return id text twice
				if _, ok := images[imageID]; ok {
					continue
				}
				// image id after image generated
				if fileData != nil {
					images[imageID] = fileData
					fileData = nil
				}
			}
		}
		if part.InlineData != nil {
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
