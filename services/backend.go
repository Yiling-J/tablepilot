package services

import (
	"fmt"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/provider"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type Backend struct {
	Config          *config.Config
	DB              *ent.Client
	Logger          *zap.SugaredLogger
	AIService       ai.AiService
	TableService    table.TableService
	ProviderService provider.ProviderService
}

func NewBackend(
	config *config.Config, db *ent.Client,
	logger *zap.SugaredLogger, aiService ai.AiService, tableService table.TableService, providerService provider.ProviderService,
) *Backend {
	return &Backend{
		Config:          config,
		DB:              db,
		Logger:          logger,
		AIService:       aiService,
		TableService:    tableService,
		ProviderService: providerService,
	}
}

func CreateBackend(cmd *cobra.Command, verbose bool) *Backend {
	container := dig.New()

	err := container.Provide(func() (*config.Config, error) {
		cfg, err := config.NewConfig(cmd.Flag("config").Value.String())
		if err != nil {
			return nil, fmt.Errorf("services.CreateBackend: creating config: %w", err)
		}
		return cfg, nil
	})
	if err != nil {
		panic(err)
	}

	err = container.Provide(func(config *config.Config) (*zap.SugaredLogger, error) {
		var cfg zap.Config
		if verbose {
			cfg = zap.NewDevelopmentConfig()
		} else {
			cfg = zap.NewProductionConfig()
		}
		// serve command will start a server, use JSON encoding logs for API calls
		// for CLI set log encoding to console
		if cmd.Name() != "serve" {
			cfg.Encoding = "console"
		}
		l, err := cfg.Build()
		if err != nil {
			return nil, fmt.Errorf("services.CreateBackend: building logger: %w", err)
		}
		return l.Sugar(), nil
	})
	if err != nil {
		panic(err)
	}

	err = container.Provide(db.NewDB)
	if err != nil {
		panic(err)
	}

	err = container.Provide(provider.NewProviderService, dig.As(new((provider.ProviderService))))
	if err != nil {
		panic(err)
	}

	err = container.Provide(ai.NewAiService, dig.As(new((ai.AiService))))
	if err != nil {
		panic(err)
	}

	err = container.Provide(table.NewTableService, dig.As(new((table.TableService))))
	if err != nil {
		panic(err)
	}

	err = container.Provide(NewBackend)
	if err != nil {
		panic(err)
	}

	var backend *Backend
	err = container.Invoke(func(b *Backend) {
		backend = b
	})
	if err != nil {
		panic(err)
	}
	return backend
}
