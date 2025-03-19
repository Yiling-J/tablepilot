package source

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Yiling-J/tablepilot/infra/db"

	"github.com/stretchr/testify/require"
)

func TestSource_Validate(t *testing.T) {
	cases := []struct {
		source string
		error  error
	}{
		{`{"type":"linked","table":""}`, ErrTableNameOrIdEmpty()},
		{`{"type":"linked","table":"abc"}`, ErrTableNotFound("abc")},
		{`{"type":"linked","table":"table"}`, nil},
		{`{"type":"list"}`, errors.New("no options")},
		{`{"type":"list","file":"z.txt"}`, nil},
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
				st, ok := s.(*LinkedSource)
				if ok {
					require.Equal(t, tb.Nanoid, st.Table)
				}
			}
		})
	}
}
