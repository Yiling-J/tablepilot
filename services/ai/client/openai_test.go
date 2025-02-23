package client

import (
	"context"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockOpenAICompletionService struct {
	new func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (res *openai.ChatCompletion, err error)
}

func (s *mockOpenAICompletionService) New(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (res *openai.ChatCompletion, err error) {
	return s.new(ctx, body, opts...)
}

func TestClient_OpenAIArraySchema(t *testing.T) {
	var params openai.ChatCompletionNewParams
	m := &mockOpenAICompletionService{
		new: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (res *openai.ChatCompletion, err error) {
			params = body
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{
					Content: `{"data": "foobar"}`,
				}}},
				Usage: openai.CompletionUsage{TotalTokens: 100},
			}, nil
		},
	}
	client := NewOpenAIClient(m, zap.NewNop().Sugar())
	resp, err := client.Chat(context.TODO(), &ChatRequest{
		Messages:        []*Message{UserMessage("abc")},
		Temperature:     0.32,
		PresencePenalty: 0.45,
		Model:           "model",
		MaxOutputTokens: 1200,
		Schema:          &jsonschema.Schema{Type: "array", Version: "v1"},
	})
	require.NoError(t, err)
	require.Equal(t, "foobar", resp.Content)
	require.Equal(t, 100, int(resp.Tokens))
	b, err := params.MarshalJSON()
	require.NoError(t, err)
	expected := `{"messages":[{"content":[{"text":"abc","type":"text"}],"role":"user"}],"model":"model","presence_penalty":0.45,"response_format":{"json_schema":{"description":"schema for table","name":"schema","schema":{"properties":{"data":{"$schema":"v1","type":"array"}},"type":"object"}},"type":"json_schema"},"temperature":0.32}`
	require.Equal(t, expected, string(b))
}

func TestClient_OpenAIObjectSchema(t *testing.T) {
	var params openai.ChatCompletionNewParams
	m := &mockOpenAICompletionService{
		new: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (res *openai.ChatCompletion, err error) {
			params = body
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{
					Content: "foobar",
				}}},
				Usage: openai.CompletionUsage{TotalTokens: 100},
			}, nil
		},
	}
	client := NewOpenAIClient(m, zap.NewNop().Sugar())
	resp, err := client.Chat(context.TODO(), &ChatRequest{
		Messages:        []*Message{UserMessage("abc")},
		Temperature:     0.32,
		PresencePenalty: 0.45,
		Model:           "model",
		MaxOutputTokens: 1200,
		Schema:          &jsonschema.Schema{Type: "string", Version: "v1"},
	})
	require.NoError(t, err)
	require.Equal(t, "foobar", resp.Content)
	require.Equal(t, 100, int(resp.Tokens))
	b, err := params.MarshalJSON()
	require.NoError(t, err)
	expected := `{"messages":[{"content":[{"text":"abc","type":"text"}],"role":"user"}],"model":"model","presence_penalty":0.45,"response_format":{"json_schema":{"description":"schema for table","name":"schema","schema":{"$schema":"v1","type":"string"}},"type":"json_schema"},"temperature":0.32}`
	require.Equal(t, expected, string(b))
}
