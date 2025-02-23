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
- Don't add a sequence id or unique identifier column, I will add manully later.
- column name should be spaced e.g., Recipe Name.
- "type" can be either "string", "number", "integer", "array" or "boolean".
- Add a short description to each column tell what info this column has.
<TableName>foo</TableName>
<TableDescription>bar</TableDescription>
<Columns>
  <Column name="foo" description="the foo"/>
  <Column name="bar" description="the bar"/>
</Columns>
Now generate 5 extra colums.
`
	require.Equal(t, expected, prompt)
}
