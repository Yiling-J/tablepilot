package source

import (
	"context"
	"encoding/csv"
	"os"
	"testing"
	"testing/fstest"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err = so.Init(ctx, "./")
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
	err = so.Init(ctx, "./")
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
}
