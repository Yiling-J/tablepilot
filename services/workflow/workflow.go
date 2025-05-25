package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/workflow"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/spf13/cast"
)

//go:generate moq -rm -out workflow_moq.go . WorkflowService Runner
type WorkflowService interface {
	Get(ctx context.Context, wf string) (*ent.Workflow, error)
	Delete(ctx context.Context, wf string) error
	Create(ctx context.Context, wf *Workflow) (string, error)
	Update(ctx context.Context, id string, wf *Workflow) (string, error)
	Start(ctx context.Context, id string, request StartWorklfowRequest) (Runner, error)
	List(ctx context.Context) ([]*ent.Workflow, error)
}

type WorkflowServiceImpl struct {
	db    *ent.Client
	table table.TableService
}

func NewWorkflowService(db *ent.Client, table table.TableService) *WorkflowServiceImpl {
	return &WorkflowServiceImpl{
		db:    db,
		table: table,
	}
}

func (w *WorkflowServiceImpl) Get(ctx context.Context, wf string) (*ent.Workflow, error) {
	workflow, err := w.db.Workflow.Query().Where(workflow.Or(workflow.Name(wf), workflow.Nanoid(wf))).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("workflow.Get: querying workflow: %w", err)
	}
	return workflow, nil
}

func (w *WorkflowServiceImpl) List(ctx context.Context) ([]*ent.Workflow, error) {
	workflows, err := w.db.Workflow.Query().Order(ent.Asc(workflow.FieldID)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("workflow.List: listing workflows: %w", err)
	}
	return workflows, nil
}

func (w *WorkflowServiceImpl) Delete(ctx context.Context, wf string) error {
	_, err := w.db.Workflow.Delete().Where(workflow.Or(workflow.Name(wf), workflow.Nanoid(wf))).Exec(ctx)
	if err != nil {
		return fmt.Errorf("workflow.Delete: deleting workflow: %w", err)
	}
	return nil
}

func (w *WorkflowServiceImpl) Create(ctx context.Context, wf *Workflow) (string, error) {
	if len(wf.Name) == 0 {
		return "", errors.New("name must not be empty")
	}
	if len(wf.Steps) == 0 {
		return "", errors.New("steps must not be empty")
	}
	dbwf, err := w.db.Workflow.Create().SetName(wf.Name).SetDescription(wf.Description).SetVariables(wf.Variables).SetSteps(wf.Steps).Save(ctx)
	if err != nil {
		return "", fmt.Errorf("workflow.Create: saving workflow: %w", err)
	}
	return dbwf.Nanoid, nil
}

func (w *WorkflowServiceImpl) Update(ctx context.Context, id string, wf *Workflow) (string, error) {
	if len(wf.Name) == 0 {
		return "", errors.New("name must not be empty")
	}
	if len(wf.Steps) == 0 {
		return "", errors.New("steps must not be empty")
	}
	err := w.db.Workflow.Update().Where(workflow.Nanoid(id)).SetName(wf.Name).SetDescription(wf.Description).SetVariables(wf.Variables).SetSteps(wf.Steps).Exec(ctx)
	if err != nil {
		return "", fmt.Errorf("workflow.Create: saving workflow: %w", err)
	}
	return id, nil
}

func (w *WorkflowServiceImpl) Start(ctx context.Context, id string, request StartWorklfowRequest) (Runner, error) {
	wf, err := w.db.Workflow.Query().Where(workflow.Nanoid(id)).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("workflow.Start: querying workflow: %w", err)
	}
	now := time.Now().UTC()
	if _, ok := request.Variables["date"]; !ok {
		request.Variables["date"] = now.Format("20060102")
	}
	if _, ok := request.Variables["time"]; !ok {
		request.Variables["time"] = now.Format("150405")
	}
	if _, ok := request.Variables["datetime"]; !ok {
		request.Variables["datetime"] = now.Format("20060102150405")
	}
	return &RunnerImpl{
		workflow: wf, context: request.Variables, db: w.db, imageModel: request.ImageModel,
		tableService: w.table, model: request.Model, temperature: request.Temperature,
	}, nil
}

type Runner interface {
	Next(ctx context.Context) (*WorkflowStepResult, error)
}

type RunnerImpl struct {
	workflow     *ent.Workflow
	index        int
	tableService table.TableService
	context      map[string]any
	db           *ent.Client
	model        string
	imageModel   string
	temperature  float64
}

type WorkflowAction string

const (
	WorkflowActionShowMessage WorkflowAction = "ShowMessage"
	WorkflowActionGenerate    WorkflowAction = "Genetate"
	WorkflowActionExport      WorkflowAction = "Export"
)

type WorkflowStepResult struct {
	Action     WorkflowAction
	Message    string
	ExportData string
	ExportPath string
	Generator  table.RowsGenerator
}

func (r *RunnerImpl) Next(ctx context.Context) (*WorkflowStepResult, error) {
	if r.index >= len(r.workflow.Steps) {
		return nil, nil
	}
	defer func() { r.index += 1 }()
	step := r.workflow.Steps[r.index]
	stepContext := StepContext{}
	defer func() { r.context[fmt.Sprintf("step%d", r.index+1)] = stepContext.AsMap() }()

	b, err := json.Marshal(step)
	if err != nil {
		return nil, fmt.Errorf("workflow.Next: marshaling step: %w", err)
	}
	tmpl, err := template.New("wf").Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("workflow.Next: parsing template: %w", err)
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, r.context)
	if err != nil {
		return nil, fmt.Errorf("workflow.Next: executing template: %w", err)
	}
	b = buffer.Bytes()
	err = json.Unmarshal(b, &step)
	if err != nil {
		return nil, fmt.Errorf("workflow.Next: unmarshaling step: %w", err)
	}

	switch step.Type {
	case schema.WorkflowStepTypeCreateTable:
		var payload WorkflowCreateTablePayload
		err = json.Unmarshal(step.Payload, &payload)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling table request: %w", err)
		}
		if payload.SchemaFile != "" {
			cb, err := os.ReadFile(payload.SchemaFile)
			if err != nil {
				return nil, fmt.Errorf("workflow.Next: reading schema file: %w", err)
			}
			tmpl, err := template.New("wf").Parse(string(cb))
			if err != nil {
				return nil, fmt.Errorf("workflow.Next: parsing schema template: %w", err)
			}
			var buffer bytes.Buffer
			err = tmpl.Execute(&buffer, r.context)
			if err != nil {
				return nil, fmt.Errorf("workflow.Next: executing schema template: %w", err)
			}
			cb = buffer.Bytes()
			err = json.Unmarshal(cb, &payload.Request)
			if err != nil {
				return nil, fmt.Errorf("workflow.Next: unmarshaling table request: %w", err)
			}
		}
		req := payload.Request
		req.Name = SanitizeString(req.Name)
		req.Model = r.model
		tables, err := r.db.TableMeta.Query().Where(
			tablemeta.Name(req.Name),
		).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: querying tables: %w", err)
		}
		if len(tables) > 0 {
			switch payload.OnExists {
			case schema.OnExistsRecreate:
				_, err = r.db.TableMeta.Delete().Where(tablemeta.ID(tables[0].ID)).Exec(ctx)
				if err != nil {
					return nil, fmt.Errorf("workflow.Next: deleting existing table: %w", err)
				}
			case schema.OnExistsStop:
				return nil, fmt.Errorf("table %s already exists", req.Name)
			case schema.OnExistsSkip:
				return &WorkflowStepResult{
					Action:  WorkflowActionShowMessage,
					Message: fmt.Sprintf("Table %s already exists, skip creating.", req.Name)}, nil
			default:
				return nil, fmt.Errorf("table %s already exists", req.Name)
			}
		}
		id, err := r.tableService.Create(ctx, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: creating table: %w", err)
		}
		stepContext.Table = id
		return &WorkflowStepResult{Message: fmt.Sprintf("Table created: id %s, name %s", id, req.Name), Action: WorkflowActionShowMessage}, nil
	case schema.WorkflowStepTypeImport:
		var req WorkflowImportFileParams
		err := json.Unmarshal(step.Payload, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling import file params: %w", err)
		}
		fileInfo, ok := r.context[fmt.Sprintf("%s__data", req.File)]
		if !ok {
			return nil, fmt.Errorf("workflow.Next: file %s not found", req.File)
		}
		cb, ok := fileInfo.(FileInfo)
		if !ok {
			return nil, errors.New("invalid file content")
		}
		switch filepath.Ext(cb.Name) {
		case ".csv":
			bf := bytes.NewBuffer(cb.Data)
			fileName := filepath.Base(req.File)
			fileName = strings.TrimSuffix(fileName, filepath.Ext(req.File))
			id, err := r.tableService.Import(ctx, table.ImportRequest{
				Table:    SanitizeString(req.Table),
				Truncate: req.Truncate,
				Reader:   bf,
				Filename: fileName,
				Name:     req.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("workflow.Next: importing CSV: %w", err)
			}
			stepContext.Table = id
			return &WorkflowStepResult{
				Message: fmt.Sprintf("CSV imported: %s", id),
				Action:  WorkflowActionShowMessage,
			}, nil
		case ".png", ".jpg", ".jpeg":
			fileName := filepath.Base(req.File)
			fileName = strings.TrimSuffix(fileName, filepath.Ext(req.File))
			id, err := r.tableService.ImportImage(ctx, table.ImportRequest{
				Table:    SanitizeString(req.Table),
				Truncate: req.Truncate,
				Name:     req.Name,
				Data:     cb.Data,
				Prompt:   req.Prompt,
				Model:    r.model,
				Filename: fileName,
			})
			if err != nil {
				return nil, fmt.Errorf("workflow.Next: importing image: %w", err)
			}
			stepContext.Table = id
			return &WorkflowStepResult{
				Message: fmt.Sprintf("Image imported: %s", id),
				Action:  WorkflowActionShowMessage,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported file ext %s", filepath.Ext(req.File))
		}
	case schema.WorkflowStepTypeGenerate:
		var req WorkflowGeneratePayload
		err := json.Unmarshal(step.Payload, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling generate rows request: %w", err)
		}
		generator, err := r.tableService.Genetate(ctx, table.GenerateRowsRequest{
			Table:       SanitizeString(req.Table),
			Model:       r.model,
			ImageModel:  r.imageModel,
			Temperature: r.temperature,
			Count:       cast.ToInt(req.Count),
			Batch:       cast.ToInt(req.Batch),
		})
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: generating rows: %w", err)
		}
		return &WorkflowStepResult{
			Message:   fmt.Sprintf("Start generating rows for table %s...", SanitizeString(req.Table)),
			Generator: generator, Action: WorkflowActionGenerate}, nil
	case schema.WorkflowStepTypeAutofill:
		var req WorkflowAutofillPayload
		err := json.Unmarshal(step.Payload, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling autofill request: %w", err)
		}
		genreq := table.GenerateRowsRequest{
			Table:       SanitizeString(req.Table),
			Model:       r.model,
			ImageModel:  r.imageModel,
			Temperature: r.temperature,
			Count:       cast.ToInt(req.Count),
			Batch:       cast.ToInt(req.Batch),
		}
		genreq.Autofill = table.AutofillRequest{Enable: true, Columns: req.Columns, ContextColumns: req.ContextColumns}
		generator, err := r.tableService.Genetate(ctx, genreq)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: autofilling rows: %w", err)
		}
		return &WorkflowStepResult{
			Action:    WorkflowActionGenerate,
			Message:   fmt.Sprintf("Start autofilling rows for table %s...", SanitizeString(req.Table)),
			Generator: generator}, nil
	case schema.WorkflowStepTypeDeleteTable:
		var req WorkflowDeleteTableParams
		err := json.Unmarshal(step.Payload, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling delete table params: %w", err)
		}
		_, err = r.tableService.Delete(ctx, SanitizeString(req.Table))
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: deleting table: %w", err)
		}
		return &WorkflowStepResult{
			Message: fmt.Sprintf("Table %s deleted", SanitizeString(req.Table)),
			Action:  WorkflowActionShowMessage,
		}, nil
	case schema.WorkflowStepTypeExportTable:
		var req WorkflowExportTableParams
		err := json.Unmarshal(step.Payload, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling export table params: %w", err)
		}
		data, err := r.tableService.CSV(ctx, SanitizeString(req.Table))
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: exporting table to CSV: %w", err)
		}
		return &WorkflowStepResult{
			Message:    fmt.Sprintf("Table %s exported", SanitizeString(req.Table)),
			ExportData: string(data),
			ExportPath: req.Path,
			Action:     WorkflowActionExport,
		}, nil
	case schema.WorkflowStepTypeCreateColumn:
		var req WorkflowCreateColumnParams
		err := json.Unmarshal(step.Payload, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling create column params: %w", err)
		}
		col := table.TableGenColumn{
			Name:        req.Name,
			Description: req.Description,
			Type:        req.Type,
			FillMode:    "ai",
		}
		id, err := r.tableService.CreateColumn(ctx, SanitizeString(req.Table), col)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: creating column: %w", err)
		}
		stepContext.Column = id
		return &WorkflowStepResult{
			Message: fmt.Sprintf("Column %s created", req.Name),
			Action:  WorkflowActionShowMessage,
		}, nil
	case schema.WorkflowStepTypeDeleteColumn:
		var req WorkflowDeleteColumnParams
		err := json.Unmarshal(step.Payload, &req)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: unmarshaling delete column params: %w", err)
		}
		_, err = r.tableService.DeleteColumn(ctx, SanitizeString(req.Table), req.Column)
		if err != nil {
			return nil, fmt.Errorf("workflow.Next: deleting column: %w", err)
		}
		return &WorkflowStepResult{
			Message: fmt.Sprintf("Column %s deleted", req.Column),
			Action:  WorkflowActionShowMessage,
		}, nil
	default:
		return nil, fmt.Errorf("unknown step type %s", step.Type)
	}
}
