package client

import (
	"context"
	"fmt"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/services/provider"

	"go.uber.org/zap"
)

//go:generate moq -rm -out client_moq.go . ChatClient
type ChatClient interface {
	Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)
	ImageGen(ctx context.Context, request *ChatRequest) (*ImageGenResponse, error)
	FunctionCall(ctx context.Context, request *ChatRequest) (*FunctionCallResponse, error)
}

func NewClient(p provider.Provider, logger *zap.SugaredLogger) (ChatClient, error) {
	switch p.Type {
	case "openai":
		logger.Debugw("creating new openai client", "name", p.Name)
		completion := NewOpenAIChatCompletionService(&config.OpenAI{
			Name:    p.Name,
			Type:    p.Type,
			Key:     p.Key,
			BaseURL: p.BaseURL,
		})
		oai := NewOpenAIClient(completion, logger)
		logger.Debug("openai client created")
		return oai, nil
	case "gemini":
		logger.Debugw("creating new gemini client", "name", p.Name)
		genai, err := NewGeminiClient(&config.Gemini{
			Name: p.Name,
			Type: p.Type,
			Key:  p.Key,
		}, logger)
		if err != nil {
			return nil, fmt.Errorf("client.NewClient: creating Gemini client: %w", err)
		}
		logger.Debug("gemini client created")
		return genai, nil
	default:
		return nil, fmt.Errorf("client.NewClient: unknown config type: %s", p.Type)
	}
}
