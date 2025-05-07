package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Model holds the schema definition for the Model entity.
type Model struct {
	ent.Schema
}

func (Model) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the Model.
func (Model) Fields() []ent.Field {
	return []ent.Field{
		field.String("model").Unique().NotEmpty(),
		field.String("alias").Optional(),
		field.Int("rpm").Optional(),
		field.Int("max_tokens").Optional(),
		field.Bool("default").Default(false),
		field.Bool("image").Default(false),
	}
}

// Edges of the Model.
func (Model) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("provider", Provider.Type).
			Ref("models").
			Unique(),
	}
}
