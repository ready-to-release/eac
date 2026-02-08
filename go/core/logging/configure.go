// Package logging - Unified logging configuration
//
// This file provides a single entry point for configuring the logging system.
// It replaces the scattered EnableFileLogging, EnableTUIForComponentLogger, etc.
package logging

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/core/paths"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// silentLumberjackWriter wraps lumberjack.Logger to suppress rotation error messages.
// Lumberjack uses the standard log package to print errors when it can't rotate files,
// which happens during concurrent builds when multiple processes access the same log file.
// This wrapper temporarily silences the standard logger during writes.
type silentLumberjackWriter struct {
	*lumberjack.Logger
	mu sync.Mutex
}

// Write implements io.Writer, suppressing lumberjack's rotation error messages.
func (s *silentLumberjackWriter) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Temporarily discard standard log output to suppress lumberjack's error messages
	// about failed log rotation (e.g., "can't rename log file: file in use")
	originalOutput := stdlog.Writer()
	stdlog.SetOutput(io.Discard)
	defer stdlog.SetOutput(originalOutput)

	return s.Logger.Write(p)
}

// Close implements io.Closer.
func (s *silentLumberjackWriter) Close() error {
	return s.Logger.Close()
}

// loggingState holds the current logging configuration state.
var (
	loggingClosers []io.Closer
	loggingMu      sync.Mutex
)

// ConfigureLogging sets up the logging system for a command.
//
// This is the ONE function to call for logging setup. It configures:
//   - File logging: All logs go to unified out/commands.log
//   - Target logging: For build/test, additional logs go to out/build/<module>/build.log or out/test/<module>/test.log
//   - Console logging: Info/Warn/Error always, Debug only if debugToConsole=true
//   - TUI logging: If tuiWriter provided, logs also go to TUI pane
//
// Parameters:
//   - workspaceRoot: repository root for log file location
//   - command: command name (e.g., "build", "test", "design")
//   - pathSegments: optional path segments. For build/test, first segment is the module name
//   - debugToConsole: if true, debug logs also appear on console (use with --debug flag)
//   - tuiWriter: if non-nil, logs also go to TUI pane (pass nil for non-TUI mode)
//
// Example usage in a command:
//
//	var tuiWriter io.Writer
//	if useTUI {
//	    tuiWriter = orch.GetTUIWriter(tui.PhaseInit)
//	}
//	// Any command: logs to out/commands.log
//	if err := logging.ConfigureLogging(workspaceRoot, "design", nil, debugMode, tuiWriter); err != nil {
//	    log.Warnf("Failed to configure logging: %v", err)
//	}
//	// Build with module: logs to out/commands.log + out/build/core/build.log
//	if err := logging.ConfigureLogging(workspaceRoot, "build", []string{"core"}, debugMode, tuiWriter); err != nil {
//	    log.Warnf("Failed to configure logging: %v", err)
//	}
//	// Test with module: logs to out/commands.log + out/test/core/test.log
//	if err := logging.ConfigureLogging(workspaceRoot, "test", []string{"core"}, debugMode, tuiWriter); err != nil {
//	    log.Warnf("Failed to configure logging: %v", err)
//	}
//	defer logging.CloseLogging()
func ConfigureLogging(workspaceRoot, command string, pathSegments []string, debugToConsole bool, tuiWriter io.Writer) error {
	loggingMu.Lock()
	defer loggingMu.Unlock()

	// Close any existing logging resources
	closeLoggingLocked()

	var cores []zapcore.Core
	var closers []io.Closer

	// Console encoder config - minimal output: message only (no prefix)
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        zapcore.OmitKey,
		LevelKey:       zapcore.OmitKey,
		NameKey:        zapcore.OmitKey,
		CallerKey:      zapcore.OmitKey,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	// 1. Console core (always present)
	// Info/Warn/Error always go to console
	// Debug goes to console only if debugToConsole=true
	consoleLevel := zapcore.InfoLevel
	if debugToConsole {
		consoleLevel = zapcore.DebugLevel
	}
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stderr),
		consoleLevel,
	)
	cores = append(cores, consoleCore)

	// 2. File logging cores
	// CLIE_TEST_LOGGING_ACTIVE=true disables unified log but allows target logs
	testLoggingActive := os.Getenv("CLIE_TEST_LOGGING_ACTIVE") == "true"

	if workspaceRoot != "" && command != "" {
		// Ensure out/ directory exists
		if err := os.MkdirAll(filepath.Join(workspaceRoot, paths.OutDir), 0o755); err != nil { //nolint:gosec // G301: Log directory should be world-readable
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Load logging config for rolling settings
		logCfg := LoadLoggingConfig(workspaceRoot)

		// 2a. Unified log: out/commands.log (disabled when test logging active)
		if !testLoggingActive {
			logPath := paths.CommandsLogPath(workspaceRoot)
			writer := &silentLumberjackWriter{
				Logger: &lumberjack.Logger{
					Filename:   logPath,
					MaxSize:    logCfg.File.MaxSizeMB,
					MaxBackups: logCfg.File.MaxBackups,
					MaxAge:     logCfg.File.MaxAgeDays,
					Compress:   logCfg.File.Compress != nil && *logCfg.File.Compress,
				},
			}
			closers = append(closers, writer)

			fileEncoder := CreateEncoder(logCfg.File.Formatter, command)
			fileCore := zapcore.NewCore(
				fileEncoder,
				zapcore.AddSync(writer),
				zapcore.DebugLevel, // File ALWAYS gets debug level
			)
			cores = append(cores, fileCore)
		}

		// 2b. Target file core (for build/test with module) - always enabled
		if target, ok := logCfg.GetTarget(command); ok && len(pathSegments) > 0 {
			module := pathSegments[0] // First path segment is the module
			targetPath := target.ResolveTargetPath(workspaceRoot, module)
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err == nil { //nolint:gosec // G301: Log directory should be world-readable
				targetFile, err := os.OpenFile(targetPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:gosec // G302: Log files should be world-readable
				if err == nil {
					closers = append(closers, targetFile)

					targetEncoder := CreateEncoder(target.Formatter, command)
					targetCore := zapcore.NewCore(
						targetEncoder,
						zapcore.AddSync(targetFile),
						zapcore.DebugLevel,
					)
					cores = append(cores, targetCore)
				}
			}
		}
	}

	// 3. TUI core (if TUI writer provided)
	// Logs go to TUI pane for live display
	if tuiWriter != nil {
		tuiLevel := zapcore.InfoLevel
		if debugToConsole {
			tuiLevel = zapcore.DebugLevel
		}
		tuiCore := NewTUICore(tuiWriter, tuiLevel)
		cores = append(cores, tuiCore)
	}

	// Combine all cores
	combinedCore := zapcore.NewTee(cores...)

	// Create new logger with combined core
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1), // Skip the ComponentLogger wrapper
		// Suppress zap's internal error output (e.g., write failures during log rotation)
		zap.ErrorOutput(zapcore.AddSync(io.Discard)),
	}

	newZapLogger := zap.New(combinedCore, opts...)

	// Replace the global component logger
	// This affects all existing ComponentLogger instances since they share the global
	componentGlobalLogger = &Logger{
		Logger:  newZapLogger,
		config:  DefaultConfig("eac", workspaceRoot),
		closers: closers,
	}

	// Mark the component logger as initialized to prevent initComponentGlobalLogger from overwriting it
	componentOnce.Do(func() {
		// Empty - just marks the once as done so initComponentGlobalLogger won't run
	})

	// Store closers for CloseLogging
	loggingClosers = closers

	// Only enable debug flag if debugToConsole is true
	// This ensures IsDebugEnabled() reflects the actual console debug setting
	if debugToConsole {
		EnableDebug()
	}

	return nil
}

// CloseLogging closes any open log files.
// Call this with defer after ConfigureLogging.
func CloseLogging() {
	loggingMu.Lock()
	defer loggingMu.Unlock()
	closeLoggingLocked()
}

// closeLoggingLocked closes logging resources (must hold loggingMu).
func closeLoggingLocked() {
	// Sync the zap logger to flush any buffered logs
	// Error is intentionally ignored as this is best-effort cleanup during shutdown
	if componentGlobalLogger != nil {
		if err := componentGlobalLogger.Logger.Sync(); err != nil {
			// Best-effort cleanup - nothing meaningful to do with error during shutdown
		}
	}

	for _, closer := range loggingClosers {
		closer.Close()
	}
	loggingClosers = nil
}

// ConfigureLoggingSimple is a convenience wrapper for non-TUI commands.
// Equivalent to ConfigureLogging(workspaceRoot, command, pathSegments, debugToConsole, nil).
func ConfigureLoggingSimple(workspaceRoot, command string, pathSegments []string, debugToConsole bool) error {
	return ConfigureLogging(workspaceRoot, command, pathSegments, debugToConsole, nil)
}
