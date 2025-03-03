package table

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"
	"github.com/Yiling-J/tablepilot/services/table/source"

	"github.com/invopop/jsonschema"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRowsGenerator_PrepareRow(t *testing.T) {
	sc := &source.ListSource{Options: []string{"foo"}}
	err := sc.Init(context.TODO())
	require.NoError(t, err)
	generator := &AIRowsGenerator{
		sourceMap: map[string]source.Source{
			"c1": sc,
		},
		table: &ent.TableMeta{Edges: ent.TableMetaEdges{
			Columns: []*ent.TableColumn{
				{Nanoid: "c1"},
				{Nanoid: "c2"},
			},
		}},
	}
	err = generator.prepareRow(context.TODO())
	require.NoError(t, err)
	require.Equal(t, map[string]*schema.CellValue{"c1": {Value: "foo"}}, generator.rows[0])
}

func TestRowsGenerator_PrepareContextRows(t *testing.T) {
	cases := []struct {
		generated int
		expected  []any
	}{
		{0, []any{"12", "11", "10"}},
		{2, []any{"1", "0", "12", "11", "10"}},
		{3, []any{"2", "1", "0", "12", "11"}},
		{5, []any{"4", "3", "2", "1", "0"}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			ctx := context.TODO()
			db := db.NewTestDB()
			tb, err := db.TableMeta.Create().SetName("foo").Save(ctx)
			require.NoError(t, err)
			col, err := db.TableColumn.Create().
				SetName("c1").
				SetFillMode(tablecolumn.FillModeAi).
				SetTablemeta(tb).
				SetContextLength(5).
				SetType(tablecolumn.TypeString).Save(ctx)
			require.NoError(t, err)
			generator, err := NewRowsGenerator(ctx, GenerateRowsParams{
				Table: "foo",
				Batch: 5,
				Count: 5,
			}, db, nil, zap.NewNop().Sugar())
			require.NoError(t, err)

			generator.contextLength = 5
			for i := 0; i < c.generated; i++ {
				generator.generated = append(generator.generated, map[string]*schema.CellValue{
					col.Nanoid: {Value: cast.ToString(i)},
				})
			}
			for i := 0; i < 3; i++ {
				err = db.TableRow.Create().SetTablemeta(tb).SetCells(
					[]*schema.CellValue{{Value: cast.ToString(10 + i)}},
				).Exec(ctx)
				require.NoError(t, err)
			}
			err = generator.newBatch(ctx, 5)
			require.NoError(t, err)
			p, err := generator.builder.Prompt()
			require.NoError(t, err)
			eb := promptbuilder.NewRowsBuilder(5)
			err = eb.AddColumnContextData(col.Nanoid, c.expected)
			require.NoError(t, err)
			ep, err := eb.Prompt()
			require.NoError(t, err)
			require.Equal(t, ep, p)
		})
	}
}

func TestRowsGenerator_Chat(t *testing.T) {
	ctx := context.Background()
	var schema *jsonschema.Schema
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			schema = request.Schema
			return &client.ChatResponse{
				Content: "",
			}, nil
		},
	}

	generator := &AIRowsGenerator{
		ai: aiService,
		missingColumns: []*ent.TableColumn{
			{Nanoid: "n1", Type: tablecolumn.TypeArray},
			{Nanoid: "n2", Type: tablecolumn.TypeString},
		},
		builder: promptbuilder.NewRowsBuilder(1),
		table:   &ent.TableMeta{Model: "test"},
	}
	_, err := generator.chat(ctx)
	require.NoError(t, err)
	expectedSchema := `{"properties":{"data":{"items":{"properties":{"id":{"type":"integer"},"n1":{"items":{"type":"string"},"type":"array"},"n2":{"type":"string"}},"additionalProperties":false,"type":"object","required":["n1","n2"]},"type":"array"}},"additionalProperties":false,"type":"object"}`
	bs, err := schema.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, expectedSchema, string(bs))
}

func TestRowsGenerator_Next(t *testing.T) {
	cases := []struct {
		count     int
		batch     int
		chatBatch int  // how many rows returned from chat API
		chatCount int  // how many times chat API called
		saveTo    bool // generated data will ruturn without saving
	}{
		{10, 1, 1, 10, false},
		{10, 2, 2, 5, false},
		{10, 3, 3, 4, false},
		{10, 8, 8, 2, false},
		{10, 10, 10, 1, false},
		{10, 10, 20, 1, false},
		{10, 10, 3, 4, false},
		{10, 10, 3, 4, true},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			db := db.NewTestDB()
			ctx := context.Background()
			aiService := &ai.AiServiceMock{
				ChatFunc: func(
					ctx context.Context, request *client.ChatRequest,
				) (*client.ChatResponse, error) {
					data := []map[string]any{}
					for i := 0; i < tc.chatBatch; i++ {
						data = append(data, map[string]any{
							"name": "go",
						})
					}
					b, err := json.Marshal(map[string]any{"data": data})
					require.NoError(t, err)
					return &client.ChatResponse{
						Content: string(b),
					}, nil
				},
			}
			tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
			require.NoError(t, err)
			err = db.TableColumn.Create().
				SetName("c").
				SetFillMode(tablecolumn.FillModeAi).
				SetTablemeta(tb).
				SetType(tablecolumn.TypeString).Exec(ctx)
			require.NoError(t, err)

			st := ""
			if tc.saveTo {
				st = "abc"
			}
			generator, err := NewRowsGenerator(ctx, GenerateRowsParams{
				Table:  tb.Nanoid,
				Count:  tc.count,
				Batch:  tc.batch,
				SaveTo: st,
			}, db, aiService, zap.NewNop().Sugar())
			require.NoError(t, err)
			l := []int{}
			for {
				v, err := generator.Next(ctx)
				require.NoError(t, err)
				if len(v) == 0 {
					break
				}
				l = append(l, len(v))
			}
			left := 10
			for i, v := range l {
				if i != len(l)-1 {
					require.Equal(t, v, tc.chatBatch)
					left -= v
				} else {
					require.Equal(t, v, left)
				}
			}
			require.Equal(t, tc.chatCount, len(aiService.ChatCalls()))
			c, err := tb.QueryRows().Count(ctx)
			require.NoError(t, err)
			if tc.saveTo {
				require.Equal(t, 0, c)
			} else {
				require.Equal(t, tc.count, c)
			}
		})
	}
}
