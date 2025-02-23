package cli

import (
	"encoding/csv"
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
	"tablepilot/services/table/util"
	"tablepilot/utils/tableprinter"
	"time"

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

func createBackend(cmd *cobra.Command) *Backend {
	container := dig.New()

	err := container.Provide(func() (*config.Config, error) {
		return config.NewConfig(cmd.Flag("config").Value.String())
	})
	if err != nil {
		panic(err)
	}

	err = container.Provide(func(config *config.Config) (*zap.SugaredLogger, error) {
		if config.Debug {
			cfg := zap.NewDevelopmentConfig()
			l, err := cfg.Build()
			if err != nil {
				return nil, err
			}
			return l.Sugar(), nil
		}
		l, err := zap.NewProduction()
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

	var backend *Backend
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		backend = createBackend(cmd)
		return nil
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "create <file>...",
		Short: "Create tables from schema JSON files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, fileName := range args {
				var req table.TableGenRequest
				f, err := os.ReadFile(fileName)
				if err != nil {
					return err
				}
				err = json.Unmarshal(f, &req)
				if err != nil {
					return err
				}
				id, err := backend.tableService.CreateTable(cmd.Context(), &req)
				if err != nil {
					return err
				}
				backend.Logger.Infow("table created", "id", id)
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := backend.tableService.ListTables(cmd.Context())
			if err != nil {
				return err
			}
			width, _, err := term.GetSize(0)
			if err != nil {
				return err
			}
			tp := tableprinter.New(os.Stdout, true, width, 25)
			tp.AddHeader([]string{"ID", "Name"})
			for _, table := range resp.Tables {
				tp.AddField(table.ID)
				tp.AddField(table.Name)
				tp.EndRow()
			}
			return tp.Render()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "describe <table id or name>",
		Short: "Show details about the columns in a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "show <table id or name>",
		Short: "Display the rows of a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			table := args[0]
			rows, err := backend.tableService.Rows(cmd.Context(), table)
			if err != nil {
				return err
			}
			indexer := util.NewColumnIndexer(rows.Columns)
			tp := newPrinter()
			tp.AddHeader(indexer.ColumnNames())
			for _, row := range rows.Rows {
				for _, cell := range row.Cells {
					tp.AddField(cellString(cell.Value))
				}
				tp.EndRow()
			}
			return tp.Render()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <table id or name>",
		Short: "Delete a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			table := args[0]
			count, err := backend.tableService.Delete(cmd.Context(), table)
			if err != nil {
				return err
			}
			if count > 0 {
				backend.Logger.Info("table removed")
			} else {
				backend.Logger.Info("table not found")
			}
			return nil
		},
	})

	var to string
	exportCommand := &cobra.Command{
		Use:   "export <table id or name>",
		Short: "xport the table as a CSV file",
		RunE: func(cmd *cobra.Command, args []string) error {
			table := args[0]
			rows, err := backend.tableService.Rows(cmd.Context(), table)
			if err != nil {
				return err
			}
			if to == "" {
				to = fmt.Sprintf("%s_%d.csv", table, time.Now().Unix())
			}
			csvFile, err := os.Create(to)
			if err != nil {
				return err
			}
			defer func() { _ = csvFile.Close() }()
			csvwriter := csv.NewWriter(csvFile)
			columns := []string{}
			for _, col := range rows.Columns {
				columns = append(columns, col.Name)
			}
			err = csvwriter.Write(columns)
			if err != nil {
				return err
			}
			data := [][]string{}
			for _, row := range rows.Rows {
				r := []string{}
				for _, v := range row.Cells {
					r = append(r, cellString(v.Value))
				}
				data = append(data, r)
			}
			defer backend.Logger.Infow("file exported", "path", to)
			return csvwriter.WriteAll(data)
		},
	}
	exportCommand.Flags().StringVarP(&to, "to", "t", "", "exported file path")
	cmd.AddCommand(exportCommand)

	cmd.AddCommand(&cobra.Command{
		Use:   "truncate <table id or name>",
		Short: "Remove all data from a specified table",
		RunE: func(cmd *cobra.Command, args []string) error {
			table := args[0]
			removed, err := backend.tableService.Truncate(cmd.Context(), table)
			if err != nil {
				return err
			}
			backend.Logger.Infow("table truncated", "removed", removed)
			return nil
		},
	})

	var count int
	var batch int
	generate := &cobra.Command{
		Use:   "generate <table id or name>",
		Short: "Generate data for a specified table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if batch > count {
				batch = count
			}
			table := args[0]
			generator, err := backend.tableService.Genetate(cmd.Context(), table, count, batch)
			if err != nil {
				return err
			}
			indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
			tp := newPrinter()
			tp.AddHeader(indexer.ColumnNames())
			for {
				batch, err := generator.Next(cmd.Context())
				if err != nil {
					return err
				}
				if len(batch) == 0 {
					break
				}
				for _, row := range batch {
					v, err := indexer.RowMapToSlice(row)
					if err != nil {
						return err
					}
					for _, cell := range v {
						tp.AddField(cellString(cell.Value))
					}
					tp.EndRow()
				}
				err = tp.Render()
				if err != nil {
					return err
				}
			}
			backend.Logger.Infow("generated done")
			return nil
		},
	}
	generate.Flags().IntVarP(&count, "count", "c", 0, "")
	err := generate.MarkFlagRequired("count")
	if err != nil {
		panic(err)
	}
	generate.Flags().IntVarP(&batch, "batch", "b", 10, "")

	cmd.AddCommand(generate)

	var importName string
	importCmd := &cobra.Command{
		Use:   "import <table id or name>",
		Short: "Import csv file as table",
		RunE: func(cmd *cobra.Command, args []string) error {
			tableFile := args[0]
			reader, err := os.Open(tableFile)
			if err != nil {
				return err
			}
			defer func() { _ = reader.Close() }()
			if importName == "" {
				importName = tableFile
			}
			id, err := backend.tableService.Import(cmd.Context(), importName, reader)
			if err != nil {
				return err
			}
			backend.Logger.Infow("table imported", "id", id)
			return nil
		},
	}
	importCmd.Flags().StringVarP(&importName, "name", "n", "", "")
	cmd.AddCommand(importCmd)
}
