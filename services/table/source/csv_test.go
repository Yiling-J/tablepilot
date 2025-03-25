package source

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"testing"
	"testing/fstest"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSource_CSVParsePaths(t *testing.T) {
	mockFS := fstest.MapFS{
		"file1.txt":              {},
		"file2.txt":              {},
		"dir1/file3.txt":         {},
		"dir1/file4.log":         {},
		"dir1/subdir1/file5.txt": {},
		"dir1/subdir1/file6.log": {},
		"dir2/file7.txt":         {},
	}

	tests := []struct {
		name   string
		paths  []string
		expect []string
		err    bool
	}{
		{
			name:   "Exact match",
			paths:  []string{"file1.txt"},
			expect: []string{"file1.txt"},
		},
		{
			name:   "Single star wildcard",
			paths:  []string{"*.txt"},
			expect: []string{"file1.txt", "file2.txt"},
		},
		{
			name:   "Double star wildcard",
			paths:  []string{"**/*.txt"},
			expect: []string{"file1.txt", "file2.txt", "dir1/file3.txt", "dir1/subdir1/file5.txt", "dir2/file7.txt"},
		},
		{
			name:   "Mixed patterns",
			paths:  []string{"dir1/**/*.txt", "file2.txt"},
			expect: []string{"file2.txt", "dir1/file3.txt", "dir1/subdir1/file5.txt"},
		},
		{
			name:   "No matches",
			paths:  []string{"nonexistent*.txt"},
			expect: []string{},
		},
		{
			name:   "Duplicate files across patterns",
			paths:  []string{"**/*.txt", "dir1/**/*.txt"},
			expect: []string{"file1.txt", "file2.txt", "dir1/file3.txt", "dir1/subdir1/file5.txt", "dir2/file7.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parsePaths(mockFS, tt.paths)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expect, result)
			}
		})
	}
}

func TestSource_CSV(t *testing.T) {
	t.Run("default root", func(t *testing.T) {
		ctx := context.TODO()

		tmpFile, err := os.CreateTemp("./", "test_*.csv")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		writer := csv.NewWriter(tmpFile)
		require.NoError(t, writer.Write([]string{"Name", "Job", "Age"}))
		require.NoError(t, writer.Write([]string{"me", "Engineer", "1"}))
		require.NoError(t, writer.Write([]string{"you", "Doctor", "2"}))
		writer.Flush()
		require.NoError(t, writer.Error())
		require.NoError(t, tmpFile.Close())

		so := &CsvSource{Paths: []string{"test*.csv"}}
		err = so.Init(ctx, zap.NewNop().Sugar(), "./")
		require.NoError(t, err)
		indexer := NewIndexer(so, &ent.TableColumn{Random: false, LinkedColumn: "Name", LinkedContextColumns: []string{}})
		v, err := indexer.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "me", v.Value)
		require.Equal(t, map[string]any{}, v.ContextValue)
		v, err = indexer.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "you", v.Value)
		require.Equal(t, map[string]any{}, v.ContextValue)

		so = &CsvSource{Paths: []string{"test*.csv"}}
		err = so.Init(ctx, zap.NewNop().Sugar(), "./")
		require.NoError(t, err)
		indexer = NewIndexer(so, &ent.TableColumn{Random: false, LinkedColumn: "Name", LinkedContextColumns: []string{"Name", "Job"}})
		v, err = indexer.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "me", v.Value)
		require.Equal(t, map[string]any{"Name": "me", "Job": "Engineer"}, v.ContextValue)
		v, err = indexer.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "you", v.Value)
		require.Equal(t, map[string]any{"Name": "you", "Job": "Doctor"}, v.ContextValue)
	})

	t.Run("different root", func(t *testing.T) {
		ctx := context.TODO()

		err := os.Mkdir("./gogo", os.ModePerm)
		require.NoError(t, err)
		tmpFile, err := os.CreateTemp("./gogo", "test_*.csv")
		require.NoError(t, err)
		defer os.RemoveAll("./gogo")

		writer := csv.NewWriter(tmpFile)
		require.NoError(t, writer.Write([]string{"Name", "Job", "Age"}))
		require.NoError(t, writer.Write([]string{"me", "Engineer", "1"}))
		require.NoError(t, writer.Write([]string{"you", "Doctor", "2"}))
		writer.Flush()
		require.NoError(t, writer.Error())
		require.NoError(t, tmpFile.Close())

		so := &CsvSource{Paths: []string{"test*.csv"}}
		err = so.Init(ctx, zap.NewNop().Sugar(), "./gogo")
		require.NoError(t, err)
		indexer := NewIndexer(so, &ent.TableColumn{Random: false, LinkedColumn: "Name", LinkedContextColumns: []string{}})
		v, err := indexer.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "me", v.Value)
		require.Equal(t, map[string]any{}, v.ContextValue)
		v, err = indexer.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, "you", v.Value)
		require.Equal(t, map[string]any{}, v.ContextValue)
	})
}

func TestSource_GetColumns(t *testing.T) {
	t.Run("default root", func(t *testing.T) {
		ctx := context.TODO()

		tmpFile, err := os.CreateTemp("./", "test_*.csv")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		writer := csv.NewWriter(tmpFile)
		require.NoError(t, writer.Write([]string{"Name", "Job", "Age"}))
		require.NoError(t, writer.Write([]string{"me", "Engineer", "1"}))
		require.NoError(t, writer.Write([]string{"you", "Doctor", "2"}))
		writer.Flush()
		require.NoError(t, writer.Error())
		require.NoError(t, tmpFile.Close())

		so := &CsvSource{Paths: []string{"test*.csv"}}
		columns, err := so.GetColumns(ctx, zap.NewNop().Sugar(), "./")
		require.NoError(t, err)
		require.Equal(t, []string{"Name", "Job", "Age"}, columns)
	})

	t.Run("different root", func(t *testing.T) {
		ctx := context.TODO()

		err := os.Mkdir("./gogo", os.ModePerm)
		require.NoError(t, err)
		tmpFile, err := os.CreateTemp("./gogo", "test_*.csv")
		require.NoError(t, err)
		defer os.RemoveAll("./gogo")

		writer := csv.NewWriter(tmpFile)
		require.NoError(t, writer.Write([]string{"Name", "Job", "Age"}))
		require.NoError(t, writer.Write([]string{"me", "Engineer", "1"}))
		require.NoError(t, writer.Write([]string{"you", "Doctor", "2"}))
		writer.Flush()
		require.NoError(t, writer.Error())
		require.NoError(t, tmpFile.Close())

		so := &CsvSource{Paths: []string{"test*.csv"}}
		columns, err := so.GetColumns(ctx, zap.NewNop().Sugar(), "./gogo")
		require.NoError(t, err)
		require.Equal(t, []string{"Name", "Job", "Age"}, columns)
	})
}

func createTestZipCSV(t *testing.T) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	csvFile, err := zipWriter.Create("foo_bar.csv")
	require.NoError(t, err)

	writer := csv.NewWriter(csvFile)
	err = writer.Write([]string{"Name", "Age"})
	require.NoError(t, err)
	err = writer.Write([]string{"Alice", "30"})
	require.NoError(t, err)
	writer.Flush()

	err = zipWriter.Close()
	require.NoError(t, err)
	return buf.Bytes()
}

func TestSource_KaggleCSV(t *testing.T) {
	ctx := context.TODO()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockZip := createTestZipCSV(t)
	httpmock.RegisterResponder("GET", "https://www.kaggle.com/api/v1/datasets/download/foo/bar",
		httpmock.NewBytesResponder(200, mockZip).Once())

	so := &CsvSource{Paths: []string{"foo_*.csv"}, Kaggle: "foo/bar"}
	err := so.Init(ctx, zap.NewNop().Sugar(), "./")
	defer func() { _ = os.RemoveAll("tablepilot_kaggle_cache/foo--bar") }()
	require.NoError(t, err)
	indexer := NewIndexer(so, &ent.TableColumn{Random: false, LinkedColumn: "Name", LinkedContextColumns: []string{}})
	v, err := indexer.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, "Alice", v.Value)
	require.Equal(t, map[string]any{}, v.ContextValue)
}

func TestSource_KaggleGetColumns(t *testing.T) {
	ctx := context.TODO()
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockZip := createTestZipCSV(t)
	httpmock.RegisterResponder("GET", "https://www.kaggle.com/api/v1/datasets/download/foo/bar",
		httpmock.NewBytesResponder(200, mockZip).Once())

	so := &CsvSource{Paths: []string{"foo_*.csv"}, Kaggle: "foo/bar"}
	columns, err := so.GetColumns(ctx, zap.NewNop().Sugar(), "./")
	defer func() { _ = os.RemoveAll("tablepilot_kaggle_cache/foo--bar") }()
	require.NoError(t, err)
	require.Equal(t, []string{"Name", "Age"}, columns)
}
