// Package design provides architecture documentation commands using Structurizr DSL.
//
// This file contains shared logging utilities for the design command and its subcommands.
package design

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// WriteDebugFile writes content to a debug file when debug mode is enabled.
// Files are written to out/logs/design/<filename> in the workspace root.
func WriteDebugFile(workspaceRoot string, logger *logging.Logger, filename string, content string) {
	if logger == nil || !logger.IsDebugMode() {
		return
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to load config: %v", err))
		return
	}
	debugDir := cfg.Repository.LogsPathAbs(workspaceRoot, "design")
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
func WriteDebugFilef(workspaceRoot string, logger *logging.Logger, format string, content string, args ...interface{}) {
	if logger == nil || !logger.IsDebugMode() {
		return
	}
	filename := fmt.Sprintf(format, args...)
	WriteDebugFile(workspaceRoot, logger, filename, content)
}
