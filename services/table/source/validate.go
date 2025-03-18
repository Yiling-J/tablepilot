package source

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/tablemeta"

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
		s = &ls
	case "csv":
		var ls CsvSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		if len(ls.Paths) == 0 {
			return nil, errors.New("paths is empty")
		}
		s = &ls
	}
	return s, nil
}
