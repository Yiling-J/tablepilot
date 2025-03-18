**CSV Data Source:** [IMDB Movies Dataset](https://www.kaggle.com/datasets/harshitshankhdhar/imdb-dataset-of-top-1000-movies-and-tv-shows)

### Download Dataset
```shell
curl -L -o imdb1000.zip https://www.kaggle.com/api/v1/datasets/download/harshitshankhdhar/imdb-dataset-of-top-1000-movies-and-tv-shows
```

### Unzip
```shell
unzip imdb1000.zip -d movies
```

### Create Table Schema
```shell
tablepilot create examples/imdb_movie_haiku/haiku.json
```

### Generate Haiku for 30 Random Movies
```shell
generate imdb_haiku -c=30 -b=10 -t=0.9 --saveto=examples/imdb_movie_haiku/haiku.csv
```
