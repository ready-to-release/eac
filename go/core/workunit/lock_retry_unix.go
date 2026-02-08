//go:build !windows

package workunit

// isWindowsAccessDenied always returns false on non-Windows platforms.
func isWindowsAccessDenied(_ error) bool {
	return false
}
