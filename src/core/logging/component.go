package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

// consoleConfig holds the lazily-loaded console configuration
var (
	consoleConfig     *SinkConfig
	consoleConfigOnce sync.Once
)

// getConsoleConfig lazily loads the console logging configuration
func getConsoleConfig() *SinkConfig {
	// If already set (e.g., by test), return it
	if consoleConfig != nil {
		return consoleConfig
	}

	consoleConfigOnce.Do(func() {
		// Try to find workspace root by looking for .r2r directory
		wd, err := os.Getwd()
		if err != nil {
			cfg := DefaultLoggingConfig()
			consoleConfig = &cfg.Console
			return
		}

		// Walk up to find .r2r/eac/logging.yml
		dir := wd
		for {
			configPath := filepath.Join(dir, ".r2r", "eac", "logging.yml")
			if _, err := os.Stat(configPath); err == nil {
				cfg := LoadLoggingConfig(dir)
				consoleConfig = &cfg.Console
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}

		// Fall back to defaults
		cfg := DefaultLoggingConfig()
		consoleConfig = &cfg.Console
	})
	return consoleConfig
}

// formatMessage formats a log message based on console config
func formatMessage(level, component, msg string) string {
	cfg := getConsoleConfig()

	switch cfg.Formatter {
	case FormatterRaw:
		return msg
	case FormatterJSON:
		// Simple JSON for console (rare use case)
		return fmt.Sprintf(`{"level":"%s","component":"%s","message":"%s"}`, level, component, msg)
	case FormatterTimestamped:
		fallthrough
	default:
		// Default to timestamped: "HH:MM:SS.mmm  LEVEL  component:message"
		return fmt.Sprintf("%s  %-5s  %s:%s", debugTime(), level, component, msg)
	}
}

// ComponentLogger provides logging for a specific module/component.
// It is designed to be created once per package and reused.
//
// Usage:
//
//	var log = logging.C()  // Infers component from call site (recommended)
//
//	func MyFunc() {
//	    log.Debug("starting operation")
//	    // ... work ...
//	    log.Debug("operation complete")
//	}
type ComponentLogger struct {
	// component is the inferred or explicit component path
	// (e.g., "commands/impl/security/sast", "core/logging")
	component string
}

// C creates a ComponentLogger, inferring the component name from the call site.
// If an explicit component name is provided, it uses that instead.
//
// The inferred name is derived from the caller's package path relative to src/,
// giving unique, unambiguous component identifiers like:
//   - "commands/impl/security/sast"
//   - "core/logging"
//   - "commands/impl/release"
//
// Example:
//
//	var log = logging.C()              // infers from call site (recommended)
//	var log = logging.C("custom-name") // explicit override
func C(component ...string) *ComponentLogger {
	if len(component) > 0 && component[0] != "" {
		return &ComponentLogger{component: component[0]}
	}
	return &ComponentLogger{component: inferComponent()}
}

// Component creates a ComponentLogger, inferring from call site or using explicit name.
// This is the long form of C() - use when clarity is preferred.
func Component(component ...string) *ComponentLogger {
	if len(component) > 0 && component[0] != "" {
		return &ComponentLogger{component: component[0]}
	}
	return &ComponentLogger{component: inferComponent()}
}

// inferComponent extracts the component path from the caller's package.
// Uses runtime reflection to get the full package path, then extracts
// the relative path from src/ for unambiguous identification.
//
// Call stack: runtime.Caller(0)=inferComponent, (1)=C/Component, (2)=caller
func inferComponent() string {
	// skip=2: skip inferComponent itself and C()/Component()
	pc, file, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	// Primary: extract from function/package path (most reliable)
	if fn := runtime.FuncForPC(pc); fn != nil {
		name := fn.Name()
		// name looks like: github.com/ready-to-release/eac/src/commands/impl/security/sast.init
		// or: github.com/ready-to-release/eac/src/commands/impl/security/sast.SAST

		// Find the /src/ boundary in the package path
		if idx := strings.Index(name, "/src/"); idx != -1 {
			// Extract everything after /src/ up to the function name
			pkgPath := name[idx+5:] // "commands/impl/security/sast.init"
			// Remove function name (after last dot)
			if dotIdx := strings.LastIndex(pkgPath, "."); dotIdx != -1 {
				pkgPath = pkgPath[:dotIdx] // "commands/impl/security/sast"
			}
			return pkgPath
		}
	}

	// Fallback: extract from file path (for edge cases)
	file = filepath.ToSlash(file)

	// Look for /src/ in the file path
	if idx := strings.Index(file, "/src/"); idx != -1 {
		relPath := file[idx+5:]               // "commands/impl/security/sast/sast.go"
		relPath = filepath.Dir(relPath)       // "commands/impl/security/sast"
		return filepath.ToSlash(relPath)
	}

	// Ultimate fallback: directory name only
	return filepath.Base(filepath.Dir(file))
}

// Debug logs a debug message if debug logging is enabled.
// This has a fast-path check - if debug is disabled, it returns immediately
// with only a single atomic load (no allocations, no formatting).
func (c *ComponentLogger) Debug(msg string) {
	if atomic.LoadUint32(&debugEnabled) == 0 {
		return
	}
	fmt.Fprintln(debugOutput, formatMessage("DEBUG", c.component, msg))
}

// Debugf logs a formatted debug message if debug logging is enabled.
// The format string and args are only evaluated if debug is enabled.
func (c *ComponentLogger) Debugf(format string, args ...interface{}) {
	if atomic.LoadUint32(&debugEnabled) == 0 {
		return
	}
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(debugOutput, formatMessage("DEBUG", c.component, msg))
}

// Info logs an informational message.
// Info messages are always shown (not controlled by debug flag).
// Info goes to stdout (normal output), while Debug/Warn/Error go to stderr.
func (c *ComponentLogger) Info(msg string) {
	fmt.Fprintln(stdOutput, formatMessage("INFO", c.component, msg))
}

// Infof logs a formatted informational message.
// Info goes to stdout (normal output), while Debug/Warn/Error go to stderr.
func (c *ComponentLogger) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(stdOutput, formatMessage("INFO", c.component, msg))
}

// Warn logs a warning message.
// Warning messages are always shown.
func (c *ComponentLogger) Warn(msg string) {
	fmt.Fprintln(debugOutput, formatMessage("WARN", c.component, msg))
}

// Warnf logs a formatted warning message.
func (c *ComponentLogger) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(debugOutput, formatMessage("WARN", c.component, msg))
}

// Error logs an error message.
// Error messages are always shown.
func (c *ComponentLogger) Error(msg string) {
	fmt.Fprintln(debugOutput, formatMessage("ERROR", c.component, msg))
}

// Errorf logs a formatted error message.
func (c *ComponentLogger) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(debugOutput, formatMessage("ERROR", c.component, msg))
}

// WithSuffix creates a new ComponentLogger with an additional suffix.
// Useful for sub-components or specific functions within a module.
//
// Example:
//
//	var log = logging.C("git")
//	funcLog := log.WithSuffix("StagedFiles")
//	funcLog.Debug("start")  // Output: "15:04:05.000  DEBUG  git.StagedFiles:start"
func (c *ComponentLogger) WithSuffix(suffix string) *ComponentLogger {
	return &ComponentLogger{component: c.component + "." + suffix}
}
