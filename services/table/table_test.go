package table

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/dataset"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/tablerow"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/ai/client"
	"github.com/Yiling-J/tablepilot/services/ai/promptbuilder"
	dataset_service "github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/spf13/cast"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func requireColumnEqual(t *testing.T, expcted, column *ent.TableColumn) {
	require.Equal(t, expcted.Name, column.Name)
	require.Equal(t, expcted.Description, column.Description)
	require.Equal(t, expcted.Type, column.Type)
	require.Equal(t, expcted.FillMode, column.FillMode)
	require.Equal(t, expcted.SourceID, column.SourceID)
	require.Equal(t, expcted.SourceType, column.SourceType)
	if column.SourceID != "" {
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
	columns := []*TableGenColumn{
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
			FillMode: "pick", SourceID: "tags", SourceType: tablecolumn.SourceTypeDataset, Random: true, Replacement: true, Repeat: 3,
		},
		{
			Name: "country", Description: "recipe country", Type: "string",
			FillMode: "pick", SourceType: tablecolumn.SourceTypeOptions,
			Options: []string{"China", "Italy"},
		},
		{
			Name: "user", Description: "recipe user", Type: "boolean",
			FillMode: "pick", SourceID: "users", SourceType: tablecolumn.SourceTypeTable, LinkedColumn: "name", LinkedContextColumns: []string{"age"},
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
			require.Equal(t, prompt, request.Messages[0].Content[0].Data)
			return &client.ChatResponse{
				Content: `[{"name":"extra","type":"string"},{"name":"extra2","type":"string"}]`,
				Tokens:  100,
			}, nil
		},
	}
	srv, err := NewTableService(&config.Config{}, db, aiService, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	ds, err := db.Dataset.Create().SetName("tags").SetType(dataset.TypeList).SetValues([]string{
		"a", "b", "c",
	}).Save(ctx)
	require.NoError(t, err)
	userTable, err := db.TableMeta.Create().SetName("users").Save(ctx)
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
	require.Equal(t, "test", table.Name)
	require.Equal(t, "test table", table.Description)
	require.Equal(t, id, table.Nanoid)
	require.Equal(t, 6, len(table.Edges.Columns))
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
			SourceID: ds.Nanoid, SourceType: tablecolumn.SourceTypeDataset,
			Random: true, Replacement: true, Repeat: 3,
		},
		table.Edges.Columns[2],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "country", Description: "recipe country", ContextLength: 0,
			Type: tablecolumn.TypeString, FillMode: tablecolumn.FillModePick,
			Options: []string{"China", "Italy"}, SourceType: tablecolumn.SourceTypeOptions,
		},
		table.Edges.Columns[3],
	)
	requireColumnEqual(
		t,
		&ent.TableColumn{
			Name: "user", Description: "recipe user", ContextLength: 0,
			Type: tablecolumn.TypeBoolean, FillMode: tablecolumn.FillModePick,
			SourceID: userTable.Nanoid, SourceType: tablecolumn.SourceTypeTable,
			LinkedColumn: "name", LinkedContextColumns: []string{"age"},
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
		&config.Config{}, db, nil, nil, zap.NewNop().Sugar(),
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
		&config.Config{}, db, nil, nil, zap.NewNop().Sugar(),
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
		&config.Config{}, db, nil, nil, zap.NewNop().Sugar(),
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
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
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

	// create table when importing with given table name
	id, err := srv.Import(ctx, ImportRequest{
		Filename: "bar.csv",
		Table:    "",
		Reader:   strings.NewReader(buffer.String()),
		Name:     "foo",
	})
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

	// create table when importing, file name
	id, err = srv.Import(ctx, ImportRequest{
		Filename: "bar",
		Reader:   strings.NewReader(buffer.String()),
	})
	require.NoError(t, err)

	t2, err := db.TableMeta.Query().Where(tablemeta.Nanoid(id)).Only(ctx)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(t2.Name, "bar_"))

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

	id, err = srv.Import(ctx, ImportRequest{
		Table:  "foo",
		Reader: strings.NewReader(buffer.String()),
	})
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

	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
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

	id, err := srv.Import(ctx, ImportRequest{
		Reader: strings.NewReader(buffer.String()),
		Table:  "foo",
	})
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

func TestTableService_ImportSourceColumn(t *testing.T) {
	defer func() { _ = os.RemoveAll("datasets") }()
	cases := []struct {
		name       string
		sourceType tablecolumn.SourceType
	}{
		{"list", tablecolumn.SourceTypeDataset},
		{"table", tablecolumn.SourceTypeTable},
		{"csv", tablecolumn.SourceTypeDataset},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := db.NewTestDB()
			ctx := context.Background()
			linkedColumn := "col"
			dsid := ""
			switch tc.name {
			case "list":
				dsv := dataset_service.NewDatasetService(db, &config.Config{})
				id, err := dsv.Create(t.Context(), &dataset_service.CreateDatasetRequest{
					Name:        "s1",
					Description: "ds",
					Type:        dataset.TypeList,
				})
				require.NoError(t, err)
				dsid = id
			case "table":
				tb, err := db.TableMeta.Create().SetName("s1").Save(ctx)
				require.NoError(t, err)
				dsid = tb.Nanoid
				col, err := db.TableColumn.Create().
					SetTablemeta(tb).SetTablemeta(tb).
					SetName("col").SetType(tablecolumn.TypeString).
					SetDescription("c1d").
					SetFillMode(tablecolumn.FillModeAi).Save(ctx)
				require.NoError(t, err)
				linkedColumn = col.Nanoid
				err = db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
					{Value: "bar"}},
				).Exec(ctx)
				require.NoError(t, err)
			case "csv":
				var buf []byte
				b := bytes.NewBuffer(buf)
				writer := csv.NewWriter(b)
				require.NoError(t, writer.Write([]string{"col"}))
				require.NoError(t, writer.Write([]string{"bar"}))
				writer.Flush()
				require.NoError(t, writer.Error())
				dsv := dataset_service.NewDatasetService(db, &config.Config{})
				id, err := dsv.Create(t.Context(), &dataset_service.CreateDatasetRequest{
					Name:        "s1",
					Description: "ds",
					Type:        dataset.TypeCsv,
					Files:       []io.Reader{b},
				})
				require.NoError(t, err)
				dsid = id
			default:
				require.FailNow(t, "unknown source")
			}

			tb, err := db.TableMeta.Create().SetName("foo").Save(ctx)
			require.NoError(t, err)
			err = db.TableColumn.Create().SetName("string").SetTablemeta(tb).SetFillMode(tablecolumn.FillModePick).
				SetType(tablecolumn.TypeString).SetSourceID(dsid).SetSourceType(tc.sourceType).
				SetLinkedColumn(linkedColumn).SetLinkedContextColumns([]string{linkedColumn}).
				Exec(ctx)
			require.NoError(t, err)

			srv, err := NewTableService(&config.Config{Common: config.Common{SourceDataDir: "./"}}, db, nil, nil, zap.NewNop().Sugar())
			require.NoError(t, err)
			records := [][]string{
				{"string"},
				{"bar"},
			}

			buffer := bytes.NewBuffer([]byte(""))
			w := csv.NewWriter(buffer)
			for _, record := range records {
				err := w.Write(record)
				require.NoError(t, err)
			}
			w.Flush()

			id, err := srv.Import(ctx, ImportRequest{
				Table:  "foo",
				Reader: strings.NewReader(buffer.String()),
			})
			require.NoError(t, err)

			table, err := db.TableMeta.Query().WithRows(func(trq *ent.TableRowQuery) {
				trq.Order(ent.Asc(tablerow.FieldID))
			}).Where(tablemeta.Nanoid(id)).First(ctx)
			require.NoError(t, err)
			cell := table.Edges.Rows[0].Cells[0]
			switch tc.name {
			case "list":
				require.Equal(t, &schema.CellValue{
					Value: "bar",
				}, cell)
			case "table":
				require.Equal(t, &schema.CellValue{
					Value: "bar",
					ContextValue: map[string]any{"col": map[string]any{
						"data":        "bar",
						"description": "c1d",
					}},
				}, cell)
			default:
				require.Equal(t, &schema.CellValue{
					Value:        "bar",
					ContextValue: map[string]any{"col": "bar"},
				}, cell)
			}
		})
	}
}

func TestTableService_ImportImage(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()

	srv, err := NewTableService(&config.Config{}, db, &ai.AiServiceMock{
		ChatFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
			builder := promptbuilder.NewNewImageToTableBuilder("foobar")
			prompt, err := builder.Prompt()
			require.NoError(t, err)
			require.Equal(t, request.Messages[0].Content[0].Data, prompt)
			require.Equal(t, 0.1, request.Temperature)
			require.Equal(t, "m1", request.Model)
			require.Equal(t, int64(6000), request.MaxOutputTokens)
			generated := ImageExtractionOutput{
				TableName:        "table_test",
				TableDescription: "abc",
				Columns: []ImageExtractionColumn{
					{Name: "col1", Description: "d1", Type: "string"},
					{Name: "col2", Description: "d2", Type: "integer"},
				},
				Rows: [][]string{
					{"a", "1"},
					{"b", "2"},
				},
			}
			b, err := json.Marshal(generated)
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
			}, nil
		},
	}, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	pb, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC")
	require.NoError(t, err)
	id, err := srv.ImportImage(ctx, ImportRequest{
		Data:   pb,
		Prompt: "foobar",
		Model:  "m1",
	})
	require.NoError(t, err)

	table, err := db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
		tcq.Order(ent.Asc(tablecolumn.FieldID))
	}).WithRows(func(trq *ent.TableRowQuery) {
		trq.Order(ent.Asc(tablerow.FieldID))
	}).Where(tablemeta.Name("table_test")).Only(ctx)
	require.NoError(t, err)
	require.Equal(t, id, table.Nanoid)
	require.Equal(t, "abc", table.Description)
	columnNames := []string{}
	columnTypes := []string{}
	for _, col := range table.Edges.Columns {
		require.Equal(t, tablecolumn.TypeString, col.Type)
		columnNames = append(columnNames, col.Name)
		columnTypes = append(columnTypes, col.Type.String())
	}
	require.Equal(t, []string{"col1", "col2"}, columnNames)
	require.Equal(t, []string{"string", "string"}, columnTypes)
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

func TestTableService_ImportImageToTable(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	table, err := db.TableMeta.Create().SetName("foo").SetDescription("bar").Save(ctx)
	require.NoError(t, err)
	col, err := db.TableColumn.Create().
		SetName("c1").
		SetFillMode(tablecolumn.FillModeAi).
		SetTablemeta(table).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	col2, err := db.TableColumn.Create().
		SetName("c2").
		SetFillMode(tablecolumn.FillModePick).
		SetTablemeta(table).
		SetType(tablecolumn.TypeString).Save(ctx)
	require.NoError(t, err)
	_ = col2

	data := ""
	srv, err := NewTableService(&config.Config{}, db, &ai.AiServiceMock{
		ChatFunc: func(ctx context.Context, request *client.ChatRequest) (*client.ChatResponse, error) {
			data = request.Messages[0].Content[0].Data
			require.Equal(t, 0.1, request.Temperature)
			require.Equal(t, "m1", request.Model)
			require.Equal(t, int64(6000), request.MaxOutputTokens)

			d := []map[string]any{{"__id__": 0, col.Nanoid: "d"}, {"__id__": 1, col.Nanoid: "e"}}
			b, err := json.Marshal(map[string]any{"data": d})
			require.NoError(t, err)
			require.NoError(t, err)
			return &client.ChatResponse{
				Content: string(b),
			}, nil
		},
	}, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	pb, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAIAAAACDbGyAAAAEklEQVR4nGJiQAWU8gEBAAD//wIwAAtSRUCpAAAAAElFTkSuQmCC")
	require.NoError(t, err)
	id, err := srv.ImportImage(ctx, ImportRequest{
		Data:   pb,
		Prompt: "foobar",
		Model:  "m1",
		Table:  "foo",
	})
	require.NoError(t, err)
	require.Equal(t, table.Nanoid, id)

	rows, err := table.QueryRows().All(ctx)
	require.NoError(t, err)
	nw := [][]any{}
	for _, row := range rows {
		r := []any{}
		for _, cell := range row.Cells {
			r = append(r, cell.Value)
		}
		nw = append(nw, r)
	}
	require.Equal(t, [][]any{{"d", nil}, {"e", nil}}, nw)
	builder := promptbuilder.NewNewImageToTableBuilder("foobar")
	tm, err := db.TableMeta.Query().WithColumns().Where(tablemeta.ID(table.ID)).Only(ctx)
	require.NoError(t, err)
	builder.ToTable(tm)
	prompt, err := builder.Prompt()
	require.NoError(t, err)
	require.Equal(t, prompt, data)
}

func TestTableService_ListTables(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
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
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
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
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
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

func TestTableService_ImportLinked(t *testing.T) {
	defer func() { _ = os.RemoveAll("datasets") }()
	db := db.NewTestDB()
	ctx := context.Background()
	srv, err := NewTableService(&config.Config{Common: config.Common{SourceDataDir: "./"}}, db, nil, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	tid, err := srv.Create(ctx, &TableGenRequest{Name: "lt", Columns: []*TableGenColumn{
		{Name: "c1", FillMode: "ai", Type: "string", Description: "c1c1"},
		{Name: "c2", FillMode: "ai", Type: "string", Description: "c2c2"},
	}})
	_ = tid
	require.NoError(t, err)
	err = srv.CreateRows(ctx, "lt", []map[string]any{
		{"c1": "aa", "c2": "foo"},
		{"c1": "bb", "c2": "bar"},
	})
	require.NoError(t, err)

	var buf []byte
	b := bytes.NewBuffer(buf)
	writer := csv.NewWriter(b)
	require.NoError(t, writer.Write([]string{"c1", "c2"}))
	require.NoError(t, writer.Write([]string{"a", "v1"}))
	require.NoError(t, writer.Write([]string{"b", "v2"}))
	writer.Flush()
	require.NoError(t, writer.Error())
	ds := dataset_service.NewDatasetService(db, &config.Config{})
	did, err := ds.Create(t.Context(), &dataset_service.CreateDatasetRequest{
		Name:  "ds",
		Type:  dataset.TypeCsv,
		Files: []io.Reader{b},
	})
	require.NoError(t, err)

	tb, err := db.TableMeta.Create().SetName("t1").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("col1").
		SetFillMode(tablecolumn.FillModePick).
		SetTablemeta(tb).
		SetSourceID(did).SetSourceType(tablecolumn.SourceTypeDataset).
		SetLinkedColumn("c1").SetLinkedContextColumns([]string{"c1", "c2"}).
		SetType(tablecolumn.TypeString).Exec(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetName("col2").
		SetFillMode(tablecolumn.FillModePick).
		SetTablemeta(tb).
		SetSourceID(tid).SetSourceType(tablecolumn.SourceTypeTable).
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

	id, err := srv.Import(ctx, ImportRequest{
		Table:  "t1",
		Reader: strings.NewReader(buffer.String()),
	})
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
	columns := []*TableGenColumn{
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
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
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
	columns = []*TableGenColumn{
		{
			Name: "steps", Description: "recipe steps go", Type: "array",
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
		{ID: "EfhxLZ", Name: "steps", Description: "recipe steps go", Type: "array", FillMode: "ai"},
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

func TestTableService_GetTableSchema(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.Background()
	columns := []*TableGenColumn{
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
			FillMode: "pick", SourceID: "tags", SourceType: tablecolumn.SourceTypeDataset, Random: true, Replacement: true, Repeat: 3,
		},
		{
			Name: "country", Description: "recipe country", Type: "string",
			FillMode: "pick", SourceID: "countries", SourceType: tablecolumn.SourceTypeDataset,
		},
		{
			Name: "user", Description: "recipe user", Type: "boolean", SourceType: tablecolumn.SourceTypeTable,
			FillMode: "pick", SourceID: "users", LinkedColumn: "name", LinkedContextColumns: []string{"age"},
		},
		{
			Name: "dish_type", Description: "recipe dish type", Type: "string", SourceType: tablecolumn.SourceTypeOptions,
			FillMode: "pick", Options: []string{"foo", "bar"},
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
			require.Equal(t, prompt, request.Messages[0].Content[0].Data)
			return &client.ChatResponse{
				Content: `[{"name":"extra","type":"string"},{"name":"extra2","type":"string"}]`,
				Tokens:  100,
			}, nil
		},
	}
	srv, err := NewTableService(&config.Config{}, db, aiService, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	ds1, err := db.Dataset.Create().SetName("countries").SetType(dataset.TypeList).SetValues([]string{
		"China", "Japan", "Englland",
	}).Save(ctx)
	require.NoError(t, err)
	ds2, err := db.Dataset.Create().SetName("tags").SetType(dataset.TypeList).SetValues([]string{
		"a", "b", "c",
	}).Save(ctx)
	require.NoError(t, err)
	_ = ds1 == ds2
	userTable, err := db.TableMeta.Create().SetName("users").Save(ctx)
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
	schema, err := srv.GetTableSchema(ctx, id)
	require.NoError(t, err)
	expected := `{"name":"test","model":"","description":"test table","columns":[{"name":"name","description":"recipe name","type":"string","fill_mode":"ai","source_type":"","source_id":"","options":null,"random":false,"replacement":false,"repeat":1,"context_length":5,"linked_column":"","linked_context_columns":[]},{"name":"count","description":"recipe count","type":"integer","fill_mode":"ai","source_type":"","source_id":"","options":null,"random":false,"replacement":false,"repeat":1,"context_length":3,"linked_column":"","linked_context_columns":[]},{"name":"tag","description":"recipe tag","type":"array","fill_mode":"pick","source_type":"dataset","source_id":"gbHJdm","options":null,"random":true,"replacement":true,"repeat":3,"context_length":0,"linked_column":"","linked_context_columns":null},{"name":"country","description":"recipe country","type":"string","fill_mode":"pick","source_type":"dataset","source_id":"UkLWZg","options":null,"random":false,"replacement":false,"repeat":0,"context_length":0,"linked_column":"","linked_context_columns":null},{"name":"user","description":"recipe user","type":"boolean","fill_mode":"pick","source_type":"table","source_id":"UkLWZg","options":null,"random":false,"replacement":false,"repeat":0,"context_length":0,"linked_column":"name","linked_context_columns":["age"]},{"name":"dish_type","description":"recipe dish type","type":"string","fill_mode":"pick","source_type":"options","source_id":"","options":["foo","bar"],"random":false,"replacement":false,"repeat":0,"context_length":0,"linked_column":"","linked_context_columns":null},{"name":"extra","description":"","type":"string","fill_mode":"ai","source_type":"","source_id":"","options":null,"random":false,"replacement":false,"repeat":1,"context_length":0,"linked_column":"","linked_context_columns":[]}]}`
	b, err := json.Marshal(schema)
	require.NoError(t, err)
	require.Equal(t, expected, string(b))
}

func TestTableService_Validate(t *testing.T) {
	cases := []struct {
		req *TableGenRequest
		err string
	}{
		{
			req: &TableGenRequest{},
			err: "table.Validate: columns should not be empty",
		},
		{
			req: &TableGenRequest{
				Columns: []*TableGenColumn{
					{SourceID: "s1", SourceType: tablecolumn.SourceTypeTable, FillMode: "pick"},
				},
			},
			err: "source table s1 not found",
		},
	}

	db := db.NewTestDB()
	ctx := context.Background()
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
	require.NoError(t, err)

	for _, tc := range cases {
		err := srv.Validate(ctx, tc.req)
		require.Equal(t, tc.err, err.Error())
	}
}

func TestTableService_CSV(t *testing.T) {
	db := db.NewTestDB()
	userTable, err := db.TableMeta.Create().SetName("user").Save(t.Context())
	require.NoError(t, err)
	_, err = db.TableColumn.CreateBulk(
		db.TableColumn.Create().SetTablemeta(userTable).SetName("name").SetType(
			tablecolumn.TypeString,
		).SetFillMode(tablecolumn.FillModeAi),
		db.TableColumn.Create().SetTablemeta(userTable).SetName("age").SetType(
			tablecolumn.TypeInteger,
		).SetFillMode(tablecolumn.FillModeAi),
	).Save(t.Context())
	require.NoError(t, err)
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
	err = srv.CreateRows(t.Context(), "user", []map[string]any{
		{"name": "aa", "age": 1},
		{"name": "bb", "age": 2},
	})
	require.NoError(t, err)
	data, err := srv.CSV(t.Context(), "user")
	require.NoError(t, err)
	require.Equal(t, "name,age\naa,1\nbb,2\n", string(data))
}

func TestTableService_CreateColumn(t *testing.T) {
	db := db.NewTestDB()
	userTable, err := db.TableMeta.Create().SetName("user").Save(t.Context())
	require.NoError(t, err)
	_, err = db.TableColumn.CreateBulk(
		db.TableColumn.Create().SetTablemeta(userTable).SetName("name").SetType(
			tablecolumn.TypeString,
		).SetFillMode(tablecolumn.FillModeAi),
		db.TableColumn.Create().SetTablemeta(userTable).SetName("age").SetType(
			tablecolumn.TypeInteger,
		).SetFillMode(tablecolumn.FillModeAi),
	).Save(t.Context())
	require.NoError(t, err)
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
	err = srv.CreateRows(t.Context(), "user", []map[string]any{
		{"name": "aa", "age": 1},
		{"name": "bb", "age": 2},
	})
	require.NoError(t, err)
	_, err = srv.CreateColumn(t.Context(), "user", TableGenColumn{
		Name:        "job",
		Description: "user job",
		Type:        "string",
		FillMode:    "ai",
	})
	require.NoError(t, err)

	columns, err := userTable.QueryColumns().All(t.Context())
	require.NoError(t, err)
	require.Equal(t, 3, len(columns))
	found := false
	for _, col := range columns {
		if col.Name == "job" {
			require.Equal(t, "user job", col.Description)
			require.Equal(t, tablecolumn.TypeString, col.Type)
			require.Equal(t, tablecolumn.FillModeAi, col.FillMode)
			found = true
			break
		}
	}
	require.True(t, found)
	rows, err := userTable.QueryRows().All(t.Context())
	require.NoError(t, err)
	for _, row := range rows {
		cells := row.Cells
		require.Equal(t, "", cells[2].Value)
	}
}
func TestTableService_DeleteColumn(t *testing.T) {
	db := db.NewTestDB()
	userTable, err := db.TableMeta.Create().SetName("user").Save(t.Context())
	require.NoError(t, err)
	_, err = db.TableColumn.CreateBulk(
		db.TableColumn.Create().SetTablemeta(userTable).SetName("name").SetType(
			tablecolumn.TypeString,
		).SetFillMode(tablecolumn.FillModeAi),
		db.TableColumn.Create().SetTablemeta(userTable).SetName("age").SetType(
			tablecolumn.TypeInteger,
		).SetFillMode(tablecolumn.FillModeAi),
		db.TableColumn.Create().SetTablemeta(userTable).SetName("job").SetType(
			tablecolumn.TypeString,
		).SetFillMode(tablecolumn.FillModeAi),
	).Save(t.Context())
	require.NoError(t, err)
	srv, err := NewTableService(&config.Config{}, db, nil, nil, zap.NewNop().Sugar())
	require.NoError(t, err)
	err = srv.CreateRows(t.Context(), "user", []map[string]any{
		{"name": "aa", "age": 1, "job": "a"},
		{"name": "bb", "age": 2, "job": "b"},
	})
	require.NoError(t, err)

	_, err = srv.DeleteColumn(t.Context(), "user", "age")
	require.NoError(t, err)

	columns, err := userTable.QueryColumns().All(t.Context())
	require.NoError(t, err)
	require.Equal(t, 2, len(columns))
	require.Equal(t, "name", columns[0].Name)
	require.Equal(t, "job", columns[1].Name)
	rows, err := userTable.QueryRows().All(t.Context())
	require.NoError(t, err)
	ns := []string{}
	js := []string{}
	for _, row := range rows {
		cells := row.Cells
		require.Equal(t, 2, len(cells))
		ns = append(ns, cast.ToString(cells[0].Value))
		js = append(js, cast.ToString(cells[1].Value))
	}
	require.ElementsMatch(t, []string{"aa", "bb"}, ns)
	require.ElementsMatch(t, []string{"a", "b"}, js)
}
