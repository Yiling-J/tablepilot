package dataset

import (
	"context"
	"errors"
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

func (s DatasetServiceImpl) buildCreateDatasetReq(ctx context.Context, req *CreateDatasetRequest, sr *ent.Dataset) error {
	switch req.Type {
	case db_dataset.TypeCsv:
		relativePath := filepath.Join("datasets/shared", sr.Nanoid)
		dirPath := filepath.Join(s.cfg.Common.DataDir, relativePath)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if len(req.Files) == 0 {
			return errors.New("dataset.Create: files should not be empty")
		}

		for _, file := range req.Files {
			outFile, err := os.Create(filepath.Join(dirPath, file.Name))
			if err != nil {
				return fmt.Errorf("failed to create file %w", err)
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, file.Reader)
			if err != nil {
				return fmt.Errorf("failed to write to file %w", err)
			}
		}

		// build index
		indexer, err := csvindexer.NewCSVIndexer(os.DirFS(dirPath), req.Data)
		if err != nil {
			return fmt.Errorf("table.Create: build csv index: %w", err)
		}
		err = sr.Update().SetPath(relativePath).SetIndexer(indexer.CSVIndexer).SetValues(req.Data).Exec(ctx)
		if err != nil {
			return fmt.Errorf("table.Create: update dataset metadata: %w", err)
		}
	case db_dataset.TypeImage:
		relativePath := filepath.Join("datasets/shared", sr.Nanoid)
		dirPath := filepath.Join(s.cfg.Common.DataDir, relativePath)
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if len(req.Files) == 0 {
			return errors.New("dataset.Create: files should not be empty")
		}
		for _, file := range req.Files {
			outFile, err := os.Create(filepath.Join(dirPath, file.Name))
			if err != nil {
				return fmt.Errorf("failed to create file %w", err)
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, file.Reader)
			if err != nil {
				return fmt.Errorf("failed to write to file %w", err)
			}
		}
		err = sr.Update().SetPath(relativePath).SetValues(req.Data).Exec(ctx)
		if err != nil {
			return fmt.Errorf("table.Create: update dataset metadata: %w", err)
		}
	case db_dataset.TypeList:
		err := sr.Update().SetValues(req.Data).Exec(ctx)
		if err != nil {
			return fmt.Errorf("table.Create: update dataset values: %w", err)
		}
	}
	return nil
}

func (s *DatasetServiceImpl) Create(ctx context.Context, req *CreateDatasetRequest) (string, error) {
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("dataset.Create: starting a transaction: %w", err)
	}
	sr, err := tx.Dataset.Create().SetName(req.Name).SetDescription(req.Description).SetType(
		req.Type,
	).Save(ctx)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("dataset.Create: save dataset: %w", err))
	}
	err = s.buildCreateDatasetReq(ctx, req, sr)
	if err != nil {
		return "", ent.Rollback(tx, fmt.Errorf("dataset.Create: buildCreateDatasetReq: %w", err))
	}
	return sr.Nanoid, tx.Commit()
}

func (s *DatasetServiceImpl) List(ctx context.Context) ([]*DatasetInfo, error) {
	datasets, err := s.db.Dataset.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataset.List: query all: %w", err)
	}
	datasetInfos := []*DatasetInfo{}
	for _, ds := range datasets {
		// backward compatible
		if ds.Type == db_dataset.TypeCsv && len(ds.Values) == 0 {
			ds.Values = []string{"data.csv"}
		}
		datasetInfos = append(datasetInfos, &DatasetInfo{
			ID:          ds.Nanoid,
			Name:        ds.Name,
			Description: ds.Description,
			Type:        string(ds.Type),
			ColumnCount: len(ds.Indexer.ColumnNames),
			ValueCount:  len(ds.Values),
			Data:        ds.Values,
			Columns:     ds.Indexer.ColumnNames,
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
		return ent.Rollback(tx, fmt.Errorf("dataset.Update: query dataset: %w", err)) // Clarified error
	}

	if req.Type != "" && req.Type != ds.Type {
		return ent.Rollback(tx, errors.New("dataset type cannot be changed via update"))
	}

	originalPath := ds.Path
	updater := ds.Update()
	processDataRebuild := false

	for _, f := range req.Fields {
		switch f {
		case "name":
			updater.SetName(req.Name)
		case "description":
			updater.SetDescription(req.Description)
		case "data", "files":
			processDataRebuild = true
			updater.ClearIndexer().ClearPath().SetValues(nil)
		}
	}

	updatedDsEntity, err := updater.Save(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("dataset.Update: save changes: %w", err))
	}

	if processDataRebuild {
		if originalPath != "" {
			oldDirPath := filepath.Join(s.cfg.Common.DataDir, originalPath)
			if _, statErr := os.Stat(oldDirPath); !os.IsNotExist(statErr) {
				keep := map[string]bool{}
				for _, file := range req.Data {
					keep[file] = true
				}
				entries, err := os.ReadDir(oldDirPath)
				if err != nil {
					return ent.Rollback(tx, fmt.Errorf("dataset.Update: read dir: %w", err))
				}
				for _, e := range entries {
					if _, ok := keep[e.Name()]; !ok {
						err = os.Remove(filepath.Join(oldDirPath, e.Name()))
						if err != nil {
							return ent.Rollback(tx, fmt.Errorf("dataset.Update: remove file: %w", err))
						}
					}
				}
			}
		}

		buildReq := req.CreateDatasetRequest
		buildReq.Type = updatedDsEntity.Type

		err = s.buildCreateDatasetReq(ctx, &buildReq, updatedDsEntity)
		if err != nil {
			return ent.Rollback(tx, fmt.Errorf("dataset.Update: buildCreateDatasetReq: %w", err))
		}
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
		return ent.Rollback(tx, fmt.Errorf("dataset.Delete: query dataset: %w", err)) // Clarified error
	}

	if ds.Type == db_dataset.TypeCsv && ds.Path != "" {
		dirPath := filepath.Join(s.cfg.Common.DataDir, ds.Path)
		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			if err := os.RemoveAll(dirPath); err != nil {
				return ent.Rollback(tx, fmt.Errorf("dataset.Delete: failed to remove directory %s: %w", dirPath, err))
			}
		}
	}

	err = tx.Dataset.DeleteOne(ds).Exec(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("dataset.Delete: execute delete: %w", err))
	}

	return tx.Commit()
}

func (s *DatasetServiceImpl) Get(ctx context.Context, source string) (*DatasetInfo, error) {
	sr, err := s.db.Dataset.Query().Where(db_dataset.Or(
		db_dataset.Name(source),
		db_dataset.Nanoid(source),
	)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataset.Get: query dataset: %w", err)
	}
	return &DatasetInfo{
		Name:        sr.Name,
		Description: sr.Description,
		Type:        string(sr.Type),
		ColumnCount: len(sr.Indexer.ColumnNames),
		ValueCount:  len(sr.Values),
		Data:        sr.Values,
	}, nil
}

// preview source data, csv will show top 100 rows and list will show all data
func (s *DatasetServiceImpl) Preview(ctx context.Context, dataset string) (*DatasetRows, error) {
	sr, err := s.db.Dataset.Query().Where(db_dataset.Or(
		db_dataset.Name(dataset),
		db_dataset.Nanoid(dataset),
	)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("dataset.Preview: query dataset: %w", err)
	}
	switch sr.Type {
	case db_dataset.TypeCsv:
		dir := fmt.Sprintf("%s/datasets/shared/%s", s.cfg.Common.DataDir, sr.Nanoid)
		s := &source.CsvSource{RandomCSV: &csvindexer.CSVIndexer{
			FS:         os.DirFS(dir),
			CSVIndexer: sr.Indexer,
		}}
		counter := 0
		rows := []map[string]any{}
		columns := sr.Indexer.ColumnNames
		err = s.Range(func(row []any) bool {
			rowm := map[string]any{}
			for i, v := range row {
				rowm[columns[i]] = v
			}
			rows = append(rows, rowm)
			counter += 1
			return counter != 100
		})
		if err != nil {
			return nil, fmt.Errorf("dataset.Preview: range csv data: %w", err)
		}
		return &DatasetRows{
			Type: sr.Type,
			Rows: rows,
		}, nil
	case db_dataset.TypeImage:
		dir := fmt.Sprintf("%s/datasets/shared/%s", s.cfg.Common.DataDir, sr.Nanoid)
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("dataset.Preview: read images dir: %w", err)
		}
		images := []string{}
		for _, entry := range entries {
			if !entry.IsDir() {
				images = append(
					images, fmt.Sprintf("datasets/shared/%s/%s", sr.Nanoid, entry.Name()),
				)
			}
		}
		return &DatasetRows{
			Type: sr.Type,
			Data: images,
		}, nil
	case db_dataset.TypeList:
		return &DatasetRows{
			Type: sr.Type,
			Data: sr.Values,
		}, nil
	}
	return nil, fmt.Errorf("unknown source type: %s", sr.Type)
}
