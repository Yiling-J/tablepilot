package schema

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type WorkflowVariableType string
type WorkflowStepType string
type WorkflowTableOnExists string

const (
	WorkflowVariableTypeString  WorkflowVariableType = "string"
	WorkflowVariableTypeNumber  WorkflowVariableType = "number"
	WorkflowVariableTypeInteger WorkflowVariableType = "integer"
	WorkflowVariableTypeFile    WorkflowVariableType = "file"
)

const (
	WorkflowStepTypeCreateTable  WorkflowStepType = "CreateTable"
	WorkflowStepTypeImport       WorkflowStepType = "Import"
	WorkflowStepTypeCreateColumn WorkflowStepType = "CreateColumn"
	WorkflowStepTypeDeleteColumn WorkflowStepType = "DeleteColumn"
	WorkflowStepTypeGenerate     WorkflowStepType = "Generate"
	WorkflowStepTypeAutofill     WorkflowStepType = "Autofill"
	WorkflowStepTypeExportTable  WorkflowStepType = "ExportTable"
	WorkflowStepTypeDeleteTable  WorkflowStepType = "DeleteTable"
)

const (
	OnExistsStop     WorkflowTableOnExists = "Stop"
	OnExistsRecreate WorkflowTableOnExists = "Recreate"
	OnExistsSkip     WorkflowTableOnExists = "Skip"
)

type WorkflowVariable struct {
	Name         string               `json:"name"`
	Type         WorkflowVariableType `json:"type"`
	DefaultValue any                  `json:"default_value"`
	Options      []any                `json:"options"`
}

type WorkflowStep struct {
	Type WorkflowStepType `json:"type"`
	// SchemaFile string                `json:"schema_file"`
	// OnExists   WorkflowTableOnExists `json:"on_exists"`
	Payload json.RawMessage `json:"payload"`
}

// Workflow holds the schema definition for the Workflow entity.
type Workflow struct {
	ent.Schema
}

func (Workflow) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		NanoidMixin{},
	}
}

// Fields of the Workflow.
func (Workflow) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Unique(),
		field.String("description").Optional(),
		field.JSON("variables", []WorkflowVariable{}),
		field.JSON("steps", []WorkflowStep{}),
	}
}

// Edges of the Workflow.
func (Workflow) Edges() []ent.Edge {
	return nil
}
