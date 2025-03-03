package client

import (
	"context"
	"errors"

	"github.com/Yiling-J/tablepilot/config"

	"go.uber.org/zap"
)

type ChatClient interface {
	Chat(ctx context.Context, request *ChatRequest) (*ChatResponse, error)
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
		default:
			return nil, errors.New("")
		}
	}
	return clients, nil
}
