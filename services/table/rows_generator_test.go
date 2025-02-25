package table

import (
	"context"
	"fmt"
	"tablepilot/ent/schema"
	"tablepilot/ent/tablecolumn"
	"tablepilot/infra/db"
	"tablepilot/services/ai/promptbuilder"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRowsGenerator_PrepareCOntextRows(t *testing.T) {
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
			generator, err := NewRowsGenerator(ctx, "foo", "", 5, 5, db, nil, zap.NewNop().Sugar())
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
