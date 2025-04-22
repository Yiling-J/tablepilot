package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
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
	{"recipes_simple", "recipes_simple/recipes.json", [][]string{}},
	{"recipes_ai_columns", "recipes_ai_columns/recipes.json", [][]string{}},
	{"recipes_cmi", "recipes_cuisine_meal_ingredients/recipes.json", [][]string{}},
	{"recipes_by_country", "recipes_by_country/recipes.json", [][]string{}},
	{"recipes_by_country_kaggle", "recipes_by_country_kaggle/recipes.json", [][]string{}},
	{"recipes_for_customers", "recipes_for_customers/recipes.json", [][]string{
		{"create", "../../examples/recipes_for_customers/customers.json"},
		{"generate", "customers", "-c", "5", "-b", "5"},
	}},
	{"imdb_movie_haiku", "imdb_movie_haiku/haiku.json", [][]string{}},
	{"chinese_qa_parquet", "chinese_qa_parquet/table.json", [][]string{}},
	{"chinese_qa_parquet_huggingface", "chinese_qa_parquet_huggingface/table.json", [][]string{}},
	{"icon_jokes", "icon_jokes/icon_jokes.json", [][]string{}},
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
		if sn.snapshot == "recipes_by_country_kaggle" {
			b, err := os.ReadFile("examples/countries.csv")
			if err != nil {
				panic(err)
			}
			buf := new(bytes.Buffer)
			zipWriter := zip.NewWriter(buf)
			defer zipWriter.Close()

			csvFile, err := zipWriter.Create("countries.csv")
			if err != nil {
				panic(err)
			}
			_, err = csvFile.Write(b)
			if err != nil {
				panic(err)
			}
			err = zipWriter.Close()
			if err != nil {
				panic(err)
			}
			httpmock.RegisterResponder("GET", "https://www.kaggle.com/api/v1/datasets/download/fernandol/countries-of-the-world",
				httpmock.NewBytesResponder(200, buf.Bytes()).Once())
			defer func() { _ = os.RemoveAll("tablepilot_kaggle_cache") }()
		}
		if sn.snapshot == "chinese_qa_parquet_huggingface" {
			hfcounter := 0
			hfsnapshots := []snapshot{}
			hfraw, err := os.ReadFile("../snapshots/" + sn.snapshot + "_hf.json")
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(hfraw, &hfsnapshots)
			if err != nil {
				panic(err)
			}
			httpmock.RegisterResponder(
				"GET", `=~^https://datasets-server\.huggingface\.co`,
				func(r *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, hfsnapshots[hfcounter].Response)
					resp.Header.Add("content-type", "application/json")
					hfcounter++
					return resp, nil
				},
			)
		}
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
		_ = os.Remove(fmt.Sprintf("../snapshots/%s.json", sn.snapshot))
		os.Setenv("TABLEPILOT_SNAPSHOT_RECORD", sn.snapshot)

		for _, command := range sn.prepare {
			command = append(command, "--config", "test.toml")
			root.SetArgs(command)
			err := root.Execute()
			if err != nil {
				panic(err)
			}
		}

		p := "../../examples/" + sn.example
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
		err = os.Rename("tmp.csv", fmt.Sprintf("../snapshots/%s.csv", sn.snapshot))
		if err != nil {
			panic(err)
		}
	}

	autofills := []struct {
		snapshot string
		example  string
		commands [][]string
	}{
		{"pokemons", "pokemons/pokemons.json", [][]string{
			{"create", "../../examples/pokemons/pokemons.json"},
			{"import", "examples/pokemons/pokemons.csv"},
			{"autofill", "pokemons", "-c", "5", "-b", "3", "--columns", "Ecology"},
		}},
		{"pokemons_autofill", "pokemons_autofill/pokemons.json", [][]string{
			{"create", "../../examples/pokemons_autofill/pokemons.json"},
			{"import", "examples/pokemons_autofill/data.csv", "-t", "pokemon_stories"},
			{"autofill", "pokemon_stories", "-c", "5", "-b", "3", "--columns", "Story"},
		}},
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
		_ = os.Remove(fmt.Sprintf("../snapshots/%s.json", af.snapshot))
		os.Setenv("TABLEPILOT_SNAPSHOT_RECORD", af.snapshot)

		for _, command := range af.commands {
			command = append(command, "--config", "test.toml")
			root.SetArgs(command)
			err := root.Execute()
			if err != nil {
				panic(err)
			}
		}

		p := "../../examples/" + af.example
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
		err = os.Rename("tmp.csv", fmt.Sprintf("../snapshots/%s.csv", af.snapshot))
		if err != nil {
			panic(err)
		}
	}
}
