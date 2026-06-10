package config

import (
	"os"
	"testing"
)

func TestCheckTOMLOutput(t *testing.T) {
	cfg := &Config{
		Type:    "fabric",
		Version: "1.21.1",
		Created: "2026-06-10T00:00:00Z",
		Mods: ModsConfig{
			AutoUpdate:      true,
			AutoInstallDeps: true,
			Installed: []InstalledMod{
				{Slug: "fabric-api", ProjectID: "P7dR8mMs", VersionID: "abc", VersionNumber: "0.100.0"},
			},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("nether.toml")

	data, err := os.ReadFile("nether.toml")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TOML output:\n%s", data)
}

func TestLoadOldFormat(t *testing.T) {
	content := `# Nether server config
type = "paper"
version = "1.21.1"
created = "2026-06-10T00:00:00Z"
`
	if err := os.WriteFile("nether.toml", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("nether.toml")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Type != "paper" {
		t.Errorf("expected paper, got %s", cfg.Type)
	}
	if cfg.Version != "1.21.1" {
		t.Errorf("expected 1.21.1, got %s", cfg.Version)
	}
	if len(cfg.Mods.Installed) != 0 {
		t.Errorf("expected 0 mods, got %d", len(cfg.Mods.Installed))
	}
	if len(cfg.Modpacks.Installed) != 0 {
		t.Errorf("expected 0 modpacks, got %d", len(cfg.Modpacks.Installed))
	}
}

func TestSaveAndLoad(t *testing.T) {
	cfg := &Config{
		Type:    "fabric",
		Version: "1.21.1",
		Created: "2026-06-10T00:00:00Z",
		Mods: ModsConfig{
			AutoUpdate:      true,
			AutoInstallDeps: true,
			Installed: []InstalledMod{
				{Slug: "fabric-api", ProjectID: "P7dR8mMs", VersionID: "abc123", VersionNumber: "0.100.0"},
			},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("nether.toml")

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Type != "fabric" {
		t.Errorf("expected fabric, got %s", loaded.Type)
	}
	if loaded.Mods.AutoUpdate != true {
		t.Errorf("expected auto_update true, got %v", loaded.Mods.AutoUpdate)
	}
	if len(loaded.Mods.Installed) != 1 {
		t.Fatalf("expected 1 mod, got %d", len(loaded.Mods.Installed))
	}
	if loaded.Mods.Installed[0].Slug != "fabric-api" {
		t.Errorf("expected fabric-api, got %s", loaded.Mods.Installed[0].Slug)
	}
}

func TestExists(t *testing.T) {
	os.Remove("nether.toml")
	if Exists() {
		t.Error("expected false for missing file")
	}

	os.WriteFile("nether.toml", []byte("type = \"test\""), 0644)
	defer os.Remove("nether.toml")

	if !Exists() {
		t.Error("expected true for existing file")
	}
}

func TestLoadMissingFile(t *testing.T) {
	os.Remove("nether.toml")
	_, err := Load()
	if err == nil {
		t.Error("expected error for missing file")
	}
}
