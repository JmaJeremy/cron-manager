package runlog

import (
	"bytes"
	"sync"
)

// MaxLineBytes is the longest line passed to emit; longer lines are split.
// syslog datagrams much larger than this are truncated or dropped.
const MaxLineBytes = 4096

// LineWriter is an io.Writer that calls emit once per complete line written to it,
// without the trailing newline. Call Flush after the last write to emit a final
// line that has no newline.
type LineWriter struct {
	emit func(line string)
	mu   sync.Mutex
	buf  []byte
}

// NewLineWriter returns a LineWriter that passes each line to emit.
func NewLineWriter(emit func(line string)) *LineWriter {
	return &LineWriter{emit: emit}
}

// Write implements io.Writer.
func (w *LineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		w.emitLocked(w.buf[:i])
		w.buf = w.buf[i+1:]
	}
	for len(w.buf) >= MaxLineBytes {
		w.emit(string(w.buf[:MaxLineBytes]))
		w.buf = w.buf[MaxLineBytes:]
	}
	return len(p), nil
}

// Flush emits any buffered partial line.
func (w *LineWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) > 0 {
		w.emitLocked(w.buf)
		w.buf = nil
	}
}

func (w *LineWriter) emitLocked(line []byte) {
	line = bytes.TrimSuffix(line, []byte("\r"))
	for len(line) > MaxLineBytes {
		w.emit(string(line[:MaxLineBytes]))
		line = line[MaxLineBytes:]
	}
	w.emit(string(line))
}
