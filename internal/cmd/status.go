package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/madhukraft/nether/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server status in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Println("No nether.toml found — server not initialized by Nether")
			return
		}

		fmt.Printf("Server: %s %s\n", cfg.Type, cfg.Version)
		fmt.Printf("Mods installed: %d\n", len(cfg.Mods.Installed))
		if cfg.Modpacks.Installed != nil {
			fmt.Printf("Modpacks installed: %d\n", len(cfg.Modpacks.Installed))
		}

		if _, err := os.Stat("server.jar"); err == nil {
			fmt.Println("server.jar: present")
		} else {
			fmt.Println("server.jar: missing")
		}

		if _, err := os.Stat("eula.txt"); err == nil {
			fmt.Println("eula.txt: present")
		} else {
			fmt.Println("eula.txt: missing")
		}

		modsDir, err := os.ReadDir("mods")
		if err == nil {
			count := 0
			for _, f := range modsDir {
				if !f.IsDir() && filepath.Ext(f.Name()) == ".jar" {
					count++
				}
			}
			fmt.Printf("mods/: %d jar files\n", count)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
