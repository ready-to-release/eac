package templates

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"go.uber.org/zap"
)

// writeDebugFile writes content to a debug file when debug mode is enabled.
// Files are written to out/logs/templates/<subcommand>/<filename> in the workspace root.
//
// Parameters:
//   - subcommand: The subcommand name (e.g., "list", "apply", "install")
//   - workspaceRoot: The repository root directory
//   - logger: The logger instance to use for debug output
//   - filename: The name of the debug file to create
//   - content: The content to write to the file
func writeDebugFile(subcommand, workspaceRoot string, logger *logging.Logger, filename string, content string) {
	if !logger.IsDebugMode() {
		return
	}

	cfg, err := config.Load(config.LoadOptions{RepoRoot: workspaceRoot})
	if err != nil {
		logger.Warn("Failed to load config", zap.Error(err))
		return
	}
	debugDir := filepath.Join(cfg.Repository.LogsPathAbs(workspaceRoot, "templates"), subcommand)
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		logger.Warn("Failed to create debug directory", zap.String("dir", debugDir), zap.Error(err))
		return
	}

	debugFile := filepath.Join(debugDir, filename)
	if err := os.WriteFile(debugFile, []byte(content), 0644); err != nil {
		logger.Warn("Failed to write debug file", zap.String("file", debugFile), zap.Error(err))
	} else {
		logger.Debug("Saved debug file", zap.String("file", debugFile))
	}
}

// writeDebugFilef writes content to a debug file with formatted filename.
// This is a convenience wrapper around writeDebugFile that formats the filename.
//
// Parameters:
//   - subcommand: The subcommand name (e.g., "list", "apply", "install")
//   - workspaceRoot: The repository root directory
//   - logger: The logger instance to use for debug output
//   - format: Format string for the filename (printf-style)
//   - content: The content to write to the file
//   - args: Arguments for the format string
func writeDebugFilef(subcommand, workspaceRoot string, logger *logging.Logger, format string, content string, args ...interface{}) {
	if !logger.IsDebugMode() {
		return
	}
	filename := fmt.Sprintf(format, args...)
	writeDebugFile(subcommand, workspaceRoot, logger, filename, content)
}

// logDebugInfo logs debug information when debug mode is enabled
// This is a helper for common debug logging patterns
func logDebugInfo(logger *logging.Logger, message string, fields ...zap.Field) {
	if logger.IsDebugMode() {
		logger.Debug(message, fields...)
	}
}

// logInfo logs an info message
// This is a wrapper for consistency
func logInfo(logger *logging.Logger, message string, fields ...zap.Field) {
	logger.Info(message, fields...)
}

// logWarn logs a warning message
// This is a wrapper for consistency
func logWarn(logger *logging.Logger, message string, fields ...zap.Field) {
	logger.Warn(message, fields...)
}

// logError logs an error message
// This is a wrapper for consistency
func logError(logger *logging.Logger, message string, fields ...zap.Field) {
	logger.Error(message, fields...)
}
