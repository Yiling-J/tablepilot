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
	  Specify the model used to generate rows. If not provided, the default model will be used.

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
  tablepilot update recipes.json -t=recipes
  ```
  - `-t, --table string`
	 Table ID or name to update. If not provided, the name field in the JSON file will be used.

  The update command compares the existing columns in the database with those in the JSON file and matches them by **column name**. Columns present in both the database and JSON file will be updated. Columns missing from the JSON file will be removed. And if the table already contains data: Removed columns will also have their data deleted, newly added columns will be initialized with their zero value.

- **delete**
  Delete a specified table.
  ```console
  tablepilot delete recipes
  ```

- **describe**
  Show details about the columns in a specified table.
  ```console
  tablepilot describe recipes
  ```
  - `-o, --output string`
	Specifies the output format. Possible values are 'table' or 'json'. Defaults to 'table'.

- **export**
  Export the table as a CSV file.
  ```console
  tablepilot export recipes
  ```
  - `-t, --to string`
	Specifies the output file path.

- **generate**
  Generate data for a specified table.
  ```console
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
  ```console
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

- **regenerate**
  Regenerate specified rows/columns for a table. For each row in `rows`, the provided `--columns` will be regenerated.
  ```console
  tablepilot regenerate recipes -b=5 --columns=ingredients --columns=tags --rows=5CQZnC --rows=rsClYt
  ```
    - <model/batch/temperature generate command flgs>

	- `--columns string`
	  Specifies the columns to autofill; existing values will be ignored and regenerated. This flag can be set multiple times to specify multiple columns (see example).

	- `--rows string`
	  Specifies the rows(ID) to be regenerated. This flag can be set multiple times to specify multiple rows (see example).

	- `-p, --prompt string`
	  prompt text send to LLM.

	- `-t, --temperature float`
	  The sampling temperature. Higher values will make the output more random. (default 0.6)

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

  **Important**: When importing into an existing table, if the table contains a pick-type column and the imported value for this column is not empty, Tablepilot will scan the entire source (all CSV/Parquet files, the entire database table, or loop through the Hugging Face Rows API) to find a matching value. This process may take a significant amount of time if your source is large.

- **show**
  Display the rows of a specified table.
  ```console
  tablepilot show recipes
  ```

- **truncate**
  Remove all data from a specified table.
  ```console
  tablepilot truncate recipes
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
  tablepilot show recipes --config custom_config.toml
  ```
- **-v, --verbose**
  Verbose output, this will show detailed debug info including LLM prompt/response (default: false).
  ```console
  tablepilot generate recipes -v
  ```
