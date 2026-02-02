// Package ansi provides utilities for handling ANSI escape sequences in output.
// Based on go/cli/r2r/internal/docker/ansi_filter.go with added configurability.
package ansi

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"sync"

	"github.com/MatusOllah/stripansi"
)

// FilterMode controls how ANSI sequences are handled.
type FilterMode int

const (
	// FilterBadOnly strips only problematic ANSI (default) - preserves colors.
	FilterBadOnly FilterMode = iota
	// FilterAll strips ALL ANSI sequences including colors.
	FilterAll
)

// Filter wraps an io.Writer and filters ANSI sequences based on mode.
// Uses buffering to handle sequences split across Write() boundaries.
//
// FilterBadOnly (default): Preserves colors, strips control sequences.
// FilterAll: Strips everything (uses github.com/MatusOllah/stripansi).
type Filter struct {
	w          io.Writer
	source     string
	mode       FilterMode
	buffer     bytes.Buffer
	warnedOnce bool
	mu         sync.Mutex
}

// Bad ANSI patterns - control sequences that cause problems.
// Colors (ESC[...m) are NOT included - they pass through.
var (
	// Cursor position reports (CPR) like ESC[61;1R - main issue on macOS.
	cursorPositionReportRegex = regexp.MustCompile(`\x1b\[\d+;\d+R`)

	// Device status report requests/responses that trigger CPR.
	deviceStatusRegex = regexp.MustCompile(`\x1b\[[56]n`)

	// Cursor position queries that trigger CPR responses.
	cursorQueryRegex = regexp.MustCompile(`\x1b\[\d*;?\d*[Hf]`)

	// Terminal title sequences (OSC sequences).
	titleSequenceRegex = regexp.MustCompile(`\x1b\]\d+;[^\x07]*\x07`)

	// Alternate screen buffer switches.
	altScreenRegex = regexp.MustCompile(`\x1b\[\?(?:1049|47|1047)[hl]`)

	// Mouse tracking sequences.
	mouseTrackingRegex = regexp.MustCompile(`\x1b\[\?100[0-6][hl]`)

	// Bracketed paste mode.
	bracketedPasteRegex = regexp.MustCompile(`\x1b\[\?2004[hl]`)

	// Cursor visibility changes.
	cursorVisibilityRegex = regexp.MustCompile(`\x1b\[\?25[hl]`)

	// Cursor save/restore.
	cursorSaveRestoreRegex = regexp.MustCompile(`\x1b[78]|\x1b\[[su]`)

	// DEC private modes (problematic terminal control).
	decPrivateModeRegex = regexp.MustCompile(`\x1b\[\?\d+[hl]`)

	// Terminal mode switches.
	terminalModeRegex = regexp.MustCompile(`\x1b[>=]`)

	// Combined pattern for quick detection of any bad ANSI.
	combinedBadAnsi = regexp.MustCompile(
		`\x1b\[\d+;\d+R` + // CPR
			`|\x1b\[[56]n` + // Device status
			`|\x1b\[\d*;?\d*[Hf]` + // Cursor queries
			`|\x1b\]\d+;[^\x07]*\x07` + // OSC/title
			`|\x1b\[\?(?:1049|47|1047)[hl]` + // Alt screen
			`|\x1b\[\?100[0-6][hl]` + // Mouse tracking
			`|\x1b\[\?2004[hl]` + // Bracketed paste
			`|\x1b\[\?25[hl]` + // Cursor visibility
			`|\x1b[78]|\x1b\[[su]` + // Cursor save/restore
			`|\x1b\[\?\d+[hl]` + // DEC private modes
			`|\x1b[>=]`) // Terminal mode
)

// badPatterns is the ordered list of patterns to strip.
var badPatterns = []*regexp.Regexp{
	cursorPositionReportRegex,
	deviceStatusRegex,
	cursorQueryRegex,
	titleSequenceRegex,
	altScreenRegex,
	mouseTrackingRegex,
	bracketedPasteRegex,
	cursorVisibilityRegex,
	cursorSaveRestoreRegex,
	decPrivateModeRegex,
	terminalModeRegex,
}

// NewFilter creates a filter with the specified mode.
// source identifies where output originates (for warning messages).
func NewFilter(w io.Writer, source string, mode FilterMode) *Filter {
	return &Filter{w: w, source: source, mode: mode}
}

// NewBadOnlyFilter creates a filter that strips only bad ANSI (default).
func NewBadOnlyFilter(w io.Writer, source string) *Filter {
	return NewFilter(w, source, FilterBadOnly)
}

// NewStripAllFilter creates a filter that strips ALL ANSI.
func NewStripAllFilter(w io.Writer, source string) *Filter {
	return NewFilter(w, source, FilterAll)
}

// Write filters ANSI escape sequences from p based on mode.
// Uses buffering to handle sequences split across Write() calls.
func (f *Filter) Write(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Add new data to buffer (handles split sequences)
	f.buffer.Write(p)
	data := f.buffer.Bytes()

	var output []byte

	switch f.mode {
	case FilterAll:
		// Strip ALL ANSI using external library
		output = stripansi.Bytes(data)
		if len(output) != len(data) && !f.warnedOnce {
			f.warnedOnce = true
			caller := getCallerInfo(3)
			msg := fmt.Sprintf("[NOTE] ANSI stripped (strip-all mode) from: %s [caller: %s]\n", f.source, caller)
			_, _ = f.w.Write([]byte(msg)) //nolint:errcheck
		}

	case FilterBadOnly:
		fallthrough
	default:
		// Strip only bad ANSI, preserve colors
		output = data
		hadBadAnsi := combinedBadAnsi.Match(data)
		for _, pattern := range badPatterns {
			output = pattern.ReplaceAll(output, nil)
		}

		if hadBadAnsi && !f.warnedOnce {
			f.warnedOnce = true
			caller := getCallerInfo(3)

			// Check if filtering worked
			if combinedBadAnsi.Match(output) {
				// WARNING: bad ANSI still present after filtering
				msg := fmt.Sprintf("[WARNING] Bad ANSI detected but could not fully filter from: %s [caller: %s]\n", f.source, caller)
				_, _ = f.w.Write([]byte(msg)) //nolint:errcheck
			} else {
				// NOTE: successfully filtered
				msg := fmt.Sprintf("[NOTE] Bad ANSI control sequences filtered from: %s [caller: %s]\n", f.source, caller)
				_, _ = f.w.Write([]byte(msg)) //nolint:errcheck
			}
		}
	}

	// Write filtered output
	_, err = f.w.Write(output)
	if err != nil {
		f.buffer.Reset()
		return 0, err
	}

	// Clear buffer after successful write
	f.buffer.Reset()

	// Return original length to satisfy io.Writer contract
	return len(p), nil
}

// getCallerInfo returns file:line of the caller at the given depth.
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	// Shorten to just filename, not full path
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			file = file[i+1:]
			break
		}
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// --- Utility functions ---

// StripBad removes only bad ANSI sequences, preserving colors.
func StripBad(data []byte) []byte {
	result := data
	for _, pattern := range badPatterns {
		result = pattern.ReplaceAll(result, nil)
	}
	return result
}

// StripAll removes ALL ANSI sequences (uses external library).
func StripAll(data []byte) []byte {
	return stripansi.Bytes(data)
}

// StripAllString removes ALL ANSI sequences from a string.
func StripAllString(s string) string {
	return stripansi.String(s)
}
