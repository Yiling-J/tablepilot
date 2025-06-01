package dataset

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"
	"testing"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/stretchr/testify/require"
)

func TestDatasetService_Get(t *testing.T) {}

func TestDatasetService_List(t *testing.T) {}

func TestDatasetService_Create(t *testing.T) {
	db := db.NewTestDB()
	srv := NewDatasetService(db, &config.Config{
		Common: config.Common{
			SourceDataDir: "./dstest",
		},
	})

	ds1, err := srv.Create(t.Context(), &CreateDatasetRequest{
		Name:        "ds",
		Description: "dataset",
		Type:        "list",
		Data:        []string{"foo", "bar"},
	})
	require.NoError(t, err)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"Name", "Age", "City"})
	_ = writer.Write([]string{"Alice", "30", "New York"})
	_ = writer.Write([]string{"Bob", "25", "San Francisco"})
	writer.Flush()
	var buf2 bytes.Buffer
	writer = csv.NewWriter(&buf2)
	_ = writer.Write([]string{"Name", "Age", "City"})
	_ = writer.Write([]string{"Tommy", "65", "Apple"})
	writer.Flush()

	ds2, err := srv.Create(t.Context(), &CreateDatasetRequest{
		Name:        "ds2",
		Description: "dataset2",
		Type:        "csv",
		Files:       []io.Reader{bytes.NewReader(buf.Bytes()), bytes.NewReader(buf2.Bytes())},
	})
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll("./dstest")
	}()

	rows, err := srv.Preview(t.Context(), ds1)
	require.NoError(t, err)
	require.Equal(t, []string{"foo", "bar"}, rows.Data)
	rows, err = srv.Preview(t.Context(), ds2)
	require.NoError(t, err)
	require.Equal(t, []map[string]any{
		{"Name": "Alice", "Age": "30", "City": "New York"},
		{"Name": "Bob", "Age": "25", "City": "San Francisco"},
		{"Name": "Tommy", "Age": "65", "City": "Apple"},
	}, rows.Rows)
}

func TestDatasetService_Update(t *testing.T) {}

func TestDatasetService_Delete(t *testing.T) {}

func TestDatasetService_Preview(t *testing.T) {}
