package source

import (
	"context"
	"fmt"
	"tablepilot/ent/tablecolumn"
	"tablepilot/infra/db"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSource_ValidateLinked(t *testing.T) {
	cases := []struct {
		source string
		error  error
	}{
		{`{"type":"linked","table":""}`, ErrTableNameOrIdEmpty()},
		{`{"type":"linked","table":"abc"}`, ErrTableNotFound("abc")},
		{`{"type":"linked","table":"table"}`, ErrColumnNameOrIdEmpty()},
		{`{"type":"linked","table":"table","column":"abc"}`, ErrColumnNotFound("abc")},
		{`{"type":"linked","table":"table","column":"col1"}`, nil},
		{`{"type":"linked","table":"table","column":"col1","context_columns":["abc"]}`, ErrColumnNotFound("abc")},
		{`{"type":"linked","table":"table","column":"col1","context_columns":["col1"]}`, nil},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			ctx := context.TODO()
			db := db.NewTestDB()
			tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
			require.NoError(t, err)
			col, err := db.TableColumn.Create().SetTablemeta(tb).SetName("col1").SetFillMode(tablecolumn.FillModeAi).SetType(tablecolumn.TypeInteger).Save(ctx)
			require.NoError(t, err)

			s, err := ValidateSource(ctx, []byte(c.source), db)
			require.Equal(t, c.error, err)
			if err == nil {
				st := s.(*LinkedSource)
				require.Equal(t, tb.Nanoid, st.Table)
				require.Equal(t, col.Nanoid, st.Column)
				if len(st.ContextColumns) > 0 {
					require.Equal(t, col.Nanoid, st.ContextColumns[0])
				}
			}
		})
	}
}
