package services

import (
	"context"
	"encoding/json"

	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/Yiling-J/tablepilot/services/table"
	"github.com/Yiling-J/tablepilot/services/workflow"
)

func createAppStartExample(ctx context.Context, tableService table.TableService, datasetService dataset.DatasetService, workflowService workflow.WorkflowService) error {
	ds, err := datasetService.Create(ctx, &dataset.CreateDatasetRequest{
		Name:        "cuisines",
		Description: "Common recipe cuisines",
		Type:        "list",
		Data: []string{
			"Italian",
			"Chinese",
			"Mexican",
			"Indian",
			"Japanese",
			"Thai",
			"French",
			"Greek",
			"Korean",
			"Spanish",
			"Vietnamese",
			"Turkish",
			"Moroccan",
			"Lebanese",
			"American",
			"Brazilian",
			"Ethiopian",
			"German",
			"Indonesian",
			"Caribbean",
			"Persian",
			"Cuban",
			"Russian",
			"Pakistani",
			"Malaysian",
			"Polish",
			"Filipino",
			"Syrian",
			"Afghan",
			"Argentinian",
		},
	})
	if err != nil {
		return err
	}
	_, err = tableService.Create(ctx, &table.TableGenRequest{
		Name:        "recipes",
		Description: "A table of delicious and diverse recipes.",
		Columns: []*table.TableGenColumn{
			{
				Name:        "name",
				Description: "The name of the recipe.",
				FillMode:    "ai",
				Type:        "string",
			},
			{
				Name:        "ingredients",
				Description: "A list of ingredients required for the recipe.",
				FillMode:    "ai",
				Type:        "array",
			},
			{
				Name:        "instructions",
				Description: "Step-by-step cooking instructions, detailed enough for beginners.",
				FillMode:    "ai",
				Type:        "array",
			},
			{
				Name:        "cuisine",
				Description: "The cuisine category this recipe belongs to.",
				Type:        "string",
				FillMode:    "pick",
				SourceType:  tablecolumn.SourceTypeDataset,
				SourceID:    ds,
				Random:      true,
				Replacement: true,
			},
			{
				Name:        "tags",
				Description: "Five descriptive tags for this recipe (e.g., vegan, spicy, quick).",
				FillMode:    "ai",
				Type:        "array",
			},
		},
	})
	if err != nil {
		return err
	}

	createTablePayload := &workflow.WorkflowCreateTablePayload{
		OnExists: schema.OnExistsRecreate,
		Request: table.TableGenRequest{
			Name:        "{{.fruit}}_recipes",
			Description: "table of recipes, all recipes **must include {{.fruit}}** in ingredients",
			Columns: []*table.TableGenColumn{
				{Name: "name", Description: "recipe name", Type: "string", FillMode: "ai", ContextLength: 5},
				{
					Name:        "ingredients",
					Description: "list of ingredients, must include {{.fruit}}",
					Type:        "array",
					FillMode:    "ai",
				},
				{
					Name:        "instructions",
					Description: "list of steps for this recipe, try as detailed as possible for beginners",
					Type:        "array",
					FillMode:    "ai",
				},
			},
		},
	}
	genratePayload := &workflow.WorkflowGeneratePayload{Table: "{{.fruit}}_recipes", Count: 5, Batch: 2}
	exportPayload := &workflow.WorkflowExportTableParams{Table: "{{.fruit}}_recipes"}

	m1, err := json.Marshal(createTablePayload)
	if err != nil {
		return err
	}
	m2, err := json.Marshal(genratePayload)
	if err != nil {
		return err
	}
	m3, err := json.Marshal(exportPayload)
	if err != nil {
		return err
	}

	_, err = workflowService.Create(ctx, &workflow.Workflow{
		Name:        "fruit recipe gen",
		Description: "generate a fruit recipe based on selected fruit",
		Variables: []schema.WorkflowVariable{
			{
				Name: "fruit", Type: schema.WorkflowVariableTypeString,
				DefaultValue: "apple",
				Options:      []any{"apple", "banana", "orange", "grape", "watermelon", "strawberry"},
			},
		},
		Steps: []schema.WorkflowStep{
			{Type: schema.WorkflowStepTypeCreateTable, Payload: json.RawMessage(m1)},
			{Type: schema.WorkflowStepTypeGenerate, Payload: json.RawMessage(m2)},
			{Type: schema.WorkflowStepTypeExportTable, Payload: json.RawMessage(m3)},
		},
	})
	return err
}
