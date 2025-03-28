package main

import (
	"log"

	"github.com/Yiling-J/tablepilot/cmd/cli"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "github.com/Yiling-J/tablepilot",
		Short: "A CLI tool designed to generate tables using AI",
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}
	_ = cli.BuildCLI(cmd)
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
