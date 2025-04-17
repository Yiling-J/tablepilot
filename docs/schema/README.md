## Table Schema

A table schema JSON file contains four main fields: `name`, `description`, `sources`, and `columns`.

```json
{
  "name": "{name of the table}",
  "description": "{description of the table}",
  "sources": [source objects],
  "columns": [column objects]
}
```

- For the definition of a source object, see [source.md](./source.md).
- For the definition of a column object, see [column.md](./column.md).
