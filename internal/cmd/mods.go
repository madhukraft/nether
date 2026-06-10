package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/madhukraft/nether/internal/config"
	"github.com/madhukraft/nether/internal/modrinth"
	"github.com/spf13/cobra"
)

var modsCmd = &cobra.Command{
	Use:   "mods",
	Short: "Manage server mods via Modrinth",
}

var modsInstallCmd = &cobra.Command{
	Use:   "install [mod]",
	Short: "Install a mod from Modrinth (slug or URL)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]

		cfg := ensureServerInitialized()

		autoDeps, _ := cmd.Flags().GetBool("auto-deps")
		reinstall, _ := cmd.Flags().GetBool("reinstall")

		slug = modrinth.ParseSlug(slug)

		for _, m := range cfg.Mods.Installed {
			if m.Slug == slug || m.ProjectID == slug {
				if !reinstall {
					fmt.Printf("%s %s already installed. Use --reinstall to update.\n", m.Slug, m.VersionNumber)
					return
				}
				break
			}
		}

		c := modrinth.NewClient()

		result, err := c.InstallMod(slug, cfg.Version, cfg.Type, autoDeps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfg.Mods.AutoInstallDeps = autoDeps

		found := false
		for i, m := range cfg.Mods.Installed {
			if m.Slug == result.Mod.Slug || m.ProjectID == result.Mod.ProjectID {
				cfg.Mods.Installed[i] = result.Mod
				found = true
				break
			}
		}
		if !found {
			cfg.Mods.Installed = append(cfg.Mods.Installed, result.Mod)
		}

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

		if reinstall {
			fmt.Printf("Updated %s %s\n", result.Mod.Slug, result.Mod.VersionNumber)
		} else {
			fmt.Printf("Installed %s %s\n", result.Mod.Slug, result.Mod.VersionNumber)
		}
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

		cfg := ensureServerInitialized()

		var kept []modrinth.InstalledMod
		for _, m := range cfg.Mods.Installed {
			if m.Slug == slug || m.ProjectID == slug {
				if m.FileName != "" {
					jarPath := filepath.Join("mods", m.FileName)
					if err := os.Remove(jarPath); err == nil {
						fmt.Printf("Removed %s\n", jarPath)
					}
				} else {
					for _, pattern := range []string{
						fmt.Sprintf("mods/%s*.jar", m.Slug),
						fmt.Sprintf("mods/%s-*.jar", m.Slug),
					} {
						if matches, err := filepath.Glob(pattern); err == nil {
							for _, f := range matches {
								os.Remove(f)
								fmt.Printf("Removed %s\n", f)
							}
						}
					}
				}
			} else {
				kept = append(kept, m)
			}
		}

		if len(kept) == len(cfg.Mods.Installed) {
			fmt.Printf("Mod %q not found\n", slug)
			os.Exit(1)
		}

		cfg.Mods.Installed = kept
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
			status := "✓"
			if m.FileName != "" {
				if _, err := os.Stat(filepath.Join("mods", m.FileName)); err != nil {
					status = "✗"
				}
			} else {
				status = "?"
			}
			fmt.Printf("  %s %s %s\n", status, m.Slug, m.VersionNumber)
		}
	},
}

var modsSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search mods on Modrinth",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := modrinth.NewClient()
		facets := map[string][]string{"project_type": {"mod"}}
		results, err := c.Search(args[0], 10, facets)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(results.Hits) == 0 {
			fmt.Println("No results found")
			return
		}

		for _, hit := range results.Hits {
			loaders := ""
			if len(hit.Loaders) > 0 {
				loaders = " [" + strings.Join(hit.Loaders, ", ") + "]"
			}
			fmt.Printf("  %s - %s%s\n", hit.Slug, hit.Title, loaders)
			fmt.Printf("    %s\n", hit.Description)
			fmt.Printf("    Downloads: %-9s | Follows: %s\n", humanNumber(hit.Downloads), humanNumber(hit.Follows))
			fmt.Println()
		}
	},
}

func humanNumber(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}


func init() {
	modsInstallCmd.Flags().Bool("auto-deps", true, "Automatically install required and optional dependencies")
	modsInstallCmd.Flags().Bool("reinstall", false, "Reinstall mod even if already installed")
	modsCmd.AddCommand(modsInstallCmd)
	modsCmd.AddCommand(modsRemoveCmd)
	modsCmd.AddCommand(modsListCmd)
	modsCmd.AddCommand(modsSearchCmd)
	rootCmd.AddCommand(modsCmd)
}
