package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/ai"
	"github.com/Yiling-J/tablepilot/services/provider"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/http"

	services_dataset "github.com/Yiling-J/tablepilot/services/dataset"
	db_dataset "github.com/Yiling-J/tablepilot/ent/dataset"

	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/workflow"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
)

// --- DatasetService Mock ---
type DatasetServiceMock struct {
	CreateFunc_  func(ctx context.Context, req *services_dataset.CreateDatasetRequest) (string, error)
	GetFunc_     func(ctx context.Context, source string) (*services_dataset.DatasetInfo, error)
	ListFunc_    func(ctx context.Context) ([]*services_dataset.DatasetInfo, error)
	UpdateFunc_  func(ctx context.Context, datasetID string, req *services_dataset.CreateDatasetRequest) error
	DeleteFunc_  func(ctx context.Context, datasetID string) error
	PreviewFunc_ func(ctx context.Context, source string) (*services_dataset.DatasetRows, error)
	// Add call counters if needed, or use testify/mock
	CreateCalls  int
	GetCalls     int
	ListCalls    int
	UpdateCalls  int
	DeleteCalls  int
	PreviewCalls int
}

func (m *DatasetServiceMock) Create(ctx context.Context, req *services_dataset.CreateDatasetRequest) (string, error) {
	m.CreateCalls++
	if m.CreateFunc_ != nil {
		return m.CreateFunc_(ctx, req)
	}
	panic("CreateFunc_ not implemented")
}
func (m *DatasetServiceMock) Get(ctx context.Context, source string) (*services_dataset.DatasetInfo, error) {
	m.GetCalls++
	if m.GetFunc_ != nil {
		return m.GetFunc_(ctx, source)
	}
	panic("GetFunc_ not implemented")
}
func (m *DatasetServiceMock) List(ctx context.Context) ([]*services_dataset.DatasetInfo, error) {
	m.ListCalls++
	if m.ListFunc_ != nil {
		return m.ListFunc_()
	}
	panic("ListFunc_ not implemented")
}
func (m *DatasetServiceMock) Update(ctx context.Context, datasetID string, req *services_dataset.CreateDatasetRequest) error {
	m.UpdateCalls++
	if m.UpdateFunc_ != nil {
		return m.UpdateFunc_(ctx, datasetID, req)
	}
	panic("UpdateFunc_ not implemented")
}
func (m *DatasetServiceMock) Delete(ctx context.Context, datasetID string) error {
	m.DeleteCalls++
	if m.DeleteFunc_ != nil {
		return m.DeleteFunc_(ctx, datasetID)
	}
	panic("DeleteFunc_ not implemented")
}
func (m *DatasetServiceMock) Preview(ctx context.Context, source string) (*services_dataset.DatasetRows, error) {
	m.PreviewCalls++
	if m.PreviewFunc_ != nil {
		return m.PreviewFunc_(ctx, source)
	}
	panic("PreviewFunc_ not implemented")
}
func (m *DatasetServiceMock) First(ctx context.Context) { panic("First method not expected to be called in API tests") }
// --- End DatasetService Mock ---


func TestAPI_CreateTable(t *testing.T) {
	expectedRequest := &table.TableGenRequest{
		Name:        "recipes",
		Model:       "m1",
		Description: "all recipes",
		Columns: []table.TableGenColumn{
			{Name: "col1", Description: "desc", Type: "string", FillMode: "ai"},
		},
		Sources: []json.RawMessage{[]byte(`{"source":"s"}`)},
	}
	expectedRequest.MarkAPIRequest()
	tableMock := &table.TableServiceMock{
		CreateFunc: func(ctx context.Context, req *table.TableGenRequest) (string, error) {
			require.Equal(t, expectedRequest, req)
			return "foo", nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewPostRequest("/api/v1/tables", expectedRequest)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, map[string]string{"id": "foo"})
}

func TestAPI_UpdateTable(t *testing.T) {
	expectedRequest := &table.TableGenRequest{
		Name:        "recipes",
		Model:       "m1",
		Description: "all recipes",
		Columns: []table.TableGenColumn{
			{Name: "col1", Description: "desc", Type: "string", FillMode: "ai"},
		},
		Sources: []json.RawMessage{[]byte(`{"source":"s"}`)},
	}
	expectedRequest.MarkAPIRequest()
	tableMock := &table.TableServiceMock{
		UpdateFunc: func(ctx context.Context, tb string, req *table.TableGenRequest) (string, error) {
			require.Equal(t, "foo", tb)
			require.Equal(t, expectedRequest, req)
			return "foo", nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewPatchRequest("/api/v1/tables/foo", expectedRequest)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, map[string]string{"id": "foo"})
}

func TestAPI_Generate(t *testing.T) {
	var counter int
	mockRowGen := &table.RowsGeneratorMock{
		NextFunc: func(ctx context.Context) ([]map[string]*schema.CellValue, error) {
			defer func() { counter += 1 }()
			if counter < 2 {
				return []map[string]*schema.CellValue{
					{
						"1": &schema.CellValue{Value: cast.ToString(counter), ContextValue: map[string]any{"a": "b"}},
						"2": &schema.CellValue{Value: "t" + cast.ToString(counter)},
					},
				}, nil
			}
			return []map[string]*schema.CellValue{}, nil
		},
		TableFunc: func() *ent.TableMeta {
			return &ent.TableMeta{
				Name: "foo",
				Edges: ent.TableMetaEdges{
					Columns: []*ent.TableColumn{
						{Nanoid: "1", Name: "c1"},
						{Nanoid: "2", Name: "c2"},
					},
				},
			}
		},
	}
	tableMock := &table.TableServiceMock{
		GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
			require.Equal(t, "foo", params.Table)
			require.Equal(t, 4, params.Count)
			require.Equal(t, 2, params.Batch)
			require.Equal(t, 0.56, params.Temperature)
			require.Equal(t, "aiai", params.Model)
			return mockRowGen, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewPostRequest("/api/v1/generate/tables/foo", &table.GenerateRowsRequest{
		Batch:       2,
		Count:       4,
		Temperature: 0.56,
		Model:       "aiai",
	})
	require.NoError(t, err)
	resp := server.Send(req)
	expectedRows := []map[string]any{
		{"1": "0", "2": "t0"},
		{"1": "1", "2": "t1"},
	}
	resp.ResponseEq(t, 200, map[string]any{"data": expectedRows})
}

func TestAPI_GenerateStreaming(t *testing.T) {
	var counter int
	mockRowGen := &table.RowsGeneratorMock{
		NextFunc: func(ctx context.Context) ([]map[string]*schema.CellValue, error) {
			defer func() { counter += 1 }()
			if counter < 2 {
				return []map[string]*schema.CellValue{
					{
						"1": &schema.CellValue{Value: cast.ToString(counter), ContextValue: map[string]any{"a": "b"}},
						"2": &schema.CellValue{Value: "t" + cast.ToString(counter)},
					},
				}, nil
			}
			return []map[string]*schema.CellValue{}, nil
		},
		TableFunc: func() *ent.TableMeta {
			return &ent.TableMeta{
				Name: "foo",
				Edges: ent.TableMetaEdges{
					Columns: []*ent.TableColumn{
						{Nanoid: "1", Name: "c1"},
						{Nanoid: "2", Name: "c2"},
					},
				},
			}
		},
	}
	tableMock := &table.TableServiceMock{
		GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
			require.Equal(t, "foo", params.Table)
			require.Equal(t, 4, params.Count)
			require.Equal(t, 2, params.Batch)
			require.Equal(t, 0.56, params.Temperature)
			require.Equal(t, "aiai", params.Model)
			return mockRowGen, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewPostRequest("/api/v1/generate/tables/foo", &table.GenerateRowsRequest{
		Batch:       2,
		Count:       4,
		Temperature: 0.56,
		Model:       "aiai",
		Stream:      true,
	})
	require.NoError(t, err)
	resp := server.Send(req)
	headers := resp.response.Header()
	require.Equal(t, "text/event-stream;charset=utf-8", headers.Get("Content-Type"))
	require.Equal(t, "no-cache", headers.Get("Cache-Control"))
	require.Equal(t, "keep-alive", headers.Get("Connection"))
	require.Equal(t, "chunked", headers.Get("Transfer-Encoding"))
	expectedData := `event:message
data:{"data":[{"1":"0","2":"t0"}]}

event:message
data:{"data":[{"1":"1","2":"t1"}]}

event:message
data:[DONE]

{"data":[]}`
	require.Equal(
		t, expectedData,
		resp.response.Body.String(),
	)
}

func TestAPI_Rows(t *testing.T) {
	tableMock := &table.TableServiceMock{
		RowsFunc: func(ctx context.Context, name string) (*table.Rows, error) {
			require.Equal(t, "foo", name)
			return &table.Rows{
				Columns: []*ent.TableColumn{
					{Nanoid: "1", Name: "c1"},
					{Nanoid: "2", Name: "c2"},
				},
				Rows: []*ent.TableRow{
					{Nanoid: "r1", Cells: []*schema.CellValue{{Value: "a1"}, {Value: "b1"}}},
					{Nanoid: "r2", Cells: []*schema.CellValue{{Value: "a2"}, {Value: "b2"}}},
					{Nanoid: "r3", Cells: []*schema.CellValue{{Value: "a3"}, {Value: "b3"}}},
				},
			}, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewGetRequest("/api/v1/tables/foo/rows")
	require.NoError(t, err)
	resp := server.Send(req)
	expectedRows := []map[string]any{
		{"__id__": "r1", "1": "a1", "2": "b1"},
		{"__id__": "r2", "1": "a2", "2": "b2"},
		{"__id__": "r3", "1": "a3", "2": "b3"},
	}
	resp.ResponseEq(t, 200, map[string]any{"data": expectedRows, "total": 3})
}

func TestAPI_ListTables(t *testing.T) {
	expectedResponse := &table.ListTablesResponse{
		Total: 2,
		Tables: []table.TableInfo{
			{ID: "1", Name: "t1", Description: "d1"},
			{ID: "2", Name: "t2", Description: "d2"},
		},
	}
	tableMock := &table.TableServiceMock{
		ListTablesFunc: func(ctx context.Context) (*table.ListTablesResponse, error) {
			return expectedResponse, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewGetRequest("/api/v1/tables")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, expectedResponse)
}

func TestAPI_Describe(t *testing.T) {
	columns := []table.TableColumnInfo{
		{ID: "1", Name: "c1", Type: "string", FillMode: "ai", Description: "d1"},
		{ID: "2", Name: "c2", Type: "int", FillMode: "bi", Description: "d2"},
	}
	expected := &table.TableInfo{
		ID:          "t1",
		Name:        "tb",
		Description: "td",
		Model:       "tm",
		Columns:     columns,
	}
	tableMock := &table.TableServiceMock{
		GetTableDetailFunc: func(ctx context.Context, name string) (*table.TableInfo, error) {
			require.Equal(t, "foo", name)
			return expected, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewGetRequest("/api/v1/tables/foo")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, expected)
}

func TestAPI_Delete(t *testing.T) {
	tableMock := &table.TableServiceMock{
		DeleteFunc: func(ctx context.Context, table string) (int, error) {
			require.Equal(t, "foo", table)
			return 1, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewDeleteRequest("/api/v1/tables/foo")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 204, nil)
}

func TestAPI_Truncate(t *testing.T) {
	tableMock := &table.TableServiceMock{
		TruncateFunc: func(ctx context.Context, table string) (int, error) {
			require.Equal(t, "foo", table)
			return 5, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewPostRequest("/api/v1/tables/foo/truncate", nil)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, map[string]any{"removed": 5})
}

func TestAPI_ListModels(t *testing.T) {
	expected := &ai.ModelList{
		DefaultModel: "foo",
		Models:       []ai.ModelListItem{{Name: "foo"}, {Name: "bar"}},
	}
	aiMock := &ai.AiServiceMock{
		ListModelsFunc: func(ctx context.Context) *ai.ModelList {
			return expected
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.AIService = aiMock
	})
	req, err := server.NewGetRequest("/api/v1/models")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, expected)
}

func TestAPI_Autofill(t *testing.T) {
	for _, emptyContextColumns := range []bool{false, true} {
		t.Run(fmt.Sprintf("empty context columns %v", emptyContextColumns), func(t *testing.T) {
			var counter int
			mockRowGen := &table.RowsGeneratorMock{
				NextFunc: func(ctx context.Context) ([]map[string]*schema.CellValue, error) {
					defer func() { counter += 1 }()
					if counter < 2 {
						return []map[string]*schema.CellValue{
							{
								"1": &schema.CellValue{Value: cast.ToString(counter), ContextValue: map[string]any{"a": "b"}},
								"2": &schema.CellValue{Value: "t" + cast.ToString(counter)},
							},
						}, nil
					}
					return []map[string]*schema.CellValue{}, nil
				},
				TableFunc: func() *ent.TableMeta {
					return &ent.TableMeta{
						Name: "foo",
						Edges: ent.TableMetaEdges{
							Columns: []*ent.TableColumn{
								{Nanoid: "1", Name: "c1"},
								{Nanoid: "2", Name: "c2"},
							},
						},
					}
				},
			}
			tableMock := &table.TableServiceMock{
				GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
					require.Equal(t, "foo", params.Table)
					require.Equal(t, 4, params.Count)
					require.Equal(t, 2, params.Batch)
					require.Equal(t, 0.56, params.Temperature)
					require.Equal(t, "aiai", params.Model)
					require.Equal(t, true, params.Autofill.Enable)
					require.Equal(t, 3, params.Autofill.Offset)
					require.Equal(t, []string{"c1", "c2"}, params.Autofill.Columns)
					if emptyContextColumns {
						require.Equal(t, 1, len(params.Autofill.ContextColumns))
						require.Equal(t, 36, len(params.Autofill.ContextColumns[0]))
					} else {
						require.Equal(t, []string{"c3", "c4"}, params.Autofill.ContextColumns)
					}
					return mockRowGen, nil
				},
			}
			server := NewTestServer(t, func(s *services.Backend) {
				s.TableService = tableMock
			})
			contextColumns := []string{"c3", "c4"}
			if emptyContextColumns {
				contextColumns = []string{}
			}
			req, err := server.NewPostRequest("/api/v1/autofill/tables/foo", &table.GenerateRowsRequest{
				Batch:       2,
				Count:       4,
				Temperature: 0.56,
				Model:       "aiai",
				Autofill: table.AutofillRequest{
					Offset:         3,
					Columns:        []string{"c1", "c2"},
					ContextColumns: contextColumns,
				},
			})
			require.NoError(t, err)
			resp := server.Send(req)
			expectedRows := []map[string]any{
				{"1": "0", "2": "t0"},
				{"1": "1", "2": "t1"},
			}
			resp.ResponseEq(t, 200, map[string]any{"data": expectedRows})
		})
	}
}

func TestAPI_CreateRows(t *testing.T) {
	expectedRequest := &table.CreateRowsRequest{
		Rows: []map[string]any{{"name": "foo"}},
	}
	tableMock := &table.TableServiceMock{
		CreateRowsFunc: func(ctx context.Context, table string, rows []map[string]any) error {
			require.Equal(t, expectedRequest.Rows, rows)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewPostRequest("/api/v1/tables/foo", expectedRequest)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, "")
}

func TestAPI_Sources(t *testing.T) {
	sources := []*table.SharedSource{
		{Name: "s1", Columns: []string{"c1"}, Data: json.RawMessage(`{"foo": "bar"}`)},
	}
	tableMock := &table.TableServiceMock{
		SharedSourcesFunc: func(ctx context.Context) []*table.SharedSource {
			return sources
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewGetRequest("/api/v1/sources")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, map[string]any{"sources": sources})
}

func TestAPI_GetTableSchema(t *testing.T) {
	tableMock := &table.TableServiceMock{
		GetTableSchemaFunc: func(ctx context.Context, tb string) (*table.TableGenRequest, error) {
			require.Equal(t, "foo", tb)
			return &table.TableGenRequest{Name: "bar"}, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewGetRequest("/api/v1/tables/foo/schema")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, map[string]any{
			"name": "bar", "model": "", "description": "", "columns": nil, "sources": nil,
		},
	)
}

func TestAPI_GetProviders(t *testing.T) {
	providers := []provider.Provider{
		{ID: 1, Name: "p"},
	}
	providerMock := &provider.ProviderServiceMock{
		ListProvidersFunc: func(ctx context.Context) ([]provider.Provider, error) {
			return providers, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewGetRequest("/api/v1/providers")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, providers,
	)
}

func TestAPI_CreateProvider(t *testing.T) {
	pr := provider.Provider{
		Name: "p",
	}
	providerMock := &provider.ProviderServiceMock{
		CreateProviderFunc: func(ctx context.Context, provider provider.Provider) error {
			require.Equal(t, pr, provider)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewPostRequest("/api/v1/providers", pr)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, "",
	)
}


// --- Dataset API Tests ---

func TestAPI_CreateDataset_ListType_Success(t *testing.T) {
	datasetMock := &DatasetServiceMock{}
	datasetMock.CreateFunc_ = func(ctx context.Context, req *services_dataset.CreateDatasetRequest) (string, error) {
		require.Equal(t, "My List Dataset", req.Name)
		require.Equal(t, "A simple list dataset", req.Description)
		require.Equal(t, "list", req.Type)
		require.Equal(t, []string{"item1", "item2"}, req.Data)
		require.Nil(t, req.Files)
		return "list_ds_nanoid", nil
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})

	payload := services_dataset.DatasetAPIRequest{
		Name:        "My List Dataset",
		Description: "A simple list dataset",
		Type:        "list",
		Data:        []string{"item1", "item2"},
	}

	req, err := server.NewPostRequest("/api/v1/datasets", payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusCreated, gin.H{"id": "list_ds_nanoid", "name": "My List Dataset"})
	require.Equal(t, 1, datasetMock.CreateCalls)
}

func TestAPI_CreateDataset_CSVType_Success(t *testing.T) {
	csvContent := "headerA,headerB\nvalA1,valB1"
	base64CSV := base64.StdEncoding.EncodeToString([]byte(csvContent))

	datasetMock := &DatasetServiceMock{}
	datasetMock.CreateFunc_ = func(ctx context.Context, req *services_dataset.CreateDatasetRequest) (string, error) {
		require.Equal(t, "My CSV Dataset", req.Name)
		require.Equal(t, "csv", req.Type)
		require.NotNil(t, req.Files)
		require.Len(t, req.Files, 1)

		fileContent, readErr := io.ReadAll(req.Files[0])
		require.NoError(t, readErr)
		require.Equal(t, csvContent, string(fileContent))

		return "csv_ds_nanoid", nil
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})

	payload := services_dataset.DatasetAPIRequest{
		Name:        "My CSV Dataset",
		Description: "A simple csv dataset",
		Type:        "csv",
		Files:       []string{base64CSV},
		FileNames:   []string{"test.csv"},
	}

	req, err := server.NewPostRequest("/api/v1/datasets", payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusCreated, gin.H{"id": "csv_ds_nanoid", "name": "My CSV Dataset"})
	require.Equal(t, 1, datasetMock.CreateCalls)
}


func TestAPI_CreateDataset_BadRequest_InvalidJSON(t *testing.T) {
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = &DatasetServiceMock{} // Will not be called
	})

	// Manually create a request with malformed JSON
	rawReq, _ := http.NewRequest(http.MethodPost, "/api/v1/datasets", bytes.NewBufferString(`{"name": "test", "type": "list",`)) // Malformed
	rawReq.Header.Set("Content-Type", "application/json")

	resp := server.Send(rawReq)
	require.Equal(t, http.StatusBadRequest, resp.Response().Code)
	// Error message from Gin's binding is not always simple to assert precisely, check for non-empty body
	require.NotEmpty(t, resp.Response().Body.String())
}


func TestAPI_CreateDataset_BadRequest_MissingRequiredField(t *testing.T) {
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = &DatasetServiceMock{}
	})

	payload := services_dataset.DatasetAPIRequest{
		// Name is missing, Type is missing (both required by binding tag in DatasetAPIRequest)
		Description: "Missing name and type",
	}
	req, err := server.NewPostRequest("/api/v1/datasets", payload)
	require.NoError(t, err)
	resp := server.Send(req)

	require.Equal(t, http.StatusBadRequest, resp.Response().Code)
	// Check if body contains something about validation error
	bodyBytes, _ := io.ReadAll(resp.Response().Body)
	require.Contains(t, string(bodyBytes), "binding:required")
}


func TestAPI_CreateDataset_CSVType_NoFiles(t *testing.T) {
	datasetMock := &DatasetServiceMock{} // Should not be called
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})

	payload := services_dataset.DatasetAPIRequest{
		Name:        "My CSV Dataset No File",
		Type:        "csv",
		Files:       []string{}, // Empty files array
	}

	req, err := server.NewPostRequest("/api/v1/datasets", payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusBadRequest, "at least one file is required for CSV dataset type")
	require.Equal(t, 0, datasetMock.CreateCalls)
}


func TestAPI_CreateDataset_ServiceError(t *testing.T) {
	datasetMock := &DatasetServiceMock{}
	datasetMock.CreateFunc_ = func(ctx context.Context, req *services_dataset.CreateDatasetRequest) (string, error) {
		return "", errors.New("internal service error")
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.DatasetService = datasetMock
	})
	payload := services_dataset.DatasetAPIRequest{Name: "Error DS", Type: "list", Data: []string{"a"}}
	req, err := server.NewPostRequest("/api/v1/datasets", payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusInternalServerError, "failed to create dataset: internal service error")
	require.Equal(t, 1, datasetMock.CreateCalls)
}

func TestAPI_GetDataset_Success(t *testing.T) {
	expectedDataset := &services_dataset.DatasetInfo{
		Name:        "Test Dataset",
		Description: "Description",
		Type:        "list",
		ColumnCount: 0,
		ValueCount:  3,
	}
	datasetMock := &DatasetServiceMock{}
	datasetMock.GetFunc_ = func(ctx context.Context, source string) (*services_dataset.DatasetInfo, error) {
		require.Equal(t, "ds_123", source)
		return expectedDataset, nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest("/api/v1/datasets/ds_123")
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusOK, expectedDataset)
	require.Equal(t, 1, datasetMock.GetCalls)
}

func TestAPI_GetDataset_NotFound(t *testing.T) {
	datasetMock := &DatasetServiceMock{}
	datasetMock.GetFunc_ = func(ctx context.Context, source string) (*services_dataset.DatasetInfo, error) {
		require.Equal(t, "unknown_ds", source)
		return nil, &ent.NotFoundError{} // Simulate ent.NotFoundError
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest("/api/v1/datasets/unknown_ds")
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusNotFound, "dataset 'unknown_ds' not found: ent: not found")
	require.Equal(t, 1, datasetMock.GetCalls)
}

func TestAPI_ListDatasets_Success_HasData(t *testing.T) {
	expectedDatasets := []*services_dataset.DatasetInfo{
		{Name: "DS1", Type: "list", ValueCount: 2},
		{Name: "DS2", Type: "csv", ColumnCount: 3},
	}
	datasetMock := &DatasetServiceMock{}
	datasetMock.ListFunc_ = func(ctx context.Context) ([]*services_dataset.DatasetInfo, error) {
		return expectedDatasets, nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest("/api/v1/datasets")
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusOK, expectedDatasets)
	require.Equal(t, 1, datasetMock.ListCalls)
}

func TestAPI_ListDatasets_Success_NoData(t *testing.T) {
	expectedDatasets := []*services_dataset.DatasetInfo{} // Empty slice
	datasetMock := &DatasetServiceMock{}
	datasetMock.ListFunc_ = func(ctx context.Context) ([]*services_dataset.DatasetInfo, error) {
		return expectedDatasets, nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest("/api/v1/datasets")
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusOK, expectedDatasets)
	require.Equal(t, 1, datasetMock.ListCalls)
}

func TestAPI_ListDatasets_ServiceError(t *testing.T) {
	datasetMock := &DatasetServiceMock{}
	datasetMock.ListFunc_ = func(ctx context.Context) ([]*services_dataset.DatasetInfo, error) {
		return nil, errors.New("service error listing datasets")
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest("/api/v1/datasets")
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusInternalServerError, "failed to list datasets: service error listing datasets")
	require.Equal(t, 1, datasetMock.ListCalls)
}

func TestAPI_UpdateDataset_ListType_Success(t *testing.T) {
	datasetID := "listds_to_update"
	datasetMock := &DatasetServiceMock{}
	datasetMock.UpdateFunc_ = func(ctx context.Context, id string, req *services_dataset.CreateDatasetRequest) error {
		require.Equal(t, datasetID, id)
		require.Equal(t, "Updated List Name", req.Name)
		require.Equal(t, "Updated Description", req.Description)
		require.Equal(t, "list", req.Type)
		require.Equal(t, []string{"updated_item"}, req.Data)
		return nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	payload := services_dataset.DatasetAPIRequest{
		Name:        "Updated List Name",
		Description: "Updated Description",
		Type:        "list",
		Data:        []string{"updated_item"},
	}
	req, err := server.NewPatchRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID), payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusOK, gin.H{"message": fmt.Sprintf("Dataset '%s' updated successfully.", datasetID)})
	require.Equal(t, 1, datasetMock.UpdateCalls)
}

func TestAPI_UpdateDataset_CSVType_WithNewFile(t *testing.T) {
	datasetID := "csvds_to_update"
	newCSVContent := "new_header\nnew_value"
	base64NewCSV := base64.StdEncoding.EncodeToString([]byte(newCSVContent))

	datasetMock := &DatasetServiceMock{}
	datasetMock.UpdateFunc_ = func(ctx context.Context, id string, req *services_dataset.CreateDatasetRequest) error {
		require.Equal(t, datasetID, id)
		require.Equal(t, "Updated CSV Name", req.Name)
		require.Equal(t, "csv", req.Type)
		require.NotNil(t, req.Files)
		require.Len(t, req.Files, 1)
		fileContent, _ := io.ReadAll(req.Files[0])
		require.Equal(t, newCSVContent, string(fileContent))
		return nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	payload := services_dataset.DatasetAPIRequest{
		Name:  "Updated CSV Name",
		Type:  "csv",
		Files: []string{base64NewCSV},
	}
	req, err := server.NewPatchRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID), payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusOK, gin.H{"message": fmt.Sprintf("Dataset '%s' updated successfully.", datasetID)})
	require.Equal(t, 1, datasetMock.UpdateCalls)
}

func TestAPI_UpdateDataset_CSVType_ClearFiles(t *testing.T) {
	datasetID := "csvds_clear_files"
	datasetMock := &DatasetServiceMock{}
	datasetMock.UpdateFunc_ = func(ctx context.Context, id string, req *services_dataset.CreateDatasetRequest) error {
		require.Equal(t, datasetID, id)
		require.Equal(t, "CSV Clear Files", req.Name)
		require.Equal(t, "csv", req.Type)
		// Service expects empty slice if files are to be cleared by API when 'Files' is an empty array in payload
		require.NotNil(t, req.Files)
		require.Len(t, req.Files, 0)
		return nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	// Payload with "Files: []" explicitly asking to clear/replace with no files.
	payload := services_dataset.DatasetAPIRequest{
		Name:  "CSV Clear Files",
		Type:  "csv",
		Files: []string{}, // Explicitly empty
	}
	req, err := server.NewPatchRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID), payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusOK, gin.H{"message": fmt.Sprintf("Dataset '%s' updated successfully.", datasetID)})
	require.Equal(t, 1, datasetMock.UpdateCalls)
}


func TestAPI_UpdateDataset_NotFound(t *testing.T) {
	datasetID := "non_existent_ds"
	datasetMock := &DatasetServiceMock{}
	datasetMock.UpdateFunc_ = func(ctx context.Context, id string, req *services_dataset.CreateDatasetRequest) error {
		return &ent.NotFoundError{} // Simulate not found from service
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	payload := services_dataset.DatasetAPIRequest{Name: "Trying to update", Type: "list"}
	req, err := server.NewPatchRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID), payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusNotFound, fmt.Sprintf("dataset '%s' not found for update: ent: not found", datasetID))
	require.Equal(t, 1, datasetMock.UpdateCalls)
}

func TestAPI_UpdateDataset_ServiceError(t *testing.T) {
	datasetID := "ds_with_service_error"
	datasetMock := &DatasetServiceMock{}
	datasetMock.UpdateFunc_ = func(ctx context.Context, id string, req *services_dataset.CreateDatasetRequest) error {
		return errors.New("internal service update error")
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	payload := services_dataset.DatasetAPIRequest{Name: "Error Update", Type: "list"}
	req, err := server.NewPatchRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID), payload)
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusInternalServerError, fmt.Sprintf("failed to update dataset '%s': internal service update error", datasetID))
	require.Equal(t, 1, datasetMock.UpdateCalls)
}

func TestAPI_DeleteDataset_Success(t *testing.T) {
	datasetID := "ds_to_delete"
	datasetMock := &DatasetServiceMock{}
	datasetMock.DeleteFunc_ = func(ctx context.Context, id string) error {
		require.Equal(t, datasetID, id)
		return nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewDeleteRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID))
	require.NoError(t, err)
	resp := server.Send(req)

	require.Equal(t, http.StatusNoContent, resp.Response().Code)
	require.Equal(t, 1, datasetMock.DeleteCalls)
}

func TestAPI_DeleteDataset_NotFound(t *testing.T) {
	datasetID := "ds_not_found_for_delete"
	datasetMock := &DatasetServiceMock{}
	datasetMock.DeleteFunc_ = func(ctx context.Context, id string) error {
		return &ent.NotFoundError{} // Simulate not found
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewDeleteRequest(fmt.Sprintf("/api/v1/datasets/%s", datasetID))
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusNotFound, fmt.Sprintf("dataset '%s' not found for deletion: ent: not found", datasetID))
	require.Equal(t, 1, datasetMock.DeleteCalls)
}

func TestAPI_PreviewDataset_ListType_Success(t *testing.T) {
	datasetID := "list_ds_for_preview"
	expectedPreview := &services_dataset.DatasetRows{
		Type: db_dataset.TypeList,
		Data: []string{"prev_item1", "prev_item2"},
	}
	datasetMock := &DatasetServiceMock{}
	datasetMock.PreviewFunc_ = func(ctx context.Context, source string) (*services_dataset.DatasetRows, error) {
		require.Equal(t, datasetID, source)
		return expectedPreview, nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest(fmt.Sprintf("/api/v1/datasets/%s/preview", datasetID))
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusOK, expectedPreview)
	require.Equal(t, 1, datasetMock.PreviewCalls)
}

func TestAPI_PreviewDataset_CSVType_Success(t *testing.T) {
	datasetID := "csv_ds_for_preview"
	expectedPreview := &services_dataset.DatasetRows{
		Type: db_dataset.TypeCSV,
		Rows: []map[string]any{
			{"h1": "v1a", "h2": "v1b"},
			{"h1": "v2a", "h2": "v2b"},
		},
	}
	datasetMock := &DatasetServiceMock{}
	datasetMock.PreviewFunc_ = func(ctx context.Context, source string) (*services_dataset.DatasetRows, error) {
		require.Equal(t, datasetID, source)
		return expectedPreview, nil
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest(fmt.Sprintf("/api/v1/datasets/%s/preview", datasetID))
	require.NoError(t, err)
	resp := server.Send(req)

	// Need to marshal expectedPreview to map[string]any for ResponseEq if it contains non-basic types not handled by default.
	// However, DatasetRows should be directly marshallable by gin/json.
	var actualPreview services_dataset.DatasetRows
	err = json.Unmarshal(resp.Response().Body.Bytes(), &actualPreview)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.Response().Code)
	require.Equal(t, expectedPreview.Type, actualPreview.Type)
	require.Equal(t, expectedPreview.Rows, actualPreview.Rows) // This works if map order is same or using a comparison that ignores order for maps if necessary.
	require.Equal(t, 1, datasetMock.PreviewCalls)
}

func TestAPI_PreviewDataset_NotFound(t *testing.T) {
	datasetID := "preview_ds_not_found"
	datasetMock := &DatasetServiceMock{}
	datasetMock.PreviewFunc_ = func(ctx context.Context, source string) (*services_dataset.DatasetRows, error) {
		return nil, &ent.NotFoundError{}
	}
	server := NewTestServer(t, func(s *services.Backend) { s.DatasetService = datasetMock })

	req, err := server.NewGetRequest(fmt.Sprintf("/api/v1/datasets/%s/preview", datasetID))
	require.NoError(t, err)
	resp := server.Send(req)

	resp.ResponseEq(t, http.StatusNotFound, fmt.Sprintf("dataset '%s' not found for preview: ent: not found", datasetID))
	require.Equal(t, 1, datasetMock.PreviewCalls)
}

func TestAPI_UpdateProvider(t *testing.T) {
	pr := provider.Provider{
		Name: "p",
	}
	providerMock := &provider.ProviderServiceMock{
		UpdateProviderFunc: func(ctx context.Context, id int, provider provider.Provider) error {
			require.Equal(t, 2, id)
			require.Equal(t, pr, provider)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewPatchRequest("/api/v1/providers/2", pr)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, "",
	)
}

func TestAPI_DeleteProvider(t *testing.T) {
	providerMock := &provider.ProviderServiceMock{
		DeleteProviderFunc: func(ctx context.Context, id int) error {
			require.Equal(t, 2, id)
			return nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.ProviderService = providerMock
	})
	req, err := server.NewDeleteRequest("/api/v1/providers/2")
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(
		t, 200, "",
	)
}

func TestAPI_Regenerate(t *testing.T) {
	var counter int
	mockRowGen := &table.RowsGeneratorMock{
		NextFunc: func(ctx context.Context) ([]map[string]*schema.CellValue, error) {
			defer func() { counter += 1 }()
			if counter < 2 {
				return []map[string]*schema.CellValue{
					{
						"1": &schema.CellValue{Value: cast.ToString(counter), ContextValue: map[string]any{"a": "b"}},
						"2": &schema.CellValue{Value: "t" + cast.ToString(counter)},
					},
				}, nil
			}
			return []map[string]*schema.CellValue{}, nil
		},
		TableFunc: func() *ent.TableMeta {
			return &ent.TableMeta{
				Name: "foo",
				Edges: ent.TableMetaEdges{
					Columns: []*ent.TableColumn{
						{Nanoid: "1", Name: "c1"},
						{Nanoid: "2", Name: "c2"},
					},
				},
			}
		},
	}
	tableMock := &table.TableServiceMock{
		GetTableDetailFunc: func(ctx context.Context, tb string) (*table.TableInfo, error) {
			require.Equal(t, "foo", tb)
			return &table.TableInfo{
				Columns: []table.TableColumnInfo{
					{ID: "cc1"}, {ID: "cc2"},
				},
			}, nil
		},
		GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
			require.Equal(t, "foo", params.Table)
			require.Equal(t, 4, params.Count)
			require.Equal(t, 2, params.Batch)
			require.Equal(t, 0.56, params.Temperature)
			require.Equal(t, "aiai", params.Model)
			require.Equal(t, true, params.Autofill.Enable)
			require.Equal(t, "foobar", params.Autofill.Prompt)
			require.Equal(t, []string{"c1", "c2"}, params.Autofill.Columns)
			require.Equal(t, []string{"cc1", "cc2"}, params.Autofill.ContextColumns)
			require.Equal(t, []string{"r1", "r2", "r3", "r4"}, params.Autofill.Rows)
			return mockRowGen, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	req, err := server.NewPostRequest("/api/v1/regenerate/tables/foo", &table.GenerateRowsRequest{
		Batch:       2,
		Temperature: 0.56,
		Model:       "aiai",
		Autofill: table.AutofillRequest{
			Columns: []string{"c1", "c2"},
			Rows:    []string{"r1", "r2", "r3", "r4"},
			Prompt:  "foobar",
		},
	})
	require.NoError(t, err)
	resp := server.Send(req)
	expectedRows := []map[string]any{
		{"1": "0", "2": "t0"},
		{"1": "1", "2": "t1"},
	}

	resp.ResponseEq(t, 200, map[string]any{"data": expectedRows})
}

func TestAPI_ImageImport(t *testing.T) {
	req := table.ImportRequest{
		Data:   []byte("data"),
		Prompt: "pm",
		Model:  "mm",
	}

	tableMock := &table.TableServiceMock{
		ImportImageFunc: func(ctx context.Context, request table.ImportRequest) (string, error) {
			require.Equal(t, req, request)
			return "foobar", nil
		},
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = tableMock
	})
	r, err := server.NewPostRequest("/api/v1/image_import/tables", req)
	require.NoError(t, err)
	resp := server.Send(r)
	resp.ResponseEq(
		t, 200, gin.H{"id": "foobar"},
	)
}

func TestAPI_ListWorkflows(t *testing.T) {
	w := []*ent.Workflow{{Nanoid: "i1", Name: "w1", Description: "dw1"}, {Nanoid: "i2", Name: "w2", Description: "dw2"}}
	workflowMock := &workflow.WorkflowServiceMock{
		ListFunc: func(ctx context.Context) ([]*ent.Workflow, error) {
			return w, nil
		},
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.WorkflowService = workflowMock
	})
	r, err := server.NewGetRequest("/api/v1/workflows")
	require.NoError(t, err)
	resp := server.Send(r)
	ws := []workflow.WorkflowSimple{
		{ID: "i1", Name: "w1", Description: "dw1"}, {ID: "i2", Name: "w2", Description: "dw2"},
	}
	resp.ResponseEq(
		t, 200, gin.H{"total": 2, "workflows": ws},
	)
}

func TestAPI_GetWorkflow(t *testing.T) {
	w := &ent.Workflow{
		Nanoid: "i1", Name: "w1", Description: "dw1",
		Variables: []schema.WorkflowVariable{{Name: "v1"}},
		Steps:     []schema.WorkflowStep{{Type: schema.WorkflowStepTypeAutofill}},
	}
	workflowMock := &workflow.WorkflowServiceMock{
		GetFunc: func(ctx context.Context, wf string) (*ent.Workflow, error) {
			require.Equal(t, "foo", wf)
			return w, nil
		},
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.WorkflowService = workflowMock
	})
	r, err := server.NewGetRequest("/api/v1/workflows/foo")
	require.NoError(t, err)
	resp := server.Send(r)
	ws := workflow.Workflow{
		ID: "i1", Name: "w1", Description: "dw1",
		Variables: []schema.WorkflowVariable{{Name: "v1"}},
		Steps:     []schema.WorkflowStep{{Type: schema.WorkflowStepTypeAutofill}},
	}
	resp.ResponseEq(
		t, 200, ws,
	)
}

func TestAPI_RunWorkflow(t *testing.T) {
	var counter int
	mockRunner := &workflow.RunnerMock{
		NextFunc: func(ctx context.Context) (*workflow.WorkflowStepResult, error) {
			if counter > 0 {
				return nil, nil
			}
			counter += 1
			return &workflow.WorkflowStepResult{
				Action:  workflow.WorkflowActionShowMessage,
				Message: "foobar",
			}, nil
		},
	}
	mockWorkflow := &workflow.WorkflowServiceMock{
		StartFunc: func(ctx context.Context, id string, request workflow.StartWorklfowRequest) (workflow.Runner, error) {
			require.Equal(t, request, workflow.StartWorklfowRequest{
				Variables:   map[string]any{"a": "b"},
				Model:       "aiai",
				ImageModel:  "aiia",
				Temperature: 0.56,
			})
			return mockRunner, nil
		},
		GetFunc: func(ctx context.Context, wf string) (*ent.Workflow, error) {
			require.Equal(t, "foo", wf)
			return &ent.Workflow{}, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.WorkflowService = mockWorkflow
	})
	req, err := server.NewPostRequest("/api/v1/workflows/foo/run", &workflow.StartWorklfowRequest{
		Temperature: 0.56,
		Model:       "aiai",
		ImageModel:  "aiia",
		Variables:   map[string]any{"a": "b"},
	})
	require.NoError(t, err)
	resp := server.Send(req)
	headers := resp.response.Header()
	require.Equal(t, "text/event-stream;charset=utf-8", headers.Get("Content-Type"))
	require.Equal(t, "no-cache", headers.Get("Cache-Control"))
	require.Equal(t, "keep-alive", headers.Get("Connection"))
	require.Equal(t, "chunked", headers.Get("Transfer-Encoding"))
	expectedData := `event:message
data:{"data":"foobar","type":"MESSAGE"}

event:message
data:{"type":"STEP_DONE"}

event:message
data:{"type":"WORKFLOW_DONE"}

event:message
data:[DONE]

""`
	require.Equal(
		t, expectedData,
		resp.response.Body.String(),
	)
}

func TestAPI_CreateWorkflow(t *testing.T) {
	wf := &workflow.Workflow{
		Name:        "w1",
		Description: "www",
	}
	workflowMock := &workflow.WorkflowServiceMock{
		CreateFunc: func(ctx context.Context, wff *workflow.Workflow) (string, error) {
			require.Equal(t, wf, wff)
			return "di", nil
		},
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.WorkflowService = workflowMock
	})
	r, err := server.NewPostRequest("/api/v1/workflows", wf)
	require.NoError(t, err)
	resp := server.Send(r)
	require.Equal(t, 1, len(workflowMock.CreateCalls()))
	resp.ResponseEq(
		t, 200, gin.H{"id": "di"},
	)
}

func TestAPI_UpdateWorkflow(t *testing.T) {
	wf := &workflow.Workflow{
		Name:        "w1",
		Description: "www",
	}
	workflowMock := &workflow.WorkflowServiceMock{
		UpdateFunc: func(ctx context.Context, id string, wff *workflow.Workflow) (string, error) {
			require.Equal(t, "abc", id)
			require.Equal(t, wf, wff)
			return "di", nil
		},
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.WorkflowService = workflowMock
	})
	r, err := server.NewPatchRequest("/api/v1/workflows/abc", wf)
	require.NoError(t, err)
	resp := server.Send(r)
	require.Equal(t, 1, len(workflowMock.UpdateCalls()))
	resp.ResponseEq(
		t, 200, gin.H{"id": "di"},
	)
}

func TestAPI_DeleteWorkflow(t *testing.T) {
	workflowMock := &workflow.WorkflowServiceMock{
		DeleteFunc: func(ctx context.Context, wf string) error {
			require.Equal(t, "abc", wf)
			return nil
		},
	}

	server := NewTestServer(t, func(s *services.Backend) {
		s.WorkflowService = workflowMock
	})
	r, err := server.NewDeleteRequest("/api/v1/workflows/abc")
	require.NoError(t, err)
	resp := server.Send(r)
	require.Equal(t, 1, len(workflowMock.DeleteCalls()))
	resp.ResponseEq(
		t, 200, "",
	)
}

func TestAPI_RunWorkflowFileVar(t *testing.T) {
	mockRunner := &workflow.RunnerMock{
		NextFunc: func(ctx context.Context) (*workflow.WorkflowStepResult, error) {
			return nil, nil
		},
	}
	mockWorkflow := &workflow.WorkflowServiceMock{
		StartFunc: func(ctx context.Context, id string, request workflow.StartWorklfowRequest) (workflow.Runner, error) {
			require.Equal(t, request, workflow.StartWorklfowRequest{
				Variables: map[string]any{
					"image": "go.csv",
					"go.csv__data": workflow.FileInfo{
						Name: "go.csv",
						Data: []byte("Hello, World!"),
					},
				},
				Model:       "aiai",
				ImageModel:  "aiia",
				Temperature: 0.56,
			})
			return mockRunner, nil
		},
		GetFunc: func(ctx context.Context, wf string) (*ent.Workflow, error) {
			require.Equal(t, "foo", wf)
			return &ent.Workflow{
				Variables: []schema.WorkflowVariable{
					{Name: "image", Type: schema.WorkflowVariableTypeFile},
				},
			}, nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.WorkflowService = mockWorkflow
	})
	req, err := server.NewPostRequest("/api/v1/workflows/foo/run", &workflow.StartWorklfowRequest{
		Temperature: 0.56,
		Model:       "aiai",
		ImageModel:  "aiia",
		Variables: map[string]any{"image": map[string]any{
			"name": "go.csv",
			"data": "data:text/csv;base64,SGVsbG8sIFdvcmxkIQ==",
		}},
	})
	require.NoError(t, err)
	resp := server.Send(req)
	require.Equal(t, 200, resp.response.Code)
	require.Equal(t, 1, len(mockWorkflow.StartCalls()))
}
