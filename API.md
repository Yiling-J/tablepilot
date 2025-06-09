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
