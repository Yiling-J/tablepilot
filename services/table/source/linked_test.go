package source

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/infra/db"

	"github.com/stretchr/testify/require"
)

func TestLinkedSource_ContextColumns(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.TODO()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	c1, err := db.TableColumn.Create().
		SetTablemeta(tb).SetTablemeta(tb).
		SetName("c1").SetType(tablecolumn.TypeString).
		SetDescription("c1d").
		SetFillMode(tablecolumn.FillModeAi).Save(ctx)
	require.NoError(t, err)
	c2, err := db.TableColumn.Create().
		SetTablemeta(tb).SetTablemeta(tb).
		SetName("c2").SetType(tablecolumn.TypeString).
		SetDescription("c2d").
		SetFillMode(tablecolumn.FillModeAi).Save(ctx)
	require.NoError(t, err)
	err = db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
		{Value: "foo"}, {Value: 1}},
	).Exec(ctx)
	require.NoError(t, err)
	err = db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
		{Value: "bar"}, {Value: 2},
	}).Exec(ctx)
	require.NoError(t, err)
	so := &LinkedSource{
		db:    db,
		Table: tb.Nanoid,
	}

	require.NoError(t, err)
	err = so.Init(ctx, db)
	require.NoError(t, err)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false, LinkedColumn: c1.Nanoid, LinkedContextColumns: []string{c1.Nanoid, c2.Nanoid}})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "foo", v.Value)
	require.Equal(t, map[string]any{
		"c1": map[string]any{"data": "foo", "description": "c1d"},
		"c2": map[string]any{"data": float64(1), "description": "c2d"},
	}, v.ContextValue)

	v, err = indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "bar", v.Value)
	require.Equal(t, map[string]any{
		"c1": map[string]any{"data": "bar", "description": "c1d"},
		"c2": map[string]any{"data": float64(2), "description": "c2d"},
	}, v.ContextValue)
}

func TestLinkedSource_GetLinkedCellValue(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.TODO()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetTablemeta(tb).SetTablemeta(tb).
		SetName("c1").SetType(tablecolumn.TypeString).
		SetDescription("c1d").
		SetFillMode(tablecolumn.FillModeAi).Exec(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetTablemeta(tb).SetTablemeta(tb).
		SetName("c2").SetType(tablecolumn.TypeString).
		SetDescription("c2d").
		SetFillMode(tablecolumn.FillModeAi).Exec(ctx)
	require.NoError(t, err)

	so := &LinkedSource{
		db:    db,
		Table: tb.Nanoid,
	}
	require.NoError(t, err)
	err = so.Init(ctx, db)
	require.NoError(t, err)
	cv := so.GetLinkedCellValue(&ent.TableRow{Cells: []*schema.CellValue{
		{Value: "v1"},
		{Value: "v2"},
	}}, "c1", []string{"c1", "c2"})
	require.Equal(t, "v1", cv.Value)
	require.Equal(t, map[string]any{
		"c1": map[string]any{"data": "v1", "description": "c1d"}, "c2": map[string]any{"data": "v2", "description": "c2d"},
	}, cv.ContextValue)

	cv = so.GetLinkedCellValue(&ent.TableRow{Cells: []*schema.CellValue{
		{Value: "v1"},
		{Value: "v2"},
	}}, "c1", []string{"c1", "foo"})
	require.Equal(t, "v1", cv.Value)
	require.Equal(t, map[string]any{
		"c1": map[string]any{"data": "v1", "description": "c1d"},
	}, cv.ContextValue)

	cv = so.GetLinkedCellValue(&ent.TableRow{Cells: []*schema.CellValue{
		{Value: "v1"},
		{Value: "v2"},
	}}, "c1", []string{"bar", "foo"})
	require.Equal(t, "v1", cv.Value)
	require.Equal(t, map[string]any{}, cv.ContextValue)
}

func TestLinkedSource_Range(t *testing.T) {
	db := db.NewTestDB()
	ctx := context.TODO()
	tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetTablemeta(tb).SetTablemeta(tb).
		SetName("c1").SetType(tablecolumn.TypeString).
		SetDescription("c1d").
		SetFillMode(tablecolumn.FillModeAi).Exec(ctx)
	require.NoError(t, err)
	err = db.TableColumn.Create().
		SetTablemeta(tb).SetTablemeta(tb).
		SetName("c2").SetType(tablecolumn.TypeString).
		SetDescription("c2d").
		SetFillMode(tablecolumn.FillModeAi).Exec(ctx)
	require.NoError(t, err)

	err = db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
		{Value: "foo"}, {Value: 1}},
	).Exec(ctx)
	require.NoError(t, err)
	err = db.TableRow.Create().SetTablemeta(tb).SetCells([]*schema.CellValue{
		{Value: "bar"}, {Value: 2},
	}).Exec(ctx)

	so := &LinkedSource{
		db:    db,
		Table: tb.Nanoid,
	}
	require.NoError(t, err)
	err = so.Init(ctx, db)
	require.NoError(t, err)
	rows := [][]*schema.CellValue{}
	so.Range(func(row *ent.TableRow) bool {
		rows = append(rows, row.Cells)
		return true
	})
	require.Equal(t, [][]*schema.CellValue{
		{&schema.CellValue{Value: "foo"}, &schema.CellValue{Value: float64(1)}},
		{&schema.CellValue{Value: "bar"}, &schema.CellValue{Value: float64(2)}},
	}, rows)

	rows = [][]*schema.CellValue{}
	so.Range(func(row *ent.TableRow) bool {
		rows = append(rows, row.Cells)
		return false
	})
	require.Equal(t, [][]*schema.CellValue{
		{&schema.CellValue{Value: "foo"}, &schema.CellValue{Value: float64(1)}},
	}, rows)
}
