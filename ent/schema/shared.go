package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type TimeMixin struct {
	mixin.Schema
}

func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").Optional().
			Immutable().
			Default(time.Now),
		field.Time("updated_at").Optional().
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

type NanoidMixin struct {
	mixin.Schema
}

func (NanoidMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("nanoid").Unique().Optional(),
	}
}
