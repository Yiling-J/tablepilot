**CSV Data Source:** [Kaggle - Pokémon Dataset](https://www.kaggle.com/datasets/mohitbansal31s/pokemon-dataset)

### Create Table Schema
The table consists of four columns:
- **Pokemon1**, **Pokemon2**, and **Pokemon3**: These columns store the three Pokémon featured in the story. Each is a *pick-type* column, meaning their values are randomly selected from the Pokémon CSV dataset.
- **Story**: This column is initially empty and will be generated based on the selected Pokémons.

To create the table, run:
```shell
tablepilot create examples/pokemons_autofill/pokemons.json
```

### Import CSV Data
This step imports a CSV file containing 30 rows into the table. Since the Pokémon columns are *pick-type*, additional context data, such as Pokémon type, classification, and abilities, will also be imported.

```shell
tablepilot import examples/pokemons_autofill/data.csv -t pokemon_stories
```

### Autofill Story
This command generates the missing **Story** column based on the Pokémon’s attributes.

```shell
tablepilot autofill pokemon_stories -c=30 -b=5 --columns=Story -t=1.2
```
