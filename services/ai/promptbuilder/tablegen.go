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
	tb.AddText("Based on the provided `<Name>`, `<Description>`, and `<ExistingTables>`, create a table with the appropriate columns. You are expected to **decide on the appropriate columns automatically**, without asking for further input.")
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
- **linked_context_columns**: The linked-table columns used to providing context to AI when generating data (e.g., user name, user age, job, nationality). One column can be linked_column and also in linked_context_columns(and this happen very ofter), because the display value is also useful context to guide AI.

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

**Important Note**: Each table should have at least one AI-type column. Use the "AddAiColumn" function to add this column.

The "depends" array lists other tables that this table **relies on for meaningful data**. For each dependent table, you must create a pick-type column and a related source using the "AddLinkedSource" and "AddPickColumn" functions. Use the information from the "<ExistingTables>" section to guide this creation.

#### Pick Mode Column
Pick Mode columns help generate diverse rows automatically. To ensure variety, you can add columns like cuisine or meal type to serve as context during AI generation. For example, if you're generating 200 unique recipes, instead of relying solely on context, you can randomize the values of these columns (e.g., "Chinese" and "Lunch") for each recipe. This approach increases diversity without requiring context from previous generations.

Examples:

**Using Predefined List**:

AddListSource("cuisines", ["Chinese", "Japan", "Italian", "Mexican"])

# here repeat is 3 because we want 3 recipes for each cuisine.
AddPickColumn("cuisine", "cuisine of the recipe", "string", true, 3, 0, "cuisines", "", [])

**Using AI to Generate Options**:

AddAiSource("cuisines","Generate 30 common recipe cuisines")

# here repeat is 3 because we want 3 recipes for each cuisine.
AddPickColumn("cuisine", "cuisine of the recipe", "string", true, 3, 0, "cuisines", "", [])


### Best Practices

1. **Pick Type and Linked Sources**: If the column type is "pick" and the source is a linked table, both "linked_column" and "linked_context_columns" must be specified. The "linked_column" is used for display, but if you also want it as context to guide the LLM, include it in "linked_context_columns".

2. **Context Length**: Avoid excessively long context lengths, as they consume too many tokens and may not perform well due to LLM limitations. For diverse results, it's often more effective to use pick-mode columns from a source rather than relying on a large context. Context length should remain moderate for consistency. For example, if each row represents a chapter in a story, use a context length of 1 or 2 to ensure consistency.

3. **linked_context_columns**: Linked context columns are used to bring additional context data from another table into the current table when generating a new row. This feature is only available for pick-type columns where the source type is "linked".

For example, imagine you have a "customer" table that includes columns like age, job, salary, and other relevant information. Now, assume there is a separate "workout_plan" table where each row represents a workout plan recommended to a customer. The "workout_plan" table would include a "user" column, which links to the "customers" table. When generating a workout plan, it is beneficial to include available customer data.

In this case, the column definition might look like this:

AddPickColumn("customers", "Customer for whom the plan is created", "string", true, 1, 0, "customers", "name", ["name", "age", "job", "salary"])`

type TablePolishBuilder struct {
	Builder
}

func NewTablePolishBuilder(prompt string, name, description, sourcesJSON string, columnsJSON string, tables []TableInfoSimple) *TablePolishBuilder {
	tb := &TablePolishBuilder{}
	tb.AddText(tablePolishPrompt)
	el := tb.NewXML("Name")
	el.CreateText(name)
	el = tb.NewXML("Description")
	el.CreateText(description)
	el = tb.NewXML("SourcesJSON")
	el.CreateText(sourcesJSON)
	el = tb.NewXML("ColumnsJSON")
	el.CreateText(columnsJSON)
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
	el = tb.NewXML("Requirement")
	el.CreateText(prompt)
	tb.AddText("Now update the table schema based on <Requirement>. Existing sources or columns of the table are in <SourcesJson> and <ColumnsJson>. You are expected to **decide on the appropriate columns to modify automatically**, without asking for further input.")
	return tb
}

const tablePolishPrompt = `
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
- **linked_context_columns**: The linked-table columns used to providing context to AI when generating data (e.g., user name, user age, job, nationality). One column can be linked_column and also in linked_context_columns(and this happen very ofter), because the display value is also useful context to guide AI.

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
- AddAiColumn(name, description, contextLength, type)
- AddPickColumn(name, description, type, random, repeat, contextLength, source, linkedColumn, linkedContextColumns)
- AddListSource(name, options)
- AddAiSource(name,prompt)
- AddLinkedSource(name, table)
- RemoveColumn(name)
- RemoveSource(name)

The "depends" array lists other tables that this table **relies on for meaningful data**. For each dependent table, you must create a pick-type column and a related source using the "AddLinkedSource" and "AddPickColumn" functions. Use the information from the "<ExistingTables>" section to guide this creation.

#### Pick Mode Column
Pick Mode columns help generate diverse rows automatically. To ensure variety, you can add columns like cuisine or meal type to serve as context during AI generation. For example, if you're generating 200 unique recipes, instead of relying solely on context, you can randomize the values of these columns (e.g., "Chinese" and "Lunch") for each recipe. This approach increases diversity without requiring context from previous generations.

Examples:

**Using Predefined List**:

AddListSource("cuisines", ["Chinese", "Japan", "Italian", "Mexican"])

# here repeat is 3 because we want 3 recipes for each cuisine.
AddPickColumn("cuisine", "cuisine of the recipe", "string", true, 3, 0, "cuisines", "", [])

**Using AI to Generate Options**:

AddAiSource("cuisines","Generate 30 common recipe cuisines")

# here repeat is 3 because we want 3 recipes for each cuisine.
AddPickColumn("cuisine", "cuisine of the recipe", "string", true, 3, 0, "cuisines", "", [])


### Best Practices

1. **Pick Type and Linked Sources**: If the column type is "pick" and the source is a linked table, both "linked_column" and "linked_context_columns" must be specified. The "linked_column" is used for display, but if you also want it as context to guide the LLM, include it in "linked_context_columns".

2. **Changing Column**: If you change the type or other parameters of a column, but the column name remains the same, you must first remove the column and then add it again with the new type. Example: If the "tags" column type changes from string to array:

   RemoveColumn("tags")
   AddAiColumn("tags", "xxx", "array")

3. **Context Length**: Avoid excessively long context lengths, as they consume too many tokens and may not perform well due to LLM limitations. For diverse results, it's often more effective to use pick-mode columns from a source rather than relying on a large context. Context length should remain moderate for consistency. For example, if each row represents a chapter in a story, use a context length of 1 or 2 to ensure consistency.

4. **linked_context_columns**: Linked context columns are used to bring additional context data from another table into the current table when generating a new row. This feature is only available for pick-type columns where the source type is "linked".

For example, imagine you have a "customer" table that includes columns like age, job, salary, and other relevant information. Now, assume there is a separate "workout_plan" table where each row represents a workout plan recommended to a customer. The "workout_plan" table would include a "user" column, which links to the "customers" table. When generating a workout plan, it is beneficial to include available customer data.

In this case, the column definition might look like this:

AddPickColumn("customers", "Customer for whom the plan is created", "string", true, 1, 0, "customers", "name", ["name", "age", "job", "salary"])`
