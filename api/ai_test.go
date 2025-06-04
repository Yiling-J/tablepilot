package api

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/stretchr/testify/require"
)

func TestAPI_GenerateList(t *testing.T) {
	aiMock := &ai.AiServiceMock{
		GenerateListOptionsFunc: func(ctx context.Context, req ai.GenerateListOptionsRequest) ([]string, error) {
			require.Equal(t, "m1", req.Model)
			require.Equal(t, "foobar", req.Prompt)
			require.Equal(t, []string{"z"}, req.Options)
			return []string{"foo", "bar"}, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.AIService = aiMock
	})

	req, err := server.NewPostRequest("/api/v1/ai/list_gen", &ai.GenerateListOptionsRequest{
		Model:   "m1",
		Prompt:  "foobar",
		Options: []string{"z"},
	})
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, []string{"foo", "bar"})
}
