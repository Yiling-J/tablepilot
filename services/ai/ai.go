package ai

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/services/ai/client"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

//go:generate moq -rm -out ai_moq.go . AiService
type AiService interface {
	Chat(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error)
	ImageGen(ctx context.Context, request *client.ChatRequest) (*client.ImageGenResponse, error)
	ListModels(ctx context.Context) *ModelList
	FunctionCall(ctx context.Context, request *client.ChatRequest) (*client.FunctionCallResponse, error)
}

type model struct {
	model     string
	alias     string
	client    string
	maxTokens int64
	limiter   *rate.Limiter
}

type AiServiceImpl struct {
	clients           map[string]client.ChatClient
	models            map[string]*model
	logger            *zap.SugaredLogger
	defaultModel      string
	defaultImageModel string
	config            *config.Config
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
			if m.Alias != "" {
				srv.defaultModel = m.Alias
			} else {
				srv.defaultModel = m.Model
			}
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
			model:     m.Model,
			maxTokens: m.MaxTokens,
			alias:     m.Alias,
			limiter:   limiter,
			client:    m.Client,
		}
		srv.models[md.model] = md
		if md.alias != "" {
			srv.models[md.alias] = md
		}
	}

	for i, m := range cfg.ImageModels {
		if i == 0 || m.Default {
			if m.Alias != "" {
				srv.defaultImageModel = m.Alias
			} else {
				srv.defaultImageModel = m.Model
			}
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
			model:     m.Model,
			maxTokens: m.MaxTokens,
			alias:     m.Alias,
			limiter:   limiter,
			client:    m.Client,
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
	client, err := ai.getChatClientByModel(ctx, request.Model)
	if err != nil {
		return nil, err
	}

	request.Model = ai.models[request.Model].model
	modelMaxTokens := ai.models[request.Model].maxTokens
	if modelMaxTokens != 0 {
		request.MaxOutputTokens = modelMaxTokens
	}
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

func (ai *AiServiceImpl) FunctionCall(ctx context.Context, request *client.ChatRequest) (*client.FunctionCallResponse, error) {
	if request.Model == "" {
		request.Model = ai.defaultModel
	}
	client, err := ai.getChatClientByModel(ctx, request.Model)
	if err != nil {
		return nil, err
	}

	request.Model = ai.models[request.Model].model
	modelMaxTokens := ai.models[request.Model].maxTokens
	if modelMaxTokens != 0 {
		request.MaxOutputTokens = modelMaxTokens
	}
	ai.logger.Debugln("send function call request", "model", request.Model, "temperature", request.Temperature)
	for _, message := range request.Messages {
		ai.logger.Debugf("[%s]message: \n%s", message.Role, message.Content)
	}
	resp, err := client.FunctionCall(ctx, request)
	if err != nil {
		return nil, err
	}
	ai.logger.Debugln("function call response reveived", "total_tokens", resp.Tokens, "text_message", resp.Text)
	ai.logger.Debugln("content:", resp.FunctionCalls)
	return resp, nil
}

func (ai *AiServiceImpl) ImageGen(ctx context.Context, request *client.ChatRequest) (*client.ImageGenResponse, error) {
	if request.ImageModel == "" {
		request.ImageModel = ai.defaultImageModel
	}
	aiClient, err := ai.getChatClientByModel(ctx, request.ImageModel)
	if err != nil {
		return nil, err
	}

	request.ImageModel = ai.models[request.ImageModel].model
	modelMaxTokens := ai.models[request.ImageModel].maxTokens
	if modelMaxTokens != 0 {
		request.MaxOutputTokens = modelMaxTokens
	}

	ai.logger.Debugln("send image generate request", "model", request.ImageModel, "temperature", request.Temperature)
	for _, message := range request.Messages {
		ai.logger.Debugf("[%s]message\n", message.Role)
		for _, c := range message.Content {
			data := c.Data
			if c.Type == client.ContentTypeImage {
				data = "{image data}"
			}
			ai.logger.Debugf("[%s]message part: \n%s", c.Type, data)
		}
	}
	resp, err := aiClient.ImageGen(ctx, request)
	if err != nil {
		return nil, err
	}
	ai.logger.Debugln("image generate response reveived", "total_tokens", resp.Tokens)
	return resp, nil
}

func (ai *AiServiceImpl) getChatClientByModel(ctx context.Context, model string) (client.ChatClient, error) {
	for _, m := range ai.models {
		if m.model == model || m.alias == model {
			if m.limiter != nil {
				err := m.limiter.Wait(ctx)
				if err != nil {
					return nil, err
				}
			}
			return ai.clients[m.client], nil
		}
	}
	return nil, fmt.Errorf("client not found for %s", model)
}

type ModelList struct {
	Models  []string `json:"models"`
	Default string   `json:"default"`
}

func (ai *AiServiceImpl) ListModels(ctx context.Context) *ModelList {
	models := []string{}
	var defaultModel string
	for key, model := range ai.models {
		if model.alias != "" && key != model.alias {
			// only keep alias name in return list
			continue
		}
		var name string
		if model.alias != "" {
			name = model.alias
		} else {
			name = model.model
		}
		models = append(models, name)
		if key == ai.defaultModel {
			defaultModel = name
		}
	}
	slices.Sort(models)
	return &ModelList{
		Models:  models,
		Default: defaultModel,
	}
}
