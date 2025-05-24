# Tablepilot Workflows

Workflows in Tablepilot allow you to define and automate a series of data operations. This is useful for creating reproducible data processing pipelines, complex data generation tasks, or any multi-step procedure involving your tables.

## Running Workflows via CLI

The primary command for interacting with workflows is `tablepilot workflow`.

### Listing Workflows
To see all available workflows:
```bash
tablepilot workflow list
```

### Creating a Workflow
To create a new workflow, you define its structure in a JSON file and then use the `create` command:
```bash
tablepilot workflow create <path_to_your_workflow_definition.json>
```
For example:
```bash
tablepilot workflow create my_workflow.json
```

### Running a Workflow
To execute a defined workflow:
```bash
tablepilot workflow run <workflow_name_or_id> [flags]
```
Replace `<workflow_name_or_id>` with the actual name (as defined in its JSON file) or the ID of the workflow.

**Flags:**
*   `-t, --temperature <float>`: Specifies the sampling temperature for AI generation steps (default: 0.6). Higher values (e.g., 0.8) make the output more random, while lower values (e.g., 0.2) make it more focused and deterministic.
*   `-m, --model <string>`: Specifies the primary AI model to use for generation tasks within the workflow. If not provided, the default model configured in Tablepilot will be used.
*   `-i, --image_model <string>`: Specifies the AI model to use for image-related tasks (like importing data from images) within the workflow. If not provided, the default image model will be used.

Example:
```bash
tablepilot workflow run my_data_pipeline --temperature 0.7 --model "gemini-pro"
```

### Deleting a Workflow
To remove a workflow:
```bash
tablepilot workflow delete <workflow_name_or_id>
```
Example:
```bash
tablepilot workflow delete old_workflow
```

## Workflow JSON Schema Definition

A workflow is defined by a JSON object with the following top-level properties:

*   `name` (string, required): A unique name for your workflow.
*   `description` (string, optional): A brief description of what the workflow does.
*   `variables` (array of `WorkflowVariable` objects, optional): Defines variables that can be used throughout the workflow steps.
*   `steps` (array of `WorkflowStep` objects, required): Defines the sequence of operations to be performed.

### Variables (`variables` array)

Variables allow you to define parameters that can be reused or easily changed without modifying the core logic of your workflow steps. Each object in the `variables` array defines a single variable:

*   `name` (string, required): The name of the variable. You'll use this name to reference the variable in your workflow steps.
*   `type` (string, required): The data type of the variable. Supported types:
    *   `"string"`: A text string.
    *   `"number"`: A floating-point number.
    *   `"integer"`: A whole number.
    *   `"file"`: Represents a file path. When the workflow runs, it may prompt the user for a file if a default value isn't provided or if the file type is used in a step expecting a file input (e.g., `Import` step).
*   `default_value` (any, optional): A default value for the variable. This value will be used if no other value is provided at runtime. The type of `default_value` should match the specified `type`.
*   `options` (array, optional): A list of predefined options for the variable. This can be useful for "string" or "number" types to suggest or restrict possible values, often used by UI integrations.

**Using Variables in Steps:**
Variables can be referenced within the `payload` of workflow steps using Go's template syntax: `{{.your_variable_name}}`.

Consider this example:
```json
{
  "name": "recipe_generator",
  "variables": [
    {
      "name": "cuisine_type",
      "type": "string",
      "default_value": "Italian"
    },
    {
      "name": "num_recipes",
      "type": "integer",
      "default_value": 3
    }
  ],
  "steps": [
    {
      "type": "Generate",
      "payload": {
        "table": "recipes_table",
        "count": "{{.num_recipes}}"
      }
    }
  ]
}
```
In the `Generate` step, `recipes_table` is the explicit name of the target table. The `count` parameter uses the value of the `num_recipes` variable defined above.

### Steps (`steps` array)

The `steps` array defines the actual operations the workflow will perform, in sequence. Each object in the array is a `WorkflowStep`:

*   `type` (string, required): Specifies the type of operation for this step.
*   `payload` (object, required): An object containing parameters specific to the `type` of this step.

**Referencing Entities from Previous Steps:**
When a step depends on an entity (like a table) created or modified by a previous step, you must provide the explicit name or identifier of that entity in the payload. The workflow system does not automatically pipe outputs like table names using special templates. You need to ensure the names used in subsequent steps match the names defined or generated in prior steps.

For example, if step 1 creates a table named "customer_data" (defined in its schema file or request payload), a subsequent step that modifies this table must refer to it by the exact name "customer_data" in its `table` parameter.

### Step Types and their `payload`

Here are the available `WorkflowStepType` values and the expected structure for their `payload`:

1.  **`CreateTable`**
    Creates a new table. The name of the table created will be taken from the `name` field within the JSON file specified by `schema_file`, or from the `name` field within the `request.name` if using programmatic creation.
    *   `payload`:
        *   `schema_file` (string): Path to a table schema JSON file (e.g., `examples/my_table_schema.json`). This file defines the columns, their types, and other table metadata, including the table name.
        *   `on_exists` (string, optional, defaults to "Stop"): Defines behavior if a table with the same name already exists.
            *   `"Stop"`: The workflow will stop with an error.
            *   `"Recreate"`: The existing table will be deleted, and a new one will be created.
            *   `"Skip"`: This step will be skipped if the table already exists.
        *   `request` (object, optional): For programmatic table creation without a `schema_file`. Contains detailed table generation parameters (structure defined by `table.TableGenRequest` in the backend). This is an advanced option. The table name would be specified within this request object.

2.  **`Import`**
    Imports data into a table from a file (e.g., CSV, image).
    *   `payload`:
        *   `table` (string, optional): The name or ID of an existing table to import data into. If not provided, a new table will be created (see `name`).
        *   `name` (string, optional): The name for the new table if `table` is not specified or refers to a non-existent table. If `name` is also not set, the new table name might be derived from the filename.
        *   `file` (string, required): Path to the file to import. Can be a direct path or use a variable: `{{.my_import_file}}`.
        *   `prompt` (string, optional): An AI prompt used when importing data from images, to guide the extraction process.
        *   `truncate` (boolean, optional, defaults to `false`): If `true`, all existing rows in the target table will be deleted before new data is imported.

3.  **`CreateColumn`**
    Adds a new column to an existing table.
    *   `payload`:
        *   `table` (string, required): The name or ID of the table to add the column to (e.g., `"my_target_table"`).
        *   `column` (object, required): Defines the new column:
            *   `name` (string, required): Name of the new column.
            *   `type` (string, required): Data type of the new column (e.g., "string", "number", "integer", "image", "boolean").
            *   `description` (string, optional): A description for the new column.

4.  **`DeleteColumn`**
    Removes a column from a table.
    *   `payload`:
        *   `table` (string, required): The name or ID of the table.
        *   `column` (string, required): The name of the column to delete.

5.  **`Generate`**
    Generates rows of data for a table using AI.
    *   `payload`:
        *   `table` (string, required): The name or ID of the table to generate data for (e.g., `"products_table"`).
        *   `count` (integer, required): The total number of rows to generate. Can use variables: `{{.num_rows}}`.
        *   `batch` (integer, optional, defaults to 10): The number of rows to generate in a single batch request to the AI.

6.  **`Autofill`**
    Autofills missing values in specified columns of a table using AI, based on existing data in other columns.
    *   `payload`:
        *   `table` (string, required): The name or ID of the table.
        *   `columns` (array of strings, required): A list of column names that need to be autofilled.
        *   `count` (integer, optional): The total number of rows to process for autofilling. If 0 or not provided, all existing rows (or a default limit) might be processed.
        *   `batch` (integer, optional, defaults to 10): The number of rows to process in a single batch.

7.  **`ExportTable`**
    Exports a table's data to a file (typically CSV).
    *   `payload`:
        *   `table` (string, required): The name or ID of the table to export (e.g., `"customer_orders"`).
        *   `path` (string, required): The file path where the data will be saved (e.g., `output/my_export.csv`). Can use variables: `{{.export_filepath}}`.

8.  **`DeleteTable`**
    Deletes a table.
    *   `payload`:
        *   `table` (string, required): The name or ID of the table to delete (e.g., `"temporary_data_table"`).

## Example Workflow JSON

Here is an example of a complete `workflow.json` file. This workflow first creates a table for fruit recipes, then adds an "Image" column to this table, and finally generates two recipes for it. It assumes that the `schema_file` (`recipes_v2.json`) defines the table name, potentially using variables, and that the `target_table_name` variable in the workflow is set to match this resulting table name.

```json
{
  "name": "fruit_recipe_generator_v2",
  "description": "Creates a fruit recipe table, adds an image column, and generates a few recipes. Uses explicit table name.",
  "variables": [
    {
      "name": "fruit_type_for_schema",
      "type": "string",
      "default_value": "generic_fruit"
    },
    {
      "name": "recipes_to_generate",
      "type": "integer",
      "default_value": 2
    },
    {
      "name": "target_table_name",
      "type": "string",
      "default_value": "FruitRecipes_From_generic_fruit"
    }
  ],
  "steps": [
    {
      "type": "CreateTable",
      "payload": {
        "on_exists": "Recreate",
        "schema_file": "examples/workflows/fruit_recipes/recipes_v2.json"
      }
    },
    {
      "type": "CreateColumn",
      "payload": {
        "table": "{{.target_table_name}}",
        "column": {
          "name": "Image",
          "type": "image",
          "description": "Generated image of the recipe based on its name and ingredients."
        }
      }
    },
    {
      "type": "Generate",
      "payload": {
        "table": "{{.target_table_name}}",
        "count": "{{.recipes_to_generate}}",
        "batch": 1
      }
    },
    {
      "type": "ExportTable",
      "payload": {
          "table": "{{.target_table_name}}",
          "path": "output/{{.fruit_type_for_schema}}_recipes_export.csv"
      }
    }
  ]
}
```

**Note on `examples/workflows/fruit_recipes/recipes_v2.json` (referenced in `CreateTable` step):**

This path is relative to the root of the Tablepilot application when the CLI is run. For the `CreateTable` step, you would need a JSON file at that path defining the table schema. The `name` field within this JSON schema file determines the name of the table created. Variables (like `{{.fruit_type_for_schema}}`) can be used in this schema file, and their values will be substituted from the workflow's variables.

For instance, if `recipes_v2.json` looks like this:
```json
{
  "name": "FruitRecipes_From_{{.fruit_type_for_schema}}",
  "description": "A table to store delicious fruit-based recipes.",
  "columns": [
    {
      "name": "RecipeName",
      "type": "string",
      "description": "The name of the fruit recipe."
    },
    {
      "name": "PrimaryFruit",
      "type": "string",
      "description": "The main fruit used in the recipe. Default: {{.fruit_type_for_schema}}"
    },
    {
      "name": "Ingredients",
      "type": "list_string",
      "description": "A list of ingredients for the recipe."
    },
    {
      "name": "Instructions",
      "type": "string",
      "description": "Step-by-step instructions to prepare the recipe."
    }
  ]
}
```
And if the workflow's `fruit_type_for_schema` variable is "apple", the table created by the `CreateTable` step will be named "FruitRecipes_From_apple". The `target_table_name` variable in the workflow should then be set to "FruitRecipes_From_apple" for subsequent steps to correctly reference this table.

This `Workflow.md` should provide a comprehensive guide to understanding and using workflows in Tablepilot.
