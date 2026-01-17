package logging

import (
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// configLevelEnabler implements zapcore.LevelEnabler based on config.
type configLevelEnabler struct {
	levels map[zapcore.Level]bool
}

// Enabled returns true if the level is in the configured levels.
func (e *configLevelEnabler) Enabled(lvl zapcore.Level) bool {
	return e.levels[lvl]
}

// newConfigLevelEnabler creates a level enabler from config level strings.
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
	encoder := CreateEncoder(logCfg.Console.Formatter, cfg.Command)
	enabler := newConfigLevelEnabler(logCfg.Console.Levels)

	// If debug mode is enabled, also show debug on console
	if cfg.DebugMode && !logCfg.Console.HasLevel("debug") {
		levels := append([]string{"debug"}, logCfg.Console.Levels...)
		enabler = newConfigLevelEnabler(levels)
	}

	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), enabler)
}

// buildFileCore creates a file output core for unified commands.log with rolling support.
// Returns the core and the writer that needs to be closed.
func buildFileCore(cfg Config, logCfg LoggingConfig) (zapcore.Core, io.Closer, error) {
	// Ensure out/ directory exists
	if err := os.MkdirAll(filepath.Join(cfg.WorkspaceRoot, paths.OutDir), 0o755); err != nil { //nolint:gosec // G301: Log directory should be world-readable
		return nil, nil, err
	}

	// Unified log: out/commands.log with rolling
	logPath := paths.CommandsLogPath(cfg.WorkspaceRoot)
	writer := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    logCfg.File.MaxSizeMB, // MB
		MaxBackups: logCfg.File.MaxBackups,
		MaxAge:     logCfg.File.MaxAgeDays, // days
		Compress:   logCfg.File.Compress != nil && *logCfg.File.Compress,
	}

	encoder := CreateEncoder(logCfg.File.Formatter, cfg.Command)
	enabler := newConfigLevelEnabler(logCfg.File.Levels)

	return zapcore.NewCore(encoder, zapcore.AddSync(writer), enabler), writer, nil
}

// buildTargetFileCore creates extra log core from logging.yml targets config.
// Returns nil if no target is configured for this command or no module is specified.
// Target logs use simple file output (no rolling) since they're per-module.
func buildTargetFileCore(cfg Config, logCfg LoggingConfig) (zapcore.Core, io.Closer, error) {
	// Check if command has a target configured
	target, ok := logCfg.GetTarget(cfg.Command)
	if !ok || cfg.Module == "" {
		return nil, nil, nil
	}

	// Resolve path pattern with module
	targetPath := target.ResolveTargetPath(cfg.WorkspaceRoot, cfg.Module)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil { //nolint:gosec // G301: Log directory should be world-readable
		return nil, nil, err
	}

	file, err := os.OpenFile(targetPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:gosec // G302: Log files should be world-readable
	if err != nil {
		return nil, nil, err
	}

	encoder := CreateEncoder(target.Formatter, cfg.Command)
	enabler := newConfigLevelEnabler(target.Levels)

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

	// R2R_TEST_LOGGING_ACTIVE=true disables unified log but allows target logs
	testLoggingActive := os.Getenv("R2R_TEST_LOGGING_ACTIVE") == "true"

	var closers []io.Closer
	cores := []zapcore.Core{consoleCore}

	// Add unified file core (disabled when test logging active)
	if logCfg.File.IsEnabled() && cfg.EnableFileLogging && !testLoggingActive {
		fileCore, file, err := buildFileCore(cfg, logCfg)
		if err == nil {
			cores = append(cores, fileCore)
			closers = append(closers, file)
		}
	}

	// Add target core if configured (for build/test with module) - always enabled
	if logCfg.File.IsEnabled() && cfg.EnableFileLogging {
		targetCore, targetFile, err := buildTargetFileCore(cfg, logCfg)
		if err == nil && targetCore != nil {
			cores = append(cores, targetCore)
			closers = append(closers, targetFile)
		}
	}

	if len(cores) == 1 {
		return consoleCore, nil, nil
	}
	return zapcore.NewTee(cores...), closers, nil
}
