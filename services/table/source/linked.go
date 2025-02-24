package source

import (
	"context"
	"encoding/json"
	"tablepilot/ent"
	"tablepilot/ent/schema"
	"tablepilot/ent/tablecolumn"
	"tablepilot/ent/tablemeta"
	"tablepilot/ent/tablerow"
	"tablepilot/services/table/util"
)

type LinkedSource struct {
	indexer
	Type           string `json:"type"`
	db             *ent.Client
	Table          string   `json:"table"`
	Column         string   `json:"column"`
	ContextColumns []string `json:"context_columns"`
	data           []*ent.TableRow
	_column        *ent.TableColumn // the column which need to be get
}

func (ls *LinkedSource) Init(ctx context.Context, db *ent.Client, column *ent.TableColumn) error {
	ls.db = db
	data, err := ls.db.TableRow.Query().Where(
		tablerow.HasTablemetaWith(tablemeta.Nanoid(ls.Table)),
	).Order(ent.Asc(tablerow.FieldID)).All(ctx)
	if err != nil {
		return err
	}
	ls.data = data
	ls._column = column
	ls.indexer = newIndexer(ls.Random, ls.Replacement, len(ls.data), ls.Repeat)
	return nil
}

func linkedSource(ctx context.Context, db *ent.Client, column *ent.TableColumn) (*LinkedSource, error) {
	var s LinkedSource
	err := json.Unmarshal(column.Source, &s)
	if err != nil {
		return nil, err
	}
	s.db = db
	err = s.Init(ctx, db, column)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func GetLinkedCellValue(ctx context.Context, db *ent.Client, column *ent.TableColumn, linkedRow string) (*schema.CellValue, error) {
	so, err := linkedSource(ctx, db, column)
	if err != nil {
		return nil, err
	}
	linkedTable := so.Table
	columns, err := db.TableColumn.Query().Where(
		tablecolumn.HasTablemetaWith(tablemeta.Nanoid(linkedTable)),
	).All(ctx)
	if err != nil {
		return nil, err
	}
	indexer := util.NewColumnIndexer(columns)
	row, err := db.TableRow.Query().Where(
		tablerow.HasTablemetaWith(tablemeta.Nanoid(linkedTable)),
		tablerow.Nanoid(linkedRow),
	).First(ctx)
	if err != nil {
		return nil, err
	}
	lv := &schema.CellValue{}
	idx, err := indexer.GetColumnIndexByNanoid(so.Column)
	if err != nil {
		return nil, err
	}
	lv.Value = row.Cells[idx].Value

	if len(so.ContextColumns) > 0 {
		values := map[string]any{}
		for _, c := range so.ContextColumns {
			cc, err := indexer.GetColumnByNanoid(c)
			if err != nil {
				return nil, err
			}
			idx, err := indexer.GetColumnIndexByNanoid(c)
			if err != nil {
				return nil, err
			}
			v := row.Cells[idx].Value
			if row.Cells[idx].ContextValue != nil {
				v = row.Cells[idx].ContextValue
			}
			values[cc.Name] = map[string]any{
				"description": cc.Description,
				"data":        v,
			}
		}
		lv.ContextValue = values
	}
	return lv, nil
}

func (ls *LinkedSource) Next(ctx context.Context) (*schema.CellValue, error) {
	row := ls.data[ls.nextIndex()]
	return GetLinkedCellValue(ctx, ls.db, ls._column, row.Nanoid)
}
