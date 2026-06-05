package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/madhukraft/nether/internal/server"
)

var serverType string
var serverVersion string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Minecraft server in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		// check if already a nether server
		if _, err := os.Stat("nether.toml"); err == nil {
			fmt.Println("error: a nether server already exists in this directory")
			os.Exit(1)
		}

		switch serverType {
		case "paper":
			if err := server.DownloadPaper(serverVersion); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "vanilla":
			fmt.Println("Vanilla support coming soon")
		default:
			fmt.Fprintf(os.Stderr, "error: unknown server type %q\n", serverType)
			os.Exit(1)
		}

		fmt.Println("Done! server.jar downloaded.")
	},
}

func init() {
	createCmd.Flags().StringVarP(&serverType, "type", "t", "paper", "Server type (paper, vanilla)")
	createCmd.Flags().StringVarP(&serverVersion, "version", "v", "", "Minecraft version (e.g. 1.21.1)")
	createCmd.MarkFlagRequired("version")
	rootCmd.AddCommand(createCmd)
}
