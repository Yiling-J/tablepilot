package source

import (
	"context"
	"encoding/json"
	"tablepilot/ent"
	"tablepilot/ent/schema"
	"tablepilot/services/ai"
	"tablepilot/services/ai/client"
	"tablepilot/services/ai/promptbuilder"

	"github.com/invopop/jsonschema"
)

var reflector = jsonschema.Reflector{
	AllowAdditionalProperties: false,
	DoNotReference:            true,
}

type AISource struct {
	indexer
	Type    string   `json:"type"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
}

type data struct {
	Options []string
}

func (as *AISource) Init(ctx context.Context, ai ai.AiService, column *ent.TableColumn) error {
	builder := promptbuilder.NewColumnOptionsBuilder(column.Name, column.Description, as.Prompt)

	if len(as.Options) > 0 {
		builder.AddExampleOptions(as.Options)
	}
	p, err := builder.Prompt()
	if err != nil {
		return err
	}
	resp, err := ai.Chat(ctx, &client.ChatRequest{
		Messages:        []*client.Message{client.UserMessage(p)},
		Schema:          reflector.Reflect(data{}),
		Temperature:     0.8,
		PresencePenalty: 1.0,
	})
	if err != nil {
		return err
	}
	var d data
	err = json.Unmarshal([]byte(resp.Content), &d)
	if err != nil {
		return err
	}
	as.Options = append(as.Options, d.Options...)
	as.indexer = newIndexer(as.Random, as.Replacement, len(as.Options))
	return nil
}

func (as *AISource) Next(ctx context.Context) (*schema.CellValue, error) {
	return &schema.CellValue{Value: as.Options[as.nextIndex()]}, nil
}
