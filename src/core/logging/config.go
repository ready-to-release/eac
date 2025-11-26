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
	// Module name used for log file path: out/logs/<module>/
	Module string

	// WorkspaceRoot is the repository root path
	WorkspaceRoot string

	// DebugMode enables Debug level output on console and file logging.
	// When false (default), Debug only goes to file, and no file is created.
	// When true, Debug appears on console and all levels are written to file.
	DebugMode bool

	// Development enables development mode (more verbose, stack traces)
	Development bool
}

// DefaultConfig returns configuration with default settings.
// Debug mode is disabled by default (no file logging, Debug hidden from console).
func DefaultConfig(module, workspaceRoot string) Config {
	return Config{
		Module:        module,
		WorkspaceRoot: workspaceRoot,
		DebugMode:     false,
		Development:   false,
	}
}

// WithDebugMode returns a new config with debug mode enabled.
func (c Config) WithDebugMode(enabled bool) Config {
	c.DebugMode = enabled
	return c
}
