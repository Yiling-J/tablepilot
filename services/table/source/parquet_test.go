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

func TestParquetSource_Basic(t *testing.T) {
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

func TestParquetSource_HuggingFace(t *testing.T) {
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

func TestParquetSource_GetLinkedCellValue(t *testing.T) {
	so := &ParquetSource{Paths: []string{"*.parquet"}}
	err := so.Init(context.TODO(), nil, zap.NewNop().Sugar(), "./parquet/test_data")
	require.NoError(t, err)
	cv := so.GetLinkedCellValue(map[string]any{"Id": "1", "Name": "Foo"}, "Id", []string{"Name"})
	require.Equal(t, "1", cv.Value)
	require.Equal(t, map[string]any{"Name": "Foo"}, cv.ContextValue)

	cv = so.GetLinkedCellValue(map[string]any{"Id": "1", "Name": "Foo"}, "Id", []string{"Foo"})
	require.Equal(t, "1", cv.Value)
	require.Equal(t, map[string]any{}, cv.ContextValue)
}

func TestParquetSource_Range(t *testing.T) {
	so := &ParquetSource{Paths: []string{"0.parquet"}}
	err := so.Init(context.TODO(), nil, zap.NewNop().Sugar(), "./parquet/test_data")
	require.NoError(t, err)
	rows := []map[string]any{}
	err = so.Range(context.TODO(), func(row map[string]any) bool {
		rows = append(rows, row)
		return true
	})
	require.NoError(t, err)
	expected := []map[string]any{}
	for i := range 20 {
		expected = append(expected, map[string]any{"Id": cast.ToString(i), "Name": cast.ToString(i)})
	}
	require.Equal(t, expected, rows)

	rows = []map[string]any{}
	err = so.Range(context.TODO(), func(row map[string]any) bool {
		rows = append(rows, row)
		return false
	})
	require.NoError(t, err)
	require.Equal(t, []map[string]any{{"Id": "0", "Name": "0"}}, rows)
}

func TestParquetSource_HuggingFaceRange(t *testing.T) {
	ctx := context.TODO()
	so := &ParquetSource{Huggingface: &Huggingface{
		Dataset:         "foo/bar",
		rangeBatchCount: 1,
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
			if offset > 19 {
				return &huggingface.RowResponse{Rows: []huggingface.RowInfo{}}, nil
			}
			return &huggingface.RowResponse{Rows: []huggingface.RowInfo{
				{Row: map[string]any{"Id": cast.ToString(offset), "Name": "n" + cast.ToString(offset)}},
			}}, nil
		},
	}
	err := so.Init(ctx, hfClient, zap.NewNop().Sugar(), "")
	require.NoError(t, err)
	rows := []map[string]any{}
	err = so.Range(ctx, func(row map[string]any) bool {
		rows = append(rows, row)
		return true
	})
	require.NoError(t, err)
	expected := []map[string]any{}
	for i := range 20 {
		expected = append(expected, map[string]any{"Id": cast.ToString(i), "Name": "n" + cast.ToString(i)})
	}
	require.Equal(t, expected, rows)

	rows = []map[string]any{}
	err = so.Range(context.TODO(), func(row map[string]any) bool {
		rows = append(rows, row)
		return false
	})
	require.NoError(t, err)
	require.Equal(t, []map[string]any{{"Id": "0", "Name": "n" + "0"}}, rows)
}
