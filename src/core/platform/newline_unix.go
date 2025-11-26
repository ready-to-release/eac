//go:build !windows

package platform

// LineEnding is the platform-specific line ending for console output.
// Unix systems use \n.
const LineEnding = "\n"
