package source

import (
	"context"
	"tablepilot/ent/schema"
)

type ListSource struct {
	indexer
	Type    string   `json:"type"`
	Options []string `json:"options"`
}

func (ls *ListSource) Init(ctx context.Context) error {
	ls.indexer = newIndexer(ls.Random, ls.Replacement, len(ls.Options))
	return nil
}

func (ls *ListSource) Next(ctx context.Context) (*schema.CellValue, error) {
	return &schema.CellValue{Value: ls.Options[ls.nextIndex()]}, nil
}
