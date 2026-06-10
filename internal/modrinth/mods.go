package modrinth

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/madhukraft/nether/internal/ui"
)

type InstallResult struct {
	Mod    InstalledMod
	Files  []string
	Deps   []InstallResult
}

func (c *Client) slugToID(slug string) (string, error) {
	proj, err := c.GetProject(slug)
	if err != nil {
		return "", err
	}
	return proj.ID, nil
}

func mapServerTypeToLoader(serverType string) string {
	switch serverType {
	case "fabric":
		return "fabric"
	case "forge":
		return "forge"
	case "neoforge":
		return "neoforge"
	case "quilt":
		return "quilt"
	default:
		return ""
	}
}

func (c *Client) InstallMod(input, mcVersion, serverType string, autoDeps bool) (*InstallResult, error) {
	slug := parseSlug(input)

	proj, err := c.GetProject(slug)
	if err != nil {
		return nil, fmt.Errorf("project %q not found on Modrinth", slug)
	}

	if proj.ProjectType != "mod" {
		return nil, fmt.Errorf("%q is a %s, not a mod", slug, proj.ProjectType)
	}

	if proj.ServerSide == "unsupported" {
		return nil, fmt.Errorf("%q does not support servers", slug)
	}

	loader := mapServerTypeToLoader(serverType)
	if loader == "" {
		return nil, fmt.Errorf("server type %q does not support mods (use fabric, forge, neoforge, or quilt)", serverType)
	}

	versions, err := c.GetVersions(proj.ID, []string{loader}, []string{mcVersion})
	if err != nil {
		return nil, fmt.Errorf("fetching versions for %q: %w", slug, err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no version of %q found for Minecraft %s with %s", slug, mcVersion, loader)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].DatePublished > versions[j].DatePublished
	})

	ver := &versions[0]

	primaryFile := findPrimaryFile(ver.Files)
	if primaryFile == nil {
		return nil, fmt.Errorf("no downloadable file found for %s %s", slug, ver.VersionNumber)
	}

	installed := InstalledMod{
		Slug:          proj.Slug,
		ProjectID:     proj.ID,
		VersionID:     ver.ID,
		VersionNumber: ver.VersionNumber,
		FileName:      primaryFile.Filename,
	}

	result := &InstallResult{
		Mod:   installed,
		Files: []string{},
		Deps:  []InstallResult{},
	}

	destDir := "mods"
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("creating mods directory: %w", err)
	}

	filePath := filepath.Join(destDir, primaryFile.Filename)
	if err := downloadFile(primaryFile.URL, filePath); err != nil {
		return nil, fmt.Errorf("downloading %s: %w", primaryFile.Filename, err)
	}
	result.Files = append(result.Files, filePath)

	for _, dep := range ver.Dependencies {
		if dep.DependencyType == "incompatible" {
			return nil, fmt.Errorf("%q is incompatible with a dependency, aborting", slug)
		}
		if dep.DependencyType == "embedded" {
			continue
		}
		if dep.DependencyType == "optional" && !autoDeps {
			continue
		}
		if dep.ProjectID == "" {
			continue
		}
		if dep.DependencyType == "required" || (dep.DependencyType == "optional" && autoDeps) {
			depResult, err := c.installDependency(dep.ProjectID, mcVersion, loader)
			if err != nil {
				return nil, fmt.Errorf("installing dependency for %q: %w", slug, err)
			}
			if depResult != nil {
				result.Deps = append(result.Deps, *depResult)
			}
		}
	}

	return result, nil
}

func (c *Client) installDependency(projectID, mcVersion, loader string) (*InstallResult, error) {
	proj, err := c.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("dependency project %q not found", projectID)
	}

	if proj.ServerSide == "unsupported" {
		return nil, nil
	}

	versions, err := c.GetVersions(proj.ID, []string{loader}, []string{mcVersion})
	if err != nil || len(versions) == 0 {
		return nil, nil
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].DatePublished > versions[j].DatePublished
	})

	ver := &versions[0]
	primaryFile := findPrimaryFile(ver.Files)
	if primaryFile == nil {
		return nil, nil
	}

	installed := InstalledMod{
		Slug:          proj.Slug,
		ProjectID:     proj.ID,
		VersionID:     ver.ID,
		VersionNumber: ver.VersionNumber,
		FileName:      primaryFile.Filename,
	}

	result := &InstallResult{
		Mod:   installed,
		Files: []string{},
		Deps:  []InstallResult{},
	}

	filePath := filepath.Join("mods", primaryFile.Filename)
	if _, err := os.Stat(filePath); err != nil {
		if err := downloadFile(primaryFile.URL, filePath); err != nil {
			return nil, fmt.Errorf("downloading %s: %w", primaryFile.Filename, err)
		}
		result.Files = append(result.Files, filePath)
	}

	for _, dep := range ver.Dependencies {
		if dep.DependencyType != "required" || dep.ProjectID == "" {
			continue
		}
		sub, err := c.installDependency(dep.ProjectID, mcVersion, loader)
		if err != nil {
			continue
		}
		if sub != nil {
			result.Deps = append(result.Deps, *sub)
		}
	}

	return result, nil
}

func findPrimaryFile(files []VersionFile) *VersionFile {
	for _, f := range files {
		if f.Primary {
			return &f
		}
	}
	if len(files) > 0 {
		return &files[0]
	}
	return nil
}

func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	label := filepath.Base(destPath)
	body := ui.NewProgressReader(resp.Body, resp.ContentLength, label)
	defer body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, body)
	return err
}

func FindInstalledMod(slug string, installed []InstalledMod) *InstalledMod {
	for _, m := range installed {
		if m.Slug == slug || m.ProjectID == slug {
			return &m
		}
	}
	return nil
}

func RemoveMod(slug string, installed []InstalledMod) ([]InstalledMod, []string) {
	var remaining []InstalledMod
	var removedFiles []string
	for _, m := range installed {
		if m.Slug == slug || m.ProjectID == slug {
			removedFiles = append(removedFiles, filepath.Join("mods", m.Slug+"-*.jar"))
		} else {
			remaining = append(remaining, m)
		}
	}
	return remaining, removedFiles
}


