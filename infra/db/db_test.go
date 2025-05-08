package db

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/tablerow"
	"github.com/stretchr/testify/require"
)

func TestDB_InsertManyOrder(t *testing.T) {
	db := NewTestDB()
	ctx := context.TODO()
	tb, err := db.TableMeta.Create().SetName("user").Save(ctx)
	require.NoError(t, err)
	creates := []*ent.TableRowCreate{}
	for i := range 50 {
		creates = append(creates, db.TableRow.Create().SetCells([]*schema.CellValue{{Value: i}}).SetTablemeta(tb))
	}
	saved, err := db.TableRow.CreateBulk(creates...).Save(ctx)
	require.NoError(t, err)
	require.Equal(t, 50, len(saved))
	for i, row := range saved {
		require.Equal(t, i, row.Cells[0].Value)
	}

	dbrows, err := db.TableRow.Query().Where(tablerow.HasTablemetaWith(tablemeta.Nanoid(tb.Nanoid))).Order(ent.Asc("id")).All(ctx)
	require.NoError(t, err)
	require.Equal(t, 50, len(dbrows))
	for i, row := range dbrows {
		require.Equal(t, float64(i), row.Cells[0].Value)
	}
}
