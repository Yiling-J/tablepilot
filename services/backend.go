package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/Yiling-J/tablepilot/services/provider"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/workflow"
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
	WorkflowService workflow.WorkflowService
	DatasetService  dataset.DatasetService
}

func NewBackend(
	config *config.Config, db *ent.Client,
	logger *zap.SugaredLogger, aiService ai.AiService, tableService table.TableService,
	providerService provider.ProviderService, workflowService workflow.WorkflowService, datasetService dataset.DatasetService,
) *Backend {
	go func() {
		err := tableService.RemoveUnboundDatasets(context.Background())
		if err != nil {
			logger.Errorw("delete unbound datasets error", "error", err)
		}
		time.Sleep(10 * time.Minute)
	}()

	return &Backend{
		Config:          config,
		DB:              db,
		Logger:          logger,
		AIService:       aiService,
		TableService:    tableService,
		ProviderService: providerService,
		WorkflowService: workflowService,
		DatasetService:  datasetService,
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

	err = container.Provide(workflow.NewWorkflowService, dig.As(new((workflow.WorkflowService))))
	if err != nil {
		panic(err)
	}

	err = container.Provide(dataset.NewDatasetService, dig.As(new((dataset.DatasetService))))
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
