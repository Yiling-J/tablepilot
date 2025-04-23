package table

import (
	"context"
	"encoding/json"

	"github.com/Yiling-J/tablepilot/services/table/source"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
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
		tb.table.Columns = append(tb.table.Columns, TableGenColumn{
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
		source := cast.ToString(args["source"])
		linkedColumn := cast.ToString(args["linkedColumn"])
		linkedContextColumns := cast.ToStringSlice(args["linkedContextColumns"])
		tb.table.Columns = append(tb.table.Columns, TableGenColumn{
			Name:                 columnName,
			Description:          columnDescription,
			Type:                 columnType,
			FillMode:             "pick",
			ContextLength:        contextLength,
			Random:               random,
			Repeat:               repeat,
			Replacement:          replacement,
			Source:               source,
			LinkedColumn:         linkedColumn,
			LinkedContextColumns: linkedContextColumns,
		})
	case "AddListSource":
		source := &source.ListSource{
			BasicSource: source.BasicSource{
				Name: cast.ToString(args["name"]),
				Type: "list",
			},
			Options: cast.ToStringSlice(args["options"]),
		}
		b, err := json.Marshal(source)
		if err != nil {
			return err
		}
		tb.table.Sources = append(tb.table.Sources, b)
	case "AddAiSource":
		source := &source.AISource{
			BasicSource: source.BasicSource{
				Name: cast.ToString(args["name"]),
				Type: "ai",
			},
			Prompt: cast.ToString(args["prompt"]),
		}
		b, err := json.Marshal(source)
		if err != nil {
			return err
		}
		tb.table.Sources = append(tb.table.Sources, b)
	case "AddLinkedSource":
		source := &source.LinkedSource{
			BasicSource: source.BasicSource{
				Name: cast.ToString(args["name"]),
				Type: "linked",
			},
			Table: cast.ToString(args["table"]),
		}
		b, err := json.Marshal(source)
		if err != nil {
			return err
		}
		tb.table.Sources = append(tb.table.Sources, b)
	case "RemoveColumn":
		name := cast.ToString(args["name"])
		columns := []TableGenColumn{}
		for _, col := range tb.table.Columns {
			if col.Name == name {
				continue
			}
			columns = append(columns, col)
		}
		tb.table.Columns = columns
	case "RemoveSource":
		name := cast.ToString(args["name"])
		sources := []json.RawMessage{}
		for _, s := range tb.table.Sources {
			if gjson.GetBytes(s, "name").String() == name {
				continue
			}
			sources = append(sources, s)
		}
		tb.table.Sources = sources
	default:
	}
	return nil
}
