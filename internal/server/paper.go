package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const paperAPI = "https://api.papermc.io/v2/projects/paper"

type paperBuild struct {
    Build     int    `json:"build"`
    Channel   string `json:"channel"`
    Downloads struct {
        Application struct {
            Name string `json:"name"`
        } `json:"application"`
    } `json:"downloads"`
}

type paperBuildsResponse struct {
    Builds []paperBuild `json:"builds"`
}

type paperBuildResponse struct {
	Downloads struct {
		Application struct {
			Name string `json:"name"`
		} `json:"application"`
	} `json:"downloads"`
}

func getLatestPaperBuild(version string) (int, string, error) {
    url := fmt.Sprintf("%s/versions/%s/builds", paperAPI, version)
    resp, err := http.Get(url)
    if err != nil {
        return 0, "", fmt.Errorf("failed to reach Paper API: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == 404 {
        return 0, "", fmt.Errorf("version %s not found on Paper API", version)
    }

    var data paperBuildsResponse
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return 0, "", err
    }

    if len(data.Builds) == 0 {
        return 0, "", fmt.Errorf("no builds found for version %s", version)
    }

    // get latest stable build
    for i := len(data.Builds) - 1; i >= 0; i-- {
        if data.Builds[i].Channel == "STABLE" {
            b := data.Builds[i]
            return b.Build, b.Downloads.Application.Name, nil
        }
    }

    // fall back to latest if no stable found
    b := data.Builds[len(data.Builds)-1]
    return b.Build, b.Downloads.Application.Name, nil
}

func getJarName(version string, build int) (string, error) {
	url := fmt.Sprintf("%s/versions/%s/builds/%d", paperAPI, version, build)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data paperBuildResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Downloads.Application.Name, nil
}

func DownloadPaper(version string) error {
    fmt.Println("Fetching latest Paper build...")
    build, jarName, err := getLatestPaperBuild(version)
    if err != nil {
        return err
    }

    url := fmt.Sprintf("%s/versions/%s/builds/%d/downloads/%s", paperAPI, version, build, jarName)
    fmt.Printf("Downloading %s (build %d)...\n", jarName, build)

    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    out, err := os.Create("server.jar")
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}
