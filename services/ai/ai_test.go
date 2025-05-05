package ai

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/services/ai/client"

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
			srv, err := NewAiService(&config.Config{
				Models: []config.Model{
					{Model: "default", Client: "chat"},
					{Model: "exists", Alias: "alias", Client: "chat"},
				},
			}, map[string]client.ChatClient{
				"chat": chatClient,
			}, zap.NewNop().Sugar())
			require.NoError(t, err)
			srv.defaultModel = "default"

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
	srv, err := NewAiService(&config.Config{
		Models: []config.Model{
			{Model: "m1", Client: "chat"},
			{Model: "m2", Alias: "m2a", Client: "chat"},
		},
	}, map[string]client.ChatClient{
		"chat": nil,
	}, zap.NewNop().Sugar())
	require.NoError(t, err)
	ml := srv.ListModels(ctx)
	require.Equal(t, &ModelList{
		Default: "m1",
		Models:  []string{"m1", "m2a"},
	}, ml)

	srv, err = NewAiService(&config.Config{
		Models: []config.Model{
			{Model: "m2", Alias: "m2a", Client: "chat"},
			{Model: "m1", Client: "chat"},
		},
	}, map[string]client.ChatClient{
		"chat": nil,
	}, zap.NewNop().Sugar())
	require.NoError(t, err)
	ml = srv.ListModels(ctx)
	require.Equal(t, &ModelList{
		Default: "m2a",
		Models:  []string{"m1", "m2a"},
	}, ml)

	srv, err = NewAiService(&config.Config{
		Models: []config.Model{
			{Model: "m1", Client: "chat"},
			{Model: "m2", Client: "chat", Default: true},
		},
	}, map[string]client.ChatClient{
		"chat": nil,
	}, zap.NewNop().Sugar())
	require.NoError(t, err)
	ml = srv.ListModels(ctx)
	require.Equal(t, &ModelList{
		Default: "m2",
		Models:  []string{"m1", "m2"},
	}, ml)
}

func TestAIService_ChatModelLimiter(t *testing.T) {
	chatClient := &client.ChatClientMock{
		ChatFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
			return &client.ChatResponse{}, nil
		},
	}
	srv, err := NewAiService(&config.Config{
		Models: []config.Model{
			{Model: "default", Client: "chat", RPM: 20},
			{Model: "default2", Client: "chat"},
		},
	}, map[string]client.ChatClient{
		"chat": chatClient,
	}, zap.NewNop().Sugar())
	require.NoError(t, err)
	srv.defaultModel = "default"

	now := time.Now()
	for i := 0; i < 25; i++ {
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
	for i := 0; i < 25; i++ {
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
