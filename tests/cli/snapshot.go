package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Yiling-J/tablepilot/cmd/cli"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

type snapshot struct {
	Request  string `json:"request"`
	Response string `json:"response"`
}

var snapshots = []struct {
	snapshot string
	example  string
	prepare  [][]string
}{
	// test auto column gen, pick from list dataset and pick from options
	{"recipes", "recipes.json", [][]string{
		{"dataset", "create", "--name", "ings", "--type", "list", "--path", "cases/ingredients.txt"},
	}},
	// test pick from table with context
	{"recipes_for_customers", "recipes_for_customers.json", [][]string{
		{"create", "cases/customers.json"},
		{"generate", "customers", "-c", "5", "-b", "5"},
	}},
	// test pick from csv dataset with wildcard path
	{"imdb_movie_haiku", "haiku.json", [][]string{
		{"dataset", "create", "--name", "movies", "--type", "csv", "--path", "cases/movies/*.csv"},
	}},
	// vision
	{"icon_jokes", "icon_jokes.json", [][]string{
		{"dataset", "create", "--name", "icons", "--type", "csv", "--path", "cases/icons/icons.csv"},
	}},
}

var autofills = []struct {
	snapshot string
	example  string
	commands [][]string
}{
	// autofill
	{"pokemons", "pokemons.json", [][]string{
		{"create", "cases/pokemons.json"},
		{"import", "cases/pokemons.csv", "-t", "pokemons"},
		{"autofill", "pokemons", "-c", "5", "-b", "3", "--columns", "Ecology"},
	}},
	// autofill based on linked column with context
	{"pokemons_autofill", "pokemons_autofill.json", [][]string{
		{"dataset", "create", "--name", "pokemons", "--type", "csv", "--path", "cases/pokemons.csv"},
		{"create", "cases/pokemons_autofill.json"},
		{"import", "cases/stories.csv", "-t", "pokemon_stories"},
		{"autofill", "pokemon_stories", "-c", "5", "-b", "3", "--columns", "Story"},
	}},
}

func main() {
	defer func() {
		_ = os.Remove("test.db")
		_ = os.Remove("tmp.csv")
	}()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST", "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
		func(r *http.Request) (*http.Response, error) {
			client := &http.Client{
				Transport: &http.Transport{
					Proxy:                 http.ProxyFromEnvironment,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
			}
			return client.Do(r)
		},
	)

	for _, sn := range snapshots {
		fmt.Println("Generating snapshot for " + sn.snapshot)
		root := &cobra.Command{
			Use:   "",
			Short: "",
			CompletionOptions: cobra.CompletionOptions{
				HiddenDefaultCmd: true,
			},
		}
		_ = cli.BuildCLI(root)
		_ = os.Remove("test.db")
		_ = os.Remove("tmp.csv")
		_ = os.Remove(fmt.Sprintf("snapshots/%s.json", sn.snapshot))
		_ = os.MkdirAll("snapshots", os.ModePerm)
		os.Setenv("TABLEPILOT_SNAPSHOT_RECORD", sn.snapshot)

		for _, command := range sn.prepare {
			command = append(command, "--config", "test.toml")
			root.SetArgs(command)
			err := root.Execute()
			if err != nil {
				panic(err)
			}
		}

		p := "cases/" + sn.example
		b, err := os.ReadFile(p)
		if err != nil {
			panic(err)
		}
		tableName := gjson.GetBytes(b, "name").String()
		root.SetArgs([]string{
			"create", p, "--config", "test.toml",
		})
		err = root.Execute()
		if err != nil {
			panic(err)
		}

		root.SetArgs([]string{"generate", tableName, "-c", "5", "-b", "3", "--config", "test.toml", "--saveto", "tmp.csv"})
		err = root.Execute()
		if err != nil {
			panic(err)
		}
		err = os.Rename("tmp.csv", fmt.Sprintf("snapshots/%s.csv", sn.snapshot))
		if err != nil {
			panic(err)
		}
	}

	for _, af := range autofills {
		fmt.Println("Generating snapshot for " + af.snapshot)
		root := &cobra.Command{
			Use:   "",
			Short: "",
			CompletionOptions: cobra.CompletionOptions{
				HiddenDefaultCmd: true,
			},
		}
		_ = cli.BuildCLI(root)
		_ = os.Remove("test.db")
		_ = os.Remove("tmp.csv")
		_ = os.Remove(fmt.Sprintf("snapshots/%s.json", af.snapshot))
		_ = os.MkdirAll("snapshots", os.ModePerm)
		os.Setenv("TABLEPILOT_SNAPSHOT_RECORD", af.snapshot)

		for _, command := range af.commands {
			command = append(command, "--config", "test.toml")
			root.SetArgs(command)
			err := root.Execute()
			if err != nil {
				panic(err)
			}
		}

		p := "cases/" + af.example
		b, err := os.ReadFile(p)
		if err != nil {
			panic(err)
		}
		tableName := gjson.GetBytes(b, "name").String()
		root.SetArgs([]string{"export", tableName, "--config", "test.toml", "--to", "tmp.csv"})
		err = root.Execute()
		if err != nil {
			panic(err)
		}
		err = os.Rename("tmp.csv", fmt.Sprintf("snapshots/%s.csv", af.snapshot))
		if err != nil {
			panic(err)
		}
	}
}
