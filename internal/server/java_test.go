package server

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestGetJavaVersionForMinecraft(t *testing.T) {
	tests := []struct {
		mcVersion string
		expected  int
	}{
		// 1.17 -> Java 16
		{"1.17", 16},
		{"1.17.1", 16},
		// 1.18 - 1.20.4 -> Java 17
		{"1.18", 17},
		{"1.18.2", 17},
		{"1.19.4", 17},
		{"1.20", 17},
		{"1.20.4", 17},
		// 1.20.5+ -> Java 21
		{"1.20.5", 21},
		{"1.20.6", 21},
		{"1.21", 21},
		{"1.21.1", 21},
		// Fallbacks & edge cases
		{"1.16.5", 16}, // Fallback for older
		{"21", 21},
		{"20", 17},
		{"17", 16},
		{"invalid", 21},
	}

	for _, tt := range tests {
		t.Run(tt.mcVersion, func(t *testing.T) {
			got := getJavaVersionForMinecraft(tt.mcVersion)
			if got != tt.expected {
				t.Errorf("getJavaVersionForMinecraft(%q) = %d; want %d", tt.mcVersion, got, tt.expected)
			}
		})
	}
}

func TestGetJavaVersionFromJar(t *testing.T) {
	tmpDir := t.TempDir()
	jarPath := filepath.Join(tmpDir, "test.jar")

	f, err := os.Create(jarPath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	zw := zip.NewWriter(f)

	w, err := zw.Create("Main.class")
	if err != nil {
		t.Fatalf("failed to create main class in zip: %v", err)
	}

	// Class version 61 (Java 17) -> CAFEBABE + 0000 (minor) + 003D (major = 61)
	classHeader := []byte{
		0xCA, 0xFE, 0xBA, 0xBE, // Magic number
		0x00, 0x00,             // Minor version
		0x00, 0x3D,             // Major version (61 -> Java 17)
	}

	if _, err := w.Write(classHeader); err != nil {
		t.Fatalf("failed to write header: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	f.Close()

	version, err := GetJavaVersionFromJar(jarPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if version != 17 {
		t.Errorf("expected Java version 17, got %d", version)
	}
}
