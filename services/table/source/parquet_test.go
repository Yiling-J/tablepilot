package source

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSource_Parquet(t *testing.T) {
	ctx := context.TODO()

	so := &ParquetSource{Paths: []string{"*.parquet"}}
	err := so.Init(ctx, zap.NewNop().Sugar(), "./parquet/test_data")
	require.NoError(t, err)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false, LinkedColumn: "Id", LinkedContextColumns: []string{"Name"}})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "0", v.Value)
	require.Equal(t, map[string]any{"Name": "0"}, v.ContextValue)
	v, err = indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "1", v.Value)
	require.Equal(t, map[string]any{"Name": "1"}, v.ContextValue)
}
