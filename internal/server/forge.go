package server

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
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
		parts := strings.Split(buildStr, ".")
		var nums []int
		valid := true
		for _, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil {
				valid = false
				break
			}
			nums = append(nums, n)
		}
		if !valid || len(nums) == 0 {
			continue
		}
		out = append(out, forgeVersion{mcVersion: mcVer, build: buildStr, parts: nums})
	}
	return out
}

func getForgeVersionForMC(mcVersion string) (string, error) {
	resp, err := http.Get(forgeMavenURL + "/maven-metadata.xml")
	if err != nil {
		return "", fmt.Errorf("failed to fetch Forge metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from Forge maven", resp.StatusCode)
	}

	var meta mavenMetadata
	if err := xml.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", fmt.Errorf("failed to parse Forge metadata: %w", err)
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

	fmt.Printf("Downloading Forge %s...\n", fgVersion)

	resp, err := http.Get(jarURL)
	if err != nil {
		return "", fmt.Errorf("failed to download Forge installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from Forge maven", resp.StatusCode)
	}

	out, err := os.Create(installerFile)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return "", err
	}

	return installerFile, nil
}

func InstallForge(installerFile string) error {
	fmt.Println("Running Forge installer...")
	javaPath := "java"
	if bundledJavaExists() {
		javaPath = bundledJavaPath()
	}

	cmd := exec.Command(javaPath, "-jar", installerFile, "--installServer")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Forge installer failed: %w", err)
	}

	if err := os.Remove(installerFile); err != nil {
		return fmt.Errorf("failed to remove installer: %w", err)
	}

	if err := patchScriptJava("run.sh", 0755); err != nil {
		return fmt.Errorf("failed to fix run script: %w", err)
	}

	return nil
}
