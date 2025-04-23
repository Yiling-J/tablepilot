package source

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"

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
					require.Equal(t, p, request.Messages[0].Content[0].Data)
					require.Equal(t, 1.0, request.PresencePenalty)
					require.Equal(t, OptionGenMaxTokens, int(request.MaxOutputTokens))
					return &client.ChatResponse{
						Content: string(b),
					}, nil
				},
			}
			so := &AISource{
				BasicSource: BasicSource{
					Type: "ai",
				},
				Prompt: "aiai",
			}
			if hasOption {
				so.Options = []string{"go"}
			}
			err = so.Init(ctx, aiService, &ent.TableColumn{Name: "table", Description: "a table"}, "")
			require.NoError(t, err)
			if hasOption {
				require.Equal(t, so.Options, []string{"go", "foo", "bar"})
			} else {
				require.Equal(t, so.Options, []string{"foo", "bar"})
			}
			indexer := NewIndexer(so, &ent.TableColumn{Random: false})

			v, err := indexer.Next(ctx)
			require.NoError(t, err)
			if hasOption {
				require.Equal(t, "go", v.Value)
			} else {
				require.Equal(t, "foo", v.Value)
			}
		})
	}
}
