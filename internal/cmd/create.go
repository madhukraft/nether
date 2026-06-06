package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/madhukraft/nether/internal/server"
	"github.com/spf13/cobra"
)

var serverType string
var serverVersion string

func promptEula(in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "Do you accept the Minecraft EULA (https://aka.ms/MinecraftEULA)? [y/N]: ")
	reader := bufio.NewReader(in)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	response = strings.TrimSpace(response)
	return response == "y" || response == "Y", nil
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Minecraft server in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat("nether.toml"); err == nil {
			fmt.Println("error: a nether server already exists in this directory")
			os.Exit(1)
		}

		accepted, err := promptEula(os.Stdin, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
			os.Exit(1)
		}
		if !accepted {
			fmt.Println("EULA not accepted. Aborting.")
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
	        os.Exit(0)
	    default:
	        fmt.Fprintf(os.Stderr, "error: unknown server type %q\n", serverType)
	        os.Exit(1)
	    }

	    if err := server.WriteEula(); err != nil {
	        fmt.Fprintf(os.Stderr, "error writing eula.txt: %v\n", err)
	        os.Exit(1)
	    }

	    if err := server.WriteConfig(serverType, serverVersion); err != nil {
	        fmt.Fprintf(os.Stderr, "error writing nether.toml: %v\n", err)
	        os.Exit(1)
	    }

	    fmt.Println("Done! Server ready.")
	},
}

func init() {
	createCmd.Flags().StringVarP(&serverType, "type", "t", "paper", "Server type (paper, vanilla)")
	createCmd.Flags().StringVarP(&serverVersion, "version", "v", "", "Minecraft version (e.g. 1.21.1)")
	createCmd.MarkFlagRequired("version")
	rootCmd.AddCommand(createCmd)
}
