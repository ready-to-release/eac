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

	// DebugMode enables Debug level output on console.
	// When false (default), Debug logs are hidden from console.
	// When true, Debug logs appear on console.
	DebugMode bool

	// EnableFileLogging enables writing logs to file.
	// When true, logs are written to out/logs/<module>/debug.log.
	// When false, no file logging occurs.
	EnableFileLogging bool

	// Development enables development mode (more verbose, stack traces)
	Development bool
}

// DefaultConfig returns configuration with default settings.
// Debug mode and file logging are disabled by default.
func DefaultConfig(module, workspaceRoot string) Config {
	return Config{
		Module:            module,
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
