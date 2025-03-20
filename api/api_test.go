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
	expectedRequest.MarkAPIRequest()
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
		{"1": "a1", "2": "b1"},
		{"1": "a2", "2": "b2"},
		{"1": "a3", "2": "b3"},
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
		Default: "foo",
		Models:  []string{"foo", "bar"},
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
