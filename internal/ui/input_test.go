package ui

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestInputFallback(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("hello\n"))

	got, err := inputFallback(in, &out, "Name:", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
	if !strings.Contains(out.String(), "Name:") {
		t.Errorf("expected prompt, got: %q", out.String())
	}
}

func TestInputFallbackWithDefault(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("\n"))

	got, err := inputFallback(in, &out, "Name:", "default-name")
	if err != nil {
		t.Fatal(err)
	}
	if got != "default-name" {
		t.Errorf("expected default, got %q", got)
	}
	if !strings.Contains(out.String(), "[default-name]") {
		t.Errorf("expected default in prompt, got: %q", out.String())
	}
}

func TestInputFallbackTrimWhitespace(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("  hello  \n"))

	got, err := inputFallback(in, &out, "Name:", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("expected trimmed hello, got %q", got)
	}
}

func TestInputNonTerminal(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("world\n"))

	got, err := Input(in, &out, "Message:", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "world" {
		t.Errorf("expected world, got %q", got)
	}
}

func TestConfirmFallback(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"y lowercase", "y\n", true},
		{"Y uppercase", "Y\n", true},
		{"yes", "yes\n", true},
		{"YES", "YES\n", true},
		{"n lowercase", "n\n", false},
		{"N uppercase", "N\n", false},
		{"empty", "\n", false},
		{"other text", "whatever\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := bufio.NewReader(strings.NewReader(tt.input))
			got, err := confirmFallback(in, &out, "Continue?")
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestConfirmFallbackPromptFormat(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("y\n"))

	_, err := confirmFallback(in, &out, "Continue?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "[y/N]") {
		t.Errorf("expected [y/N] in prompt, got: %q", out.String())
	}
}

func TestConfirmNonTerminal(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("y\n"))

	got, err := Confirm(in, &out, "Continue?")
	if err != nil {
		t.Fatal(err)
	}
	if got != true {
		t.Errorf("expected true, got %v", got)
	}
}
