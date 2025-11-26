//go:build windows

package platform

// LineEnding is the platform-specific line ending for console output.
// Windows console requires \r\n for proper display.
const LineEnding = "\r\n"
