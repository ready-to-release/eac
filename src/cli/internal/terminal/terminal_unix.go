// +build !windows,!aix,!plan9,!js,!nacl

package terminal

import (
	"golang.org/x/sys/unix"
)

// getTerminalSize implements terminal size detection for Unix-based systems
func getTerminalSize() (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(int(unix.Stdout), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(ws.Col), int(ws.Row), nil
}
