package logging

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// Component global logger (separate from the main global logger in logger.go)
var (
	componentGlobalLogger *Logger
	componentOnce         sync.Once
)

// initComponentGlobalLogger lazy-initializes the component global Zap logger
func initComponentGlobalLogger() {
	componentOnce.Do(func() {
		// Create default logger (console only, no file logging)
		cfg := DefaultConfig("eac", ".")
		if IsDebugEnabled() {
			cfg = cfg.WithDebugMode(true)
		}
		logger, err := New(cfg)
		if err != nil {
			// Fallback to basic logger if config fails
			logger, _ = NewDefault("eac", ".")
		}
		componentGlobalLogger = logger
	})
}

// SetTestLogger sets a custom zap logger for testing purposes.
// This replaces the global component logger. Returns a cleanup function
// that restores the previous logger.
// FOR TESTING ONLY - do not use in production code.
func SetTestLogger(zapLogger *zap.Logger) func() {
	oldLogger := componentGlobalLogger
	oldOnce := componentOnce

	// Reset the once so the test logger takes effect
	componentOnce = sync.Once{}
	componentGlobalLogger = &Logger{
		Logger: zapLogger,
		config: DefaultConfig("test", "."),
	}

	return func() {
		componentGlobalLogger = oldLogger
		componentOnce = oldOnce
	}
}

// ComponentLogger provides logging for a specific module/component.
// It is designed to be created once per package and reused.
//
// This is now backed by Zap Logger but maintains the simple API.
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
	component string // e.g., "commands/impl/build"
	// Note: We don't cache zap.Logger here because ConfigureLogging may reconfigure
	// the global logger after ComponentLogger instances are created.
	// Instead, we access componentGlobalLogger.Logger dynamically in each method.
}

// C creates a ComponentLogger, inferring the component name from the call site.
// If an explicit component name is provided, it uses that instead.
//
// The inferred name is derived from the caller's package path relative to go/eac/ or go/r2r/,
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
	initComponentGlobalLogger()

	comp := ""
	if len(component) > 0 && component[0] != "" {
		comp = component[0]
	} else {
		comp = inferComponent()
	}

	return &ComponentLogger{
		component: comp,
	}
}

// getZap returns the current zap logger.
// This is called on each log operation to ensure we use the current global logger
// (which may have been reconfigured via ConfigureLogging).
// Note: Component name is stored in ComponentLogger but not added to log output
// to avoid polluting console output with structured fields.
func (c *ComponentLogger) getZap() *zap.Logger {
	if componentGlobalLogger == nil {
		initComponentGlobalLogger()
	}
	return componentGlobalLogger.Logger
}

// Component creates a ComponentLogger, inferring from call site or using explicit name.
// This is the long form of C() - use when clarity is preferred.
func Component(component ...string) *ComponentLogger {
	return C(component...)
}

// inferComponent extracts the component path from the caller's package.
// Uses runtime reflection to get the full package path, then extracts
// the relative path from go/eac/ or go/r2r/ for unambiguous identification.
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
		// name looks like: github.com/ready-to-release/eac/go/eac/commands/impl/security/sast.init
		// or: github.com/ready-to-release/eac/go/r2r/cli/cmd.Execute

		// Find the /go/eac/ or /go/r2r/ boundary in the package path
		for _, boundary := range []string{"/go/eac/", "/go/r2r/"} {
			if idx := strings.Index(name, boundary); idx != -1 {
				// Extract everything after the boundary up to the function name
				pkgPath := name[idx+len(boundary):] // "commands/impl/security/sast.init"
				// Remove function name (after last dot)
				if dotIdx := strings.LastIndex(pkgPath, "."); dotIdx != -1 {
					pkgPath = pkgPath[:dotIdx] // "commands/impl/security/sast"
				}
				return pkgPath
			}
		}
	}

	// Fallback: extract from file path (for edge cases)
	file = filepath.ToSlash(file)

	// Look for /go/eac/ or /go/r2r/ in the file path
	for _, boundary := range []string{"/go/eac/", "/go/r2r/"} {
		if idx := strings.Index(file, boundary); idx != -1 {
			relPath := file[idx+len(boundary):] // "commands/impl/security/sast/sast.go"
			relPath = filepath.Dir(relPath)     // "commands/impl/security/sast"
			return filepath.ToSlash(relPath)
		}
	}

	// Ultimate fallback: directory name only
	return filepath.Base(filepath.Dir(file))
}

// Debug logs a debug message if debug logging is enabled.
func (c *ComponentLogger) Debug(msg string) {
	c.getZap().Debug(msg)
}

// Debugf logs a formatted debug message if debug logging is enabled.
func (c *ComponentLogger) Debugf(format string, args ...interface{}) {
	c.getZap().Sugar().Debugf(format, args...)
}

// Info logs an informational message.
func (c *ComponentLogger) Info(msg string) {
	c.getZap().Info(msg)
}

// Infof logs a formatted informational message.
func (c *ComponentLogger) Infof(format string, args ...interface{}) {
	c.getZap().Sugar().Infof(format, args...)
}

// Warn logs a warning message.
func (c *ComponentLogger) Warn(msg string) {
	c.getZap().Warn(msg)
}

// Warnf logs a formatted warning message.
func (c *ComponentLogger) Warnf(format string, args ...interface{}) {
	c.getZap().Sugar().Warnf(format, args...)
}

// Error logs an error message.
func (c *ComponentLogger) Error(msg string) {
	c.getZap().Error(msg)
}

// Errorf logs a formatted error message.
func (c *ComponentLogger) Errorf(format string, args ...interface{}) {
	c.getZap().Sugar().Errorf(format, args...)
}

// WithSuffix creates a new ComponentLogger with an additional suffix.
// Useful for sub-components or specific functions within a module.
//
// Example:
//
//	var log = logging.C("git")
//	funcLog := log.WithSuffix("StagedFiles")
//	funcLog.Debug("start")  // Component: "git.StagedFiles"
func (c *ComponentLogger) WithSuffix(suffix string) *ComponentLogger {
	newComp := c.component + "." + suffix
	return &ComponentLogger{
		component: newComp,
	}
}

// Sync flushes the component global logger.
// Call this at program exit to ensure all logs are written.
func Sync() error {
	if componentGlobalLogger != nil {
		return componentGlobalLogger.Sync()
	}
	return nil
}
