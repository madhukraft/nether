package ui

import (
	"strings"
	"testing"
)

func TestColorizeDisabled(t *testing.T) {
	orig := colorEnabled
	colorEnabled = false
	defer func() { colorEnabled = orig }()

	result := Colorize("31", "hello")
	if result != "hello" {
		t.Errorf("expected plain text, got %q", result)
	}
}

func TestColorizeEnabled(t *testing.T) {
	orig := colorEnabled
	colorEnabled = true
	defer func() { colorEnabled = orig }()

	result := Colorize("31", "hello")
	expected := "\033[31mhello\033[0m"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestColorHelpers(t *testing.T) {
	orig := colorEnabled
	colorEnabled = true
	defer func() { colorEnabled = orig }()

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Green", Green("x"), "\033[32mx\033[0m"},
		{"Yellow", Yellow("x"), "\033[33mx\033[0m"},
		{"Red", Red("x"), "\033[31mx\033[0m"},
		{"Blue", Blue("x"), "\033[34mx\033[0m"},
		{"Cyan", Cyan("x"), "\033[36mx\033[0m"},
		{"Bold", Bold("x"), "\033[1mx\033[0m"},
		{"Dim", Dim("x"), "\033[2mx\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %q, want %q", tt.got, tt.expected)
			}
		})
	}
}

func TestColorHelpersDisabled(t *testing.T) {
	orig := colorEnabled
	colorEnabled = false
	defer func() { colorEnabled = orig }()

	if g := Green("x"); g != "x" {
		t.Errorf("expected plain x, got %q", g)
	}
	if g := Bold("x"); g != "x" {
		t.Errorf("expected plain x, got %q", g)
	}
}

func TestColorizeMultipleWords(t *testing.T) {
	orig := colorEnabled
	colorEnabled = true
	defer func() { colorEnabled = orig }()

	result := Red("error: something went wrong")
	if !strings.HasPrefix(result, "\033[31m") {
		t.Errorf("expected ANSI prefix, got %q", result)
	}
	if !strings.HasSuffix(result, "\033[0m") {
		t.Errorf("expected ANSI suffix, got %q", result)
	}
}

func TestColorizeEmpty(t *testing.T) {
	orig := colorEnabled
	colorEnabled = true
	defer func() { colorEnabled = orig }()

	result := Green("")
	if result != "\033[32m\033[0m" {
		t.Errorf("unexpected result for empty string: %q", result)
	}
}
