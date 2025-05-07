package provider

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent/provider"
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

func TestProviderService_CreateProvider(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.TODO()
	srv := NewProviderService(&config.Config{}, db, zap.NewNop().Sugar())
	err := srv.CreateProvider(ctx, Provider{
		Name:    "p1",
		Type:    "openai",
		Key:     "k",
		BaseURL: "b",
		Models: []Model{
			{Model: "m1", Alias: "m", Rpm: 12, MaxTokens: 100},
		},
	})
	require.NoError(t, err)
	p, err := db.Provider.Query().Where(provider.Name("p1")).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, "k", p.Key)
	require.Equal(t, "b", p.BaseURL)
	require.Equal(t, provider.TypeOpenai, p.Type)
	models, err := p.QueryModels().All(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(models))
	require.Equal(t, "m1", models[0].Model)
	require.Equal(t, "m", models[0].Alias)
	require.Equal(t, 12, models[0].Rpm)
	require.Equal(t, 100, models[0].MaxTokens)
}

func TestProviderService_UpdateProvider(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.TODO()
	srv := NewProviderService(&config.Config{}, db, zap.NewNop().Sugar())
	err := srv.CreateProvider(ctx, Provider{
		Name:    "p1",
		Type:    "openai",
		Key:     "k",
		BaseURL: "b",
		Models: []Model{
			{Model: "m1", Alias: "m"},
			{Model: "m2"},
		},
	})
	require.NoError(t, err)
	p, err := db.Provider.Query().Where(provider.Name("p1")).Only(ctx)
	require.NoError(t, err)

	err = srv.UpdateProvider(ctx, p.ID, Provider{
		Name:    "p1",
		Type:    "openai",
		Key:     "k",
		BaseURL: "b",
		Models: []Model{
			{Model: "m1", Alias: "mm", Rpm: 12, MaxTokens: 100},
			{Model: "m3"},
		},
	})
	require.NoError(t, err)
	models, err := p.QueryModels().All(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(models))
	require.Equal(t, "m1", models[0].Model)
	require.Equal(t, "mm", models[0].Alias)
	require.Equal(t, 12, models[0].Rpm)
	require.Equal(t, 100, models[0].MaxTokens)
	require.Equal(t, "m3", models[1].Model)
}

func TestProviderService_DeleteProvider(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.TODO()
	srv := NewProviderService(&config.Config{}, db, zap.NewNop().Sugar())
	err := srv.CreateProvider(ctx, Provider{
		Name:    "p1",
		Type:    "openai",
		Key:     "k",
		BaseURL: "b",
		Models: []Model{
			{Model: "m1", Alias: "m", Rpm: 12, MaxTokens: 100},
		},
	})
	require.NoError(t, err)
	err = srv.CreateProvider(ctx, Provider{
		Name:    "p2",
		Type:    "openai",
		Key:     "k",
		BaseURL: "b",
		Models: []Model{
			{Model: "m2", Alias: "mm", Rpm: 12, MaxTokens: 100},
		},
	})
	require.NoError(t, err)
	p, err := db.Provider.Query().Where(provider.Name("p1")).Only(ctx)
	require.NoError(t, err)

	err = srv.DeleteProvider(ctx, p.ID)
	require.NoError(t, err)

	providers, err := db.Provider.Query().All(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(providers))
	require.Equal(t, "p2", providers[0].Name)

	models, err := providers[0].QueryModels().All(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(models))
	require.Equal(t, "m2", models[0].Model)
	modelsDB, err := db.Model.Query().All(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(modelsDB))
	require.Equal(t, models[0].ID, modelsDB[0].ID)
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
