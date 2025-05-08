package table

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
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
	err := sc.Init(context.TODO(), "")
	require.NoError(t, err)
	idx := source.NewIndexer(sc, &ent.TableColumn{
		Random: false,
	})
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
	require.Equal(t, map[string]*schema.CellValue{"c1": {Value: "foo"}, "__id__": {Value: 0}}, generator.rows[0])
}

func TestRowsGenerator_PrepareRowsSharedSource(t *testing.T) {
	ctx := context.TODO()
	db := db.NewTestDB()
	tb, err := db.TableMeta.Create().SetName("foo").SetDescription("bar").Save(ctx)
	require.NoError(t, err)
	col, err := db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModePick).
		SetSource("so").
		SetTablemeta(tb).
		SetContextLength(5).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
		Table: "foo",
		Batch: 5,
		Count: 5,
		sharedSources: map[string]json.RawMessage{
			"so": json.RawMessage(`{"type":"list","options":["o1"]}`),
		},
	}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	err = generator.prepareRows(context.TODO(), 1)
	require.NoError(t, err)
	require.Equal(t, map[string]*schema.CellValue{col.Nanoid: {Value: "o1"}, "__id__": {Value: 0}}, generator.rows[0])
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
	expectedSchema := `{"properties":{"data":{"items":{"properties":{"__id__":{"type":"integer"},"n1":{"items":{"type":"string"},"type":"array"},"n2":{"type":"string"}},"additionalProperties":false,"type":"object","required":["__id__","n1","n2"]},"type":"array"}},"additionalProperties":false,"type":"object"}`
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
							"name":   "go",
							"__id__": i,
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
				for _, row := range v {
					_, ok := row["__id__"]
					require.True(t, ok)
				}
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
				promptContent = request.Messages[0].Content[0].Data
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
		builder.AddMissingColumns([]*ent.TableColumn{col}, true)
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
				promptContent = request.Messages[0].Content[0].Data
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
		builder.AddMissingColumns([]*ent.TableColumn{col}, true)
		p, err := builder.Prompt()

		require.NoError(t, err)
		require.Equal(t, p, promptContent)
	})

	t.Run("pick-type column", func(t *testing.T) {
		db := db.NewTestDB()
		ctx := context.Background()
		sc := &source.ListSource{Options: []string{"a", "b"}, BasicSource: source.BasicSource{Type: "list"}}
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
				promptContent = request.Messages[0].Content[0].Data
				data := []map[string]any{{"__id__": 0, col.Nanoid: "d"}, {"__id__": 1, col.Nanoid: "e"}}
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
		builder.AddMissingColumns([]*ent.TableColumn{col}, true)
		err = builder.AddExistings([]map[string]any{
			{col2.Nanoid: "a", "__id__": 0},
			{col2.Nanoid: "b", "__id__": 1},
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
					promptContent = request.Messages[0].Content[0].Data
					data := []map[string]any{}
					for i := 0; i < 2; i++ {
						id := "x"
						if i+offset < 5 {
							id = dbrows[i+offset].Nanoid
						}
						data = append(data, map[string]any{
							"name":   "go",
							"__id__": id,
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
			builder.AddMissingColumns(missing, true)

			rows := []map[string]any{}
			for i := 0; i < 2; i++ {
				if i+tc.Offset >= 5 {
					break
				}
				row := map[string]any{"__id__": dbrows[i+tc.Offset].Nanoid}
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
							"__id__":    id,
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
					Columns: []string{"c2"},
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
				for _, row := range v {
					_, ok := row["__id__"]
					require.True(t, ok)
				}
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
				{"__id__": dbrows[0].Nanoid, col2.Nanoid: "foobar"},
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

func TestRowsGenerator_AutofillSelectedRowsWithPrompt(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
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
	col3, err := db.TableColumn.Create().
		SetName("c3").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	col4, err := db.TableColumn.Create().
		SetName("c4").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	dbrows, err := db.TableRow.CreateBulk([]*ent.TableRowCreate{
		db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
			{Value: "v10"}, {Value: "v20"}, {Value: "v30"}, {Value: "v40"},
		}),
		db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
			{Value: "v11"}, {Value: "v21"}, {Value: "v31"}, {Value: "v41"},
		}),
		db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
			{Value: "v12"}, {Value: "v22"}, {Value: "v32"}, {Value: "v42"},
		}),
		db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
			{Value: "v13"}, {Value: "v23"}, {Value: "v33"}, {Value: "v43"},
		}),
	}...).Save(ctx)
	require.NoError(t, err)
	counter := 0
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			idx := 1
			if counter == 1 {
				idx = 3
			}
			builder := promptbuilder.NewRowsBuilder(1)
			builder.AddDescription("bar")
			builder.AddUserPrompt("gogogo")
			builder.AddTableColumns([]*ent.TableColumn{col1, col2, col3, col4}, true)
			builder.AddMissingColumns([]*ent.TableColumn{col1}, true)

			rows := []map[string]any{
				{
					"__id__": dbrows[idx].Nanoid, col1.Nanoid: dbrows[idx].Cells[0].Value, col2.Nanoid: dbrows[idx].Cells[1].Value,
					col3.Nanoid: dbrows[idx].Cells[2].Value, col4.Nanoid: dbrows[idx].Cells[3].Value,
				},
			}
			err = builder.AddExistings(rows)
			require.NoError(t, err)
			p, err := builder.Prompt()
			require.NoError(t, err)
			require.Equal(t, p, request.Messages[0].Content[0].Data)

			data := []map[string]any{
				{"__id__": dbrows[idx].Nanoid, col1.Nanoid: "foobar"},
			}
			counter += 1
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
		Batch: 1,
		Autofill: AutofillRequest{
			Enable:         true,
			Columns:        []string{"c1"},
			ContextColumns: []string{"c1", "c2", "c3", "c4"},
			Rows:           []string{dbrows[1].Nanoid, dbrows[3].Nanoid},
			Prompt:         "gogogo",
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
	require.Equal(t, 2, counter)
	rows, err := tb.QueryRows().Order(ent.Asc("id")).All(ctx)
	require.NoError(t, err)
	require.Equal(t, 4, len(rows))
	for i, row := range rows {
		if i == 1 || i == 3 {
			require.Equal(t, row.Cells, []*schema.CellValue{
				{Value: "foobar"}, {Value: dbrows[i].Cells[1].Value}, {Value: dbrows[i].Cells[2].Value}, {Value: dbrows[i].Cells[3].Value},
			})
		} else {
			require.Equal(t, row.Cells, []*schema.CellValue{
				{Value: dbrows[i].Cells[0].Value}, {Value: dbrows[i].Cells[1].Value}, {Value: dbrows[i].Cells[2].Value}, {Value: dbrows[i].Cells[3].Value},
			})
		}
	}
}

func createPNG(filePath string) error {
	img := image.NewRGBA(image.Rect(0, 0, 5, 5))
	black := color.RGBA{0, 0, 0, 255}
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			img.Set(x, y, black)
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func TestRowsGenerator_PrepareImageRows(t *testing.T) {
	t.Run("generate", func(t *testing.T) {
		files := []string{"i1.png", "i2.png", "i1.png", "i3.png", "i2.png"}
		for _, f := range files {
			err := createPNG(f)
			require.NoError(t, err)
			defer func() { _ = os.Remove(f) }()
		}
		sc := &source.ListSource{Options: files}
		err := sc.Init(context.TODO(), "")
		require.NoError(t, err)
		idx := source.NewIndexer(sc, &ent.TableColumn{
			Random: false,
		})
		generator := &AIRowsGenerator{
			sourceDataDir: "./",
			indexerMap: map[string]*source.Indexer{
				"c1": idx,
			},
			table: &ent.TableMeta{Edges: ent.TableMetaEdges{
				Columns: []*ent.TableColumn{
					{Nanoid: "c1", Type: tablecolumn.TypeImage, Source: "c1", Random: false},
				},
			}},
			images: make(map[string]string),
		}
		err = generator.prepareRows(context.TODO(), 5)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"i1.png": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC", "i2.png": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC", "i3.png": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC"}, generator.images)
		for i, f := range files {
			require.Equal(t, map[string]*schema.CellValue{"c1": {Value: f}, "__id__": {Value: i}}, generator.rows[i])
		}
	})

	t.Run("generate-url", func(t *testing.T) {
		files := []string{"i1.png", "i2.png", "i1.png", "i3.png", "i2.png"}
		for i, f := range files {
			files[i] = "https://images.com/" + f
		}
		sc := &source.ListSource{Options: files}
		err := sc.Init(context.TODO(), "")
		require.NoError(t, err)
		idx := source.NewIndexer(sc, &ent.TableColumn{
			Random: false,
		})
		generator := &AIRowsGenerator{
			sourceDataDir: "./",
			indexerMap: map[string]*source.Indexer{
				"c1": idx,
			},
			table: &ent.TableMeta{Edges: ent.TableMetaEdges{
				Columns: []*ent.TableColumn{
					{Nanoid: "c1", Type: tablecolumn.TypeImage, Source: "c1", Random: false},
				},
			}},
			images: make(map[string]string),
		}
		err = generator.prepareRows(context.TODO(), 5)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"https://images.com/i1.png": "https://images.com/i1.png", "https://images.com/i2.png": "https://images.com/i2.png", "https://images.com/i3.png": "https://images.com/i3.png"}, generator.images)
		for i, f := range files {
			require.Equal(t, map[string]*schema.CellValue{"c1": {Value: f}, "__id__": {Value: i}}, generator.rows[i])
		}
	})

	t.Run("generate-dataurl", func(t *testing.T) {
		files := []string{
			"data:image/jpeg;base64,abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcbacbabbshcfudsfuibcugcguidkgkgfdsgfuigfuiehkadjagfdgfsdkfdksksdfgjksdgfkgfksdfksdgfwieufhgsfkdjfbskhfuwehfwesofhweiofhhfjksdfkjsgfjksfbwhefwohfshfwoifhiowhfsklfhlshfwiehfiowshfiowshfgwiehfiwhfowihgwioihnchfwifhv",
			"data:image/jpeg;base64,abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcbacbabbshcfersfuibcugcguidkgkgfdsgfuigfuiehkadjagfdgfsdkfdksksdfgjksdgfkgfksdfksdgfwieufhgsfkdjfbskhfuwehfwesofhweiofhhfjksdfkjsgfjksfbwhefwohfshfwoifhiowhfsklfhlshfwiehfiowshfiowshfgwiehfiwhfowihgwioihnchfwifhv",
		}
		sc := &source.ListSource{Options: files}
		err := sc.Init(context.TODO(), "")
		require.NoError(t, err)
		idx := source.NewIndexer(sc, &ent.TableColumn{
			Random: false,
		})
		generator := &AIRowsGenerator{
			sourceDataDir: "./",
			indexerMap: map[string]*source.Indexer{
				"c1": idx,
			},
			table: &ent.TableMeta{Edges: ent.TableMetaEdges{
				Columns: []*ent.TableColumn{
					{Nanoid: "c1", Type: tablecolumn.TypeImage, Source: "c1", Random: false},
				},
			}},
			images: make(map[string]string),
		}
		err = generator.prepareRows(context.TODO(), 2)
		require.NoError(t, err)
		require.Equal(t, map[string]string{
			"89fc4c78a70dc188887832116393e225": "data:image/jpeg;base64,abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcbacbabbshcfudsfuibcugcguidkgkgfdsgfuigfuiehkadjagfdgfsdkfdksksdfgjksdgfkgfksdfksdgfwieufhgsfkdjfbskhfuwehfwesofhweiofhhfjksdfkjsgfjksfbwhefwohfshfwoifhiowhfsklfhlshfwiehfiowshfiowshfgwiehfiwhfowihgwioihnchfwifhv",
			"d7f4f6b429c06ba9339ceabf1dfc9d95": "data:image/jpeg;base64,abcabcabcabcabcabcabcabcabcabcabcabcabcabcabcbacbabbshcfersfuibcugcguidkgkgfdsgfuigfuiehkadjagfdgfsdkfdksksdfgjksdgfkgfksdfksdgfwieufhgsfkdjfbskhfuwehfwesofhweiofhhfjksdfkjsgfjksfbwhefwohfshfwoifhiowhfsklfhlshfwiehfiowshfiowshfgwiehfiwhfowihgwioihnchfwifhv",
		}, generator.images)
		require.Equal(t, map[string]*schema.CellValue{"c1": {Value: "89fc4c78a70dc188887832116393e225"}, "__id__": {Value: 0}}, generator.rows[0])
		require.Equal(t, map[string]*schema.CellValue{"c1": {Value: "d7f4f6b429c06ba9339ceabf1dfc9d95"}, "__id__": {Value: 1}}, generator.rows[1])
	})

	t.Run("autofill", func(t *testing.T) {
		files := []string{"i1.png", "i2.png", "i1.png", "i3.png", "i2.png"}
		for _, f := range files {
			err := createPNG(f)
			require.NoError(t, err)
			defer func() { _ = os.Remove(f) }()
		}
		db := db.NewTestDB()
		ctx := context.TODO()
		tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
		require.NoError(t, err)
		err = db.TableColumn.Create().
			SetName("c0").
			SetFillMode(tablecolumn.FillModeAi).
			SetTablemeta(tb).
			SetType(tablecolumn.TypeString).Exec(ctx)
		require.NoError(t, err)
		err = db.TableColumn.Create().
			SetName("c1").
			SetFillMode(tablecolumn.FillModeAi).
			SetTablemeta(tb).
			SetType(tablecolumn.TypeImage).Exec(ctx)
		require.NoError(t, err)
		rows, err := db.TableRow.CreateBulk([]*ent.TableRowCreate{
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: "i1.png"}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: "i2.png"}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: "i1.png"}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: "i3.png"}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: "i2.png"}}),
		}...).Save(ctx)
		require.NoError(t, err)
		tb, err = db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
			tcq.Order(ent.Asc(tablecolumn.FieldID))
		}).Where(tablemeta.Name("table")).First(ctx)
		require.NoError(t, err)

		sc := &source.ListSource{Options: files}
		err = sc.Init(ctx, "")
		require.NoError(t, err)
		idx := source.NewIndexer(sc, &ent.TableColumn{
			Random: false,
		})
		generator := &AIRowsGenerator{
			autofill: AutofillRequest{
				Enable:         true,
				Columns:        []string{"c0"},
				ContextColumns: []string{"c1"},
			},
			sourceDataDir: "./",
			indexerMap: map[string]*source.Indexer{
				"c1": idx,
			},
			table:  tb,
			images: make(map[string]string),
		}
		err = generator.prepareRows(context.TODO(), 5)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"i1.png": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC", "i2.png": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC", "i3.png": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC"}, generator.images)
		for i, f := range files {
			require.Equal(t, map[string]*schema.CellValue{
				tb.Edges.Columns[0].Nanoid: {Value: "v"}, tb.Edges.Columns[1].Nanoid: {Value: f}, "__id__": {Value: rows[i].Nanoid},
			}, generator.rows[i])
		}
	})

	t.Run("autofill-url", func(t *testing.T) {
		files := []string{"i1.png", "i2.png", "i1.png", "i3.png", "i2.png"}
		for i, f := range files {
			files[i] = "https://images.com/" + f
		}
		db := db.NewTestDB()
		ctx := context.TODO()
		tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
		require.NoError(t, err)
		err = db.TableColumn.Create().
			SetName("c0").
			SetFillMode(tablecolumn.FillModeAi).
			SetTablemeta(tb).
			SetType(tablecolumn.TypeString).Exec(ctx)
		require.NoError(t, err)
		err = db.TableColumn.Create().
			SetName("c1").
			SetFillMode(tablecolumn.FillModeAi).
			SetTablemeta(tb).
			SetType(tablecolumn.TypeImage).Exec(ctx)
		require.NoError(t, err)
		rows, err := db.TableRow.CreateBulk([]*ent.TableRowCreate{
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: files[0]}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: files[1]}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: files[2]}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: files[3]}}),
			db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "v"}, {Value: files[4]}}),
		}...).Save(ctx)
		require.NoError(t, err)
		tb, err = db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
			tcq.Order(ent.Asc(tablecolumn.FieldID))
		}).Where(tablemeta.Name("table")).First(ctx)
		require.NoError(t, err)

		sc := &source.ListSource{Options: files}
		err = sc.Init(ctx, "")
		require.NoError(t, err)
		idx := source.NewIndexer(sc, &ent.TableColumn{
			Random: false,
		})
		generator := &AIRowsGenerator{
			autofill: AutofillRequest{
				Enable:         true,
				Columns:        []string{"c0"},
				ContextColumns: []string{"c1"},
			},
			sourceDataDir: "./",
			indexerMap: map[string]*source.Indexer{
				"c1": idx,
			},
			table:  tb,
			images: make(map[string]string),
		}
		err = generator.prepareRows(context.TODO(), 5)
		require.NoError(t, err)
		require.Equal(t, map[string]string{"https://images.com/i1.png": "https://images.com/i1.png", "https://images.com/i2.png": "https://images.com/i2.png", "https://images.com/i3.png": "https://images.com/i3.png"}, generator.images)
		for i, f := range files {
			require.Equal(t, map[string]*schema.CellValue{
				tb.Edges.Columns[0].Nanoid: {Value: "v"}, tb.Edges.Columns[1].Nanoid: {Value: f}, "__id__": {Value: rows[i].Nanoid},
			}, generator.rows[i])
		}
	})
}

func TestRowsGenerator_ImageColumn(t *testing.T) {
	ctx := context.TODO()
	db := db.NewTestDB()
	tb, err := db.TableMeta.Create().SetName("foo").SetDescription("bar").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetSource("so").
		SetTablemeta(tb).
		SetType(tablecolumn.TypeImage).Exec(ctx)
	require.NoError(t, err)
	generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
		Table: "foo",
		Batch: 5,
		Count: 5,
	}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
	require.Equal(t, 0, len(generator.missingColumns))
	require.Equal(t, 1, len(generator.missingImageColumns))
	require.Equal(t, generator.missingImageColumns[0].Name, "c1")
}

func TestRowsGenerator_ChatWithImage(t *testing.T) {
	ctx := context.Background()
	var messages []*client.Message
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			messages = request.Messages
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
		images:  map[string]string{"i.png": "i,png"},
	}
	_, err := generator.chat(ctx)
	require.NoError(t, err)
	p, err := generator.builder.Prompt()
	require.NoError(t, err)
	require.Equal(t, client.UserMessageWithImages(p, map[string]string{"i.png": "i,png"}), messages[0])
}

func TestRowsGenerator_GenerateImageWithText(t *testing.T) {
	defer func() { _ = os.RemoveAll("tablepilot_images") }()
	db := db.NewTestDB()
	ctx := context.Background()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	c1, err := db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	c2, err := db.TableColumn.Create().
		SetName("c2").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeImage).Save(ctx)
	require.NoError(t, err)
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			data := []map[string]any{
				{"__id__": 0, c1.Nanoid: "foobar"},
			}
			b, err := json.Marshal(map[string]any{"data": data})
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
			}, nil
		},
		ImageGenFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ImageGenResponse, error) {
			builder := promptbuilder.NewRowsBuilder(1)
			builder.AddDescription("")
			builder.AddTableColumns([]*ent.TableColumn{c1, c2}, false)
			builder.AddMissingColumns([]*ent.TableColumn{c2}, false)
			err = builder.AddExistings([]map[string]any{
				{c1.Nanoid: "foobar", "__id__": 0},
			})
			require.NoError(t, err)
			p, err := builder.ImageGenPrompt()
			require.NoError(t, err)
			require.Nil(t, request.Schema)
			require.Equal(t, []*client.Message{
				{Role: "user", Content: []client.Content{
					{Type: client.ContentTypeText, Data: p},
				}},
			}, request.Messages)
			id := fmt.Sprintf("%d-%s", 0, c2.Nanoid)
			return &client.ImageGenResponse{
				Images: map[string][]byte{id: []byte("bar")},
			}, nil
		},
	}

	generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
		Table: tb.Nanoid,
		Count: 1,
		Batch: 1,
	}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)
	v, err := generator.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(v))
	require.Equal(t, 1, len(aiService.ChatCalls()))
	require.Equal(t, 1, len(aiService.ImageGenCalls()))
	require.Equal(t, 3, len(v[0]))
	require.True(t, strings.HasPrefix(cast.ToString(v[0][c2.Nanoid].Value), "tablepilot_images/UkLWZg/0-gbHJdm-"))
	require.Equal(t, &schema.CellValue{
		Value: "foobar",
	}, v[0][c1.Nanoid])
	require.Equal(t, &schema.CellValue{Value: "UkLWZg"}, v[0]["__id__"])
}

func TestRowsGenerator_AutofillImageWithText(t *testing.T) {
	defer func() { _ = os.RemoveAll("tablepilot_images") }()
	db := db.NewTestDB()
	ctx := context.Background()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	c1, err := db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	c2, err := db.TableColumn.Create().
		SetName("c2").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeImage).Save(ctx)
	require.NoError(t, err)
	c3, err := db.TableColumn.Create().
		SetName("c3").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	row, err := db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: ""}, {Value: ""}, {Value: "bar"}}).Save(ctx)
	require.NoError(t, err)
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			data := []map[string]any{
				{"__id__": row.Nanoid, c1.Nanoid: "foobar"},
			}
			b, err := json.Marshal(map[string]any{"data": data})
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
			}, nil
		},
		ImageGenFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ImageGenResponse, error) {
			builder := promptbuilder.NewRowsBuilder(1)
			builder.AddDescription("")
			builder.AddTableColumns([]*ent.TableColumn{c1, c2, c3}, true)
			builder.AddMissingColumns([]*ent.TableColumn{c2}, false)
			err = builder.AddExistings([]map[string]any{
				{c1.Nanoid: "foobar", c3.Nanoid: "bar", "__id__": row.Nanoid},
			})
			require.NoError(t, err)
			p, err := builder.ImageGenPrompt()
			require.NoError(t, err)
			require.Nil(t, request.Schema)
			require.Equal(t, []*client.Message{
				{Role: "user", Content: []client.Content{
					{Type: client.ContentTypeText, Data: p},
				}},
			}, request.Messages)
			id := fmt.Sprintf("%s-%s", row.Nanoid, c2.Nanoid)
			return &client.ImageGenResponse{
				Images: map[string][]byte{id: []byte("bar")},
			}, nil
		},
	}

	generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
		Table: tb.Nanoid,
		Count: 1,
		Batch: 1,
		Autofill: AutofillRequest{
			Enable:         true,
			Columns:        []string{"c1", "c2"},
			ContextColumns: []string{"c3"},
		},
	}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)
	v, err := generator.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(v))
	require.Equal(t, 1, len(aiService.ChatCalls()))
	require.Equal(t, 1, len(aiService.ImageGenCalls()))
	require.Equal(t, 4, len(v[0]))
	require.True(t, strings.HasPrefix(cast.ToString(v[0][c2.Nanoid].Value), "tablepilot_images/UkLWZg/UkLWZg-gbHJdm-"))
	require.Equal(t, &schema.CellValue{
		Value: "foobar",
	}, v[0][c1.Nanoid])
	require.Equal(t, &schema.CellValue{Value: row.Nanoid}, v[0]["__id__"])
}

func TestRowsGenerator_AutofillImageOnly(t *testing.T) {
	defer func() { _ = os.RemoveAll("tablepilot_images") }()
	db := db.NewTestDB()
	ctx := context.Background()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	c1, err := db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	c2, err := db.TableColumn.Create().
		SetName("c2").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetType(tablecolumn.TypeImage).Save(ctx)
	require.NoError(t, err)
	row, err := db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{{Value: "bar"}, {Value: ""}}).Save(ctx)
	require.NoError(t, err)
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			data := []map[string]any{
				{"__id__": row.Nanoid, c1.Nanoid: "foobar"},
			}
			b, err := json.Marshal(map[string]any{"data": data})
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
			}, nil
		},
		ImageGenFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ImageGenResponse, error) {
			builder := promptbuilder.NewRowsBuilder(1)
			builder.AddDescription("")
			builder.AddTableColumns([]*ent.TableColumn{c1, c2}, true)
			builder.AddMissingColumns([]*ent.TableColumn{c2}, false)
			err = builder.AddExistings([]map[string]any{
				{c1.Nanoid: "bar", "__id__": row.Nanoid},
			})
			require.NoError(t, err)
			p, err := builder.ImageGenPrompt()
			require.NoError(t, err)
			require.Nil(t, request.Schema)
			require.Equal(t, []*client.Message{
				{Role: "user", Content: []client.Content{
					{Type: client.ContentTypeText, Data: p},
				}},
			}, request.Messages)
			id := fmt.Sprintf("%s-%s", row.Nanoid, c2.Nanoid)
			return &client.ImageGenResponse{
				Images: map[string][]byte{id: []byte("bar")},
			}, nil
		},
	}

	generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
		Table: tb.Nanoid,
		Count: 1,
		Batch: 1,
		Autofill: AutofillRequest{
			Enable:         true,
			Columns:        []string{"c2"},
			ContextColumns: []string{"c1"},
		},
	}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)
	v, err := generator.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(v))
	require.Equal(t, 0, len(aiService.ChatCalls()))
	require.Equal(t, 1, len(aiService.ImageGenCalls()))
	require.Equal(t, 3, len(v[0]))
	require.True(t, strings.HasPrefix(cast.ToString(v[0][c2.Nanoid].Value), "tablepilot_images/UkLWZg/UkLWZg-gbHJdm-"))
	require.Equal(t, &schema.CellValue{
		Value: "bar",
	}, v[0][c1.Nanoid])
	require.Equal(t, &schema.CellValue{Value: row.Nanoid}, v[0]["__id__"])
}

func TestRowsGenerator_PrepareContextRowsWithImage(t *testing.T) {
	defer func() { _ = os.RemoveAll("tablepilot_images") }()
	ctx := context.TODO()
	db := db.NewTestDB()
	tb, err := db.TableMeta.Create().SetName("foo").SetDescription("bar").Save(ctx)
	require.NoError(t, err)
	col1, err := db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetContextLength(1).
		SetType(tablecolumn.TypeImage).Save(ctx)
	require.NoError(t, err)
	col2, err := db.TableColumn.Create().
		SetName("c2").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetContextLength(0).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)

	counter := 0
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			imagePath := "fooo.png"
			if counter == 1 {
				row, err := tb.QueryRows().Order(ent.Desc("id")).First(ctx)
				require.NoError(t, err)
				imagePath = cast.ToString(row.Cells[0].Value)
				require.True(t, strings.HasPrefix(imagePath, "tablepilot_images/UkLWZg/0-UkLWZg"), imagePath)
			}
			builder := promptbuilder.NewRowsBuilder(1)
			builder.AddDescription("bar")
			err = builder.AddColumnContextData(col1.Nanoid, []any{imagePath})
			require.NoError(t, err)
			builder.AddTableColumns([]*ent.TableColumn{col1, col2}, false)
			// image will be generated in separate call(ImageGen)
			builder.AddMissingColumns([]*ent.TableColumn{col2}, true)
			p, err := builder.Prompt()
			require.NoError(t, err)
			require.Equal(t, []*client.Message{
				{Role: "user", Content: []client.Content{
					{Type: client.ContentTypeText, Data: p},
					{Type: client.ContentTypeText, Data: "\nBelow is the image with ID: " + fmt.Sprintf("<%s>", imagePath)},
					{Type: client.ContentTypeImage, Data: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC"},
				}},
			}, request.Messages)
			data := []map[string]any{
				{"__id__": "0", col2.Nanoid: "baz"},
			}
			b, err := json.Marshal(map[string]any{"data": data})
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
			}, nil
		},
		ImageGenFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ImageGenResponse, error) {
			imagePath := "fooo.png"
			if counter == 1 {
				row, err := tb.QueryRows().Order(ent.Desc("id")).First(ctx)
				require.NoError(t, err)
				imagePath = cast.ToString(row.Cells[0].Value)
				require.True(t, strings.HasPrefix(imagePath, "tablepilot_images/UkLWZg/0-UkLWZg"), imagePath)
			}
			builder := promptbuilder.NewRowsBuilder(1)
			builder.AddDescription("bar")
			err = builder.AddColumnContextData(col1.Nanoid, []any{imagePath})
			require.NoError(t, err)
			builder.AddTableColumns([]*ent.TableColumn{col1, col2}, false)
			builder.AddMissingColumns([]*ent.TableColumn{col1}, false)
			err = builder.AddExistings([]map[string]any{
				{col2.Nanoid: "baz", "__id__": "0"},
			})
			require.NoError(t, err)
			p, err := builder.ImageGenPrompt()
			require.NoError(t, err)
			require.Equal(t, []*client.Message{
				{Role: "user", Content: []client.Content{
					{Type: client.ContentTypeText, Data: p},
					{Type: client.ContentTypeText, Data: "\nBelow is the image with ID: " + fmt.Sprintf("<%s>", imagePath)},
					{Type: client.ContentTypeImage, Data: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC"},
				}},
			}, request.Messages)
			id := fmt.Sprintf("%s-%s", "0", col1.Nanoid)
			pb, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC")
			require.NoError(t, err)
			return &client.ImageGenResponse{
				Images: map[string][]byte{id: pb},
			}, nil
		},
	}
	generator, err := NewRowsGenerator(ctx, GenerateRowsRequest{
		Table: "foo",
		Batch: 1,
		Count: 2,
	}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)

	err = db.TableRow.Create().SetTablemeta(tb).SetCells(
		[]*schema.CellValue{{Value: "fooo.png"}, {Value: "v1"}},
	).Exec(ctx)
	err = createPNG("fooo.png")
	require.NoError(t, err)
	defer func() { _ = os.Remove("fooo.png") }()
	v, err := generator.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(v))
	require.Equal(t, 1, len(aiService.ChatCalls()))
	require.Equal(t, 1, len(aiService.ImageGenCalls()))

	counter += 1
	v, err = generator.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(v))
	require.Equal(t, 2, len(aiService.ChatCalls()))
	require.Equal(t, 2, len(aiService.ImageGenCalls()))
}
