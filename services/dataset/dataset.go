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

//go:generate moq -rm -out dataset_moq.go . DatasetService
type DatasetService interface {
	Get(ctx context.Context, source string) (*DatasetInfo, error)
	List(ctx context.Context) ([]*DatasetInfo, error)
	Create(ctx context.Context, req *CreateDatasetRequest) (string, error)
	Update(ctx context.Context, dataset string, req *UpdateDatasetRequest) error
	Delete(ctx context.Context, dataset string) error
	Preview(ctx context.Context, source string) (*DatasetRows, error)
}

type DatasetServiceImpl struct {
	db  *ent.Client
	cfg *config.Config
}

func NewDatasetService(db *ent.Client, cfg *config.Config) *DatasetServiceImpl {
	return &DatasetServiceImpl{
		db:  db,
		cfg: cfg,
	}
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

func (s *DatasetServiceImpl) List(ctx context.Context) ([]*DatasetInfo, error) {
	datasets, err := s.db.Dataset.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataset.List: query all: %w", err)
	}
	var datasetInfos []*DatasetInfo
	for _, ds := range datasets {
		datasetInfos = append(datasetInfos, &DatasetInfo{
			Name:        ds.Name,
			Description: ds.Description,
			Type:        string(ds.Type),
			ColumnCount: len(ds.Indexer.ColumnNames),
			ValueCount:  len(ds.Values),
		})
	}
	return datasetInfos, nil
}

func (s *DatasetServiceImpl) Update(ctx context.Context, dataset string, req *UpdateDatasetRequest) error {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("dataset.Update: starting a transaction: %w", err)
	}
	ds, err := tx.Dataset.Query().Where(db_dataset.Or(
		db_dataset.Name(dataset),
		db_dataset.Nanoid(dataset),
	)).Only(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("dataset.Update: get from db: %w", err))
	}

	updater := ds.Update()
	updateData := false
	for _, f := range req.Fields {
		switch f {
		case "name":
			updater.SetName(req.Name)
		case "description":
			updater.SetDescription(req.Description)
		case "data", "files":
			updateData = true
			updater.ClearIndexer().ClearPath().SetValues(nil)
		}
	}
	if !updateData {
		_, err = updater.Save(ctx)
		if err != nil {
			return ent.Rollback(tx, fmt.Errorf("dataset.Update: ent update: %w", err))
		}
		return tx.Commit()
	}

	if ds.Path != "" {
		oldDirPath := filepath.Join(s.cfg.Common.SourceDataDir, ds.Path)
		if _, err := os.Stat(oldDirPath); !os.IsNotExist(err) {
			if err := os.RemoveAll(oldDirPath); err != nil {
				return ent.Rollback(tx, fmt.Errorf("dataset.Update: failed to remove old directory %s: %w", oldDirPath, err))
			}
		}
	}

	switch req.Type {
	case "csv":
		dirPath := filepath.Join(s.cfg.Common.SourceDataDir, "shared", ds.Nanoid)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return ent.Rollback(tx, fmt.Errorf("dataset.Update: failed to create directory: %w", err))
		}

		files := []string{}
		for i, file := range req.Files {
			filePath := filepath.Join(dirPath, fmt.Sprintf("%d.csv", i))
			files = append(files, filePath)
			outFile, err := os.Create(filePath)
			if err != nil {
				return ent.Rollback(tx, fmt.Errorf("dataset.Update: failed to create file %s: %w", filePath, err))
			}

			_, err = io.Copy(outFile, file)
			outFile.Close()
			if err != nil {
				return ent.Rollback(tx, fmt.Errorf("dataset.Update: failed to write to file %s: %w", filePath, err))
			}
		}
		indexer, err := csvindexer.NewCSVIndexer(os.DirFS(dirPath), files)
		if err != nil {
			return ent.Rollback(tx, fmt.Errorf("dataset.Update: build csv index: %w", err))
		}
		updater.SetPath(fmt.Sprintf("shared/%s", ds.Nanoid)).SetIndexer(indexer.CSVIndexer)
	case "list":
		updater.SetValues(req.Data)
	default:
		return ent.Rollback(tx, fmt.Errorf("dataset.Update: unknown dataset type: %s", req.Type))
	}

	_, err = updater.Save(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("dataset.Update: ent update: %w", err))
	}

	return tx.Commit()
}

func (s *DatasetServiceImpl) Delete(ctx context.Context, dataset string) error {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("dataset.Delete: starting a transaction: %w", err)
	}
	ds, err := tx.Dataset.Query().Where(db_dataset.Or(
		db_dataset.Name(dataset),
		db_dataset.Nanoid(dataset),
	)).Only(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("dataset.Delete: get from db: %w", err))
	}

	if ds.Type == db_dataset.TypeCsv && ds.Path != "" {
		dirPath := filepath.Join(s.cfg.Common.SourceDataDir, ds.Path)
		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			if err := os.RemoveAll(dirPath); err != nil {
				return ent.Rollback(tx, fmt.Errorf("dataset.Delete: failed to remove directory %s: %w", dirPath, err))
			}
		}
	}

	err = tx.Dataset.DeleteOne(ds).Exec(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("dataset.Delete: ent delete: %w", err))
	}

	return tx.Commit()
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
