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
	expected := `{"messages":[{"content":"abc","role":"user"}],"model":"model","max_completion_tokens":1200,"presence_penalty":0.45,"temperature":0.32,"response_format":{"json_schema":{"name":"schema","description":"schema for table","schema":{"properties":{"data":{"$schema":"v1","type":"array"}},"type":"object"}},"type":"json_schema"}}`
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
	expected := `{"messages":[{"content":"abc","role":"user"}],"model":"model","max_completion_tokens":1200,"presence_penalty":0.45,"temperature":0.32,"response_format":{"json_schema":{"name":"schema","description":"schema for table","schema":{"$schema":"v1","type":"string"}},"type":"json_schema"}}`
	require.Equal(t, expected, string(b))
}

func TestClient_OpenAIContentTypeImage(t *testing.T) {
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
		Messages: []*Message{UserMessageWithImages("foo", map[string]string{
			"i1.png": "i1",
			"i2.png": "i2",
		})},
		Schema: &jsonschema.Schema{Type: "string", Version: "v1"},
	})
	require.NoError(t, err)
	require.Equal(t, "foobar", resp.Content)
	b, err := params.MarshalJSON()
	require.NoError(t, err)
	expected := `{"messages":[{"content":"foo","role":"user"},{"content":"\nBelow is the image with ID: \u003ci1.png\u003e","role":"user"},{"content":[{"image_url":{"url":"i1"},"type":"image_url"}],"role":"user"},{"content":"\nBelow is the image with ID: \u003ci2.png\u003e","role":"user"},{"content":[{"image_url":{"url":"i2"},"type":"image_url"}],"role":"user"}],"max_completion_tokens":0,"presence_penalty":0,"temperature":0,"response_format":{"json_schema":{"name":"schema","description":"schema for table","schema":{"$schema":"v1","type":"string"}},"type":"json_schema"}}`
	require.Equal(t, expected, string(b))
}

func TestClient_OpenAIFunctionCall(t *testing.T) {
	m := &mockOpenAICompletionService{
		new: func(ctx context.Context, body openai.ChatCompletionNewParams, opts ...option.RequestOption) (res *openai.ChatCompletion, err error) {
			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{{Message: openai.ChatCompletionMessage{
					Content: "foobar",
					ToolCalls: []openai.ChatCompletionMessageToolCall{
						{Function: openai.ChatCompletionMessageToolCallFunction{
							Name:      "t1",
							Arguments: `{"p1": 12}`,
						}},
					},
				}}},
				Usage: openai.CompletionUsage{TotalTokens: 100},
			}, nil
		},
	}
	client := NewOpenAIClient(m, zap.NewNop().Sugar())
	resp, err := client.FunctionCall(context.TODO(), &ChatRequest{
		Messages: []*Message{UserMessage("abc")},
		Tools: []ChatTool{
			{Name: "t1", Description: "td1", Parameters: []ChatToolParam{{
				Name: "p1",
				Type: "int",
			}}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "foobar", resp.Text)
	require.Equal(t, FunctionCall{Name: "t1", Arguments: map[string]any{"p1": 12.0}}, resp.FunctionCalls[0])
}
