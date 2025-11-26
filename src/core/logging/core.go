package logging

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// coreResult holds the core and any closers that need to be closed on Sync.
type coreResult struct {
	core    zapcore.Core
	closers []io.Closer
}

// buildConsoleCore creates a console output core.
// Console shows Info/Warn/Error by default, or all levels in debug mode.
func buildConsoleCore(cfg Config) zapcore.Core {
	encoderCfg := zap.NewProductionEncoderConfig()
	if cfg.Development {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(encoderCfg)
	enabler := newConsoleEnabler(cfg.DebugMode)

	return zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), enabler)
}

// buildFileCore creates a file output core.
// File receives all levels (Debug, Info, Warn, Error) when debug mode is enabled.
// Returns the core and the file that needs to be closed.
func buildFileCore(cfg Config) (zapcore.Core, *os.File, error) {
	// Create log directory: out/logs/<module>/
	logDir := filepath.Join(cfg.WorkspaceRoot, "out", "logs", cfg.Module)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, err
	}

	logPath := filepath.Join(logDir, "debug.log")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	enabler := newFileEnabler()

	return zapcore.NewCore(encoder, zapcore.AddSync(file), enabler), file, nil
}

// buildCore creates the appropriate core based on configuration.
// Default mode: console only (no file logging)
// Debug mode: console + file (tee core)
// Returns the core and any closers that need to be closed on Sync.
func buildCore(cfg Config) (zapcore.Core, []io.Closer, error) {
	consoleCore := buildConsoleCore(cfg)

	// Only add file core when debug mode is enabled
	if !cfg.DebugMode {
		return consoleCore, nil, nil
	}

	fileCore, file, err := buildFileCore(cfg)
	if err != nil {
		// Log warning but continue with console only
		return consoleCore, nil, nil
	}

	return zapcore.NewTee(consoleCore, fileCore), []io.Closer{file}, nil
}
