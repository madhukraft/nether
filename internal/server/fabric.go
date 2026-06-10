package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/madhukraft/nether/internal/modrinth"
)

const fabricMetaURL = "https://meta.fabricmc.net/v2"

type fabricVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

func getLatestStable(versions []fabricVersion) (string, error) {
	for _, v := range versions {
		if v.Stable {
			return v.Version, nil
		}
	}
	if len(versions) > 0 {
		return versions[0].Version, nil
	}
	return "", fmt.Errorf("no versions found")
}

func fetchFabricVersions(endpoint string) ([]fabricVersion, error) {
	resp, err := http.Get(fabricMetaURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from Fabric meta API", resp.StatusCode)
	}

	var versions []fabricVersion
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", endpoint, err)
	}
	return versions, nil
}

func downloadFabricAPI(mcVersion string) error {
	c := modrinth.NewClient()
	versions, err := c.GetVersions("fabric-api", []string{"fabric"}, []string{mcVersion})
	if err != nil {
		return fmt.Errorf("failed to fetch Fabric API version: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Fabric API version found for Minecraft %s", mcVersion)
	}

	ver := &versions[0]
	primaryFile := modrinth.FindPrimaryFile(ver.Files)
	if primaryFile == nil {
		return fmt.Errorf("no primary file found for Fabric API version %s", ver.VersionNumber)
	}

	return downloadFile(primaryFile.URL, "mods/"+primaryFile.Filename, "Fabric API jar")
}

func DownloadFabric(mcVersion string) error {
	loaderVersions, err := fetchFabricVersions("/versions/loader")
	if err != nil {
		return err
	}
	loaderVer, err := getLatestStable(loaderVersions)
	if err != nil {
		return fmt.Errorf("no Fabric loader versions available: %w", err)
	}

	installerVersions, err := fetchFabricVersions("/versions/installer")
	if err != nil {
		return err
	}
	installerVer, err := getLatestStable(installerVersions)
	if err != nil {
		return fmt.Errorf("no Fabric installer versions available: %w", err)
	}

	jarURL := fmt.Sprintf("%s/versions/loader/%s/%s/%s/server/jar", fabricMetaURL, mcVersion, loaderVer, installerVer)

	label := fmt.Sprintf("Fabric loader %s", loaderVer)
	if err := downloadFile(jarURL, "server.jar", label); err != nil {
		if isNotFound(err) {
			return fmt.Errorf("Minecraft version %s is not supported by Fabric", mcVersion)
		}
		return err
	}

	if err := os.MkdirAll("mods", 0755); err != nil {
		return fmt.Errorf("failed to create mods directory: %w", err)
	}

	fmt.Println("Downloading Fabric API...")
	if err := downloadFabricAPI(mcVersion); err != nil {
		return fmt.Errorf("failed to download Fabric API: %w", err)
	}

	return nil
}
