package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"tablepilot/config"
	"tablepilot/ent"
	_ "tablepilot/ent/runtime"
	"tablepilot/infra/db"
	"tablepilot/services/ai"
	"tablepilot/services/ai/client"
	"tablepilot/services/table"
	"tablepilot/utils/tableprinter"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"golang.org/x/term"
	_ "modernc.org/sqlite"
)

type Backend struct {
	config       *config.Config
	db           *ent.Client
	Logger       *zap.SugaredLogger
	aiService    ai.AiService
	tableService table.TableService
}

func NewBackend(
	config *config.Config, db *ent.Client,
	logger *zap.SugaredLogger, aiService ai.AiService, tableService table.TableService,
) *Backend {
	return &Backend{
		config:       config,
		db:           db,
		Logger:       logger,
		aiService:    aiService,
		tableService: tableService,
	}
}

func createBackend(cmd *cobra.Command, verbose bool) *Backend {
	container := dig.New()

	err := container.Provide(func() (*config.Config, error) {
		return config.NewConfig(cmd.Flag("config").Value.String())
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
		cfg.Encoding = "console"
		l, err := cfg.Build()
		if err != nil {
			return nil, err
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

	err = container.Provide(client.NewClients)
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

func newPrinter() tableprinter.TablePrinter {
	maxWidth, _, _ := term.GetSize(0)
	return tableprinter.New(os.Stdout, term.IsTerminal(int(os.Stdout.Fd())), maxWidth, 25)
}

func cellString(v any) string {
	vs, err := cast.ToStringE(v)
	if err != nil {
		vb, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%+v", v)
		}
		return string(vb)
	}
	return vs
}

func BuildCLI(root *cobra.Command) {
	var configPath string

	cmd := root
	cmd.PersistentFlags().StringVarP(&configPath, "config", "", "config.toml", "path to the config file")

	var verbose bool
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (default: false)")

	var handler *Handler
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		backend := createBackend(cmd, verbose)
		handler = NewHandler(backend)
		return nil
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "create <file>...",
		Short: "Create tables from schema JSON files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Create(cmd, args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.List(cmd, args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "describe <table id or name>",
		Short: "Show details about the columns in a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Describe(cmd, args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show <table id or name>",
		Short: "Display the rows of a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Show(cmd, args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <table id or name>",
		Short: "Delete a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Delete(cmd, args)
		},
	})

	exportCommand := &cobra.Command{
		Use:   "export <table id or name>",
		Short: "export the table as a CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Export(cmd, args)
		},
	}
	exportCommand.Flags().StringP("to", "t", "", "exported file path")
	cmd.AddCommand(exportCommand)

	cmd.AddCommand(&cobra.Command{
		Use:   "truncate <table id or name>",
		Short: "Remove all data from a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Truncate(cmd, args)
		},
	})

	generate := &cobra.Command{
		Use:   "generate <table id or name>",
		Short: "Generate data for a specified table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Generate(cmd, args)
		},
	}
	generate.Flags().IntP("count", "c", 0, "total number of rows to generate")
	err := generate.MarkFlagRequired("count")
	if err != nil {
		panic(err)
	}
	generate.Flags().IntP("batch", "b", 10, "number of rows to generate in a batch")
	generate.Flags().StringP(
		"saveto", "s", "",
		"specify a file to save output, instead of storing in the database",
	)
	generate.Flags().Float64P("temperature", "t", 0.6, "The sampling temperature. Higher values will make the output more random.")

	cmd.AddCommand(generate)

	importCmd := &cobra.Command{
		Use:   "import <table id or name>",
		Short: "Import csv file as table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Import(cmd, args)
		},
	}
	importCmd.Flags().StringP("name", "n", "", "")
	cmd.AddCommand(importCmd)
}
