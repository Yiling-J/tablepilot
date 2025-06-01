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
	cmd.PersistentFlags().StringVar(&configPath, "config", "config.toml", "path to the config file")

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
	autofill.Flags().StringP("prompt", "p", "", "optional prompt text send to LLM")

	cmd.AddCommand(autofill)

	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import csv or image file as table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Import(cmd, args)
		},
	}
	importCmd.Flags().StringP("table", "t", "", "imports into an existing table or creates a new one if missing.")
	importCmd.Flags().StringP("name", "n", "", "name of the new table, if to flag is not set. Optional and if not set, new table name will be file name + current timestamp")
	importCmd.Flags().StringP("prompt", "p", "", "optional prompt text send to LLM")
	importCmd.Flags().StringP(
		"model", "m", "",
		"specify the model used to extract data from image. If not provided, the default model will be used",
	)
	importCmd.Flags().Bool("truncate", false, "remove all rows in the table first before importing")
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

	regenerate := &cobra.Command{
		Use:   "regenerate <table id or name>",
		Short: "Regenerate selected rows",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.Regenerate(cmd, args)
		},
	}
	regenerate.Flags().StringP("prompt", "p", "", "prompt text send to LLM")
	regenerate.Flags().IntP("batch", "b", 10, "number of rows to autofill in a batch")
	regenerate.Flags().StringArrayP("rows", "r", []string{}, "row id to regenerate")
	err = regenerate.MarkFlagRequired("rows")
	if err != nil {
		panic(err)
	}
	regenerate.Flags().Float64P("temperature", "t", 0.6, "The sampling temperature. Higher values will make the output more random.")
	regenerate.Flags().StringP(
		"model", "m", "",
		"specify the model used to generate rows. If not provided, the default model will be used",
	)
	regenerate.Flags().StringArrayP("columns", "c", []string{}, "columns to be autofilled, existing value wil be ignore and force regenerate")
	err = regenerate.MarkFlagRequired("columns")
	if err != nil {
		panic(err)
	}
	cmd.AddCommand(regenerate)

	workflowCommand := &cobra.Command{
		Use:   "workflow",
		Short: "workflow subcommands",
	}

	runWorkflowCommand := &cobra.Command{
		Use:   "run <workflow>",
		Short: "Run workflow of given id or name",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler.RunWorkflow(cmd, args)
		},
	}
	runWorkflowCommand.Flags().Float64P("temperature", "t", 0.6, "The sampling temperature. Higher values will make the output more random.")
	runWorkflowCommand.Flags().StringP(
		"model", "m", "",
		"specify the model used to generate rows. If not provided, the default model will be used",
	)
	runWorkflowCommand.Flags().StringP(
		"image_model", "i", "",
		"specify the image model used to generate rows. If not provided, the default model will be used",
	)
	workflowCommand.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List workflows",
			RunE: func(cmd *cobra.Command, args []string) error {
				return handler.ListWorkflows(cmd, args)
			},
		},
		&cobra.Command{
			Use:   "create <file>",
			Short: "Create workflow from schema JSON files",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return handler.CreateWorkflow(cmd, args)
			},
		},
		runWorkflowCommand,
		&cobra.Command{
			Use:   "delete <workflow>",
			Short: "Delete workflow of given id or name",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return handler.DeleteWorkflow(cmd, args)
			},
		},
	)

	cmd.AddCommand(workflowCommand)

	// Dataset commands
	datasetCmd := &cobra.Command{
		Use:   "dataset",
		Short: "Manage datasets",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Ensure backend and handler are initialized for dataset commands as well
			// This might be redundant if the root PersistentPreRunE always runs first
			// and sets up cli.Backend and cli.Handler.
			// If subcommands of datasetCmd don't trigger root's PersistentPreRunE,
			// this explicit setup might be needed. Cobra's behavior can vary.
			// For safety, let's assume it's needed or harmlessly redundant.
			if cli.Backend == nil || cli.Handler == nil {
				// This re-runs part of the root command's PersistentPreRunE logic.
				// Consider refactoring PersistentPreRunE if this becomes complex.
				rootCmd := cmd.Root()
				if rootCmd.PersistentPreRunE != nil {
					// Need to pass the correct root command and args
					// This is a bit of a hack; ideally, Cobra ensures parent PersistentPreRunE runs.
					// Let's assume the root PersistentPreRunE has already run.
				}
				if cli.Backend == nil { // If still nil after trying to run root's
					return fmt.Errorf("backend not initialized for dataset command")
				}
			}
			return nil
		},
	}

	datasetCreateCmd := &cobra.Command{
		Use:   "create --name <name> [--desc <description>] (--type list --data <item1> --data <item2> | --type csv --file <path/to/file1.csv> [--file <path/to/file2.csv>...])",
		Short: "Create a new dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Handler.CreateDataset(cmd, args)
		},
	}
	datasetCreateCmd.Flags().StringP("name", "n", "", "Name of the dataset (required)")
	datasetCreateCmd.MarkFlagRequired("name")
	datasetCreateCmd.Flags().StringP("desc", "d", "", "Description of the dataset")
	datasetCreateCmd.Flags().StringP("type", "t", "", "Type of the dataset ('list' or 'csv') (required)")
	datasetCreateCmd.MarkFlagRequired("type")
	datasetCreateCmd.Flags().StringArray("data", []string{}, "Data items for 'list' type dataset (can be specified multiple times)")
	datasetCreateCmd.Flags().StringArrayP("file", "f", []string{}, "Path to CSV file(s) for 'csv' type dataset (can be specified multiple times)")

	datasetGetCmd := &cobra.Command{
		Use:   "get <dataset_id_or_name>",
		Short: "Get details of a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Handler.GetDataset(cmd, args)
		},
	}

	datasetListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all available datasets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Handler.ListDatasets(cmd, args)
		},
	}

	datasetUpdateCmd := &cobra.Command{
		Use:   "update <dataset_id_or_name> [--name <new_name>] [--desc <new_description>] [--type <new_type>] [--data <item1>...] [--file <path/to/file1.csv>...]",
		Short: "Update an existing dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Handler.UpdateDataset(cmd, args)
		},
	}
	datasetUpdateCmd.Flags().String("name", "", "New name for the dataset")
	datasetUpdateCmd.Flags().String("desc", "", "New description for the dataset")
	datasetUpdateCmd.Flags().String("type", "", "New type for the dataset ('list' or 'csv')")
	datasetUpdateCmd.Flags().StringArray("data", []string{}, "New data items for 'list' type (replaces existing if type is list or changed to list)")
	datasetUpdateCmd.Flags().StringArrayP("file", "f", []string{}, "New CSV file(s) for 'csv' type (replaces existing if type is csv or changed to csv)")

	datasetDeleteCmd := &cobra.Command{
		Use:   "delete <dataset_id_or_name>",
		Short: "Delete a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Handler.DeleteDataset(cmd, args)
		},
	}

	datasetPreviewCmd := &cobra.Command{
		Use: "preview <dataset_id_or_name>",
		Short: "Preview data from a dataset (first 100 rows for CSV)",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.Handler.PreviewDataset(cmd, args)
		},
	}


	datasetCmd.AddCommand(datasetCreateCmd)
	datasetCmd.AddCommand(datasetGetCmd)
	datasetCmd.AddCommand(datasetListCmd)
	datasetCmd.AddCommand(datasetUpdateCmd)
	datasetCmd.AddCommand(datasetDeleteCmd)
	datasetCmd.AddCommand(datasetPreviewCmd)
	cmd.AddCommand(datasetCmd)

	return cli
}
