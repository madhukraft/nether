package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/madhukraft/nether/internal/ui"
)

type httpError struct {
	StatusCode int
	URL        string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP %d from %s", e.StatusCode, e.URL)
}

func isNotFound(err error) bool {
	var he *httpError
	if ok := errors.As(err, &he); ok {
		return he.StatusCode == http.StatusNotFound
	}
	return false
}

// downloadFile downloads url to destPath with a progress indicator.
// If label is empty, destPath is used as the progress label.
func downloadFile(url, destPath, label string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return &httpError{StatusCode: resp.StatusCode, URL: url}
	}

	if label == "" {
		label = destPath
	}

	body := ui.NewProgressReader(resp.Body, resp.ContentLength, label)
	defer body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, body); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	return nil
}

// downloadFileToTemp downloads url to a temporary file and returns its path.
// Caller is responsible for removing the temp file.
func downloadFileToTemp(url, label string) (string, error) {
	tmpFile, err := os.CreateTemp("", "nether-download-*")
	if err != nil {
		return "", fmt.Errorf("temp file failed: %w", err)
	}
	tmpPath := tmpFile.Name()

	resp, err := http.Get(url)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", &httpError{StatusCode: resp.StatusCode, URL: url}
	}

	body := ui.NewProgressReader(resp.Body, resp.ContentLength, label)
	defer body.Close()

	if _, err := io.Copy(tmpFile, body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("download failed: %w", err)
	}

	tmpFile.Close()
	return tmpPath, nil
}
