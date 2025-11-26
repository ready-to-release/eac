//go:build windows

package orchestrator

// LineEnding is the platform-specific line ending for console output.
// Windows console requires \r\n for proper display.
const LineEnding = "\r\n"

// LineEndingPrefix is \r on Windows, prepended before log.Logger's automatic \n.
const LineEndingPrefix = "\r"
