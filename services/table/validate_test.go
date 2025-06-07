package table

import (
	"context"
	"fmt"
	"testing"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/infra/db"

	"github.com/stretchr/testify/require"
)

func TestTableService_ValidateLinkedColumnInfo(t *testing.T) {
	cases := []struct {
		column TableGenColumn
		error  error
	}{
		{TableGenColumn{LinkedColumn: "abc"}, ErrColumnNotFound("abc")},
		{TableGenColumn{LinkedColumn: "col1"}, nil},
		{TableGenColumn{LinkedColumn: "__id__"}, nil},
		{TableGenColumn{LinkedColumn: "col1", LinkedContextColumns: []string{"abc"}}, ErrColumnNotFound("abc")},
		{TableGenColumn{LinkedColumn: "col1", LinkedContextColumns: []string{"col1"}}, nil},
		{TableGenColumn{LinkedColumn: "col1", LinkedContextColumns: []string{"__id__"}}, nil},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			ctx := context.TODO()
			db := db.NewTestDB()
			tx, _ := db.Tx(ctx)
			tb, err := tx.TableMeta.Create().SetName("table").Save(ctx)
			require.NoError(t, err)
			col, err := tx.TableColumn.Create().SetTablemeta(tb).SetName("col1").SetFillMode(tablecolumn.FillModeAi).SetType(tablecolumn.TypeInteger).Save(ctx)
			require.NoError(t, err)
			if c.column.LinkedColumn == "__id__" {
				c.column.LinkedColumn = col.Nanoid
			}
			for i, cc := range c.column.LinkedContextColumns {
				if cc == "__id__" {
					c.column.LinkedContextColumns[i] = col.Nanoid
				}
			}
			c.column.FillMode = "pick"
			c.column.SourceID = "table"
			c.column.SourceType = tablecolumn.SourceTypeTable
			err = validateLinkedColumnInfo(ctx, tx.Client(), []*TableGenColumn{&c.column})
			require.Equal(t, c.error, err)
		})
	}
}
