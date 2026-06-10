package server

import (
	"fmt"
	"sort"
	"strings"
)

const forgeMavenURL = "https://maven.minecraftforge.net/net/minecraftforge/forge"

type forgeVersion struct {
	mcVersion string
	build     string
	parts     []int
}

func parseForgeVersions(versions []string) []forgeVersion {
	var out []forgeVersion
	for _, v := range versions {
		idx := strings.LastIndex(v, "-")
		if idx < 0 {
			continue
		}
		mcVer := v[:idx]
		buildStr := v[idx+1:]
		nums, ok := parseVersionParts(buildStr)
		if !ok {
			continue
		}
		out = append(out, forgeVersion{mcVersion: mcVer, build: buildStr, parts: nums})
	}
	return out
}

func getForgeVersionForMC(mcVersion string) (string, error) {
	meta, err := fetchMavenMetadata(forgeMavenURL)
	if err != nil {
		return "", fmt.Errorf("Forge: %w", err)
	}

	parsed := parseForgeVersions(meta.Versioning.Versions)

	var candidates []forgeVersion
	prefix := mcVersion + "-"
	for _, v := range parsed {
		if strings.HasPrefix(v.mcVersion+"-", prefix) {
			candidates = append(candidates, v)
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no Forge version found for Minecraft %s", mcVersion)
	}

	sort.Slice(candidates, func(i, j int) bool {
		ip := candidates[i].parts
		jp := candidates[j].parts
		for k := 0; k < len(ip) && k < len(jp); k++ {
			if ip[k] != jp[k] {
				return ip[k] > jp[k]
			}
		}
		return len(ip) > len(jp)
	})

	return candidates[0].mcVersion + "-" + candidates[0].build, nil
}

func DownloadForge(mcVersion string) (string, error) {
	fgVersion, err := getForgeVersionForMC(mcVersion)
	if err != nil {
		return "", err
	}

	jarURL := fmt.Sprintf("%s/%s/forge-%s-installer.jar", forgeMavenURL, fgVersion, fgVersion)
	installerFile := fmt.Sprintf("forge-%s-installer.jar", fgVersion)

	if err := downloadFile(jarURL, installerFile, "Forge "+fgVersion); err != nil {
		return "", err
	}
	return installerFile, nil
}

func InstallForge(installerFile string) error {
	return installForgeLike(installerFile, "Forge")
}
