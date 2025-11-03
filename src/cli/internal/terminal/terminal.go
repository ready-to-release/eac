package terminal

import (
	"os"
)

// GetWidth returns the current terminal width in columns
// Returns 80 as default if detection fails
func GetWidth() int {
	width, _, err := GetSize()
	if err != nil {
		return 80 // Default fallback
	}
	return width
}

// GetSize returns the terminal width and height
func GetSize() (width, height int, err error) {
	return getTerminalSize()
}

// IsTerminal returns true if stdout is a terminal
func IsTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
