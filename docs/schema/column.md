#### Column:
A list of column definitions. Each column is an object that can contain the following fields:

- **name**: The name of the column (e.g., `"Name"`, `"Ingredients"`). This will be used in the prompt when generating rows.
- **description**: A brief description of what data the column contains (e.g., `"recipe name"`). This will also be used in the prompt when generating rows.
- **type**: The data type for the column. Possible values include:
	- `"string"`: For text values.
	- `"array"`: For lists.
	- `"integer"`: For integral numbers.
	- `"number"`: For any numeric type, either integers or floating point numbers.
	- `"image"`: For local image files or image URLs. The value of this column should be either a file path relative to {source_data_dir} or a valid image URL when use as input. Your model must support [Images and vision](https://platform.openai.com/docs/guides/images?api-mode=chat) in order to use this column as input(image understanding), and you must specify the `image_models` in your config to generate or edit image.
- **fill_mode**: Specifies how the column is populated. Possible values:
	- `"ai"`: AI will generate values for this column.
	- `"pick"`: Values are picked from an existing source (e.g., a list of cuisines).
- **context_length** (Optional): Defines how many previous values in this column will be sent to the LLM when generating a new batch of rows. This helps provide context for the generation. If you aim for diverse results, using tag-like columns from the source may be more effective than increasing the context length. The context_length parameter is best used to ensure consistency in generation format and should typically remain moderate rather than excessively large.
- **source** (Optional): Specifies the source to pull data from when `fill_mode` is set to `"pick"`. This should match a source name defined in the `sources` section (e.g., `"cuisines"`).

**Additional Fields for `pick` Mode**

When `fill_mode` is set to `"pick"`, the following fields are available:

- **random**: If `true`, a random value is selected for each row from all available options in the source. Default: `false`.
- **replacement**: Determines whether sampling is with or without replacement:
  - `true`: Items can be selected multiple times.
  - `false`: Once an item is selected, it cannot be chosen again.
  - Default: `false`.
- **repeat**: Specifies how many times a picked value is reused before switching to the next one. The minimum and default value is `1`, meaning each value is used once before moving to the next.

When `source` type is `linked` or `csv` or `parquet`, the following fields are available:

- **linked_column**: The linked-table column used for display text in the generated cell(e.g., user name).
- **linked_context_columns**: The linked-table columns providing context when generating data (e.g., user age, job, nationality).

**Shared AI Type Source Behavior**

If multiple columns use the same AI source but have different `random`, `replacement`, or `repeat` settings, the source is initialized only once. For example, if a `"tags"` source generates 20 tag options via AI and three columns reference it, the tag generation process runs once, and all three columns share the same selection pool.
