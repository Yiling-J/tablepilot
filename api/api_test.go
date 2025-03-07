package api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
)

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
	tableMock := &table.TableServiceMock{
		CreateTableFunc: func(ctx context.Context, req *table.TableGenRequest) (string, error) {
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
		{"c1": "0", "c2": "t0"},
		{"c1": "1", "c2": "t1"},
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
data:{"data":[{"c1":"0","c2":"t0"}]}

event:message
data:{"data":[{"c1":"1","c2":"t1"}]}

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
					{Cells: []*schema.CellValue{{Value: "a1"}, {Value: "b1"}}},
					{Cells: []*schema.CellValue{{Value: "a2"}, {Value: "b2"}}},
					{Cells: []*schema.CellValue{{Value: "a3"}, {Value: "b3"}}},
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
		{"c1": "a1", "c2": "b1"},
		{"c1": "a2", "c2": "b2"},
		{"c1": "a3", "c2": "b3"},
	}
	resp.ResponseEq(t, 200, map[string]any{"data": expectedRows, "total": 3})
}

func TestAPI_ListTables(t *testing.T) {
	expectedResponse := &table.ListTablesResponse{
		Total: 2,
		Tables: []table.TableInfoSimple{
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
