package source

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
		if len(ls.Options) == 0 && ls.File == "" {
			return nil, fmt.Errorf("souce %s options should not be empty", ls.Name)
		}
		s = &ls
	case "ai":
		var ls AISource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		if len(ls.Prompt) == 0 {
			return nil, fmt.Errorf("souce %s prompt should not be empty", ls.Name)
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
	case "parquet":
		var ls ParquetSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		if ls.Huggingface == nil && len(ls.Paths) == 0 {
			return nil, errors.New("paths is empty")
		}
		if ls.Huggingface != nil && ls.Huggingface.Dataset == "" {
			return nil, errors.New("hugging Face dataset is empty")
		}
		s = &ls
	case "files":
		var ls FilesSource
		err := json.Unmarshal(raw, &ls)
		if err != nil {
			return nil, err
		}
		if len(ls.Paths) == 0 {
			return nil, errors.New("paths is empty")
		}
		s = &ls
	default:
		return nil, fmt.Errorf("invalid source type %s", sourceType)
	}
	return s, nil
}
