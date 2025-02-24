package source

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSource_List(t *testing.T) {
	ctx := context.TODO()
	so := &ListSource{
		indexer: newIndexer(false, false, 5, 0),
		Type:    "list",
		Options: []string{"a", "b", "c"},
	}
	err := so.Init(ctx)
	require.NoError(t, err)
	v, err := so.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "a", v.Value)
}
