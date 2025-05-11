package table

import (
	"encoding/json"

	"github.com/Yiling-J/tablepilot/ent"
)

type TableGenColumn struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Type                 string   `json:"type"`
	FillMode             string   `json:"fill_mode"`
	Source               string   `json:"source"`
	Random               bool     `json:"random"`
	Replacement          bool     `json:"replacement"`
	Repeat               int      `json:"repeat"`
	ContextLength        int      `json:"context_length"`
	LinkedColumn         string   `json:"linked_column"`
	LinkedContextColumns []string `json:"linked_context_columns"`
}

type TableGenRequest struct {
	Name        string            `json:"name"`
	Model       string            `json:"model"`
	Description string            `json:"description"`
	Columns     []TableGenColumn  `json:"columns"`
	Sources     []json.RawMessage `json:"sources"`
	apiRequest  bool
}

func (r *TableGenRequest) MarkAPIRequest() {
	r.apiRequest = true
}

func (r *TableGenRequest) APIRequest() bool {
	return r.apiRequest
}

type AutofillRequest struct {
	Enable         bool     `json:"enable"`
	Offset         int      `json:"offset"`
	Columns        []string `json:"columns"`
	ContextColumns []string `json:"context_columns"`

	// fields below are used in regenerate API
	Rows   []string `json:"rows"`
	Prompt string   `json:"prompt"`
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
	// shared sources from config file
	sharedSources map[string]json.RawMessage
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

type ImageImportRequest struct {
	Data   []byte
	Prompt string
	Model  string
}

type ImageExtractionColumn struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type" jsonschema:"enum=string,enum=number,enum=integer,enum=boolean,enum=array"`
}

type ImageExtractionOutput struct {
	TableName string                  `json:"table_name"`
	Columns   []ImageExtractionColumn `json:"columns"`
	Rows      [][]string              `json:"rows"`
}
