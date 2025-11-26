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

// NewDefault creates a logger with default configuration (debug mode disabled).
func NewDefault(module, workspaceRoot string) (*Logger, error) {
	return New(DefaultConfig(module, workspaceRoot))
}

// NewWithDebug creates a logger with debug mode enabled.
func NewWithDebug(module, workspaceRoot string) (*Logger, error) {
	cfg := DefaultConfig(module, workspaceRoot).WithDebugMode(true)
	return New(cfg)
}

// Debug logs a debug message.
// Console: only shown when debug mode is enabled
// File: only written when debug mode is enabled
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs an info message.
// Console: always shown
// File: only written when debug mode is enabled
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a warning message.
// Console: always shown
// File: only written when debug mode is enabled
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message.
// Console: always shown
// File: only written when debug mode is enabled
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
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
