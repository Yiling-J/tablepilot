**CSV Data Source:** [IMDB Movies Dataset](https://www.kaggle.com/datasets/harshitshankhdhar/imdb-dataset-of-top-1000-movies-and-tv-shows)

### Import CSV as Dataset
```console
tablepilot dataset create examples/imdb_movie_haiku/imdb_top_1000.csv -n=movies
```

### Create Table Schema
```console
tablepilot create examples/imdb_movie_haiku/haiku.json
```

### Generate Haiku for 30 Random Movies
```console
tablepilot generate imdb_haiku -c=30 -b=10 -t=0.9 --saveto=examples/imdb_movie_haiku/haiku.csv
```
