package logging

import "go.uber.org/zap/zapcore"

// Level represents log severity levels.
type Level = zapcore.Level

// Level constants for external use.
const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
)

// ChannelType identifies an output channel.
type ChannelType string

const (
	ConsoleChannel ChannelType = "console"
	FileChannel    ChannelType = "file"
)

// Config holds logger configuration.
type Config struct {
	// Command name (required) - first part of log path
	// Examples: "design", "build", "templates", "test"
	Command string

	// PathSegments (optional) - additional path components
	// Examples:
	//   [] → out/<command>/debug.log
	//   ["eac-core"] → out/<command>/eac-core/debug.log
	//   ["apply"] → out/<command>/apply/debug.log
	PathSegments []string

	// WorkspaceRoot is the repository root path
	WorkspaceRoot string

	// DebugMode enables Debug level output on console.
	// When false (default), Debug logs are hidden from console.
	// When true, Debug logs appear on console.
	DebugMode bool

	// EnableFileLogging enables writing logs to file.
	// When true, logs are written to out/<command>/<pathSegments>/debug.log.
	// When false, no file logging occurs.
	EnableFileLogging bool

	// Development enables development mode (more verbose, stack traces)
	Development bool
}

// DefaultConfig returns configuration with default settings.
// Debug mode and file logging are disabled by default.
// pathSegments is optional - omit for non-module-aware commands
func DefaultConfig(command, workspaceRoot string, pathSegments ...string) Config {
	return Config{
		Command:           command,
		PathSegments:      pathSegments,
		WorkspaceRoot:     workspaceRoot,
		DebugMode:         false,
		EnableFileLogging: false,
		Development:       false,
	}
}

// WithDebugMode returns a new config with debug mode enabled.
// This enables debug logs on console.
func (c Config) WithDebugMode(enabled bool) Config {
	c.DebugMode = enabled
	return c
}

// WithFileLogging returns a new config with file logging enabled.
// This enables writing all logs to file.
func (c Config) WithFileLogging(enabled bool) Config {
	c.EnableFileLogging = enabled
	return c
}
