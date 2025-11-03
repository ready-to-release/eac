// +build windows

package terminal

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// GetConsoleSizeViaMode tries to get console size using the 'mode con' command
// This works even when stdout is redirected
func GetConsoleSizeViaMode() (width, height int, err error) {
	cmd := exec.Command("cmd", "/c", "mode", "con")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for "Columns:" (case-insensitive, handles different languages)
		lineLower := strings.ToLower(line)
		if strings.Contains(lineLower, "columns:") || strings.Contains(line, "Columns:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if w, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					width = w
				}
			}
		}

		// Look for "Lines:" (case-insensitive)
		if strings.Contains(lineLower, "lines:") || strings.Contains(line, "Lines:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if h, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
					height = h
				}
			}
		}
	}

	if width > 0 && height > 0 {
		return width, height, nil
	}

	return 0, 0, err
}
