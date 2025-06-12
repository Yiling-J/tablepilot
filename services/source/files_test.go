package source

import (
	"context"
	"os"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSource_Files(t *testing.T) {
	ctx := context.TODO()
	tmpFile, err := os.CreateTemp("./", "test_*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	so := &FilesSource{
		BasicSource: BasicSource{
			Type: "files",
		},
		Paths: []string{"parquet/test_data/*.parquet", "test_*.txt"},
	}
	err = so.Init(ctx, zap.NewNop().Sugar(), "./")
	require.NoError(t, err)
	require.Equal(t, []string{
		"parquet/test_data/0.parquet", "parquet/test_data/1.parquet", "parquet/test_data/2.parquet",
		"parquet/test_data/3.parquet", "parquet/test_data/4.parquet", "parquet/test_data/5.parquet",
		"parquet/test_data/6.parquet", "parquet/test_data/7.parquet", "parquet/test_data/8.parquet",
		"parquet/test_data/9.parquet", tmpFile.Name()[2:],
	}, so.Files)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "parquet/test_data/0.parquet", v.Value)
}
