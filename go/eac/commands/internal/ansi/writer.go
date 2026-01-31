// Package ansi provides utilities for handling ANSI escape sequences in output.
package ansi

import (
	"fmt"
	"io"
	"regexp"
	"sync"
)

// escapeRegex matches ANSI escape sequences for stripping from log files.
// Matches sequences like: ESC[0m, ESC[31m, ESC[1;31m, etc.
var escapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StrippingWriter wraps an io.Writer and strips ANSI escape sequences
// from all data written through it. Use this to ensure log files remain
// clean and readable without terminal color codes.
//
// When ANSI sequences are detected and stripped, a warning is logged
// to help identify the source of the leak.
type StrippingWriter struct {
	w           io.Writer
	source      string
	warnedOnce  bool
	mu          sync.Mutex
}

// NewStrippingWriter creates a new writer that strips ANSI escape sequences.
// The source parameter identifies where output originates (for warning messages).
func NewStrippingWriter(w io.Writer, source string) *StrippingWriter {
	return &StrippingWriter{w: w, source: source}
}

// Write strips ANSI escape sequences from p and writes the result to the
// underlying writer. Returns the original length of p (not the stripped length)
// to satisfy io.Writer contract expectations from callers.
//
// On first detection of ANSI codes, logs a warning to help identify the leak source.
func (sw *StrippingWriter) Write(p []byte) (n int, err error) {
	// Check if data contains ANSI escape sequences
	if escapeRegex.Match(p) {
		sw.mu.Lock()
		if !sw.warnedOnce {
			sw.warnedOnce = true
			// Write warning before the stripped content
			warning := fmt.Sprintf("[WARNING] ANSI escape codes detected in log stream from: %s\n", sw.source)
			_, _ = sw.w.Write([]byte(warning))
		}
		sw.mu.Unlock()
	}

	stripped := escapeRegex.ReplaceAll(p, nil)
	_, err = sw.w.Write(stripped)
	if err != nil {
		return 0, err
	}
	// Return original length to satisfy caller expectations
	return len(p), nil
}

// Strip removes ANSI escape sequences from a byte slice.
func Strip(data []byte) []byte {
	return escapeRegex.ReplaceAll(data, nil)
}

// StripString removes ANSI escape sequences from a string.
func StripString(s string) string {
	return escapeRegex.ReplaceAllString(s, "")
}

// Contains checks if data contains ANSI escape sequences.
func Contains(data []byte) bool {
	return escapeRegex.Match(data)
}

// ContainsString checks if a string contains ANSI escape sequences.
func ContainsString(s string) bool {
	return escapeRegex.MatchString(s)
}
