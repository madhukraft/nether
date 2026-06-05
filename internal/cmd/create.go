package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var serverType string
var serverVersion string

var createCmd = &cobra.Command{
    Use:   "create",
    Short: "Create a new Minecraft server in the current directory",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Creating %s server version %s...\n", serverType, serverVersion)
    },
}

func init() {
    createCmd.Flags().StringVarP(&serverType, "type", "t", "paper", "Server type (paper, vanilla)")
    createCmd.Flags().StringVarP(&serverVersion, "version", "v", "", "Minecraft version (e.g. 1.21.1)")
    createCmd.MarkFlagRequired("version")
    rootCmd.AddCommand(createCmd)
}
