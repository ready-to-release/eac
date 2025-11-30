//go:build !windows

package platform

// WrapCommand returns the command unchanged on Unix systems.
func WrapCommand(name string, args ...string) (string, []string) {
	return name, args
}
