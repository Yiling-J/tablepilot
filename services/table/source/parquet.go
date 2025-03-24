package source

import (
	"context"
	"errors"
	"os"
	"slices"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/table/source/parquet"
	"go.uber.org/zap"
)

type ParquetSource struct {
	reader         *parquet.ParquetGoReader
	Type           string   `json:"type"`
	Paths          []string `json:"paths"`
	Column         string   `json:"column"`
	ContextColumns []string `json:"context_columns"`
}

func (ps *ParquetSource) getRoot(ctx context.Context, logger *zap.SugaredLogger, dir string) (*os.Root, string, error) {
	rootPath := dir
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, "", err
	}
	return root, rootPath, nil
}

func (ps *ParquetSource) Init(ctx context.Context, logger *zap.SugaredLogger, dir string) error {
	root, _, err := ps.getRoot(ctx, logger, dir)
	if err != nil {
		return err
	}
	fileSystem := root.FS()
	files, err := parsePaths(fileSystem, ps.Paths)
	if err != nil {
		return err
	}
	ps.reader, err = parquet.NewParquetGoReader(ctx, fileSystem, files)
	if err != nil {
		return err
	}
	return nil
}

func (ps *ParquetSource) GetColumns(ctx context.Context, logger *zap.SugaredLogger, dir string) ([]string, error) {
	root, _, err := ps.getRoot(ctx, logger, dir)
	if err != nil {
		return nil, err
	}
	fileSystem := root.FS()
	files, err := parsePaths(fileSystem, ps.Paths)
	if err != nil {
		return nil, err
	}
	ps.reader, err = parquet.NewParquetGoReader(ctx, fileSystem, files)
	if err != nil {
		return nil, err
	}
	return ps.reader.Columns(), nil
}

func (ps *ParquetSource) NextLinked(ctx context.Context, idx int, column string, contextColumns []string) (*schema.CellValue, error) {
	rows, err := ps.reader.Rows(ctx, 1, int64(idx))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("parquet reader get 0 rows")
	}
	row := rows[0]
	cv := &schema.CellValue{ContextValue: map[string]any{}}
	columns := ps.reader.Columns()
	if err != nil {
		return nil, err
	}

	for i, col := range columns {
		if col == column {
			cv.Value = row[i]
		}
		if slices.Contains(contextColumns, col) {
			cv.ContextValue[col] = row[i]
		}
	}

	return cv, nil
}

func (ps *ParquetSource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return nil, errors.New("not implemented")
}

func (ps *ParquetSource) Total() int {
	total, _ := ps.reader.Total(context.TODO())
	return int(total)
}
