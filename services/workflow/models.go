package workflow

import (
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/table"
)

type Workflow struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Variables   []schema.WorkflowVariable `json:"variables"`
	Steps       []schema.WorkflowStep     `json:"steps"`
}

type WorkflowSimple struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StepContext struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

func (s StepContext) AsMap() map[string]string {
	return map[string]string{
		"table":  s.Table,
		"column": s.Column,
	}
}

type WorkflowDeleteTableParams struct {
	Table string `json:"table"`
}

type WorkflowCreateColumnParams struct {
	Table       string `json:"table"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type WorkflowDeleteColumnParams struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

type WorkflowExportTableParams struct {
	Table string `json:"table"`
	Path  string `json:"path"`
}

type WorkflowImportFileParams struct {
	Table    string `json:"table"`
	Name     string `json:"name"`
	File     string `json:"file"`
	Prompt   string `json:"prompt"`
	Truncate bool   `json:"truncate"`
}

type StartWorklfowRequest struct {
	Variables   map[string]any `json:"variables"`
	Model       string         `json:"model"`
	ImageModel  string         `json:"image_model"`
	Temperature float64        `json:"temperature"`
}

type WorkflowCreateTablePayload struct {
	SchemaFile string                       `json:"schema_file"`
	OnExists   schema.WorkflowTableOnExists `json:"on_exists"`
	Request    table.TableGenRequest        `json:"request"`
}

type WorkflowAutofillPayload struct {
	Count   int      `json:"count"`
	Batch   int      `json:"batch"`
	Table   string   `json:"table"`
	Columns []string `json:"columns"`
}

type FileInfo struct {
	Name string
	Data []byte
}
