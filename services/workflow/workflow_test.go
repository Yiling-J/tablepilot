package workflow

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/infra/db"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
)

func TestWorkflow_Create(t *testing.T) {
	db := db.NewTestDB()
	wf := NewWorkflowService(db, &table.TableServiceMock{})
	ww := &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{{Name: "v1", Type: schema.WorkflowVariableTypeString, DefaultValue: "foo"}},
		Steps:     []schema.WorkflowStep{{Type: schema.WorkflowStepTypeCreateColumn, Payload: json.RawMessage(`{"a":"b"}`)}},
	}
	id, err := wf.Create(t.Context(), ww)
	require.NoError(t, err)
	wg, err := wf.Get(t.Context(), "wf1")
	require.NoError(t, err)
	require.Equal(t, id, wg.Nanoid)
	require.Equal(t, ww.Variables, wg.Variables)
	require.Equal(t, ww.Steps, wg.Steps)
}

func TestWorkflow_Get(t *testing.T) {
	db := db.NewTestDB()
	wf := NewWorkflowService(db, &table.TableServiceMock{})
	w, err := db.Workflow.Create().SetName("wf1").SetVariables([]schema.WorkflowVariable{}).SetSteps([]schema.WorkflowStep{}).Save(t.Context())
	require.NoError(t, err)
	wg, err := wf.Get(t.Context(), "wf1")
	require.NoError(t, err)
	require.Equal(t, w.ID, wg.ID)
	wg, err = wf.Get(t.Context(), w.Nanoid)
	require.NoError(t, err)
	require.Equal(t, w.ID, wg.ID)
}

func TestWorkflow_Delete(t *testing.T) {
	db := db.NewTestDB()
	wf := NewWorkflowService(db, &table.TableServiceMock{})
	_, err := db.Workflow.Create().SetName("wf1").SetVariables([]schema.WorkflowVariable{}).SetSteps([]schema.WorkflowStep{}).Save(t.Context())
	require.NoError(t, err)

	err = wf.Delete(t.Context(), "wf1")
	require.NoError(t, err)
	c, err := db.Workflow.Query().Count(t.Context())
	require.NoError(t, err)
	require.Equal(t, 0, c)

	w, err := db.Workflow.Create().SetName("wf1").SetVariables([]schema.WorkflowVariable{}).SetSteps([]schema.WorkflowStep{}).Save(t.Context())
	require.NoError(t, err)

	err = wf.Delete(t.Context(), w.Nanoid)
	require.NoError(t, err)
	c, err = db.Workflow.Query().Count(t.Context())
	require.NoError(t, err)
	require.Equal(t, 0, c)
}

func TestWorkflow_Start(t *testing.T) {
	db := db.NewTestDB()
	wf := NewWorkflowService(db, &table.TableServiceMock{})
	id, err := wf.Create(t.Context(), &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{{Name: "v1", Type: schema.WorkflowVariableTypeString, DefaultValue: "foo"}},
		Steps:     []schema.WorkflowStep{{Type: schema.WorkflowStepTypeCreateColumn, Payload: json.RawMessage(`{"a":"b"}`)}},
	})
	require.NoError(t, err)

	vars := map[string]any{"foo": "bar"}
	runner, err := wf.Start(t.Context(), id, vars)
	require.NoError(t, err)
	require.True(t, len(cast.ToString(vars["date"])) > 0)
	require.True(t, len(cast.ToString(vars["time"])) > 0)
	require.True(t, len(cast.ToString(vars["datetime"])) > 0)
	require.Equal(t, vars, runner.context)
}

func TestWorkflowRunner_CreateTable(t *testing.T) {
	testCases := []struct {
		name    string
		step    schema.WorkflowStep
		err     bool
		prepare func(db *ent.Client)
		assert  func(db *ent.Client, req *table.TableGenRequest, result *WorkflowStepResult)
	}{
		{
			name: "create table", step: schema.WorkflowStep{
				Type:    schema.WorkflowStepTypeCreateTable,
				Payload: json.RawMessage(`{"name": "foo"}`),
			},
			assert: func(db *ent.Client, req *table.TableGenRequest, result *WorkflowStepResult) {
				require.Equal(t, "foo", req.Name)
				require.Equal(t, &WorkflowStepResult{
					Action:  WorkflowActionShowMessage,
					Message: "Table created: id tb, name foo",
				}, result)
			},
		},
		{
			name: "create table invalid name", step: schema.WorkflowStep{
				Type:    schema.WorkflowStepTypeCreateTable,
				Payload: json.RawMessage(`{"name": "foo bar"}`),
			},
			assert: func(db *ent.Client, req *table.TableGenRequest, result *WorkflowStepResult) {
				require.Equal(t, "foo_bar", req.Name)
				require.Equal(t, &WorkflowStepResult{
					Action:  WorkflowActionShowMessage,
					Message: "Table created: id tb, name foo_bar",
				}, result)
			},
		},
		{
			name: "create table exists stop (default)",
			step: schema.WorkflowStep{
				Type:    schema.WorkflowStepTypeCreateTable,
				Payload: json.RawMessage(`{"name": "foo","description": "bar"}`),
			},
			err: true,
			prepare: func(db *ent.Client) {
				err := db.TableMeta.Create().SetName("foo").Exec(t.Context())
				require.NoError(t, err)
			},
			assert: func(db *ent.Client, req *table.TableGenRequest, result *WorkflowStepResult) {
				require.Nil(t, req)
			},
		},
		{
			name: "create table exists recreate",
			step: schema.WorkflowStep{
				Type:     schema.WorkflowStepTypeCreateTable,
				Payload:  json.RawMessage(`{"name": "foo","description":"xyz"}`),
				OnExists: schema.OnExistsRecreate,
			},
			prepare: func(db *ent.Client) {
				err := db.TableMeta.Create().SetName("foo").Exec(t.Context())
				require.NoError(t, err)
			},
			assert: func(db *ent.Client, req *table.TableGenRequest, result *WorkflowStepResult) {
				require.Equal(t, "foo", req.Name)
				require.Equal(t, "xyz", req.Description)
				require.Equal(t, &WorkflowStepResult{
					Action:  WorkflowActionShowMessage,
					Message: "Table created: id tb, name foo",
				}, result)
			},
		},
		{
			name: "create table exists skip",
			step: schema.WorkflowStep{
				Type:     schema.WorkflowStepTypeCreateTable,
				Payload:  json.RawMessage(`{"name": "foo","description":"xyz"}`),
				OnExists: schema.OnExistsSkip,
			},
			prepare: func(db *ent.Client) {
				err := db.TableMeta.Create().SetName("foo").Exec(t.Context())
				require.NoError(t, err)
			},
			assert: func(db *ent.Client, req *table.TableGenRequest, result *WorkflowStepResult) {
				require.Nil(t, req)
				require.Equal(t, &WorkflowStepResult{
					Action:  WorkflowActionShowMessage,
					Message: "Table foo already exists, skip creating.",
				}, result)
			},
		},
		{
			name: "create table from schema file", step: schema.WorkflowStep{
				Type:       schema.WorkflowStepTypeCreateTable,
				SchemaFile: "wf.json",
			},
			prepare: func(db *ent.Client) {
				f, err := os.Create("wf.json")
				require.NoError(t, err)
				defer f.Close()
				require.NoError(t, err)
				_, err = f.WriteString(`{"name": "foo"}`)
				require.NoError(t, err)
			},
			assert: func(db *ent.Client, req *table.TableGenRequest, result *WorkflowStepResult) {
				defer os.Remove("wf.json")
				require.Equal(t, "foo", req.Name)
				require.Equal(t, &WorkflowStepResult{
					Action:  WorkflowActionShowMessage,
					Message: "Table created: id tb, name foo",
				}, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := db.NewTestDB()
			var genreq *table.TableGenRequest
			wf := NewWorkflowService(db, &table.TableServiceMock{
				CreateFunc: func(ctx context.Context, req *table.TableGenRequest) (string, error) {
					genreq = req
					return "tb", nil
				},
			})
			if tc.prepare != nil {
				tc.prepare(db)
			}
			id, err := wf.Create(t.Context(), &Workflow{
				Name:      "wf1",
				Variables: []schema.WorkflowVariable{},
				Steps:     []schema.WorkflowStep{tc.step},
			})
			require.NoError(t, err)
			runner, err := wf.Start(t.Context(), id, map[string]any{})
			require.NoError(t, err)
			r, err := runner.Next(t.Context())
			if tc.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			tc.assert(db, genreq, r)
		})
	}
}
