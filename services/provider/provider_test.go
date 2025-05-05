package provider

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestProviderService_ListProviders(t *testing.T) {
	db := db.NewTestDB()
	srv := NewProviderService(&config.Config{}, db, zap.NewNop().Sugar())
	srv.providers = []Provider{
		{
			Name: "p1", Type: "openai", Key: "k", BaseURL: "b",
			Models: []Model{
				{Model: "m", Alias: "ma", MaxTokens: 100, Rpm: 5},
				{Model: "m2", Alias: "ma2", MaxTokens: 200, Rpm: 25},
				{Model: "mi", Alias: "mia", MaxTokens: 10, Rpm: 3, Image: true},
			},
		},
		{
			ID: 2, Name: "p2", Type: "gemini", Key: "k2", BaseURL: "b2",
			Models: []Model{
				{Model: "mg", Alias: "mga", MaxTokens: 100, Rpm: 5},
			},
		},
	}

	providers, err := srv.ListProviders(context.TODO())
	require.NoError(t, err)
	require.Equal(t, srv.providers, providers)
}

func TestProviderService_genProviders(t *testing.T) {
	db := db.NewTestDB()
	srv := NewProviderService(&config.Config{
		Clients: []config.Client{
			&config.OpenAI{
				Name:    "p1",
				Key:     "k",
				Type:    "openai",
				BaseURL: "b",
			},
		},
		Models: []config.Model{
			{Model: "m", Alias: "ma", Client: "p1", MaxTokens: 100, RPM: 5},
			{Model: "m2", Alias: "ma2", Client: "p1", MaxTokens: 200, RPM: 25},
			{Model: "mi", Alias: "mia", Client: "p1", MaxTokens: 10, RPM: 3, Image: true},
		},
	}, db, zap.NewNop().Sugar())
	err := srv.CreateProvider(context.TODO(), Provider{
		ID: 2, Name: "p2", Type: "gemini", Key: "k2", BaseURL: "b2",
		Models: []Model{
			{Model: "mg", Alias: "mga", MaxTokens: 100, Rpm: 5},
			{Model: "mgi", Alias: "mgia", Client: "p2", MaxTokens: 1, Rpm: 1, Image: true},
		},
	})
	require.NoError(t, err)
	p, err := db.Provider.Query().First(context.TODO())
	require.NoError(t, err)

	expected := []Provider{
		{
			Name: "p1", Type: "openai", Key: "k", BaseURL: "b", Editable: false,
			Models: []Model{
				{Model: "m", Alias: "ma", MaxTokens: 100, Rpm: 5, Client: "p1"},
				{Model: "m2", Alias: "ma2", MaxTokens: 200, Rpm: 25, Client: "p1"},
				{Model: "mi", Alias: "mia", MaxTokens: 10, Rpm: 3, Client: "p1", Image: true},
			},
		},
		{
			ID: p.ID, Name: "p2", Type: "gemini", Key: "k2", BaseURL: "b2", Editable: true,
			Models: []Model{
				{Model: "mg", Alias: "mga", MaxTokens: 100, Rpm: 5, Client: "p2"},
				{Model: "mgi", Alias: "mgia", Client: "p2", MaxTokens: 1, Rpm: 1, Image: true},
			},
		},
	}

	providers, err := srv.genProviders(context.TODO())
	require.NoError(t, err)
	require.Equal(t, expected, providers)
}
