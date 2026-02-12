package display

import core "github.com/ready-to-release/eac/contracts/core/0.1.0"

// Config configures TUI console behavior.
type Config struct {
	// Display
	Height     int  // Total height (default: DefaultHeight)
	BufferSize int  // Line buffer size (default: DefaultBufferSize)
	ASCIIMode  bool // Use ASCII-only characters
	TUI3Demo   bool // Use experimental tui3 layout

	// Behavior
	SkipTUIDelay bool // Skip exit delay

	// Command Context
	RunPhaseName string // Custom name for Run phase (e.g., "building")
	CommandName  string // Name of the command being run
	ActionType   core.ActionType // ActionBuild, ActionTest, ActionLint, ActionScan, etc.
}

// Default TUI configuration values.
const (
	// DefaultHeight is the initial TUI height before WindowSizeMsg arrives.
	DefaultHeight = 40

	// DefaultBufferSize is the default line buffer capacity.
	DefaultBufferSize = 1000
)
