package dataset

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"testing"

	"errors"
	"path/filepath"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDatasetService_Get(t *testing.T) {
	db := db.NewTestDB()
	testDir := "./dstest_get_" + uuid.NewString()
	cfg := &config.Config{
		Common: config.Common{
			DataDir: testDir,
		},
	}
	srv := NewDatasetService(db, cfg)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	ctx := t.Context()

	listDatasetName := "test-list-dataset"
	listDatasetDesc := "A list dataset for testing Get"
	listDatasetData := []string{"item1", "item2", "item3"}
	listNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
		Name:        listDatasetName,
		Description: listDatasetDesc,
		Type:        "list",
		Data:        listDatasetData,
	})
	require.NoError(t, err)
	require.NotEmpty(t, listNanoid)

	retrievedListDS, err := srv.Get(ctx, listNanoid)
	require.NoError(t, err)
	require.NotNil(t, retrievedListDS)
	require.Equal(t, listDatasetName, retrievedListDS.Name)
	require.Equal(t, listDatasetDesc, retrievedListDS.Description)
	require.Equal(t, len(listDatasetData), retrievedListDS.ValueCount)
	require.Equal(t, "list", retrievedListDS.Type) // Also check type

	nonExistentNanoid := "non-existent-nanoid"
	retrievedNonExistentDS, err := srv.Get(ctx, nonExistentNanoid)
	require.Error(t, err)
	require.Nil(t, retrievedNonExistentDS)

	csvDatasetName := "test-csv-dataset"
	csvDatasetDesc := "A csv dataset for testing Get by name"

	var csvBuf bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuf)
	csvHeaders := []string{"col1", "col2"}
	csvRow1 := []string{"val1", "val2"}
	csvRow2 := []string{"val3", "val4"}
	_ = csvWriter.Write(csvHeaders)
	_ = csvWriter.Write(csvRow1)
	_ = csvWriter.Write(csvRow2)
	csvWriter.Flush()

	csvNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
		Name:        csvDatasetName,
		Description: csvDatasetDesc,
		Type:        "csv",
		Files:       []io.Reader{bytes.NewReader(csvBuf.Bytes())},
	})
	require.NoError(t, err)
	require.NotEmpty(t, csvNanoid)

	retrievedCsvDS, err := srv.Get(ctx, csvDatasetName) // Get by name
	require.NoError(t, err)
	require.NotNil(t, retrievedCsvDS)
	require.Equal(t, csvDatasetName, retrievedCsvDS.Name)
	require.Equal(t, csvDatasetDesc, retrievedCsvDS.Description)
	require.Equal(t, len(csvHeaders), retrievedCsvDS.ColumnCount)
	require.Equal(t, "csv", retrievedCsvDS.Type) // Also check type
}

func TestDatasetService_List(t *testing.T) {
	db := db.NewTestDB()
	testDir := "./dstest_list_" + uuid.NewString()
	cfg := &config.Config{
		Common: config.Common{
			DataDir: testDir,
		},
	}
	srv := NewDatasetService(db, cfg)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	ctx := t.Context()

	listedDatasets, err := srv.List(ctx)
	require.NoError(t, err)
	require.Empty(t, listedDatasets)

	listDatasetName_l := "test-list-dataset-for-list"
	listDatasetDesc_l := "A list dataset for testing List"
	listDatasetData_l := []string{"entryA", "entryB"}
	_, err = srv.Create(ctx, &CreateDatasetRequest{
		Name:        listDatasetName_l,
		Description: listDatasetDesc_l,
		Type:        "list",
		Data:        listDatasetData_l,
	})
	require.NoError(t, err)

	csvDatasetName_l := "test-csv-dataset-for-list"
	csvDatasetDesc_l := "A csv dataset for testing List"
	var csvBuf_l bytes.Buffer
	csvWriter_l := csv.NewWriter(&csvBuf_l)
	csvHeaders_l := []string{"header1", "header2", "header3"}
	_ = csvWriter_l.Write(csvHeaders_l)
	_ = csvWriter_l.Write([]string{"r1c1", "r1c2", "r1c3"})
	csvWriter_l.Flush()
	_, err = srv.Create(ctx, &CreateDatasetRequest{
		Name:        csvDatasetName_l,
		Description: csvDatasetDesc_l,
		Type:        "csv",
		Files:       []io.Reader{bytes.NewReader(csvBuf_l.Bytes())},
	})
	require.NoError(t, err)

	listedDatasets, err = srv.List(ctx)
	require.NoError(t, err)
	require.Len(t, listedDatasets, 2)

	foundListDs := false
	foundCsvDs := false
	for _, dsInfo := range listedDatasets {
		switch dsInfo.Name {
		case listDatasetName_l:
			foundListDs = true
			require.Equal(t, listDatasetDesc_l, dsInfo.Description)
			require.Equal(t, "list", dsInfo.Type)
			require.Equal(t, len(listDatasetData_l), dsInfo.ValueCount)
			require.Equal(t, 0, dsInfo.ColumnCount)
		case csvDatasetName_l:
			foundCsvDs = true
			require.Equal(t, csvDatasetDesc_l, dsInfo.Description)
			require.Equal(t, "csv", dsInfo.Type)
			require.Equal(t, len(csvHeaders_l), dsInfo.ColumnCount)
			require.Equal(t, 0, dsInfo.ValueCount)
			require.Equal(t, []string{"header1", "header2", "header3"}, dsInfo.Columns)
		}
	}
	require.True(t, foundListDs, "List dataset was not found in the list")
	require.True(t, foundCsvDs, "CSV dataset was not found in the list")
}

func TestDatasetService_Create(t *testing.T) {
	db := db.NewTestDB()
	srv := NewDatasetService(db, &config.Config{
		Common: config.Common{
			DataDir: "./dstest",
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

func TestDatasetService_Update(t *testing.T) {
	db := db.NewTestDB()
	testDirBase := "./dstest_update_" + uuid.NewString()
	err := os.MkdirAll(testDirBase, os.ModePerm)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(testDirBase)
	}()

	ctx := t.Context()

	newServiceForTest := func(subDir string) DatasetService {
		specificTestDir := filepath.Join(testDirBase, subDir)
		err := os.MkdirAll(specificTestDir, os.ModePerm)
		require.NoError(t, err)

		cfg := &config.Config{
			Common: config.Common{
				DataDir: specificTestDir,
			},
		}
		return NewDatasetService(db, cfg)
	}

	t.Run("update existing list dataset", func(t *testing.T) {
		srv := newServiceForTest("list_update")
		listName := "initial-list"
		listDesc := "Initial list description"
		listData := []string{"a", "b", "c"}
		listNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
			Name: listName, Description: listDesc, Type: "list", Data: listData,
		})
		require.NoError(t, err)
		require.NotEmpty(t, listNanoid)

		updatedListName := "updated-list-name"
		updatedListDesc := "Updated list description"
		err = srv.Update(ctx, listNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{Name: updatedListName, Description: updatedListDesc},
			Fields:               []string{"name", "description"},
		})
		require.NoError(t, err)

		retrieved, err := srv.Get(ctx, listNanoid)
		require.NoError(t, err)
		require.Equal(t, updatedListName, retrieved.Name)
		require.Equal(t, updatedListDesc, retrieved.Description)
		require.Equal(t, len(listData), retrieved.ValueCount) // ValueCount should be original

		updatedListData := []string{"x", "y"}
		err = srv.Update(ctx, listNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{Type: "list", Data: updatedListData},
			Fields:               []string{"data"},
		})
		require.NoError(t, err)

		retrieved, err = srv.Get(ctx, listNanoid)
		require.NoError(t, err)
		require.Equal(t, len(updatedListData), retrieved.ValueCount)

		preview, err := srv.Preview(ctx, listNanoid)
		require.NoError(t, err)
		require.Equal(t, updatedListData, preview.Data)
	})

	t.Run("update existing csv dataset", func(t *testing.T) {
		srv := newServiceForTest("csv_update")
		csvName := "initial-csv"
		csvDesc := "Initial CSV description"
		var initialCsvBuf bytes.Buffer
		csvW := csv.NewWriter(&initialCsvBuf)
		_ = csvW.Write([]string{"h1", "h2"})
		_ = csvW.Write([]string{"r1v1", "r1v2"})
		csvW.Flush()
		initialHeaders := []string{"h1", "h2"}

		csvNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
			Name: csvName, Description: csvDesc, Type: "csv", Files: []io.Reader{bytes.NewReader(initialCsvBuf.Bytes())},
		})
		require.NoError(t, err)
		require.NotEmpty(t, csvNanoid)

		updatedCsvName := "updated-csv-name"
		updatedCsvDesc := "Updated CSV description"
		err = srv.Update(ctx, csvNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{Name: updatedCsvName, Description: updatedCsvDesc},
			Fields:               []string{"name", "description"},
		})
		require.NoError(t, err)

		retrieved, err := srv.Get(ctx, csvNanoid)
		require.NoError(t, err)
		require.Equal(t, updatedCsvName, retrieved.Name)
		require.Equal(t, updatedCsvDesc, retrieved.Description)
		require.Equal(t, len(initialHeaders), retrieved.ColumnCount)

		var updatedCsvBuf bytes.Buffer
		csvW = csv.NewWriter(&updatedCsvBuf)
		updatedHeaders := []string{"new_h1", "new_h2", "new_h3"}
		_ = csvW.Write(updatedHeaders)
		_ = csvW.Write([]string{"new_r1v1", "new_r1v2", "new_r1v3"})
		csvW.Flush()
		err = srv.Update(ctx, csvNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{Type: "csv", Files: []io.Reader{bytes.NewReader(updatedCsvBuf.Bytes())}},
			Fields:               []string{"files"},
		})
		require.NoError(t, err)

		retrieved, err = srv.Get(ctx, csvNanoid)
		require.NoError(t, err)
		require.Equal(t, len(updatedHeaders), retrieved.ColumnCount)

		preview, err := srv.Preview(ctx, csvNanoid)
		require.NoError(t, err)
		require.Len(t, preview.Rows, 1)
		require.Equal(t, map[string]any{"new_h1": "new_r1v1", "new_h2": "new_r1v2", "new_h3": "new_r1v3"}, preview.Rows[0])
	})

	t.Run("update non-existent dataset", func(t *testing.T) {
		srv := newServiceForTest("non_existent_update")
		err := srv.Update(ctx, "fake-nanoid", &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{Name: "anything"},
			Fields:               []string{"name"},
		})
		require.Error(t, err)
	})

	t.Run("convert list to csv not allowed", func(t *testing.T) {
		srv := newServiceForTest("list_to_csv")
		listName := "list-to-convert"
		listNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
			Name: listName, Description: "list to be csv", Type: "list", Data: []string{"q", "w", "e"},
		})
		require.NoError(t, err)

		var csvBuf bytes.Buffer
		csvW := csv.NewWriter(&csvBuf)
		csvHeaders := []string{"c1", "c2"}
		_ = csvW.Write(csvHeaders)
		_ = csvW.Write([]string{"d1", "d2"})
		csvW.Flush()
		err = srv.Update(ctx, listNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{Type: "csv", Files: []io.Reader{bytes.NewReader(csvBuf.Bytes())}, Data: nil},
			Fields:               []string{"type", "files", "data"},
		})
		require.Error(t, err)
		require.EqualError(t, err, "dataset type cannot be changed via update")

		// Verify that the dataset was NOT changed
		retrieved, err := srv.Get(ctx, listNanoid)
		require.NoError(t, err)
		require.Equal(t, "list", retrieved.Type) // Should still be list
		originalData, err := srv.Preview(ctx, listNanoid)
		require.NoError(t, err)
		require.Equal(t, []string{"q", "w", "e"}, originalData.Data)
	})

	t.Run("convert csv to list not allowed", func(t *testing.T) {
		srv := newServiceForTest("csv_to_list")
		csvName := "csv-to-convert"
		originalCsvHeaders := []string{"h_old1", "h_old2"}
		originalCsvRow := []string{"r_old1", "r_old2"}

		var initialCsvBuf bytes.Buffer
		csvW := csv.NewWriter(&initialCsvBuf)
		_ = csvW.Write(originalCsvHeaders)
		_ = csvW.Write(originalCsvRow)
		csvW.Flush()
		csvNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
			Name: csvName, Description: "csv to be list", Type: "csv", Files: []io.Reader{bytes.NewReader(initialCsvBuf.Bytes())},
		})
		require.NoError(t, err)

		listData := []string{"new_list_item1", "new_list_item2"}
		err = srv.Update(ctx, csvNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{Type: "list", Data: listData, Files: nil},
			Fields:               []string{"type", "data", "files"},
		})
		require.Error(t, err)
		require.EqualError(t, err, "dataset type cannot be changed via update")

		// Verify that the dataset was NOT changed
		retrieved, err := srv.Get(ctx, csvNanoid)
		require.NoError(t, err)
		require.Equal(t, "csv", retrieved.Type)
		require.Equal(t, len(originalCsvHeaders), retrieved.ColumnCount)
		require.Equal(t, 0, retrieved.ValueCount)

		preview, err := srv.Preview(ctx, csvNanoid)
		require.NoError(t, err)
		require.Empty(t, preview.Data)
		require.Len(t, preview.Rows, 1)
		expectedRow := make(map[string]any)
		for i, h := range originalCsvHeaders {
			expectedRow[h] = originalCsvRow[i]
		}
		require.Equal(t, expectedRow, preview.Rows[0])
	})
}

func TestDatasetService_Delete(t *testing.T) {
	db := db.NewTestDB()
	testDirBase := "./dstest_delete_" + uuid.NewString()
	defer func() {
		_ = os.RemoveAll(testDirBase)
	}()

	ctx := t.Context()

	newServiceForTest := func(subDir string) (DatasetService, string) {
		specificTestDir := filepath.Join(testDirBase, subDir)
		err := os.MkdirAll(specificTestDir, os.ModePerm)
		require.NoError(t, err)

		cfg := &config.Config{
			Common: config.Common{
				DataDir: specificTestDir,
			},
		}
		return NewDatasetService(db, cfg), specificTestDir
	}

	t.Run("delete existing list dataset", func(t *testing.T) {
		srv, _ := newServiceForTest("list_delete")
		listName := "list-to-delete"
		listNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
			Name: listName, Description: "temp list", Type: "list", Data: []string{"1", "2"},
		})
		require.NoError(t, err)
		require.NotEmpty(t, listNanoid)

		err = srv.Delete(ctx, listNanoid)
		require.NoError(t, err)

		_, err = srv.Get(ctx, listNanoid)
		require.Error(t, err)
		require.True(t, ent.IsNotFound(err) || errors.Is(err, os.ErrNotExist), "Error should indicate not found: %v", err)

		allDS, err := srv.List(ctx)
		require.NoError(t, err)
		for _, ds := range allDS {
			require.NotEqual(t, listNanoid, ds.Name, "Found deleted dataset by nanoid in Name field - unexpected")
			require.NotEqual(t, listName, ds.Name, "Found deleted dataset by original name")
		}
	})

	t.Run("delete existing csv dataset", func(t *testing.T) {
		srv, sourceDataDirUsed := newServiceForTest("csv_delete")
		csvName := "csv-to-delete"
		var csvBuf bytes.Buffer
		csvW := csv.NewWriter(&csvBuf)
		_ = csvW.Write([]string{"h1"})
		_ = csvW.Write([]string{"v1"})
		csvW.Flush()

		csvNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
			Name: csvName, Description: "temp csv", Type: "csv", Files: []io.Reader{bytes.NewReader(csvBuf.Bytes())},
		})
		require.NoError(t, err)
		require.NotEmpty(t, csvNanoid)

		// Construct expected data directory path
		csvDataPath := filepath.Join(sourceDataDirUsed, "datasets/shared", csvNanoid)
		_, err = os.Stat(csvDataPath)
		require.NoError(t, err, "CSV data directory should exist after creation")

		err = srv.Delete(ctx, csvName)
		require.NoError(t, err)

		_, err = srv.Get(ctx, csvName)
		require.Error(t, err)
		require.True(t, ent.IsNotFound(err) || errors.Is(err, os.ErrNotExist), "Error should indicate not found: %v", err)

		allDS, err := srv.List(ctx)
		require.NoError(t, err)
		for _, ds := range allDS {
			require.NotEqual(t, csvName, ds.Name)
		}

		_, err = os.Stat(csvDataPath)
		require.Error(t, err)
		require.True(t, os.IsNotExist(err), "CSV data directory should not exist after deletion")
	})

	t.Run("delete non-existent dataset", func(t *testing.T) {
		srv, _ := newServiceForTest("non_existent_delete")
		nonExistentName := "absolutely-does-not-exist-" + uuid.NewString()
		err := srv.Delete(ctx, nonExistentName)
		require.Error(t, err)
		require.True(t, ent.IsNotFound(err), "Expected a 'not found' error type, got: %v", err)
	})
}

func TestDatasetService_Preview(t *testing.T) {
	db := db.NewTestDB()
	srv := NewDatasetService(db, &config.Config{
		Common: config.Common{
			DataDir: "./dstest",
		},
	})

	t.Run("list show all options", func(t *testing.T) {
		data := []string{}
		for i := range 120 {
			data = append(data, fmt.Sprintf("%d", i))
		}
		ds1, err := srv.Create(t.Context(), &CreateDatasetRequest{
			Name:        "ds",
			Description: "dataset",
			Type:        "list",
			Data:        data,
		})
		require.NoError(t, err)
		rows, err := srv.Preview(t.Context(), ds1)
		require.NoError(t, err)
		require.Equal(t, data, rows.Data)
	})

	t.Run("csv show first 100 rows", func(t *testing.T) {
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		_ = writer.Write([]string{"Name"})
		for i := range 200 {
			_ = writer.Write([]string{fmt.Sprintf("%d", i)})
		}
		writer.Flush()

		ds2, err := srv.Create(t.Context(), &CreateDatasetRequest{
			Name:        "ds2",
			Description: "dataset2",
			Type:        "csv",
			Files:       []io.Reader{bytes.NewReader(buf.Bytes())},
		})
		require.NoError(t, err)
		defer func() {
			_ = os.RemoveAll("./dstest")
		}()

		rows, err := srv.Preview(t.Context(), ds2)
		require.NoError(t, err)
		require.Equal(t, 100, len(rows.Rows))
		expected := []map[string]any{}
		for i := range 100 {
			expected = append(expected, map[string]any{"Name": fmt.Sprintf("%d", i)})
		}
		require.Equal(t, expected, rows.Rows)
	})
}
