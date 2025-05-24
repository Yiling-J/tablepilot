package promptbuilder

import (
	"testing"

	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/tablecolumn"
	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_ImageToTableBuilder(t *testing.T) {
	builder := NewNewImageToTableBuilder("")
	prompt, err := builder.Prompt()
	require.NoError(t, err)
	expected := `Please extract a table from the provided image. All extracted data should be placed in a single table with a consistent schema.
<ExampleOutput>{&quot;table_name&quot;:&quot;user_info&quot;,&quot;table_description&quot;:&quot;A table of basic user information&quot;,&quot;columns&quot;:[{&quot;name&quot;:&quot;name&quot;,&quot;description&quot;:&quot;The full name of the user&quot;,&quot;type&quot;:&quot;string&quot;},{&quot;name&quot;:&quot;age&quot;,&quot;description&quot;:&quot;The age of the user in years&quot;,&quot;type&quot;:&quot;integer&quot;},{&quot;name&quot;:&quot;job&quot;,&quot;description&quot;:&quot;The occupation of the user&quot;,&quot;type&quot;:&quot;string&quot;}],&quot;rows&quot;:[[&quot;Jack&quot;,24,&quot;Doctor&quot;],[&quot;Amy&quot;,36,&quot;Teacher&quot;]]}
</ExampleOutput>
- table_name **must** start with a letter and contain only letters, numbers, or underscores.
`
	require.Equal(t, expected, prompt)

	builder = NewNewImageToTableBuilder("gogo")
	prompt, err = builder.Prompt()
	require.NoError(t, err)
	expected = `Please extract a table from the provided image. All extracted data should be placed in a single table with a consistent schema.
## User Requirements:
<Requirement>gogo</Requirement>
<ExampleOutput>{&quot;table_name&quot;:&quot;user_info&quot;,&quot;table_description&quot;:&quot;A table of basic user information&quot;,&quot;columns&quot;:[{&quot;name&quot;:&quot;name&quot;,&quot;description&quot;:&quot;The full name of the user&quot;,&quot;type&quot;:&quot;string&quot;},{&quot;name&quot;:&quot;age&quot;,&quot;description&quot;:&quot;The age of the user in years&quot;,&quot;type&quot;:&quot;integer&quot;},{&quot;name&quot;:&quot;job&quot;,&quot;description&quot;:&quot;The occupation of the user&quot;,&quot;type&quot;:&quot;string&quot;}],&quot;rows&quot;:[[&quot;Jack&quot;,24,&quot;Doctor&quot;],[&quot;Amy&quot;,36,&quot;Teacher&quot;]]}
</ExampleOutput>
- table_name **must** start with a letter and contain only letters, numbers, or underscores.
`
	require.Equal(t, expected, prompt)

	builder = NewNewImageToTableBuilder("gogo")
	builder.AddExistingTableNames([]string{"t1", "t2"})
	prompt, err = builder.Prompt()
	require.NoError(t, err)
	expected = `Please extract a table from the provided image. All extracted data should be placed in a single table with a consistent schema.
## User Requirements:
<Requirement>gogo</Requirement>
<ExampleOutput>{&quot;table_name&quot;:&quot;user_info&quot;,&quot;table_description&quot;:&quot;A table of basic user information&quot;,&quot;columns&quot;:[{&quot;name&quot;:&quot;name&quot;,&quot;description&quot;:&quot;The full name of the user&quot;,&quot;type&quot;:&quot;string&quot;},{&quot;name&quot;:&quot;age&quot;,&quot;description&quot;:&quot;The age of the user in years&quot;,&quot;type&quot;:&quot;integer&quot;},{&quot;name&quot;:&quot;job&quot;,&quot;description&quot;:&quot;The occupation of the user&quot;,&quot;type&quot;:&quot;string&quot;}],&quot;rows&quot;:[[&quot;Jack&quot;,24,&quot;Doctor&quot;],[&quot;Amy&quot;,36,&quot;Teacher&quot;]]}
</ExampleOutput>
- table_name **must** start with a letter and contain only letters, numbers, or underscores.
- Avoid using table names that already exist: <tables>t1,t2</tables>
`
	require.Equal(t, expected, prompt)

	builder = NewNewImageToTableBuilder("gogo")
	builder.AddExistingTableNames([]string{})
	builder.ToTable(&ent.TableMeta{Edges: ent.TableMetaEdges{
		Columns: []*ent.TableColumn{
			{Name: "col1", Nanoid: "id1", FillMode: tablecolumn.FillModeAi, Description: "foo"},
			{Name: "col2", Nanoid: "id2"},
		},
	}})
	prompt, err = builder.Prompt()
	require.NoError(t, err)
	expected = `Please extract a table from the provided image. All extracted data should be placed in a single table with a consistent schema.
## User Requirements:
<Requirement>gogo</Requirement>
<ExampleOutput>{&quot;table_name&quot;:&quot;user_info&quot;,&quot;table_description&quot;:&quot;A table of basic user information&quot;,&quot;columns&quot;:[{&quot;name&quot;:&quot;name&quot;,&quot;description&quot;:&quot;The full name of the user&quot;,&quot;type&quot;:&quot;string&quot;},{&quot;name&quot;:&quot;age&quot;,&quot;description&quot;:&quot;The age of the user in years&quot;,&quot;type&quot;:&quot;integer&quot;},{&quot;name&quot;:&quot;job&quot;,&quot;description&quot;:&quot;The occupation of the user&quot;,&quot;type&quot;:&quot;string&quot;}],&quot;rows&quot;:[[&quot;Jack&quot;,24,&quot;Doctor&quot;],[&quot;Amy&quot;,36,&quot;Teacher&quot;]]}
</ExampleOutput>
- table_name **must** start with a letter and contain only letters, numbers, or underscores.
## Below is the table info that extracted data should match.
### Table Information:
<TableName></TableName>
<TableDescription></TableDescription>
### Columns:
<Columns>
  <Column name="col1" id="id1" description="foo"/>
</Columns>
`
	require.Equal(t, expected, prompt)
}
