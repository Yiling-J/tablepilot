package source

import (
	"context"
	"errors"
	"os"
	"slices"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/table/source/huggingface"
	"github.com/Yiling-J/tablepilot/services/table/source/parquet"
	"go.uber.org/zap"
)

type Huggingface struct {
	Dataset string `json:"dataset"`
	Config  string `json:"config"`
	Split   string `json:"split"`
	client  huggingface.Client
	size    int
}

type ParquetSource struct {
	reader         *parquet.ParquetGoReader
	Type           string   `json:"type"`
	Paths          []string `json:"paths"`
	Column         string   `json:"column"`
	ContextColumns []string `json:"context_columns"`
	Huggingface    *Huggingface
}

func (ps *ParquetSource) getRoot(ctx context.Context, logger *zap.SugaredLogger, dir string) (*os.Root, string, error) {
	rootPath := dir
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, "", err
	}
	return root, rootPath, nil
}

func (ps *ParquetSource) Init(ctx context.Context, hfClient huggingface.Client, logger *zap.SugaredLogger, dir string) error {
	if hfClient != nil {
		ps.Huggingface.client = hfClient
		resp, err := ps.Huggingface.client.GetDatasetSize(ctx)
		if err != nil {
			return err
		}
		for _, split := range resp.Size.Splits {
			if split.Split == ps.Huggingface.Split && split.Config == ps.Huggingface.Config {
				ps.Huggingface.size = split.NumRows
			}
		}
		return nil
	}
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
	if ps.Huggingface != nil {
		info, err := ps.Huggingface.client.GetDatasetInfo(ctx)
		if err != nil {
			return nil, err
		}
		columns := []string{}
		for column := range info.Features {
			columns = append(columns, column)
		}
		return columns, nil
	}
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
	if ps.Huggingface != nil {
		rp, err := ps.Huggingface.client.GetDatasetRows(ctx, idx, 1)
		if err != nil {
			return nil, err
		}
		if len(rp.Rows) == 0 {
			return nil, errors.New("huggingface rows API return 0 rows")
		}
		row := rp.Rows[0]
		cv := &schema.CellValue{ContextValue: map[string]any{}}
		for col, val := range row.Row {
			if col == column {
				cv.Value = val
			}
			if slices.Contains(contextColumns, col) {
				cv.ContextValue[col] = val
			}
		}
		return cv, nil
	}

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
	if ps.Huggingface != nil {
		return ps.Huggingface.size
	}
	total, _ := ps.reader.Total(context.TODO())
	return int(total)
}
