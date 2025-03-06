package api

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/table"
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
	mockTable := table.TableServiceMock{
		CreateTableFunc: func(ctx context.Context, req *table.TableGenRequest) (string, error) {
			require.Equal(t, expectedRequest, req)
			return "foo", nil
		},
	}
	server := NewTestServer(t, func(s *services.Backend) {
		s.TableService = &mockTable
	})
	req, err := server.NewPostRequest("/api/v1/tables", expectedRequest)
	require.NoError(t, err)
	resp := server.Send(req)
	resp.ResponseEq(t, 200, map[string]string{"id": "foo"})
}
