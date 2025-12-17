package logging

import (
	"io"
	"sync"

	"go.uber.org/zap"
)

// Logger wraps zap.Logger with dual-output support.
type Logger struct {
	*zap.Logger
	config  Config
	closers []io.Closer
}

// New creates a new Logger with the given configuration.
func New(cfg Config) (*Logger, error) {
	core, closers, err := buildCore(cfg)
	if err != nil {
		return nil, err
	}

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	zapLogger := zap.New(core, opts...)

	return &Logger{
		Logger:  zapLogger,
		config:  cfg,
		closers: closers,
	}, nil
}

// NewDefault creates a logger with file logging enabled.
// Logs go to: out/commands.log
func NewDefault(command, workspaceRoot string) (*Logger, error) {
	cfg := DefaultConfig(command, workspaceRoot).WithFileLogging(true)
	return New(cfg)
}

// NewForModule creates a logger for build/test with module-based logging.
// Logs go to: out/commands.log + target from logging.yml (if configured)
// Example for build: out/commands.log + out/build/eac-core/build.log
// Example for test: out/commands.log + out/test/eac-core/test.log
func NewForModule(command, workspaceRoot, module string) (*Logger, error) {
	cfg := DefaultConfig(command, workspaceRoot).
		WithModule(module).
		WithFileLogging(true)
	return New(cfg)
}

// NewWithDebug creates a logger with debug mode and file logging enabled.
// Debug logs appear on console in addition to file.
func NewWithDebug(command, workspaceRoot string) (*Logger, error) {
	cfg := DefaultConfig(command, workspaceRoot).
		WithDebugMode(true).
		WithFileLogging(true)
	return New(cfg)
}

// NewWithFileLogging creates a logger with file logging enabled but console debug disabled.
// This is useful for commands that want debug logs in files only.
func NewWithFileLogging(command, workspaceRoot string) (*Logger, error) {
	cfg := DefaultConfig(command, workspaceRoot).WithFileLogging(true)
	return New(cfg)
}

// Debug logs a debug message.
// Console: only shown when debug mode is enabled
// File: written when file logging is enabled
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs an info message.
// Console: always shown
// File: written when file logging is enabled
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a warning message.
// Console: always shown
// File: written when file logging is enabled
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message.
// Console: always shown
// File: written when file logging is enabled
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// LogDebugContent logs debug content to the logger instead of writing to a file.
// This replaces the pattern: os.WriteFile("out/logs/debug-*.md", content, 0644)
//
// The content is logged as a debug message with structured fields for filtering.
// Only logs when debug mode is enabled.
//
// Example:
//
//	if logger != nil && logger.IsDebugMode() {
//	    logger.LogDebugContent("full-prompt", promptText)
//	}
func (l *Logger) LogDebugContent(contentType, content string) {
	if l == nil || !l.IsDebugMode() {
		return
	}

	l.Debug("Debug content",
		zap.String("type", contentType),
		zap.String("content", content))
}

// LogDebugContentLines logs large content in chunks for readability.
// Each line is logged separately with a line number for easier viewing.
// Only logs when debug mode is enabled.
//
// Example:
//
//	if logger != nil && logger.IsDebugMode() {
//	    lines := strings.Split(largeContent, "\n")
//	    logger.LogDebugContentLines("analysis-output", lines)
//	}
func (l *Logger) LogDebugContentLines(contentType string, lines []string) {
	if l == nil || !l.IsDebugMode() {
		return
	}

	for i, line := range lines {
		l.Debug("Debug content line",
			zap.String("type", contentType),
			zap.Int("line", i+1),
			zap.String("content", line))
	}
}

// With creates a child logger with additional fields.
// Note: child loggers share the same closers with parent.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger:  l.Logger.With(fields...),
		config:  l.config,
		closers: nil, // closers are owned by parent logger
	}
}

// Sync flushes any buffered log entries and closes file handles.
func (l *Logger) Sync() error {
	err := l.Logger.Sync()
	for _, c := range l.closers {
		if c != nil {
			c.Close()
		}
	}
	return err
}

// IsDebugMode returns true if debug mode is enabled.
func (l *Logger) IsDebugMode() bool {
	return l.config.DebugMode
}

// --- Global Logger Support ---

var (
	globalLogger *Logger
	globalMu     sync.RWMutex
)

// Initialize sets up the global logger.
func Initialize(cfg Config) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}

	globalMu.Lock()
	globalLogger = logger
	globalMu.Unlock()

	return nil
}

// Get returns the global logger.
// Panics if Initialize has not been called.
func Get() *Logger {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if globalLogger == nil {
		panic("logging: Get called before Initialize")
	}
	return globalLogger
}

// L is a shorthand for Get().
func L() *Logger {
	return Get()
}
