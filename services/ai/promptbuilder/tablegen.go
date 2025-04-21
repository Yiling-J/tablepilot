package promptbuilder

import (
	"fmt"
)

type TableColumnSimple struct {
	Name        string
	Description string
}

type TableInfoSimple struct {
	Name        string
	Description string
	Columns     []TableColumnSimple
}

type TableGenBuilder struct {
	Builder
}

func NewTableGenBuilder(name, description string, depends []string, tables []TableInfoSimple) *TableGenBuilder {
	tb := &TableGenBuilder{}
	tb.AddText(tableGenPrompt)
	el := tb.NewXML("Name")
	el.CreateText(name)
	el = tb.NewXML("Description")
	el.CreateText(description)
	el = tb.NewXML("Depends")
	for _, dp := range depends {
		el.CreateText(fmt.Sprintf("<Table name=\"%s\">", dp))
	}
	el = tb.NewXML("ExistingTables")
	for _, tb := range tables {
		et := el.CreateElement("Table")
		et.CreateAttr("name", tb.Name)
		et.CreateAttr("description", tb.Description)
		for _, col := range tb.Columns {
			ec := et.CreateElement("Column")
			ec.CreateAttr("name", col.Name)
			ec.CreateAttr("description", col.Description)
		}
	}
	tb.AddText("Create a table based on <Name>, <Description> and <ExistingTables>. You are smart and you should decide the proper columns in this table, don't ask me.")
	return tb
}

const tableGenPrompt = `
#### Columns

Each column object contain the following fields:

- **name**: The name of the column (e.g., "Name", "Ingredients"). This will be used in the prompt when generating rows.
- **description**: A brief description of what data the column contains (e.g., "recipe name"). This will also be used in the prompt when generating rows.
- **type**: The data type for the column. Possible values include:
	- "string": For text values.
	- "array": For lists.
	- "integer": For integral numbers.
	- "number": For any numeric type, either integers or floating point numbers.
	- "image": For image generation.
- **fill_mode**: Specifies how the column is populated. Possible values:
	- "ai": AI will generate values for this column.
	- "pick": Values are picked from an existing source (e.g., a list of cuisines).
- **context_length** (Optional): Defines how many previous values in this column will be sent to the LLM when generating a new batch of rows. This helps provide context for the generation. If you aim for diverse results, using tag-like columns from the source may be more effective than increasing the context length. The context_length parameter is best used to ensure consistency in generation format and should typically remain moderate rather than excessively large.
- **source** (Optional): Specifies the source to pull data from when "fill_mode" is set to "pick". This should match a source name defined in the "sources" section (e.g., "cuisines").

**Additional Fields for "pick" Mode**

When "fill_mode" is set to "pick", the following fields are available:

- **random**: If true, a random value is selected for each row from all available options in the source. Default: false.
- **replacement**: Determines whether sampling is with or without replacement:
  - true: Items can be selected multiple times.
  - false: Once an item is selected, it cannot be chosen again.
  - Default: false.
- **repeat**: Specifies how many times a picked value is reused before switching to the next one. The minimum and default value is 1, meaning each value is used once before moving to the next.

When source type is "linked", the following fields are available:

- **linked_column**: The linked-table column used for display text in the generated cell(e.g., user name).
- **linked_context_columns**: The linked-table columns providing context when generating data (e.g., user age, job, nationality). Note that only values from linked_context_columns are sent as context, not from linked_column. If you want a column to be both displayed and sent as context, it must be included in both linked_column and linked_context_columns.

#### Sources

A sourc object from which pick-type columns can select values. There are currently 3 types of sources:

- **ai**: Uses AI to generate a list of options dynamically.
- **list**: Uses a predefined list of options.
- **linked**: Uses rows from another table as the source.

Each source is an object with the following fields:

Common fields:

- **name**: The name of the source (e.g., "cuisines").
- **type**: The type of the source, which can be "ai", "list", "linked".

Special fields for different types:

- **ai**:
	- **prompt**: The prompt used to generate options by AI, e.g., Give me 50 common ingredients.
- **list**:
	- **options**: A list of predefined options to pick from.
- **linked**:
	- **table**: The name of the linked table.

#### Functions

Here are functions to create a table:
- AddAiColumn(name, description, type): Add a ai type column.
- AddPickColumn(name, description, type, random, repeat, source, linkedColumn, linkedContextColumns): add a pick type column.
- AddListSource(name, options): add a list type source.
- AddAiSource(name,prompt): add a ai type source.
- AddLinkedSource(name, table): add a linked type source.

The "depends" array list other table names that this table **relies on for meaningful data**. For each depend table, create a pick type column and related source using "AddLinkedSource" and "AddPickColumn" functions, based on the info in <ExistingTables>.`

type TablePolishBuilder struct {
	Builder
}

func NewTablePolishBuilder(prompt string, sources string, columns string) *TablePolishBuilder {
	tb := &TablePolishBuilder{}
	tb.AddText(tablePolishPrompt)
	el := tb.NewXML("SourcesJson")
	el.CreateText(sources)
	el = tb.NewXML("ColumnsJson")
	el.CreateText(columns)
	el = tb.NewXML("Requirement")
	el.CreateText(prompt)
	tb.AddText("Now Update the table schema based on <Requirement>. Existing sourcs or columns of the table are in <SourcesJson> and <ColumnsJson>.")
	return tb
}

const tablePolishPrompt = `Update the  columns or sources in table schema based on user requirements.

#### Columns

Each column object contain the following fields:

- **name**: The name of the column (e.g., "Name", "Ingredients"). This will be used in the prompt when generating rows.
- **description**: A brief description of what data the column contains (e.g., "recipe name"). This will also be used in the prompt when generating rows.
- **type**: The data type for the column. Possible values include:
	- "string": For text values.
	- "array": For lists.
	- "integer": For integral numbers.
	- "number": For any numeric type, either integers or floating point numbers.
	- "image": For image generation.
- **fill_mode**: Specifies how the column is populated. Possible values:
	- "ai": AI will generate values for this column.
	- "pick": Values are picked from an existing source (e.g., a list of cuisines).
- **context_length** (Optional): Defines how many previous values in this column will be sent to the LLM when generating a new batch of rows. This helps provide context for the generation. If you aim for diverse results, using tag-like columns from the source may be more effective than increasing the context length. The context_length parameter is best used to ensure consistency in generation format and should typically remain moderate rather than excessively large.
- **source** (Optional): Specifies the source to pull data from when "fill_mode" is set to "pick". This should match a source name defined in the "sources" section (e.g., "cuisines").

**Additional Fields for "pick" Mode**

When "fill_mode" is set to "pick", the following fields are available:

- **random**: If true, a random value is selected for each row from all available options in the source. Default: false.
- **replacement**: Determines whether sampling is with or without replacement:
  - true: Items can be selected multiple times.
  - false: Once an item is selected, it cannot be chosen again.
  - Default: false.
- **repeat**: Specifies how many times a picked value is reused before switching to the next one. The minimum and default value is 1, meaning each value is used once before moving to the next.

When source type is "linked", the following fields are available:

- **linked_column**: The linked-table column used for display text in the generated cell(e.g., user name).
- **linked_context_columns**: The linked-table columns providing context when generating data (e.g., user age, job, nationality). Note that only values from linked_context_columns are sent as context, not from linked_column. If you want a column to be both displayed and sent as context, it must be included in both linked_column and linked_context_columns.

#### Sources

A sourc object from which pick-type columns can select values. There are currently 3 types of sources:

- **ai**: Uses AI to generate a list of options dynamically.
- **list**: Uses a predefined list of options.
- **linked**: Uses rows from another table as the source.

Each source is an object with the following fields:

Common fields:

- **name**: The name of the source (e.g., "cuisines").
- **type**: The type of the source, which can be "ai", "list", "linked".

Special fields for different types:

- **ai**:
	- **prompt**: The prompt used to generate options by AI, e.g., Give me 50 common ingredients.
- **list**:
	- **options**: A list of predefined options to pick from.
- **linked**:
	- **table**: The name of the linked table.

Here are functions available, which you can use to update table schema:
- AddAiColumn(name, description, type)
- AddPickColumn(name, description, type, random, repeat, source, linkedColumn, linkedContextColumns)
- AddListSource(name, options)
- AddAiSource(name,prompt)
- AddLinkedSource(name, table)
- RemoveColumn(name)
- RemoveSource(name)

The "depends" array list other table names that this table **relies on for meaningful data**. For each depend table, create a pick type column and related source using "AddLinkedSource" and "AddPickColumn" functions, based on the info in <ExistingTables>. Example: If current table is "" and depends is [""], you should first find the "" table in <ExistingTables>, then find the suitable linkedColumn and linkedContextColumns, in this case is "" and [""], then create the source first: AddLinkedSource. then Create the column: AddPickColumn.`
