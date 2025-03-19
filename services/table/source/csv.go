package source

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"slices"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/table/source/csvindexer"
	"github.com/bmatcuk/doublestar/v4"
)

type CsvSource struct {
	randomCSV      *csvindexer.CSVIndexer
	Type           string   `json:"type"`
	Paths          []string `json:"paths"`
	Column         string   `json:"column"`
	ContextColumns []string `json:"context_columns"`
}

func (cs *CsvSource) Init(ctx context.Context) error {
	fileSystem := os.DirFS("./")
	files, err := parsePaths(fileSystem, cs.Paths)
	if err != nil {
		return err
	}
	rs, err := csvindexer.NewCSVIndexer(files)
	if err != nil {
		return err
	}
	cs.randomCSV = rs
	return nil
}

func (cs *CsvSource) GetColumns(ctx context.Context) ([]string, error) {
	fileSystem := os.DirFS("./")
	files, err := parsePaths(fileSystem, cs.Paths)
	if err != nil {
		return nil, err
	}
	return csvindexer.GetColumnsFromFiles(files)
}

func (cs *CsvSource) NextLinked(ctx context.Context, idx int, column string, contextColumns []string) (*schema.CellValue, error) {
	row, err := cs.randomCSV.Fetch(idx)
	if err != nil {
		return nil, err
	}
	cv := &schema.CellValue{ContextValue: map[string]any{}}
	for i, col := range cs.randomCSV.Columns() {
		if col == column {
			cv.Value = row[i]
		}
		if slices.Contains(contextColumns, col) {
			cv.ContextValue[col] = row[i]
		}
	}

	return cv, nil
}

func (cs *CsvSource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return nil, errors.New("not implemented")
}

func (cs *CsvSource) Total() int {
	return cs.randomCSV.Total()
}

func parsePaths(fileSystem fs.FS, paths []string) ([]string, error) {
	fm := map[string]bool{}
	results := []string{}
	for _, p := range paths {
		files, err := doublestar.Glob(fileSystem, p)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if _, ok := fm[f]; !ok {
				fm[f] = true
				results = append(results, f)
			}
		}
	}
	return results, nil
}
