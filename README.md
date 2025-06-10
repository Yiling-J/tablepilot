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

#### Install with Docker (WebUI)

Create a directory called `tablepilot`. Make `tablepilot` your current working directory:

```console
mkdir tablepilot
cd tablepilot
```

Copy and paste the following commands into your command line to start the Docker container:

```console
docker run --name tablepilot -d -v $(pwd):/app/data -p 8083:8083 yilingj/tablepilot:latest
```

Verify that your containers are running:

```console
docker container ls
```

You should see something similar to the following:

```console
CONTAINER ID   IMAGE                       COMMAND                CREATED         STATUS         PORTS                    NAMES
acf9d361460c   yilingj/tablepilot:latest   "./tablepilot serve"   7 seconds ago   Up 7 seconds   0.0.0.0:8083->8083/tcp   tablepilot
```

Open your browser and go to http://127.0.0.1:8083 You should see the Tablepilot web interface, including a built-in example table of recipes.

#### Download Binary Release

Pre-built binaries for various operating systems are available on the [Releases](https://github.com/Yiling-J/tablepilot/releases) page.

* Files with the `tablepilot_cli` prefix are for command-line interface (CLI) use. These include the CLI itself, as well as the API and WebUI.
* To use the Tablepilot desktop app on macOS or Windows, download the file with the `tablepilot_app` prefix that matches your platform (`.dmg` for macOS, `.exe` or `.msi` for Windows).

#### Install with Go
Ensure that Go is installed on your system. Then run `go install github.com/Yiling-J/tablepilot@latest`. Only **CLI and API** are supported.

#### Install from Source
Ensure that Go is installed on your system. Then, clone the repository and run `make install`. After installation, the `tablepilot` command should be available for use. This includes **CLI and API**. To use the WebUI, you need to **build the frontend first, before running make install**. Ensure you have `pnpm`, `tsc` and `node` installed, then run `make build-ui`.

To build the Desktop App, you'll need everything required for the WebUI, plus Rust and Tauri. Once set up, run  `make tauri-dev`, this will build and launch the Tauri app in development mode.

#### Run Tablepilot WebUI

```console
tablepilot serve
```

The serve command starts the backend server and, if the frontend has already been built, it will also serve the Web UI. Once running, you can access the interface at:
http://localhost:8083

#### CLI and API Documentation

Tablepilot provides a full set of CLI commands, including `builder`, `create`, `update`, `autofill` and many more. Most CLI commands have corresponding API endpoints, and most operations can also be performed through the WebUI or App.

When using the API or CLI without the UI, you must first prepare a TOML config file specifying the model providers and models used by Tablepilot. Example:

```toml
[[providers]]
name = "gemini"
type = "Gemini"
key = "your_api_key"

[[models]]
model = "gemini-2.0-flash-001"
provider = "gemini"
rpm = 20
```

- For more config details, check the [documentation](docs/config).
- For a complete list of CLI commands, see [this doc](CLI.md).
- For all available API endpoints, see [this doc](API.md).
- For workflow schema and CLI, see [this doc](Workflow.md).

Check out the [examples folder](examples) for some CLI use cases. The syntax is simple and intuitive, you can easily understand how it works without reading the full documentation.

## Guide

Tablepilot has four core components: Models, Tables, Datasets, and Workflows.

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

**Columns**

> See the [column config readme](docs/schema/column.md) for full details.

Each column has `name`, `description`, and `type`. You also specify how the column should be populated using the `fill_mode` property.

All column names and descriptions are sent to the AI model, regardless of whether the column is AI-generated. This helps the model better understand the overall table schema and generate more relevant content, so make your descriptions as clear and descriptive as possible.

A column's `fill_mode` can be:

-   **AI Generated**: The content for this column is generated by an AI model.

-   **Pick from a Source**: The content for this column is selected from a specified source. This is further defined by the `source_type` property.
        *   **select from table**: Pulls data from another existing Tablepilot table.
        *   **select from dataset**: Pulls data from an existing Dataset.
        *   **select from options**: Picks from a predefined list of values.

If the column is filled from a source, you can define how values are selected: randomly, randomly with replacement(same value can be selected more than once), or sequentially. For tabular sources (table or csv-dataset), you must specify a `linked_column`, and you may optionally define `linked_context_columns`. You can control how often a value is reused using the `repeat` parameter. For example, if your recipe table has a `Tag` column with `repeat` set to 5, the "Vegan" tag will be used for 5 rows before moving on to the next value from the source.

You can also configure the **context length**, which determines how many previous values from the column are passed to the AI during generation. For example, if the "Name" column has a context length of 20, the last 20 names will be included when generating the next row.

**Generate and Autofill**

Tablepilot has two modes: generate and autofill. Use generate mode when you want to create new rows from scratch. Use autofill mode when you already have a table with data, and you’ve added new columns that needs to be filled in or regenerate values for columns.

-   **Generate Mode**: Generate mode creates new rows for a table. You specify the table, the number of rows, and the batch size (how many rows per AI call). During generation, each column is populated according to its defined `fill_mode` – either by AI or by picking values from a specified source. If you expect large outputs per row (e.g., from AI-generated columns), use smaller batch sizes to avoid hitting the model’s maximum token limit. Larger batch sizes can improve consistency and token efficiency.

-   **Autofill Mode**:  Fills in missing values for existing columns in a table. No new rows are created. When starting an autofill task, you must specify both the columns to be filled and the columns to be used as context. A column can be used as both a target and context, this means the value will be regenerated based on its existing content. You can also provide an optional prompt to help the AI better understand the desired behavior.

---

### Datasets

A dataset is a structured collection of data, such as a CSV file of customers or a list of recipe cuisines.

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
