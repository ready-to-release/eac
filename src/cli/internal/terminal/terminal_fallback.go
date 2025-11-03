// +build !windows,!linux,!darwin,!freebsd,!openbsd,!netbsd,!dragonfly,!solaris aix plan9 js nacl

package terminal

import "errors"

// Fallback for unsupported platforms
func getTerminalSize() (width, height int, err error) {
	return 80, 24, errors.New("terminal size detection not implemented for this platform")
}
