package table

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
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

func TestRowsGenerator_PrepareRows(t *testing.T) {
	sc := &source.ListSource{Options: []string{"foo"}}
	err := sc.Init(context.TODO())
	require.NoError(t, err)
	idx := source.NewIndexer(sc, false, false, 1)
	generator := &AIRowsGenerator{
		indexerMap: map[string]*source.Indexer{
			"c1": idx,
		},
		table: &ent.TableMeta{Edges: ent.TableMetaEdges{
			Columns: []*ent.TableColumn{
				{Nanoid: "c1"},
				{Nanoid: "c2"},
			},
		}},
	}
	err = generator.prepareRows(context.TODO(), 1)
	require.NoError(t, err)
	require.Equal(t, map[string]*schema.CellValue{"c1": {Value: "foo"}, "id": {Value: 0}}, generator.rows[0])
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
			tb, err := db.TableMeta.Create().SetName("foo").SetDescription("bar").Save(ctx)
			require.NoError(t, err)
			col, err := db.TableColumn.Create().
				SetName("c1").
				SetFillMode(tablecolumn.FillModeAi).
				SetTablemeta(tb).
				SetContextLength(5).
				SetType(tablecolumn.TypeString).Save(ctx)
			require.NoError(t, err)
			generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
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
			eb.AddDescription("bar")
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
	expectedSchema := `{"properties":{"data":{"items":{"properties":{"id":{"type":"integer"},"n1":{"items":{"type":"string"},"type":"array"},"n2":{"type":"string"}},"additionalProperties":false,"type":"object","required":["id","n1","n2"]},"type":"array"}},"additionalProperties":false,"type":"object"}`
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
							"id":   i,
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
			generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
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

func TestRowsGenerator_Prompt(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		db := db.NewTestDB()
		ctx := context.Background()
		var promptContent string
		aiService := &ai.AiServiceMock{
			ChatFunc: func(
				ctx context.Context, request *client.ChatRequest,
			) (*client.ChatResponse, error) {
				promptContent = request.Messages[0].Content
				b, err := json.Marshal(map[string]any{"data": "d"})
				require.NoError(t, err)
				return &client.ChatResponse{
					Content: string(b),
				}, nil
			},
		}
		tb, err := db.TableMeta.Create().SetName("table").SetDescription("bar").Save(ctx)
		require.NoError(t, err)
		col, err := db.TableColumn.Create().
			SetName("c").
			SetFillMode(tablecolumn.FillModeAi).
			SetTablemeta(tb).
			SetType(tablecolumn.TypeString).Save(ctx)
		require.NoError(t, err)

		generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
			Table: tb.Nanoid,
			Count: 2,
			Batch: 2,
		}, db, aiService, zap.NewNop().Sugar())
		require.NoError(t, err)
		_, err = generator.Next(ctx)
		require.NoError(t, err)

		builder := promptbuilder.NewRowsBuilder(2)
		builder.AddDescription("bar")
		builder.AddTableColumns([]*ent.TableColumn{col}, false)
		builder.AddMissingColumns([]*ent.TableColumn{col})
		p, err := builder.Prompt()

		require.NoError(t, err)
		require.Equal(t, p, promptContent)
	})

	t.Run("with context", func(t *testing.T) {
		db := db.NewTestDB()
		ctx := context.Background()
		var promptContent string
		aiService := &ai.AiServiceMock{
			ChatFunc: func(
				ctx context.Context, request *client.ChatRequest,
			) (*client.ChatResponse, error) {
				promptContent = request.Messages[0].Content
				b, err := json.Marshal(map[string]any{"data": "d"})
				require.NoError(t, err)
				return &client.ChatResponse{
					Content: string(b),
				}, nil
			},
		}
		tb, err := db.TableMeta.Create().SetName("table").SetDescription("bar").Save(ctx)
		require.NoError(t, err)
		col, err := db.TableColumn.Create().
			SetName("c").
			SetFillMode(tablecolumn.FillModeAi).
			SetTablemeta(tb).
			SetContextLength(2).
			SetType(tablecolumn.TypeString).Save(ctx)
		require.NoError(t, err)

		generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
			Table: tb.Nanoid,
			Count: 2,
			Batch: 2,
		}, db, aiService, zap.NewNop().Sugar())
		require.NoError(t, err)
		for i := 0; i < 5; i++ {
			generator.generated = append(generator.generated, map[string]*schema.CellValue{
				col.Nanoid: {Value: cast.ToString(i)},
			})
		}
		_, err = generator.Next(ctx)
		require.NoError(t, err)

		builder := promptbuilder.NewRowsBuilder(2)
		builder.AddDescription("bar")
		err = builder.AddColumnContextData(col.Nanoid, []any{4, 3})
		require.NoError(t, err)
		builder.AddTableColumns([]*ent.TableColumn{col}, false)
		builder.AddMissingColumns([]*ent.TableColumn{col})
		p, err := builder.Prompt()

		require.NoError(t, err)
		require.Equal(t, p, promptContent)
	})

	t.Run("pick-type column", func(t *testing.T) {
		db := db.NewTestDB()
		ctx := context.Background()
		sc := &source.ListSource{Options: []string{"a", "b"}, Type: "list"}
		sb, err := json.Marshal(sc)
		require.NoError(t, err)
		tb, err := db.TableMeta.Create().SetName("table").SetDescription("bar").SetSources(map[string]json.RawMessage{"so": sb}).Save(ctx)
		require.NoError(t, err)
		col, err := db.TableColumn.Create().
			SetName("c").
			SetFillMode(tablecolumn.FillModeAi).
			SetTablemeta(tb).
			SetType(tablecolumn.TypeString).Save(ctx)
		require.NoError(t, err)
		col2, err := db.TableColumn.Create().
			SetName("c2").
			SetFillMode(tablecolumn.FillModePick).
			SetSource("so").
			SetTablemeta(tb).
			SetType(tablecolumn.TypeString).Save(ctx)
		require.NoError(t, err)
		var promptContent string
		aiService := &ai.AiServiceMock{
			ChatFunc: func(
				ctx context.Context, request *client.ChatRequest,
			) (*client.ChatResponse, error) {
				promptContent = request.Messages[0].Content
				data := []map[string]any{{"id": 0, col.Nanoid: "d"}, {"id": 1, col.Nanoid: "e"}}
				b, err := json.Marshal(map[string]any{"data": data})
				require.NoError(t, err)
				return &client.ChatResponse{
					Content: string(b),
				}, nil
			},
		}
		generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
			Table: tb.Nanoid,
			Count: 2,
			Batch: 2,
		}, db, aiService, zap.NewNop().Sugar())
		require.NoError(t, err)
		_, err = generator.Next(ctx)
		require.NoError(t, err)

		builder := promptbuilder.NewRowsBuilder(2)
		builder.AddDescription("bar")
		builder.AddTableColumns([]*ent.TableColumn{col, col2}, false)
		builder.AddMissingColumns([]*ent.TableColumn{col})
		err = builder.AddExistings([]map[string]any{
			{col2.Nanoid: "a", "id": 0},
			{col2.Nanoid: "b", "id": 1},
		})
		require.NoError(t, err)
		p, err := builder.Prompt()

		require.NoError(t, err)
		require.Equal(t, p, promptContent)
	})
}

func TestRowsGenerator_Autofill(t *testing.T) {
	cases := []AutofillRequest{
		{Offset: 0, Columns: []string{"c1"}},
		{Offset: 0, Columns: []string{"c1"}, ContextColumns: []string{""}},
		{Offset: 2, Columns: []string{"c1"}},
		{Offset: 4, Columns: []string{"c1"}}, // autofill the last row
		{Offset: 5, Columns: []string{"c1"}}, // autofill no rows, because offset > total row count
		{Offset: 8, Columns: []string{"c1"}}, // autofill no rows, because offset > total row count
		{Offset: 0, Columns: []string{"c1", "c2"}},
		{Offset: 0, Columns: []string{"c1"}, ContextColumns: []string{"c2"}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			db := db.NewTestDB()
			ctx := context.Background()
			var promptContent string
			tb, err := db.TableMeta.Create().SetName("table").SetDescription("bar").Save(ctx)
			require.NoError(t, err)
			col1, err := db.TableColumn.Create().
				SetName("c1").
				SetFillMode(tablecolumn.FillModeAi).
				SetTablemeta(tb).
				SetType(tablecolumn.TypeString).Save(ctx)
			require.NoError(t, err)
			col2, err := db.TableColumn.Create().
				SetName("c2").
				SetFillMode(tablecolumn.FillModeAi).
				SetTablemeta(tb).
				SetType(tablecolumn.TypeString).Save(ctx)
			require.NoError(t, err)
			dbrows, err := db.TableRow.CreateBulk([]*ent.TableRowCreate{
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v10"}, {Value: "v20"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v11"}, {Value: "v21"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v12"}, {Value: "v22"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v13"}, {Value: "v23"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v14"}, {Value: "v24"}}),
			}...).Save(ctx)
			require.NoError(t, err)
			offset := tc.Offset
			aiService := &ai.AiServiceMock{
				ChatFunc: func(
					ctx context.Context, request *client.ChatRequest,
				) (*client.ChatResponse, error) {
					promptContent = request.Messages[0].Content
					data := []map[string]any{}
					for i := 0; i < 2; i++ {
						id := "x"
						if i+offset < 5 {
							id = dbrows[i+offset].Nanoid
						}
						data = append(data, map[string]any{
							"name": "go",
							"id":   id,
						})
					}
					offset += 2
					b, err := json.Marshal(map[string]any{"data": data})
					require.NoError(t, err)
					return &client.ChatResponse{
						Content: string(b),
					}, nil
				},
			}

			generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
				Table: tb.Nanoid,
				Count: 2,
				Batch: 2,
				Autofill: AutofillRequest{
					Enable:         true,
					Offset:         tc.Offset,
					Columns:        tc.Columns,
					ContextColumns: tc.ContextColumns,
				},
			}, db, aiService, zap.NewNop().Sugar())
			require.NoError(t, err)
			for {
				v, err := generator.Next(ctx)
				require.NoError(t, err)
				if len(v) == 0 {
					break
				}
			}
			if tc.Offset < 5 {
				require.Equal(t, 1, len(aiService.ChatCalls()))
			} else {
				require.Equal(t, 0, len(aiService.ChatCalls()))
			}
			c, err := tb.QueryRows().Count(ctx)
			require.NoError(t, err)
			require.Equal(t, 5, c)

			if tc.Offset >= 5 {
				return
			}

			builder := promptbuilder.NewRowsBuilder(2)
			builder.AddDescription("bar")
			builder.AddTableColumns([]*ent.TableColumn{col1, col2}, true)
			missing := []*ent.TableColumn{}
			for _, col := range tc.Columns {
				if col == "c1" {
					missing = append(missing, col1)
				}
				if col == "c2" {
					missing = append(missing, col2)
				}
			}
			builder.AddMissingColumns(missing)

			rows := []map[string]any{}
			for i := 0; i < 2; i++ {
				if i+tc.Offset >= 5 {
					break
				}
				row := map[string]any{"id": dbrows[i+tc.Offset].Nanoid}
				if len(tc.ContextColumns) == 0 && !slices.Contains(tc.Columns, "c2") {
					tc.ContextColumns = []string{"c2"}
				}
				for _, col := range tc.ContextColumns {
					if col == "c2" {
						row[col2.Nanoid] = dbrows[i+tc.Offset].Cells[1].Value
					}
				}
				rows = append(rows, row)
			}
			err = builder.AddExistings(rows)
			require.NoError(t, err)

			p, err := builder.Prompt()

			require.NoError(t, err)
			require.Equal(t, p, promptContent)
		})
	}
}

func TestRowsGenerator_AutofillNext(t *testing.T) {
	cases := []struct {
		count     int
		batch     int
		chatBatch int // how many rows returned from chat API
		chatCount int // how many times chat API called
	}{
		{15, 3, 3, 4},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			db := db.NewTestDB()
			ctx := context.Background()
			tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
			require.NoError(t, err)
			err = db.TableColumn.Create().
				SetName("c1").
				SetFillMode(tablecolumn.FillModeAi).
				SetTablemeta(tb).
				SetType(tablecolumn.TypeString).Exec(ctx)
			require.NoError(t, err)
			col2, err := db.TableColumn.Create().
				SetName("c2").
				SetFillMode(tablecolumn.FillModeAi).
				SetTablemeta(tb).
				SetType(tablecolumn.TypeString).Save(ctx)
			require.NoError(t, err)
			dbrows, err := db.TableRow.CreateBulk([]*ent.TableRowCreate{
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v10"}, {Value: "v20"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v11"}, {Value: "v21"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v12"}, {Value: "v22"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v13"}, {Value: "v23"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v14"}, {Value: "v24"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v15"}, {Value: "v25"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v16"}, {Value: "v26"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v17"}, {Value: "v27"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v18"}, {Value: "v28"}}),
				db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v19"}, {Value: "v29"}}),
			}...).Save(ctx)
			require.NoError(t, err)
			offset := 0
			aiService := &ai.AiServiceMock{
				ChatFunc: func(
					ctx context.Context, request *client.ChatRequest,
				) (*client.ChatResponse, error) {
					data := []map[string]any{}
					for i := 0; i < tc.chatBatch; i++ {
						id := "x"
						if i+offset < 10 {
							id = dbrows[i+offset].Nanoid
						}
						data = append(data, map[string]any{
							col2.Nanoid: "go_" + id,
							"id":        id,
						})
					}
					offset += tc.chatBatch
					b, err := json.Marshal(map[string]any{"data": data})
					require.NoError(t, err)
					return &client.ChatResponse{
						Content: string(b),
					}, nil
				},
			}

			generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
				Table: tb.Nanoid,
				Count: tc.count,
				Batch: tc.batch,
				Autofill: AutofillRequest{
					Enable:  true,
					Columns: []string{"c"},
				},
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
			rows, err := tb.QueryRows().All(ctx)
			require.NoError(t, err)
			require.Equal(t, 10, len(rows))
			for _, row := range rows {
				require.Equal(t, row.Nanoid, strings.TrimPrefix(cast.ToString(row.Cells[1].Value), "go_"))
			}
		})
	}
}

func TestRowsGenerator_AutofillPartial(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Exec(ctx)
	require.NoError(t, err)
	col2, err := db.TableColumn.Create().
		SetName("c2").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	_, err = db.TableColumn.Create().
		SetName("c3").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	_, err = db.TableColumn.Create().
		SetName("c4").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	dbrows, err := db.TableRow.CreateBulk([]*ent.TableRowCreate{
		db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
			{Value: "v10"}, {Value: "v20"}, {Value: "v30"}, {Value: "v40"},
		}),
	}...).Save(ctx)
	require.NoError(t, err)
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			data := []map[string]any{
				{"id": dbrows[0].Nanoid, col2.Nanoid: "foobar"},
			}
			b, err := json.Marshal(map[string]any{"data": data})
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
			}, nil
		},
	}

	generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
		Table: tb.Nanoid,
		Count: 1,
		Batch: 1,
		Autofill: AutofillRequest{
			Enable:         true,
			Columns:        []string{"c1"},
			ContextColumns: []string{col2.Nanoid},
		},
	}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)
	for {
		v, err := generator.Next(ctx)
		require.NoError(t, err)
		if len(v) == 0 {
			break
		}
	}
	rows, err := tb.QueryRows().All(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(rows))
	require.Equal(t, rows[0].Cells, []*schema.CellValue{
		{Value: "v10"}, {Value: "foobar"}, {Value: "v30"}, {Value: "v40"},
	})
}
