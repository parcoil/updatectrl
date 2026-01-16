package main

import (
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "updatectrl",
		Version: version,
	}
	rootCmd.AddCommand(initCmd, watchCmd, buildCmd, listCmd, logsCmd)
	rootCmd.Execute()
}
