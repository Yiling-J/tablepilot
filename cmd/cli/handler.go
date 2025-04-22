package cli

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/table/util"
	"github.com/Yiling-J/tablepilot/utils/tableprinter"
	"github.com/gammazero/toposort"

	"github.com/spf13/cobra"
)

type Handler struct {
	backend    *services.Backend
	getPrinter func() tableprinter.TablePrinter
}

func NewHandler(backend *services.Backend) *Handler {
	return &Handler{
		backend:    backend,
		getPrinter: newPrinter,
	}
}

func (h *Handler) Create(cmd *cobra.Command, args []string) error {
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
		id, err := h.backend.TableService.Create(cmd.Context(), &req)
		if err != nil {
			return err
		}
		h.backend.Logger.Infow("table created", "id", id)
	}
	return nil
}

func (h *Handler) Update(cmd *cobra.Command, args []string) error {
	fileName := args[0]
	tb, err := cmd.Flags().GetString("table")
	if err != nil {
		return err
	}
	var req table.TableGenRequest
	f, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(f, &req)
	if err != nil {
		return err
	}
	id, err := h.backend.TableService.Update(cmd.Context(), tb, &req)
	if err != nil {
		return err
	}
	h.backend.Logger.Infow("table updated", "id", id)
	return nil
}

func (h *Handler) Show(cmd *cobra.Command, args []string) error {
	table := args[0]
	rows, err := h.backend.TableService.Rows(cmd.Context(), table)
	if err != nil {
		return err
	}
	indexer := util.NewColumnIndexer(rows.Columns)
	tp := h.getPrinter()
	tp.AddHeader(indexer.ColumnNames())
	for _, row := range rows.Rows {
		for _, cell := range row.Cells {
			tp.AddField(cellString(cell.Value))
		}
		tp.EndRow()
	}
	return tp.Render()
}

func (h *Handler) List(cmd *cobra.Command, args []string) error {
	resp, err := h.backend.TableService.ListTables(cmd.Context())
	if err != nil {
		return err
	}
	tp := h.getPrinter()
	tp.AddHeader([]string{"ID", "Name", "Description"})
	for _, table := range resp.Tables {
		tp.AddField(table.ID)
		tp.AddField(table.Name)
		tp.AddField(table.Description)
		tp.EndRow()
	}
	return tp.Render()
}

func (h *Handler) Describe(cmd *cobra.Command, args []string) error {
	detail, err := h.backend.TableService.GetTableDetail(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	tp := h.getPrinter()
	tp.AddHeader([]string{"ID", "Name", "Type", "Fill Mode", "Description"})
	for _, column := range detail.Columns {
		tp.AddField(column.ID)
		tp.AddField(column.Name)
		tp.AddField(column.Type)
		tp.AddField(column.FillMode)
		tp.AddField(column.Description)
		tp.EndRow()
	}
	return tp.Render()
}

func (h *Handler) Delete(cmd *cobra.Command, args []string) error {
	table := args[0]
	count, err := h.backend.TableService.Delete(cmd.Context(), table)
	if err != nil {
		return err
	}
	if count > 0 {
		h.backend.Logger.Info("table removed")
	} else {
		h.backend.Logger.Info("table not found")
	}
	return nil
}

func (h *Handler) Export(cmd *cobra.Command, args []string) error {
	table := args[0]
	rows, err := h.backend.TableService.Rows(cmd.Context(), table)
	if err != nil {
		return err
	}
	to, err := cmd.Flags().GetString("to")
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
	defer h.backend.Logger.Infow("file exported", "path", to)
	return csvwriter.WriteAll(data)
}

func (h *Handler) Generate(cmd *cobra.Command, args []string) error {
	batch, err := cmd.Flags().GetInt("batch")
	if err != nil {
		return err
	}
	count, err := cmd.Flags().GetInt("count")
	if err != nil {
		return err
	}
	saveTo, err := cmd.Flags().GetString("saveto")
	if err != nil {
		return err
	}
	if batch > count {
		batch = count
	}
	temperature, err := cmd.Flags().GetFloat64("temperature")
	if err != nil {
		return err
	}
	model, err := cmd.Flags().GetString("model")
	if err != nil {
		return err
	}
	generator, err := h.backend.TableService.Genetate(
		cmd.Context(), table.GenerateRowsRequest{
			Table:       args[0],
			SaveTo:      saveTo,
			Count:       count,
			Batch:       batch,
			Temperature: temperature,
			Model:       model,
		},
	)
	if err != nil {
		return err
	}
	indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
	tp := h.getPrinter()
	tp.AddHeader(indexer.ColumnNames())
	var csvWriter *csv.Writer
	if saveTo != "" {
		file, err := os.Create(saveTo)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		csvWriter = csv.NewWriter(file)
		columns := []string{}
		for _, col := range generator.Table().Edges.Columns {
			columns = append(columns, col.Name)
		}
		err = csvWriter.Write(columns)
		if err != nil {
			return err
		}
	}
	for {
		batch, err := generator.Next(cmd.Context())
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			break
		}
		for _, row := range batch {
			sr := []string{}
			v, err := indexer.RowMapToSlice(row)
			if err != nil {
				return err
			}
			for _, cell := range v {
				sv := cellString(cell.Value)
				sr = append(sr, sv)
				tp.AddField(sv)
			}
			tp.EndRow()
			if csvWriter != nil {
				err = csvWriter.Write(sr)
				if err != nil {
					return err
				}
			}
		}
		err = tp.Render()
		if err != nil {
			return err
		}
		if csvWriter != nil {
			csvWriter.Flush()
		}
	}
	h.backend.Logger.Infow("generated done")
	return nil
}

func (h *Handler) Import(cmd *cobra.Command, args []string) error {
	table, err := cmd.Flags().GetString("table")
	if err != nil {
		return err
	}
	tableFile := args[0]
	reader, err := os.Open(tableFile)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()
	if table == "" {
		table = filepath.Base(tableFile)
		table = strings.TrimSuffix(table, filepath.Ext(table))
	}
	id, err := h.backend.TableService.Import(cmd.Context(), table, reader)
	if err != nil {
		return err
	}
	h.backend.Logger.Infow("table imported", "id", id)
	return nil
}

func (h *Handler) Truncate(cmd *cobra.Command, args []string) error {
	table := args[0]
	removed, err := h.backend.TableService.Truncate(cmd.Context(), table)
	if err != nil {
		return err
	}
	h.backend.Logger.Infow("table truncated", "removed", removed)
	return nil
}

func (h *Handler) Autofill(cmd *cobra.Command, args []string) error {
	batch, err := cmd.Flags().GetInt("batch")
	if err != nil {
		return err
	}
	count, err := cmd.Flags().GetInt("count")
	if err != nil {
		return err
	}
	offset, err := cmd.Flags().GetInt("offset")
	if err != nil {
		return err
	}
	saveTo, err := cmd.Flags().GetString("saveto")
	if err != nil {
		return err
	}
	if batch > count {
		batch = count
	}
	temperature, err := cmd.Flags().GetFloat64("temperature")
	if err != nil {
		return err
	}
	model, err := cmd.Flags().GetString("model")
	if err != nil {
		return err
	}
	columns, err := cmd.Flags().GetStringArray("columns")
	if err != nil {
		return err
	}
	contextColumns, err := cmd.Flags().GetStringArray("context_columns")
	if err != nil {
		return err
	}

	generator, err := h.backend.TableService.Genetate(
		cmd.Context(), table.GenerateRowsRequest{
			Table:       args[0],
			SaveTo:      saveTo,
			Count:       count,
			Batch:       batch,
			Temperature: temperature,
			Model:       model,
			Autofill: table.AutofillRequest{
				Enable:         true,
				Offset:         offset,
				Columns:        columns,
				ContextColumns: contextColumns,
			},
		},
	)
	if err != nil {
		return err
	}

	indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
	tp := h.getPrinter()
	tp.AddHeader(indexer.ColumnNames())
	var csvWriter *csv.Writer
	if saveTo != "" {
		file, err := os.Create(saveTo)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		csvWriter = csv.NewWriter(file)
		columns := []string{}
		for _, col := range generator.Table().Edges.Columns {
			columns = append(columns, col.Name)
		}
		err = csvWriter.Write(columns)
		if err != nil {
			return err
		}
	}
	for {
		batch, err := generator.Next(cmd.Context())
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			break
		}
		for _, row := range batch {
			sr := []string{}
			v, err := indexer.RowMapToSlice(row)
			if err != nil {
				return err
			}
			for _, cell := range v {
				sv := cellString(cell.Value)
				sr = append(sr, sv)
				tp.AddField(sv)
			}
			tp.EndRow()
			if csvWriter != nil {
				err = csvWriter.Write(sr)
				if err != nil {
					return err
				}
			}
		}
		err = tp.Render()
		if err != nil {
			return err
		}
		if csvWriter != nil {
			csvWriter.Flush()
		}
	}
	h.backend.Logger.Infow("autofill done")
	return nil
}

func printTables(tables []table.BuilderTable) {
	for _, t := range tables {
		fmt.Printf("- %s: %s\n", t.Name, t.Description)
		if len(t.Depends) > 0 {
			fmt.Printf("depends on: %s\n", strings.Join(t.Depends, ", "))
		}
		fmt.Println("----------")
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	b, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func topoSortTables(tables []table.BuilderTable) ([]table.BuilderTable, error) {
	var edges []toposort.Edge
	tm := map[string]table.BuilderTable{}
	for _, table := range tables {
		tm[table.Name] = table
		if len(table.Depends) == 0 {
			edges = append(edges, toposort.Edge{nil, table.Name})
			continue
		}
		for _, dep := range table.Depends {
			edges = append(edges, toposort.Edge{dep, table.Name})
		}
	}

	sortedNames, err := toposort.Toposort(edges)
	if err != nil {
		return nil, err
	}
	sorted := []table.BuilderTable{}
	for _, name := range sortedNames {
		sorted = append(sorted, tm[name.(string)])
	}
	return sorted, nil
}

func (h *Handler) Builder(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This command will create a set of tables based on your requirement.")
	fmt.Println("Please describe what you want to build, press Enter to finish")
	fmt.Print("> ")

	userPrompt, err := readLine(reader)
	if err != nil {
		return fmt.Errorf("failed to read prompt: %w", err)
	}
	userPrompt = strings.TrimSpace(userPrompt)

	// Generate tables
	tables, err := h.backend.TableService.GenerateBuilderTables(cmd.Context(), userPrompt)
	if err != nil {
		return fmt.Errorf("failed to generate tables: %w", err)
	}

	for {
		fmt.Println("\nGenerated Tables:")
		printTables(tables)

		fmt.Print("\nIs this table list acceptable? Enter any improvements you'd like, or leave blank if it's good enough. Press Enter when you're done: ")
		input, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		input = strings.TrimSpace(input)

		if input == "" {
			break
		}

		tables, err = h.backend.TableService.PolishBuilderTables(cmd.Context(), tables, input)
		if err != nil {
			return fmt.Errorf("failed to polish tables: %w", err)
		}
	}
	tables, err = topoSortTables(tables)
	if err != nil {
		return fmt.Errorf("failed to topo sort tables: %w", err)
	}

	fmt.Println("Final table list accepted. Start creating tables...")
	created := map[string]*table.TableInfo{}
	for _, tb := range tables {
		fmt.Printf("Generate table schema for %s\n", tb.Name)
		exists := []*table.TableInfo{}
		for _, c := range created {
			exists = append(exists, c)
		}
		info, err := h.backend.TableService.BuildTable(cmd.Context(), tb.Name, tb.Description, tb.Depends, exists)
		if err != nil {
			return err
		}
		fmt.Printf("Table Name: %s\n", tb.Name)
		fmt.Printf("Table Description: %s\n", tb.Description)
		fmt.Println("")
		for {
			if len(info.Sources) > 0 {
				fmt.Println("Here is the sources of the table in JSON format:")
				for _, source := range info.Sources {
					b, err := json.Marshal(source)
					if err != nil {
						return err
					}
					fmt.Println(string(b))
				}
			}
			fmt.Println("Here is the columns of the table in JSON format:")
			for _, column := range info.Columns {
				b, err := json.Marshal(column)
				if err != nil {
					return err
				}
				fmt.Println(string(b))
			}
			fmt.Print("\nIs this table acceptable? Enter any improvements you'd like, or leave blank if it's good enough. Press Enter when you're done: ")
			input, err := readLine(reader)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			input = strings.TrimSpace(input)
			if input == "" {
				break
			}

			info, err = h.backend.TableService.PolishBuilderTable(cmd.Context(), info, input)
			if err != nil {
				return fmt.Errorf("failed to polish tables: %w", err)
			}
		}
		tid, err := h.backend.TableService.Create(cmd.Context(), info)
		if err != nil {
			return err
		}
		fmt.Println("table created")
		detail, err := h.backend.TableService.GetTableDetail(cmd.Context(), tid)
		if err != nil {
			return err
		}
		created[info.Name] = detail
	}
	return nil
}
