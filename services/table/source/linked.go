package source

import (
	"context"
	"errors"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/schema"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"
	"github.com/Yiling-J/tablepilot/ent/tablerow"
	"github.com/Yiling-J/tablepilot/services/table/util"
)

type LinkedSource struct {
	Type    string `json:"type"`
	db      *ent.Client
	Table   string `json:"table"`
	data    []*ent.TableRow
	columns []*ent.TableColumn
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

func (ls *LinkedSource) getLinkedCellValueByIndex(idx int, column string, contextColumns []string) (*schema.CellValue, error) {
	indexer := util.NewColumnIndexer(ls.columns)
	row := ls.data[idx]
	lv := &schema.CellValue{}
	idx, err := indexer.GetColumnIndexByNanoid(column)
	if err != nil {
		return nil, err
	}
	lv.Value = row.Cells[idx].Value

	if len(contextColumns) > 0 {
		values := map[string]any{}
		for _, c := range contextColumns {
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

func (ls *LinkedSource) GetLinkedCellValue(row *ent.TableRow, column string, contextColumns []string) *schema.CellValue {
	indexer := util.NewColumnIndexer(ls.columns)
	ids := map[string]bool{}
	names := map[string]string{}
	if len(contextColumns) == 0 {
		for _, col := range ls.columns {
			contextColumns = append(contextColumns, col.Nanoid)
		}
	}
	for _, col := range ls.columns {
		ids[col.Nanoid] = true
		names[col.Name] = col.Nanoid
	}

	// column name -> column id
	if v, ok := names[column]; ok {
		column = v
	}
	for i, c := range contextColumns {
		if v, ok := names[c]; ok {
			contextColumns[i] = v
		}
	}

	cv := &schema.CellValue{}
	idx, err := indexer.GetColumnIndexByNanoid(column)
	if err != nil {
		return cv
	}
	cv.Value = row.Cells[idx].Value

	if len(contextColumns) > 0 {
		values := map[string]any{}
		for _, c := range contextColumns {
			cc, err := indexer.GetColumnByNanoid(c)
			if err != nil {
				continue
			}
			idx, err := indexer.GetColumnIndexByNanoid(c)
			if err != nil {
				continue
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
		cv.ContextValue = values
	}
	return cv
}

func (ls *LinkedSource) Next(ctx context.Context, idx int) (*schema.CellValue, error) {
	return nil, errors.New("not implemented")
}

func (ls *LinkedSource) NextLinked(ctx context.Context, idx int, column string, contextColumns []string) (*schema.CellValue, error) {
	ids := map[string]bool{}
	names := map[string]string{}
	for _, col := range ls.columns {
		ids[col.Nanoid] = true
		names[col.Name] = col.Nanoid
	}

	// column name -> column id
	if v, ok := names[column]; ok {
		column = v
	}
	for i, c := range contextColumns {
		if v, ok := names[c]; ok {
			contextColumns[i] = v
		}
	}

	return ls.getLinkedCellValueByIndex(idx, column, contextColumns)
}

func (ls *LinkedSource) Total() int {
	return len(ls.data)
}

func (ls *LinkedSource) Range(fn func(row *ent.TableRow) bool) {
	for _, row := range ls.data {
		if !fn(row) {
			break
		}
	}
}
