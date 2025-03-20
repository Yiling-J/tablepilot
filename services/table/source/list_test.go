package source

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSource_List(t *testing.T) {
	ctx := context.TODO()
	so := &ListSource{
		Type:    "list",
		Options: []string{"a", "b", "c"},
	}
	err := so.Init(ctx, "./")
	require.NoError(t, err)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "a", v.Value)
}

func TestSource_ListFile(t *testing.T) {
	ctx := context.TODO()
	tmpFile, err := os.CreateTemp("./", "test_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	w := bufio.NewWriter(tmpFile)
	for _, line := range []string{"a", "b", "c"} {
		fmt.Fprintln(w, line)
	}
	err = w.Flush()
	require.NoError(t, err)

	so := &ListSource{
		Type: "list",
		File: strings.TrimPrefix(tmpFile.Name(), "./"),
	}
	err = so.Init(ctx, "./")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, so.Options)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "a", v.Value)
}
