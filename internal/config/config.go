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

func Save(cfg *Config) error {
	f, err := os.Create(filename())
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}
