package server

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type mavenMetadata struct {
	XMLName    xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning struct {
		Latest   string   `xml:"latest"`
		Release  string   `xml:"release"`
		Versions []string `xml:"versions>version"`
	} `xml:"versioning"`
}

func fetchMavenMetadata(mavenURL string) (*mavenMetadata, error) {
	resp, err := http.Get(mavenURL + "/maven-metadata.xml")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maven metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from maven", resp.StatusCode)
	}

	var meta mavenMetadata
	if err := xml.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, fmt.Errorf("failed to parse maven metadata: %w", err)
	}

	return &meta, nil
}

func parseVersionParts(s string) ([]int, bool) {
	parts := strings.Split(s, ".")
	var nums []int
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		nums = append(nums, n)
	}
	if len(nums) == 0 {
		return nil, false
	}
	return nums, true
}

func installForgeLike(installerFile, label string) error {
	fmt.Printf("Running %s installer...\n", label)
	javaPath := "java"
	if bundledJavaExists() {
		javaPath = bundledJavaPath()
	}

	cmd := exec.Command(javaPath, "-jar", installerFile, "--installServer")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s installer failed: %w", label, err)
	}

	if err := os.Remove(installerFile); err != nil {
		return fmt.Errorf("failed to remove installer: %w", err)
	}

	if err := patchScriptJava("run.sh", 0755); err != nil {
		return fmt.Errorf("failed to fix run script: %w", err)
	}
	if err := patchScriptJava("run.bat", 0644); err != nil {
		return fmt.Errorf("failed to fix run.bat: %w", err)
	}

	return nil
}
