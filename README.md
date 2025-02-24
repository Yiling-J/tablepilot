# Tablepilot(WIP)

Tablepilot is a CLI tool designed to generate tables using AI.

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
  tablepilot generate recipes
  ```

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

### Flags:

- **--config string**  
  Path to the config file (default is `config.toml`).
  
  ```
  tablepilot show recipes --config custom_config.toml
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
