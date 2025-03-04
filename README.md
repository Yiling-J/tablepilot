<div align="center">
  <a href="https://ollama.com" />
    <img alt="ollama" height="150px" src="icon.png">
  </a>
</div>

# Tablepilot

Tablepilot is a CLI tool designed to generate tables using AI.
- Reusable JSON table schema.
- Accurate column context length control and batch generation control.
- AI-powered columns generation.
- Bring other table row data as context to generate new data.
- Easy switching between LLM providers and models for flexibility.

#### Download Binary Release
Pre-built binaries for different operating systems are available on the [Releases](https://github.com/Yiling-J/tablepilot/releases) page.

#### Install with Go
Ensure that Go is installed on your system. Then run `go install github.com/Yiling-J/tablepilot@latest`.

#### Install from Source
Ensure that Go is installed on your system. Then, clone the repository and run `make install`. After installation, the `tablepilot` command should be available for use.

## How to Use

To generate a table, you need to prepare a TOML config file and a table schema JSON file. The config file defines the LLM clients used to generate tables, as well as the database where the table schema and data will be stored. The JSON schema file includes the table name, columns, and other information about your table.

Below is an example TOML config for using OpenAI GPT-4o and an SQLite3 database. Replace the `key` field with your OpenAI API key and save the file as `config.toml`:
```toml
[database]
driver = "sqlite3"
dsn = "data.db?_pragma=foreign_keys(1)"

[[clients]]
name = "openai"
type = "openai"
key = "your_api_key"
base_url = "https://api.openai.com/v1/"

[[models]]
model = "gpt-4o"
client = "openai"
rpm = 10
```
For this example, we'll use the schema file located at `examples/recipes_simple/recipes.json`.

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


## Examples

A number of examples demonstrating various use cases of Tablepilot are available in the [examples directory](https://github.com/Yiling-J/tablepilot/tree/main/examples).

## Usage

### CLI Commands:

- **create**
  Create tables from schema JSON files.
  ```
  tablepilot create recipes.json
  ```

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

- **import**
  Import a CSV file as a table.
  ```
  tablepilot import users.csv
  ```

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

Tablepilot uses a TOML configuration file to customize its behavior. The default config file is `config.toml`, but you can specify a custom config file using the `--config` flag.

The configuration consists of three main sections: `database`, `clients`, and `models`.

### Database

- **driver**: Specifies the database driver (e.g., `"sqlite3"`).
- **dsn**: The data source name (DSN) for the database connection.

### Clients

You can define multiple clients, and different models can use different clients.

- **name**: The name of the client. This name is referenced in the `models` section to select which client the model uses.
- **type**: The client type. Currently, only `"openai"` is supported, which should includes all OpenAI-compatible APIs.
- **key**: The API key used to authenticate requests.
- **base_url**: The base URL of the API.

### Models

You can define multiple models and assign them to different clients or different generations.

- **model**: The name of the model as used in the LLM API (e.g., `"gemini-2.0-flash-001"`).
- **alias**: An alias for the model (e.g., `"gemini-pro"`). This allows you to upgrade the model without changing the alias in the table JSON schema, making it easier to manage.
- **client**: The name of the client to be used for this model (must match a name from the `clients` section).
- **default**: Set to `true` if this is the default model. Only one model can be set as `default`. If no model is marked as `default`, the first model in the list will be used. The default model is used when no specific model is provided in the table JSON schema or the `--model` flag.
- **rpm**: The rate limit for this model, specified in requests per minute. This is used to control the rate of API calls and enforce a model-specific rate limiter.

**Important**: All models must support [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs).

## Table Schema

A Table schema JSON file consists of five main parts: `name`, `description`, `model`, `sources`, and `columns`.

### Schema Breakdown

#### name:

The name of the table. This serves as a unique identifier for the table (e.g., `"recipes"` in the example above).

#### description: 
A description of what the table represents. It provides context for the data (e.g., `"table of recipes"`). This description will be used in the prompt, so it should be clear and easy for the LLM to understand. It's helpful to include relevant details to ensure accurate and meaningful generation.

#### model (Optional): 
This section allows you to specify a default model for AI-generated columns. If not defined, the default model will be selected based on the configuration file.

#### sources: 
A list of sources from which `pick`-type columns can select values. Each source is an object with the following fields:

Common fields:

- **name**: The name of the source (e.g., `"cuisines"`).
- **type**: The type of the source, which can be `"ai"`, `"list"`, or `"linked"`.
- **random**: When set to `true`, each row generation will pick a random value from all available values in the source.
- **replacement**: Defines whether the sampling is with or without replacement. When set to `true`, items can be selected multiple times; when set to `false`, once an item is selected, it cannot be chosen again.

Special fields for different types:

- **ai**:
	- **prompt**: The prompt used to generate data from the AI model.
- **list**:
	- **options**: A list of predefined options to pick from.
- **linked**:
	- **table**: The name of the linked table.
	- **column**: The column used for display text in the generated cell(e.g., user name).
	- **context_columns**: The columns providing context when generating data (e.g., user age, job, nationality).

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
- **context_length** (Optional): Defines how many previous values in this column will be sent to the LLM when generating a new row. This helps provide context for the generation.
- **source** (Optional): Specifies the source to pull data from when `fill_mode` is set to `"pick"`. This should match a source name defined in the `sources` section (e.g., `"cuisines"`).
