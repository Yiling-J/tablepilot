package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type FileOffset struct {
	File   uint16
	Total  uint8 // each chunk has max 50 rows
	Offset int64
}

type CSVIndexer struct {
	Files       []string
	Count       int
	Positions   []FileOffset
	ColumnNames []string
}

// Dataset holds the schema definition for the Dataset entity.
type Dataset struct {
	ent.Schema
}

func (Dataset) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
		NanoidMixin{},
	}
}

// Fields of the Dataset.
func (Dataset) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().NotEmpty(),
		field.String("path").Optional(),
		field.String("description").Default(""),
		field.Enum("type").Values("list", "csv"),
		field.JSON("indexer", CSVIndexer{}).Optional(),
		field.Strings("values"),
	}
}

// Edges of the Dataset.
func (Dataset) Edges() []ent.Edge {
	return nil
}
