package workflow

import (
	"context"
	"encoding/json"
	"io"
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

func TestWorkflow_List(t *testing.T) {
	db := db.NewTestDB()
	wf := NewWorkflowService(db, &table.TableServiceMock{})
	err := db.Workflow.Create().SetName("wf1").SetVariables([]schema.WorkflowVariable{}).SetSteps([]schema.WorkflowStep{}).Exec(t.Context())
	require.NoError(t, err)
	err = db.Workflow.Create().SetName("wf2").SetVariables([]schema.WorkflowVariable{}).SetSteps([]schema.WorkflowStep{}).Exec(t.Context())
	require.NoError(t, err)
	wfs, err := wf.List(t.Context())
	require.NoError(t, err)
	require.Equal(t, 2, len(wfs))
	require.Equal(t, "wf1", wfs[0].Name)
	require.Equal(t, "wf2", wfs[1].Name)
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
	runner, err := wf.Start(t.Context(), id, vars, "", 0.5)
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
					require.Equal(t, "m1", req.Model)

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
			runner, err := wf.Start(t.Context(), id, map[string]any{}, "m1", 0.5)
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

func TestWorkflowRunner_Import(t *testing.T) {
	testCases := []struct {
		name    string
		step    schema.WorkflowStep
		message string
		tp      string
	}{
		{
			name: "import csv", step: schema.WorkflowStep{
				Type:    schema.WorkflowStepTypeImport,
				Payload: json.RawMessage(`{"table": "foo","file":"test.csv","prompt":"bar"}`),
			},
			message: "CSV imported: z",
			tp:      "csv",
		},
		{
			name: "import image", step: schema.WorkflowStep{
				Type:    schema.WorkflowStepTypeImport,
				Payload: json.RawMessage(`{"table": "foo","file":"test.png","prompt":"bar"}`),
			},
			message: "Image imported: z",
			tp:      "img",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := db.NewTestDB()
			tm := &table.TableServiceMock{
				ImportFunc: func(ctx context.Context, table string, reader io.Reader) (string, error) {
					require.Equal(t, "foo", table)
					d, err := io.ReadAll(reader)
					require.NoError(t, err)
					require.Equal(t, "csv", string(d))
					return "z", nil
				},
				ImportImageFunc: func(ctx context.Context, request table.ImageImportRequest) (string, error) {
					require.Equal(t, "bar", request.Prompt)
					require.Equal(t, "m1", request.Model)
					require.Equal(t, []byte("png"), request.Data)
					return "z", nil
				},
			}
			wf := NewWorkflowService(db, tm)
			id, err := wf.Create(t.Context(), &Workflow{
				Name:      "wf1",
				Variables: []schema.WorkflowVariable{},
				Steps:     []schema.WorkflowStep{tc.step},
			})
			require.NoError(t, err)
			runner, err := wf.Start(t.Context(), id, map[string]any{
				"test.csv": []byte("csv"),
				"test.png": []byte("png"),
			}, "m1", 0.5)
			require.NoError(t, err)
			r, err := runner.Next(t.Context())
			require.NoError(t, err)
			switch tc.tp {
			case "csv":
				require.Equal(t, 1, len(tm.ImportCalls()))
			case "img":
				require.Equal(t, 1, len(tm.ImportImageCalls()))
			default:
				require.FailNow(t, "unknown import type")
			}
			require.Equal(t, WorkflowActionShowMessage, r.Action)
			require.Equal(t, tc.message, r.Message)
		})
	}
}

func TestWorkflowRunner_Generate(t *testing.T) {
	db := db.NewTestDB()
	tm := &table.TableServiceMock{
		GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
			require.Equal(t, "m1", params.Model)
			require.Equal(t, 0.5, params.Temperature)
			require.Equal(t, "foo", params.Table)
			require.Equal(t, 5, params.Count)
			require.Equal(t, 2, params.Batch)
			return nil, nil
		},
	}
	wf := NewWorkflowService(db, tm)
	id, err := wf.Create(t.Context(), &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{},
		Steps: []schema.WorkflowStep{{
			Type:    schema.WorkflowStepTypeGenerate,
			Payload: json.RawMessage(`{"table": "foo","count":5,"batch":2}`),
		}},
	})
	require.NoError(t, err)
	runner, err := wf.Start(t.Context(), id, map[string]any{}, "m1", 0.5)
	require.NoError(t, err)
	r, err := runner.Next(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, len(tm.GenetateCalls()))
	require.Equal(t, &WorkflowStepResult{
		Action:  WorkflowActionGenerate,
		Message: "Start generating rows for table foo...",
	}, r)
}

func TestWorkflowRunner_Autofill(t *testing.T) {
	db := db.NewTestDB()
	tm := &table.TableServiceMock{
		GenetateFunc: func(ctx context.Context, params table.GenerateRowsRequest) (table.RowsGenerator, error) {
			require.Equal(t, "m1", params.Model)
			require.Equal(t, 0.5, params.Temperature)
			require.Equal(t, "foo", params.Table)
			require.Equal(t, 5, params.Count)
			require.Equal(t, 2, params.Batch)
			return nil, nil
		},
	}
	wf := NewWorkflowService(db, tm)
	id, err := wf.Create(t.Context(), &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{},
		Steps: []schema.WorkflowStep{{
			Type:    schema.WorkflowStepTypeAutofill,
			Payload: json.RawMessage(`{"table": "foo","count":5,"batch":2}`),
		}},
	})
	require.NoError(t, err)
	runner, err := wf.Start(t.Context(), id, map[string]any{}, "m1", 0.5)
	require.NoError(t, err)
	r, err := runner.Next(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, len(tm.GenetateCalls()))
	require.Equal(t, &WorkflowStepResult{
		Action:  WorkflowActionGenerate,
		Message: "Start autofilling rows for table foo...",
	}, r)
}

func TestWorkflowRunner_DeleteTable(t *testing.T) {
	db := db.NewTestDB()
	tm := &table.TableServiceMock{
		DeleteFunc: func(ctx context.Context, table string) (int, error) {
			require.Equal(t, "foo", table)
			return 1, nil
		},
	}
	wf := NewWorkflowService(db, tm)
	id, err := wf.Create(t.Context(), &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{},
		Steps: []schema.WorkflowStep{{
			Type:    schema.WorkflowStepTypeDeleteTable,
			Payload: json.RawMessage(`{"table": "foo"}`),
		}},
	})
	require.NoError(t, err)
	runner, err := wf.Start(t.Context(), id, map[string]any{}, "m1", 0.5)
	require.NoError(t, err)
	r, err := runner.Next(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, len(tm.DeleteCalls()))
	require.Equal(t, &WorkflowStepResult{
		Action:  WorkflowActionShowMessage,
		Message: "Table foo deleted",
	}, r)
}

func TestWorkflowRunner_ExportTable(t *testing.T) {
	db := db.NewTestDB()
	tm := &table.TableServiceMock{
		CSVFunc: func(ctx context.Context, table string) ([]byte, error) {
			require.Equal(t, "foo", table)
			return []byte("csv"), nil
		},
	}
	wf := NewWorkflowService(db, tm)
	id, err := wf.Create(t.Context(), &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{},
		Steps: []schema.WorkflowStep{{
			Type:    schema.WorkflowStepTypeExportTable,
			Payload: json.RawMessage(`{"table": "foo","path":"tmp.csv"}`),
		}},
	})
	require.NoError(t, err)
	runner, err := wf.Start(t.Context(), id, map[string]any{}, "m1", 0.5)
	require.NoError(t, err)
	r, err := runner.Next(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, len(tm.CSVCalls()))
	require.Equal(t, &WorkflowStepResult{
		Action:     WorkflowActionExport,
		Message:    "Table foo exported",
		ExportData: "csv",
		ExportPath: "tmp.csv",
	}, r)
}

func TestWorkflowRunner_CreateColumn(t *testing.T) {
	db := db.NewTestDB()
	tm := &table.TableServiceMock{
		CreateColumnFunc: func(ctx context.Context, table string, col table.TableGenColumn) (string, error) {
			require.Equal(t, "foo", table)
			require.Equal(t, "col", col.Name)
			return "", nil
		},
	}
	wf := NewWorkflowService(db, tm)
	id, err := wf.Create(t.Context(), &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{},
		Steps: []schema.WorkflowStep{{
			Type:    schema.WorkflowStepTypeCreateColumn,
			Payload: json.RawMessage(`{"table": "foo","column":{"name":"col"}}`),
		}},
	})
	require.NoError(t, err)
	runner, err := wf.Start(t.Context(), id, map[string]any{}, "m1", 0.5)
	require.NoError(t, err)
	r, err := runner.Next(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, len(tm.CreateColumnCalls()))
	require.Equal(t, &WorkflowStepResult{
		Action:  WorkflowActionShowMessage,
		Message: "Column col created",
	}, r)
}

func TestWorkflowRunner_DeleteColumn(t *testing.T) {
	db := db.NewTestDB()
	tm := &table.TableServiceMock{
		DeleteColumnFunc: func(ctx context.Context, table, column string) (string, error) {
			require.Equal(t, "foo", table)
			require.Equal(t, "col", column)
			return "", nil
		},
	}
	wf := NewWorkflowService(db, tm)
	id, err := wf.Create(t.Context(), &Workflow{
		Name:      "wf1",
		Variables: []schema.WorkflowVariable{},
		Steps: []schema.WorkflowStep{{
			Type:    schema.WorkflowStepTypeDeleteColumn,
			Payload: json.RawMessage(`{"table": "foo","column":"col"}`),
		}},
	})
	require.NoError(t, err)
	runner, err := wf.Start(t.Context(), id, map[string]any{}, "m1", 0.5)
	require.NoError(t, err)
	r, err := runner.Next(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, len(tm.DeleteColumnCalls()))
	require.Equal(t, &WorkflowStepResult{
		Action:  WorkflowActionShowMessage,
		Message: "Column col deleted",
	}, r)
}
