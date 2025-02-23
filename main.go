package main

import (
	"log"
	"tablepilot/cmd/cli"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "tablepilot",
		Short: "AI table generator",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
	cli.BuildCLI(cmd)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
