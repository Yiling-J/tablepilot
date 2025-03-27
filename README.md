<div align="center">
  <a href="https://ollama.com" />
    <img alt="ollama" height="100px" src="icon.png">
  </a>
</div>

# Tablepilot

Tablepilot is a CLI/API/WebUI tool designed to generate and autofill tables using AI in a scalable and diverse way.

- **Scalable:** Supports batch processing and allows fine-grained control over column context length to generate large datasets efficiently.
- **Diverse:** Easily integrates columns from other table or CSV/Parquet files as context, enabling richer and more dynamic table generation.
- **Easy use of remote data sources:** Supports Kaggle datasets and Hugging Face datasets, allowing seamless integration of real-world data. For more details about local/remote file sources, check out the [guide](contribute/csv-and-parquet.md).
- **Easily switch between different LLMs and models:** You can switch between models like OpenAI, Gemini or other OpenAI compatible API(Ollama/vLLM...) easily.
- **Seamless API & WebUI:** With a single command: `tablepilot serve`, you can launch both the API server and the WebUI effortlessly. Everything is included by default in the released binary. Check out the WebUI in action! This demo showcases both table generation and autofill:

![Demo](./demo.gif)

#### Download Binary Release
Pre-built binaries for different operating systems are available on the [Releases](https://github.com/Yiling-J/tablepilot/releases) page. The binary includes everything - CLI/API/WebUI, so you can start using Tablepilot instantly.

#### Install with Go
Ensure that Go is installed on your system. Then run `go install github.com/Yiling-J/tablepilot@latest`. Only CLI/API are supported.

#### Install from Source
Ensure that Go is installed on your system. Then, clone the repository and run `make install`. After installation, the `tablepilot` command should be available for use.  This includes CLI, API, and WebUI. However, to use the WebUI, you need to build the frontend first. Ensure you have `pnpm`, `tsc` and `node` installed, then run `make build-ui`, Once built, you can start the server using `serve` command.

## How to Use

To generate a table, you’ll need a **TOML config file**, and in the case of the CLI, a **table schema JSON file**:

- **Config File (Required):** Defines the LLM clients for table generation and specifies the database for storing table schemas and data.
- **Schema File (Only Required for CLI):** If using the CLI, you'll need a JSON schema file to define the table name, columns, and other details. For WebUI users, you can build the schema interactively in the UI. For API users, send the schema as JSON in the request body when calling the API.

Below is an example `config.toml` file using an SQLite3 database (`data.db`). This configuration defines two clients, **OpenAI** and **Gemini**, and assigns two models to them:  

- **Gemini** → `gemini-2.0-flash-001` (default)
- **OpenAI** → `gpt-4o`

You can modify the configuration by selecting a single client/model pair, removing the other, or adjusting the settings to fit your needs.

Make sure to replace the `key` field with your actual OpenAI/Gemini API key before saving the file as `config.toml`.

```toml
[database]
driver = "sqlite3"
dsn = "data.db?_pragma=foreign_keys(1)"

[[clients]]
name = "gemini"
type = "openai"
key = "your_api_key"
base_url = "https://generativelanguage.googleapis.com/v1beta/openai/"

[[clients]]
name = "openai"
type = "openai"
key = "your_api_key"
base_url = "https://api.openai.com/v1/"

[[models]]
model = "gemini-2.0-flash-001"
client = "gemini"
rpm = 20

[[models]]
model = "gpt-4o"
client = "openai"
rpm = 5
```
For CLI, we'll use the schema file located at `examples/recipes_simple/recipes.json`.

Once you have everything prepared, follow these steps:

**Save the Table Schema**: Use the `create` command to save the table schema into your database. After this step, the JSON file is no longer needed, as the schema is already stored in the database.
```
tablepilot create examples/recipes_simple/recipes.json
```

**View the Saved Schema**
```
tablepilot describe recipes
```
   
**Generate Rows**: Use the `generate` command to create rows. The rows will be stored automatically in the database. However, you can use the `saveas` flag to save the generated rows directly into a CSV file, instead of the database. In this case, the database acts as a schema store and does not store any row data. In this example we generate 20 recipes, 5 recipes a batch.
```
tablepilot generate recipes -c=20 -b=5
```

**Export Data**: If you are storing data in the database, you can use the `export` command to export the data as a CSV file.
```
tablepilot export recipes -t recipes.csv
```

For the API server, send a request with the schema as JSON in the request body. For the WebUI, you can build your schema interactively, see the demo above.

## Examples

A number of examples demonstrating various use cases of Tablepilot are available in the [examples directory](https://github.com/Yiling-J/tablepilot/tree/main/examples). Below are some interesting ones:

- **recipes_for_customers**
  This example illustrating how to use another table as a reference. The `customers.json` file is used to generate a customer table, and then the recipes table is generated based on customer data. Each customer will receive a unique recipe tailored to their information.
  
- **pokémons**
  This example demonstrates how to create a table, import an existing CSV of 1000 Pokémons, and autofill column data. Tablepilot will generate ecological information for each Pokémon based on the existing row data.
  
- **imdb_movie_haiku**
  This example takes an IMDb movie CSV table and generates haiku poems inspired by movie titles and overviews, blending structured data with artistic expression.

## Usage

### CLI Commands:

- **create**
  Create tables from schema JSON files.
  ```
  tablepilot create recipes.json
  ```
  **Important**: If you modify the JSON schema after creating a table, be sure to delete the existing table then recreating it.

- **delete**
  Delete a specified table.
  ```
  tablepilot delete recipes
  ```

- **describe**
  Show details about the columns in a specified table.
  ```
  tablepilot describe recipes
  ```

- **export**
  Export the table as a CSV file.
  ```
  tablepilot export recipes
  ```

- **generate**
  Generate data for a specified table.
  ```
  tablepilot generate recipes -c=50 -b=10
  ```

	- `-m, --model string`
	  Specify the model used to generate rows. If not provided, the default model will be used.

	- `-b, --batch int`
	  Number of rows to generate in a batch (default: 10).

	- `-c, --count int`
	  Total number of rows to generate.

	- `-s, --saveto string`
	  Specify a file to save the output, instead of storing it in the database.

	- `-t, --temperature float`
	  The sampling temperature. Higher values will make the output more random. (default 0.6)

- **autofill**
  Autofill specified columns for a table. For each existing row, the provided `--columns` will be generated.
  ```
  tablepilot autofill recipes -c=50 -b=10 --columns=ingredients --columns=tags --context_columns=name --context_columns=steps
  ```
    - <all generate command flgs>

	- `--columns string`
	  Specifies the columns to autofill; existing values will be ignored and regenerated. This flag can be set multiple times to specify multiple columns (see example).

	- `--context_columns string`
	  Specifies the columns to be used as context info when autofilling. This flag can be set multiple times to specify multiple columns (see example).
	  If you don't want any context columns, just set it to a non-existent column (`--context_columns=notexists`).

	- `-o, --offset int`
	  Start offset for autofilling rows. (default 0)

- **import**
  Import a CSV file into a table.
  ```
  tablepilot import users.csv
  ```

  	- `-t, --table string`
	 Imports into an existing table or creates a new one if missing. Defaults to the file name if not set.
	 - If the table exists, Tablepilot matches columns by name and tries to convert data types automatically, if a column exists in table but not in csv file, the default empty value of the column type will be used. Errors occur if conversion fails.
	 - If the table doesn't exist, all columns are treated as strings.

- **show**
  Display the rows of a specified table.
  ```
  tablepilot show recipes
  ```

- **truncate**
  Remove all data from a specified table.
  ```
  tablepilot truncate recipes
  ```
  
- **serve**
  Start an API server. See [API.md](API.md) for available endpoints. If you installed Tablepilot from a binary release or built the frontend when installing from source, the WebUI will be accessible at the root URL, such as: http://127.0.0.1:8080/
  ```
  tablepilot serve
  ```
  By default, the API server listens on `:8080`. You can customize the address by adding a `server` section to your TOML config:
  ```
  [server]
  address = ":9901"
  ```

### Common Flags:

- **--config string**
  Path to the config file (default is `config.toml`).
  
  ```
  tablepilot show recipes --config custom_config.toml
  ```
- **-v, --verbose**
  Verbose output, this will show detailed debug info including LLM prompt/response (default: false).
  ```
  tablepilot generate recipes -v
  ```

## Configuration

Tablepilot requires a TOML configuration file. The default config file is `config.toml`, but you can specify a custom config file using the `--config` flag.

The configuration consists of following sections:

### Common (Optional)

- **source_data_dir**: The root search dir for CSV source `paths` field and List source `file` field. Default "./".

### Database

- **driver**: Specifies the database driver (e.g., `"sqlite3"`).
- **dsn**: The data source name (DSN) for the database connection.

### Clients

You can define multiple clients, and different models can use different clients.

- **name**: The name of the client. This name is referenced in the `models` section to select which client the model uses.
- **type**: The client type. Currently, only `"openai"` is supported, which should includes all OpenAI-compatible APIs.
- **key**: The API key used to authenticate requests.
- **base_url**: The base URL of the API.

### Server (Optional)

This section configures the API server when running `tablepilot serve`.

- **address**: TCP network address. Used in `http.ListenAndServe`. Default `:8080`.

### Models

You can define multiple models and assign them to different clients or different generations.

- **model**: The name of the model as used in the LLM API (e.g., `"gemini-2.0-flash-001"`).
- **alias**: An alias for the model (e.g., `"gemini-pro"`). This allows you to upgrade the model without changing the alias in the table JSON schema, making it easier to manage. Optional.
- **client**: The name of the client to be used for this model (must match a name from the `clients` section).
- **default**: Set to `true` if this is the default model. Only one model can be set as `default`. If no model is marked as `default`, the first model in the list will be used. The default model is used when no specific model is provided in the table JSON schema or the `--model` flag. Optional.
- **max_tokens**: The maximum number of tokens that can be generated in the chat completion (default 6000). Optional.
- **rpm**: The rate limit for this model, specified in requests per minute. This is used to control the rate of API calls and enforce a model-specific rate limiter (default no limit). Optional.

**Important**: All models must support [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs).

### Sources (Optional)

You can also define shared sources here. These sources will be accessible to all tables. For more details on source definitions, see [Sources](#sources). Example:
```toml
[[sources]]
name = "customers"
type = "linked"
table = "customers"

[[sources]]
name = "movies"
type = "csv"
paths = ["movies/*.csv"]
```

## Table Schema

A Table schema JSON file consists of five main parts: `name`, `description`, `model`, `sources`, and `columns`.

### Schema Breakdown

#### name:

The name of the table. This serves as a unique identifier for the table (e.g., `"recipes"` in the example above). Use only letters, numbers, and underscores (_), and start with a letter.

#### description: 
A description of what the table represents. It provides context for the data (e.g., `"table of recipes"`). This description will be used in the prompt, so it should be clear and easy for the LLM to understand. It's helpful to include relevant details to ensure accurate and meaningful generation.

#### model (Optional): 
This section allows you to specify a default model for AI-generated columns. If not defined, the default model will be selected based on the configuration file.

#### sources:
A list of sources from which `pick`-type columns can select values. Tablepilot currently 4 types of sources:

- **AI**: Uses AI to generate a list of options dynamically. Each time a new generation starts (via the `generate` command, generate API call, or start button in UI), the options will be regenerated.
- **LIST**: Uses a predefined list of options.
- **LINKED**: Uses rows from another table as the source.
- **CSV**: Uses rows from one or more CSV files as the source. All CSV files must have a header row with column names and share the same column structure.

Each source is an object with the following fields:

Common fields:

- **name**: The name of the source (e.g., `"cuisines"`).
- **type**: The type of the source, which can be `"ai"`, `"list"`, `"linked"`, `"csv"` or `"parquet"`.

Special fields for different types:

- **ai**:
	- **prompt**: The prompt used to generate options by AI, e.g., Give me 50 common ingredients.
- **list**:
	- **options**: A list of predefined options to pick from.
	- **file**: Use file content as options, each line in the file will be one option, if this field is not empty then `options` field will be ignored. File path is **relative to the source_data_dir config**, e.g., `countries.text` or `data/countries.txt`.
- **linked**:
	- **table**: The name of the linked table.
- **csv**:
  - **paths**: A list of path patterns **relative to the source_data_dir config**. Supports exact matches, single asterisk (`*`), and double asterisk (`**`) patterns. Examples:
    - Full match: `"data/cuisines.csv"`
    - Single asterisk: `"data/*.csv"` (matches all CSV files in `data/`)
    - Double asterisk: `"data/**/*.csv"` (matches all CSV files in `data/` and subdirectories)
  - **kaggle**: The Kaggle dataset name, e.g., `"fernandol/countries-of-the-world"`.

    You can use a Kaggle dataset as a CSV data source. When doing so, Tablepilot first downloads the dataset into a cache folder:

    ```
    {source_data_dir}/tablepilot_kaggle_cache/{dataset_name (with `/` replaced by `--`)}
    ```

    Once downloaded, it functions like a local CSV source. The only difference is that the search root for `paths` is relative to the downloaded folder. The cached dataset will always be used unless you remove it manually.

    **Example:**
    ```json
    {
      "name": "countries",
      "type": "csv",
      "kaggle": "fernandol/countries-of-the-world",
      "paths": ["*.csv"]
    }
    ```

    You can find the dataset name by clicking the **Download** button in the top-right corner of the dataset page on Kaggle.

- **parquet**:
  - **paths**: Same as `CSV` source.
  - **huggingface**: If specified, uses a Hugging Face dataset and ignores `paths`.
    - **dataset**: The dataset to use (e.g., `facebook/natural_reasoning`).
    - **config**: (Optional) The dataset configuration to use, defaulting to `"default"`.
    - **split**: (Optional) The dataset split to use, defaulting to `"train"`.

#### columns: 
A list of column definitions. Each column is an object that can contain the following fields:

- **name**: The name of the column (e.g., `"Name"`, `"Ingredients"`). This will be used in the prompt when generating rows.
- **description**: A brief description of what data the column contains (e.g., `"recipe name"`). This will also be used in the prompt when generating rows.
- **type**: The data type for the column. Possible values include:
	- `"string"`: For text values.
	- `"array"`: For lists.
	- `"integer"`: For integral numbers.
	- `"number"`: For any numeric type, either integers or floating point numbers.
- **fill_mode**: Specifies how the column is populated. Possible values:
	- `"ai"`: AI will generate values for this column.
	- `"pick"`: Values are picked from an existing source (e.g., a list of cuisines).
- **context_length** (Optional): Defines how many previous values in this column will be sent to the LLM when generating a new batch of rows. This helps provide context for the generation. If you aim for diverse results, using tag-like columns from the source may be more effective than increasing the context length. The context_length parameter is best used to ensure consistency in generation format and should typically remain moderate rather than excessively large.
- **source** (Optional): Specifies the source to pull data from when `fill_mode` is set to `"pick"`. This should match a source name defined in the `sources` section (e.g., `"cuisines"`).

**Additional Fields for `pick` Mode**

When `fill_mode` is set to `"pick"`, the following fields are available:

- **random**: If `true`, a random value is selected for each row from all available options in the source. Default: `false`.
- **replacement**: Determines whether sampling is with or without replacement:
  - `true`: Items can be selected multiple times.
  - `false`: Once an item is selected, it cannot be chosen again.
  - Default: `false`.
- **repeat**: Specifies how many times a picked value is reused before switching to the next one. The minimum and default value is `1`, meaning each value is used once before moving to the next.

When `source` type is `linked` or `csv` or `parquet`, the following fields are available:

- **linked_column**: The linked-table column used for display text in the generated cell(e.g., user name).
- **linked_context_columns**: The linked-table columns providing context when generating data (e.g., user age, job, nationality).

**Shared AI Type Source Behavior**

If multiple columns use the same AI source but have different `random`, `replacement`, or `repeat` settings, the source is initialized only once. For example, if a `"tags"` source generates 20 tag options via AI and three columns reference it, the tag generation process runs once, and all three columns share the same selection pool.
