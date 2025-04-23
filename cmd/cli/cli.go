package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Yiling-J/tablepilot/api"
	_ "github.com/Yiling-J/tablepilot/ent/runtime"
	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/utils/tableprinter"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	_ "modernc.org/sqlite"
)

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

type CLI struct {
	Backend *services.Backend
	Handler *Handler
}

func BuildCLI(root *cobra.Command) *CLI {
	var configPath string
	cli := &CLI{}

	cmd := root
	cmd.PersistentFlags().StringVarP(&configPath, "config", "", "config.toml", "path to the config file")

	var verbose bool
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output (default: false)")

	var handler *Handler
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		backend := services.CreateBackend(cmd, verbose)
		cli.Backend = backend
		handler = NewHandler(backend)
		cli.Handler = handler
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

	updateCommand := &cobra.Command{
		Use:   "update <file>...",
		Short: "Update an existing table using a schema JSON file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Update(cmd, args)
		},
	}
	updateCommand.Flags().StringP("table", "t", "", "table ID or name to update; defaults to the name field in the JSON if not specified")
	cmd.AddCommand(updateCommand)

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.List(cmd, args)
		},
	})

	describeCommand := &cobra.Command{
		Use:   "describe <table id or name>",
		Short: "Show details about the columns in a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Describe(cmd, args)
		},
	}
	describeCommand.Flags().StringP("output", "o", "table", "specifies the output format. Possible values are 'table' or 'json'. Defaults to 'table'")
	cmd.AddCommand(describeCommand)

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

	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start tablepilot server",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := api.NewHttpServer(handler.backend, verbose)
			server.RegisterRoutes()

			err := server.Engine.Run(handler.backend.Config.Server.Address)
			if err != nil {
				return err
			}
			return nil
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
	generate.Flags().StringP(
		"model", "m", "",
		"specify the model used to generate rows. If not provided, the default model will be used",
	)

	cmd.AddCommand(generate)

	autofill := &cobra.Command{
		Use:   "autofill <table id or name>",
		Short: "Autofill missing columns specified table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Autofill(cmd, args)
		},
	}
	autofill.Flags().IntP("count", "c", 0, "total number of rows to autofill, default all existing rows")
	autofill.Flags().IntP("batch", "b", 10, "number of rows to autofill in a batch")
	autofill.Flags().IntP("offset", "o", 0, "start offset")
	autofill.Flags().StringP(
		"saveto", "s", "",
		"specify a file to save output, instead of storing in the database",
	)
	autofill.Flags().Float64P("temperature", "t", 0.6, "The sampling temperature. Higher values will make the output more random.")
	autofill.Flags().StringP(
		"model", "m", "",
		"specify the model used to generate rows. If not provided, the default model will be used",
	)
	autofill.Flags().StringArray("columns", []string{}, "columns to be autofilled, existing value wil be ignore and force regenerate")
	err = autofill.MarkFlagRequired("columns")
	if err != nil {
		panic(err)
	}
	autofill.Flags().StringArray("context_columns", []string{}, "columns that should be put in prompt as context, default to all other columns")

	cmd.AddCommand(autofill)

	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import csv file as table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Import(cmd, args)
		},
	}
	importCmd.Flags().StringP("table", "t", "", "imports into an existing table or creates a new one if missing. Defaults to file name if not set")
	cmd.AddCommand(importCmd)

	builder := &cobra.Command{
		Use:   "builder",
		Short: "Start tablepilot builder, create tables interactively using natural language",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Builder(cmd, args)
		},
	}
	builder.Flags().Float64P("temperature", "t", 0.3, "The sampling temperature. Higher values will make the output more random.")
	builder.Flags().StringP(
		"model", "m", "",
		"specify the model used by builder. If not provided, the default model will be used",
	)
	cmd.AddCommand(builder)
	return cli
}
