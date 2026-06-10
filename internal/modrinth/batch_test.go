package modrinth

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestBatchProgressNonTerminal(t *testing.T) {
	var stderr bytes.Buffer
	old := batchStderr
	batchStderr = &stderr
	defer func() { batchStderr = old }()

	bp := newBatchProgress("test-pack", 3)
	if bp.isTTY {
		t.Error("expected isTTY false when stderr is not a terminal")
	}

	// Simulate a skipped file
	bp.skipFile()

	// Simulate a downloaded file
	bp.nextFile("lithium.jar")
	data := []byte(strings.Repeat("x", 4096))
	rc := io.NopCloser(bytes.NewReader(data))
	wrapped := bp.wrapReader(rc, int64(len(data)))
	io.ReadAll(wrapped)
	wrapped.Close()

	// Simulate another downloaded file
	bp.nextFile("fabric-api.jar")
	rc2 := io.NopCloser(bytes.NewReader(data))
	wrapped2 := bp.wrapReader(rc2, int64(len(data)))
	io.ReadAll(wrapped2)
	wrapped2.Close()

	bp.done(2, 1, 8192)

	output := stderr.String()
	if !strings.Contains(output, "=>") {
		t.Errorf("expected done() to print summary, got: %q", output)
	}
	if !strings.Contains(output, "2 downloaded") {
		t.Errorf("expected '2 downloaded' in summary, got: %q", output)
	}
	if !strings.Contains(output, "1 up-to-date") {
		t.Errorf("expected '1 up-to-date' in summary, got: %q", output)
	}
}

func TestBatchProgressWrapReader(t *testing.T) {
	old := batchStderr
	var stderr bytes.Buffer
	batchStderr = &stderr
	defer func() { batchStderr = old }()

	bp := newBatchProgress("test", 1)
	bp.nextFile("file.jar")

	data := []byte("hello world")
	rc := io.NopCloser(bytes.NewReader(data))
	wrapped := bp.wrapReader(rc, int64(len(data)))

	got, err := io.ReadAll(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	wrapped.Close()

	if !bytes.Equal(got, data) {
		t.Errorf("expected %q, got %q", data, got)
	}
}
