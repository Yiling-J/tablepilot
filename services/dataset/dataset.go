package dataset

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	db_dataset "github.com/Yiling-J/tablepilot/ent/dataset"
	"github.com/Yiling-J/tablepilot/services/source"
	"github.com/Yiling-J/tablepilot/services/source/csvindexer"
)

type DatasetService interface {
	Get(ctx context.Context, source string)
	List(ctx context.Context)
	Create(ctx context.Context, req *CreateDatasetRequest)
	Update(ctx context.Context, source string, req *CreateDatasetRequest)
	Preview(ctx context.Context, source string) (*DatasetRows, error)
	First(ctx context.Context)
}

type DatasetServiceImpl struct {
	db  *ent.Client
	cfg *config.Config
}

func (s *DatasetServiceImpl) Create(ctx context.Context, req *CreateDatasetRequest) (string, error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("dataset.Create: starting a transaction: %w", err)
	}
	sr, err := tx.Dataset.Create().SetName(req.Name).SetDescription(req.Description).SetType(
		db_dataset.Type(req.Type),
	).Save(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("dataset.Create: ent create: %w", err))
	}
	switch req.Type {
	case "csv":
		dirPath := filepath.Join(s.cfg.Common.SourceDataDir, "shared", sr.Nanoid)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		files := []string{}
		for i, file := range req.Files {
			filePath := filepath.Join(dirPath, fmt.Sprintf("%d.csv", i))
			files = append(files, filePath)
			outFile, err := os.Create(filePath)
			if err != nil {
				return "", fmt.Errorf("failed to create file %s: %w", filePath, err)
			}

			_, err = io.Copy(outFile, file)
			outFile.Close()
			if err != nil {
				return "", fmt.Errorf("failed to write to file %s: %w", filePath, err)
			}
		}
		// build index
		indexer, err := csvindexer.NewCSVIndexer(os.DirFS(dirPath), files)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.Create: build csv index: %w", err))
		}
		err = sr.Update().SetPath(fmt.Sprintf("shared/%s", sr.Nanoid)).SetIndexer(indexer.CSVIndexer).Exec(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.Create: ent update dataset: %w", err))
		}
	case "list":
		err = sr.Update().SetValues(req.Data).Exec(ctx)
		if err != nil {
			return "", ent.Rollback(tx, fmt.Errorf("table.Create: ent update dataset: %w", err))
		}
	}
	return sr.Nanoid, tx.Commit()
}

func (s *DatasetServiceImpl) Get(ctx context.Context, source string) (*DatasetInfo, error) {
	sr, err := s.db.Dataset.Query().Where(db_dataset.Or(
		db_dataset.Name(source),
		db_dataset.Nanoid(source),
	)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataset.Get: get from db: %w", err)
	}
	return &DatasetInfo{
		Name:        sr.Name,
		Description: sr.Description,
		ColumnCount: len(sr.Indexer.ColumnNames),
		ValueCount:  len(sr.Values),
	}, nil
}

// preview source data, csv will show top 100 rows and list will show all data
func (s *DatasetServiceImpl) Preview(ctx context.Context, dataset string) (*DatasetRows, error) {
	sr, err := s.db.Dataset.Query().Where(db_dataset.Or(
		db_dataset.Name(dataset),
		db_dataset.Nanoid(dataset),
	)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataset.Preview: get from db: %w", err)
	}
	switch sr.Type {
	case "csv":
		s := &source.CsvSource{RandomCSV: &csvindexer.CSVIndexer{
			CSVIndexer: sr.Indexer,
		}}
		counter := 0
		rows := []map[string]any{}
		columns := sr.Indexer.ColumnNames
		s.Range(func(row []any) bool {
			rowm := map[string]any{}
			for i, v := range row {
				rowm[columns[i]] = v
			}
			rows = append(rows, rowm)
			counter += 1
			return counter != 100
		})
		return &DatasetRows{
			Type: sr.Type,
			Rows: rows,
		}, nil
	case "list":
		return &DatasetRows{
			Type: sr.Type,
			Data: sr.Values,
		}, nil
	}
	return nil, fmt.Errorf("unknown source type: %s", sr.Type)
}
