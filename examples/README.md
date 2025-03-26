## Examples

All examples are generated using gemini-2.0-flash-001 with a tempature of 0.6.

- **recipes_simple**
  A simple example that generates columns defined in the schema.

- **recipes_ai_columns**
  This example demonstrates the ability to automatically generate columns.

- **recipes_cuisine**
  This example shows how the AI generates 20 different cuisines and uses these generated values as the column values. Each generated row will randomly pick one cuisine from the 20 options.

- **recipes_cuisine_meal**
  Building on the `recipes_cuisine` example, this variation adds a "meal" column, where the value is selected from 6 meal options, creating more diversity in the results.
  
- **recipes_cuisine_meal_ingredients**
  Building on the `recipes_cuisine_meal` example, this variation adds two must-have-ingredient column, where the value is selected from 50 AI generated ingredients, creating even more diversity in the results. (But it doesn't seem very creative, maybe write a better prompt in table description or try a different model.)

- **recipes_by_country**
  This example demonstrates how to repeat column values. The generated table will include 6 recipes for each country: 2 Breakfast recipes, 2 Lunch recipes, and 2 Dinner recipes.

- **recipes_by_country_kaggle**
  This example demonstrates how to use a kaggle country dataset(https://www.kaggle.com/datasets/fernandol/countries-of-the-world) as context. The generated table will include 6 recipes for each country: 2 Breakfast recipes, 2 Lunch recipes, and 2 Dinner recipes.

- **recipes_for_customers**
  This is the most complex example, illustrating how to use another table as a reference. The `customers.json` file is used to generate a customer table, and then the recipes table is generated based on customer data. Each customer will receive a unique recipe tailored to their information.
  
- **pokémons**
  This example demonstrates how to create a table, import an existing CSV of 1000 Pokémons, and autofill column data. Tablepilot will generate ecological information for each Pokémon based on the existing row data.

- **imdb_movie_haiku**
  This example takes an IMDb movie CSV table and generates haiku poems inspired by movie titles and overviews, blending structured data with artistic expression.

- **chinese_qa_parquet**
  This example downloads Parquet files from a [Hugging Face dataset](https://huggingface.co/datasets/Congliu/Chinese-DeepSeek-R1-Distill-data-110k) and uses them as the data source.

- **chinese_qa_parquet_huggingface**
  Similar to the previous example, but this version uses Hugging Face integration directly, eliminating the need for manual downloads.
