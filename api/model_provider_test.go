package api

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/provider"
	"github.com/stretchr/testify/require"
)

func TestAPI_ListModels(t *testing.T) {
	expected := &ai.ModelList{
		DefaultModel: "foo",
		Models:       []ai.ModelListItem{{Name: "foo"}, {Name: "bar"}},
	}
	aiMock := &ai.AiServiceMock{
		ListModelsFunc: func(ctx context.Context) *ai.ModelList {
			return expected
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.AIService = aiMock
	})
	req, err := server.NewGetRequest("/api/v1/models")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, expected)
}

func TestAPI_GetProviders(t *testing.T) {
	providers := []provider.Provider{
		{ID: 1, Name: "p"},
	}
	providerMock := &provider.ProviderServiceMock{
		ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
			return providers, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewGetRequest("/api/v1/providers")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, providers,
	)
}

func TestAPI_CreateProvider(t *testing.T) {
	pr := provider.Provider{
		Name: "p",
	}
	providerMock := &provider.ProviderServiceMock{
		CreateProviderFunc: func(ctx context.Context, provider provider.Provider) error {
			require.Equal(t, pr, provider)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewPostRequest("/api/v1/providers", pr)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, "",
	)
}

func TestAPI_UpdateProvider(t *testing.T) {
	pr := provider.Provider{
		Name: "p",
	}
	providerMock := &provider.ProviderServiceMock{
		UpdateProviderFunc: func(ctx context.Context, id int, provider provider.Provider) error {
			require.Equal(t, 2, id)
			require.Equal(t, pr, provider)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewPatchRequest("/api/v1/providers/2", pr)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, "",
	)
}

func TestAPI_DeleteProvider(t *testing.T) {
	providerMock := &provider.ProviderServiceMock{
		DeleteProviderFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 2, id)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewDeleteRequest("/api/v1/providers/2")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, "",
	)
}
