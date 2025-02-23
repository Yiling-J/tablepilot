package ai

import (
	"context"
	"errors"
	"time"

	"tablepilot/config"
	"tablepilot/services/ai/client"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

//go:generate moq -out ai_moq.go . AiService
type AiService interface {
	Chat(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error)
}

type model struct {
	model   string
	alias   string
	client  string
	limiter *rate.Limiter
}

type AiServiceImpl struct {
	clients      map[string]client.ChatClient
	models       map[string]*model
	logger       *zap.SugaredLogger
	defaultModel string
	config       *config.Config
}

func NewAiService(cfg *config.Config, clients map[string]client.ChatClient, logger *zap.SugaredLogger) (*AiServiceImpl, error) {
	srv := &AiServiceImpl{
		clients: clients,
		models:  map[string]*model{},
		logger:  logger,
		config:  cfg,
	}

	for i, m := range cfg.Models {
		if i == 0 || m.Default {
			srv.defaultModel = m.Model
		}
		_, ok := srv.clients[m.Client]
		if !ok {
			return nil, errors.New("not found")
		}

		var limiter *rate.Limiter
		if m.RPM > 0 {
			limiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(m.RPM)), m.RPM)
		}
		md := &model{
			model:   m.Model,
			alias:   m.Alias,
			limiter: limiter,
			client:  m.Client,
		}
		srv.models[md.model] = md
		if md.alias != "" {
			srv.models[md.alias] = md
		}
	}
	return srv, nil
}

func (ai *AiServiceImpl) Chat(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
	if request.Model == "" {
		request.Model = ai.defaultModel
	}
	client, err := ai.getChatClientByModel(request.Model)
	if err != nil {
		return nil, err
	}

	request.Model = ai.models[request.Model].model
	ai.logger.Debugln("send chat request", "model", request.Model, "temperature", request.Temperature)
	for _, message := range request.Messages {
		ai.logger.Debugf("[%s]message: \n%s", message.Role, message.Content)
	}
	resp, err := client.Chat(ctx, request)
	if err != nil {
		return nil, err
	}
	ai.logger.Debugln("chat response reveived", "total_tokens", resp.Tokens)
	ai.logger.Debugln("content:", resp.Content)
	return resp, nil
}

func (ai *AiServiceImpl) getChatClientByModel(model string) (client.ChatClient, error) {
	for _, m := range ai.models {
		if m.model == model || m.alias == model {
			return ai.clients[m.client], nil
		}
	}
	return nil, errors.New("")
}
