package cli

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Yiling-J/tablepilot/cmd/cli"
	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func captureStderr(f func() error) (string, error) {
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := f()
	os.Stderr = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func captureStdout(f func() error) (string, error) {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out), err
}

func TestIntegrationCLI_Basic(t *testing.T) {
	defer func() { _ = os.Remove("test.db") }()
	root := &cobra.Command{
		Use:   "",
		Short: "",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
	cl := cli.BuildCLI(root)
	root.SetArgs([]string{"list", "--config", "test.toml"})
	out, err := captureStdout(root.Execute)
	require.NoError(t, err)
	require.Equal(t, "", out)

	root.SetArgs([]string{"create", "test.json"})
	out, err = captureStderr(root.Execute)
	require.NoError(t, err)
	tables, err := cl.Backend.DB.TableMeta.Query().All(context.TODO())
	require.NoError(t, err)
	require.Equal(t, 1, len(tables))
	require.Contains(t, out, fmt.Sprintf(`table created	{"id": "%s"}`, tables[0].Nanoid))

	root.SetArgs([]string{"list", "--config", "test.toml"})
	out, err = captureStdout(root.Execute)
	require.NoError(t, err)
	require.Equal(t, `UkLWZg	recipes	table of recipes`, strings.TrimSpace(out))

	root.SetArgs([]string{"describe", tables[0].Name, "--config", "test.toml"})
	out, err = captureStdout(root.Execute)
	require.NoError(t, err)
	expectedDescribe := `
UkLWZg	Name	string	ai	recipe name
gbHJdm	Ingredients	array	ai	list of ingredients
EfhxLZ	Cooking Time	integer	ai	time required to cook the recipe (minutes)
VqXmZF	Cuisine	string	ai	type of cuisine
uw2YK1	Difficulty Level	integer	ai	difficulty level of the recipe (1-5)
`
	require.Equal(t, strings.TrimSpace(expectedDescribe), strings.TrimSpace(out))

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	counter := 0
	httpmock.RegisterResponder(
		"POST", "https://models.inference.ai.azure.com/chat/completions",
		func(r *http.Request) (*http.Response, error) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, simpleRecipeRequest, string(b))
			resp := httpmock.NewStringResponse(200, simpleRecipeResponse1)
			if counter > 0 {
				resp = httpmock.NewStringResponse(200, simpleRecipeResponse2)
			}
			resp.Header.Add("content-type", "application/json")
			counter++
			return resp, nil
		},
	)
	root.SetArgs([]string{"generate", tables[0].Name, "-c", "4", "-b", "2", "--config", "test.toml"})
	out, err = captureStdout(root.Execute)
	require.NoError(t, err)
	expectedGenerate := `
foo	["ing1","ing2"]	10	cc	1
foo1	["ing3","ing4"]	12	cc1	2
bar	["ing-1","ing-2"]	3	cd	2
bar1	["ing-3","ing-4"]	15	cd1	2
`
	require.Equal(t, strings.TrimSpace(expectedGenerate), strings.TrimSpace(out))

	counter = 0
	root.SetArgs([]string{"generate", tables[0].Name, "-c", "4", "-b", "2", "--config", "test.toml", "-s", "gen.csv"})
	out, err = captureStdout(root.Execute)
	require.NoError(t, err)
	expectedGenerate = `
foo	["ing1","ing2"]	10	cc	1
foo1	["ing3","ing4"]	12	cc1	2
bar	["ing-1","ing-2"]	3	cd	2
bar1	["ing-3","ing-4"]	15	cd1	2
`
	require.Equal(t, strings.TrimSpace(expectedGenerate), strings.TrimSpace(out))
	f, err := os.Open("gen.csv")
	require.NoError(t, err)
	defer f.Close()
	defer func() { _ = os.Remove("gen.csv") }()
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	require.NoError(t, err)
	require.Equal(t, [][]string{{"Name", "Ingredients", "Cooking Time", "Cuisine", "Difficulty Level"}, {"foo", "[\"ing1\",\"ing2\"]", "10", "cc", "1"}, {"foo1", "[\"ing3\",\"ing4\"]", "12", "cc1", "2"}, {"bar", "[\"ing-1\",\"ing-2\"]", "3", "cd", "2"}, {"bar1", "[\"ing-3\",\"ing-4\"]", "15", "cd1", "2"}}, records)

	root.SetArgs([]string{"show", tables[0].Name, "--config", "test.toml"})
	out, err = captureStdout(root.Execute)
	require.NoError(t, err)
	expectedShow := `
foo	["ing1","ing2"]	10	cc	1
foo1	["ing3","ing4"]	12	cc1	2
bar	["ing-1","ing-2"]	3	cd	2
bar1	["ing-3","ing-4"]	15	cd1	2
`
	require.Equal(t, strings.TrimSpace(expectedShow), strings.TrimSpace(out))

	root.SetArgs([]string{"export", tables[0].Name, "--config", "test.toml", "-t", "recipes.csv"})
	err = root.Execute()
	require.NoError(t, err)
	f, err = os.Open("recipes.csv")
	require.NoError(t, err)
	defer f.Close()
	defer func() { _ = os.Remove("recipes.csv") }()
	csvReader = csv.NewReader(f)
	records, err = csvReader.ReadAll()
	require.NoError(t, err)
	require.Equal(t, [][]string{{"Name", "Ingredients", "Cooking Time", "Cuisine", "Difficulty Level"}, {"foo", "[\"ing1\",\"ing2\"]", "10", "cc", "1"}, {"foo1", "[\"ing3\",\"ing4\"]", "12", "cc1", "2"}, {"bar", "[\"ing-1\",\"ing-2\"]", "3", "cd", "2"}, {"bar1", "[\"ing-3\",\"ing-4\"]", "15", "cd1", "2"}}, records)

	root.SetArgs([]string{"truncate", tables[0].Name, "--config", "test.toml"})
	out, err = captureStderr(root.Execute)
	require.NoError(t, err)
	require.Contains(t, strings.TrimSpace(out), `{"removed": 4}`)
	count, err := tables[0].QueryRows().Count(context.TODO())
	require.NoError(t, err)
	require.Equal(t, 0, count)

	root.SetArgs([]string{"delete", tables[0].Name, "--config", "test.toml"})
	err = root.Execute()
	require.NoError(t, err)
	total, err := cl.Backend.DB.TableMeta.Query().Count(context.TODO())
	require.NoError(t, err)
	require.Equal(t, 0, total)
}

func TestIntegrationCLI_Import(t *testing.T) {
	defer func() { _ = os.Remove("test.db") }()
	root := &cobra.Command{
		Use:   "",
		Short: "",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
	_ = cli.BuildCLI(root)

	records := [][]string{
		{"first_name", "last_name", "username"},
		{"Rob", "Pike", "rob"},
		{"Ken", "Thompson", "ken"},
		{"Robert", "Griesemer", "gri"},
	}

	f, err := os.Create("users.csv")
	require.NoError(t, err)
	defer os.Remove("users.csv")
	w := csv.NewWriter(f)
	err = w.WriteAll(records)
	require.NoError(t, err)

	root.SetArgs([]string{"import", "users.csv", "--config", "test.toml"})
	err = root.Execute()
	require.NoError(t, err)

	root.SetArgs([]string{"show", "users", "--config", "test.toml"})
	out, err := captureStdout(root.Execute)
	require.NoError(t, err)
	expectedShow := `
Rob	Pike	rob
Ken	Thompson	ken
Robert	Griesemer	gri
`
	require.Equal(t, strings.TrimSpace(expectedShow), strings.TrimSpace(out))

	root.SetArgs([]string{"import", "users.csv", "--config", "test.toml", "-t", "foo"})
	err = root.Execute()
	require.NoError(t, err)
	root.SetArgs([]string{"show", "foo", "--config", "test.toml"})
	out, err = captureStdout(root.Execute)
	require.NoError(t, err)
	expectedShow = `
Rob	Pike	rob
Ken	Thompson	ken
Robert	Griesemer	gri
`
	require.Equal(t, strings.TrimSpace(expectedShow), strings.TrimSpace(out))
}
