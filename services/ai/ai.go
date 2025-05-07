package ai

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/provider"

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

type AiServiceImpl struct {
	providerService   provider.ProviderService
	clients           map[string]client.ChatClient
	models            map[string]*model
	logger            *zap.SugaredLogger
	defaultModel      string
	defaultImageModel string
	config            *config.Config
}

func (ai *AiServiceImpl) syncProviders(ctx context.Context, providers []provider.Provider) {
	clients := map[string]client.ChatClient{}
	models := map[string]*model{}
	defaultModel := ""
	defaultImageModel := ""
	for _, p := range providers {
		if ac, ok := ai.clients[p.Name]; !ok {
			c, err := client.NewClient(p, ai.logger)
			if err != nil {
				ai.logger.Errorw("sync providers failed", "error", err)
				return
			}
			clients[p.Name] = c
		} else {
			clients[p.Name] = ac
		}

		for _, m := range p.Models {
			_, ok := clients[m.Client]
			if !ok {
				ai.logger.Error("provider not found")
				return
			}

			var limiter *rate.Limiter
			if m.Rpm > 0 {
				limiter = rate.NewLimiter(
					rate.Every(time.Minute/time.Duration(m.Rpm)), m.Rpm,
				)
			}
			md := &model{
				model:     m.Model,
				maxTokens: m.MaxTokens,
				alias:     m.Alias,
				limiter:   limiter,
				client:    m.Client,
				image:     m.Image,
			}
			models[md.model] = md
			if md.alias != "" {
				models[md.alias] = md
			}
			if md.image {
				if defaultImageModel == "" || m.Default {
					if m.Alias != "" {
						defaultImageModel = m.Alias
					} else {
						defaultImageModel = m.Model
					}
				}
			} else {
				if defaultModel == "" || m.Default {
					if m.Alias != "" {
						defaultModel = m.Alias
					} else {
						defaultModel = m.Model
					}
				}
			}
		}
	}
	ai.clients = clients
	ai.models = models
	ai.defaultModel = defaultModel
	ai.defaultImageModel = defaultImageModel
}

func NewAiService(cfg *config.Config, providerService provider.ProviderService, logger *zap.SugaredLogger) (*AiServiceImpl, error) {
	srv := &AiServiceImpl{
		models:          map[string]*model{},
		logger:          logger.With("service", "ai"),
		config:          cfg,
		providerService: providerService,
	}
	ctx := context.Background()
	err := srv.providerService.BuildProviders(ctx)
	if err != nil {
		return nil, err
	}
	providers, err := srv.providerService.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	srv.syncProviders(ctx, providers)
	srv.providerService.WithSyncCallback(srv.syncProviders)
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

func (ai *AiServiceImpl) ListModels(ctx context.Context) *ModelList {
	list := &ModelList{}
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
		if _, ok := ai.clients[model.client]; !ok {
			continue
		}
		list.Models = append(list.Models, ModelListItem{
			Name:  name,
			Image: model.image,
		})
		if key == ai.defaultModel {
			list.DefaultModel = name
		}
		if key == ai.defaultImageModel {
			list.DefaultImageModel = name
		}
	}
	slices.SortFunc(list.Models, func(a, b ModelListItem) int {
		return strings.Compare(a.Name, b.Name)
	})
	return list
}
