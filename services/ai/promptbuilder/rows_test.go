package promptbuilder

import (
	"tablepilot/ent"
	"tablepilot/ent/tablecolumn"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_RowsBuilder(t *testing.T) {
	builder := NewRowsBuilder(5)
	builder.AddDescription("rows gen")
	builder.AddMissingColumns([]*ent.TableColumn{
		{Name: "c1", Nanoid: "n1", Description: "col1", Type: tablecolumn.TypeString},
		{Name: "c2", Nanoid: "n2", Description: "col2", Type: tablecolumn.TypeInteger},
	})
	err := builder.AddColumnContextData("n1", []any{
		"abc", 123, []string{"go"}, map[string]any{"name": "bar"},
	})
	require.NoError(t, err)
	prompt, err := builder.Prompt()
	require.NoError(t, err)
	expected := `Give me 5 new rows for the table in JSON array format.
<TableDescription>rows gen</TableDescription>
Generate values for the following missing columns:
<MissingColumns>
  <Column id="n1" name="c1" description="col1" type="string"/>
  <Column id="n2" name="c2" description="col2" type="integer"/>
</MissingColumns>
Consider the following existing values for column n1, collected from previous rows. Try not to repeat any of these values in your output for column n1:
<Values column_id="n1">
  <Value>abc</Value>
  <Value>123</Value>
  <Value>[&quot;go&quot;]</Value>
  <Value>{&quot;name&quot;:&quot;bar&quot;}</Value>
</Values>
`
	require.Equal(t, expected, prompt)
}

func TestPromptBuilder_RowsBuilderExists(t *testing.T) {
	builder := NewRowsBuilder(5)
	builder.AddDescription("rows gen")
	builder.AddMissingColumns([]*ent.TableColumn{
		{Name: "c1", Nanoid: "n1", Description: "col1", Type: tablecolumn.TypeString},
		{Name: "c2", Nanoid: "n2", Description: "col2", Type: tablecolumn.TypeInteger},
	})
	err := builder.AddColumnContextData("n1", []any{
		"abc", 123, []string{"go"}, map[string]any{"name": "bar"},
	})
	require.NoError(t, err)
	err = builder.AddExistings([]map[string]any{{"n2": "foo"}, {"n2": "foofoo"}})
	require.NoError(t, err)
	prompt, err := builder.Prompt()
	require.NoError(t, err)
	expected := `<TableDescription>rows gen</TableDescription>
Generate values for the following missing columns:
<MissingColumns>
  <Column id="n1" name="c1" description="col1" type="string"/>
  <Column id="n2" name="c2" description="col2" type="integer"/>
</MissingColumns>
Consider the following existing values for column n1, collected from previous rows. Try not to repeat any of these values in your output for column n1:
<Values column_id="n1">
  <Value>abc</Value>
  <Value>123</Value>
  <Value>[&quot;go&quot;]</Value>
  <Value>{&quot;name&quot;:&quot;bar&quot;}</Value>
</Values>
Below is the rows data, each row contains existing columns data, and help me fill missing columns for each row. In the return rows array, provide id field and missing column data.
<Rows>
  <Row id="0">{&quot;n2&quot;:&quot;foo&quot;}</Row>
  <Row id="1">{&quot;n2&quot;:&quot;foofoo&quot;}</Row>
</Rows>
`
	require.Equal(t, expected, prompt)
}
