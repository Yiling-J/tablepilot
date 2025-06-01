package parquet_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Yiling-J/tablepilot/services/source/parquet"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
)

func initReader(files []string) (*parquet.ParquetGoReader, error) {
	return parquet.NewParquetGoReader(context.TODO(), os.DirFS("./"), files)
}

// test data: 10 parquet files, each file contains 20 rows of data, row id start from 0
func TestGoReader_All(t *testing.T) {
	cases := []struct {
		limit         int64
		offset        int64
		expectedRange string
	}{
		{10, 0, "0-9"},
		{10, 10, "10-19"},
		{10, 18, "18-27"},
		{30, 18, "18-47"},
		{60, 185, "185-199"},
		{100, 75, "75-174"},
	}

	paths := []string{}
	for i := range 10 {
		paths = append(paths, fmt.Sprintf("test_data/%d.parquet", i))
	}

	reader, err := initReader(paths)
	require.NoError(t, err)
	require.Equal(t, []string{"Id", "Name"}, reader.Columns())
	total, err := reader.Total(context.TODO())
	require.NoError(t, err)
	require.Equal(t, int64(200), total)
	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v", c), func(t *testing.T) {
			data, err := reader.Rows(
				context.TODO(),
				c.limit,
				c.offset,
			)
			require.NoError(t, err)

			rg := strings.Split(c.expectedRange, "-")
			start := cast.ToInt64(rg[0])
			end := cast.ToInt64(rg[1])
			current := start
			for _, row := range data {
				id := cast.ToInt64(row[0])
				name := cast.ToInt64(row[1])
				require.Equal(t, current, id)
				require.Equal(t, current, name)
				current += 1
			}
			require.Equal(t, end+1, current)
		})
	}
}
