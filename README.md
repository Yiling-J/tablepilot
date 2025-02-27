# Tablepilot(WIP)

Tablepilot is a CLI tool designed to generate tables using AI.


## How to Use

To generate a table, you need to prepare a TOML config file and a table schema JSON file. The config file defines the LLM clients used to generate tables, as well as the database where the table schema and data will be stored. The JSON schema file includes the table name, columns, and other information about your table.

Once you have everything prepared, follow these steps:

1. **Save the Table Schema**: Use the `create` command to save the table schema into your database. After this step, the JSON file is no longer needed, as the schema is already stored in the database.
   
2. **Generate Rows**: Use the `generate` command to create rows. The rows will be stored automatically in the database. However, you can use the `saveas` flag to save the generated rows directly into a CSV file, instead of the database. In this case, the database acts as a schema store and does not store any row data.

3. **Export Data**: If you are storing data in the database, you can use the `export` command to export the data as a CSV file.

4. **Regenerate Data**: To regenerate data, use the `truncate` command to remove all rows from the table first. 

5. **Modify Schema**: If you modify the schema JSON, `delete` the existing table first and then recreate it using the updated JSON file.


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

#### Example Config

```toml
[database]
driver = "sqlite3"
dsn = "data.db?_pragma=foreign_keys(1)"

[[clients]]
name = "openai"
type = "openai"
key = ""
base_url = "https://models.inference.ai.azure.com"

[[models]]
model = "gpt-4o"
alias = "gpt4o"
client = "openai"
rpm = 10
```


## Table Schema
Tablepilot uses JSON files to define the schema for tables. A schema file specifies the structure of your table, including column names, data types, and any other relevant information. This allows you to create tables with a predefined structure.
