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

		if cfg.HasMod(slug) && !reinstall {
			for _, m := range cfg.Mods.Installed {
				if m.Slug == slug || m.ProjectID == slug {
					fmt.Printf("%s %s already installed. Use --reinstall to update.\n", m.Slug, m.VersionNumber)
					return
				}
			}
		}

		c := modrinth.NewClient()

		result, err := c.InstallMod(slug, cfg.Version, cfg.Type, autoDeps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cfg.Mods.AutoInstallDeps = autoDeps
		cfg.UpsertMod(result.Mod)
		for _, dep := range result.Deps {
			cfg.AddModIfMissing(dep.Mod)
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

var modsUpdateCmd = &cobra.Command{
	Use:   "update [mod]",
	Short: "Update installed mods to latest versions",
	Long: `Check Modrinth for newer versions of installed mods and update them.
If a mod slug is given, only that mod is updated. Otherwise all installed
mods are checked.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := ensureServerInitialized()

		loader := modrinth.MapServerTypeToLoader(cfg.Type)
		if loader == "" {
			fmt.Fprintf(os.Stderr, "Server type %q does not support mods\n", cfg.Type)
			os.Exit(1)
		}

		c := modrinth.NewClient()

		var toUpdate []modrinth.InstalledMod
		if len(args) == 1 {
			slug := modrinth.ParseSlug(args[0])
			m := modrinth.FindInstalledMod(slug, cfg.Mods.Installed)
			if m == nil {
				fmt.Fprintf(os.Stderr, "Mod %q not found\n", slug)
				os.Exit(1)
			}
			toUpdate = append(toUpdate, *m)
		} else {
			toUpdate = cfg.Mods.Installed
		}

		updated := 0
		upToDate := 0
		skipped := 0

		for _, m := range toUpdate {
			latest, err := c.GetLatestVersion(m.ProjectID, cfg.Version, loader)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", m.Slug, err)
				skipped++
				continue
			}
			if latest == nil {
				fmt.Printf("%s — no version for %s %s\n", m.Slug, cfg.Type, cfg.Version)
				skipped++
				continue
			}

			if latest.ID == m.VersionID {
				upToDate++
				continue
			}

			primaryFile := modrinth.FindPrimaryFile(latest.Files)
			if primaryFile == nil {
				fmt.Fprintf(os.Stderr, "%s %s has no downloadable file, skipping\n", m.Slug, latest.VersionNumber)
				skipped++
				continue
			}

			if m.FileName != "" {
				oldPath := filepath.Join("mods", m.FileName)
				if err := os.Remove(oldPath); err == nil {
					fmt.Printf("Removed old %s\n", m.FileName)
				}
			} else {
				for _, pattern := range []string{
					fmt.Sprintf("mods/%s*.jar", m.Slug),
					fmt.Sprintf("mods/%s-*.jar", m.Slug),
				} {
					if matches, err := filepath.Glob(pattern); err == nil {
						for _, f := range matches {
							os.Remove(f)
							fmt.Printf("Removed old %s\n", filepath.Base(f))
						}
					}
				}
			}

			newPath := filepath.Join("mods", primaryFile.Filename)
			if err := modrinth.DownloadModFile(primaryFile.URL, newPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", m.Slug, err)
				skipped++
				continue
			}

			cfg.UpsertMod(modrinth.InstalledMod{
				Slug:          m.Slug,
				ProjectID:     m.ProjectID,
				VersionID:     latest.ID,
				VersionNumber: latest.VersionNumber,
				FileName:      primaryFile.Filename,
			})

			fmt.Printf("Updated %s %s → %s\n", m.Slug, m.VersionNumber, latest.VersionNumber)
			updated++
		}

		if updated > 0 {
			if err := config.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}
		}

		if len(args) == 1 {
			slug := modrinth.ParseSlug(args[0])
			if updated > 0 {
				fmt.Printf("%s updated\n", slug)
			} else if skipped == 0 {
				fmt.Printf("%s is up to date\n", slug)
			}
		} else if updated > 0 || skipped > 0 {
			parts := []string{}
			if updated > 0 {
				parts = append(parts, fmt.Sprintf("%d updated", updated))
			}
			if upToDate > 0 {
				parts = append(parts, fmt.Sprintf("%d up to date", upToDate))
			}
			if skipped > 0 {
				parts = append(parts, fmt.Sprintf("%d skipped", skipped))
			}
			fmt.Println(strings.Join(parts, ", "))
		} else {
			fmt.Println("All mods up to date")
		}
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

func searchCmd(projectType string) *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: fmt.Sprintf("Search %ss on Modrinth", projectType),
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := modrinth.NewClient()
			facets := map[string][]string{"project_type": {projectType}}
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
	modsCmd.AddCommand(modsUpdateCmd)
	modsCmd.AddCommand(modsListCmd)
	modsCmd.AddCommand(searchCmd("mod"))
	rootCmd.AddCommand(modsCmd)
}
