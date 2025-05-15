package workflow

import (
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/table"
)

type Workflow struct {
	Name      string                    `json:"name"`
	Variables []schema.WorkflowVariable `json:"variables"`
	Steps     []schema.WorkflowStep     `json:"steps"`
}

type StepContext struct {
	Table  string
	Column string
}

type WorkflowContext struct {
	Steps    []StepContext
	Date     string
	Time     string
	Datetime string
	Vars     map[string]any
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
