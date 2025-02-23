package ai

import (
	"context"
	"fmt"
	"tablepilot/config"
	"tablepilot/services/ai/client"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockChatClient struct {
	chat func(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error)
}

func (m *MockChatClient) Chat(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
	return m.chat(ctx, request)
}

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
			chatClient := &MockChatClient{
				chat: func(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
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
