package tui

// Default TUI configuration values.
const (
	// DefaultHeight is the initial TUI height before WindowSizeMsg arrives.
	// In alt-screen mode, the actual terminal height is used instead.
	DefaultHeight = 40

	// DefaultBufferSize is the default line buffer capacity.
	DefaultBufferSize = 1000
)

// Config configures TUI console behavior.
// All fields have sensible defaults; zero values are valid.
type Config struct {
	// Display
	Height     int  // Total height (default: DefaultHeight)
	BufferSize int  // Line buffer size (default: DefaultBufferSize)
	ASCIIMode  bool // Use ASCII-only characters

	// Behavior
	SkipTUIDelay bool // Skip exit delay

	// Command Context
	RunPhaseName string // Custom name for Run phase (e.g., "building")
	CommandName  string // Name of the command being run
	CommandType  string // "build", "test", "lint", "scan", "interactive"
}

// WithDefaults returns a copy of Config with default values applied.
func (c Config) WithDefaults() Config {
	if c.Height <= 0 {
		c.Height = DefaultHeight
	}
	if c.BufferSize <= 0 {
		c.BufferSize = DefaultBufferSize
	}
	return c
}
