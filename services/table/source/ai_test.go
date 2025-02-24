package source

import (
	"context"
	"encoding/json"
	"fmt"
	"tablepilot/ent"
	"tablepilot/services/ai"
	"tablepilot/services/ai/client"
	"tablepilot/services/ai/promptbuilder"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSource_AI(t *testing.T) {
	for _, hasOption := range []bool{false, true} {
		t.Run(fmt.Sprintf("has option %v", hasOption), func(t *testing.T) {
			ctx := context.TODO()
			mockData := data{
				Options: []string{"foo", "bar"},
			}
			b, err := json.Marshal(mockData)
			require.NoError(t, err)
			builder := promptbuilder.NewColumnOptionsBuilder("table", "a table", "aiai")
			if hasOption {
				builder.AddExampleOptions([]string{"go"})
			}
			p, err := builder.Prompt()
			require.NoError(t, err)
			aiService := &ai.AiServiceMock{
				ChatFunc: func(
					ctx context.Context, request *client.ChatRequest,
				) (*client.ChatResponse, error) {
					require.Equal(t, p, request.Messages[0].Content)
					require.Equal(t, 1.0, request.PresencePenalty)
					return &client.ChatResponse{
						Content: string(b),
					}, nil
				},
			}
			so := &AISource{
				indexer: newIndexer(false, false, 20, 0),
				Type:    "ai",
				Prompt:  "aiai",
			}
			if hasOption {
				so.Options = []string{"go"}
			}
			err = so.Init(ctx, aiService, &ent.TableColumn{Name: "table", Description: "a table"})
			require.NoError(t, err)
			if hasOption {
				require.Equal(t, so.Options, []string{"go", "foo", "bar"})
			} else {
				require.Equal(t, so.Options, []string{"foo", "bar"})
			}
			v, err := so.Next(ctx)
			require.NoError(t, err)
			if hasOption {
				require.Equal(t, "go", v.Value)
			} else {
				require.Equal(t, "foo", v.Value)
			}
		})
	}
}
