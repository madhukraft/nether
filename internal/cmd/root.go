package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
    Use:     "nether",
    Short:   "Nether - Minecraft server setup tool",
    Version: Version,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
