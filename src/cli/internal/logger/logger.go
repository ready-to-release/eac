package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hitoshi44/go-uid64"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds logger configuration
type Config struct {
	Level       string
	File        string
	Console     bool
	JSON        bool
	Environment string
	RunID       string
	Minimal     bool // Minimal console output for CLI usage
}

// Logger wraps zerolog.Logger with additional context capabilities
type Logger struct {
	zerolog.Logger
	config Config
	mu     sync.RWMutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// Context keys for structured logging
	operationIDKey contextKey = "operation_id"
	commandKey     contextKey = "command"
	userKey        contextKey = "user"
	componentKey   contextKey = "component"
)

// LevelWriter filters log output based on the log level
type LevelWriter struct {
	io.Writer
	level zerolog.Level
}

// WriteLevel filters the log messages by level
func (lw *LevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= lw.level {
		return lw.Writer.Write(p)
	}
	return len(p), nil
}

// Initialize creates and configures the global logger instance
func Initialize(cfg Config) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(cfg)
		if err != nil {
			return
		}
		// Set global zerolog logger
		log.Logger = defaultLogger.Logger
	})
	return err
}

// New creates a new logger instance with the given configuration
func New(cfg Config) (*Logger, error) {
	// Set default values
	if cfg.RunID == "" {
		cfg.RunID, _ = uid64.NewString()
	}
	if cfg.Environment == "" {
		cfg.Environment = detectEnvironment()
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.File == "" {
		cfg.File = getLogFilePath()
	}

	// Parse log level
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Set global log level
	zerolog.SetGlobalLevel(level)

	// Configure time format based on environment
	if cfg.Environment == "production" || cfg.JSON {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
	}

	// Create writers
	writers := []io.Writer{}

	// File writer
	if cfg.File != "" {
		fileWriter, err := createFileWriter(cfg.File)
		if err != nil {
			return nil, fmt.Errorf("failed to create file writer: %w", err)
		}
		writers = append(writers, fileWriter)
	}

	// Console writer
	if cfg.Console {
		// Auto-detect minimal mode if not explicitly set
		if !cfg.Minimal && os.Getenv("R2R_VERBOSE_LOG") != "true" {
			cfg.Minimal = true
		}
		consoleWriter := createConsoleWriter(cfg, level)
		writers = append(writers, consoleWriter)
	}

	// Create multi-writer
	var writer io.Writer
	if len(writers) == 0 {
		// Default to console if no writers configured
		writer = createConsoleWriter(cfg, level)
	} else if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = zerolog.MultiLevelWriter(writers...)
	}

	// Create logger with standard fields
	zlog := zerolog.New(writer).With().
		Str("run_id", cfg.RunID).
		Str("environment", cfg.Environment).
		Str("version", getVersion()).
		Timestamp().
		Logger()

	return &Logger{
		Logger: zlog,
		config: cfg,
	}, nil
}

// Get returns the global logger instance
func Get() *Logger {
	if defaultLogger == nil {
		// Initialize with defaults if not already initialized
		Initialize(Config{Console: true})
	}
	return defaultLogger
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	// Add operation ID if present
	if opID := ctx.Value(operationIDKey); opID != nil {
		logger = logger.With().Str("operation_id", opID.(string)).Logger()
	}

	// Add command if present
	if cmd := ctx.Value(commandKey); cmd != nil {
		logger = logger.With().Str("command", cmd.(string)).Logger()
	}

	// Add user if present
	if user := ctx.Value(userKey); user != nil {
		logger = logger.With().Str("user", user.(string)).Logger()
	}

	// Add component if present
	if comp := ctx.Value(componentKey); comp != nil {
		logger = logger.With().Str("component", comp.(string)).Logger()
	}

	return &Logger{
		Logger: logger,
		config: l.config,
	}
}

// WithField adds a single field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.With().Interface(key, value).Logger(),
		config: l.config,
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.Logger
	for k, v := range fields {
		logger = logger.With().Interface(k, v).Logger()
	}
	return &Logger{
		Logger: logger,
		config: l.config,
	}
}

// SetLevel changes the logger's level at runtime
func (l *Logger) SetLevel(level string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	lvl, err := parseLevel(level)
	if err != nil {
		return err
	}

	l.config.Level = level
	zerolog.SetGlobalLevel(lvl)
	return nil
}

// Helper functions

func parseLevel(level string) (zerolog.Level, error) {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel, nil
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}

func detectEnvironment() string {
	// Check for CI environment
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		return "ci"
	}

	// Check for production indicators
	if os.Getenv("R2R_ENV") == "production" || os.Getenv("ENV") == "production" {
		return "production"
	}

	// Default to development
	return "development"
}

func createFileWriter(filename string) (io.Writer, error) {
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Open file for appending
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func createConsoleWriter(cfg Config, minLevel zerolog.Level) io.Writer {
	if cfg.JSON || cfg.Environment == "ci" || cfg.Environment == "production" {
		// Use plain JSON output for CI/production or when JSON is requested
		return &LevelWriter{Writer: os.Stdout, level: minLevel}
	}

	// Check if console output should be suppressed entirely
	if os.Getenv("R2R_NO_CONSOLE_LOG") == "true" {
		return &LevelWriter{Writer: io.Discard, level: minLevel}
	}

	// Use minimal console output for normal CLI usage
	if cfg.Minimal || (os.Getenv("R2R_VERBOSE_LOG") != "true" && minLevel >= zerolog.InfoLevel) {
		// Only show warnings and errors in minimal mode
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout, // Send to stdout in minimal mode to avoid shell error interpretation
			TimeFormat: "", // No timestamp for minimal output
			NoColor:    os.Getenv("NO_COLOR") != "",
			FormatLevel: func(i interface{}) string {
				if level, ok := i.(string); ok {
					switch level {
					case "warn":
						return "⚠️ "
					case "error":
						return "❌ "
					case "fatal":
						return "💀 "
					case "info":
						return "ℹ️ "
					default:
						return ""
					}
				}
				return ""
			},
			FormatMessage: func(i interface{}) string {
				if msg, ok := i.(string); ok {
					return msg
				}
				return fmt.Sprintf("%s", i)
			},
			FormatFieldName: func(i interface{}) string {
				// Only show error field in minimal mode
				if str, ok := i.(string); ok && str == "error" {
					return "" // We'll handle error in FormatFieldValue
				}
				return "" // Hide other field names for cleaner output
			},
			FormatFieldValue: func(i interface{}) string {
				return "" // Hide field values in minimal mode
			},
			FormatTimestamp: func(i interface{}) string {
				return "" // No timestamp for minimal output
			},
		}
		// Only show warnings and errors in minimal mode
		if minLevel < zerolog.WarnLevel {
			return &LevelWriter{Writer: consoleWriter, level: zerolog.WarnLevel}
		}
		return &LevelWriter{Writer: consoleWriter, level: minLevel}
	}

	// Use colored console writer for development/verbose mode
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr, // Send logs to stderr to keep stdout clean
		TimeFormat: time.Kitchen,
		NoColor:    os.Getenv("NO_COLOR") != "",
	}

	return &LevelWriter{Writer: consoleWriter, level: minLevel}
}

func getVersion() string {
	// This will be set by build flags
	return version
}

func getLogFilePath() string {
	// Try to find repository root by looking for .git directory first
	currentDir, err := os.Getwd()
	if err != nil {
		return "r2r.log" // fallback to current directory
	}

	// First pass: look for .git directory (most reliable indicator of repo root)
	dir := currentDir
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return filepath.Join(dir, "r2r.log")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory, couldn't find .git
			break
		}
		dir = parent
	}

	// Second pass: look for go.mod if no .git found (fallback)
	dir = currentDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return filepath.Join(dir, "r2r.log")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	// Fallback to current directory
	return "r2r.log"
}

// Context helper functions

// ContextWithOperationID adds an operation ID to the context
func ContextWithOperationID(ctx context.Context, opID string) context.Context {
	return context.WithValue(ctx, operationIDKey, opID)
}

// ContextWithCommand adds a command name to the context
func ContextWithCommand(ctx context.Context, cmd string) context.Context {
	return context.WithValue(ctx, commandKey, cmd)
}

// ContextWithUser adds a user to the context
func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// ContextWithComponent adds a component name to the context
func ContextWithComponent(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, componentKey, component)
}

// Package-level convenience functions that use the default logger

// Debug logs a debug message
func Debug(msg string) {
	Get().Debug().Msg(msg)
}

// Info logs an info message
func Info(msg string) {
	Get().Info().Msg(msg)
}

// Warn logs a warning message
func Warn(msg string) {
	Get().Warn().Msg(msg)
}

// Error logs an error message
func Error(msg string) {
	Get().Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string) {
	Get().Fatal().Msg(msg)
}

// WithContext returns a logger with context fields
func WithContext(ctx context.Context) *Logger {
	return Get().WithContext(ctx)
}

// WithField adds a field to the default logger
func WithField(key string, value interface{}) *Logger {
	return Get().WithField(key, value)
}

// WithFields adds multiple fields to the default logger
func WithFields(fields map[string]interface{}) *Logger {
	return Get().WithFields(fields)
}

// Variable to hold version information (set by build)
var version = "dev"
