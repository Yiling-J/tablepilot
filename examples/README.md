## Examples

All examples are generated using gemini-2.0-flash-001 with a tempature of 0.6.

#### Text

- **recipes**
  A simple example that generates columns defined in the schema.

- **recipes_for_customers**
  This is a more complex example, illustrating how to use another table as a reference. The `customers.json` file is used to generate a customer table, and then the recipes table is generated based on customer data. Each customer will receive a unique recipe tailored to their information.

- **pokémons**
  This example demonstrates how to create a table, import an existing CSV of 1000 Pokémons as dataset, and autofill column data. Tablepilot will generate ecological information for each Pokémon based on the existing row data.

- **imdb_movie_haiku**
  This example takes an IMDb movie CSV table and generates haiku poems inspired by movie titles and overviews, blending structured data with artistic expression.

#### Image understanding

- **icon_jokes**
  This example create a joke from two random icon images, shows how to use images as input prompts(column type image). Include both generate and autofill examples.

#### Generate/Edit image

- **100_recipes**
  A comprehensive, step-by-step tutorial on how to generate 100 recipes with images.

- **outfit_preview**
  This example shows how to combine a hat image, a glasses image, and a model image to create an outfit preview.
