package csvindexer

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func createTestFiles(t *testing.T, fileContents map[string][]string) []string {
	fileNames := []string{}
	for fileName, contents := range fileContents {
		f, err := os.Create(fileName)
		require.NoError(t, err)
		defer f.Close()
		for _, line := range contents {
			_, err := f.WriteString(line + "\n")
			require.NoError(t, err)
		}
		fileNames = append(fileNames, fileName)
	}
	return fileNames
}

func deleteTestFiles(t *testing.T, fileNames []string) {
	t.Helper()
	for _, fileName := range fileNames {
		err := os.Remove(fileName)
		require.NoError(t, err)
	}
}

func TestCSVIndexer_New(t *testing.T) {
	tests := []struct {
		name          string
		fileContents  map[string][]string
		wantTotal     int
		wantPositions []FileOffset
		wantErr       bool
	}{
		{
			name: "Single file",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2", "value3,value4"},
			},
			wantTotal:     2,
			wantPositions: []FileOffset{{File: 0, Offset: 16}},
			wantErr:       false,
		},
		{
			name: "Multiple files",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2", "value3,value4"},
				"test2.csv": {"header1,header2", "value5,value6"},
			},
			wantTotal:     3,
			wantPositions: []FileOffset{{File: 0, Offset: 16}, {File: 1, Offset: 16}},
			wantErr:       false,
		},
		{
			name: "Empty file",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2"},
			},
			wantTotal:     0,
			wantPositions: *new([]FileOffset),
			wantErr:       false,
		},
		{
			name: "File with only header and exactly chunksize lines",
			fileContents: map[string][]string{
				"test1.csv": append([]string{"header1,header2"}, strings.Split(strings.Repeat("value1,value2\n", chunkSize), "\n")...),
			},
			wantTotal:     chunkSize,
			wantPositions: []FileOffset{{File: 0, Offset: 16}},
			wantErr:       false,
		},
		{
			name: "File with more than chunksize lines",
			fileContents: map[string][]string{
				"test1.csv": append([]string{"header1,header2"}, strings.Split(strings.Repeat("value1,value2\n", chunkSize+1), "\n")...),
			},
			wantTotal:     chunkSize + 1,
			wantPositions: []FileOffset{{File: 0, Offset: 16}, {File: 0, Offset: 716}},
			wantErr:       false,
		},
		{
			name: "Multiple files with total lines exceeding chunksize",
			fileContents: map[string][]string{
				"test1.csv": append([]string{"header1,header2"}, strings.Split(strings.Repeat("value1,value2\n", chunkSize-1), "\n")...),
				"test2.csv": {"header1,header2", "value5,value6", "value7,value8"},
			},
			wantTotal:     chunkSize + 1,
			wantPositions: []FileOffset{{File: 0, Offset: 16}, {File: 1, Offset: 16}},
			wantErr:       false,
		},
		{
			name: "Multiple files with total lines match chunksize",
			fileContents: map[string][]string{
				"test1.csv": append([]string{"header1,header2"}, strings.Split(strings.Repeat("value1,value2\n", chunkSize), "\n")...),
				"test2.csv": {"header1,header2", "value5,value6"},
			},
			wantTotal:     chunkSize + 1,
			wantPositions: []FileOffset{{File: 0, Offset: 16}, {File: 1, Offset: 16}},
			wantErr:       false,
		},
		{
			name: "Mismatched headers",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2"},
				"test2.csv": {"header3,header4", "value3,value4"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileNames := createTestFiles(t, tt.fileContents)
			defer deleteTestFiles(t, fileNames)

			selector, err := NewCSVIndexer(fileNames)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			if selector.Total() != tt.wantTotal {
				require.FailNow(t, "NewCSVSelector() total = %v, want %v", selector.Total(), tt.wantTotal)
			}
			require.Equal(t, tt.wantPositions, selector.positions)
		})
	}
}

func TestCSVIndexer_Fetch(t *testing.T) {
	tests := []struct {
		name         string
		fileContents map[string][]string
		fetchIdx     int
		wantRecord   []string
		wantErr      bool
	}{
		{
			name: "Fetch first record",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2", "value3,value4"},
			},
			fetchIdx:   0,
			wantRecord: []string{"value1", "value2"},
			wantErr:    false,
		},
		{
			name: "Fetch second record",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2", "value3,value4"},
			},
			fetchIdx:   1,
			wantRecord: []string{"value3", "value4"},
			wantErr:    false,
		},
		{
			name: "Fetch record from second file",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2"},
				"test2.csv": {"header1,header2", "value3,value4", "value5,value6"},
			},
			fetchIdx:   2,
			wantRecord: []string{"value5", "value6"},
			wantErr:    false,
		},
		{
			name: "Fetch record at chunk boundary",
			fileContents: map[string][]string{
				"test1.csv": append([]string{"header1,header2"}, strings.Split(strings.Repeat("value1,value2\n", chunkSize), "\n")...),
			},
			fetchIdx:   chunkSize - 1,
			wantRecord: []string{"value1", "value2"},
			wantErr:    false,
		},
		{
			name: "Fetch record after chunk boundary",
			fileContents: map[string][]string{
				"test1.csv": append([]string{"header1,header2"}, strings.Split(strings.Repeat("value1,value2\n", chunkSize+1), "\n")...),
			},
			fetchIdx:   chunkSize,
			wantRecord: []string{"value1", "value2"},
			wantErr:    false,
		},
		{
			name: "Fetch record from second file after chunk boundary",
			fileContents: map[string][]string{
				"test1.csv": append([]string{"header1,header2"}, strings.Split(strings.Repeat("value1,value2\n", chunkSize-1), "\n")...),
				"test2.csv": {"header1,header2", "value3,value4", "value5,value6"},
			},
			fetchIdx:   chunkSize,
			wantRecord: []string{"value5", "value6"},
			wantErr:    false,
		},
		{
			name: "Fetch out of bounds",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2", "value3,value4"},
			},
			fetchIdx: 2,
			wantErr:  true,
		},
		{
			name: "Fetch from empty file",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2"},
			},
			fetchIdx: 0,
			wantErr:  true,
		},
		{
			name: "Fetch with negative index",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2", "value3,value4"},
			},
			fetchIdx: -1,
			wantErr:  true,
		},
		{
			name: "File becomes inaccessible during fetch",
			fileContents: map[string][]string{
				"test1.csv": {"header1,header2", "value1,value2", "value3,value4"},
			},
			fetchIdx:   1,
			wantRecord: []string{"value3", "value4"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileNames := createTestFiles(t, tt.fileContents)
			defer deleteTestFiles(t, fileNames)

			selector, err := NewCSVIndexer(fileNames)
			if err != nil {
				require.FailNow(t, "NewCSVSelector() failed: %v", err)
			}

			// Simulate file becoming inaccessible for a specific test case
			if tt.name == "File becomes inaccessible during fetch" {
				deleteTestFiles(t, fileNames)
			}

			record, err := selector.Fetch(tt.fetchIdx)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.Equal(t, tt.wantRecord, record)
		})
	}
}
