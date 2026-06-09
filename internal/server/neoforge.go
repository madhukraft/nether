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

const neoforgeMavenURL = "https://maven.neoforged.net/releases/net/neoforged/neoforge"

type mavenMetadata struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning struct {
		Latest  string   `xml:"latest"`
		Release string   `xml:"release"`
		Versions []string `xml:"versions>version"`
	} `xml:"versioning"`
}

type neoForgeVersion struct {
	mcMinor int
	full    string
	parts   []int
}

func parseNeoForgeVersions(versions []string) []neoForgeVersion {
	var out []neoForgeVersion
	for _, v := range versions {
		// Skip beta/snapshot versions for clean auto-detection
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
		if !valid {
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

	resp, err := http.Get(neoforgeMavenURL + "/maven-metadata.xml")
	if err != nil {
		return "", fmt.Errorf("failed to fetch NeoForge metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from NeoForge maven", resp.StatusCode)
	}

	var meta mavenMetadata
	if err := xml.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return "", fmt.Errorf("failed to parse NeoForge metadata: %w", err)
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

	fmt.Printf("Downloading NeoForge %s...\n", nfVersion)

	resp, err := http.Get(jarURL)
	if err != nil {
		return "", fmt.Errorf("failed to download NeoForge installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d from NeoForge maven", resp.StatusCode)
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

func InstallNeoForge(installerFile string) error {
	fmt.Println("Running NeoForge installer...")
	javaPath := "java"
	if bundledJavaExists() {
		javaPath = bundledJavaPath()
	}

	cmd := exec.Command(javaPath, "-jar", installerFile, "--installServer")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("NeoForge installer failed: %w", err)
	}

	if err := os.Remove(installerFile); err != nil {
		return fmt.Errorf("failed to remove installer: %w", err)
	}

	if err := patchScriptJava("run.sh", 0755); err != nil {
		return fmt.Errorf("failed to fix run script: %w", err)
	}

	return nil
}

func WriteUserJVMArgs(minRAM, maxRAM string) error {
	flags := []string{
		"-Xms" + minRAM,
		"-Xmx" + maxRAM,
		"-XX:+AlwaysPreTouch",
		"-XX:+DisableExplicitGC",
		"-XX:+ParallelRefProcEnabled",
		"-XX:+PerfDisableSharedMem",
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:+UseG1GC",
		"-XX:G1HeapRegionSize=8M",
		"-XX:G1HeapWastePercent=5",
		"-XX:G1MaxNewSizePercent=40",
		"-XX:G1MixedGCCountTarget=4",
		"-XX:G1MixedGCLiveThresholdPercent=90",
		"-XX:G1NewSizePercent=30",
		"-XX:G1RSetUpdatingPauseTimePercent=5",
		"-XX:G1ReservePercent=20",
		"-XX:InitiatingHeapOccupancyPercent=15",
		"-XX:MaxGCPauseMillis=200",
		"-XX:MaxTenuringThreshold=1",
		"-XX:SurvivorRatio=32",
		"-Dusing.aikars.flags=https://mcflags.emc.gs",
		"-Daikars.new.flags=true",
	}
	content := strings.Join(flags, " ") + "\n"
	return os.WriteFile("user_jvm_args.txt", []byte(content), 0644)
}
