# Tablepilot API Documentation

## Overview
Tablepilot provides a REST API for managing AI-generated tables. You can start server using `tablepilot serve`

## Base URL
```
http://127.0.0.1:8080/api/v1
```

---
## Endpoints

### 1. Create a Table
#### Endpoint
```
POST /tables
```
#### Request Body

See [examples directory](https://github.com/Yiling-J/tablepilot/tree/main/examples) for more examples.

```json
{
  "name": "recipes",
  "model": "m1",
  "description": "all recipes",
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
### 2. Generate Rows
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
### 3. Stream Row Generation
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
### 4. Fetch Table Rows
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
### 5. List All Tables
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
### 6. Delete a Table
#### Endpoint
```
DELETE /tables/{table_id or table_name}
```
#### Response
```
204 No Content
```

---
### 7. Truncate a Table
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
### 8. Describe a Table
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

