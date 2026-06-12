package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestPromptEula(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Accept lowercase y", "y\n", true},
		{"Accept uppercase Y", "Y\n", true},
		{"Accept lowercase y with space", "y \n", true},
		{"Accept yes", "yes\n", true},
		{"Reject lowercase n", "n\n", false},
		{"Reject uppercase N", "N\n", false},
		{"Reject empty input", "\n", false},
		{"Reject other text", "other\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bufio.NewReader(strings.NewReader(tt.input))
			var out bytes.Buffer
			accepted, err := promptEula(in, &out)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if accepted != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, accepted)
			}
			expectedPrompt := "Accept the Minecraft EULA (https://aka.ms/MinecraftEULA)? [y/N]: "
			if out.String() != expectedPrompt {
				t.Errorf("expected prompt %q, got %q", expectedPrompt, out.String())
			}
		})
	}
}

func TestParseRAM(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"1G", "1024M", false},
		{"2g", "2048M", false},
		{"2048M", "2048M", false},
		{"1024m", "1024M", false},
		{" 4g ", "4096M", false},
		{"4gb", "4096M", false},
		{"4GiB", "4096M", false},
		{"1024mb", "1024M", false},
		{"1024mib", "1024M", false},
		{"", "", true},
		{"g", "", true},
		{"m", "", true},
		{"-2g", "", true},
		{"0g", "", true},
		{"2", "", true},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseRAM(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRAM(%q) error = %v; wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("parseRAM(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPromptPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Default port", "\n", 25565},
		{"Custom port", "12345\n", 12345},
		{"Default with whitespace", "  \n", 25565},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bufio.NewReader(strings.NewReader(tt.input))
			var out bytes.Buffer
			port, err := promptPort(in, &out)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if port != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, port)
			}
		})
	}
}

