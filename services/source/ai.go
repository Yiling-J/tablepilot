package source

import (
	"context"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"
)

type AISource struct {
	BasicSource
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
}

func (as *AISource) Init(ctx context.Context, ai ai.AiService, column *ent.TableColumn, model string) error {
	builder := promptbuilder.NewColumnOptionsBuilder(column.Name, column.Description, as.Prompt)

	if len(as.Options) > 0 {
		builder.AddExampleOptions(as.Options)
	}
	p, err := builder.Prompt()
	if err != nil {
		return err
	}
	ops, err := ai.GenerateListOptions(ctx, model, p)
	if err != nil {
		return err
	}
	as.Options = append(as.Options, ops...)
	return nil
}

func (as *AISource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return &schema.CellValue{Value: as.Options[idx]}, nil
}

func (as *AISource) Total() int {
	return len(as.Options)
}
