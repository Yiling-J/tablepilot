package dataset

import (
	"io"
	"mime/multipart"

	db_dataset "github.com/Yiling-J/tablepilot/ent/dataset"
)

type CreateDatasetRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        db_dataset.Type `json:"type"`
	Data        []string        `json:"data"`  // for list type
	Files       []io.Reader     `json:"files"` // for csv type
	Private     bool            `json:"private"`
}

type UpdateDatasetRequest struct {
	CreateDatasetRequest
	Fields []string `json:"fields"`
}

type DatasetInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	ColumnCount int      `json:"column_count"`
	ValueCount  int      `json:"value_count"`
	Data        []string `json:"data"`
}

type DatasetRows struct {
	Rows []map[string]any `json:"rows"`
	Data []string         `json:"data"`
	Type db_dataset.Type  `json:"type"`
}

// DatasetAPIRequest is used for API calls where files are base64 encoded strings
type DatasetAPIRequest struct {
	Name        string                  `form:"name"`
	Description string                  `form:"description"`
	Type        string                  `form:"type"`
	Data        []string                `form:"data"`
	Files       []*multipart.FileHeader `form:"files"`
}
