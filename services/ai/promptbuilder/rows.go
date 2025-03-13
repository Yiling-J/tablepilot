package promptbuilder

import (
	"encoding/json"
	"fmt"

	"github.com/Yiling-J/tablepilot/ent"

	"github.com/spf13/cast"
)

type RowsBuilder struct {
	Builder
	count    int
	existing map[string][]any
}

func NewRowsBuilder(count int) *RowsBuilder {
	p := fmt.Sprintf(
		"Give me %d new rows for the table in JSON array format.", count,
	)
	rb := &RowsBuilder{existing: map[string][]any{}, count: count}
	rb.AddText(p)
	return rb
}

func (rb *RowsBuilder) AddDescription(v string) {
	xml := rb.NewXML("TableDescription")
	xml.SetText(v)
}

func (rb *RowsBuilder) AddExistings(rows []map[string]any) error {
	if len(rows) == 0 {
		return nil
	}
	rb.AddText("Below is the rows data, each row contains existing columns data, and help me fill missing columns for each row. In the return rows array, provide id field and missing column data.")
	rx := rb.NewXML("Rows")
	for i, row := range rows {
		r := rx.CreateElement("Row")
		r.CreateAttr("id", cast.ToString(i))
		b, err := json.Marshal(row)
		if err != nil {
			return err
		}
		r.CreateText(string(b))
	}
	rb.elements = rb.elements[1:]
	return nil
}

func (rb *RowsBuilder) AddTableColumns(v []*ent.TableColumn) {
	if len(v) == 0 {
		return
	}
	rb.AddText("Columns of the table:")
	el := rb.NewXML("Columns")
	// add id column
	cel := el.CreateElement("Column")
	cel.CreateAttr("id", "id")
	cel.CreateAttr("name", "id")
	cel.CreateAttr("description", "index of the row, always starting from 0 in each generation")
	cel.CreateAttr("type", "integer")
	for _, col := range v {
		cel := el.CreateElement("Column")
		cel.CreateAttr("id", col.Nanoid)
		cel.CreateAttr("name", col.Name)
		cel.CreateAttr("description", col.Description)
		cel.CreateAttr("type", string(col.Type))
	}
}

func (rb *RowsBuilder) AddMissingColumns(v []*ent.TableColumn) {
	if len(v) == 0 {
		return
	}
	rb.AddText("Generate values for the following missing columns:")
	el := rb.NewXML("MissingColumns")
	// add id column
	cel := el.CreateElement("Column")
	cel.CreateAttr("id", "id")
	for _, col := range v {
		cel := el.CreateElement("Column")
		cel.CreateAttr("id", col.Nanoid)
	}
}

func (rb *RowsBuilder) AddColumnContextData(columnId string, values []any) error {
	if len(values) == 0 {
		return nil
	}
	rb.AddText(fmt.Sprintf("Consider the following existing values for column %s, collected from previous rows. Try not to repeat any of these values in your output for column %s:", columnId, columnId))
	el := rb.NewXML("Values")
	el.CreateAttr("column_id", columnId)
	for _, v := range values {
		val := el.CreateElement("Value")
		var vs string
		switch tmp := v.(type) {
		case string:
			vs = tmp
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return err
			}
			vs = string(b)
		}
		val.CreateText(vs)
	}
	return nil
}
