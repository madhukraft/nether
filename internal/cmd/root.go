package cmd

import (
    "fmt"
    "os"
    "runtime/debug"

    "github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
    Use:     "nether",
    Short:   "Nether - Minecraft server setup tool",
}

func init() {
    if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
        if Version == "dev" {
            Version = info.Main.Version
        }
    }
    rootCmd.Version = Version
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
