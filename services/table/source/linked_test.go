package source

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/infra/db"

	"github.com/stretchr/testify/require"
)

func TestSource_LinkedContextColumns(t *testing.T) {
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
		db:             db,
		Table:          tb.Nanoid,
		Column:         c1.Nanoid,
		ContextColumns: []string{c1.Nanoid, c2.Nanoid},
	}

	require.NoError(t, err)
	err = so.Init(ctx, db)
	require.NoError(t, err)
	indexer := NewIndexer(so, false, false, 0)
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
