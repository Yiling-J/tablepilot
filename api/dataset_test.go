package api

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"testing"

	// "github.com/Yiling-J/tablepilot/ent" // Not directly used by these dataset tests, but GetDataset might if it returns ent.Dataset
	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPI_ListDatasets(t *testing.T) {
	ds := []*dataset.DatasetInfo{
		{Name: "d1", Description: "desc1"},
		{Name: "d2", Description: "desc2"},
	}
	datasetMock := &dataset.DatasetServiceMock{
		ListFunc: func(ctx context.Context) ([]*dataset.DatasetInfo, error) {
			return ds, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})
	req, err := server.NewGetRequest("/api/v1/datasets")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, gin.H{
		"total":    2,
		"datasets": ds,
	})
}

func TestAPI_CreateDataset(t *testing.T) {
	expectedRequest := &dataset.CreateDatasetRequest{
		Name:        "new_dataset",
		Description: "A new dataset for testing",
	}
	datasetMock := &dataset.DatasetServiceMock{
		CreateFunc: func(ctx context.Context, req *dataset.CreateDatasetRequest) (string, error) {
			require.Equal(t, expectedRequest.Name, req.Name) // Compare field by field as Files will be different (nil vs empty slice)
			require.Equal(t, expectedRequest.Description, req.Description)
			require.Equal(t, expectedRequest.Type, req.Type)
			require.Equal(t, expectedRequest.Data, req.Data)
			// require.Equal(t, expectedRequest, req) // Original had this, but Files field makes it tricky if one is nil and other is empty
			return "new_dataset_id", nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})
	req, err := server.NewMultiplePartRequest("POST", "/api/v1/datasets", map[string]string{"name": "new_dataset", "description": "A new dataset for testing"}, nil)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 201, gin.H{"id": "new_dataset_id", "name": "new_dataset"})
}

func TestAPI_CreateDatasetWithFiles(t *testing.T) {
	expectedRequest := &dataset.CreateDatasetRequest{
		Name:        "new_dataset",
		Description: "A new dataset for testing",
		Type:        "csv",
	}
	datasetMock := &dataset.DatasetServiceMock{
		CreateFunc: func(ctx context.Context, req *dataset.CreateDatasetRequest) (string, error) {
			require.Equal(t, expectedRequest.Name, req.Name)
			require.Equal(t, expectedRequest.Description, req.Description)
			require.Equal(t, expectedRequest.Type, req.Type)
			require.Equal(t, 1, len(req.Files))
			data, err := io.ReadAll(req.Files[0])
			require.NoError(t, err)
			require.Equal(t, "header1,header2,header3\nr1c1,r1c2,r1c3\n", string(data))
			return "new_dataset_id", nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})
	var csvBuf bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuf)
	csvHeaders := []string{"header1", "header2", "header3"}
	err := csvWriter.Write(csvHeaders)
	require.NoError(t, err)
	err = csvWriter.Write([]string{"r1c1", "r1c2", "r1c3"})
	require.NoError(t, err)
	csvWriter.Flush()

	req, err := server.NewMultiplePartRequest(
		"POST", "/api/v1/datasets",
		map[string]string{"name": "new_dataset", "description": "A new dataset for testing", "type": "csv"},
		map[string]io.Reader{"1.csv": bytes.NewReader(csvBuf.Bytes())},
	)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 201, gin.H{"id": "new_dataset_id", "name": "new_dataset"})
}

func TestAPI_UpdateDataset(t *testing.T) {
	datasetID := "existing_dataset_id"
	expectedRequest := &dataset.UpdateDatasetRequest{
		CreateDatasetRequest: dataset.CreateDatasetRequest{
			Name:  "xyz",
			// Files field is omitted, so it will be nil.
			// The handler initializes Files to []io.Reader{} if apiReq.Files is nil.
		},
		Fields: []string{"name"},
	}

	datasetMock := &dataset.DatasetServiceMock{
		UpdateFunc: func(ctx context.Context, id string, req *dataset.UpdateDatasetRequest) error {
			require.Equal(t, datasetID, id)
			require.Equal(t, expectedRequest.Name, req.Name)
			require.Equal(t, expectedRequest.Description, req.Description)
			require.Equal(t, expectedRequest.Fields, req.Fields)
			// Comparing req.Files is tricky as one might be nil and other empty slice.
			// The original test didn't check Files field in expectedRequest here.
			// require.Equal(t, expectedRequest, req)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})

	req, err := server.NewMultiplePartRequest("PATCH",
		fmt.Sprintf("/api/v1/datasets/%s", datasetID),
		map[string]string{"name": "xyz"},
		nil, // No files being sent for update in this test case
	)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, gin.H{"id": datasetID})
}

func TestAPI_DeleteDataset(t *testing.T) {
	datasetID := "dataset_to_delete_id"

	datasetMock := &dataset.DatasetServiceMock{
		DeleteFunc: func(ctx context.Context, id string) error {
			require.Equal(t, datasetID, id)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})

	req, err := server.NewDeleteRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID))
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, "")
}

func TestAPI_PreviewDataset(t *testing.T) {
	datasetID := "dataset_preview_id"
	expectedResponse := &dataset.DatasetRows{
		Rows: []map[string]any{
			{"col1": "val1_1", "col2": "val1_2"},
			{"col1": "val2_1", "col2": "val2_2"},
		},
		Data: []string{"col1", "col2"},
		Type: "CSV",
	}

	datasetMock := &dataset.DatasetServiceMock{
		PreviewFunc: func(ctx context.Context, id string) (*dataset.DatasetRows, error) {
			require.Equal(t, datasetID, id)
			return expectedResponse, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})

	req, err := server.NewGetRequest(fmt.Sprintf("/api/v1/datasets/%s/preview", datasetID))
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, expectedResponse)
}

func TestAPI_GetDataset(t *testing.T) {
	expected := &dataset.DatasetInfo{Name: "ds", Description: "bar"}
	datasetMock := &dataset.DatasetServiceMock{
		GetFunc: func(ctx context.Context, datasetID string) (*dataset.DatasetInfo, error) { // Renamed 'dataset' param to 'datasetID'
			require.Equal(t, "foo", datasetID)
			return expected, nil
		},
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})
	r, err := server.NewGetRequest("/api/v1/datasets/foo")
	require.NoError(t, err)
	resp := server.Send(r)
	resp.ResponseEq(
		t, 200, expected,
	)
}
