package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const vanillaManifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest_v2.json"

type versionManifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		URL  string `json:"url"`
	} `json:"versions"`
}

type versionMeta struct {
	Downloads struct {
		Server struct {
			SHA1 string `json:"sha1"`
			Size int    `json:"size"`
			URL  string `json:"url"`
		} `json:"server"`
	} `json:"downloads"`
}

func fetchJSON(url string, target any) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func getServerJarURL(version string) (string, error) {
	// Step 1: fetch the version manifest
	var manifest versionManifest
	if err := fetchJSON(vanillaManifestURL, &manifest); err != nil {
		return "", fmt.Errorf("failed to fetch version manifest: %w", err)
	}

	// resolve "latest" alias
	if version == "latest" {
		version = manifest.Latest.Release
	}

	// Step 2: find the URL for the requested version
	var versionURL string
	for _, v := range manifest.Versions {
		if v.ID == version {
			versionURL = v.URL
			break
		}
	}
	if versionURL == "" {
		return "", fmt.Errorf("version %q not found in manifest", version)
	}

	// Step 3: fetch the version metadata and extract the server jar URL
	var meta versionMeta
	if err := fetchJSON(versionURL, &meta); err != nil {
		return "", fmt.Errorf("failed to fetch version metadata: %w", err)
	}

	if meta.Downloads.Server.URL == "" {
		return "", fmt.Errorf("no server jar available for version %s", version)
	}

	return meta.Downloads.Server.URL, nil
}

func DownloadVanilla(version string) error {
	fmt.Printf("Fetching server info for version %q...\n", version)

	jarURL, err := getServerJarURL(version)
	if err != nil {
		return err
	}

	fmt.Println("Downloading vanilla server...")

	resp, err := http.Get(jarURL)
	if err != nil {
		return fmt.Errorf("failed to download server jar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d when downloading jar", resp.StatusCode)
	}

	out, err := os.Create("server.jar")
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}