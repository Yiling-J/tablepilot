#### Generate Example
```
tablepilot create examples/icon_jokes/icon_jokes.json

tablepilot generate icon_jokes -c 15 -b 15
```

#### Autofill Example
```
tablepilot create examples/icon_jokes/icon_jokes_autofill.json

tablepilot import examples/icon_jokes/icon_jokes_autofill.csv

tablepilot autofill icon_jokes -c 15 -b 15 --columns=Joke
```
