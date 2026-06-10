package modrinth

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ModrinthIndex struct {
	FormatVersion int                    `json:"formatVersion"`
	Game          string                 `json:"game"`
	VersionID     string                 `json:"versionId"`
	Name          string                 `json:"name"`
	Summary       string                 `json:"summary,omitempty"`
	Files         []ModrinthIndexFile    `json:"files"`
	Dependencies  map[string]string      `json:"dependencies"`
}

type ModrinthIndexFile struct {
	Path      string            `json:"path"`
	Hashes    map[string]string `json:"hashes"`
	Env       map[string]string `json:"env,omitempty"`
	Downloads []string          `json:"downloads"`
	FileSize  int64             `json:"fileSize"`
}

func (c *Client) InstallModpack(input string) (*InstallResult, error) {
	slug := parseSlug(input)

	proj, err := c.GetProject(slug)
	if err != nil {
		return nil, fmt.Errorf("modpack %q not found on Modrinth", slug)
	}

	if proj.ProjectType != "modpack" {
		return nil, fmt.Errorf("%q is a %s, not a modpack", slug, proj.ProjectType)
	}

	versions, err := c.GetVersions(proj.ID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching versions for %q: %w", slug, err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for %q", slug)
	}

	ver := &versions[0]

	primaryFile := findPrimaryFile(ver.Files)
	if primaryFile == nil || !strings.HasSuffix(primaryFile.Filename, ".mrpack") {
		return nil, fmt.Errorf("no .mrpack file found for %q", slug)
	}

	return c.installMrpack(proj.Slug, proj.ID, ver.ID, ver.VersionNumber, primaryFile)
}

func (c *Client) installMrpack(slug, projectID, versionID, versionNumber string, file *VersionFile) (*InstallResult, error) {
	mrpackPath := filepath.Join(".nether", file.Filename)
	if err := os.MkdirAll(filepath.Dir(mrpackPath), 0755); err != nil {
		return nil, fmt.Errorf("creating temp directory: %w", err)
	}

	if err := downloadFile(file.URL, mrpackPath); err != nil {
		return nil, fmt.Errorf("downloading modpack: %w", err)
	}

	index, err := readMrpackIndex(mrpackPath)
	if err != nil {
		os.Remove(mrpackPath)
		return nil, fmt.Errorf("reading modpack: %w", err)
	}

	result := &InstallResult{
		Mod: InstalledMod{
			Slug:          slug,
			ProjectID:     projectID,
			VersionID:     versionID,
			VersionNumber: versionNumber,
		},
		Files: []string{},
		Deps:  []InstallResult{},
	}

	var serverFiles []ModrinthIndexFile
	for _, f := range index.Files {
		env := f.Env
		if env == nil {
			serverFiles = append(serverFiles, f)
			continue
		}
		side, ok := env["server"]
		if !ok || side == "required" || side == "optional" {
			serverFiles = append(serverFiles, f)
		}
	}

	for _, f := range serverFiles {
		destPath := filepath.Join(".", f.Path)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, fmt.Errorf("creating directory for %s: %w", f.Path, err)
		}

		downloaded := false
		for _, url := range f.Downloads {
			if err := downloadFile(url, destPath); err == nil {
				downloaded = true
				result.Files = append(result.Files, destPath)

				if sha1Hex, ok := f.Hashes["sha1"]; ok {
					if err := verifySha1(destPath, sha1Hex); err != nil {
						os.Remove(destPath)
						return nil, fmt.Errorf("checksum mismatch for %s: %w", f.Path, err)
					}
				}
				break
			}
		}
		if !downloaded {
			return nil, fmt.Errorf("failed to download %s from any mirror", f.Path)
		}
	}

	if err := extractOverrides(mrpackPath); err != nil {
		return nil, fmt.Errorf("extracting overrides: %w", err)
	}

	os.Remove(mrpackPath)

	return result, nil
}

func readMrpackIndex(mrpackPath string) (*ModrinthIndex, error) {
	r, err := zip.OpenReader(mrpackPath)
	if err != nil {
		return nil, fmt.Errorf("opening modpack: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "modrinth.index.json" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening index: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("reading index: %w", err)
			}

			var index ModrinthIndex
			if err := json.Unmarshal(data, &index); err != nil {
				return nil, fmt.Errorf("parsing index: %w", err)
			}
			return &index, nil
		}
	}

	return nil, fmt.Errorf("modrinth.index.json not found in modpack")
}

func extractOverrides(mrpackPath string) error {
	r, err := zip.OpenReader(mrpackPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "overrides/") {
			relPath := strings.TrimPrefix(f.Name, "overrides/")
			if relPath == "" {
				continue
			}

			destPath := filepath.Join(".", relPath)

			if f.FileInfo().IsDir() {
				os.MkdirAll(destPath, 0755)
				continue
			}

			os.MkdirAll(filepath.Dir(destPath), 0755)

			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("opening %s: %w", f.Name, err)
			}

			out, err := os.Create(destPath)
			if err != nil {
				rc.Close()
				return fmt.Errorf("creating %s: %w", destPath, err)
			}

			_, err = io.Copy(out, rc)
			rc.Close()
			out.Close()
			if err != nil {
				return fmt.Errorf("writing %s: %w", destPath, err)
			}
		}
	}

	return nil
}

func verifySha1(path, expectedHex string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	got := hex.EncodeToString(h.Sum(nil))
	if got != expectedHex {
		return fmt.Errorf("expected %s, got %s", expectedHex, got)
	}
	return nil
}
