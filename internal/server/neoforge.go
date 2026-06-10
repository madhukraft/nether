package server

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

const neoforgeMavenURL = "https://maven.neoforged.net/releases/net/neoforged/neoforge"

type neoForgeVersion struct {
	mcMinor int
	full    string
	parts   []int
}

func parseNeoForgeVersions(versions []string) []neoForgeVersion {
	var out []neoForgeVersion
	for _, v := range versions {
		if strings.Contains(v, "beta") || strings.Contains(v, "craftmine") {
			continue
		}
		parts := strings.Split(v, ".")
		if len(parts) < 2 {
			continue
		}
		mcMinor, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		nums, ok := parseVersionParts(v)
		if !ok {
			continue
		}
		out = append(out, neoForgeVersion{mcMinor: mcMinor, full: v, parts: nums})
	}
	return out
}

func getNeoForgeVersionForMC(mcVersion string) (string, error) {
	parts := strings.Split(mcVersion, ".")
	if len(parts) < 2 || parts[0] != "1" {
		return "", fmt.Errorf("unsupported Minecraft version format: %s", mcVersion)
	}
	mcMinor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid Minecraft version: %s", mcVersion)
	}

	meta, err := fetchMavenMetadata(neoforgeMavenURL)
	if err != nil {
		return "", fmt.Errorf("NeoForge: %w", err)
	}

	parsed := parseNeoForgeVersions(meta.Versioning.Versions)

	var candidates []neoForgeVersion
	for _, v := range parsed {
		if v.mcMinor == mcMinor {
			candidates = append(candidates, v)
		}
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("no NeoForge version found for Minecraft %s", mcVersion)
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

	return candidates[0].full, nil
}

func DownloadNeoForge(mcVersion string) (string, error) {
	nfVersion, err := getNeoForgeVersionForMC(mcVersion)
	if err != nil {
		return "", err
	}

	jarURL := fmt.Sprintf("%s/%s/neoforge-%s-installer.jar", neoforgeMavenURL, nfVersion, nfVersion)
	installerFile := fmt.Sprintf("neoforge-%s-installer.jar", nfVersion)

	if err := downloadFile(jarURL, installerFile, "NeoForge "+nfVersion); err != nil {
		return "", err
	}
	return installerFile, nil
}

func InstallNeoForge(installerFile string) error {
	return installForgeLike(installerFile, "NeoForge")
}

func WriteUserJVMArgs(minRAM, maxRAM string) error {
	flags := make([]string, 0, len(DefaultJVMFlags)+2)
	flags = append(flags, "-Xms"+minRAM, "-Xmx"+maxRAM)
	flags = append(flags, DefaultJVMFlags...)
	content := strings.Join(flags, " ") + "\n"
	return os.WriteFile("user_jvm_args.txt", []byte(content), 0644)
}
