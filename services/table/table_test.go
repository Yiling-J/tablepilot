package table

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
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
	require.Equal(t, expcted.Source, column.Source)
	if column.Source != "" {
		require.Equal(t, expcted.Random, column.Random)
		require.Equal(t, expcted.Replacement, column.Replacement)
		require.Equal(t, expcted.Repeat, column.Repeat)
		require.Equal(t, expcted.LinkedColumn, column.LinkedColumn)
		require.Equal(t, expcted.LinkedContextColumns, column.LinkedContextColumns)
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

func TestTableService_Create(t *testing.T) {
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
			FillMode: "pick", Source: "tags", Random: true, Replacement: true, Repeat: 3,
		},
		{
			Name: "country", Description: "recipe country", Type: "string",
			FillMode: "pick", Source: "countries",
		},
		{
			Name: "user", Description: "recipe user", Type: "boolean",
			FillMode: "pick", Source: "users", LinkedColumn: "name", LinkedContextColumns: []string{"age"},
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
	srv, err := NewTableService(&config.Config{}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)

	sources := []json.RawMessage{
		[]byte(`{
      "name": "countries",
      "type": "list",
      "options": ["China", "Japan", "England", "Thai", "France"]
    }`),
		[]byte(`
    {
      "name": "tags",
      "type": "ai",
      "prompt": "Generate 20 tags."
    }`),
		[]byte(`
    {
      "name": "users",
      "type": "linked",
      "table": "user",
      "filters": { "must": [{ "name": "a" }, { "age": 12 }, { "should": [] }] },
      "sorts": [{ "column": "name", "desc": true }]
    }
`),
	}

	userTable, err := db.TableMeta.Create().SetName("user").Save(ctx)
	require.NoError(t, err)
	_, err = db.TableColumn.CreateBulk(
		db.TableColumn.Create().SetTablemeta(userTable).SetName("name").SetType(
			tablecolumn.TypeString,
		).SetFillMode(tablecolumn.FillModeAi),
		db.TableColumn.Create().SetTablemeta(userTable).SetName("age").SetType(
			tablecolumn.TypeInteger,
		).SetFillMode(tablecolumn.FillModeAi),
	).Save(ctx)
	require.NoError(t, err)

	savedSources := map[string]json.RawMessage{
		"countries": []byte(`{"type":"list","options":["China","Japan","England","Thai","France"]}`),
		"tags":      []byte(`{"type":"ai","prompt":"Generate 20 tags.","options":null}`),
		"users":     []byte(fmt.Sprintf(`{"type":"linked","table":"%s"}`, userTable.Nanoid)),
	}

	id, err := srv.Create(ctx, &TableGenRequest{
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
	require.Equal(t, savedSources, table.Sources)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "name", Description: "recipe name", ContextLength: 5,
			Type: tablecolumn.TypeString, FillMode: tablecolumn.FillModeAi, Source: "",
		},
		table.Edges.Columns[0],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "count", Description: "recipe count", ContextLength: 3,
			Type: tablecolumn.TypeInteger, FillMode: tablecolumn.FillModeAi, Source: "",
		},
		table.Edges.Columns[1],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "tag", Description: "recipe tag", ContextLength: 0,
			Type: tablecolumn.TypeArray, FillMode: tablecolumn.FillModePick,
			Source: "tags", Random: true, Replacement: true, Repeat: 3,
		},
		table.Edges.Columns[2],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "country", Description: "recipe country", ContextLength: 0,
			Type: tablecolumn.TypeString, FillMode: tablecolumn.FillModePick,
			Source: "countries",
		},
		table.Edges.Columns[3],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "user", Description: "recipe user", ContextLength: 0,
			Type: tablecolumn.TypeBoolean, FillMode: tablecolumn.FillModePick,
			Source: "users", LinkedColumn: "name", LinkedContextColumns: []string{"age"},
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

func TestTableService_CreateTableSharedSource(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	columns := []TableGenColumn{
		{
			Name: "user", Description: "recipe user", Type: "boolean",
			FillMode: "pick", Source: "users", LinkedColumn: "name", LinkedContextColumns: []string{"age"},
		},
	}
	aiService := &ai.AiServiceMock{
		ChatFunc: func(
			ctx context.Context, request *client.ChatRequest,
		) (*client.ChatResponse, error) {
			return &client.ChatResponse{
				Content: `[{"name":"extra","type":"string"},{"name":"extra2","type":"string"}]`,
				Tokens:  100,
			}, nil
		},
	}
	userTable, err := db.TableMeta.Create().SetName("user").Save(ctx)
	require.NoError(t, err)
	srv, err := NewTableService(&config.Config{
		Sources: []map[string]any{{
			"name":  "users",
			"type":  "linked",
			"table": "user",
		}},
	}, db, aiService, zap.NewNop().Sugar())
	require.NoError(t, err)

	_, err = db.TableColumn.CreateBulk(
		db.TableColumn.Create().SetTablemeta(userTable).SetName("name").SetType(
			tablecolumn.TypeString,
		).SetFillMode(tablecolumn.FillModeAi),
		db.TableColumn.Create().SetTablemeta(userTable).SetName("age").SetType(
			tablecolumn.TypeInteger,
		).SetFillMode(tablecolumn.FillModeAi),
	).Save(ctx)
	require.NoError(t, err)

	id, err := srv.Create(ctx, &TableGenRequest{
		Name:        "test",
		Description: "test table",
		Columns:     columns,
		Model:       "aiai",
	})
	require.NoError(t, err)
	table, err := db.TableMeta.Query().WithColumns().Where(tablemeta.Nanoid(id)).Only(ctx)
	require.NoError(t, err)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "user", Description: "recipe user", ContextLength: 0,
			Type: tablecolumn.TypeBoolean, FillMode: tablecolumn.FillModePick,
			Source: "users", LinkedColumn: "name", LinkedContextColumns: []string{"age"},
		},
		table.Edges.Columns[0],
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
		table:          tb,
		contextLength:  1,
		contextColumns: []*ent.TableColumn{c},
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
	srv, err := NewTableService(
		&config.Config{}, db, nil, zap.NewNop().Sugar(),
	)
	require.NoError(t, err)
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

	srv, err := NewTableService(
		&config.Config{}, db, nil, zap.NewNop().Sugar(),
	)
	require.NoError(t, err)
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

	srv, err := NewTableService(
		&config.Config{}, db, nil, zap.NewNop().Sugar(),
	)
	require.NoError(t, err)
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
	srv, err := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
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

	// import some data again, test table exists case
	records = [][]string{
		{"col2", "col4", "col1"},
		{"3", "n1", "c"},
		{"4", "n2", "d"},
	}

	buffer = bytes.NewBuffer([]byte(""))
	w = csv.NewWriter(buffer)
	for _, record := range records {
		err := w.Write(record)
		require.NoError(t, err)
	}
	w.Flush()

	id, err = srv.Import(ctx, "foo", strings.NewReader(buffer.String()))
	require.NoError(t, err)
	require.Equal(t, table.Nanoid, id)
	table, err = db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).WithRows(func(trq *ent.TableRowQuery) {
		trq.Order(ent.Asc(tablerow.FieldID))
	}).Where(tablemeta.Nanoid(id)).First(ctx)
	require.NoError(t, err)
	rows = [][]any{}
	for _, row := range table.Edges.Rows {
		r := []any{}
		for _, cell := range row.Cells {
			r = append(r, cell.Value)
		}
		rows = append(rows, r)
	}
	require.Equal(t, [][]any{{"a", "1"}, {"b", "2"}, {"c", "3"}, {"d", "4"}}, rows)
}

func TestTableService_ImportAutoType(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()

	tb, err := db.TableMeta.Create().SetName("foo").Save(ctx)
	require.NoError(t, err)
	creates := []*ent.TableColumnCreate{
		db.TableColumn.Create().SetName("int").SetTablemeta(tb).SetFillMode(tablecolumn.FillModeAi).SetType(tablecolumn.TypeInteger),
		db.TableColumn.Create().SetName("string").SetTablemeta(tb).SetFillMode(tablecolumn.FillModeAi).SetType(tablecolumn.TypeString),
		db.TableColumn.Create().SetName("number").SetTablemeta(tb).SetFillMode(tablecolumn.FillModeAi).SetType(tablecolumn.TypeNumber),
		db.TableColumn.Create().SetName("array").SetTablemeta(tb).SetFillMode(tablecolumn.FillModeAi).SetType(tablecolumn.TypeArray),
	}
	err = db.TableColumn.CreateBulk(creates...).Exec(ctx)
	require.NoError(t, err)

	srv, err := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
	records := [][]string{
		{"int", "string", "number", "array"},
		{"1", "2", "3.2", `["a", "b"]`},
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
		columnNames = append(columnNames, col.Name)
	}
	require.Equal(t, []string{"int", "string", "number", "array"}, columnNames)
	rows := [][]any{}
	for _, row := range table.Edges.Rows {
		r := []any{}
		for _, cell := range row.Cells {
			r = append(r, cell.Value)
		}
		rows = append(rows, r)
	}
	require.Equal(t, [][]any{{1.0, "2", 3.2, []any{"a", "b"}}}, rows)
}

func TestTableService_ListTables(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	srv, err := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
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
	srv, err := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
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

func TestTableService_CreateRows(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	srv, err := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
	tb, err := db.TableMeta.Create().SetName("t1").Save(ctx)
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
		SetType(tablecolumn.TypeInteger).Save(ctx)
	require.NoError(t, err)
	err = srv.CreateRows(ctx, "t1", []map[string]any{
		{"c1": "v1", "c2": 1},
		{col1.Nanoid: "v2", col2.Nanoid: 2},
		{"c1": "v3"},
	})
	require.NoError(t, err)
	rows, err := tb.QueryRows().All(ctx)
	require.NoError(t, err)
	expected := [][]any{
		{"v1", float64(1)},
		{"v2", float64(2)},
		{"v3", float64(0)},
	}
	data := [][]any{}
	for _, row := range rows {
		r := []any{}
		for _, c := range row.Cells {
			r = append(r, c.Value)
		}
		data = append(data, r)
	}
	require.Equal(t, expected, data)
}

func TestTableService_NewServiceSharedSource(t *testing.T) {
	tmpFile, err := os.CreateTemp("./", "test_*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	writer := csv.NewWriter(tmpFile)
	require.NoError(t, writer.Write([]string{"Name", "Job", "Age"}))
	require.NoError(t, writer.Write([]string{"me", "Engineer", "1"}))
	require.NoError(t, writer.Write([]string{"you", "Doctor", "2"}))
	writer.Flush()
	require.NoError(t, writer.Error())
	require.NoError(t, tmpFile.Close())

	db := db.NewTestDB()
	srv, err := NewTableService(&config.Config{Sources: []map[string]any{
		{"name": "s1", "type": "list", "options": []string{"a", "b"}},
		{"name": "s2", "type": "csv", "paths": []string{strings.TrimPrefix(tmpFile.Name(), "./")}},
	}, Common: config.Common{SourceDataDir: "./"}}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
	require.ElementsMatch(t, []*SharedSource{
		{Name: "s1", Columns: nil, Data: json.RawMessage(`{"name":"s1","options":["a","b"],"type":"list"}`)},
		{Name: "s2", Columns: []string{"Name", "Job", "Age"}, Data: json.RawMessage(
			fmt.Sprintf(`{"name":"s2","paths":["%s"],"type":"csv"}`, strings.TrimPrefix(tmpFile.Name(), "./")),
		)},
	}, srv.sharedSources)
}

func TestTableService_CreateTableAPIRequest(t *testing.T) {
	for _, tc := range []struct {
		source string
		error  string
	}{
		{`{"name":"so","type":"list","file":"go.txt"}`, "file field for list source is only allowed in CLI"},
		{`{"name":"so","type":"csv","paths":["z.csv"]}`, "paths field for csv source is only allowed in CLI"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			db := db.NewTestDB()
			ctx := context.Background()
			columns := []TableGenColumn{
				{
					Name: "user", Description: "recipe user", Type: "boolean",
					FillMode: "ai",
				},
			}
			aiService := &ai.AiServiceMock{
				ChatFunc: func(
					ctx context.Context, request *client.ChatRequest,
				) (*client.ChatResponse, error) {
					return &client.ChatResponse{
						Content: `[{"name":"extra","type":"string"},{"name":"extra2","type":"string"}]`,
						Tokens:  100,
					}, nil
				},
			}
			userTable, err := db.TableMeta.Create().SetName("user").Save(ctx)
			require.NoError(t, err)
			srv, err := NewTableService(&config.Config{}, db, aiService, zap.NewNop().Sugar())
			require.NoError(t, err)

			_, err = db.TableColumn.CreateBulk(
				db.TableColumn.Create().SetTablemeta(userTable).SetName("name").SetType(
					tablecolumn.TypeString,
				).SetFillMode(tablecolumn.FillModeAi),
				db.TableColumn.Create().SetTablemeta(userTable).SetName("age").SetType(
					tablecolumn.TypeInteger,
				).SetFillMode(tablecolumn.FillModeAi),
			).Save(ctx)
			require.NoError(t, err)

			_, err = srv.Create(ctx, &TableGenRequest{
				Name:        "test",
				Description: "test table",
				Columns:     columns,
				Sources: []json.RawMessage{
					[]byte(tc.source),
				},
				Model:      "aiai",
				apiRequest: true,
			})
			require.Equal(t, tc.error, err.Error())
		})
	}
}

func TestTableService_ImportLinked(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	srv, err := NewTableService(&config.Config{Common: config.Common{SourceDataDir: "./"}}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	tid, err := srv.Create(ctx, &TableGenRequest{Name: "lt", Columns: []TableGenColumn{
		{Name: "c1", FillMode: "ai", Type: "string", Description: "c1c1"},
		{Name: "c2", FillMode: "ai", Type: "string", Description: "c2c2"},
	}})
	require.NoError(t, err)
	err = srv.CreateRows(ctx, "lt", []map[string]any{
		{"c1": "aa", "c2": "foo"},
		{"c1": "bb", "c2": "bar"},
	})
	require.NoError(t, err)

	tmpFile, err := os.CreateTemp("./", "test_*.csv")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	writer := csv.NewWriter(tmpFile)
	require.NoError(t, writer.Write([]string{"c1", "c2"}))
	require.NoError(t, writer.Write([]string{"a", "v1"}))
	require.NoError(t, writer.Write([]string{"b", "v2"}))
	writer.Flush()
	require.NoError(t, writer.Error())
	require.NoError(t, tmpFile.Close())

	tb, err := db.TableMeta.Create().SetName("t1").SetSources(map[string]json.RawMessage{
		"s1": []byte(`{"type":"csv","paths":["test_*.csv"]}`),
		"s2": []byte(fmt.Sprintf(`{"type":"linked","table":"%s"}`, tid)),
	}).Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("col1").
		SetFillMode(tablecolumn.FillModePick).
		SetTablemeta(tb).
		SetSource("s1").
		SetLinkedColumn("c1").SetLinkedContextColumns([]string{"c1", "c2"}).
		SetType(tablecolumn.TypeString).Exec(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("col2").
		SetFillMode(tablecolumn.FillModePick).
		SetTablemeta(tb).
		SetSource("s2").
		SetLinkedColumn("c1").SetLinkedContextColumns([]string{"c1", "c2"}).
		SetType(tablecolumn.TypeString).Exec(ctx)
	require.NoError(t, err)
	records := [][]string{
		{"col1", "col2"},
		{"a", "aa"},
		{"b", "bb"},
		{"c", "cc"},
	}

	buffer := bytes.NewBuffer([]byte(""))
	w := csv.NewWriter(buffer)
	for _, record := range records {
		err := w.Write(record)
		require.NoError(t, err)
	}
	w.Flush()

	id, err := srv.Import(ctx, "t1", strings.NewReader(buffer.String()))
	require.NoError(t, err)
	table, err := db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).WithRows(func(trq *ent.TableRowQuery) {
		trq.Order(ent.Asc(tablerow.FieldID))
	}).Where(tablemeta.Nanoid(id)).First(ctx)
	require.NoError(t, err)
	rows := [][]*schema.CellValue{}
	for _, row := range table.Edges.Rows {
		rows = append(rows, row.Cells)
		for _, cell := range row.Cells {
			fmt.Println(cell.Value, cell.ContextValue)
		}
	}
	require.Equal(t, [][]*schema.CellValue{
		{&schema.CellValue{Value: "a", ContextValue: map[string]any{"c1": "a", "c2": "v1"}}, &schema.CellValue{Value: "aa", ContextValue: map[string]any{
			"c1": map[string]any{"data": "aa", "description": "c1c1"},
			"c2": map[string]any{"data": "foo", "description": "c2c2"},
		}}},
		{&schema.CellValue{Value: "b", ContextValue: map[string]any{"c1": "b", "c2": "v2"}}, &schema.CellValue{Value: "bb", ContextValue: map[string]any{
			"c1": map[string]any{"data": "bb", "description": "c1c1"},
			"c2": map[string]any{"data": "bar", "description": "c2c2"},
		}}},
		{&schema.CellValue{Value: "c", ContextValue: nil}, &schema.CellValue{Value: "cc", ContextValue: nil}},
	}, rows)
}

func TestTableService_Update(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	columns := []TableGenColumn{
		{
			Name: "name", Description: "recipe name", Type: "string",
			FillMode: "ai", ContextLength: 5,
		},
		{
			Name: "description", Description: "recipe description", Type: "string",
			FillMode: "ai", ContextLength: 5,
		},
		{
			Name: "steps", Description: "recipe steps", Type: "string",
			FillMode: "ai", ContextLength: 3,
		},
		{
			Name: "ingredients", Description: "recipe ingredients", Type: "string",
			FillMode: "ai",
		},
	}
	srv, err := NewTableService(&config.Config{}, db, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	id, err := srv.Create(ctx, &TableGenRequest{
		Name:        "test",
		Description: "test table",
		Columns:     columns,
		Model:       "aiai",
	})
	require.NoError(t, err)
	detail, err := srv.GetTableDetail(ctx, id)
	require.NoError(t, err)
	require.Equal(t, []TableColumnInfo{
		{ID: "UkLWZg", Name: "name", Description: "recipe name", Type: "string", FillMode: "ai"},
		{ID: "gbHJdm", Name: "description", Description: "recipe description", Type: "string", FillMode: "ai"},
		{ID: "EfhxLZ", Name: "steps", Description: "recipe steps", Type: "string", FillMode: "ai"},
		{ID: "VqXmZF", Name: "ingredients", Description: "recipe ingredients", Type: "string", FillMode: "ai"},
	}, detail.Columns)
	err = srv.CreateRows(ctx, id, []map[string]any{
		{"name": "r1", "description": "d1", "steps": "s1", "ingredients": "i1"},
		{"name": "r2", "description": "d2", "steps": "s2", "ingredients": "i2"},
	})
	require.NoError(t, err)

	// change order, remove one column and add two columns
	columns = []TableGenColumn{
		{
			Name: "steps", Description: "recipe steps", Type: "string",
			FillMode: "ai", ContextLength: 3,
		},
		{
			Name: "ingredients", Description: "recipe ingredients", Type: "string",
			FillMode: "ai",
		},
		{
			Name: "name", Description: "recipe name", Type: "string",
			FillMode: "ai", ContextLength: 5,
		},
		{
			Name: "tags", Description: "recipe tags", Type: "string",
			FillMode: "ai",
		},
		{
			Name: "difficulty", Description: "recipe difficulty", Type: "integer",
			FillMode: "ai",
		},
	}
	id, err = srv.Update(ctx, "test", &TableGenRequest{
		Name:        "test_go",
		Description: "test table go",
		Columns:     columns,
	})
	require.NoError(t, err)
	detail, err = srv.GetTableDetail(ctx, id)
	require.NoError(t, err)
	require.Equal(t, []TableColumnInfo{
		{ID: "UkLWZg", Name: "name", Description: "recipe name", Type: "string", FillMode: "ai"},
		{ID: "EfhxLZ", Name: "steps", Description: "recipe steps", Type: "string", FillMode: "ai"},
		{ID: "VqXmZF", Name: "ingredients", Description: "recipe ingredients", Type: "string", FillMode: "ai"},
		{ID: "p6klVe", Name: "tags", Description: "recipe tags", Type: "string", FillMode: "ai"},
		{ID: "nJqfPa", Name: "difficulty", Description: "recipe difficulty", Type: "integer", FillMode: "ai"},
	}, detail.Columns)

	dbrows, err := db.TableRow.Query().Where(tablerow.HasTablemetaWith(tablemeta.Nanoid(id))).Order(ent.Asc("id")).All(ctx)
	require.NoError(t, err)
	rows := [][]*schema.CellValue{}
	for _, row := range dbrows {
		rows = append(rows, row.Cells)
	}
	require.Equal(t, [][]*schema.CellValue{
		{
			&schema.CellValue{Value: "r1"}, &schema.CellValue{Value: "s1"},
			&schema.CellValue{Value: "i1"}, &schema.CellValue{Value: ""}, &schema.CellValue{Value: float64(0)},
		},
		{
			&schema.CellValue{Value: "r2"}, &schema.CellValue{Value: "s2"},
			&schema.CellValue{Value: "i2"}, &schema.CellValue{Value: ""}, &schema.CellValue{Value: float64(0)},
		},
	}, rows)
}
