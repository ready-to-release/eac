package interfaces

// OutputBufferPort provides controlled stdout/stderr capture at the OS level.
// When active, captures all writes to stdout/stderr regardless of source.
// When inactive, writes pass through normally.
//
// This is designed for TUI mode where external tools, libraries, and workers
// may write directly to os.Stdout/os.Stderr, bypassing the observer pattern.
// The buffer intercepts these writes and holds them until the TUI exits.
type OutputBufferPort interface {
	// Start begins capturing stdout/stderr.
	// All writes are buffered until Flush() or Stop().
	// Returns an error if capture cannot be started.
	Start() error

	// Stop ends capture and restores original stdout/stderr.
	// Buffered content is flushed to real stdout/stderr.
	// Safe to call multiple times.
	Stop()

	// Flush writes buffered content to real stdout/stderr without stopping.
	// Useful for showing captured output during execution.
	Flush()

	// IsActive returns true if capture is currently active.
	IsActive() bool

	// GetBuffered returns current buffered content without flushing.
	// Returns copies of the stdout and stderr buffers.
	GetBuffered() (stdout, stderr []byte)
}
