# Tablepilot API Documentation

## Overview
Tablepilot provides a REST API for managing AI-generated tables. You can start server using `tablepilot serve`

## Base URL
```
http://127.0.0.1:8080/api/v1
```

---
## Endpoints

### Create a Table
#### Endpoint
```
POST /tables
```
#### Request Body

The JSON body of this API follows the same syntax as the CLI schema file. See [examples directory](https://github.com/Yiling-J/tablepilot/tree/main/examples) for more examples.

```json
{
  "name": "recipes",
  "model": "m1",
  "description": "all recipes",
  "columns": [
    {"name": "recipe_name", "description": "Name of the recipe", "type": "string", "fill_mode": "ai"}
  ]
}
```
#### Response
```json
{
  "id": "foo"
}
```

---
### Get Table Schema
#### Endpoint
```
GET /tables/{table_id or table_name}/schema
```
#### Description
Retrieves the schema definition of a specific table. This includes the table's name, description, and a detailed list of its columns with their configurations (name, type, fill mode, source details, etc.). This is useful for understanding the structure and data generation rules for a table.

#### Response
```json
{
  "name": "users",
  "description": "Table containing user information",
  "columns": [
    {
      "name": "user_id",
      "description": "Unique identifier for the user",
      "type": "integer",
      "fill_mode": "ai",
      "random": false,
      "replacement": false,
      "repeat": false,
      "context_length": 0,
      "linked_column": "",
      "linked_context_columns": [],
      "source_id": "",
      "source_type": "",
      "options": []
    },
    {
      "name": "username",
      "description": "User's chosen username",
      "type": "string",
      "fill_mode": "ai",
      "random": false,
      "replacement": false,
      "repeat": false,
      "context_length": 0,
      "linked_column": "",
      "linked_context_columns": [],
      "source_id": "",
      "source_type": "",
      "options": []
    },
    {
      "name": "email",
      "description": "User's email address",
      "type": "string",
      "fill_mode": "pick",
      "random": true,
      "replacement": false,
      "repeat": false,
      "context_length": 0,
      "linked_column": "email_address",
      "linked_context_columns": ["domain"],
      "source_id": "dataset_emails",
      "source_type": "dataset",
      "options": []
    }
  ]
}
```

---
### Update a Table
#### Endpoint
```
PATCH /tables/{table_id or table_name}
```
#### Request Body

```json
{
  "columns": [
    {"name": "recipe_name", "description": "Name of the recipe", "type": "string", "fill_mode": "ai"}
  ],
  "sources": [
    {"name": "cuisines"}
  ]
}
```
#### Response
```json
{
  "id": "foo"
}
```

---
### Generate Rows
#### Endpoint
```
POST /generate/tables/{table_id or table_name}
```
#### Request Body
```json
{
  "batch": 2,
  "count": 4,
  "temperature": 0.56,
  "model": "aiai"
}
```
#### Response
```json
{
  "data": [
    {"recipe_name": "0", "ingredient": "t0"},
    {"recipe_name": "1", "ingredient": "t1"}
  ]
}
```

---
### Autofill Rows
#### Endpoint
```
POST /autofill/tables/{table_id or table_name}
```
#### Request Body
```json
{
  "batch": 2,
  "count": 4,
  "temperature": 0.56,
  "model": "aiai"
  "autofill": {"columns": ["ingredients"], "context_columns": ["steps"]}
}
```
**Important**: Unlike the autofill CLI command, if `context_columns` is empty, other columns will not be used automatically. You must explicitly specify which `context_columns` to use.

#### Response
```json
{
  "data": [
    {"recipe_name": "0", "ingredient": "t0"},
    {"recipe_name": "1", "ingredient": "t1"}
  ]
}
```

---
### Regenerate Rows
#### Endpoint
```
POST /regenerate/tables/{table_id or table_name}
```
#### Request Body
```json
{
  "batch": 2,
  "temperature": 0.56,
  "model": "aiai"
  "autofill": {"columns": ["ingredients"], "rows": ["rsClYt", "8cR0I7"], "prompt": "foo bar"}
}
```

#### Response
```json
{
  "data": [
    {"recipe_name": "0", "ingredient": "t0"},
    {"recipe_name": "1", "ingredient": "t1"}
  ]
}
```

---
### Stream Row Generation
#### Endpoint
```
POST /generate/tables/{table_id or table_name}
```
#### Request Body
```json
{
  "batch": 2,
  "count": 4,
  "temperature": 0.56,
  "model": "aiai",
  "stream": true
}
```
#### Response (Event Stream)
```
event:message
data:{"data":[{"recipe_name":"0","ingredient":"t0"}]}

event:message
data:{"data":[{"recipe_name":"1","ingredient":"t1"}]}

event:message
data:[DONE]
```

---
### Fetch Table Rows
#### Endpoint
```
GET /tables/{table_id or table_name}/rows
```
#### Response
```json
{
  "data": [
    {"recipe_name": "Spaghetti Bolognese", "ingredient": "Tomato Sauce"},
    {"recipe_name": "Pancakes", "ingredient": "Flour"},
    {"recipe_name": "Salad", "ingredient": "Lettuce"}
  ],
  "total": 3
}
```

---
### List All Tables
#### Endpoint
```
GET /tables
```
#### Response
```json
{
  "total": 2,
  "tables": [
    {"id": "1", "name": "recipes", "description": "Collection of recipes"},
    {"id": "2", "name": "ingredients", "description": "Ingredient lists"}
  ]
}
```

---
### Delete a Table
#### Endpoint
```
DELETE /tables/{table_id or table_name}
```
#### Response
```
204 No Content
```

---
### Truncate a Table
#### Endpoint
```
POST /tables/{table_id or table_name}/truncate
```
#### Response
```json
{
  "deleted_rows": 5
}
```

---
### Describe a Table
#### Endpoint
```
GET /tables/{table_id or table_name}
```
#### Response
```json
{
  "columns": [
    {
      "ID": "1",
      "Name": "name",
      "Type": "string",
      "FillMode": "ai",
      "Description": "recipe name"
    },
    {
      "ID": "2",
      "Name": "description",
      "Type": "string",
      "FillMode": "ai",
      "Description": "recipe description"
    }
  ]
}
```

---
### List Models
#### Endpoint
```
GET /models
```
#### Response
```json
{
  "default": "4o"
  "models": ["gemini-2","4o","vllm-llama3"]
}
```

---
## Providers

Manages AI model providers.

### List All Providers
#### Endpoint
```
GET /providers
```
#### Description
Retrieves a list of all configured AI model providers.
#### Response
```json
[
  {
    "id": 1,
    "name": "openai_official",
    "type": "openaicompatible",
    "editable": false,
    "key": "sk-...",
    "base_url": "https://api.openai.com/v1",
    "enabled": true,
    "models": [
      {
        "model": "gpt-4",
        "alias": "gpt4",
        "client": "openai_official",
        "max_tokens": 8192,
        "rpm": 100,
        "image": true,
        "default": true
      }
    ]
  },
  {
    "id": 2,
    "name": "custom_ollama",
    "type": "openaicompatible",
    "editable": true,
    "key": "ollama_key_if_needed",
    "base_url": "http://localhost:11434/v1",
    "enabled": true,
    "models": [
      {
        "model": "llama3",
        "alias": "",
        "client": "custom_ollama",
        "max_tokens": 4096,
        "rpm": 0,
        "image": false,
        "default": false
      }
    ]
  }
]
```

---
### Create a Provider
#### Endpoint
```
POST /providers
```
#### Description
Creates a new AI model provider. The `id` and `editable` fields in the request body are typically ignored as they are server-assigned. `models` can be initially empty or provided.
#### Request Body
```json
{
  "name": "new_provider_xyz",
  "type": "openaicompatible",
  "key": "provider_api_key",
  "base_url": "https://api.example.com/v1",
  "enabled": true,
  "models": [
    {
      "model": "custom-model-1",
      "alias": "cm1",
      "max_tokens": 2048,
      "rpm": 50,
      "image": false
    }
  ]
}
```
#### Response
A `200 OK` response with an empty JSON string:
```json
""
```

---
### Update a Provider
#### Endpoint
```
PATCH /providers/{provider_id}
```
#### Description
Updates an existing AI model provider. The `id` in the URL specifies the provider to update. The request body should contain the fields to be updated. `name` and `type` are usually not changed, but other fields like `key`, `base_url`, `enabled`, and `models` can be updated.
#### Request Body
```json
{
  "key": "updated_api_key",
  "enabled": false,
  "models": [
    {
      "model": "custom-model-1",
      "alias": "cm1_updated",
      "max_tokens": 4000,
      "rpm": 60,
      "image": true
    },
    {
      "model": "new-model-2",
      "alias": "nm2",
      "max_tokens": 2000,
      "rpm": 30,
      "image": false
    }
  ]
}
```
#### Response
A `200 OK` response with an empty JSON string:
```json
""
```

---
### Delete a Provider
#### Endpoint
```
DELETE /providers/{provider_id}
```
#### Description
Deletes an AI model provider specified by its `id`.
#### Response
A `200 OK` response with an empty JSON string:
```json
""
```

---
## Workflows

Manages and executes automated workflows involving table operations and AI tasks.

### List All Workflows
#### Endpoint
```
GET /workflows
```
#### Description
Retrieves a list of all saved workflows.
#### Response
```json
{
  "total": 1,
  "workflows": [
    {
      "id": "wf_abc123",
      "name": "Daily Data Processing",
      "description": "Imports data, generates insights, and exports results."
    }
  ]
}
```

---
### Create a Workflow
#### Endpoint
```
POST /workflows
```
#### Description
Creates a new workflow.
#### Request Body (`workflow.Workflow`)
```json
{
  "name": "New User Onboarding Workflow",
  "description": "Prepares tables and initial data for new users.",
  "variables": [
    {
      "name": "username",
      "description": "The name of the new user.",
      "type": "string",
      "default_value": "guest"
    },
    {
      "name": "user_data_file",
      "description": "CSV file containing user details.",
      "type": "file"
    }
  ],
  "steps": [
    {
      "name": "Create User Table",
      "type": "create_table",
      "description": "Creates a table specific for the user {{.username}}.",
      "payload": {
        "request": {
          "name": "user_{{.username}}_details",
          "description": "Details for user {{.username}}",
          "columns": [
            {"name": "email", "type": "string", "fill_mode": "ai"},
            {"name": "registration_date", "type": "date", "fill_mode": "ai"}
          ]
        },
        "on_exists": "skip"
      }
    },
    {
      "name": "Import User Data",
      "type": "import",
      "description": "Imports data from the provided CSV.",
      "payload": {
        "table": "user_{{.username}}_details",
        "file": "{{.user_data_file}}",
        "truncate": true
      }
    }
  ]
}
```
**Note on `payload`**: The structure of `payload` within each step depends on the step's `type`. Variables defined in the `variables` section can be referenced in step payloads using Go template syntax (e.g., `{{.variable_name}}`).

#### Response
```json
{
  "id": "wf_xyz789"
}
```

---
### Get a Specific Workflow
#### Endpoint
```
GET /workflows/{workflow_id}
```
#### Description
Retrieves the complete details of a specific workflow by its ID.
#### Response (`workflow.Workflow`)
```json
{
  "id": "wf_xyz789",
  "name": "New User Onboarding Workflow",
  "description": "Prepares tables and initial data for new users.",
  "variables": [
    {
      "name": "username",
      "description": "The name of the new user.",
      "type": "string",
      "default_value": "guest"
    },
    {
      "name": "user_data_file",
      "description": "CSV file containing user details.",
      "type": "file"
    }
  ],
  "steps": [
    {
      "name": "Create User Table",
      "type": "create_table",
      "description": "Creates a table specific for the user {{.username}}.",
      "payload": {
        "request": {
          "name": "user_{{.username}}_details",
          "description": "Details for user {{.username}}",
          "columns": [
            {"name": "email", "type": "string", "fill_mode": "ai"},
            {"name": "registration_date", "type": "date", "fill_mode": "ai"}
          ]
        },
        "on_exists": "skip"
      }
    },
    {
      "name": "Import User Data",
      "type": "import",
      "description": "Imports data from the provided CSV.",
      "payload": {
        "table": "user_{{.username}}_details",
        "file": "{{.user_data_file}}",
        "truncate": true
      }
    }
  ]
}
```

---
### Update a Workflow
#### Endpoint
```
PATCH /workflows/{workflow_id}
```
#### Description
Updates an existing workflow. The request body should be a complete `workflow.Workflow` object containing the desired changes.
#### Request Body (`workflow.Workflow`)
(Similar to the Create Workflow request body, with potentially updated fields)
```json
{
  "id": "wf_xyz789", // Usually ignored by server, ID from URL is used
  "name": "Updated User Onboarding Workflow",
  "description": "Prepares tables and initial data for new users, now with more automation.",
  "variables": [
    {
      "name": "username",
      "description": "The name of the new user.",
      "type": "string",
      "default_value": "default_user"
    },
    {
      "name": "user_data_file",
      "description": "CSV file containing user details.",
      "type": "file"
    },
    {
      "name": "notification_email",
      "description": "Email to notify upon completion.",
      "type": "string"
    }
  ],
  "steps": [
    // Potentially updated or new steps
    {
      "name": "Create User Table",
      "type": "create_table",
      "description": "Creates a table specific for the user {{.username}}.",
      "payload": {
        "request": {
          "name": "user_{{.username}}_main_data", // Renamed table
          "description": "Main data for user {{.username}}",
          "columns": [
            {"name": "email", "type": "string", "fill_mode": "ai"},
            {"name": "registration_date", "type": "date", "fill_mode": "ai"},
            {"name": "status", "type": "string", "fill_mode": "ai"} // New column
          ]
        },
        "on_exists": "recreate"
      }
    }
  ]
}
```
#### Response
```json
{
  "id": "wf_xyz789"
}
```

---
### Delete a Workflow
#### Endpoint
```
DELETE /workflows/{workflow_id}
```
#### Description
Deletes a workflow specified by its ID.
#### Response
A `200 OK` response with an empty JSON string:
```json
""
```

---
### Run a Workflow
#### Endpoint
```
POST /workflows/{workflow_id_or_name}/run
```
#### Description
Executes a specified workflow. Input variables for the workflow are provided in the request body. The response is a Server-Sent Events (SSE) stream providing real-time updates on the workflow's execution.
#### Request Body (`workflow.StartWorklfowRequest`)
```json
{
  "variables": {
    "username": "john_doe",
    "user_data_file": {
        "name": "john_doe_details.csv",
        "data": "data:text/csv;base64,YXMNCg=="
    },
    "notification_email": "john.doe@example.com"
  },
  "model": "gpt-4-turbo",
  "image_model": "dall-e-3",
  "temperature": 0.7
}
```
**Note on `user_data_file`**: File variables are passed as objects with `name` and `data` (Data URL encoded string). Other variables are passed directly. `model`, `image_model`, and `temperature` are optional overrides for AI steps within the workflow.

#### Response (Event Stream)
The server responds with a stream of events. Each event has a `type` and `data`.
Example events:
```
event:message
data:{"type":"MESSAGE","data":"Starting step: Create User Table..."}

event:message
data:{"type":"STEP_DONE"}

event:message
data:{"type":"MESSAGE","data":"Starting step: Import User Data..."}

event:message
data:{"type":"ROWS","data":[{"email":{"value":"john.doe@example.com","type":"string"},"registration_date":{"value":"2023-10-26","type":"date"}}]}

event:message
data:{"type":"STEP_DONE"}

event:message
data:{"type":"EXPORT","data":"CSV content here...","path":"/path/to/export/file.csv"}
// The 'path' field in EXPORT is optional and indicates the desired save path if applicable.

event:message
data:{"type":"MESSAGE","data":"Table user_john_doe_main_data exported"}

event:message
data:{"type":"STEP_DONE"}

event:message
data:{"type":"ERROR","data":"Error message if something went wrong."}

event:message
data:{"type":"WORKFLOW_DONE"}

event:message
data:[DONE]
```
Possible event `type` values:
- `MESSAGE`: General message or log from a step.
- `ROWS`: Data generated by a step (e.g., from a "generate" or "autofill" step), typically an array of row objects.
- `EXPORT`: Contains exported data (e.g., CSV content) and an optional `path`.
- `ERROR`: Indicates an error occurred during a step.
- `STEP_DONE`: Signals the completion of a workflow step.
- `WORKFLOW_DONE`: Signals the completion of the entire workflow.
The stream is terminated by `[DONE]`.

---
## Datasets

Manages datasets that can be used as sources in table generation or other operations. Datasets can be of type 'list' (a simple list of strings) or 'csv' (data uploaded from a CSV file).

### List All Datasets
#### Endpoint
```
GET /datasets
```
#### Description
Retrieves a list of all available datasets.
#### Response
```json
{
  "total": 2,
  "datasets": [
    {
      "id": "ds_abc123",
      "name": "Country Names",
      "description": "A list of country names.",
      "type": "list",
      "column_count": 0,
      "value_count": 195,
      "data": ["United States", "Canada", "Mexico", "..."],
      "columns": []
    },
    {
      "id": "ds_xyz789",
      "name": "Product Catalog",
      "description": "CSV data for product information.",
      "type": "csv",
      "column_count": 3,
      "value_count": 0,
      "data": [],
      "columns": ["product_id", "product_name", "price"]
    }
  ]
}
```
**Note**: For 'list' type, `data` contains the list items and `column_count` is 0, `columns` is empty. For 'csv' type, `columns` contains the header names, `data` is empty, and `value_count` is 0 (as actual row count isn't stored directly in this summary).

---
### Create a Dataset
#### Endpoint
```
POST /datasets
```
#### Description
Creates a new dataset. The request must be `multipart/form-data`.
- For `type: "list"`, provide `name`, `description`, `type`, and `data` (as form fields, `data` can be repeated for multiple items).
- For `type: "csv"`, provide `name`, `description`, `type`, and one or more `files` (as file uploads).

#### Request Body (`multipart/form-data`)
**Example for `type: "list"`:**
- `name`: (string) "My List Dataset"
- `description`: (string) "A simple list of items"
- `type`: (string) "list"
- `data`: (string) "item1"
- `data`: (string) "item2"
  (repeat `data` field for each item in the list)

**Example for `type: "csv"`:**
- `name`: (string) "My CSV Dataset"
- `description`: (string) "Data uploaded from a CSV file"
- `type`: (string) "csv"
- `files`: (file) [upload content of data.csv]
  (can attach multiple files if the server merges them, or just one primary CSV)

#### Response
```json
{
  "id": "ds_new456",
  "name": "My List Dataset"
}
```

---
### Get a Specific Dataset
#### Endpoint
```
GET /datasets/{dataset_id}
```
#### Description
Retrieves detailed information about a specific dataset by its ID.
#### Response (`DatasetInfo`)
```json
{
  "id": "ds_abc123",
  "name": "Country Names",
  "description": "A list of country names.",
  "type": "list",
  "column_count": 0,
  "value_count": 195,
  "data": ["United States", "Canada", "Mexico", "Brazil", "United Kingdom", "..."],
  "columns": []
}
```
Or for a CSV dataset:
```json
{
  "id": "ds_xyz789",
  "name": "Product Catalog",
  "description": "CSV data for product information.",
  "type": "csv",
  "column_count": 3,
  "value_count": 0,
  "data": [],
  "columns": ["product_id", "product_name", "price"]
}
```

---
### Update a Dataset
#### Endpoint
```
PATCH /datasets/{dataset_id}
```
#### Description
Updates an existing dataset. The request must be `multipart/form-data`.
You can update `name`, `description`. For `type: "list"`, you can update `data`. For `type: "csv"`, you can upload new `files` (this typically replaces the existing CSV data). The dataset `type` cannot be changed.

#### Request Body (`multipart/form-data`)
**Example for updating description and data of a `list` dataset:**
- `description`: (string) "An updated list of items"
- `data`: (string) "updated_item1"
- `data`: (string) "updated_item2"
  (The provided `data` fields will replace the existing list.)

**Example for updating files of a `csv` dataset:**
- `files`: (file) [upload content of new_data.csv]
  (The new file(s) will replace the old CSV data.)

#### Response
```json
{
  "id": "ds_xyz789"
}
```

---
### Delete a Dataset
#### Endpoint
```
DELETE /datasets/{dataset_id}
```
#### Description
Deletes a dataset specified by its ID.
#### Response
A `200 OK` response with an empty JSON string:
```json
""
```

---
### Preview a Dataset
#### Endpoint
```
GET /datasets/{dataset_id}/preview
```
#### Description
Retrieves a preview of the dataset's content.
- For `type: "list"`, it returns all items.
- For `type: "csv"`, it typically returns the first 100 rows.
#### Response (`DatasetRows`)
**For a `list` dataset:**
```json
{
  "type": "list",
  "data": ["item1", "item2", "item3", "..."],
  "rows": null
}
```
**For a `csv` dataset:**
```json
{
  "type": "csv",
  "data": null,
  "rows": [
    {"product_id": "P1001", "product_name": "Laptop", "price": "1200.00"},
    {"product_id": "P1002", "product_name": "Mouse", "price": "25.00"},
    // ... up to 100 rows
  ]
}
```

---
### Import Image
#### Endpoint
```
POST /image_import/tables
```
#### Request Body
```json
{
  "data": <base64 encoded image data>,
  "model": "aiai",
  "prompt": "create a table if 4 columns, ...."
}
```

#### Response
```json
{"id": "cx5zty"}
```

---
### Generate List Options
#### Endpoint
```
POST /ai/list_gen
```
#### Description
Uses AI to generate a list of string options based on a given prompt. It can optionally take existing options to avoid duplicates and generate new, complementary options.
#### Request Body
```json
{
  "model": "gpt-4",
  "prompt": "Generate a list of common fruits.",
  "options": ["apple", "banana"]
}
```
If no specific AI model is provided, the system's default model will be used. The `options` field is optional; if provided, the AI will try to generate options not present in this list.

#### Response
A JSON array of generated string options.
```json
[
  "orange",
  "grape",
  "strawberry",
  "mango"
]
```