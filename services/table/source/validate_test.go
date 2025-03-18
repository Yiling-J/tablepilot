package source

import (
	"context"
	"fmt"
	"testing"

	"github.com/Yiling-J/tablepilot/infra/db"

	"github.com/stretchr/testify/require"
)

func TestSource_ValidateLinked(t *testing.T) {
	cases := []struct {
		source string
		error  error
	}{
		{`{"type":"linked","table":""}`, ErrTableNameOrIdEmpty()},
		{`{"type":"linked","table":"abc"}`, ErrTableNotFound("abc")},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			ctx := context.TODO()
			db := db.NewTestDB()
			tb, err := db.TableMeta.Create().SetName("table").Save(ctx)
			require.NoError(t, err)

			s, err := ValidateSource(ctx, []byte(c.source), db)
			require.Equal(t, c.error, err)
			if err == nil {
				st := s.(*LinkedSource)
				require.Equal(t, tb.Nanoid, st.Table)
			}
		})
	}
}
