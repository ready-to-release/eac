//go:build windows

package platform

// WrapCommand returns the command unchanged on Windows.
// Go's exec.LookPath correctly uses PATHEXT to find .cmd/.bat files.
func WrapCommand(name string, args ...string) (string, []string) {
	return name, args
}
