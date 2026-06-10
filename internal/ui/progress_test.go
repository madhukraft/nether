package ui

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type readCloser struct {
	io.Reader
	closed bool
}

func (rc *readCloser) Close() error {
	rc.closed = true
	return nil
}

func TestProgressReaderWrapsRead(t *testing.T) {
	data := []byte("hello world")
	rc := &readCloser{Reader: bytes.NewReader(data)}
	pr := NewProgressReader(rc, int64(len(data)), "test")
	defer pr.Close()

	got, err := io.ReadAll(pr)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("expected %q, got %q", data, got)
	}
}

func TestProgressReaderClosesUnderlying(t *testing.T) {
	rc := &readCloser{Reader: bytes.NewReader([]byte("data"))}
	pr := NewProgressReader(rc, 4, "test")
	io.ReadAll(pr)
	pr.Close()

	if !rc.closed {
		t.Error("expected underlying reader to be closed")
	}
}

func TestProgressReaderNonTerminalNoOutput(t *testing.T) {
	var buf bytes.Buffer
	old := stderr
	stderr = &buf
	defer func() { stderr = old }()

	data := []byte(strings.Repeat("x", 4096))
	rc := &readCloser{Reader: bytes.NewReader(data)}
	pr := NewProgressReader(rc, int64(len(data)), "test")
	defer pr.Close()

	io.ReadAll(pr)
	pr.Close()

	if buf.Len() > 0 {
		t.Errorf("expected no progress output on non-terminal, got %d bytes", buf.Len())
	}
}

func TestProgressReaderShortDownloadSkipsProgress(t *testing.T) {
	var buf bytes.Buffer
	old := stderr
	stderr = &buf
	defer func() { stderr = old }()

	pr := &ProgressReader{
		r:     &readCloser{Reader: bytes.NewReader([]byte("hi"))},
		total: 2,
		label: "tiny",
		start: timeNow(),
		isTTY: true,
		width: 80,
	}

	io.ReadAll(pr)
	pr.Close()

	output := buf.String()
	// completion line with => is OK, but no intermediate progress bar (with / and %)
	if strings.Contains(output, "/") && strings.Contains(output, "%") {
		t.Errorf("expected no intermediate progress for tiny download, got: %q", output)
	}
	if !strings.Contains(output, "=>") {
		t.Errorf("expected completion line for tiny download, got: %q", output)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestProgressReaderUnknownTotal(t *testing.T) {
	var buf bytes.Buffer
	old := stderr
	stderr = &buf
	defer func() { stderr = old }()

	rc := &readCloser{Reader: bytes.NewReader([]byte(strings.Repeat("x", 8192)))}
	pr := NewProgressReader(rc, -1, "unknown")
	defer pr.Close()

	io.ReadAll(pr)
	pr.Close()

	if buf.Len() > 0 && !strings.Contains(buf.String(), "@") {
		t.Errorf("expected speed indicator for unknown total, got: %q", buf.String())
	}
}
