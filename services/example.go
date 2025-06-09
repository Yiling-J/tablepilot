package services

import (
	"context"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/services/dataset"
	"github.com/Yiling-J/tablepilot/services/table"
)

func createAppStartExample(ctx context.Context, tableService table.TableService, datasetService dataset.DatasetService) error {
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
	return err
}
