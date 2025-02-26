## Examples

- **recipes_simple**
  A simple example that generates columns defined in the schema.

- **recipes_ai_columns**
  This example demonstrates the ability to automatically generate columns.

- **recipes_cuisine**
  This example shows how the AI generates 20 different cuisines and uses these generated values as the column values. Each generated row will randomly pick one cuisine from the 20 options.

- **recipes_cuisine_meal**
  Building on the `recipes_cuisine` example, this variation adds a "meal" column, where the value is selected from 6 meal options, creating more diversity in the results.

- **recipes_by_country**
  This example demonstrates how to repeat column values. The generated table will include 6 recipes for each country: 2 Breakfast recipes, 2 Lunch recipes, and 2 Dinner recipes.

- **recipes_for_customers**
  This is the most complex example, illustrating how to use another table as a reference. The `customers.json` file is used to generate a customer table, and then the recipes table is generated based on customer data. Each customer will receive a unique recipe tailored to their information.
