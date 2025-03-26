package huggingface

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHFClient_GetDatasetSize(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mockResponse := DatasetSizeResponse{
		Size: SizeInfo{
			Splits: []SplitInfo{{NumRows: 1000, Config: "default", Split: "train"}},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient("dataset", "default", "train", logger)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	resp, err := client.GetDatasetSize(context.TODO())
	require.NoError(t, err)
	require.Equal(t, 1000, resp.Size.Splits[0].NumRows)
}

func TestHFCLient_GetDatasetRows(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mockResponse := RowResponse{
		Rows: []RowInfo{{Row: map[string]any{"text": "sample data"}}},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient("dataset", "default", "train", logger)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	resp, err := client.GetDatasetRows(context.TODO(), 0, 1)
	require.NoError(t, err)
	require.Equal(t, "sample data", resp.Rows[0].Row["text"])
}

func TestHFClient_GetDatasetInfo(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mockResponse := DatasetInfoResponse{
		Features: map[string]any{"feature1": "type1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient("dataset", "default", "train", logger)
	client.httpClient = server.Client()
	client.baseURL = server.URL

	resp, err := client.GetDatasetInfo(context.TODO())
	require.NoError(t, err)
	require.Equal(t, "type1", resp.Features["feature1"])
}
