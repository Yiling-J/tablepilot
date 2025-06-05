package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Column holds the schema definition for the Column entity.
type TableColumn struct {
	ent.Schema
}

func (TableColumn) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		NanoidMixin{},
	}
}

// Fields of the Column.
func (TableColumn) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("description").Optional(),
		field.Enum("type").Values("string", "number", "integer", "boolean", "array", "image"),
		field.Enum("fill_mode").Values("ai", "pick"),
		field.String("source").Optional(), // deprecated
		field.String("source_id").Optional(),
		field.Enum("source_type").Optional().Values("table", "dataset"),
		field.Int("context_length").Default(0),
		field.Int("table_id"),
		field.Bool("random").Default(false),
		field.Bool("replacement").Default(false),
		field.Int("repeat").Default(1),
		field.String("linked_column").Default(""),
		field.Strings("linked_context_columns").Default([]string{}),
	}
}

// Edges of the Column.
func (TableColumn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tablemeta", TableMeta.Type).
			Ref("columns").Unique().Required().Field("table_id"),
	}
}

func (TableColumn) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "table_id").Unique(),
	}
}
