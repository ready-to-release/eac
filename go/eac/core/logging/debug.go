// Package logging provides a unified logging system for the EAC codebase.
//
// Debug Logging Design:
// - Global state for zero-cost disabled path (single bool check)
// - Module moniker identification for tracing
// - Consistent output format: "HH:MM:SS.mmm DEBUG module:message"
// - Thread-safe initialization
package logging

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ExecutionContext identifies where the CLI is running
type ExecutionContext string

const (
	// ContextImplicitCLI indicates running locally outside Docker
	ContextImplicitCLI ExecutionContext = "implicit-cli"
	// ContextR2RCLI indicates running inside Docker via r2r CLI
	ContextR2RCLI ExecutionContext = "r2r-cli"
)

// executionContext holds the detected execution context
var executionContext ExecutionContext

// originalCommand holds the original command line captured at init time
var originalCommand string

// contextOnce ensures context is detected only once
var contextOnce sync.Once

// contextLogged tracks if we've already logged the execution context
var contextLogged uint32

func init() {
	// Capture original command at package init, before any os.Args manipulation
	originalCommand = strings.Join(os.Args, " ")
}

// debugEnabled is the global debug state.
// Using atomic for thread-safe reads without locks.
// 0 = disabled, 1 = enabled
var debugEnabled uint32

// stdOutput is where Info messages are written (stdout by default).
// Can be changed for testing.
var stdOutput io.Writer = os.Stdout

// debugOutput is where Debug/Warn/Error messages are written (stderr by default).
// Can be changed for testing.
var debugOutput io.Writer = os.Stderr

// detectExecutionContext determines the execution context based on environment
func detectExecutionContext() {
	contextOnce.Do(func() {
		// Check explicit environment variable first
		if os.Getenv("R2R_DOCKER_MODE") == "true" {
			executionContext = ContextR2RCLI
			return
		}
		// Check for Docker container indicators
		if _, err := os.Stat("/.dockerenv"); err == nil {
			executionContext = ContextR2RCLI
			return
		}
		// Check if running from /app path (Docker container convention)
		if exe, err := os.Executable(); err == nil && len(exe) > 4 && exe[:4] == "/app" {
			executionContext = ContextR2RCLI
			return
		}
		executionContext = ContextImplicitCLI
	})
}

// GetExecutionContext returns the detected execution context
func GetExecutionContext() ExecutionContext {
	detectExecutionContext()
	return executionContext
}

// GetFullCommand returns the original command line as a string.
// This returns the command as it was at process startup, before any os.Args manipulation.
func GetFullCommand() string {
	return originalCommand
}

// LogExecutionContext emits a debug log with execution context info.
// This is called automatically on first logger creation, but only logs once.
func LogExecutionContext() {
	// Only log once across all logger creations
	if !atomic.CompareAndSwapUint32(&contextLogged, 0, 1) {
		return
	}

	detectExecutionContext()
	fullCommand := GetFullCommand()
	DebugDirectf("logging", "Execution context: %s. Command: \"%s\"", executionContext, fullCommand)
}

// EnableDebug turns on debug logging globally.
// This should be called once at application startup.
func EnableDebug() {
	atomic.StoreUint32(&debugEnabled, 1)
}

// DisableDebug turns off debug logging globally.
func DisableDebug() {
	atomic.StoreUint32(&debugEnabled, 0)
}

// IsDebugEnabled returns true if debug logging is enabled.
// This is a fast atomic read - safe to call frequently.
func IsDebugEnabled() bool {
	return atomic.LoadUint32(&debugEnabled) == 1
}

// InitFromEnv initializes debug state from EAC_DEBUG environment variable.
// Call this early in main() before any logging occurs.
func InitFromEnv() {
	if os.Getenv("EAC_DEBUG") != "" {
		EnableDebug()
	}
}

// debugTime returns the current time formatted for debug output.
// Format: "HH:MM:SS.mmm" (15:04:05.000)
func debugTime() string {
	return time.Now().Format("15:04:05.000")
}

// DebugDirect writes a debug message directly to stderr.
// This is the lowest-level debug function - use ComponentLogger for normal usage.
// Format: "HH:MM:SS.mmm  DEBUG  module:message"
func DebugDirect(module, msg string) {
	if atomic.LoadUint32(&debugEnabled) == 0 {
		return
	}
	fmt.Fprintf(debugOutput, "%s  DEBUG  %s:%s\n", debugTime(), module, msg)
}

// DebugDirectf writes a formatted debug message directly to stderr.
func DebugDirectf(module, format string, args ...interface{}) {
	if atomic.LoadUint32(&debugEnabled) == 0 {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(debugOutput, "%s  DEBUG  %s:%s\n", debugTime(), module, msg)
}
