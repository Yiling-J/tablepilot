package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/Yiling-J/tablepilot/cmd/cli"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func compareCSVFiles(t *testing.T, file1, file2 string) {
	f1, err := os.Open(file1)
	require.NoError(t, err)
	defer f1.Close()

	f2, err := os.Open(file2)
	require.NoError(t, err)
	defer f2.Close()

	r1 := csv.NewReader(f1)
	r2 := csv.NewReader(f2)

	for {
		row1, err1 := r1.Read()
		row2, err2 := r2.Read()

		if err1 != nil || err2 != nil {
			require.Equal(t, err1, err2)
			return
		}
		require.Equal(t, row1, row2)
	}
}

func TestIntegrationCLI_Snapshots(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.snapshot, func(t *testing.T) {
			t.Setenv("TABLEPILOT_SNAPSHOT_TEST", tt.snapshot)
			defer func() { _ = os.Remove("test.db") }()
			defer func() { _ = os.Remove("tmp.csv") }()

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			counter := 0
			snapshots := []snapshot{}
			raw, err := os.ReadFile("../snapshots/" + tt.snapshot + ".json")
			require.NoError(t, err)
			err = json.Unmarshal(raw, &snapshots)
			require.NoError(t, err)

			httpmock.RegisterResponder(
				"POST", "https://models.inference.ai.azure.com/chat/completions",
				func(r *http.Request) (*http.Response, error) {
					b, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.JSONEq(t, string(b), snapshots[counter].Request)
					resp := httpmock.NewStringResponse(200, snapshots[counter].Response)
					resp.Header.Add("content-type", "application/json")
					counter++
					return resp, nil
				},
			)
			if tt.snapshot == "recipes_by_country_kaggle" {
				b, err := os.ReadFile("examples/countries.csv")
				require.NoError(t, err)
				buf := new(bytes.Buffer)
				zipWriter := zip.NewWriter(buf)
				defer zipWriter.Close()

				csvFile, err := zipWriter.Create("countries.csv")
				require.NoError(t, err)
				_, err = csvFile.Write(b)
				require.NoError(t, err)
				err = zipWriter.Close()
				require.NoError(t, err)
				httpmock.RegisterResponder("GET", "https://www.kaggle.com/api/v1/datasets/download/fernandol/countries-of-the-world",
					httpmock.NewBytesResponder(200, buf.Bytes()).Once())
				defer func() { _ = os.RemoveAll("tablepilot_kaggle_cache") }()
			}
			if tt.snapshot == "chinese_qa_parquet_huggingface" {
				hfcounter := 0
				hfsnapshots := []snapshot{}
				hfraw, err := os.ReadFile("../snapshots/" + tt.snapshot + "_hf.json")
				require.NoError(t, err)
				err = json.Unmarshal(hfraw, &hfsnapshots)
				require.NoError(t, err)
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

			for _, command := range tt.prepare {
				command = append(command, "--config", "test.toml")
				root.SetArgs(command)
				err = root.Execute()
				require.NoError(t, err)
			}

			p := "../../examples/" + tt.example
			b, err := os.ReadFile(p)
			require.NoError(t, err)
			tableName := gjson.GetBytes(b, "name").String()
			root.SetArgs([]string{
				"create", p, "--config", "test.toml",
			})
			err = root.Execute()
			require.NoError(t, err)

			root.SetArgs([]string{"generate", tableName, "-c", "5", "-b", "3", "--config", "test.toml", "--saveto", "tmp.csv"})
			err = root.Execute()
			require.NoError(t, err)

			compareCSVFiles(t, "tmp.csv", "../snapshots/"+tt.snapshot+".csv")
		})
	}
}

func TestIntegrationCLI_SnapshotsAutofill(t *testing.T) {
	tests := []struct {
		snapshot string
		example  string
		commands [][]string
	}{
		{"pokemons", "pokemons/pokemons.json", [][]string{
			{"create", "../../examples/pokemons/pokemons.json"},
			{"import", "examples/pokemons/pokemons.csv", "-t", "pokemons"},
			{"autofill", "pokemons", "-c", "5", "-b", "3", "--columns", "Ecology"},
		}},
		{"pokemons_autofill", "pokemons_autofill/pokemons.json", [][]string{
			{"create", "../../examples/pokemons_autofill/pokemons.json"},
			{"import", "examples/pokemons_autofill/data.csv", "-t", "pokemon_stories"},
			{"autofill", "pokemon_stories", "-c", "5", "-b", "3", "--columns", "Story"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.snapshot, func(t *testing.T) {
			t.Setenv("TABLEPILOT_SNAPSHOT_TEST", tt.snapshot)
			defer func() { _ = os.Remove("test.db") }()
			defer func() { _ = os.Remove("tmp.csv") }()

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			counter := 0
			snapshots := []snapshot{}
			raw, err := os.ReadFile("../snapshots/" + tt.snapshot + ".json")
			require.NoError(t, err)
			err = json.Unmarshal(raw, &snapshots)
			require.NoError(t, err)

			httpmock.RegisterResponder(
				"POST", "https://models.inference.ai.azure.com/chat/completions",
				func(r *http.Request) (*http.Response, error) {
					b, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.JSONEq(t, string(b), snapshots[counter].Request)
					resp := httpmock.NewStringResponse(200, snapshots[counter].Response)
					resp.Header.Add("content-type", "application/json")
					counter++
					return resp, nil
				},
			)

			root := &cobra.Command{
				Use:   "",
				Short: "",
				CompletionOptions: cobra.CompletionOptions{
					HiddenDefaultCmd: true,
				},
			}
			_ = cli.BuildCLI(root)

			for _, command := range tt.commands {
				command = append(command, "--config", "test.toml")
				root.SetArgs(command)
				err = root.Execute()
				require.NoError(t, err)
			}

			p := "../../examples/" + tt.example
			b, err := os.ReadFile(p)
			require.NoError(t, err)
			tableName := gjson.GetBytes(b, "name").String()
			root.SetArgs([]string{"export", tableName, "--config", "test.toml", "--to", "tmp.csv"})
			err = root.Execute()
			require.NoError(t, err)

			compareCSVFiles(t, "tmp.csv", "../snapshots/"+tt.snapshot+".csv")
		})
	}
}
