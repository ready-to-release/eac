package docker

import (
	"bytes"
	"io"
	"regexp"
)

// AnsiFilter wraps an io.Writer and filters out problematic ANSI sequences
// while preserving useful ones like colors and basic formatting.
//
// Preserved sequences:
// - SGR (Select Graphic Rendition) for colors and text formatting: ESC[...m
// - Basic cursor movement: ESC[A (up), ESC[B (down), ESC[C (forward), ESC[D (backward)
// - Line clearing: ESC[K, ESC[2K
// - Screen clearing: ESC[2J
//
// Filtered sequences:
// - Cursor position reports (CPR): ESC[row;colR
// - Device status reports: ESC[6n, ESC[5n
// - Cursor position queries: ESC[H, ESC[f
// - Terminal mode changes
// - Mouse tracking
// - Alternate screen buffer switches
// - Terminal title changes
// - Bracketed paste mode
// - Cursor visibility changes
type AnsiFilter struct {
	writer io.Writer
	buffer bytes.Buffer
}

// NewAnsiFilter creates a new ANSI filter that wraps the given writer
func NewAnsiFilter(w io.Writer) *AnsiFilter {
	return &AnsiFilter{writer: w}
}

var (
	// Sequences to filter out completely

	// Cursor position reports (CPR) like ESC[61;1R - this is the main issue on macOS
	cursorPositionReportRegex = regexp.MustCompile(`\x1b\[[0-9]+;[0-9]+R`)

	// Device status report requests/responses that trigger CPR
	deviceStatusRegex = regexp.MustCompile(`\x1b\[[56]n`)

	// Cursor position queries that trigger CPR responses
	cursorQueryRegex = regexp.MustCompile(`\x1b\[[0-9]*;?[0-9]*[Hf]`)

	// Terminal title sequences (OSC sequences)
	titleSequenceRegex = regexp.MustCompile(`\x1b\][0-9]+;[^\x07]*\x07`)

	// Alternate screen buffer switches
	altScreenRegex = regexp.MustCompile(`\x1b\[\?(?:1049|47|1047)[hl]`)

	// Mouse tracking sequences
	mouseTrackingRegex = regexp.MustCompile(`\x1b\[\?100[0-6][hl]`)

	// Bracketed paste mode
	bracketedPasteRegex = regexp.MustCompile(`\x1b\[\?2004[hl]`)

	// Cursor visibility changes
	cursorVisibilityRegex = regexp.MustCompile(`\x1b\[\?25[hl]`)

	// Cursor save/restore
	cursorSaveRestoreRegex = regexp.MustCompile(`\x1b[78]|\x1b\[[su]`)

	// DEC private modes that are problematic
	// This now catches ALL DEC private mode sequences like [?1h, [?1l, etc.
	decPrivateModeRegex = regexp.MustCompile(`\x1b\[\?[0-9]+[hl]`)
)

// Write filters ANSI escape sequences from the input
func (f *AnsiFilter) Write(p []byte) (n int, err error) {
	// Add new data to buffer
	f.buffer.Write(p)
	data := f.buffer.Bytes()

	// Apply filters in order - most specific first
	filtered := cursorPositionReportRegex.ReplaceAll(data, []byte{})
	filtered = deviceStatusRegex.ReplaceAll(filtered, []byte{})
	filtered = cursorQueryRegex.ReplaceAll(filtered, []byte{})
	filtered = titleSequenceRegex.ReplaceAll(filtered, []byte{})
	filtered = altScreenRegex.ReplaceAll(filtered, []byte{})
	filtered = mouseTrackingRegex.ReplaceAll(filtered, []byte{})
	filtered = bracketedPasteRegex.ReplaceAll(filtered, []byte{})
	filtered = cursorVisibilityRegex.ReplaceAll(filtered, []byte{})
	filtered = cursorSaveRestoreRegex.ReplaceAll(filtered, []byte{})
	filtered = decPrivateModeRegex.ReplaceAll(filtered, []byte{})

	// Be more selective with terminal mode to avoid removing SGR sequences
	// Only remove specific problematic modes, not all ? sequences
	filtered = regexp.MustCompile(`\x1b[>=]`).ReplaceAll(filtered, []byte{})

	// Write the filtered output
	_, err = f.writer.Write(filtered)
	if err != nil {
		return n, err
	}

	// Clear the buffer
	f.buffer.Reset()

	// Return original length to satisfy io.Writer contract
	return len(p), nil
}
