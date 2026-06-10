package cmd

import (
	"fmt"
	"os"

	"github.com/madhukraft/nether/internal/config"
	"github.com/madhukraft/nether/internal/modrinth"
	"github.com/spf13/cobra"
)

var modpackCmd = &cobra.Command{
	Use:   "modpack",
	Short: "Manage modpack installations",
}

var modpackInstallCmd = &cobra.Command{
	Use:   "install [modpack]",
	Short: "Install a modpack from Modrinth (slug or URL)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]

		cfg := ensureServerInitialized()

		c := modrinth.NewClient()
		result, err := c.InstallModpack(slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfg.Modpacks.Installed = append(cfg.Modpacks.Installed, modrinth.InstalledModpack{
			Slug:          result.Mod.Slug,
			ProjectID:     result.Mod.ProjectID,
			VersionID:     result.Mod.VersionID,
			VersionNumber: result.Mod.VersionNumber,
		})

		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
		}

		fmt.Printf("Installed modpack: %s %s\n", result.Mod.Slug, result.Mod.VersionNumber)
		fmt.Printf("Downloaded %d files\n", len(result.Files))
	},
}

var modpackListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed modpacks",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("No nether.toml found.")
			os.Exit(1)
		}

		if len(cfg.Modpacks.Installed) == 0 {
			fmt.Println("No modpacks installed")
			return
		}

		fmt.Println("Installed modpacks:")
		for _, m := range cfg.Modpacks.Installed {
			fmt.Printf("  %s %s\n", m.Slug, m.VersionNumber)
		}
	},
}

func init() {
	modpackCmd.AddCommand(modpackInstallCmd)
	modpackCmd.AddCommand(modpackListCmd)
	rootCmd.AddCommand(modpackCmd)
}
