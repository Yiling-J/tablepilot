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
		{`{"name":"s1","type":"list"}`, errors.New("souce s1 options should not be empty")},
		{`{"type":"list","file":"z.txt"}`, nil},
		{`{"type": "csv","paths":["foo.csv"]}`, nil},
		{`{"type": "csv"}`, errors.New("paths is empty")},
		{`{"type": "csv","kaggle":"foo/bar"}`, errors.New("paths is empty")},
		{`{"type": "parquet","paths":["foo.parquet"]}`, nil},
		{`{"type": "parquet"}`, errors.New("paths is empty")},
		{`{"type": "parquet","huggingface":{"dataset":"abc"}}`, nil},
		{`{"type": "parquet","huggingface":{"dataset":""}}`, errors.New("Hugging Face dataset is empty")},
		{`{"type": "files","paths":["foo.csv"]}`, nil},
		{`{"type": "files"}`, errors.New("paths is empty")},
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
