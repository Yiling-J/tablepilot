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

// DatasetAPIRequest is used for API calls where files are base64 encoded strings
type DatasetAPIRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Type        string   `json:"type" binding:"required,oneof=list csv"`
	Data        []string `json:"data"` // For list type
	Files       []string `json:"files"`      // For csv type, array of base64 encoded file contents
	FileNames   []string `json:"file_names"` // Optional corresponding file names for CSV type
}
