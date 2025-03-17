package source

import (
	"context"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/tablerow"
	"github.com/Yiling-J/tablepilot/services/table/util"
)

type LinkedSource struct {
	Type           string `json:"type"`
	db             *ent.Client
	Table          string   `json:"table"`
	Column         string   `json:"column"`
	ContextColumns []string `json:"context_columns"`
	data           []*ent.TableRow
	columns        []*ent.TableColumn
}

func (ls *LinkedSource) Init(ctx context.Context, db *ent.Client) error {
	ls.db = db
	data, err := ls.db.TableRow.Query().Where(
		tablerow.HasTablemetaWith(tablemeta.Nanoid(ls.Table)),
	).Order(ent.Asc(tablerow.FieldID)).All(ctx)
	if err != nil {
		return err
	}
	ls.data = data
	columns, err := ls.db.TableColumn.Query().Where(
		tablecolumn.HasTablemetaWith(tablemeta.Nanoid(ls.Table)),
	).Order(ent.Asc(tablerow.FieldID)).All(ctx)
	if err != nil {
		return err
	}
	ls.columns = columns
	return nil
}

func (ls *LinkedSource) getLinkedCellValue(idx int) (*schema.CellValue, error) {
	indexer := util.NewColumnIndexer(ls.columns)
	row := ls.data[idx]
	lv := &schema.CellValue{}
	idx, err := indexer.GetColumnIndexByNanoid(ls.Column)
	if err != nil {
		return nil, err
	}
	lv.Value = row.Cells[idx].Value

	if len(ls.ContextColumns) > 0 {
		values := map[string]any{}
		for _, c := range ls.ContextColumns {
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

func (ls *LinkedSource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return ls.getLinkedCellValue(idx)
}

func (ls *LinkedSource) Total() int {
	return len(ls.data)
}
