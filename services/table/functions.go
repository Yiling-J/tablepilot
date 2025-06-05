package table

import (
	"context"
	"encoding/json"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/Yiling-J/tablepilot/services/source"
	"github.com/spf13/cast"
)

type tableBuilder struct {
	table *TableGenRequest
}

func (tb *tableBuilder) run(ctx context.Context, name string, args map[string]any) error {
	switch name {
	case "AddAiColumn":
		columnName := cast.ToString(args["name"])
		columnDescription := cast.ToString(args["description"])
		columnType := cast.ToString(args["type"])
		contextLength := cast.ToInt(args["contextLength"])
		tb.table.Columns = append(tb.table.Columns, &TableGenColumn{
			Name:          columnName,
			Description:   columnDescription,
			Type:          columnType,
			FillMode:      "ai",
			ContextLength: contextLength,
		})
	case "AddPickColumn":
		columnName := cast.ToString(args["name"])
		columnDescription := cast.ToString(args["description"])
		columnType := cast.ToString(args["type"])
		contextLength := cast.ToInt(args["contextLength"])
		random := cast.ToBool(args["random"])
		repeat := cast.ToInt(args["repeat"])
		replacement := cast.ToBool(args["replacement"])
		linkedColumn := cast.ToString(args["linkedColumn"])
		linkedContextColumns := cast.ToStringSlice(args["linkedContextColumns"])
		sourceID := cast.ToString(args["sourceID"])
		sourceType := cast.ToString(args["sourceType"])
		tb.table.Columns = append(tb.table.Columns, &TableGenColumn{
			Name:                 columnName,
			Description:          columnDescription,
			Type:                 columnType,
			FillMode:             "pick",
			ContextLength:        contextLength,
			Random:               random,
			Repeat:               repeat,
			Replacement:          replacement,
			LinkedColumn:         linkedColumn,
			LinkedContextColumns: linkedContextColumns,
			SourceID:             sourceID,
			SourceType:           tablecolumn.SourceType(sourceType),
		})
	case "AddListDataset":
		source := &source.ListSource{
			BasicSource: source.BasicSource{
				Name: cast.ToString(args["name"]),
				Type: "list",
			},
			Options: cast.ToStringSlice(args["options"]),
		}
		b, err := json.Marshal(source)
		_ = b
		if err != nil {
			return err
		}
	case "RemoveColumn":
		name := cast.ToString(args["name"])
		columns := []*TableGenColumn{}
		for _, col := range tb.table.Columns {
			if col.Name == name {
				continue
			}
			columns = append(columns, col)
		}
		tb.table.Columns = columns
	case "RemoveSource":
	default:
	}
	return nil
}
