package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func downloadFabricAPI(mcVersion string) error {
	v := url.Values{}
	v.Set("game_versions", `["`+mcVersion+`"]`)
	v.Set("loaders", `["fabric"]`)
	u := "https://api.modrinth.com/v2/project/fabric-api/version?" + v.Encode()

	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("failed to fetch Fabric API version: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("no Fabric API version found for Minecraft %s", mcVersion)
		}
		return fmt.Errorf("unexpected status %d from Modrinth API", resp.StatusCode)
	}

	var versions []struct {
		Files []struct {
			URL      string `json:"url"`
			Filename string `json:"filename"`
			Primary  bool   `json:"primary"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return fmt.Errorf("failed to parse Modrinth response: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Fabric API versions available for Minecraft %s", mcVersion)
	}

	for _, v := range versions {
		for _, f := range v.Files {
			if f.Primary {
				return downloadFile(f.URL, "mods/"+f.Filename)
			}
		}
	}

	f := versions[0].Files[0]
	return downloadFile(f.URL, "mods/"+f.Filename)
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

	fmt.Printf("Downloading Fabric server (loader %s)...\n", loaderVer)

	resp, err := http.Get(jarURL)
	if err != nil {
		return fmt.Errorf("failed to download Fabric server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Minecraft version %s is not supported by Fabric", mcVersion)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from Fabric meta API", resp.StatusCode)
	}

	out, err := os.Create("server.jar")
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
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
