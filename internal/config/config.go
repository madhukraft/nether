package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/madhukraft/nether/internal/modrinth"
)

type ModsConfig struct {
	AutoUpdate      bool                  `toml:"auto_update"`
	AutoInstallDeps bool                  `toml:"auto_install_deps"`
	Installed       []modrinth.InstalledMod `toml:"installed"`
}

type ModpackConfig struct {
	Installed []modrinth.InstalledModpack `toml:"installed"`
}

type Config struct {
	Type     string         `toml:"type"`
	Version  string         `toml:"version"`
	Created  string         `toml:"created"`
	TargetOS string         `toml:"target_os,omitempty"`
	Mods     ModsConfig     `toml:"mods"`
	Modpacks ModpackConfig  `toml:"modpacks"`
}

func filename() string {
	return "nether.toml"
}

func Exists() bool {
	_, err := os.Stat(filename())
	return err == nil
}

func Load() (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(filename(), &cfg); err != nil {
		return nil, err
	}
	if cfg.Mods.Installed == nil {
		cfg.Mods.Installed = []modrinth.InstalledMod{}
	}
	if cfg.Modpacks.Installed == nil {
		cfg.Modpacks.Installed = []modrinth.InstalledModpack{}
	}
	return &cfg, nil
}

func (c *Config) UpsertMod(mod modrinth.InstalledMod) {
	for i, m := range c.Mods.Installed {
		if m.Slug == mod.Slug || m.ProjectID == mod.ProjectID {
			c.Mods.Installed[i] = mod
			return
		}
	}
	c.Mods.Installed = append(c.Mods.Installed, mod)
}

func (c *Config) AddModIfMissing(mod modrinth.InstalledMod) {
	for _, m := range c.Mods.Installed {
		if m.ProjectID == mod.ProjectID {
			return
		}
	}
	c.Mods.Installed = append(c.Mods.Installed, mod)
}

func (c *Config) HasMod(slugOrID string) bool {
	for _, m := range c.Mods.Installed {
		if m.Slug == slugOrID || m.ProjectID == slugOrID {
			return true
		}
	}
	return false
}

func (c *Config) UpsertModpack(mod modrinth.InstalledModpack) {
	for i, m := range c.Modpacks.Installed {
		if m.Slug == mod.Slug || m.ProjectID == mod.ProjectID {
			c.Modpacks.Installed[i] = mod
			return
		}
	}
	c.Modpacks.Installed = append(c.Modpacks.Installed, mod)
}

func Save(cfg *Config) error {
	f, err := os.Create(filename())
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
