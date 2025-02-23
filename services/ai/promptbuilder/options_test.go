package promptbuilder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_ColumnOptionsBuilder(t *testing.T) {
	builder := NewColumnOptionsBuilder("foo", "Bar", "gen")
	builder.AddExampleOptions([]string{"a", "b"})
	p, err := builder.Prompt()
	require.NoError(t, err)
	expected := `Please generate some options for this table Column. Don't duplicate.
<Requirement>gen</Requirement>
<Column name="foo" description="Bar"/>
<ExampleOptions>
  <Option value="a"/>
  <Option value="b"/>
</ExampleOptions>
`
	require.Equal(t, expected, p)
}
