package schema

import (
	"encoding/json"
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
		field.String("name").Unique().NotEmpty().Match(regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]*$")),
		field.String("description").Default(""),
		field.String("model").Default(""),
		field.JSON("sources", map[string]json.RawMessage{}).Optional(),
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
