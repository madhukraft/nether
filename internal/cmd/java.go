package cmd

import (
	"fmt"
	"os"

	"github.com/madhukraft/nether/internal/server"
	"github.com/spf13/cobra"
)

var javaCmd = &cobra.Command{
	Use:   "java",
	Short: "Manage Java runtime",
}

var javaInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Download and install a Java runtime",
	Long: `Downloads a Java runtime from Adoptium and installs it in the current
directory under java/. If no version is specified, Java 21 is used.`,
	Run: func(cmd *cobra.Command, args []string) {
		ver, _ := cmd.Flags().GetInt("version")
		jdk, _ := cmd.Flags().GetBool("jdk")

		if err := server.DownloadJavaVersion(ver, jdk, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	javaInstallCmd.Flags().Int("version", 21, "Java major version (e.g. 21)")
	javaInstallCmd.Flags().Bool("jdk", false, "Download JDK instead of JRE")
	javaCmd.AddCommand(javaInstallCmd)
	rootCmd.AddCommand(javaCmd)
}
