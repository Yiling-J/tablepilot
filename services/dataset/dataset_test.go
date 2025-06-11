package dataset

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"testing"

	"errors"
	"path/filepath"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/dataset"
	"github.com/Yiling-J/tablepilot/ent/schema"
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
		Files: []CreateDatasetFile{{
			Name:   "file.csv",
			Reader: bytes.NewReader(csvBuf.Bytes()),
		}},
		Data: []string{"file.csv"},
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

	listDatasetName := "test-list-dataset-for-list"
	listDatasetDesc := "A list dataset for testing List"
	listDatasetData := []string{"entryA", "entryB"}
	_, err = srv.Create(ctx, &CreateDatasetRequest{
		Name:        listDatasetName,
		Description: listDatasetDesc,
		Type:        "list",
		Data:        listDatasetData,
	})
	require.NoError(t, err)

	csvDatasetName := "test-csv-dataset-for-list"
	csvDatasetDesc := "A csv dataset for testing List"
	var csvBuf bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuf)
	csvHeaders := []string{"header1", "header2", "header3"}
	_ = csvWriter.Write(csvHeaders)
	_ = csvWriter.Write([]string{"r1c1", "r1c2", "r1c3"})
	csvWriter.Flush()
	_, err = srv.Create(ctx, &CreateDatasetRequest{
		Name:        csvDatasetName,
		Description: csvDatasetDesc,
		Type:        "csv",
		Files: []CreateDatasetFile{{
			Name:   "file",
			Reader: bytes.NewReader(csvBuf.Bytes()),
		}},
		Data: []string{"file"},
	})
	require.NoError(t, err)

	listedDatasets, err = srv.List(ctx)
	require.NoError(t, err)
	require.Len(t, listedDatasets, 2)

	foundListDs := false
	foundCsvDs := false
	for _, dsInfo := range listedDatasets {
		switch dsInfo.Name {
		case listDatasetName:
			foundListDs = true
			require.Equal(t, listDatasetDesc, dsInfo.Description)
			require.Equal(t, "list", dsInfo.Type)
			require.Equal(t, len(listDatasetData), dsInfo.ValueCount)
			require.Equal(t, 0, dsInfo.ColumnCount)
		case csvDatasetName:
			foundCsvDs = true
			require.Equal(t, csvDatasetDesc, dsInfo.Description)
			require.Equal(t, "csv", dsInfo.Type)
			require.Equal(t, len(csvHeaders), dsInfo.ColumnCount)
			require.Equal(t, 1, dsInfo.ValueCount)
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
		Files: []CreateDatasetFile{
			{Name: "c2.csv", Reader: bytes.NewReader(buf.Bytes())},
			{Name: "c1.csv", Reader: bytes.NewReader(buf2.Bytes())},
		},
		Data: []string{"c2.csv", "c1.csv"},
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

	di, err := db.Dataset.Query().Where(dataset.Nanoid(ds2)).Only(t.Context())
	require.NoError(t, err)
	require.Equal(t, []string{"c2.csv", "c1.csv"}, di.Values)
	require.Equal(t, schema.FileOffset{
		File:   0,
		Total:  2,
		Offset: 14,
	}, di.Indexer.Positions[0])

	ds3, err := srv.Create(t.Context(), &CreateDatasetRequest{
		Name:        "ds3",
		Description: "dataset3",
		Type:        "csv",
		Files: []CreateDatasetFile{
			{Name: "c2.csv", Reader: bytes.NewReader(buf.Bytes())},
			{Name: "c1.csv", Reader: bytes.NewReader(buf2.Bytes())},
		},
		Data: []string{"c1.csv", "c2.csv"},
	})
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll("./dstest")
	}()

	rows, err = srv.Preview(t.Context(), ds3)
	require.NoError(t, err)
	require.Equal(t, []map[string]any{
		{"Name": "Tommy", "Age": "65", "City": "Apple"},
		{"Name": "Alice", "Age": "30", "City": "New York"},
		{"Name": "Bob", "Age": "25", "City": "San Francisco"},
	}, rows.Rows)

	di, err = db.Dataset.Query().Where(dataset.Nanoid(ds3)).Only(t.Context())
	require.NoError(t, err)
	require.Equal(t, []string{"c1.csv", "c2.csv"}, di.Values)
	require.Equal(t, schema.FileOffset{
		File:   0,
		Total:  1,
		Offset: 14,
	}, di.Indexer.Positions[0])
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

		var csv1buf bytes.Buffer
		csvW := csv.NewWriter(&csv1buf)
		_ = csvW.Write([]string{"h1", "h2"})
		_ = csvW.Write([]string{"v1", "v2"})
		csvW.Flush()
		var csv2buf bytes.Buffer
		csvW = csv.NewWriter(&csv2buf)
		_ = csvW.Write([]string{"h1", "h2"})
		_ = csvW.Write([]string{"v3", "v4"})
		csvW.Flush()
		var csv3buf bytes.Buffer
		csvW = csv.NewWriter(&csv3buf)
		_ = csvW.Write([]string{"h1", "h2"})
		_ = csvW.Write([]string{"v5", "v6"})
		csvW.Flush()

		initialHeaders := []string{"h1", "h2"}
		csvNanoid, err := srv.Create(ctx, &CreateDatasetRequest{
			Name: csvName, Description: csvDesc, Type: "csv",
			Files: []CreateDatasetFile{
				{
					Name:   "1.csv",
					Reader: bytes.NewReader(csv1buf.Bytes()),
				},
				{
					Name:   "2.csv",
					Reader: bytes.NewReader(csv2buf.Bytes()),
				},
				{
					Name:   "3.csv",
					Reader: bytes.NewReader(csv3buf.Bytes()),
				},
			},
			Data: []string{"1.csv", "2.csv", "3.csv"},
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

		// update 2.csv
		var updatedCsvBuf bytes.Buffer
		csvW = csv.NewWriter(&updatedCsvBuf)
		_ = csvW.Write([]string{"h1", "h2"})
		_ = csvW.Write([]string{"vv1", "vv2"})
		csvW.Flush()

		err = srv.Update(ctx, csvNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{
				Type: "csv",
				Files: []CreateDatasetFile{{
					Name:   "2.csv",
					Reader: bytes.NewReader(updatedCsvBuf.Bytes()),
				}},
				Data: []string{"1.csv", "2.csv", "3.csv"},
			},
			Fields: []string{"files"},
		})
		require.NoError(t, err)
		preview, err := srv.Preview(ctx, csvNanoid)
		require.NoError(t, err)
		require.Equal(t, []map[string]any{
			{"h1": "v1", "h2": "v2"},
			{"h1": "vv1", "h2": "vv2"},
			{"h1": "v5", "h2": "v6"},
		}, preview.Rows)

		// reorder
		err = srv.Update(ctx, csvNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{
				Type: "csv",
				Data: []string{"3.csv", "2.csv", "1.csv"},
			},
			Fields: []string{"files"},
		})
		require.NoError(t, err)
		preview, err = srv.Preview(ctx, csvNanoid)
		require.NoError(t, err)
		require.Equal(t, []map[string]any{
			{"h1": "v5", "h2": "v6"},
			{"h1": "vv1", "h2": "vv2"},
			{"h1": "v1", "h2": "v2"},
		}, preview.Rows)

		// remove 2.csv and add 4.csv
		var updatedCsvBuf2 bytes.Buffer
		csvW = csv.NewWriter(&updatedCsvBuf2)
		_ = csvW.Write([]string{"h1", "h2"})
		_ = csvW.Write([]string{"vv3", "vv4"})
		csvW.Flush()

		err = srv.Update(ctx, csvNanoid, &UpdateDatasetRequest{
			CreateDatasetRequest: CreateDatasetRequest{
				Type: "csv",
				Files: []CreateDatasetFile{{
					Name:   "4.csv",
					Reader: bytes.NewReader(updatedCsvBuf2.Bytes()),
				}},
				Data: []string{"1.csv", "4.csv", "3.csv"},
			},
			Fields: []string{"files"},
		})
		require.NoError(t, err)
		preview, err = srv.Preview(ctx, csvNanoid)
		require.NoError(t, err)
		require.Equal(t, []map[string]any{
			{"h1": "v1", "h2": "v2"},
			{"h1": "vv3", "h2": "vv4"},
			{"h1": "v5", "h2": "v6"},
		}, preview.Rows)
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
			Name: csvName, Description: "temp csv", Type: "csv",
			Files: []CreateDatasetFile{{
				Name:   "file",
				Reader: bytes.NewReader(csvBuf.Bytes()),
			}},
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
			Files: []CreateDatasetFile{{
				Name:   "file",
				Reader: bytes.NewReader(buf.Bytes()),
			}},
			Data: []string{"file"},
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
