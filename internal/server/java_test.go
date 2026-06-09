package server

import (
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
