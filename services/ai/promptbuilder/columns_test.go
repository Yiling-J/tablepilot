package promptbuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_ColumnsBuilder(t *testing.T) {
	builder := NewColumnsBuilder(5, "foo", "bar")
	builder.AddExistingColumns([]Column{{"foo", "the foo"}, {"bar", "the bar"}})
	prompt, err := builder.Prompt()
	require.NoError(t, err)
	expected := `I'm designing a database table, I will tell you existing columns I have,
and you help me generate extra columns. Return a list of extra columns.
### Rules:
- Do NOT add a sequence ID or unique identifier column (I will add this manually).
- Column names should use spaces (e.g., "Recipe Name").
- "type" can be one of: "string", "number", "integer", "array", or "boolean".
- Each column must have a description explaining what information it contains.
- Ensure the generated columns are **relevant** to the table's purpose.
### Table Information:
<TableName>foo</TableName>
<TableDescription>bar</TableDescription>
### Existing Columns:
<Columns>
  <Column name="foo" description="the foo"/>
  <Column name="bar" description="the bar"/>
</Columns>
Now generate 5 extra colums.
`
	require.Equal(t, expected, prompt)
}
