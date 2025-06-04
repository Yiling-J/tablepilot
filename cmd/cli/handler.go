package cli

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	db_dataset "github.com/Yiling-J/tablepilot/ent/dataset"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/table/util"
	"github.com/Yiling-J/tablepilot/services/workflow"
	"github.com/Yiling-J/tablepilot/utils/tableprinter"
	"github.com/gammazero/toposort"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

type Handler struct {
	backend          *services.Backend
	getPrinter       func() tableprinter.TablePrinter
	promptUserSelect func(prompt string, options []string, defaultValue string) (string, error)
}

func NewHandler(backend *services.Backend) *Handler {
	return &Handler{
		backend:          backend,
		getPrinter:       newPrinter,
		promptUserSelect: SelectFromSlice,
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
	headers := []string{"[ID]"}
	headers = append(headers, indexer.ColumnNames()...)
	tp.AddHeader(headers)
	for _, row := range rows.Rows {
		tp.AddField(row.Nanoid)
		for _, cell := range row.Cells {
			tp.AddField(cellString(cell.Value))
		}
		tp.EndRow()
	}
	return tp.Render()
}

func (h *Handler) CreateDataset(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("error getting name flag: %w", err)
	}
	desc, err := cmd.Flags().GetString("desc")
	if err != nil {
		return fmt.Errorf("error getting desc flag: %w", err)
	}
	datasetType, err := cmd.Flags().GetString("type")
	if err != nil {
		return fmt.Errorf("error getting type flag: %w", err)
	}

	req := &dataset.CreateDatasetRequest{
		Name:        name,
		Description: desc,
		Type:        datasetType,
	}

	switch datasetType {
	case "list":
		filePaths, err := cmd.Flags().GetStringArray("file")
		if err != nil {
			return fmt.Errorf("error getting data flag for list type: %w", err)
		}
		options := []string{}
		for _, fp := range filePaths {
			o, err := ReadLines(fp)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fp, err)
			}
			options = append(options, o...)
		}
		req.Data = options
	case "csv":
		filePaths, err := cmd.Flags().GetStringArray("file")
		if err != nil {
			return fmt.Errorf("error getting file flag for csv type: %w", err)
		}
		if len(filePaths) == 0 {
			return fmt.Errorf("at least one --file path must be provided for type 'csv'")
		}
		var files []io.Reader
		for _, filePath := range filePaths {
			file, err := os.Open(filePath)
			if err != nil {
				// Close already opened files if any
				for _, f := range files {
					if c, ok := f.(io.Closer); ok {
						c.Close()
					}
				}
				return fmt.Errorf("failed to open file %s: %w", filePath, err)
			}
			files = append(files, file)
		}
		req.Files = files
		defer func() {
			for _, f := range files {
				if c, ok := f.(io.Closer); ok {
					c.Close()
				}
			}
		}()
	}

	nanoid, err := h.backend.DatasetService.Create(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	h.backend.Logger.Infow("Dataset created successfully", "id", nanoid, "name", name)
	fmt.Printf("Dataset created successfully:\nID: %s\nName: %s\n", nanoid, name)
	return nil
}

func (h *Handler) GetDataset(cmd *cobra.Command, args []string) error {
	datasetIDOrName := args[0]
	dsInfo, err := h.backend.DatasetService.Get(cmd.Context(), datasetIDOrName)
	if err != nil {
		return fmt.Errorf("failed to get dataset '%s': %w", datasetIDOrName, err)
	}

	tp := h.getPrinter()
	tp.AddHeader([]string{"Name", "Type", "Description"})
	tp.AddField(dsInfo.Name)
	tp.AddField(dsInfo.Type)
	tp.AddField(dsInfo.Description)
	tp.EndRow()
	return tp.Render()
}

func (h *Handler) ListDatasets(cmd *cobra.Command, args []string) error {
	datasets, err := h.backend.DatasetService.List(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to list datasets: %w", err)
	}

	if len(datasets) == 0 {
		fmt.Println("No datasets found.")
		return nil
	}

	tp := h.getPrinter()
	tp.AddHeader([]string{"Name", "Description", "Type", "Columns", "Values"})
	for _, ds := range datasets {
		tp.AddField(ds.Name)
		tp.AddField(ds.Description)
		tp.AddField(ds.Type)
		tp.AddField(fmt.Sprintf("%d", ds.ColumnCount))
		tp.AddField(fmt.Sprintf("%d", ds.ValueCount))
		tp.EndRow()
	}
	return tp.Render()
}

func (h *Handler) UpdateDataset(cmd *cobra.Command, args []string) error {
	ds, err := h.backend.DatasetService.Get(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	fields := []string{}
	for _, f := range []string{"name", "desc", "file"} {
		if cmd.Flags().Changed(f) {
			switch f {
			case "name":
				fields = append(fields, "name")
			case "desc":
				fields = append(fields, "description")
			case "file":
				fields = append(fields, "data")
				fields = append(fields, "files")
			}
			fields = append(fields, f)
		}
	}
	name, _ := cmd.Flags().GetString("name")
	desc, _ := cmd.Flags().GetString("desc")

	req := &dataset.UpdateDatasetRequest{
		CreateDatasetRequest: dataset.CreateDatasetRequest{
			Name:        name,
			Description: desc,
		},
		Fields: fields,
	}

	switch ds.Type {
	case "list":
		filePaths, err := cmd.Flags().GetStringArray("file")
		if err != nil {
			return fmt.Errorf("error getting data flag for list type: %w", err)
		}
		options := []string{}
		for _, fp := range filePaths {
			o, err := ReadLines(fp)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fp, err)
			}
			options = append(options, o...)
		}
		req.Data = options
	case "csv":
		filePaths, err := cmd.Flags().GetStringArray("file")
		if err != nil {
			return fmt.Errorf("error getting file flag for csv type: %w", err)
		}
		if len(filePaths) == 0 {
			return fmt.Errorf("at least one --file path must be provided for type 'csv'")
		}
		var files []io.Reader
		for _, filePath := range filePaths {
			file, err := os.Open(filePath)
			if err != nil {
				// Close already opened files if any
				for _, f := range files {
					if c, ok := f.(io.Closer); ok {
						c.Close()
					}
				}
				return fmt.Errorf("failed to open file %s: %w", filePath, err)
			}
			files = append(files, file)
		}
		req.Files = files
		defer func() {
			for _, f := range files {
				if c, ok := f.(io.Closer); ok {
					c.Close()
				}
			}
		}()
	}

	err = h.backend.DatasetService.Update(cmd.Context(), args[0], req)
	if err != nil {
		return fmt.Errorf("failed to update dataset '%s': %w", args[0], err)
	}

	h.backend.Logger.Infow("Dataset updated successfully", "id_or_name", args[0])
	return nil
}

func (h *Handler) DeleteDataset(cmd *cobra.Command, args []string) error {
	datasetID := args[0]
	err := h.backend.DatasetService.Delete(cmd.Context(), datasetID)
	if err != nil {
		return fmt.Errorf("failed to delete dataset '%s': %w", datasetID, err)
	}
	h.backend.Logger.Infow("Dataset deleted successfully", "id_or_name", datasetID)
	return nil
}

func (h *Handler) PreviewDataset(cmd *cobra.Command, args []string) error {
	datasetIDOrName := args[0]
	previewData, err := h.backend.DatasetService.Preview(cmd.Context(), datasetIDOrName)
	if err != nil {
		return fmt.Errorf("failed to preview dataset '%s': %w", datasetIDOrName, err)
	}

	switch previewData.Type {
	case db_dataset.TypeList:
		if len(previewData.Data) == 0 {
			fmt.Println("Dataset is empty.")
		} else {
			fmt.Println("Data items:")
			for i, item := range previewData.Data {
				fmt.Printf("%d: %s\n", i+1, item)
			}
		}
	case db_dataset.TypeCsv:
		if len(previewData.Rows) == 0 {
			fmt.Println("Dataset is empty or has no rows to preview.")
		} else {
			tp := h.getPrinter()
			var headers []string
			if len(previewData.Rows) > 0 {
				for key := range previewData.Rows[0] {
					headers = append(headers, key)
				}
			}

			if len(previewData.Rows) > 0 {
				tempHeadersMap := make(map[string]bool)
				for _, row := range previewData.Rows {
					for k := range row {
						tempHeadersMap[k] = true
					}
				}
				for k := range tempHeadersMap { // This will be random order
					headers = append(headers, k)
				}

				originalDsFull, dsErr := h.backend.DB.Dataset.Query().Where(
					db_dataset.Or(db_dataset.Name(datasetIDOrName), db_dataset.Nanoid(datasetIDOrName)),
				).Only(cmd.Context())

				if dsErr == nil && len(originalDsFull.Indexer.ColumnNames) > 0 {
					headers = originalDsFull.Indexer.ColumnNames
				} else if len(previewData.Rows) > 0 {
					inferredHeaders := []string{}
					for k := range previewData.Rows[0] {
						inferredHeaders = append(inferredHeaders, k)
					}
					headers = inferredHeaders
				}
			}

			tp.AddHeader(headers)
			for _, rowMap := range previewData.Rows {
				for _, hKey := range headers {
					tp.AddField(cellString(rowMap[hKey]))
				}
				tp.EndRow()
			}
			err = tp.Render()
			if err != nil {
				return fmt.Errorf("failed to render preview table: %w", err)
			}
			if len(previewData.Rows) >= 100 {
				fmt.Println("(Preview limited to first 100 rows)")
			}
		}
	default:
		fmt.Printf("Unknown dataset type '%s' for preview.\n", previewData.Type)
	}

	return nil
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
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	if output == "json" {
		schema, err := h.backend.TableService.GetTableSchema(cmd.Context(), args[0])
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
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
	headers := []string{}
	if saveTo == "" {
		headers = append(headers, "[ID]")
	}
	headers = append(headers, indexer.ColumnNames()...)
	tp.AddHeader(headers)
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
			if saveTo == "" {
				tp.AddField(cast.ToString(row["__id__"].Value))
			}
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
	tb, err := cmd.Flags().GetString("table")
	if err != nil {
		return err
	}
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	truncate, err := cmd.Flags().GetBool("truncate")
	if err != nil {
		return err
	}
	tableFile := args[0]
	reader, err := os.Open(tableFile)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()
	fileName := filepath.Base(tableFile)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))

	_, _, err = image.DecodeConfig(reader)

	_, seekErr := reader.Seek(0, io.SeekStart)
	if seekErr != nil {
		return fmt.Errorf("failed to seek file %s to beginning: %w", tableFile, seekErr)
	}

	var id string
	var importErr error

	if err == nil {
		h.backend.Logger.Debugw("file detected as image, using ImportImage", "file", tableFile)
		model, err := cmd.Flags().GetString("model")
		if err != nil {
			return err
		}
		prompt, err := cmd.Flags().GetString("prompt")
		if err != nil {
			return err
		}
		d, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("failed to read image %s: %w", tableFile, err)
		}
		id, importErr = h.backend.TableService.ImportImage(cmd.Context(), table.ImportRequest{
			Data:     d,
			Model:    model,
			Prompt:   prompt,
			Table:    tb,
			Filename: fileName,
			Truncate: truncate,
			Name:     name,
		})
		if importErr != nil {
			return fmt.Errorf("failed to import image %s: %w", tableFile, importErr)
		}
		h.backend.Logger.Infow("image imported successfully", "id", id, "sourceFile", tableFile)
	} else {
		h.backend.Logger.Debugw("file not detected as image or error decoding image config, using default Import",
			"file", tableFile, "decode_error", err.Error())
		id, importErr = h.backend.TableService.Import(cmd.Context(), table.ImportRequest{
			Reader:   reader,
			Table:    tb,
			Filename: fileName,
			Truncate: truncate,
			Name:     name,
		})
		if importErr != nil {
			return fmt.Errorf("failed to import file %s as table %s: %w", tableFile, tb, importErr)
		}
		h.backend.Logger.Infow("table imported successfully", "id", id, "table", tb, "sourceFile", tableFile)
	}
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
	prompt, err := cmd.Flags().GetString("prompt")
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
				Prompt:         prompt,
			},
		},
	)
	if err != nil {
		return err
	}

	indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
	tp := h.getPrinter()
	headers := []string{}
	if saveTo == "" {
		headers = append(headers, "[ID]")
	}
	headers = append(headers, indexer.ColumnNames()...)
	tp.AddHeader(headers)
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
			if saveTo == "" {
				tp.AddField(cast.ToString(row["__id__"].Value))
			}
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

var tableNameRegex = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")

func (h *Handler) Builder(cmd *cobra.Command, args []string) error {
	temperature, err := cmd.Flags().GetFloat64("temperature")
	if err != nil {
		return err
	}
	model, err := cmd.Flags().GetString("model")
	if err != nil {
		return err
	}

	reader := bufio.NewReader(cmd.InOrStdin())

	fmt.Println("This command will create a set of tables based on your requirement.")
	fmt.Println("Please describe what you want to build, press Enter to finish")
	fmt.Print("> ")

	userPrompt, err := readLine(reader)
	if err != nil {
		return fmt.Errorf("failed to read prompt: %w", err)
	}
	userPrompt = strings.TrimSpace(userPrompt)

	tables, err := h.backend.TableService.GenerateBuilderTables(cmd.Context(), userPrompt, table.ModelParams{
		Model:       model,
		Temperature: temperature,
	})
	if err != nil {
		return fmt.Errorf("failed to generate tables: %w", err)
	}

	for {
		fmt.Println("\nGenerated Tables:")
		printTables(tables)

		fmt.Print("\nIs this table list acceptable? Type any suggestions for improvements, or leave blank if it's good enough. Press Enter when you're done: ")
		input, err := readLine(reader)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		input = strings.TrimSpace(input)

		if input == "" {
			break
		}

		tables, err = h.backend.TableService.PolishBuilderTables(cmd.Context(), tables, input, table.ModelParams{
			Model:       model,
			Temperature: temperature,
		})
		if err != nil {
			return fmt.Errorf("failed to polish tables: %w", err)
		}
	}
	errors := []string{}
	for _, tb := range tables {
		if !tableNameRegex.MatchString(tb.Name) {
			errors = append(errors, fmt.Sprintf("[%s]: Table name must start with a letter and contain only letters, numbers, or underscores.", tb.Name))
		}
		exist, err := h.backend.DB.TableMeta.Query().Where(tablemeta.Name(tb.Name)).Exist(cmd.Context())
		if err != nil {
			return err
		}
		if exist {
			errors = append(errors, fmt.Sprintf("[%s]: This table name already exist, please change.", tb.Name))
		}
	}
	if len(errors) > 0 {
		fmt.Printf("\nAuto fixing these errors:\n%s", strings.Join(errors, "\n"))
		msg := fmt.Sprintf(`Please fix these errors:\n %s`, strings.Join(errors, `\n`))
		tables, err = h.backend.TableService.PolishBuilderTables(cmd.Context(), tables, msg, table.ModelParams{
			Model:       model,
			Temperature: temperature,
		})
		if err != nil {
			return fmt.Errorf("failed to polish tables: %w", err)
		}
		fmt.Println("\nDone! Here is the new table list:")
		printTables(tables)
		fmt.Println("")
	}
	tables, err = topoSortTables(tables)
	if err != nil {
		return fmt.Errorf("failed to topo sort tables: %w", err)
	}

	fmt.Println("Final table list accepted. Start creating tables...")
	created := map[string]*table.TableInfo{}
	for _, tb := range tables {
		fmt.Println("Start generating table schema")
		exists := []*table.TableInfo{}
		for _, c := range created {
			exists = append(exists, c)
		}
		info, err := h.backend.TableService.BuildTable(cmd.Context(), tb.Name, tb.Description, tb.Depends, exists, table.ModelParams{
			Model:       model,
			Temperature: temperature,
		})
		if err != nil {
			return err
		}
		err = h.backend.TableService.Validate(cmd.Context(), info)
		if err != nil {
			fmt.Println("Validate failed, try auto fixing table...")

			info, err = h.backend.TableService.PolishBuilderTable(
				cmd.Context(), info, fmt.Sprintf("Please fix this error: %s", err.Error()), exists,
				table.ModelParams{
					Model:       model,
					Temperature: temperature,
				})
			if err != nil {
				return fmt.Errorf("failed to fix table: %w", err)
			}
		}
		fmt.Println("")
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
			fmt.Print("\nIs this table acceptable? Type any suggestions for improvements, or leave blank if it's good enough. Press Enter when you're done: ")
			input, err := readLine(reader)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			input = strings.TrimSpace(input)
			if input == "" {
				break
			}

			info, err = h.backend.TableService.PolishBuilderTable(cmd.Context(), info, input, exists, table.ModelParams{
				Model:       model,
				Temperature: temperature,
			})
			if err != nil {
				return fmt.Errorf("failed to polish tables: %w", err)
			}
			for {
				err = h.backend.TableService.Validate(cmd.Context(), info)
				if err != nil {
					fmt.Println("Validate failed, try auto fixing table...")

					info, err = h.backend.TableService.PolishBuilderTable(
						cmd.Context(), info, fmt.Sprintf("Please fix this error: %s", err.Error()), exists,
						table.ModelParams{
							Model:       model,
							Temperature: temperature,
						})
					if err != nil {
						return fmt.Errorf("failed to fix table: %w", err)
					}
				} else {
					break
				}
			}
		}
		tid, err := h.backend.TableService.Create(cmd.Context(), info)
		if err != nil {
			return err
		}
		detail, err := h.backend.TableService.GetTableDetail(cmd.Context(), tid)
		if err != nil {
			return err
		}
		created[info.Name] = detail
	}
	fmt.Println("")
	fmt.Println("Table creation is complete. You can now generate content for the tables using either the CLI or WebUI. The recommended table generation order is as follows:")
	tableNames := []string{}
	for _, tb := range tables {
		tableNames = append(tableNames, tb.Name)
	}
	fmt.Println(strings.Join(tableNames, " -> "))
	return nil
}

func (h *Handler) Regenerate(cmd *cobra.Command, args []string) error {
	batch, err := cmd.Flags().GetInt("batch")
	if err != nil {
		return err
	}
	rows, err := cmd.Flags().GetStringArray("rows")
	if err != nil {
		return err
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
	prompt, err := cmd.Flags().GetString("prompt")
	if err != nil {
		return err
	}

	generator, err := h.backend.TableService.Genetate(
		cmd.Context(), table.GenerateRowsRequest{
			Table:       args[0],
			Count:       len(rows),
			Batch:       batch,
			Temperature: temperature,
			Model:       model,
			Autofill: table.AutofillRequest{
				Enable:  true,
				Columns: columns,
				Rows:    rows,
				Prompt:  prompt,
			},
		},
	)
	if err != nil {
		return err
	}

	indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
	tp := h.getPrinter()
	headers := []string{"[ID]"}
	headers = append(headers, indexer.ColumnNames()...)
	tp.AddHeader(headers)
	for {
		batch, err := generator.Next(cmd.Context())
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			break
		}
		for _, row := range batch {
			tp.AddField(cast.ToString(row["__id__"].Value))
			v, err := indexer.RowMapToSlice(row)
			if err != nil {
				return err
			}
			for _, cell := range v {
				sv := cellString(cell.Value)
				tp.AddField(sv)
			}
			tp.EndRow()
		}
		err = tp.Render()
		if err != nil {
			return err
		}
	}
	h.backend.Logger.Infow("regenerate done")
	return nil
}

func (h *Handler) RunWorkflow(cmd *cobra.Command, args []string) error {
	temperature, err := cmd.Flags().GetFloat64("temperature")
	if err != nil {
		return err
	}
	model, err := cmd.Flags().GetString("model")
	if err != nil {
		return err
	}
	imageModel, err := cmd.Flags().GetString("image_model")
	if err != nil {
		return err
	}
	// collect variables interactively
	wf, err := h.backend.WorkflowService.Get(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	vars := map[string]any{}
	if len(wf.Variables) > 0 {
		for _, v := range wf.Variables {
			var input string
			reader := bufio.NewReader(cmd.InOrStdin())
			if len(v.Options) > 0 {
				ops := []string{}
				for _, v := range v.Options {
					ops = append(ops, cast.ToString(v))
				}
				input, err = h.promptUserSelect(fmt.Sprintf("Please select a value for variable %s", v.Name), ops, cast.ToString(v.DefaultValue))
				if err != nil {
					return fmt.Errorf("failed to read user selected input: %w", err)
				}
			} else {
				fmt.Println("Please input variable value (leave empty to use default one), press Enter to finish.")
				fmt.Printf("Variable Name: %s, Variable Type: %s, Default Value: %s\n", v.Name, v.Type, cast.ToString(v.DefaultValue))
				fmt.Print("> ")
				input, err = readLine(reader)
				if err != nil {
					return fmt.Errorf("failed to read prompt: %w", err)
				}
				input = strings.TrimSpace(input)
				if input == "" {
					input = cast.ToString(v.DefaultValue)
				}
			}
			var iv any = input
			switch v.Type {
			case schema.WorkflowVariableTypeInteger:
				iv, err = cast.ToIntE(input)
				if err != nil {
					return fmt.Errorf("failed to convert input value to integer: %w", err)
				}
			case schema.WorkflowVariableTypeNumber:
				iv, err = cast.ToFloat64E(input)
				if err != nil {
					return fmt.Errorf("failed to convert input value to number: %w", err)
				}
			case schema.WorkflowVariableTypeString:
			case schema.WorkflowVariableTypeFile:
				b, err := os.ReadFile(cast.ToString(iv))
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				// assign data to another var
				vars[fmt.Sprintf("%s__data", cast.ToString(iv))] = workflow.FileInfo{
					Name: cast.ToString(iv),
					Data: b,
				}
				iv = cast.ToString(iv)
			}
			vars[v.Name] = iv
		}
	}
	runner, err := h.backend.WorkflowService.Start(cmd.Context(), args[0], workflow.StartWorklfowRequest{
		Model:       model,
		ImageModel:  imageModel,
		Temperature: temperature,
		Variables:   vars,
	})
	if err != nil {
		return err
	}
	for {
		result, err := runner.Next(cmd.Context())
		if err != nil {
			return err
		}
		if result == nil {
			break
		}
		switch result.Action {
		case workflow.WorkflowActionShowMessage:
			fmt.Println(result.Message)
		case workflow.WorkflowActionExport:
			//nolint:gosec
			err := os.WriteFile(result.ExportPath, []byte(result.ExportData), 0644)
			if err != nil {
				log.Fatalf("failed to write file: %v", err)
			}
			fmt.Println(result.Message)
		case workflow.WorkflowActionGenerate:
			fmt.Println(result.Message)
			generator := result.Generator
			indexer := util.NewColumnIndexer(generator.Table().Edges.Columns)
			tp := h.getPrinter()
			headers := []string{"[ID]"}
			headers = append(headers, indexer.ColumnNames()...)
			tp.AddHeader(headers)
			for {
				batch, err := generator.Next(cmd.Context())
				if err != nil {
					return err
				}
				if len(batch) == 0 {
					break
				}
				for _, row := range batch {
					tp.AddField(cast.ToString(row["__id__"].Value))
					v, err := indexer.RowMapToSlice(row)
					if err != nil {
						return err
					}
					for _, cell := range v {
						sv := cellString(cell.Value)
						tp.AddField(sv)
					}
					tp.EndRow()
				}
				err = tp.Render()
				if err != nil {
					return err
				}
			}
		}
	}
	h.backend.Logger.Infow("workflow complete")
	return nil
}

func (h *Handler) CreateWorkflow(cmd *cobra.Command, args []string) error {
	var req workflow.Workflow
	f, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	err = json.Unmarshal(f, &req)
	if err != nil {
		return err
	}
	id, err := h.backend.WorkflowService.Create(cmd.Context(), &req)
	if err != nil {
		return err
	}
	h.backend.Logger.Infow("workflow created", "id", id)
	return nil
}

func (h *Handler) DeleteWorkflow(cmd *cobra.Command, args []string) error {
	err := h.backend.WorkflowService.Delete(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	h.backend.Logger.Infow("workflow deleted", "workflow", args[0])
	return nil
}

func (h *Handler) ListWorkflows(cmd *cobra.Command, args []string) error {
	wfs, err := h.backend.WorkflowService.List(cmd.Context())
	if err != nil {
		return err
	}
	tp := h.getPrinter()
	tp.AddHeader([]string{"ID", "Name", "Description"})
	for _, w := range wfs {
		tp.AddField(w.Nanoid)
		tp.AddField(w.Name)
		tp.AddField(w.Description)
		tp.EndRow()
	}
	return tp.Render()
}
