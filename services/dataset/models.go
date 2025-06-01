package dataset

import (
	"io"

	db_dataset "github.com/Yiling-J/tablepilot/ent/dataset"
)

type CreateDatasetRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Data        []string    `json:"data"`  // for list type
	Files       []io.Reader `json:"files"` // for csv type
}

type DatasetInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ColumnCount int    `json:"column_count"`
	ValueCount  int    `json:"value_count"`
}

type DatasetRows struct {
	Rows []map[string]any `json:"rows"`
	Data []string         `json:"data"`
	Type db_dataset.Type  `json:"type"`
}
