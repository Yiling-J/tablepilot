package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type CellValue struct {
	Value        any            `json:"v"`
	ContextValue map[string]any `json:"c,omitempty"`
}

// TableData holds the schema definition for the TableData entity.
type TableRow struct {
	ent.Schema
}

func (TableRow) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		NanoidMixin{},
	}
}

// Fields of the TableData.
func (TableRow) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("cells", []*CellValue{}).Optional(),
	}
}

// Edges of the TableData.
func (TableRow) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tablemeta", TableMeta.Type).
			Ref("rows").Unique().Required(),
	}
}

func (TableRow) Indexes() []ent.Index {
	return []ent.Index{}
}
