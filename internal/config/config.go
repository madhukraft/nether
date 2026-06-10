package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type InstalledMod struct {
	Slug          string `toml:"slug"`
	ProjectID     string `toml:"project_id"`
	VersionID     string `toml:"version_id"`
	VersionNumber string `toml:"version_number"`
}

type ModsConfig struct {
	AutoUpdate      bool           `toml:"auto_update"`
	AutoInstallDeps bool           `toml:"auto_install_deps"`
	Installed       []InstalledMod `toml:"installed"`
}

type InstalledModpack struct {
	Slug          string `toml:"slug"`
	ProjectID     string `toml:"project_id"`
	VersionID     string `toml:"version_id"`
	VersionNumber string `toml:"version_number"`
}

type ModpackConfig struct {
	Installed []InstalledModpack `toml:"installed"`
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
		cfg.Mods.Installed = []InstalledMod{}
	}
	if cfg.Modpacks.Installed == nil {
		cfg.Modpacks.Installed = []InstalledModpack{}
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
