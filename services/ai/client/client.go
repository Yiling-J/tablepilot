package client

import (
	"context"
	"errors"

	"github.com/Yiling-J/tablepilot/config"

	"go.uber.org/zap"
)

//go:generate moq -rm -out client_moq.go . ChatClient
type ChatClient interface {
	Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)
	ImageGen(ctx context.Context, request *ChatRequest) (*ImageGenResponse, error)
}

func NewClients(cfg *config.Config, logger *zap.SugaredLogger) (map[string]ChatClient, error) {
	clients := map[string]ChatClient{}
	for _, c := range cfg.Clients {
		switch v := c.(type) {
		case *config.OpenAI:
			logger.Debugw("creating new openai client", "name", v.Name)
			completion := NewOpenAIChatCompletionService(v)
			oai := NewOpenAIClient(completion, logger)
			logger.Debug("openai client created")
			clients[v.Name] = oai
		case *config.Gemini:
			logger.Debugw("creating new gemini client", "name", v.Name)
			genai, err := NewGeminiClient(v)
			if err != nil {
				return nil, err
			}
			logger.Debug("gemini client created")
			clients[v.Name] = genai
		default:
			return nil, errors.New("unknown config type")
		}
	}
	return clients, nil
}
