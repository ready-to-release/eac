package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Logger provides a simple logging interface for the r2r CLI.
// It outputs to stdout for info level and stderr for debug/warn/error.
type Logger struct {
	config LoggingConfig
	mu     sync.RWMutex
}

// Global state
var (
	globalLogger  *Logger
	loggerOnce    sync.Once
	debugEnabled  uint32 // atomic: 0 = disabled, 1 = enabled
	stdOutput     io.Writer = os.Stdout
	errOutput     io.Writer = os.Stderr
)

// Initialize creates the global logger with the given configuration.
// This should be called early in main() before any logging occurs.
// If not called, a default logger will be created on first use.
func Initialize(cfg LoggingConfig) {
	loggerOnce.Do(func() {
		globalLogger = &Logger{config: cfg}
	})
}

// InitFromEnv initializes the logger with default config and checks
// environment variables for debug mode.
func InitFromEnv() {
	loggerOnce.Do(func() {
		globalLogger = &Logger{config: DefaultConfig()}
	})

	// Check for debug mode via environment variable
	if os.Getenv("R2R_DEBUG") != "" || os.Getenv("R2R_LOG_LEVEL") == "debug" {
		EnableDebug()
	}
}

// get returns the global logger, initializing with defaults if needed
func get() *Logger {
	if globalLogger == nil {
		loggerOnce.Do(func() {
			globalLogger = &Logger{config: DefaultConfig()}
		})
	}
	return globalLogger
}

// EnableDebug turns on debug logging globally
func EnableDebug() {
	atomic.StoreUint32(&debugEnabled, 1)
	// Also enable debug level in console config
	if globalLogger != nil {
		globalLogger.mu.Lock()
		if !globalLogger.config.Console.HasLevel("debug") {
			globalLogger.config.Console.Levels = append([]string{"debug"}, globalLogger.config.Console.Levels...)
		}
		globalLogger.mu.Unlock()
	}
}

// DisableDebug turns off debug logging globally
func DisableDebug() {
	atomic.StoreUint32(&debugEnabled, 0)
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return atomic.LoadUint32(&debugEnabled) == 1
}

// SetLevel sets the minimum log level
func SetLevel(level string) error {
	switch level {
	case "debug":
		EnableDebug()
	case "info", "warn", "warning", "error":
		DisableDebug()
	default:
		return fmt.Errorf("unknown log level: %s", level)
	}
	return nil
}

// GetLevel returns the current log level as a string
func GetLevel() string {
	if IsDebugEnabled() {
		return "debug"
	}
	return "info"
}

// formatMessage formats a log message based on console config
func (l *Logger) formatMessage(level, msg string) string {
	l.mu.RLock()
	formatter := l.config.Console.Formatter
	l.mu.RUnlock()

	switch formatter {
	case FormatterRaw:
		return msg
	case FormatterJSON:
		data := map[string]string{
			"level":   level,
			"message": msg,
			"time":    time.Now().Format(time.RFC3339),
		}
		b, err := json.Marshal(data)
		if err != nil {
			return msg // Fallback to raw message on marshal error
		}
		return string(b)
	case FormatterTimestamped:
		fallthrough
	default:
		return fmt.Sprintf("%s  %-5s  %s", time.Now().Format("15:04:05.000"), level, msg)
	}
}

// shouldLog checks if a message at the given level should be logged
func (l *Logger) shouldLog(level string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.config.Console.HasLevel(level)
}

// Debug logs a debug message if debug logging is enabled.
// Debug has a fast-path check - if debug is disabled, it returns immediately.
func Debug(msg string) {
	if atomic.LoadUint32(&debugEnabled) == 0 {
		return
	}
	l := get()
	fmt.Fprintln(errOutput, l.formatMessage("DEBUG", msg))
}

// Debugf logs a formatted debug message if debug logging is enabled.
func Debugf(format string, args ...interface{}) {
	if atomic.LoadUint32(&debugEnabled) == 0 {
		return
	}
	l := get()
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(errOutput, l.formatMessage("DEBUG", msg))
}

// Info logs an informational message.
// Info goes to stdout (normal output).
func Info(msg string) {
	l := get()
	if !l.shouldLog("info") {
		return
	}
	fmt.Fprintln(stdOutput, l.formatMessage("INFO", msg))
}

// Infof logs a formatted informational message.
func Infof(format string, args ...interface{}) {
	l := get()
	if !l.shouldLog("info") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stdOutput, l.formatMessage("INFO", msg))
}

// Print outputs a message without a newline (raw output only).
// Useful for inline progress indicators like dots.
func Print(msg string) {
	l := get()
	if !l.shouldLog("info") {
		return
	}
	fmt.Fprint(stdOutput, msg)
}

// Printf outputs a formatted message without a newline (raw output only).
func Printf(format string, args ...interface{}) {
	l := get()
	if !l.shouldLog("info") {
		return
	}
	fmt.Fprintf(stdOutput, format, args...)
}

// Warn logs a warning message.
// Warn goes to stderr.
func Warn(msg string) {
	l := get()
	if !l.shouldLog("warn") {
		return
	}
	fmt.Fprintln(errOutput, l.formatMessage("WARN", msg))
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	l := get()
	if !l.shouldLog("warn") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(errOutput, l.formatMessage("WARN", msg))
}

// Error logs an error message.
// Error goes to stderr.
func Error(msg string) {
	l := get()
	if !l.shouldLog("error") {
		return
	}
	fmt.Fprintln(errOutput, l.formatMessage("ERROR", msg))
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	l := get()
	if !l.shouldLog("error") {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(errOutput, l.formatMessage("ERROR", msg))
}

// Fatal logs an error message and exits with status 1.
func Fatal(msg string) {
	Error(msg)
	os.Exit(1)
}

// Fatalf logs a formatted error message and exits with status 1.
func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
	os.Exit(1)
}

// SetOutput sets the output writers for testing purposes.
// stdOut receives Info messages, errOut receives Debug/Warn/Error messages.
func SetOutput(stdOut, errOut io.Writer) {
	stdOutput = stdOut
	errOutput = errOut
}

// ResetOutput resets output to default (stdout/stderr).
func ResetOutput() {
	stdOutput = os.Stdout
	errOutput = os.Stderr
}

// ResetForTesting resets the global logger state for testing.
// This allows tests to re-initialize the logger.
func ResetForTesting() {
	globalLogger = nil
	loggerOnce = sync.Once{}
	atomic.StoreUint32(&debugEnabled, 0)
	stdOutput = os.Stdout
	errOutput = os.Stderr
}
