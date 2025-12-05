// Package internal provides shared utilities for work commands
package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// InitLogger creates a logger for work commands.
// If debug is true, enables debug mode with file logging to out/logs/work/.
// If debug is false, creates a default logger (no file output, debug hidden).
func InitLogger(debug bool, workspaceRoot string) (*logging.Logger, error) {
	if debug {
		return logging.NewWithDebug("work", workspaceRoot)
	}
	return logging.NewDefault("work", workspaceRoot)
}

// WriteDebugFile writes content to a debug file when debug mode is enabled.
// Files are written to out/logs/work/<filename> in the workspace root.
// This is a no-op if debug mode is disabled.
func WriteDebugFile(logger *logging.Logger, workspaceRoot, filename, content string) {
	if logger == nil || !logger.IsDebugMode() {
		return
	}

	debugDir := paths.CommandLogsPath(workspaceRoot, "work")
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		logger.Warn(fmt.Sprintf("Failed to create debug directory: %v", err))
		return
	}

	debugFile := filepath.Join(debugDir, filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		logger.Warn(fmt.Sprintf("Failed to write debug file %s: %v", debugFile, err))
	} else {
		logger.Debug(fmt.Sprintf("Saved debug file: %s", debugFile))
	}
}

// WriteDebugFilef writes content to a debug file with formatted filename.
// The filename format string and args are passed to fmt.Sprintf.
func WriteDebugFilef(logger *logging.Logger, workspaceRoot, format string, content string, args ...interface{}) {
	if logger == nil || !logger.IsDebugMode() {
		return
	}
	filename := fmt.Sprintf(format, args...)
	WriteDebugFile(logger, workspaceRoot, filename, content)
}
