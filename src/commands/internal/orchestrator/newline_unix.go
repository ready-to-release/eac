//go:build !windows

package orchestrator

// LineEnding is the platform-specific line ending for console output.
// Unix systems use \n.
const LineEnding = "\n"

// LineEndingPrefix is empty on Unix (log.Logger's \n is sufficient).
const LineEndingPrefix = ""
