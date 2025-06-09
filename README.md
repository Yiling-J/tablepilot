<div align="center">
    <img alt="tablepilot" height="100px" src="icon.png">
</div>

# Tablepilot

Tablepilot is a simple yet powerful AI-native platform for tabular data generation.

### Key Features

* Easily Generate, Autofill, Regenerate Rows
* Available on CLI, Web UI, App
* Supports Vision, Image Generation, Image Editing
* Create diverse and creative content using customizable sources
* Create Workflows to automate repetitive content generation tasks

### Demos

#### Generate / Autofill / Regenerate Recipes

<div align="start">
    <img alt="tablepilot" width="950px" src="demo.gif">
</div>

---

#### Workflow

A workflow that extracts dishes from a menu image, adds an image column, autofills missing images, and exports the generated data as a CSV file

<div align="start">
    <img alt="tablepilot" width="950px" src="workflow.gif">
</div>

---

Tablepilot uses a declarative schema format to create tables. Check out the [examples folder](examples) for many interesting use cases. The syntax is simple and intuitive, you can easily understand how it works without reading the full documentation.

#### Capabilities and Model Requirements

| Mode                         | Description                                                         | Available            | Model Requirements                                                      |
|------------------------------|---------------------------------------------------------------------|----------------------|-------------------------------------------------------------------------|
| builder                      | Create tables interactively using natural language                  | CLI                  | OpenAI Chat Completion API with support for parallel function calls       |
| generate(text)                | Generate rows (text) for the table                                  | CLI, API, WebUI, App      | OpenAI Chat Completion API with support for Structured Output            |
| autofill(text)                | Autofill columns (text) for existing rows in the table              | CLI, API, WebUI, App      | OpenAI Chat Completion API with support for Structured Output            |
| generate(text + vision)     | Generate rows (text) for the table, with image context              | CLI, API, WebUI, App      | OpenAI Chat Completion API with support for Structured Output and Vision |
| autofill(text + vision)     | Autofill columns (text) for existing rows, with image context       | CLI, API, WebUI, App      | OpenAI Chat Completion API with support for Structured Output and Vision |
| generate(text + image generation/edit) | Generate rows (text or image) for the table, with image context | CLI, API, WebUI, App      | The provider type must be `gemini`, and only `gemini-2.0-flash-exp-image-generation/gemini-2.0-flash-exp` is currently supported                     |
| autofill(text + image generation/edit) | Autofill columns (text or image) for existing rows, with image context | CLI, API, WebUI, App | The provider type must be `gemini`, and only `gemini-2.0-flash-exp-image-generation/gemini-2.0-flash-exp` is currently supported                       |
| image to table     | Extract structured data from an image into a table              | CLI, API, WebUI, App      | OpenAI Chat Completion API with support for Structured Output and Vision |
| workflow                      | Automate multi-step content generation tasks                  | CLI, API, WebUI, App                  | Depend on the steps of the workflow       |

> OpenAI Chat Completion API refers to any API compatible with OpenAI, such as Gemini, vLLM, Ollama, and xAI.

#### Download Binary Release

Pre-built binaries for various operating systems are available on the [Releases](https://github.com/Yiling-J/tablepilot/releases) page.

* Files with the `tablepilot_cli` prefix are for command-line interface (CLI) use. These include the CLI itself, as well as the API and WebUI.
* To use the Tablepilot desktop app on macOS or Windows, download the file with the `tablepilot_app` prefix that matches your platform (`.dmg` for macOS, `.exe` or `.msi` for Windows).

#### Install with Go
Ensure that Go is installed on your system. Then run `go install github.com/Yiling-J/tablepilot@latest`. Only **CLI and API** are supported.

#### Install from Source
Ensure that Go is installed on your system. Then, clone the repository and run `make install`. After installation, the `tablepilot` command should be available for use. This includes **CLI and API**. To use the WebUI, you need to **build the frontend first, before running make install**. Ensure you have `pnpm`, `tsc` and `node` installed, then run `make build-ui`, Once built, you can start the server using `serve` command.

To build the Desktop App, you'll need everything required for the WebUI, plus Rust and Tauri. Once set up, run  `make tauri-dev`, this will build and launch the Tauri app in development mode.

#### CLI and API Documentation

Tablepilot provides a full set of CLI commands, including `builder`, `create`, `update`, `autofill` and many more. Most CLI commands have corresponding API endpoints, and most operations can also be performed through the WebUI or App. Use `tablepilot serve` command to start API server and WebUI.

- For a complete list of CLI commands, see [this doc](CLI.md).
- For all available API endpoints, see [this doc](API.md).
- For workflow schema and CLI, see [this doc](Workflow.md).

## Guide

If you're using Tablepilot in CLI, the first step is to prepare a TOML config file. Below is an example `config.toml` file using an SQLite3 database (`data.db`) and use `gemini-2.0-flash-001`. Make sure to replace the `key` field with your actual Gemini API key before saving the file as `config.toml`.

```toml
[database]
driver = "sqlite3"
dsn = "data.db?_pragma=foreign_keys(1)"

[[providers]]
name = "gemini"
type = "Gemini"
key = "your_api_key"

[[models]]
model = "gemini-2.0-flash-001"
provider = "gemini"
rpm = 20
```

> For more config details, check the [documentation](docs/config).


Now let's explore its four main components: Models, Tables, Datasets, Workflows.


### Models

Models are LLMs (e.g., GPT-4o, Gemini Flash 2.5, Claude Sonnet 4) grouped under providers like OpenAI or Gemini. Tablepilot is designed to work flexibly with multiple models and providers.

-   **Providers**: Define the AI service provider (e.g., Gemini, OpenAI).
-   **Models**: Specify the particular model to use (e.g., `gemini-2.0-flash-001`), its provider, and other parameters like RPM (requests per minute).

**Define models in config file**

The example above already shows how to declare providers and models in the config toml.

**Define models in UI**

If you're using the Tablepilot web or desktop UI, you can manage providers and models on the Models page. The concept is the same: create a provider, then add models under it.


---

### Tables

In Tablepilot, **tables** are the core units where AI generation or autofill takes place. Each table has a `name`, a `description`, and a list of `columns`.

* In the **UI**, you can dynamically create tables using the builder.
* In the **CLI**, you must first prepare a table schema JSON file:

```json
{
  "name": "{name of the table}",
  "description": "{description of the table}",
  "columns": [column objects]
}
```

The `name` and `description` fields are straightforward, but it's important to note that both are sent to the AI during data generation. You can include relevant context or background information in the `description` to help the model generate more accurate results.

The `columns` section is the core of the generation process.

**Columns:**

> See the [column config readme](docs/schema/column.md) for full details.

Each column has `name`, `description`, and `type`. You also specify how the column should be populated using the `fill_mode` property.

All column names and descriptions are sent to the AI model, regardless of whether the column is AI-generated. This helps the model better understand the overall table schema and generate more relevant content, so make your descriptions as clear and descriptive as possible.

A column's `fill_mode` can be:

-   **AI Generated**:
    *   The content for this column is generated by an AI model.
    *   The generation is based on the table schema, context from other columns (especially those marked as `linked_context_columns` from picked sources), and any previous values in this column (see `context_length`).

-   **Pick from a Source**:
    *   The content for this column is selected from a specified source. This is further defined by the `source_type` property:
        *   **`source_type: "select from table"`**: Pulls data from another existing Tablepilot table.
        *   **`source_type: "select from dataset"`**: Pulls data from an existing Dataset (e.g., an imported CSV or Parquet file).
        *   **`source_type: "select from options"`**: Picks from a predefined list of values.

If the column is filled from a source, you can define how values are selected: randomly, randomly with replacement(same value can be selected more than once), or sequentially. For tabular sources (table/CSV dataset), you must specify a `linked_column`, and you may optionally define `linked_context_columns`. You can control how often a value is reused using the `repeat` parameter. For example, if your recipe table has a `Tag` column with `repeat` set to 5, the "Vegan" tag will be used for 5 rows before moving on to the next value from the source.

You can also configure the **context length**, which determines how many previous values from the column are passed to the AI during generation. For example, if the "Name" column has a context length of 20, the last 20 names will be included when generating the next row.

Tablepilot has two modes: generate and autofill. Use generate mode when you want to create new rows from scratch. Use autofill mode when you already have a table with data, and you’ve added new columns that needs to be filled in. Before diving into these two modes, let’s go over a few basic concepts.

-   **Generate Mode**: Generate mode creates new rows for a table. You specify the table, the number of rows, and the batch size (how many rows per AI call). During generation, each column is populated according to its defined `fill_mode` – either by AI or by picking values from a specified source. If you expect large outputs per row (e.g., from AI-generated columns), use smaller batch sizes to avoid hitting the model’s maximum token limit. Larger batch sizes can improve consistency and token efficiency.

-   **Autofill Mode**: Fill in missing values in existing columns of a table. Specify the columns to autofill and the columns to use as context. No new rows are created.

---

### Datasets

A dataset is a structured collection of data—such as a CSV file of customers or a list of recipe cuisines.

The primary purpose of a dataset is to provide context when generating rows in a table. For example, when generating recipes, you can use a dataset to populate the cuisine column with real values instead of letting AI generate them freely. This ensures each recipe has a cuisine from the list (randomly or sequentially selected), which then guides the AI to generate other columns accordingly. This results in more diverse and controlled outputs.

Currently, two types of datasets are supported: CSV and List.

- CSV Dataset: Upload CSV files with a consistent header/schema. You can select one column as the fill value and others as context data.
- List Dataset: Manually input values or ask AI to help generate them.
When defining table columns, set the fill mode to "Select from dataset" and choose your dataset. For example, if you're generating a sales plan using a customer dataset, you might use the Name column as the fill value, and Age, Job, and Salary as context fields to help the AI generate more relevant data.

---

### Workflow
Workflows let you automate a series of tasks with ease. For example, you might want to:

1. Import images into a table
2. Add a new column
3. Autofill data
4. Generate content into a new table using the current table as a source
5. Export the final result as a CSV

Workflows also support variables, so users can input or select values at runtime—making your automation more flexible and reusable.

If you're using the **CLI**, see [CLI workflow reference](Workflow.md) for details. If you're using the **UI**, simply build your workflow step-by-step using the visual editor.

Currently, workflows support the following step types:

* `Create Table`
* `Import`
* `Create Column`
* `Delete Column`
* `Generate`
* `Autofill`
* `Export Table`
* `Delete Table`

Check out [workflow examples](examples/workflows) to see how to build powerful workflows in practice.
