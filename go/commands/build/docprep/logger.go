package docprep

import (
	"fmt"
	"io"
)

// Logger wraps an io.Writer with leveled formatting.
type Logger struct {
	w io.Writer
}

// NewLogger creates a Logger that writes to w.
func NewLogger(w io.Writer) Logger { return Logger{w: w} }

// Infof writes an informational message.
func (l Logger) Infof(format string, args ...any) {
	fmt.Fprintf(l.w, format+"\n", args...)
}

// Warnf writes a warning message.
func (l Logger) Warnf(format string, args ...any) {
	fmt.Fprintf(l.w, "  Warning: "+format+"\n", args...)
}
