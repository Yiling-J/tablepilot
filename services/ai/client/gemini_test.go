package client

import (
	"context"
	"encoding/base64"
	"testing"

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
