package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show the status of the server in the current directory",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Checking server status...")
    },
}

func init() {
    rootCmd.AddCommand(statusCmd)
}
