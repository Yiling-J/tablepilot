**CSV Data Source:** [Kaggle - Pokémon Dataset](https://www.kaggle.com/datasets/mohitbansal31s/pokemon-dataset)

### Create Table Schema
```shell
tablepilot create examples/pokemons/pokemons.json
```

### Import CSV Data
```shell
tablepilot import examples/pokemons/pokemons.csv
```

### Autofill Missing Ecology Data
This command fills in missing data for the **Ecology** column using all other columns as context. It sets the total row count to **1,000** (the CSV contains **905** rows) and processes data in batches of **15**, meaning each LLM API call generates **15** rows. The temperature is set to **1.5** for increased variability.

```shell
tablepilot autofill pokemons -c=1000 -b=15 --columns=Ecology -t=1.5
```
