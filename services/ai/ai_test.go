package ai

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/provider"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAIService_Chat(t *testing.T) {
	cases := []struct {
		inputModel string
		chatModel  string
		error      bool
	}{
		{inputModel: "", chatModel: "default"},
		{inputModel: "exists", chatModel: "exists"},
		{inputModel: "missing", error: true},
		{inputModel: "alias", chatModel: "exists"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			var req *client.ChatRequest
			chatClient := &client.ChatClientMock{
				ChatFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
					req = request
					return &client.ChatResponse{}, nil
				},
			}
			srv, err := NewAiService(&config.Config{}, &provider.ProviderServiceMock{
				BuildProvidersFunc: func(ctx context.Context) error { return nil },
				ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
					return []provider.Provider{
						{Name: "chat", Type: "openai", Models: []provider.Model{
							{Model: "default", Client: "chat"},
							{Model: "exists", Alias: "alias", Client: "chat"},
						}, Enabled: true},
					}, nil
				},
				WithSyncCallbackFunc: func(callback func(ctx context.Context, providers []provider.Provider)) {},
			}, zap.NewNop().Sugar())
			require.NoError(t, err)
			srv.defaultModel = "default"
			srv.clients = map[string]client.ChatClient{
				"chat": chatClient,
			}

			_, err = srv.Chat(context.TODO(), &client.ChatRequest{
				Messages:        []*client.Message{client.UserMessage("foo")},
				Temperature:     0.35,
				Schema:          &jsonschema.Schema{Version: "v-test"},
				Model:           c.inputModel,
				MaxOutputTokens: 1234,
				PresencePenalty: 1.0,
			})
			if c.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.chatModel, req.Model)
				require.Equal(t, 1234, int(req.MaxOutputTokens))
				require.Equal(t, 0.35, req.Temperature)
				require.Equal(t, 1.0, req.PresencePenalty)
				require.Equal(t, []*client.Message{client.UserMessage("foo")}, req.Messages)
				require.Equal(t, &jsonschema.Schema{Version: "v-test"}, req.Schema)
			}
		})
	}
}

func TestAIService_ListModels(t *testing.T) {
	ctx := context.TODO()
	srv, err := NewAiService(&config.Config{}, &provider.ProviderServiceMock{
		BuildProvidersFunc: func(ctx context.Context) error { return nil },
		ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
			return []provider.Provider{
				{Name: "chat", Type: "openai", Models: []provider.Model{
					{Model: "m1", Client: "chat"},
					{Model: "m2", Client: "chat", Alias: "m2a"},
				}, Enabled: true},
			}, nil
		},
		WithSyncCallbackFunc: func(callback func(ctx context.Context, providers []provider.Provider)) {},
	}, zap.NewNop().Sugar())
	require.NoError(t, err)
	srv.clients = map[string]client.ChatClient{
		"chat": nil,
	}
	ml := srv.ListModels(ctx)
	require.Equal(t, &ModelList{
		DefaultModel: "m1",
		Models:       []ModelListItem{{Name: "m1"}, {Name: "m2a"}},
	}, ml)

	srv, err = NewAiService(&config.Config{}, &provider.ProviderServiceMock{
		BuildProvidersFunc: func(ctx context.Context) error { return nil },
		ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
			return []provider.Provider{
				{Name: "chat", Type: "openai", Models: []provider.Model{
					{Model: "m2", Client: "chat", Alias: "m2a"},
					{Model: "m1", Client: "chat"},
				}, Enabled: true},
			}, nil
		},
		WithSyncCallbackFunc: func(callback func(ctx context.Context, providers []provider.Provider)) {},
	}, zap.NewNop().Sugar())
	require.NoError(t, err)
	srv.clients = map[string]client.ChatClient{
		"chat": nil,
	}
	ml = srv.ListModels(ctx)
	require.Equal(t, &ModelList{
		DefaultModel: "m2a",
		Models:       []ModelListItem{{Name: "m1"}, {Name: "m2a"}},
	}, ml)

	srv, err = NewAiService(&config.Config{}, &provider.ProviderServiceMock{
		BuildProvidersFunc: func(ctx context.Context) error { return nil },
		ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
			return []provider.Provider{
				{Name: "chat", Type: "openai", Models: []provider.Model{
					{Model: "m1", Client: "chat"},
					{Model: "m2", Client: "chat", Default: true},
				}, Enabled: true},
			}, nil
		},
		WithSyncCallbackFunc: func(callback func(ctx context.Context, providers []provider.Provider)) {},
	}, zap.NewNop().Sugar())
	require.NoError(t, err)
	srv.clients = map[string]client.ChatClient{
		"chat": nil,
	}
	ml = srv.ListModels(ctx)
	require.Equal(t, &ModelList{
		DefaultModel: "m2",
		Models:       []ModelListItem{{Name: "m1"}, {Name: "m2"}},
	}, ml)
}

func TestAIService_ChatModelLimiter(t *testing.T) {
	chatClient := &client.ChatClientMock{
		ChatFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
			return &client.ChatResponse{}, nil
		},
	}
	srv, err := NewAiService(&config.Config{}, &provider.ProviderServiceMock{
		BuildProvidersFunc: func(ctx context.Context) error { return nil },
		ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
			return []provider.Provider{
				{Name: "chat", Type: "openai", Models: []provider.Model{
					{Model: "default", Client: "chat", Rpm: 20},
					{Model: "default2", Client: "chat"},
				}, Enabled: true},
			}, nil
		},
		WithSyncCallbackFunc: func(callback func(ctx context.Context, providers []provider.Provider)) {},
	}, zap.NewNop().Sugar())
	srv.clients = map[string]client.ChatClient{
		"chat": chatClient,
	}
	require.NoError(t, err)
	srv.defaultModel = "default"

	now := time.Now()
	for range 25 {
		_, err = srv.Chat(context.TODO(), &client.ChatRequest{
			Messages:        []*client.Message{client.UserMessage("foo")},
			Temperature:     0.35,
			Schema:          &jsonschema.Schema{Version: "v-test"},
			Model:           "default",
			MaxOutputTokens: 1234,
			PresencePenalty: 1.0,
		})
		require.NoError(t, err)
	}
	delta := time.Since(now).Seconds()
	require.True(t, delta > 10)
	require.True(t, delta < 20)

	now = time.Now()
	for range 25 {
		_, err = srv.Chat(context.TODO(), &client.ChatRequest{
			Messages:        []*client.Message{client.UserMessage("foo")},
			Temperature:     0.35,
			Schema:          &jsonschema.Schema{Version: "v-test"},
			Model:           "default2",
			MaxOutputTokens: 1234,
			PresencePenalty: 1.0,
		})
		require.NoError(t, err)
	}
	delta = time.Since(now).Seconds()
	require.True(t, delta < 1)
}

func TestAIService_syncProviders(t *testing.T) {
	ctx := context.TODO()
	srv, err := NewAiService(&config.Config{}, &provider.ProviderServiceMock{
		BuildProvidersFunc: func(ctx context.Context) error { return nil },
		ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
			return []provider.Provider{
				{Name: "chat", Type: "openai", Models: []provider.Model{
					{Model: "default", Client: "chat"},
				}, Enabled: true},
			}, nil
		},
		WithSyncCallbackFunc: func(callback func(ctx context.Context, providers []provider.Provider)) {},
	}, zap.NewNop().Sugar())
	srv.clients = map[string]client.ChatClient{
		"chat": nil,
	}
	require.NoError(t, err)

	srv.syncProviders(ctx,
		[]provider.Provider{
			{Name: "chat", Type: "openai", Models: []provider.Model{
				{Model: "default", Client: "chat"},
			}, Enabled: true},
			{Name: "c2", Type: "openai", Models: []provider.Model{
				{Model: "c2m", Client: "c2"},
			}, Enabled: true},
		})
	models := srv.ListModels(ctx)
	require.Equal(t, []ModelListItem{{Name: "c2m", Image: false}, {Name: "default", Image: false}}, models.Models)

	srv.syncProviders(ctx,
		[]provider.Provider{
			{Name: "chat", Type: "openai", Models: []provider.Model{
				{Model: "default", Client: "chat"},
			}, Enabled: true},
			{Name: "c2", Type: "openai", Models: []provider.Model{
				{Model: "c2mm", Client: "c2"},
			}, Enabled: true},
		})
	models = srv.ListModels(ctx)
	require.Equal(t, []ModelListItem{{Name: "c2mm", Image: false}, {Name: "default", Image: false}}, models.Models)

	srv.syncProviders(ctx,
		[]provider.Provider{
			{Name: "chat", Type: "openai", Models: []provider.Model{
				{Model: "default", Client: "chat"},
			}, Enabled: true},
			{Name: "c2", Type: "openai", Models: []provider.Model{
				{Model: "c2mm", Client: "c2"},
				{Model: "c2mt", Alias: "ct", Client: "c2"},
			}, Enabled: true},
		})
	models = srv.ListModels(ctx)
	require.Equal(t, []ModelListItem{{Name: "c2mm", Image: false}, {Name: "ct", Image: false}, {Name: "default", Image: false}}, models.Models)

	srv.syncProviders(ctx,
		[]provider.Provider{
			{Name: "chat", Type: "openai", Models: []provider.Model{
				{Model: "default", Client: "chat"},
			}, Enabled: true},
			{Name: "c2", Type: "openai", Models: []provider.Model{
				{Model: "c2mm", Client: "c2"},
			}, Enabled: true},
		})
	models = srv.ListModels(ctx)
	require.Equal(t, []ModelListItem{{Name: "c2mm", Image: false}, {Name: "default", Image: false}}, models.Models)

	srv.syncProviders(ctx,
		[]provider.Provider{
			{Name: "chat", Type: "openai", Models: []provider.Model{
				{Model: "default", Client: "chat"},
			}, Enabled: true},
		})
	models = srv.ListModels(ctx)
	require.Equal(t, []ModelListItem{{Name: "default", Image: false}}, models.Models)

	srv.syncProviders(ctx,
		[]provider.Provider{
			{Name: "chat", Type: "openai", Models: []provider.Model{
				{Model: "default", Client: "chat"},
			}, Enabled: false},
		})
	models = srv.ListModels(ctx)
	require.Empty(t, models.Models)
}
