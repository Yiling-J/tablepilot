package kaggle

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func createTestZipCSV(t *testing.T) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	csvFile, err := zipWriter.Create("test.csv")
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

func TestKaggle_PrepareKaggleDataset(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockZip := createTestZipCSV(t)
	httpmock.RegisterResponder("GET", "https://www.kaggle.com/api/v1/datasets/download/jessicali9530/lfw-dataset",
		httpmock.NewBytesResponder(200, mockZip).Once())

	path, err := PrepareKaggleDataset(context.TODO(), zap.NewNop().Sugar(), "jessicali9530/lfw-dataset", "./")
	defer func() { _ = os.RemoveAll("tablepilot_kaggle_cache/jessicali9530--lfw-dataset") }()
	require.NoError(t, err)
	require.Equal(t, "tablepilot_kaggle_cache/jessicali9530--lfw-dataset", path)
	info, err := os.Stat("tablepilot_kaggle_cache/jessicali9530--lfw-dataset")
	require.NoError(t, err)
	require.Equal(t, true, info.IsDir())
	files, err := os.ReadDir("tablepilot_kaggle_cache/jessicali9530--lfw-dataset")
	require.NoError(t, err)
	var found bool
	for _, f := range files {
		if f.Name() == "test.csv" {
			found = true
			f, err := os.Open(filepath.Join(path, f.Name()))
			defer func() { _ = f.Close() }()
			require.NoError(t, err)
			csvReader := csv.NewReader(f)
			records, err := csvReader.ReadAll()
			require.NoError(t, err)
			require.Equal(t, [][]string{{"Name", "Age"}, {"Alice", "30"}}, records)
		}
	}
	require.True(t, found)

	// prepare again, not http request this time because files are cached
	_, err = PrepareKaggleDataset(context.TODO(), zap.NewNop().Sugar(), "jessicali9530/lfw-dataset", "./")
	require.NoError(t, err)
}
