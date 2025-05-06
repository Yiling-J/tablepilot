package table

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/tidwall/gjson"
)

func ErrColumnNotFound(input string) error {
	return fmt.Errorf("column not found: %s", input)
}

func validateLinkedColumnInfo(ctx context.Context, tx *ent.Tx, columns []TableGenColumn, sources map[string]json.RawMessage) error {
	// validate linked column column/context_columns exists
	for _, col := range columns {
		if col.FillMode != "pick" {
			continue
		}
		source, ok := sources[col.Source]
		if !ok {
			return fmt.Errorf("source %s not found", col.Source)
		}
		if gjson.GetBytes(source, "type").String() != "linked" {
			continue
		}
		table := gjson.GetBytes(source, "table").String()
		linkedTable, err := tx.TableMeta.Query().WithColumns(func(tcq *ent.TableColumnQuery) {
			tcq.Order(ent.Asc(tablecolumn.FieldID))
		}).Where(tablemeta.Or(
			tablemeta.Nanoid(table),
			tablemeta.Name(table),
		)).First(ctx)
		if err != nil {
			return err
		}

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
	}
	return nil
}
