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
	Table  string               `json:"table"`
	Column table.TableGenColumn `json:"column"`
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
	Table  string `json:"table"`
	File   string `json:"file"`
	Prompt string `json:"prompt"`
}

type StartWorklfowRequest struct {
	Variables   map[string]any `json:"variables"`
	Model       string         `json:"model"`
	ImageModel  string         `json:"image_model"`
	Temperature float64        `json:"temperature"`
}
