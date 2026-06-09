package server

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func getJavaVersionForMinecraft(mcVersion string) int {
	parts := strings.Split(mcVersion, ".")
	if len(parts) == 0 {
		return 21
	}

	if parts[0] != "1" {
		// If it's a major number e.g. "21"
		v, err := strconv.Atoi(parts[0])
		if err == nil {
			if v >= 26 {
				return 25
			}
			if v >= 21 {
				return 21
			}
			if v == 20 {
				return 17
			}
			if v == 17 {
				return 16
			}
		}
		return 21
	}

	if len(parts) < 2 {
		return 21 // Default fallback
	}

	major, err := strconv.Atoi(parts[1])
	if err != nil {
		return 21
	}

	if major < 17 {
		return 16
	}
	if major == 17 {
		return 16
	}

	if major >= 26 {
		return 25
	}

	if major >= 21 && major <= 25 {
		return 21
	}

	if major >= 18 && major <= 20 {
		// Check for 1.20.5+
		if major == 20 && len(parts) >= 3 {
			minor, err := strconv.Atoi(parts[2])
			if err == nil && minor >= 5 {
				return 21
			}
		}
		return 17
	}

	// major >= 21
	return 21
}

func getAdoptiumOS() string {
	osVal := runtime.GOOS
	if osVal == "darwin" {
		return "mac"
	}
	return osVal
}

func getAdoptiumArch() string {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	case "386":
		return "x32"
	default:
		return arch
	}
}

func GetJavaVersionFromJar(jarPath string) (int, error) {
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open jar: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".class") {
			rc, err := f.Open()
			if err != nil {
				return 0, err
			}
			defer rc.Close()

			buf := make([]byte, 8)
			_, err = io.ReadFull(rc, buf)
			if err != nil {
				return 0, err
			}

			// Verify CAFEBABE magic number
			if binary.BigEndian.Uint32(buf[0:4]) != 0xCAFEBABE {
				continue
			}

			// Read major version (bytes 6 and 7)
			majorVersion := binary.BigEndian.Uint16(buf[6:8])
			javaVersion := int(majorVersion) - 44
			return javaVersion, nil
		}
	}
	return 0, fmt.Errorf("no class files found in jar")
}

func DownloadJava(mcVersion string) error {
	javaVer := getJavaVersionForMinecraft(mcVersion)
	fmt.Printf("Using Java %d\n", javaVer)

	osVal := getAdoptiumOS()
	archVal := getAdoptiumArch()

	// Attempt to download JRE first
	url := fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jre/hotspot/normal/eclipse", javaVer, osVal, archVal)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch JRE from Adoptium API: %w", err)
	}

	// Fallback to JDK if JRE is not found
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		url = fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%d/ga/%s/%s/jdk/hotspot/normal/eclipse", javaVer, osVal, archVal)
		resp, err = http.Get(url)
		if err != nil {
			return fmt.Errorf("failed to fetch JDK from Adoptium API: %w", err)
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Java: HTTP %d", resp.StatusCode)
	}

	tempFile := "java_download.tmp"
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary download file: %w", err)
	}
	defer func() {
		out.Close()
		os.Remove(tempFile)
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write Java download: %w", err)
	}
	out.Close()

	// Clear existing java directory
	if err := os.RemoveAll("java"); err != nil {
		return fmt.Errorf("failed to clear existing java directory: %w", err)
	}

	if err := os.MkdirAll("java", 0755); err != nil {
		return fmt.Errorf("failed to create java directory: %w", err)
	}

	archiveType, err := detectArchiveType(tempFile)
	if err != nil {
		return fmt.Errorf("failed to detect archive type: %w", err)
	}

	fmt.Println("Installing Java...")
	if archiveType == "tar.gz" {
		if err := extractTarGz(tempFile, "java"); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	} else if archiveType == "zip" {
		if err := extractZip(tempFile, "java"); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	}

	// Make sure the java binary is executable on Unix systems
	if runtime.GOOS != "windows" {
		javaBin := filepath.Join("java", "bin", "java")
		_ = os.Chmod(javaBin, 0755)
	}

	fmt.Println("Java setup complete.")
	return nil
}

func detectArchiveType(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 4)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	if n >= 2 && buf[0] == 0x1f && buf[1] == 0x8b {
		return "tar.gz", nil
	}
	if n >= 4 && buf[0] == 0x50 && buf[1] == 0x4b && buf[2] == 0x03 && buf[3] == 0x04 {
		return "zip", nil
	}
	return "", fmt.Errorf("unknown archive format")
}

func stripFirstComponent(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[1:], "/")
}

func extractTarGz(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := stripFirstComponent(header.Name)
		if targetPath == "" {
			continue
		}

		outPath := filepath.Join(destDir, targetPath)

		// Check for directory traversal
		cleanedDest := filepath.Clean(destDir)
		cleanedTarget := filepath.Clean(outPath)
		rel, err := filepath.Rel(cleanedDest, cleanedTarget)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("illegal file path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		targetPath := stripFirstComponent(f.Name)
		if targetPath == "" {
			continue
		}

		outPath := filepath.Join(destDir, targetPath)

		// Check for directory traversal
		cleanedDest := filepath.Clean(destDir)
		cleanedTarget := filepath.Clean(outPath)
		rel, err := filepath.Rel(cleanedDest, cleanedTarget)
		if err != nil || strings.HasPrefix(rel, "..") {
			return fmt.Errorf("illegal file path: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
