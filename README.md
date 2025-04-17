<div align="center">
    <img alt="tablepilot" height="100px" src="icon.png">
</div>

# Tablepilot

Tablepilot is a CLI/API/WebUI tool for generating and autofilling tables using AI. One of the most powerful features of Tablepilot is its ability to incorporate external context: such as other tables, [local CSV/Parquet files, or datasets from Kaggle or Hugging Face](contribute/csv-and-parquet.md). Making it easy to generate diverse results.

As a CLI tool, Tablepilot uses a declarative schema format. Check out the [examples folder](examples) for many interesting use cases. The syntax is simple and intuitive, you can easily understand how it works without reading the full documentation. A WebUI is also available. See the demo below:

![Demo](./demo.gif)

Tablepilot also includes experimental support for image understanding, image generation, and image editing. See this [detailed example](examples/recipes_with_image) for more information.

#### Concept

The concept behind Tablepilot is simple yet powerful. Suppose you want to generate 1,000 unique recipes using AI. A straightforward approach might be to ask ChatGPT for 10 recipes at a time, then continue requesting more while using previously generated content as context, until you reach 1,000 recipes. However, this method has two major drawbacks: the growing context consumes a large number of tokens, and as the context expands, ChatGPT struggles to ensure uniqueness across recipes.

Instead of relying on context, suppose we add two columns to the table: cuisine and meal type, and assign random values to them (e.g., Chinese and Lunch) for each of the 1,000 recipes. These values then serve as context for generation, naturally increases diversity in the results without needing previous generations as context. The key question is: How do we get random values for columns like cuisine and meal type? This is where Tablepilot excels. You can source data from other tables, local CSV or Parquet files, AI-generated options, or remote datasets from Kaggle and Hugging Face.

#### Download Binary Release
Pre-built binaries for different operating systems are available on the [Releases](https://github.com/Yiling-J/tablepilot/releases) page. The binary includes everything - CLI/API/WebUI, so you can start using Tablepilot instantly.

#### Install with Go
Ensure that Go is installed on your system. Then run `go install github.com/Yiling-J/tablepilot@latest`. Only CLI/API are supported.

#### Install from Source
Ensure that Go is installed on your system. Then, clone the repository and run `make install`. After installation, the `tablepilot` command should be available for use.  This includes CLI, API, and WebUI. However, to use the WebUI, you need to build the frontend first. Ensure you have `pnpm`, `tsc` and `node` installed, then run `make build-ui`, Once built, you can start the server using `serve` command.

## Guide

To start, first you need to prepare a toml config file. Below is an example `config.toml` file using an SQLite3 database (`data.db`) and use `gemini-2.0-flash-001`(openai compatible API mode). Make sure to replace the `key` field with your actual Gemini API key before saving the file as `config.toml`.

```toml
[database]
driver = "sqlite3"
dsn = "data.db?_pragma=foreign_keys(1)"

[[clients]]
name = "gemini"
type = "openai"
key = "your_api_key"
base_url = "https://generativelanguage.googleapis.com/v1beta/openai/"

[[models]]
model = "gemini-2.0-flash-001"
client = "gemini"
rpm = 20
```

> For more config details, check the [documentation](docs/config).

Tablepilot has two modes: generate and autofill. Use generate mode when you want to create new rows from scratch. Use autofill mode when you already have a table with data, and you’ve added new columns that needs to be filled in. Before diving into these two modes, let’s go over a few basic concepts.

A table typically consists of multiple columns. Some columns are AI-generated, while others are contextual—sourced from external data like another table, a fixed list, a CSV file, or a remote dataset.

In Tablepilot, tables are defined using a JSON schema. You pass this JSON file via CLI, send it as a payload via API, or create it through the WebUI (which auto-generates the JSON under the hood). The schema format looks like this:
```json
{
  "name": "{name of the table}",
  "description": "{description of the table}",
  "sources": [source objects],
  "columns": [column objects]
}
```
Let’s briefly review what each part means. Full definitions are available [here](docs/schema).

### Sources

> See the [source config readme](docs/schema/source.md) for full details.

When generating a row, non-AI columns are filled first using data from **source**. These values are then passed as context to the LLM to generate the remaining columns. Pretty straightforward, right?

Sources are the core of Tablepilot's flexibility and support 6 types:

- `list`
- `ai`
- `files`
- `linked`
- `csv`
- `parquet`

The first three act like lists of options. For example, a fruits list `["apple", "banana", "orange"]` and each row picks one value from it.

- **list**: Options are predefined by you.
- **ai**: If you don’t know the options, you can prompt the LLM to generate them.
- **files**: Used for image columns—values are file paths like `["cat.png", "dog.png"]`.

The remaining three are **tabular sources** containing multiple columns. For example, a `user` table might have Name, Age, and Job. You can select which column to use as the display value and which ones to use as context. See the next section for details.

- **linked**: Pulls data from other Tablepilot tables.
- **csv**: Uses local CSV files or remote Kaggle datasets.
- **parquet**: Uses local Parquet files or remote Hugging Face datasets.

---

### Columns

> See the [column config readme](docs/schema/column.md) for full details.

Each column has `name`, `description`, and `type`. You’ll also need to specify how the column should be populated, either by AI or from a source.

All column names and descriptions are sent to the AI model, regardless of whether the column is AI-generated. This helps the model better understand the overall table schema and generate more relevant content, so make your descriptions as clear and descriptive as possible.

If the column is filled from a source, you can define how values are selected: randomly, randomly with replacement(same value can be selected more than once), or sequentially. For tabular sources (linked/CSV/Parquet), you must specify a `linked_column`, and you may optionally define `linked_context_columns`. You can control how often a value is reused using the `repeat` parameter. For example, if your recipe table has a `Tag` column with `repeat` set to 5, the "Vegan" tag will be used for 5 rows before moving on to the next value from the source.

You can also configure the **context length**, which determines how many previous values from the column are passed to the AI during generation. For example, if the "Name" column has a context length of 20, the last 20 names will be included when generating the next row.

---

### Generate Mode

Generate mode is simple: specify which table to generate, how many rows, and the batch size (how many rows per AI call). If your rows are large, use smaller batch sizes to avoid token limits. Larger batches offer better consistency and token efficiency.

Example: generate 30 recipes, 5 per batch:

```sh
tablepilot generate recipes -c=30 -b=5
```

---

### Autofill Mode

Autofill mode fills in missing columns for an existing table. No new rows are created.

You specify:

- Which **columns** to autofill
- Which **columns** to use as context

Tablepilot then processes the table row by row.

Example: autofill `ingredients` and `steps` using `name` and `meal` as context:

```sh
tablepilot autofill recipes columns=ingredients columns=steps context_columns=name context_columns=meal -c=30 -b=5
```


## CLI and API

Tablepilot provides a full set of CLI commands, including `create`, `update`, `list`, `delete`, and many more. Most CLI commands have corresponding API endpoints, and most operations can also be performed through the WebUI.

- For a complete list of CLI commands, see [this doc](CLI.md).
- For all available API endpoints, see [this doc](API.md).
