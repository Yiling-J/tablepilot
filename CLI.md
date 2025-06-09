### CLI Commands:

- **builder**
  Start the Tablepilot builder to create tables interactively using natural language. During the process, Tablepilot will guide you through:

  - Asking what you want to build.
  - Creating a draft list.
  - Asking if you'd like to improve it.
  - Generating tables and columns one by one.

  The instructions are straightforward, just follow the guide. Note: The builder command only creates the table structure; it does not populate the table with data. You’ll need to use the `generate` command separately to fill in the rows.
  ```console
  tablepilot builder
  ```
  	- `-m, --model string`
	  specify the model used by builder. If not provided, the default model will be used.

	- `-t, --temperature float`
	  The sampling temperature. Higher values will make the output more random. (default 0.3)

- **create**
  Create tables from schema JSON files.
  ```console
  tablepilot create recipes.json
  ```
  **Important**: If you modify the JSON schema after creating a table, be sure to update or recreate the table first.

- **update**
  Update table from schema JSON file.
  ```console
  tablepilot update recipes.json -t=<table id or name>
  ```
  - `-t, --table string`
	 Table ID or name to update. If not provided, the name field in the JSON file will be used.

  The update command compares the existing columns in the database with those in the JSON file and matches them by **column name**. Columns present in both the database and JSON file will be updated. Columns missing from the JSON file will be removed. And if the table already contains data: Removed columns will also have their data deleted, newly added columns will be initialized with their zero value.

- **list**
  List all available tables.
  ```console
  tablepilot list
  ```

- **delete**
  Delete a specified table.
  ```console
  tablepilot delete <table id or name>
  ```

- **describe**
  Show details about the columns in a specified table.
  ```console
  tablepilot describe <table id or name>
  ```
  - `-o, --output string`
	Specifies the output format. Possible values are 'table' or 'json'. Defaults to 'table'.

- **export**
  Export the table as a CSV file.
  ```console
  tablepilot export <table id or name>
  ```
  - `-t, --to string`
	Specifies the output file path.

- **generate**
  Generate data for a specified table.
  ```console
  tablepilot generate <table id or name> -c=50 -b=10
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
  Autofill missing columns specified table.
  ```console
  tablepilot autofill <table id or name> -c=50 -b=10 --columns=ingredients --columns=tags --context_columns=name --context_columns=steps
  ```
	- `-c, --count int`
	  total number of rows to autofill, default all existing rows. (default 0)
	- `-b, --batch int`
	  number of rows to autofill in a batch. (default 10)
	- `-o, --offset int`
	  start offset. (default 0)
	- `-s, --saveto string`
	  specify a file to save output, instead of storing in the database.
	- `-t, --temperature float`
	  The sampling temperature. Higher values will make the output more random. (default 0.6)
	- `-m, --model string`
	  specify the model used to generate rows. If not provided, the default model will be used.
	- `--columns stringArray`
	  columns to be autofilled, existing value wil be ignore and force regenerate. This flag can be set multiple times to specify multiple columns (see example). (required)
	- `--context_columns stringArray`
	  columns that should be put in prompt as context, default to all other columns. This flag can be set multiple times to specify multiple columns (see example).
	- `-p, --prompt string`
	  optional prompt text send to LLM.

- **regenerate**
  Regenerate specified rows/columns for a table. For each row in `rows`, the provided `--columns` will be regenerated.
  ```console
  tablepilot regenerate <table id or name> -b=5 --columns=ingredients --columns=tags --rows=5CQZnC --rows=rsClYt
  ```
	- `-p, --prompt string`
	  prompt text send to LLM.
	- `-b, --batch int`
	  number of rows to regenerate in a batch. (default 10)
	- `-r, --rows stringArray`
	  Specifies the rows(ID) to be regenerated. This flag can be set multiple times to specify multiple rows (see example). (required)
	- `-t, --temperature float`
	  The sampling temperature. Higher values will make the output more random. (default 0.6)
	- `-m, --model string`
	  specify the model used to generate rows. If not provided, the default model will be used.
	- `-c, --columns stringArray`
	  Specifies the columns to autofill; existing values will be ignored and regenerated. This flag can be set multiple times to specify multiple columns (see example). (required)

- **import**
  Import a CSV or image file into a table. Supported image file formats are PNG and JPEG. When importing image, a new table will always been created.
  ```console
  tablepilot import users.csv
  ```

  	- `-t, --table string` [CSV only]
	 imports into an existing table or creates a new one if missing. Defaults to the file name if not set.
	 - If the table exists, Tablepilot matches columns by name and tries to convert data types automatically, if a column exists in table but not in csv file, the default empty value of the column type will be used. Errors occur if conversion fails.
	 - If the table doesn't exist, all columns are treated as strings.

	- `-m, --model string` [Image only]
	  specify the model used to generate rows. If not provided, the default model will be used.

	- `-p, --prompt string` [Image only]
	  prompt text send to LLM, optional.

	- `-n, --name string`
	  name of the new table, if to flag is not set. Optional and if not set, new table name will be file name + current timestamp

	- `--truncate`
	  remove all rows in the table first before importing

  **Important**: When importing into an existing table, if the table contains a pick-type column and the imported value for this column is not empty, Tablepilot will scan the entire source (all CSV/Parquet files, the entire database table, or loop through the Hugging Face Rows API) to find a matching value. This process may take a significant amount of time if your source is large.

- **show**
  Display the rows of a specified table.
  ```console
  tablepilot show <table id or name>
  ```

- **truncate**
  Remove all data from a specified table.
  ```console
  tablepilot truncate <table id or name>
  ```

- **workflow**
  Workflow subcommands.

  - **workflow list**
    List workflows.
    ```console
    tablepilot workflow list
    ```

  - **workflow create <file>**
    Create workflow from schema JSON files.
    ```console
    tablepilot workflow create <file>
    ```

  - **workflow run <workflow>**
    Run workflow of given id or name.
    ```console
    tablepilot workflow run <workflow>
    ```
    - `-t, --temperature float`
      The sampling temperature. Higher values will make the output more random. (default 0.6)
    - `-m, --model string`
      Specify the model used to generate rows. If not provided, the default model will be used.
    - `-i, --image_model string`
      Specify the image model used to generate rows. If not provided, the default model will be used.

  - **workflow delete <workflow>**
    Delete workflow of given id or name.
    ```console
    tablepilot workflow delete <workflow>
    ```

- **dataset**
  Manage datasets.

  - **dataset create --name <name> [--desc <description>] --type <type> [--path <path>...]**
    Create a new dataset.
    ```console
    tablepilot dataset create --name <name> --type <type>
    ```
    - `-n, --name string`
      Name of the dataset (required).
    - `-d, --desc string`
      Description of the dataset.
    - `-t, --type string`
      Type of the dataset ('list' or 'csv') (required).
    - `-p, --path stringArray`
      Dataset file paths, for csv type all files should have same schema, and for list type, the final options will be options in all files concate together.

  - **dataset get <dataset_id>**
    Get info of a dataset by id.
    ```console
    tablepilot dataset get <dataset_id>
    ```

  - **dataset list**
    List all available datasets.
    ```console
    tablepilot dataset list
    ```

  - **dataset update <dataset_id> [--name <new_name>] [--desc <description>] [--type <type>] [--file <file>...]**
    Update an existing dataset.
    ```console
    tablepilot dataset update <dataset_id>
    ```
    - `--name string`
      New name for the dataset.
    - `--desc string`
      New description for the dataset.
    - `--type string`
      New type for the dataset ('list' or 'csv').
    - `-f, --file stringArray`
      New data files.

  - **dataset delete <dataset_id>**
    Delete a dataset.
    ```console
    tablepilot dataset delete <dataset_id>
    ```

  - **dataset preview <dataset_id>**
    Preview data from a dataset (first 100 rows for CSV, all options for List).
    ```console
    tablepilot dataset preview <dataset_id>
    ```

- **serve**
  Start an API server. See [API.md](API.md) for available endpoints. If you installed Tablepilot from a binary release or built the frontend when installing from source, the WebUI will be accessible at the root URL, such as: http://127.0.0.1:8083/
  ```console
  tablepilot serve
  ```
  By default, the API server listens on `:8083`. You can customize the address by adding a `server` section to your TOML config:
  ```
  [server]
  address = ":9901"
  ```

### Common Flags:

- **--config string**
  Path to the config file (default is `config.toml`).

  ```console
  tablepilot show <table id or name> --config custom_config.toml
  ```
- **-v, --verbose**
  Verbose output, this will show detailed debug info including LLM prompt/response (default: false).
  ```console
  tablepilot generate <table id or name> -v
  ```
