package api

import (
	"context"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/services"
	"github.com/Yiling-J/tablepilot/services/workflow"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

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
	expectedData := "event:message\ndata:{\"data\":\"foobar\",\"type\":\"MESSAGE\"}\n\nevent:message\ndata:{\"type\":\"STEP_DONE\"}\n\nevent:message\ndata:{\"type\":\"WORKFLOW_DONE\"}\n\nevent:message\ndata:[DONE]\n\n\"\""
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
			"data": "data:text/csv;base64,SGVsbG8sIFdvcmxkIQ==", // "Hello, World!" base64 encoded
		}},
	})
	require.NoError(t, err)
	resp := server.Send(req)
	require.Equal(t, 200, resp.response.Code) // Check status code
	require.Equal(t, 1, len(mockWorkflow.StartCalls()))
}
