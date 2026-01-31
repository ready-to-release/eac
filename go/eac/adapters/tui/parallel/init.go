package parallel

import (
	"github.com/ready-to-release/eac/go/eac/adapters/tui"
)

func init() {
	// Register parallel TUI for build, test, lint, scan commands.
	// These commands use the parallel task visualization TUI.
	factory := Factory()

	tui.Register("build", factory)
	tui.Register("test", factory)
	tui.Register("lint", factory)
	tui.Register("scan", factory)
}
