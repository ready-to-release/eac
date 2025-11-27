// Package design provides architecture documentation commands using Structurizr DSL.
//
// This file contains shared logging utilities for the design command and its subcommands.
package design

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/core/logging"
)

// writeDebugFile writes content to a debug file when debug mode is enabled.
// Files are written to out/logs/design/<filename> in the workspace root.
func writeDebugFile(workspaceRoot string, logger *logging.Logger, filename string, content string) {
	if logger == nil || !logger.IsDebugMode() {
		return
	}

	debugDir := filepath.Join(workspaceRoot, "out", "logs", "design")
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

// writeDebugFilef writes content to a debug file with formatted filename.
func writeDebugFilef(workspaceRoot string, logger *logging.Logger, format string, content string, args ...interface{}) {
	if logger == nil || !logger.IsDebugMode() {
		return
	}
	filename := fmt.Sprintf(format, args...)
	writeDebugFile(workspaceRoot, logger, filename, content)
}
