package server

import (
	"fmt"
	"net/http"
	"strings"
)

func ValidateVersion(serverType, version string) error {
	switch serverType {
	case "paper":
		return validatePaper(version)
	case "vanilla":
		return validateVanilla(version)
	case "fabric":
		return validateFabric(version)
	case "neoforge":
		return validateNeoForge(version)
	case "forge":
		return validateForge(version)
	default:
		return fmt.Errorf("unknown server type: %s", serverType)
	}
}

func validatePaper(version string) error {
	url := fmt.Sprintf("https://api.papermc.io/v2/projects/paper/versions/%s", version)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to reach Paper API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("version %q not found on Paper API", version)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from Paper API", resp.StatusCode)
	}
	return nil
}

func validateVanilla(version string) error {
	var manifest struct {
		Versions []struct {
			ID string `json:"id"`
		} `json:"versions"`
	}
	if err := fetchJSON(vanillaManifestURL, &manifest); err != nil {
		return fmt.Errorf("failed to fetch version manifest: %w", err)
	}
	for _, v := range manifest.Versions {
		if v.ID == version {
			return nil
		}
	}
	return fmt.Errorf("version %q not found in Mojang manifest", version)
}

func validateFabric(version string) error {
	versions, err := fetchFabricVersions("/versions/game")
	if err != nil {
		return err
	}
	for _, v := range versions {
		if v.Version == version {
			return nil
		}
	}
	return fmt.Errorf("Minecraft version %q is not supported by Fabric", version)
}

func validateNeoForge(version string) error {
	_, err := getNeoForgeVersionForMC(version)
	if err != nil {
		// Clean up error message for user-facing prompt
		msg := err.Error()
		if strings.Contains(msg, "no NeoForge version found") {
			return fmt.Errorf("no NeoForge version available for Minecraft %s", version)
		}
		return err
	}
	return nil
}

func validateForge(version string) error {
	_, err := getForgeVersionForMC(version)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "no Forge version found") {
			return fmt.Errorf("no Forge version available for Minecraft %s", version)
		}
		return err
	}
	return nil
}
