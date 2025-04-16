#### Generate Example
```
tablepilot create examples/icon_jokes/icon_jokes.json

tablepilot generate icon_jokes -c 15 -b 3
```

#### Autofill Example
```
tablepilot create examples/icon_jokes/icon_jokes_autofill.json

tablepilot import examples/icon_jokes/icon_jokes_autofill.csv

tablepilot autofill icon_jokes -c 15 -b 3 --columns=Joke
```

#### Update: v1.1

You can now use files as the source type. This means there's no need to create a CSV listing all files beforehand.

```
tablepilot create examples/icon_jokes/icon_jokes_v1.1.json

tablepilot generate icon_jokes_v1_1 -c 15 -b 3

```
