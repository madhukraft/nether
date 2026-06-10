package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// overridable for testing
var stderr io.Writer = os.Stderr
var timeNow = time.Now

type ProgressReader struct {
	r        io.ReadCloser
	total    int64
	cur      int64
	label    string
	start    time.Time
	isTTY    bool
	width    int
	finished bool
}

func NewProgressReader(r io.ReadCloser, total int64, label string) *ProgressReader {
	pr := &ProgressReader{
		r:     r,
		total: total,
		label: label,
		start: timeNow(),
	}

	pr.isTTY, pr.width = IsTerminal(stderr)

	return pr
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.cur += int64(n)

	if pr.isTTY {
		pr.draw()
	}

	if err == io.EOF {
		pr.finished = true
		if pr.isTTY {
			pr.drawComplete()
		}
	}

	return n, err
}

func (pr *ProgressReader) Close() error {
	pr.isTTY = false

	if !pr.finished {
		fmt.Fprintf(stderr, "\r%s\r", strings.Repeat(" ", pr.width))
	}

	return pr.r.Close()
}

func (pr *ProgressReader) draw() {
	elapsed := timeNow().Sub(pr.start)
	if elapsed < 100*time.Millisecond {
		return
	}

	speed := float64(pr.cur) / elapsed.Seconds()

	if pr.total > 0 {
		pct := float64(pr.cur) / float64(pr.total) * 100
		fmt.Fprintf(stderr, "\r  %s  %s / %s  (%3.0f%%)  %s/s    ",
			pr.label,
			FormatBytes(pr.cur),
			FormatBytes(pr.total),
			pct,
			FormatBytes(int64(speed)),
		)
	} else {
		fmt.Fprintf(stderr, "\r  %s  %s @ %s/s    ",
			pr.label,
			FormatBytes(pr.cur),
			FormatBytes(int64(speed)),
		)
	}
}

func (pr *ProgressReader) drawComplete() {
	elapsed := timeNow().Sub(pr.start).Truncate(time.Millisecond * 100)
	fmt.Fprintf(stderr, "\r  => %s  (%s in %v)\n", pr.label, FormatBytes(pr.cur), elapsed)
}

func IsTerminal(w io.Writer) (bool, int) {
	width := 80
	if f, ok := w.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		if w, _, err := term.GetSize(int(f.Fd())); err == nil {
			width = w
		}
		return true, width
	}
	return false, width
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
