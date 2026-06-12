package ui

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestWriteList(t *testing.T) {
	var buf bytes.Buffer
	options := []string{"paper", "vanilla", "fabric"}

	writeList(&buf, "Select type:", options, 0)
	output := buf.String()

	if !strings.Contains(output, "Select type:") {
		t.Errorf("missing title: %q", output)
	}
	if !strings.Contains(output, "> paper") {
		t.Errorf("missing selected marker on paper: %q", output)
	}
	if !strings.Contains(output, "vanilla") {
		t.Errorf("missing vanilla: %q", output)
	}
	if !strings.Contains(output, "fabric") {
		t.Errorf("missing fabric: %q", output)
	}
}

func TestWriteListSelected(t *testing.T) {
	var buf bytes.Buffer
	options := []string{"paper", "vanilla", "fabric"}

	writeList(&buf, "Select type:", options, 1)
	output := buf.String()

	if !strings.Contains(output, "> vanilla") {
		t.Errorf("expected vanilla selected, got: %q", output)
	}
	if strings.Contains(output, "> paper") {
		t.Errorf("paper should not be selected: %q", output)
	}
	if strings.Contains(output, "> fabric") {
		t.Errorf("fabric should not be selected: %q", output)
	}
}

func TestWriteListLinesCount(t *testing.T) {
	var buf bytes.Buffer
	options := []string{"a", "b", "c"}
	writeList(&buf, "title", options, 0)
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\r\n"), "\r\n")
	// Title + 3 options = 4 lines
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d: %q", len(lines), lines)
	}
}

func TestSelectFallbackEnterNumber(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("2\n"))
	options := []string{"paper", "vanilla", "fabric"}

	got, err := selectFallback(in, &out, "Select type:", options)
	if err != nil {
		t.Fatal(err)
	}
	if got != "vanilla" {
		t.Errorf("expected vanilla, got %q", got)
	}
}

func TestSelectFallbackEmptyDefaultsToFirst(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("\n"))
	options := []string{"paper", "vanilla", "fabric"}

	got, err := selectFallback(in, &out, "Select type:", options)
	if err != nil {
		t.Fatal(err)
	}
	if got != "paper" {
		t.Errorf("expected paper, got %q", got)
	}
}

func TestSelectFallbackInvalidThenValid(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("99\n1\n"))
	options := []string{"paper", "vanilla", "fabric"}

	got, err := selectFallback(in, &out, "Select type:", options)
	if err != nil {
		t.Fatal(err)
	}
	if got != "paper" {
		t.Errorf("expected paper, got %q", got)
	}
	output := out.String()
	if !strings.Contains(output, "Enter a number between 1 and 3") {
		t.Errorf("expected error message for invalid input, got: %q", output)
	}
}

func TestSelectNonTerminal(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader("2\n"))
	options := []string{"paper", "vanilla", "fabric"}

	got, err := Select(in, &out, "Select type:", options)
	if err != nil {
		t.Fatal(err)
	}
	if got != "vanilla" {
		t.Errorf("expected vanilla, got %q", got)
	}
	output := out.String()
	if !strings.Contains(output, "1) paper") {
		t.Errorf("expected numbered list, got: %q", output)
	}
}

func TestSelectEmpty(t *testing.T) {
	var out bytes.Buffer
	in := bufio.NewReader(strings.NewReader(""))
	_, err := Select(in, &out, "title", []string{})
	if err == nil {
		t.Error("expected error for empty options")
	}
}
