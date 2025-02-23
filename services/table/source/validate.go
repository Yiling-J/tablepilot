package source

import (
	"context"
	"encoding/json"
	"errors"
	"tablepilot/ent"
	"tablepilot/ent/tablecolumn"
	"tablepilot/ent/tablemeta"

	"github.com/tidwall/gjson"
)

func ValidateSource(ctx context.Context, raw json.RawMessage, db *ent.Client) (Source, error) {
	var s Source
	sourceType := gjson.GetBytes(raw, "type").String()
	switch sourceType {
	case "list":
		var ls ListSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		if len(ls.Options) == 0 {
			return nil, errors.New("no options")
		}
		s = &ls
	case "ai":
		var ls AISource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		if len(ls.Prompt) == 0 {
			return nil, errors.New("empty prompt")
		}
		s = &ls
	case "linked":
		var ls LinkedSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		if ls.Table == "" {
			return nil, ErrTableNameOrIdEmpty()
		}
		tb, err := db.TableMeta.Query().Where(tablemeta.Or(
			tablemeta.Name(ls.Table),
			tablemeta.Nanoid(ls.Table),
		)).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, ErrTableNotFound(ls.Table)
			}
			return nil, err
		}
		ls.Table = tb.Nanoid
		if ls.Column == "" {
			return nil, ErrColumnNameOrIdEmpty()
		}
		col, err := db.TableColumn.Query().Where(
			tablecolumn.HasTablemetaWith(tablemeta.Nanoid(tb.Nanoid)),
			tablecolumn.Or(
				tablecolumn.Name(ls.Column),
				tablecolumn.Nanoid(ls.Column),
			)).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil, ErrColumnNotFound(ls.Column)
			}
			return nil, err
		}
		ls.Column = col.Nanoid
		for i, c := range ls.ContextColumns {
			col, err := db.TableColumn.Query().Where(
				tablecolumn.HasTablemetaWith(tablemeta.Nanoid(tb.Nanoid)),
				tablecolumn.Or(
					tablecolumn.Name(c),
					tablecolumn.Nanoid(c),
				)).Only(ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					return nil, ErrColumnNotFound(c)
				}
				return nil, err
			}
			ls.ContextColumns[i] = col.Nanoid
		}
		s = &ls
	}
	return s, nil
}
