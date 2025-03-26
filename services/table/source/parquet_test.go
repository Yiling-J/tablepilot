package source

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/services/table/source/huggingface"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSource_Parquet(t *testing.T) {
	ctx := context.TODO()

	so := &ParquetSource{Paths: []string{"*.parquet"}}
	err := so.Init(ctx, nil, zap.NewNop().Sugar(), "./parquet/test_data")
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

func TestSource_HuggingFaceParquet(t *testing.T) {
	ctx := context.TODO()

	so := &ParquetSource{Huggingface: &Huggingface{
		Dataset: "foo/bar",
	}}
	hfClient := &huggingface.ClientMock{
		GetDatasetSizeFunc: func(ctx context.Context) (*huggingface.DatasetSizeResponse, error) {
			return &huggingface.DatasetSizeResponse{Size: huggingface.SizeInfo{
				Splits: []huggingface.SplitInfo{{Config: "default", Split: "train", NumRows: 100}},
			}}, nil
		},
		GetDatasetInfoFunc: func(ctx context.Context) (*huggingface.DatasetInfoResponse, error) {
			return &huggingface.DatasetInfoResponse{Features: map[string]any{
				"Id":   true,
				"Name": true,
			}}, nil
		},
		GetDatasetRowsFunc: func(ctx context.Context, offset, length int) (*huggingface.RowResponse, error) {
			return &huggingface.RowResponse{Rows: []huggingface.RowInfo{
				{Row: map[string]any{"Id": cast.ToString(offset), "Name": "n" + cast.ToString(offset)}},
			}}, nil
		},
	}
	err := so.Init(ctx, hfClient, zap.NewNop().Sugar(), "")
	require.NoError(t, err)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false, LinkedColumn: "Id", LinkedContextColumns: []string{"Name"}})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "0", v.Value)
	require.Equal(t, map[string]any{"Name": "n0"}, v.ContextValue)
	v, err = indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "1", v.Value)
	require.Equal(t, map[string]any{"Name": "n1"}, v.ContextValue)
}
