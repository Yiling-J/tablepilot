package huggingface

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"
)

const baseURL = "https://datasets-server.huggingface.co"

//go:generate moq -rm -out client_moq.go . Client
type Client interface {
	GetDatasetSize() (*DatasetSizeResponse, error)
	GetDatasetRows(offset, length int) (*RowResponse, error)
	GetDatasetInfo() (*DatasetInfoResponse, error)
}

type ClientImpl struct {
	httpClient *http.Client
	dataset    string
	config     string
	split      string
	logger     *zap.SugaredLogger
	baseURL    string
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
		httpClient: &http.Client{},
		dataset:    dataset,
		config:     config,
		split:      split,
		logger:     logger,
		baseURL:    baseURL,
	}
}

func (c *ClientImpl) GetDatasetSize() (*DatasetSizeResponse, error) {
	params := url.Values{}
	params.Set("dataset", c.dataset)
	params.Set("config", c.config)

	reqURL := fmt.Sprintf("%s/size?%s", c.baseURL, params.Encode())
	c.logger.Debugw("get dataset size using Hugging Face API", "url", reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

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

func (c *ClientImpl) GetDatasetRows(offset, length int) (*RowResponse, error) {
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result RowResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *ClientImpl) GetDatasetInfo() (*DatasetInfoResponse, error) {
	params := url.Values{}
	params.Set("dataset", c.dataset)
	params.Set("config", c.config)

	reqURL := fmt.Sprintf("%s/info?%s", c.baseURL, params.Encode())
	c.logger.Debugw("get dataset info using Hugging Face API", "url", reqURL)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

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
