package cmd

import (
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
		{"Reject lowercase n", "n\n", false},
		{"Reject uppercase N", "N\n", false},
		{"Reject empty input", "\n", false},
		{"Reject other text", "yes\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := strings.NewReader(tt.input)
			var out bytes.Buffer
			accepted, err := promptEula(in, &out)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if accepted != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, accepted)
			}
			expectedPrompt := "Do you accept the Minecraft EULA (https://aka.ms/MinecraftEULA)? [y/N]: "
			if out.String() != expectedPrompt {
				t.Errorf("expected prompt %q, got %q", expectedPrompt, out.String())
			}
		})
	}
}
