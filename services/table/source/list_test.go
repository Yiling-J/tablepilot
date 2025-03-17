package source

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/stretchr/testify/require"
)

func TestSource_List(t *testing.T) {
	ctx := context.TODO()
	so := &ListSource{
		Type:    "list",
		Options: []string{"a", "b", "c"},
	}
	err := so.Init(ctx)
	require.NoError(t, err)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "a", v.Value)
}
