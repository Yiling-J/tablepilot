package table

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/tablerow"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func requireColumnEqual(t *testing.T, expcted, column *ent.TableColumn) {
	require.Equal(t, expcted.Name, column.Name)
	require.Equal(t, expcted.Description, column.Description)
	require.Equal(t, expcted.Type, column.Type)
	require.Equal(t, expcted.FillMode, column.FillMode)
	if len(expcted.Source) > 0 {
		require.JSONEq(t, string(expcted.Source), string(column.Source))
	} else {
		require.Equal(t, expcted.Source, column.Source)
	}
	require.Equal(t, expcted.ContextLength, column.ContextLength)
}

func fromCells(data []*schema.CellValue) []any {
	cells := []any{}
	for _, c := range data {
		cells = append(cells, c.Value)
	}
	return cells
}

func toCells(data []any) []*schema.CellValue {
	cells := []*schema.CellValue{}
	for _, c := range data {
		cells = append(cells, &schema.CellValue{Value: c})
	}
	return cells
}

func TestTableService_CreateTable(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	columns := []TableGenColumn{
		{
			Name: "name", Description: "recipe name", Type: "string",
			FillMode: "ai", ContextLength: 5,
		},
		{
			Name: "count", Description: "recipe count", Type: "integer",
			FillMode: "ai", ContextLength: 3,
		},
		{
			Name: "tag", Description: "recipe tag", Type: "array",
			FillMode: "pick", Source: "tags",
		},
		{
			Name: "country", Description: "recipe country", Type: "string",
			FillMode: "pick", Source: "countries",
		},
		{
			Name: "user", Description: "recipe user", Type: "boolean",
			FillMode: "pick", Source: "users",
		},
		{},
	}
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			require.Equal(t, "user", request.Messages[0].Role)
			require.Equal(t, request.Model, "aiai")
			columnsGenBuilder := promptbuilder.NewColumnsBuilder(1, "test", "test table")
			bc := []promptbuilder.Column{}
			for _, c := range columns {
				if c.Name == "" {
					continue
				}
				bc = append(bc, promptbuilder.Column{
					Name:        c.Name,
					Description: c.Description,
				})
			}
			columnsGenBuilder.AddExistingColumns(bc)
			prompt, err := columnsGenBuilder.Prompt()
			require.NoError(t, err)
			require.Equal(t, prompt, request.Messages[0].Content)
			return &client.ChatResponse{
				Content: `[{"name":"extra","type":"string"},{"name":"extra2","type":"string"}]`,
				Tokens:  100,
			}, nil
		},
	}
	srv := NewTableService(&config.Config{}, db, aiService, zap.NewNop().Sugar())

	sources := []json.RawMessage{
		[]byte(`{
      "name": "countries",
      "type": "list",
      "options": ["China", "Japan", "England", "Thai", "France"],
      "random": true,
      "replacement": true
    }`),
		[]byte(`
    {
      "name": "tags",
      "type": "ai",
      "prompt": "Generate 20 tags.",
      "random": true
    }`),
		[]byte(`
    {
      "name": "users",
      "type": "linked",
      "table": "user",
      "column": "name",
      "context_columns": ["age"],
      "filters": { "must": [{ "name": "a" }, { "age": 12 }, { "should": [] }] },
      "sorts": [{ "column": "name", "desc": true }],
      "random": true,
      "replacement": true
    }
`),
	}

	userTable, err := db.TableMeta.Create().SetName("user").Save(ctx)
	require.NoError(t, err)
	userColumns, err := db.TableColumn.CreateBulk(
		db.TableColumn.Create().SetTablemeta(userTable).SetName("name").SetType(
			tablecolumn.TypeString,
		).SetFillMode(tablecolumn.FillModeAi),
		db.TableColumn.Create().SetTablemeta(userTable).SetName("age").SetType(
			tablecolumn.TypeInteger,
		).SetFillMode(tablecolumn.FillModeAi),
	).Save(ctx)
	require.NoError(t, err)

	savedSources := []json.RawMessage{
		[]byte(`{
      "type": "list",
      "options": ["China", "Japan", "England", "Thai", "France"],
      "random": true,
      "replacement": true
    }`),
		[]byte(`
    {
      "type": "ai",
      "prompt": "Generate 20 tags.",
      "random": true,
      "options": null
    }`),
		[]byte(fmt.Sprintf(`
    {
      "type": "linked",
      "table": "%s",
      "column": "%s",
      "context_columns": ["%s"],
      "random": true,
      "replacement": true
    }
`, userTable.Nanoid, userColumns[0].Nanoid, userColumns[1].Nanoid)),
	}

	id, err := srv.CreateTable(ctx, &TableGenRequest{
		Name:        "test",
		Description: "test table",
		Columns:     columns,
		Sources:     sources,
		Model:       "aiai",
	})
	require.NoError(t, err)
	table, err := db.TableMeta.Query().WithColumns().Where(tablemeta.Nanoid(id)).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, "test", table.Name)
	require.Equal(t, "test table", table.Description)
	require.Equal(t, id, table.Nanoid)
	require.Equal(t, 6, len(table.Edges.Columns))
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "name", Description: "recipe name", ContextLength: 5,
			Type: tablecolumn.TypeString, FillMode: tablecolumn.FillModeAi, Source: nil,
		},
		table.Edges.Columns[0],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "count", Description: "recipe count", ContextLength: 3,
			Type: tablecolumn.TypeInteger, FillMode: tablecolumn.FillModeAi, Source: nil,
		},
		table.Edges.Columns[1],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "tag", Description: "recipe tag", ContextLength: 0,
			Type: tablecolumn.TypeArray, FillMode: tablecolumn.FillModePick,
			Source: savedSources[1],
		},
		table.Edges.Columns[2],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "country", Description: "recipe country", ContextLength: 0,
			Type: tablecolumn.TypeString, FillMode: tablecolumn.FillModePick,
			Source: savedSources[0],
		},
		table.Edges.Columns[3],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "user", Description: "recipe user", ContextLength: 0,
			Type: tablecolumn.TypeBoolean, FillMode: tablecolumn.FillModePick,
			Source: savedSources[2],
		},
		table.Edges.Columns[4],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "extra", Description: "", ContextLength: 0,
			Type: tablecolumn.TypeString, FillMode: tablecolumn.FillModeAi,
		},
		table.Edges.Columns[5],
	)
}

func TestTableService_LinkedContextRow(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	tb, err := db.TableMeta.Create().SetName("table").SetDescription("bar").Save(ctx)
	require.NoError(t, err)
	c, err := db.TableColumn.Create().
		SetName("c").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetContextLength(1).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	tb, err = db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).Where(tablemeta.Nanoid(tb.Nanoid)).First(ctx)
	require.NoError(t, err)
	err = db.TableRow.CreateBulk(
		db.TableRow.Create().SetCells([]*schema.CellValue{{
			Value:        "foo",
			ContextValue: map[string]any{"bar": 1, "go": 2},
		}},
		).SetTablemeta(tb),
		db.TableRow.Create().SetCells([]*schema.CellValue{{
			Value:        "bar",
			ContextValue: map[string]any{"bar": 3, "go": 4},
		}},
		).SetTablemeta(tb),
	).Exec(ctx)
	require.NoError(t, err)

	generator := &AIRowsGenerator{
		table:         tb,
		contextLength: 1,
	}
	err = generator.newBatch(ctx, 1)
	require.NoError(t, err)
	pm, err := generator.builder.Prompt()
	require.NoError(t, err)
	builder := promptbuilder.NewRowsBuilder(1)
	builder.AddDescription("bar")
	err = builder.AddColumnContextData(c.Nanoid, []any{
		`{"bar":3,"go":4}`,
	})
	require.NoError(t, err)
	p, err := builder.Prompt()
	require.NoError(t, err)
	require.Equal(t, p, pm)
}

func TestTableService_Rows(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb).
		SetContextLength(3).
		SetType(tablecolumn.TypeString).Exec(ctx)
	require.NoError(t, err)
	err = db.TableRow.CreateBulk(
		db.TableRow.Create().SetCells(toCells([]any{"1"})).SetTablemeta(tb),
		db.TableRow.Create().SetCells(toCells([]any{"2"})).SetTablemeta(tb),
		db.TableRow.Create().SetCells(toCells([]any{"3"})).SetTablemeta(tb),
		db.TableRow.Create().SetCells(toCells([]any{"4"})).SetTablemeta(tb),
	).Exec(ctx)
	require.NoError(t, err)
	srv := NewTableService(
		&config.Config{}, db, nil, zap.NewNop().Sugar(),
	)
	rows, err := srv.Rows(ctx, "table")
	require.NoError(t, err)
	require.Equal(t, "c1", rows.Columns[0].Name)
	require.Equal(t, 4, len(rows.Rows))
	require.Equal(t, []any{"1"}, fromCells(rows.Rows[0].Cells))
	require.Equal(t, []any{"2"}, fromCells(rows.Rows[1].Cells))
	require.Equal(t, []any{"3"}, fromCells(rows.Rows[2].Cells))
	require.Equal(t, []any{"4"}, fromCells(rows.Rows[3].Cells))
}

func TestTableService_Truncate(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	tb1, err := db.TableMeta.Create().SetName("table1").Save(ctx)
	require.NoError(t, err)
	err = db.TableRow.CreateBulk(
		db.TableRow.Create().SetCells(toCells([]any{"1"})).SetTablemeta(tb1),
	).Exec(ctx)
	require.NoError(t, err)
	tb2, err := db.TableMeta.Create().SetName("table2").Save(ctx)
	require.NoError(t, err)
	err = db.TableRow.CreateBulk(
		db.TableRow.Create().SetCells(toCells([]any{"1"})).SetTablemeta(tb2),
	).Exec(ctx)
	require.NoError(t, err)

	srv := NewTableService(
		&config.Config{}, db, nil, zap.NewNop().Sugar(),
	)
	count, err := srv.Truncate(ctx, "table1")
	require.NoError(t, err)
	require.Equal(t, 1, count)
	c, err := tb1.QueryRows().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, c)
	c, err = tb2.QueryRows().Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, c)
}

func TestTableService_Delete(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	tb1, err := db.TableMeta.Create().SetName("table1").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb1).
		SetContextLength(3).
		SetType(tablecolumn.TypeString).Exec(ctx)
	require.NoError(t, err)
	err = db.TableRow.CreateBulk(
		db.TableRow.Create().SetCells(toCells([]any{"1"})).SetTablemeta(tb1),
	).Exec(ctx)
	require.NoError(t, err)
	tb2, err := db.TableMeta.Create().SetName("table2").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb2).
		SetContextLength(3).
		SetType(tablecolumn.TypeString).Exec(ctx)
	require.NoError(t, err)
	err = db.TableRow.CreateBulk(
		db.TableRow.Create().SetCells(toCells([]any{"1"})).SetTablemeta(tb2),
	).Exec(ctx)
	require.NoError(t, err)

	srv := NewTableService(
		&config.Config{}, db, nil, zap.NewNop().Sugar(),
	)
	count, err := srv.Delete(ctx, "table1")
	require.NoError(t, err)
	require.Equal(t, 1, count)
	n, err := db.TableMeta.Query().Where(tablemeta.Nanoid(tb1.Nanoid)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)
	n, err = db.TableColumn.Query().Where(tablecolumn.HasTablemetaWith(tablemeta.Nanoid(tb1.Nanoid))).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)
	n, err = db.TableRow.Query().Where(tablerow.HasTablemetaWith(tablemeta.Nanoid(tb1.Nanoid))).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, n)
	n, err = db.TableMeta.Query().Where(tablemeta.Nanoid(tb2.Nanoid)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	n, err = db.TableColumn.Query().Where(tablecolumn.HasTablemetaWith(tablemeta.Nanoid(tb2.Nanoid))).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	n, err = db.TableRow.Query().Where(tablerow.HasTablemetaWith(tablemeta.Nanoid(tb2.Nanoid))).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, n)
}

func TestTableService_Import(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	srv := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	records := [][]string{
		{"col1", "col2"},
		{"a", "1"},
		{"b", "2"},
	}

	buffer := bytes.NewBuffer([]byte(""))
	w := csv.NewWriter(buffer)
	for _, record := range records {
		err := w.Write(record)
		require.NoError(t, err)
	}
	w.Flush()

	id, err := srv.Import(ctx, "foo", strings.NewReader(buffer.String()))
	require.NoError(t, err)

	table, err := db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).WithRows(func(trq *ent.TableRowQuery) {
		trq.Order(ent.Asc(tablerow.FieldID))
	}).Where(tablemeta.Nanoid(id)).First(ctx)
	require.NoError(t, err)
	columnNames := []string{}
	for _, col := range table.Edges.Columns {
		require.Equal(t, tablecolumn.TypeString, col.Type)
		columnNames = append(columnNames, col.Name)
	}
	require.Equal(t, []string{"col1", "col2"}, columnNames)
	rows := [][]any{}
	for _, row := range table.Edges.Rows {
		r := []any{}
		for _, cell := range row.Cells {
			r = append(r, cell.Value)
		}
		rows = append(rows, r)
	}
	require.Equal(t, [][]any{{"a", "1"}, {"b", "2"}}, rows)
}

func TestTableService_ListTables(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	srv := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	tb1, err := db.TableMeta.Create().SetName("t1").SetDescription("tt1").SetModel("m1").Save(ctx)
	require.NoError(t, err)
	col1, err := db.TableColumn.Create().
		SetName("c1").SetDescription("cc1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb1).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	tb2, err := db.TableMeta.Create().SetName("t2").SetDescription("tt2").SetModel("m2").Save(ctx)
	require.NoError(t, err)
	col2, err := db.TableColumn.Create().
		SetName("c2").SetDescription("cc2").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb2).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)

	resp, err := srv.ListTables(ctx)
	require.NoError(t, err)
	require.Equal(t, &ListTablesResponse{
		Total: 2,
		Tables: []TableInfo{
			{ID: tb2.Nanoid, Name: "t2", Description: "tt2", Model: "m2", Columns: []TableColumnInfo{
				{ID: col2.Nanoid, Name: "c2", Description: "cc2", Type: "string", FillMode: "ai"},
			}},
			{ID: tb1.Nanoid, Name: "t1", Description: "tt1", Model: "m1", Columns: []TableColumnInfo{
				{ID: col1.Nanoid, Name: "c1", Description: "cc1", Type: "string", FillMode: "ai"},
			}},
		},
	}, resp)
}

func TestTableService_GetTableDetail(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	srv := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	tb1, err := db.TableMeta.Create().SetName("t1").SetDescription("tt1").SetModel("m1").Save(ctx)
	require.NoError(t, err)
	col1, err := db.TableColumn.Create().
		SetName("c1").SetDescription("cc1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(tb1).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)

	resp, err := srv.GetTableDetail(ctx, "t1")
	require.NoError(t, err)
	require.Equal(t,
		&TableInfo{ID: tb1.Nanoid, Name: "t1", Description: "tt1", Model: "m1", Columns: []TableColumnInfo{
			{ID: col1.Nanoid, Name: "c1", Description: "cc1", Type: "string", FillMode: "ai"},
		}}, resp)
}
