package source

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"slices"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/table/source/csvindexer"
	"github.com/Yiling-J/tablepilot/services/table/source/kaggle"
	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"
)

type CsvSource struct {
	BasicSource
	randomCSV      *csvindexer.CSVIndexer
	Paths          []string `json:"paths"`
	Kaggle         string   `json:"kaggle"`
	Column         string   `json:"column"`
	ContextColumns []string `json:"context_columns"`
}

func (cs *CsvSource) getRoot(ctx context.Context, logger *zap.SugaredLogger, dir string) (*os.Root, string, error) {
	var root *os.Root
	rootPath := dir
	if cs.Kaggle != "" {
		rp, err := kaggle.PrepareKaggleDataset(ctx, logger, cs.Kaggle, dir)
		if err != nil {
			return nil, "", err
		}
		root, err = os.OpenRoot(rp)
		if err != nil {
			return nil, "", err
		}
		rootPath = rp
	} else {
		var err error
		root, err = os.OpenRoot(dir)
		if err != nil {
			return nil, "", err
		}
	}
	return root, rootPath, nil
}

func (cs *CsvSource) Init(ctx context.Context, logger *zap.SugaredLogger, dir string) error {
	root, _, err := cs.getRoot(ctx, logger, dir)
	if err != nil {
		return err
	}
	fileSystem := root.FS()
	files, err := parsePaths(fileSystem, cs.Paths)
	if err != nil {
		return err
	}
	rs, err := csvindexer.NewCSVIndexer(fileSystem, files)
	if err != nil {
		return err
	}
	cs.randomCSV = rs

	return nil
}

func (cs *CsvSource) GetColumns(ctx context.Context, logger *zap.SugaredLogger, dir string) ([]string, error) {
	root, _, err := cs.getRoot(ctx, logger, dir)
	if err != nil {
		return nil, err
	}
	fileSystem := root.FS()
	files, err := parsePaths(fileSystem, cs.Paths)
	if err != nil {
		return nil, err
	}
	return csvindexer.GetColumnsFromFiles(fileSystem, files)
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

func (cs *CsvSource) GetLinkedCellValue(row []any, column string, contextColumns []string) *schema.CellValue {
	cv := &schema.CellValue{ContextValue: map[string]any{}}
	if len(contextColumns) == 0 {
		contextColumns = cs.randomCSV.Columns()
	}
	for i, col := range cs.randomCSV.Columns() {
		if col == column {
			cv.Value = row[i]
		}
		if slices.Contains(contextColumns, col) {
			cv.ContextValue[col] = row[i]
		}
	}

	return cv
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

func (cs *CsvSource) Range(fn func(row []any) bool) error {
	return cs.randomCSV.Range(func(row []string) bool {
		ar := []any{}
		for _, cell := range row {
			ar = append(ar, cell)
		}
		return fn(ar)
	})
}
