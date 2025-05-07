## Configuration

Tablepilot requires a TOML configuration file. The default config file is `config.toml`, but you can specify a custom config file using the `--config` flag.

> For experimental image generation and editing, see this [example](examples/recipes_with_image) for details.

The configuration consists of following sections:

### Common (Optional)

- **source_data_dir**: The root search dir for CSV source `paths` field and List source `file` field. Default "./".

```toml
[common]
source_data_dir = "./data"
```


### Database

- **driver**: Specifies the database driver (e.g., `"sqlite3"`).
- **dsn**: The data source name (DSN) for the database connection.

```toml
[database]
driver = "sqlite3"
dsn = "data.db?_pragma=foreign_keys(1)"
```

### Providers

You can define multiple providers, and different models can use different providers.

- **name**: The name of the provider. This name is referenced in the `models` section to select which provider the model uses.
- **type**: The provider type. Currently, `"openai"` and `"gemini"`(experimental and only usable for image generation) is supported. `openai` type should includes all OpenAI-compatible APIs.
- **key**: The API key used to authenticate requests.
- **base_url**: The base URL of the API.

```toml
[[providers]]
name = "gemini"
type = "openai"
key = "your_api_key"
base_url = "https://generativelanguage.googleapis.com/v1beta/openai/"

[[providers]]
name = "openai"
type = "openai"
key = "your_api_key"
base_url = "https://api.openai.com/v1/"
```

### Server (Optional)

This section configures the API server when running `tablepilot serve`.

- **address**: TCP network address. Used in `http.ListenAndServe`. Default `:8080`.

```toml
[server]
address = ":8088"
```

### Models

You can define multiple models and assign them to different providers or different generations.

- **model**: The name of the model as used in the LLM API (e.g., `"gemini-2.0-flash-001"`).
- **alias**: An alias for the model (e.g., `"gemini-pro"`). This allows you to upgrade the model without changing the alias in the table JSON schema, making it easier to manage. Optional.
- **provider**: The name of the provider to be used for this model (must match a name from the `providers` section).
- **default**: Set to `true` if this is the default model. Only one model can be set as `default`. If no model is marked as `default`, the first model in the list will be used. The default model is used when no specific model is provided in the table JSON schema or the `--model` flag. Optional.
- **max_tokens**: The maximum number of tokens that can be generated in the chat completion (default 6000). Optional.
- **rpm**: The rate limit for this model, specified in requests per minute. This is used to control the rate of API calls and enforce a model-specific rate limiter (default no limit). Optional.

**Important**: All models must support [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs).

```toml
[[models]]
model = "gemini-2.0-flash-001"
provider = "gemini"
rpm = 20

[[models]]
model = "gpt-4o"
provider = "openai"
rpm = 5
```

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
