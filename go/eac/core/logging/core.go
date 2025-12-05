package logging

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"go.uber.org/zap/zapcore"
)

// logsPath returns the path to the logs output directory
func logsPath(repoRoot string) string {
	return paths.LogsPath(repoRoot)
}

// configLevelEnabler implements zapcore.LevelEnabler based on config
type configLevelEnabler struct {
	levels map[zapcore.Level]bool
}

// Enabled returns true if the level is in the configured levels
func (e *configLevelEnabler) Enabled(lvl zapcore.Level) bool {
	return e.levels[lvl]
}

// newConfigLevelEnabler creates a level enabler from config level strings
func newConfigLevelEnabler(levels []string) zapcore.LevelEnabler {
	enabler := &configLevelEnabler{
		levels: make(map[zapcore.Level]bool),
	}
	for _, level := range levels {
		switch level {
		case "debug":
			enabler.levels[zapcore.DebugLevel] = true
		case "info":
			enabler.levels[zapcore.InfoLevel] = true
		case "warn":
			enabler.levels[zapcore.WarnLevel] = true
		case "error":
			enabler.levels[zapcore.ErrorLevel] = true
		}
	}
	return enabler
}

// buildConsoleCore creates a console output core based on logging config.
func buildConsoleCore(cfg Config, logCfg LoggingConfig) zapcore.Core {
	encoder := CreateEncoder(logCfg.Console.Formatter, cfg.Module)
	enabler := newConfigLevelEnabler(logCfg.Console.Levels)

	// If debug mode is enabled, also show debug on console
	if cfg.DebugMode && !logCfg.Console.HasLevel("debug") {
		levels := append([]string{"debug"}, logCfg.Console.Levels...)
		enabler = newConfigLevelEnabler(levels)
	}

	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), enabler)
}

// buildFileCore creates a file output core based on logging config.
// Returns the core and the file that needs to be closed.
func buildFileCore(cfg Config, logCfg LoggingConfig) (zapcore.Core, *os.File, error) {
	// Create log directory: out/logs/<module>/
	logDir := filepath.Join(logsPath(cfg.WorkspaceRoot), cfg.Module)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, err
	}

	logPath := filepath.Join(logDir, "debug.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	encoder := CreateEncoder(logCfg.File.Formatter, cfg.Module)
	enabler := newConfigLevelEnabler(logCfg.File.Levels)

	return zapcore.NewCore(encoder, zapcore.AddSync(file), enabler), file, nil
}

// buildCore creates the appropriate core based on configuration.
// Uses logging.yml config to determine:
// - Which levels go to console
// - Which levels go to file
// - What formatter each sink uses
// Returns the core and any closers that need to be closed on Sync.
func buildCore(cfg Config) (zapcore.Core, []io.Closer, error) {
	// Load logging configuration from .r2r/eac/logging.yml
	logCfg := LoadLoggingConfig(cfg.WorkspaceRoot)

	consoleCore := buildConsoleCore(cfg, logCfg)

	// Only add file core when enabled
	if !logCfg.File.IsEnabled() {
		return consoleCore, nil, nil
	}

	// File logging requires debug mode to actually write
	if !cfg.DebugMode {
		return consoleCore, nil, nil
	}

	fileCore, file, err := buildFileCore(cfg, logCfg)
	if err != nil {
		// Continue with console only on error
		return consoleCore, nil, nil
	}

	return zapcore.NewTee(consoleCore, fileCore), []io.Closer{file}, nil
}
