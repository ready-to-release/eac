package defaulttui

import (
	"github.com/ready-to-release/eac/go/eac/commands/tui"
)

func init() {
	// Register default TUI as fallback for all unbound root commands.
	// This includes: update, get, show, work, validate, create, pipeline, release, etc.
	//
	// The default TUI is an interactive command picker:
	// - Shows available subcommands
	// - User navigates with j/k or arrows
	// - Enter selects and executes the subcommand
	//
	// Only activates when:
	// - Running in devbox (interactive terminal)
	// - No subcommand provided (bare "update" not "update lint")
	// - No --help flag
	tui.Register("*", Factory())
}
