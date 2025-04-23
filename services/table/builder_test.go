package table

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"
	"github.com/Yiling-J/tablepilot/services/table/source"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuilder_GenerateBuilderTables(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	builderTables := []BuilderTable{
		{Name: "t1", Description: "d1"},
		{Name: "t2", Description: "d2", Depends: []string{"t1"}},
	}
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			require.Equal(t, "user", request.Messages[0].Role)
			require.Equal(t, "m1", request.Model)
			require.Equal(t, 0.32, request.Temperature)
			pb := promptbuilder.NewTablesBuilder("gogo")
			prompt, err := pb.Prompt()
			require.NoError(t, err)
			require.Equal(t, prompt, request.Messages[0].Content[0].Data)
			b, err := json.Marshal(builderTables)
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
				Tokens:  100,
			}, nil
		},
	}
	srv, err := NewTableService(&config.Config{}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)

	tables, err := srv.GenerateBuilderTables(ctx, "gogo", ModelParams{
		Model:       "m1",
		Temperature: 0.32,
	})
	require.NoError(t, err)
	require.Equal(t, builderTables, tables)
}

func TestBuilder_PolishBuilderTables(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	builderTables := []BuilderTable{
		{Name: "t1", Description: "d1"},
		{Name: "t2", Description: "d2", Depends: []string{"t1"}},
	}
	tr, err := json.Marshal(map[string]any{"data": builderTables})
	require.NoError(t, err)
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			require.Equal(t, "user", request.Messages[0].Role)
			require.Equal(t, "m1", request.Model)
			require.Equal(t, 0.32, request.Temperature)
			pb := promptbuilder.NewTablesPolishBuilder(string(tr), "gogo")
			prompt, err := pb.Prompt()
			require.NoError(t, err)
			require.Equal(t, reflector.Reflect([]BuilderTable{}), request.Schema)
			require.Equal(t, prompt, request.Messages[0].Content[0].Data)
			b, err := json.Marshal([]BuilderTable{
				{Name: "t1", Description: "d-1"},
				{Name: "t3", Description: "d3", Depends: []string{"t1"}},
			})
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
				Tokens:  100,
			}, nil
		},
	}
	srv, err := NewTableService(&config.Config{}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)

	tables, err := srv.PolishBuilderTables(ctx, []BuilderTable{
		{Name: "t1", Description: "d1"},
		{Name: "t2", Description: "d2", Depends: []string{"t1"}},
	}, "gogo", ModelParams{Model: "m1",
		Temperature: 0.32})
	require.NoError(t, err)
	require.Equal(t, []BuilderTable{
		{Name: "t1", Description: "d-1"},
		{Name: "t3", Description: "d3", Depends: []string{"t1"}},
	}, tables)
}

func TestBuilder_BuildTable(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	aiService := &ai.AiServiceMock{
		FunctionCallFunc: func(ctx context.Context, request *client.ChatRequest) (*client.FunctionCallResponse, error) {
			require.Equal(t, "m1", request.Model)
			require.Equal(t, 0.32, request.Temperature)
			pm := promptbuilder.NewTableGenBuilder("foo", "bar", []string{"baz"}, []promptbuilder.TableInfoSimple{
				{Name: "baz", Description: "abc", Columns: []promptbuilder.TableColumnSimple{
					{Name: "c1", Description: "c1d"},
				}},
			})
			message, err := pm.Prompt()
			require.NoError(t, err)
			require.Equal(t, message, request.Messages[0].Content[0].Data)
			require.Equal(t, getTools(false, false), request.Tools)
			return &client.FunctionCallResponse{
				FunctionCalls: []client.FunctionCall{
					{Name: "AddAiSource", Arguments: map[string]any{
						"name": "s1", "prompt": "go",
					}},
					{Name: "AddAiColumn", Arguments: map[string]any{
						"name": "c1", "description": "c1d", "type": "string", "contextLength": 3,
					}},
					{Name: "AddPickColumn", Arguments: map[string]any{
						"name":                 "c2",
						"description":          "c2d",
						"type":                 "string",
						"contextLength":        10,
						"random":               true,
						"repeat":               5,
						"replacement":          false,
						"source":               "externalSource",
						"linkedColumn":         "ln",
						"linkedContextColumns": []string{"lc1", "lc2"},
					}},
				},
			}, nil
		},
	}
	srv, err := NewTableService(&config.Config{}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)

	schema, err := srv.BuildTable(ctx, "foo", "bar", []string{"baz"}, []*TableInfo{
		{Name: "baz", Description: "abc", Columns: []TableColumnInfo{
			{Name: "c1", Description: "c1d"},
		}},
	}, ModelParams{Model: "m1",
		Temperature: 0.32})
	require.NoError(t, err)
	s := &source.AISource{
		BasicSource: source.BasicSource{
			Name: "s1",
			Type: "ai",
		},
		Prompt: "go",
	}
	bs, err := json.Marshal(s)
	require.NoError(t, err)
	require.Equal(t, &TableGenRequest{
		Name:        "foo",
		Description: "bar",
		Sources:     []json.RawMessage{bs},
		Columns: []TableGenColumn{
			{Name: "c1", Description: "c1d", Type: "string", FillMode: "ai", ContextLength: 3},
			{
				Name: "c2", Description: "c2d", Type: "string", FillMode: "pick", ContextLength: 10,
				Random: true, Repeat: 5, Source: "externalSource", LinkedColumn: "ln", LinkedContextColumns: []string{"lc1", "lc2"},
			},
		},
	}, schema)
}

func TestBuilder_PolishBuilderTable(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	s := &source.AISource{
		BasicSource: source.BasicSource{
			Name: "s1",
			Type: "ai",
		},
		Prompt: "go",
	}
	bs, err := json.Marshal(s)
	require.NoError(t, err)
	req := &TableGenRequest{
		Name:        "foo",
		Description: "bar",
		Sources:     []json.RawMessage{bs},
		Columns: []TableGenColumn{
			{Name: "c1", Description: "c1d", Type: "string", FillMode: "ai", ContextLength: 3},
			{
				Name: "c2", Description: "c2d", Type: "string", FillMode: "pick", ContextLength: 10,
				Random: true, Repeat: 5, Source: "externalSource", LinkedColumn: "ln", LinkedContextColumns: []string{"lc1", "lc2"},
			},
		},
	}
	aiService := &ai.AiServiceMock{
		FunctionCallFunc: func(ctx context.Context, request *client.ChatRequest) (*client.FunctionCallResponse, error) {
			require.Equal(t, "m1", request.Model)
			require.Equal(t, 0.32, request.Temperature)
			cb, err := json.Marshal(req.Columns)
			require.NoError(t, err)
			sb, err := json.Marshal(req.Sources)
			require.NoError(t, err)
			pm := promptbuilder.NewTablePolishBuilder("go", req.Name, req.Description, string(sb), string(cb), []promptbuilder.TableInfoSimple{
				{Name: "baz", Description: "abc", Columns: []promptbuilder.TableColumnSimple{
					{Name: "c1", Description: "c1d"},
				}},
			})
			message, err := pm.Prompt()
			require.NoError(t, err)
			require.Equal(t, message, request.Messages[0].Content[0].Data)
			require.Equal(t, getTools(true, false), request.Tools)
			return &client.FunctionCallResponse{
				FunctionCalls: []client.FunctionCall{
					{Name: "AddAiColumn", Arguments: map[string]any{
						"name": "c3", "description": "c3d", "type": "string", "contextLength": 0,
					}},
					{Name: "RemoveColumn", Arguments: map[string]any{
						"name": "c2",
					}},
				},
			}, nil
		},
	}
	srv, err := NewTableService(&config.Config{}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)

	schema, err := srv.PolishBuilderTable(ctx, req, "go", []*TableInfo{
		{Name: "baz", Description: "abc", Columns: []TableColumnInfo{
			{Name: "c1", Description: "c1d"},
		}},
	}, ModelParams{Model: "m1",
		Temperature: 0.32})
	require.NoError(t, err)
	require.Equal(t, &TableGenRequest{
		Name:        "foo",
		Description: "bar",
		Sources:     []json.RawMessage{bs},
		Columns: []TableGenColumn{
			{Name: "c1", Description: "c1d", Type: "string", FillMode: "ai", ContextLength: 3},
			{Name: "c3", Description: "c3d", Type: "string", FillMode: "ai", ContextLength: 0},
		},
	}, schema)
}
