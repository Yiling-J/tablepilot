package table

import (
	"context"

	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
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
	case "AddPickFromTableColumn":
		columnName := cast.ToString(args["name"])
		columnDescription := cast.ToString(args["description"])
		columnType := cast.ToString(args["type"])
		contextLength := cast.ToInt(args["contextLength"])
		random := cast.ToBool(args["random"])
		repeat := cast.ToInt(args["repeat"])
		replacement := cast.ToBool(args["replacement"])
		linkedColumn := cast.ToString(args["linkedColumn"])
		linkedContextColumns := cast.ToStringSlice(args["linkedContextColumns"])
		table := cast.ToString(args["table"])
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
			SourceID:             table,
			SourceType:           tablecolumn.SourceTypeTable,
		})
	case "AddPickFromOptionsColumn":
		columnName := cast.ToString(args["name"])
		columnDescription := cast.ToString(args["description"])
		columnType := cast.ToString(args["type"])
		contextLength := cast.ToInt(args["contextLength"])
		random := cast.ToBool(args["random"])
		repeat := cast.ToInt(args["repeat"])
		replacement := cast.ToBool(args["replacement"])
		options := cast.ToStringSlice(args["options"])
		tb.table.Columns = append(tb.table.Columns, &TableGenColumn{
			Name:          columnName,
			Description:   columnDescription,
			Type:          columnType,
			FillMode:      "pick",
			ContextLength: contextLength,
			Random:        random,
			Repeat:        repeat,
			Replacement:   replacement,
			SourceType:    tablecolumn.SourceTypeOptions,
			Options:       options,
		})
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
