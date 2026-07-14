package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "updatectrl",
		Version: version,
	}
	rootCmd.AddCommand(initCmd, watchCmd, buildCmd, listCmd, logsCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
