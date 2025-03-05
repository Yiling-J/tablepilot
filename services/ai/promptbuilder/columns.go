package promptbuilder

import (
	"fmt"
)

type Column struct {
	Name        string
	Description string
}

type ColumnsBuilder struct {
	Builder
	count int
}

func NewColumnsBuilder(count int, tableName, tableDescription string) *ColumnsBuilder {
	pb := &ColumnsBuilder{count: count}
	pb.AddText(columnsGenPrompt)
	pb.AddText("### Table Information:")
	el := pb.NewXML("TableName")
	el.CreateText(tableName)
	el = pb.NewXML("TableDescription")
	el.CreateText(tableDescription)
	return pb
}

func (c *ColumnsBuilder) AddExistingColumns(columns []Column) {
	c.AddText("### Existing Columns:")
	el := c.NewXML("Columns")
	for _, col := range columns {
		cel := el.CreateElement("Column")
		cel.CreateAttr("name", col.Name)
		cel.CreateAttr("description", col.Description)
	}
}

func (c *ColumnsBuilder) Prompt() (string, error) {
	c.AddText(fmt.Sprintf("Now generate %d extra colums.", c.count))
	return c.Builder.Prompt()
}

const columnsGenPrompt = `I'm designing a database table, I will tell you existing columns I have,
and you help me generate extra columns. Return a list of extra columns.
### Rules:
- Do NOT add a sequence ID or unique identifier column (I will add this manually).
- Column names should use spaces (e.g., "Recipe Name").
- "type" can be one of: "string", "number", "integer", "array", or "boolean".
- Each column must have a description explaining what information it contains.
- Ensure the generated columns are **relevant** to the table's purpose.`
