package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tablepilot/services/table"
	"tablepilot/services/table/util"
	"tablepilot/utils/tableprinter"
	"time"

	"github.com/spf13/cobra"
)

type Handler struct {
	backend    *Backend
	getPrinter func() tableprinter.TablePrinter
}

func NewHandler(backend *Backend) *Handler {
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
		id, err := h.backend.tableService.CreateTable(cmd.Context(), &req)
		if err != nil {
			return err
		}
		h.backend.Logger.Infow("table created", "id", id)
	}
	return nil
}

func (h *Handler) Show(cmd *cobra.Command, args []string) error {
	table := args[0]
	rows, err := h.backend.tableService.Rows(cmd.Context(), table)
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
	resp, err := h.backend.tableService.ListTables(cmd.Context())
	if err != nil {
		return err
	}
	tp := h.getPrinter()
	tp.AddHeader([]string{"ID", "Name"})
	for _, table := range resp.Tables {
		tp.AddField(table.ID)
		tp.AddField(table.Name)
		tp.EndRow()
	}
	return tp.Render()
}

func (h *Handler) Describe(cmd *cobra.Command, args []string) error {
	resp, err := h.backend.tableService.GetTableColumns(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	tp := h.getPrinter()
	tp.AddHeader([]string{"ID", "Name", "Type", "Fill Mode", "Description"})
	for _, column := range resp {
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
	count, err := h.backend.tableService.Delete(cmd.Context(), table)
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
	rows, err := h.backend.tableService.Rows(cmd.Context(), table)
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
	table := args[0]
	generator, err := h.backend.tableService.Genetate(
		cmd.Context(), table, saveTo, count, batch,
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
	importName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	tableFile := args[0]
	reader, err := os.Open(tableFile)
	if err != nil {
		return err
	}
	defer func() { _ = reader.Close() }()
	if importName == "" {
		importName = strings.TrimSuffix(tableFile, filepath.Ext(tableFile))
	}
	id, err := h.backend.tableService.Import(cmd.Context(), importName, reader)
	if err != nil {
		return err
	}
	h.backend.Logger.Infow("table imported", "id", id)
	return nil
}

func (h *Handler) Truncate(cmd *cobra.Command, args []string) error {
	table := args[0]
	removed, err := h.backend.tableService.Truncate(cmd.Context(), table)
	if err != nil {
		return err
	}
	h.backend.Logger.Infow("table truncated", "removed", removed)
	return nil
}
