package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madhukraft/nether/internal/config"
	"github.com/madhukraft/nether/internal/modrinth"
	"github.com/spf13/cobra"
)

var modsCmd = &cobra.Command{
	Use:   "mods",
	Short: "Manage server mods via Modrinth",
}

var modsAddCmd = &cobra.Command{
	Use:   "add [mod]",
	Short: "Add a mod from Modrinth (slug or URL)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Println("No nether.toml found. Run 'nether create' first.")
			os.Exit(1)
		}

		autoDeps, _ := cmd.Flags().GetBool("auto-deps")
		c := modrinth.NewClient()

		result, err := c.InstallMod(slug, cfg.Version, cfg.Type, autoDeps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfg.Mods.AutoInstallDeps = autoDeps
		cfg.Mods.Installed = append(cfg.Mods.Installed, result.Mod)

		for _, dep := range result.Deps {
			found := false
			for _, existing := range cfg.Mods.Installed {
				if existing.ProjectID == dep.Mod.ProjectID {
					found = true
					break
				}
			}
			if !found {
				cfg.Mods.Installed = append(cfg.Mods.Installed, dep.Mod)
			}
		}

		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Installed %s %s\n", result.Mod.Slug, result.Mod.VersionNumber)
		for _, dep := range result.Deps {
			fmt.Printf("  dependency: %s %s\n", dep.Mod.Slug, dep.Mod.VersionNumber)
		}
	},
}

var modsRemoveCmd = &cobra.Command{
	Use:   "remove [mod]",
	Short: "Remove a mod (slug or project ID)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Println("No nether.toml found.")
			os.Exit(1)
		}

		var removed []modrinth.InstalledMod
		for _, m := range cfg.Mods.Installed {
			if m.Slug == slug || m.ProjectID == slug {
				jarPath := fmt.Sprintf("mods/%s*.jar", m.Slug)
				if matches, err := filepath.Glob(jarPath); err == nil {
					for _, f := range matches {
						os.Remove(f)
						fmt.Printf("Removed %s\n", f)
					}
				}
			} else {
				removed = append(removed, m)
			}
		}

		if len(removed) == len(cfg.Mods.Installed) {
			fmt.Printf("Mod %q not found\n", slug)
			os.Exit(1)
		}

		cfg.Mods.Installed = removed
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed %s\n", slug)
	},
}

var modsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed mods",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("No nether.toml found.")
			os.Exit(1)
		}

		if len(cfg.Mods.Installed) == 0 {
			fmt.Println("No mods installed")
			return
		}

		fmt.Println("Installed mods:")
		for _, m := range cfg.Mods.Installed {
			fmt.Printf("  %s %s\n", m.Slug, m.VersionNumber)
		}
	},
}

var modsSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search mods on Modrinth",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := modrinth.NewClient()
		results, err := c.Search(args[0], 10, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(results.Hits) == 0 {
			fmt.Println("No results found")
			return
		}

		for _, hit := range results.Hits {
			fmt.Printf("  %s - %s\n", hit.Slug, hit.Title)
			fmt.Printf("    %s\n", hit.Description)
			fmt.Printf("    Downloads: %d | Follows: %d\n", hit.Downloads, hit.Follows)
			fmt.Println()
		}
	},
}

func init() {
	modsAddCmd.Flags().Bool("auto-deps", true, "Automatically install required and optional dependencies")
	modsCmd.AddCommand(modsAddCmd)
	modsCmd.AddCommand(modsRemoveCmd)
	modsCmd.AddCommand(modsListCmd)
	modsCmd.AddCommand(modsSearchCmd)
	rootCmd.AddCommand(modsCmd)
}
