package huggingface

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const baseURL = "https://datasets-server.huggingface.co"

//go:generate moq -rm -out client_moq.go . Client
type Client interface {
	GetDatasetSize(ctx context.Context) (*DatasetSizeResponse, error)
	GetDatasetRows(ctx context.Context, offset, length int) (*RowResponse, error)
	GetDatasetInfo(ctx context.Context) (*DatasetInfoResponse, error)
}

type ClientImpl struct {
	httpClient *http.Client
	dataset    string
	config     string
	split      string
	logger     *zap.SugaredLogger
	baseURL    string
	limiter    *rate.Limiter
}

type SplitInfo struct {
	NumRows int    `json:"num_rows"`
	Config  string `json:"config"`
	Split   string `json:"split"`
}

type SizeInfo struct {
	Splits []SplitInfo `json:"splits"`
}

type DatasetSizeResponse struct {
	Size SizeInfo `json:"size"`
}

type RowInfo struct {
	Row map[string]any `json:"row"`
}
type RowResponse struct {
	Rows []RowInfo `json:"rows"`
}

type DatasetInfoResponse struct {
	Features map[string]any `json:"features"`
}

func NewClient(dataset, config, split string, logger *zap.SugaredLogger) *ClientImpl {
	return &ClientImpl{
		httpClient: &http.Client{Transport: &SnapshotTransport{}},
		dataset:    dataset,
		config:     config,
		split:      split,
		logger:     logger,
		baseURL:    baseURL,
		limiter:    rate.NewLimiter(rate.Every(5*time.Second), 5),
	}
}

func (c *ClientImpl) GetDatasetSize(ctx context.Context) (*DatasetSizeResponse, error) {
	params := url.Values{}
	params.Set("dataset", c.dataset)
	params.Set("config", c.config)

	reqURL := fmt.Sprintf("%s/size?%s", c.baseURL, params.Encode())
	c.logger.Debugw("get dataset size using Hugging Face API", "url", reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DatasetSizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *ClientImpl) GetDatasetRows(ctx context.Context, offset, length int) (*RowResponse, error) {
	err := c.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("dataset", c.dataset)
	params.Set("config", c.config)
	params.Set("split", c.split)
	params.Set("offset", strconv.Itoa(offset))
	params.Set("length", strconv.Itoa(length))

	reqURL := fmt.Sprintf("%s/rows?%s", c.baseURL, params.Encode())
	c.logger.Debugw("get dataset rows using Hugging Face API", "url", reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result RowResponse
	if resp.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		c.logger.Errorw("Hugging Face get rows API error", "status", resp.StatusCode, "message", string(bodyBytes))
		return nil, errors.New("Hugging Face get rows API error")
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *ClientImpl) GetDatasetInfo(ctx context.Context) (*DatasetInfoResponse, error) {
	params := url.Values{}
	params.Set("dataset", c.dataset)
	params.Set("config", c.config)

	reqURL := fmt.Sprintf("%s/info?%s", c.baseURL, params.Encode())
	c.logger.Debugw("get dataset info using Hugging Face API", "url", reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DatasetInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func SnapshotMiddleware(name string, next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		var reqBody []byte
		if r.Body != nil {
			reqBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(reqBody))
		}

		resp, err := next.RoundTrip(r)
		if err != nil {
			return resp, err
		}

		var respBody []byte
		if resp.Body != nil {
			respBody, _ = io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewReader(respBody))
		}

		snapshot := map[string]string{
			"request":  string(reqBody),
			"response": string(respBody),
		}

		var snapshots []map[string]string
		filename := fmt.Sprintf("tests/snapshots/%s_hf.json", name)
		if fileData, err := os.ReadFile(filename); err == nil {
			_ = json.Unmarshal(fileData, &snapshots)
		}

		snapshots = append(snapshots, snapshot)

		if file, err := os.Create(filename); err == nil {
			defer file.Close()
			_ = json.NewEncoder(file).Encode(snapshots)
		}

		return resp, nil
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type SnapshotTransport struct {
	BaseTransport http.RoundTripper
}

func (s *SnapshotTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if s.BaseTransport == nil {
		s.BaseTransport = http.DefaultTransport
	}

	if name, ok := os.LookupEnv("TABLEPILOT_SNAPSHOT_RECORD"); ok && len(name) > 0 {
		return SnapshotMiddleware(name, s.BaseTransport).RoundTrip(r)
	}

	return s.BaseTransport.RoundTrip(r)
}
