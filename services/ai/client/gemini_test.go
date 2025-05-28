package client

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/genai"
)

func TestGeminiClient_ImageGen(t *testing.T) {
	cases := []struct {
		name     string
		parts    []*genai.Part
		expected map[string][]byte
	}{
		{
			name: "simple",
			parts: []*genai.Part{
				{Text: `<info row_id="r0" column_id="c0" />`},
				{InlineData: &genai.Blob{Data: []byte("foo")}},
			},
			expected: map[string][]byte{"r0-c0": []byte("foo")},
		},
		{
			name: "simple 2",
			parts: []*genai.Part{
				{Text: `<info row_id="r0" column_id="c0"/>`},
				{InlineData: &genai.Blob{Data: []byte("foo")}},
			},
			expected: map[string][]byte{"r0-c0": []byte("foo")},
		},
		{
			name: "three images simple",
			parts: []*genai.Part{
				{Text: `<info row_id="r0" column_id="c0" />`},
				{InlineData: &genai.Blob{Data: []byte("foo")}},
				{Text: `<info row_id="r0" column_id="c1" />`},
				{InlineData: &genai.Blob{Data: []byte("bar")}},
				{Text: `<info row_id="r0" column_id="c2" />`},
				{InlineData: &genai.Blob{Data: []byte("baz")}},
			},
			expected: map[string][]byte{
				"r0-c0": []byte("foo"),
				"r0-c1": []byte("bar"),
				"r0-c2": []byte("baz"),
			},
		},
		{
			name: "three images, image first",
			parts: []*genai.Part{
				{InlineData: &genai.Blob{Data: []byte("foo")}},
				{Text: `<info row_id="r0" column_id="c0" />`},
				{InlineData: &genai.Blob{Data: []byte("bar")}},
				{Text: `<info row_id="r0" column_id="c1" />`},
				{InlineData: &genai.Blob{Data: []byte("baz")}},
				{Text: `<info row_id="r0" column_id="c2" />`},
			},
			expected: map[string][]byte{
				"r0-c0": []byte("foo"),
				"r0-c1": []byte("bar"),
				"r0-c2": []byte("baz"),
			},
		},
		{
			name: "three images, mix order",
			parts: []*genai.Part{
				{InlineData: &genai.Blob{Data: []byte("foo")}},
				{Text: `<info row_id="r0" column_id="c0" />`},
				{Text: `<info row_id="r0" column_id="c1" />`},
				{InlineData: &genai.Blob{Data: []byte("bar")}},
				{InlineData: &genai.Blob{Data: []byte("baz")}},
				{Text: `<info row_id="r0" column_id="c2" />`},
			},
			expected: map[string][]byte{
				"r0-c0": []byte("foo"),
				"r0-c1": []byte("bar"),
				"r0-c2": []byte("baz"),
			},
		},
		{
			name: "three images, duplicate and unknown",
			parts: []*genai.Part{
				{Text: `<info unknown="abc" />`},
				{Text: `<info row_id="r0" column_id="c0" />`},
				{InlineData: &genai.Blob{Data: []byte("foo")}},
				{Text: `<info row_id="r0" column_id="c0" />`},
				{Text: `<info row_id="r0" column_id="c1" />`},
				{InlineData: &genai.Blob{Data: []byte("bar")}},
				{Text: `<info row_id="r0" column_id="c2" />`},
				{Text: `<info unknown="def" />`},
				{InlineData: &genai.Blob{Data: []byte("baz")}},
			},
			expected: map[string][]byte{
				"r0-c0": []byte("foo"),
				"r0-c1": []byte("bar"),
				"r0-c2": []byte("baz"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := &GenaiModelServiceMock{
				GenerateContentFunc: func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
					require.Equal(t, "bar", model)
					require.Equal(t, []*genai.Content{{
						Role: "user",
						Parts: []*genai.Part{
							genai.NewPartFromText("foo"),
							genai.NewPartFromBytes([]byte("abc"), "image/png"),
							genai.NewPartFromText("bar"),
							genai.NewPartFromBytes([]byte("def"), "image/jpeg"),
						},
					}}, contents)
					return &genai.GenerateContentResponse{
						Candidates: []*genai.Candidate{
							{Content: &genai.Content{
								Parts: tc.parts,
							}},
						},
					}, nil
				},
			}
			client := &GeminiClient{
				modelService: srv,
				logger:       zap.NewNop().Sugar(),
			}
			resp, err := client.ImageGen(context.TODO(), &ChatRequest{
				Temperature: 0.85,
				Model:       "foo",
				ImageModel:  "bar",
				Messages: []*Message{
					{Role: "user", Content: []Content{
						{Type: ContentTypeText, Data: "foo"},
						{Type: ContentTypeImage, Data: "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("abc"))},
						{Type: ContentTypeText, Data: "bar"},
						{Type: ContentTypeImage, Data: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString([]byte("def"))},
					}},
				},
			})
			require.NoError(t, err)
			require.Equal(t, tc.expected, resp.Images)
		})
	}
}

func TestGeminiClient_Chat(t *testing.T) {
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
	client := &GeminiClient{
		modelService: &GenaiModelServiceMock{},
		logger:       zap.NewNop().Sugar(),
		oaiClient:    NewOpenAIClient(m, zap.NewNop().Sugar()),
	}
	resp, err := client.Chat(t.Context(), &ChatRequest{
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

func TestGeminiClient_FunctionCall(t *testing.T) {
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
	client := &GeminiClient{
		modelService: &GenaiModelServiceMock{},
		logger:       zap.NewNop().Sugar(),
		oaiClient:    NewOpenAIClient(m, zap.NewNop().Sugar()),
	}
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
