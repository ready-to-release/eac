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
)

// loggingState holds the current logging configuration state
var (
	loggingClosers []io.Closer
	loggingMu      sync.Mutex
)

// ConfigureLogging sets up the logging system for a command.
//
// This is the ONE function to call for logging setup. It configures:
//   - File logging: Debug logs ALWAYS go to file (out/logs/module/debug.log)
//   - Console logging: Info/Warn/Error always, Debug only if debugToConsole=true
//   - TUI logging: If tuiWriter provided, logs also go to TUI pane
//
// Parameters:
//   - workspaceRoot: repository root for log file location
//   - module: name for log file directory (e.g., "build", "test")
//   - debugToConsole: if true, debug logs also appear on console (use with --debug flag)
//   - tuiWriter: if non-nil, logs also go to TUI pane (pass nil for non-TUI mode)
//
// Example usage in a command:
//
//	var tuiWriter io.Writer
//	if useTUI {
//	    tuiWriter = orch.GetTUIWriter(tui.PhaseInit)
//	}
//	if err := logging.ConfigureLogging(workspaceRoot, "build", debugMode, tuiWriter); err != nil {
//	    log.Warnf("Failed to configure logging: %v", err)
//	}
//	defer logging.CloseLogging()
func ConfigureLogging(workspaceRoot, module string, debugToConsole bool, tuiWriter io.Writer) error {
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

	// File encoder config - full details for debugging
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
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

	// 2. File core (always present when workspaceRoot provided)
	// Debug logs ALWAYS go to file regardless of debugToConsole setting
	if workspaceRoot != "" && module != "" {
		logDir := filepath.Join(paths.LogsPath(workspaceRoot), module)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		logPath := filepath.Join(logDir, "debug.log")
		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to create debug log file: %w", err)
		}
		closers = append(closers, file)

		fileEncoder := zapcore.NewConsoleEncoder(fileEncoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(file),
			zapcore.DebugLevel, // File ALWAYS gets debug level
		)
		cores = append(cores, fileCore)
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
// Equivalent to ConfigureLogging(workspaceRoot, module, debugToConsole, nil)
func ConfigureLoggingSimple(workspaceRoot, module string, debugToConsole bool) error {
	return ConfigureLogging(workspaceRoot, module, debugToConsole, nil)
}
