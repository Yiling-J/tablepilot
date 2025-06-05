package table

import (
	"context"
	"fmt"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/dataset"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
)

func ErrColumnNotFound(input string) error {
	return fmt.Errorf("column not found: %s", input)
}

func validateLinkedColumnInfo(ctx context.Context, db *ent.Client, tableID int, columns []*TableGenColumn) error {
	// validate linked column column/context_columns exists
	for _, col := range columns {
		if col.FillMode != "pick" {
			continue
		}

		switch col.SourceType {
		case tablecolumn.SourceTypeTable:
			linkedTable, err := db.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
				tcq.Order(ent.Asc(tablecolumn.FieldID))
			}).Where(tablemeta.Or(
				tablemeta.Nanoid(col.SourceID),
				tablemeta.Name(col.SourceID),
			)).First(ctx)
			if err != nil {
				return fmt.Errorf("source table %s not found", col.SourceID)
			}
			col.SourceID = linkedTable.Nanoid
			names := map[string]bool{}
			ids := map[string]bool{}

			for _, dbcol := range linkedTable.Edges.Columns {
				names[dbcol.Name] = true
				ids[dbcol.Nanoid] = true
			}

			if !names[col.LinkedColumn] && !ids[col.LinkedColumn] {
				return ErrColumnNotFound(col.LinkedColumn)
			}
			if len(col.LinkedContextColumns) > 0 {
				for _, lc := range col.LinkedContextColumns {
					if !names[lc] && !ids[lc] {
						return ErrColumnNotFound(lc)
					}
				}
			}
		case tablecolumn.SourceTypeDataset:
			linkedDataset, err := db.Dataset.Query().Where(dataset.Or(
				dataset.Nanoid(col.SourceID),
				dataset.Name(col.SourceID),
			)).First(ctx)
			if err != nil {
				return err
			}
			col.SourceID = linkedDataset.Nanoid
			if linkedDataset.Type == dataset.TypeCsv {
				names := map[string]bool{}
				for _, col := range linkedDataset.Indexer.ColumnNames {
					names[col] = true
				}
				if !names[col.LinkedColumn] {
					return ErrColumnNotFound(col.LinkedColumn)
				}
				if len(col.LinkedContextColumns) > 0 {
					for _, lc := range col.LinkedContextColumns {
						if !names[lc] {
							return ErrColumnNotFound(lc)
						}
					}
				}
			}
			if linkedDataset.Private && tableID > 0 {
				err = linkedDataset.Update().SetTableID(tableID).Exec(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
