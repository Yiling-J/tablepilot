package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type ResponseHelper struct {
	response *httptest.ResponseRecorder
}

func (h *ResponseHelper) Response() *httptest.ResponseRecorder {
	return h.response
}

func (h *ResponseHelper) ResponseEq(t *testing.T, code int, expected any) {
	b, err := json.Marshal(expected)
	require.NoError(t, err)
	require.Equal(t, code, h.response.Code, h.response.Body.String())
	if expected == nil {
		require.Equal(t, "", h.response.Body.String())
	} else {
		require.JSONEq(t, string(b), h.response.Body.String())
	}
}

type TestServer struct {
	server *HTTPServer
}

func NewTestServer(t *testing.T, option func(s *services.Backend)) *TestServer {
	gin.SetMode(gin.ReleaseMode)
	config := &config.Config{}
	backend := &services.Backend{Config: config, Logger: zap.NewNop().Sugar()}
	option(backend)
	server := NewHttpServer(backend, false)
	server.RegisterRoutes()
	return &TestServer{server: server}
}

func (ts *TestServer) NewRequest(method, url string, body any) (*http.Request, error) {
	d, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(method, url, strings.NewReader(string(d)))
}

func (ts *TestServer) NewGetRequest(url string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, url, nil)
}

func (ts *TestServer) NewPostRequest(url string, body any) (*http.Request, error) {
	d, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(http.MethodPost, url, strings.NewReader(string(d)))
}

func (ts *TestServer) NewPatchRequest(url string, body any) (*http.Request, error) {
	d, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(http.MethodPatch, url, strings.NewReader(string(d)))
}

func (ts *TestServer) NewPutRequest(url string, body any) (*http.Request, error) {
	d, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(http.MethodPut, url, strings.NewReader(string(d)))
}

func (ts *TestServer) NewDeleteRequest(url string) (*http.Request, error) {
	return http.NewRequest(http.MethodDelete, url, nil)
}

func (ts *TestServer) NewMultiplePartRequest(method, url string, fields map[string]string, files map[string]io.Reader) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, err
		}
	}
	for fieldName, fileReader := range files {
		part, err := writer.CreateFormFile("files", fieldName)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, fileReader); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func (ts *TestServer) Send(req *http.Request) *ResponseHelper {
	w := httptest.NewRecorder()
	ts.server.Engine.ServeHTTP(w, req)
	return &ResponseHelper{response: w}
}
