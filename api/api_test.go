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
