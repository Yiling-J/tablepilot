package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Provider holds the schema definition for the Provider entity.
type Provider struct {
	ent.Schema
}

func (Provider) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the Provider.
func (Provider) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique(),
		field.String("type"),
		field.String("key").Optional(),
		field.String("base_url").Optional(),
		field.Bool("enabled").Default(true),
	}
}

// Edges of the Provider.
func (Provider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("models", Model.Type).Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
