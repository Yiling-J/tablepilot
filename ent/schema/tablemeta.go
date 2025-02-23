package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TableMeta holds the schema definition for the TableMeta entity.
type TableMeta struct {
	ent.Schema
}

func (TableMeta) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		NanoidMixin{},
	}
}

// Fields of the TableMeta.
func (TableMeta) Fields() []ent.Field {
	return []ent.Field{
		field.Int("max_rows").Default(100),
		field.String("prompt_raw").Default(""),
		field.String("prompt_gen").Default(""),
		field.String("name").Unique().NotEmpty().Match(regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")),
		field.String("description").Default(""),
		field.Enum("build_status").Values("init", "success", "failed").Default("init"),
		field.String("model").Default(""),
	}
}

// Edges of the TableMeta.
func (TableMeta) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("columns", TableColumn.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("rows", TableRow.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (TableMeta) Indexes() []ent.Index {
	return []ent.Index{}
}
