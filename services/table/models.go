package table

import (
	"encoding/json"
	"io"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/dataset"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
)

type TableGenColumn struct {
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	Type                 string                 `json:"type"`
	FillMode             string                 `json:"fill_mode"`
	SourceType           tablecolumn.SourceType `json:"source_type"`
	SourceID             string                 `json:"source_id"`
	Options              []string               `json:"options"`
	Random               bool                   `json:"random"`
	Replacement          bool                   `json:"replacement"`
	Repeat               int                    `json:"repeat"`
	ContextLength        int                    `json:"context_length"`
	LinkedColumn         string                 `json:"linked_column"`
	LinkedContextColumns []string               `json:"linked_context_columns"`
}

type TableGenRequest struct {
	Name        string            `json:"name"`
	Model       string            `json:"model"`
	Description string            `json:"description"`
	Columns     []*TableGenColumn `json:"columns"`
}

type CLITableGenDataset struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Type        dataset.Type `json:"type"`
	Values      []string     `json:"values"`
	Paths       []string     `json:"paths"`
}

type CLITableGenRequest struct {
	Name        string            `json:"name"`
	Model       string            `json:"model"`
	Description string            `json:"description"`
	Columns     []*TableGenColumn `json:"columns"`
}

type AutofillRequest struct {
	Enable         bool     `json:"enable"`
	Offset         int      `json:"offset"`
	Columns        []string `json:"columns"`
	ContextColumns []string `json:"context_columns"`
	Prompt         string   `json:"prompt"`

	// fields below are used in regenerate API
	Rows []string `json:"rows"`
}

type GenerateRowsRequest struct {
	Table       string  `json:"table"`
	SaveTo      string  `json:"save_to"`
	Count       int     `json:"count"`
	Batch       int     `json:"batch"`
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"`
	ImageModel  string  `json:"image_model"`
	// used in API only to send streaming results
	Stream bool `json:"stream"`

	Autofill AutofillRequest `json:"autofill"`
	// used in file list source and csv source, the root fs for files
	sourceDataDir string
}

type ReGenerateRowsRequest struct {
	Rows        []string `json:"rows"`
	Table       string   `json:"table"`
	Temperature float64  `json:"temperature"`
	Model       string   `json:"model"`
	ImageModel  string   `json:"image_model"`
	Columns     []string `json:"columns"`
	Prompt      string   `json:"prompt"`
}

type ColumnSchema struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	LinkedTo string `json:"linked_to"`
	Mode     string `json:"mode"`
}

type ListTablesResponse struct {
	Total  int         `json:"total"`
	Tables []TableInfo `json:"tables"`
}

type TableColumnInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	FillMode    string `json:"fill_mode"`
}

type Rows struct {
	Columns []*ent.TableColumn `json:"columns"`
	Rows    []*ent.TableRow    `json:"rows"`
}

type TableInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Model       string            `json:"model"`
	Columns     []TableColumnInfo `json:"columns"`
}

type CreateRowsRequest struct {
	Rows []map[string]any `json:"rows"`
}

type SharedSource struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
	// only used in csv/parquet type source, this is required by API+WebUI, so user can select columns in UI
	Columns []string `json:"columns"`
}

type ModelParams struct {
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"`
}

type ImportRequest struct {
	Model string
	// import to existing table
	Table string
	// create a new table with given name
	Name string
	// remove existing rows in To
	Truncate bool
	// file name, if To is empty and Name is empty, the new table's name will using Filename_timestamp as new name
	Filename string

	// import image
	Data   []byte
	Prompt string
	// import csv
	Reader io.Reader
}

type ImageExtractionColumn struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type" jsonschema:"enum=string,enum=number,enum=integer,enum=boolean,enum=array"`
}

type ImageExtractionOutput struct {
	TableName        string                  `json:"table_name"`
	TableDescription string                  `json:"table_description"`
	Columns          []ImageExtractionColumn `json:"columns"`
	Rows             [][]string              `json:"rows"`
}
