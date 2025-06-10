package main

import (
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
	for _, tt := range snapshots {
		t.Run(tt.snapshot, func(t *testing.T) {
			t.Setenv("TABLEPILOT_SNAPSHOT_TEST", tt.snapshot)
			defer func() { _ = os.Remove("test.db") }()
			defer func() { _ = os.Remove("tmp.csv") }()
			defer func() { _ = os.RemoveAll("data") }()

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			counter := 0
			snapshots := []snapshot{}
			raw, err := os.ReadFile("snapshots/" + tt.snapshot + ".json")
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

			for _, command := range tt.prepare {
				command = append(command, "--config", "test.toml")
				root.SetArgs(command)
				err = root.Execute()
				require.NoError(t, err)
			}

			p := "cases/" + tt.example
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

			compareCSVFiles(t, "tmp.csv", "snapshots/"+tt.snapshot+".csv")
		})
	}
}

func TestIntegrationCLI_SnapshotsAutofill(t *testing.T) {
	for _, tt := range autofills {
		t.Run(tt.snapshot, func(t *testing.T) {
			t.Setenv("TABLEPILOT_SNAPSHOT_TEST", tt.snapshot)
			defer func() { _ = os.Remove("test.db") }()
			defer func() { _ = os.Remove("tmp.csv") }()
			defer func() { _ = os.RemoveAll("data") }()

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			counter := 0
			snapshots := []snapshot{}
			raw, err := os.ReadFile("snapshots/" + tt.snapshot + ".json")
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

			p := "cases/" + tt.example
			b, err := os.ReadFile(p)
			require.NoError(t, err)
			tableName := gjson.GetBytes(b, "name").String()
			root.SetArgs([]string{"export", tableName, "--config", "test.toml", "--to", "tmp.csv"})
			err = root.Execute()
			require.NoError(t, err)

			compareCSVFiles(t, "tmp.csv", "snapshots/"+tt.snapshot+".csv")
		})
	}
}
