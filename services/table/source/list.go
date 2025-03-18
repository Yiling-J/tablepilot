package source

import (
	"context"

	"github.com/Yiling-J/tablepilot/ent/schema"
)

type ListSource struct {
	Type    string   `json:"type"`
	Options []string `json:"options"`
}

func (ls *ListSource) Init(ctx context.Context) error {
	return nil
}

func (ls *ListSource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return &schema.CellValue{Value: ls.Options[idx]}, nil
}

func (ls *ListSource) Total() int {
	return len(ls.Options)
}
