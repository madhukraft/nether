package modrinth

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/madhukraft/nether/internal/ui"
)

var batchStderr io.Writer = os.Stderr

type batchFileReader struct {
	bp    *batchProgress
	r     io.ReadCloser
	total int64
	cur   int64
}

func (r *batchFileReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.cur += int64(n)
	r.bp.advance(r.cur, r.total)
	return n, err
}

func (r *batchFileReader) Close() error {
	return r.r.Close()
}

type batchProgress struct {
	label   string
	total   int
	current int
	file    string
	start   time.Time
	isTTY   bool
	width   int
}

func newBatchProgress(label string, total int) *batchProgress {
	bp := &batchProgress{
		label: label,
		total: total,
		start: time.Now(),
	}

	bp.isTTY, bp.width = ui.IsTerminal(batchStderr)

	return bp
}

func (bp *batchProgress) nextFile(file string) {
	bp.current++
	bp.file = file
	if bp.isTTY {
		bp.draw(file, 0, 0)
	}
}

func (bp *batchProgress) skipFile() {
	bp.current++
}

func (bp *batchProgress) advance(cur, total int64) {
	if !bp.isTTY {
		return
	}
	elapsed := time.Since(bp.start)
	if elapsed < 100*time.Millisecond {
		return
	}
	bp.draw(bp.file, cur, total)
}

func (bp *batchProgress) wrapReader(r io.ReadCloser, total int64) io.ReadCloser {
	return &batchFileReader{bp: bp, r: r, total: total}
}

func (bp *batchProgress) done(downloaded, upToDate int, totalBytes int64) {
	elapsed := time.Since(bp.start).Truncate(time.Millisecond * 100)

	if bp.isTTY {
		fmt.Fprintf(batchStderr, "\r  => Modpack %q installed (%d downloaded, %d up-to-date, %s in %v)\n",
			bp.label, downloaded, upToDate, ui.FormatBytes(totalBytes), elapsed)
	} else {
		fmt.Fprintf(batchStderr, "  => Modpack %q installed (%d downloaded, %d up-to-date, %s in %v)\n",
			bp.label, downloaded, upToDate, ui.FormatBytes(totalBytes), elapsed)
	}
}

func (bp *batchProgress) draw(file string, cur, total int64) {
	var line string
	if cur > 0 && total > 0 {
		pct := float64(cur) / float64(total) * 100
		elapsed := time.Since(bp.start).Seconds()
		speed := float64(0)
		if elapsed > 0 {
			speed = float64(cur) / elapsed
		}
		line = fmt.Sprintf("  %s: %s  (%d/%d)  %s / %s  (%3.0f%%)  %s/s",
			bp.label,
			file,
			bp.current,
			bp.total,
			ui.FormatBytes(cur),
			ui.FormatBytes(total),
			pct,
			ui.FormatBytes(int64(speed)),
		)
	} else {
		line = fmt.Sprintf("  %s: %s  (%d/%d)", bp.label, file, bp.current, bp.total)
	}

	if len(line) > bp.width {
		line = line[:bp.width]
	}
	fmt.Fprintf(batchStderr, "\r%s    ", line)
}


