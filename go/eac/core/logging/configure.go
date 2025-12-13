// Package logging - Unified logging configuration
//
// This file provides a single entry point for configuring the logging system.
// It replaces the scattered EnableFileLogging, EnableTUIForComponentLogger, etc.
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// loggingState holds the current logging configuration state
var (
	loggingClosers []io.Closer
	loggingMu      sync.Mutex
)

// ConfigureLogging sets up the logging system for a command.
//
// This is the ONE function to call for logging setup. It configures:
//   - File logging: All logs go to unified out/commands.log
//   - Target logging: For build/test, additional logs go to out/build/<unit>/build.log or out/test/<unit>/test.log
//   - Console logging: Info/Warn/Error always, Debug only if debugToConsole=true
//   - TUI logging: If tuiWriter provided, logs also go to TUI pane
//
// Parameters:
//   - workspaceRoot: repository root for log file location
//   - command: command name (e.g., "build", "test", "design")
//   - pathSegments: optional path segments. For build/test, first segment is the unit name
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
//	// Build with module: logs to out/commands.log + out/build/eac-core/build.log
//	if err := logging.ConfigureLogging(workspaceRoot, "build", []string{"eac-core"}, debugMode, tuiWriter); err != nil {
//	    log.Warnf("Failed to configure logging: %v", err)
//	}
//	// Test with suite: logs to out/commands.log + out/test/commit/test.log
//	if err := logging.ConfigureLogging(workspaceRoot, "test", []string{"commit"}, debugMode, tuiWriter); err != nil {
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
	// R2R_TEST_LOGGING_ACTIVE=true disables unified log but allows target logs
	testLoggingActive := os.Getenv("R2R_TEST_LOGGING_ACTIVE") == "true"

	if workspaceRoot != "" && command != "" {
		// Ensure out/ directory exists
		if err := os.MkdirAll(filepath.Join(workspaceRoot, paths.OutDir), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		// Load logging config for rolling settings
		logCfg := LoadLoggingConfig(workspaceRoot)

		// 2a. Unified log: out/commands.log (disabled when test logging active)
		if !testLoggingActive {
			logPath := paths.CommandsLogPath(workspaceRoot)
			writer := &lumberjack.Logger{
				Filename:   logPath,
				MaxSize:    logCfg.File.MaxSizeMB,
				MaxBackups: logCfg.File.MaxBackups,
				MaxAge:     logCfg.File.MaxAgeDays,
				Compress:   logCfg.File.Compress != nil && *logCfg.File.Compress,
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

		// 2b. Target file core (for build/test with unit) - always enabled
		if target, ok := logCfg.GetTarget(command); ok && len(pathSegments) > 0 {
			unit := pathSegments[0] // First path segment is the unit
			targetPath := target.ResolveTargetPath(workspaceRoot, unit)
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err == nil {
				targetFile, err := os.OpenFile(targetPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

// closeLoggingLocked closes logging resources (must hold loggingMu)
func closeLoggingLocked() {
	// Sync the zap logger to flush any buffered logs
	if componentGlobalLogger != nil {
		componentGlobalLogger.Logger.Sync()
	}

	for _, closer := range loggingClosers {
		closer.Close()
	}
	loggingClosers = nil
}

// ConfigureLoggingSimple is a convenience wrapper for non-TUI commands.
// Equivalent to ConfigureLogging(workspaceRoot, command, pathSegments, debugToConsole, nil)
func ConfigureLoggingSimple(workspaceRoot, command string, pathSegments []string, debugToConsole bool) error {
	return ConfigureLogging(workspaceRoot, command, pathSegments, debugToConsole, nil)
}
