package logging

import (
	"fmt"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// tuiCore is a zapcore.Core that writes formatted log entries to an io.Writer (typically a TUI pane).
// It formats messages in a simple, human-readable format suitable for TUI display.
type tuiCore struct {
	zapcore.LevelEnabler
	writer  io.Writer
	encoder zapcore.Encoder
}

// NewTUICore creates a zapcore.Core that writes to the given writer.
// This is designed for use with TUI panes, formatting messages without timestamps or caller info.
//
// The writer should be a tuiWriter or similar that forwards to the TUI.
// The minLevel determines which log levels are written to the TUI (typically DebugLevel or InfoLevel).
func NewTUICore(writer io.Writer, minLevel zapcore.Level) zapcore.Core {
	// Create a simple console encoder without timestamps (TUI doesn't need them)
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder, // Not used since TimeKey is empty
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Use console encoder for human-readable output
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	return &tuiCore{
		LevelEnabler: minLevel,
		writer:       writer,
		encoder:      encoder,
	}
}

// Enabled implements zapcore.Core.
func (c *tuiCore) Enabled(lvl zapcore.Level) bool {
	return c.LevelEnabler.Enabled(lvl)
}

// With implements zapcore.Core.
func (c *tuiCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.encoder = c.encoder.Clone()
	for _, field := range fields {
		field.AddTo(clone.encoder)
	}
	return &clone
}

// Check implements zapcore.Core.
func (c *tuiCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

// Write implements zapcore.Core.
func (c *tuiCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Encode the entry
	buf, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}
	defer buf.Free()

	// Write to the TUI writer
	_, err = c.writer.Write(buf.Bytes())
	return err
}

// Sync implements zapcore.Core.
func (c *tuiCore) Sync() error {
	// If the writer implements io.Closer, sync it
	if syncer, ok := c.writer.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}

// EnableTUIForComponentLogger reconfigures the global component logger to include TUI output.
// This should be called early, before any ComponentLoggers are created, typically in build/test commands.
//
// The minLevel parameter controls which log levels go to the TUI:
//   - zapcore.DebugLevel: All logs (Debug, Info, Warn, Error) go to TUI (use with --debug flag)
//   - zapcore.InfoLevel: Only Info, Warn, Error go to TUI (normal mode)
//
// Example usage in build.go:
//
//	if tuiWriter := orch.GetTUIWriter(tui.PhaseInit); tuiWriter != nil {
//	    level := zapcore.InfoLevel
//	    if debugMode {
//	        level = zapcore.DebugLevel
//	    }
//	    logging.EnableTUIForComponentLogger(tuiWriter, level)
//	}
func EnableTUIForComponentLogger(writer io.Writer, minLevel zapcore.Level) error {
	if writer == nil {
		return fmt.Errorf("TUI writer cannot be nil")
	}

	// Get or create the base config
	cfg := DefaultConfig("eac", ".")
	if IsDebugEnabled() {
		cfg = cfg.WithDebugMode(true)
	}

	// Build the normal cores (console + optional file)
	normalCore, closers, err := buildCore(cfg)
	if err != nil {
		return fmt.Errorf("failed to build logger cores: %w", err)
	}

	// Create TUI core
	tuiCoreInstance := NewTUICore(writer, minLevel)

	// Combine normal cores with TUI core using Tee
	combinedCore := zapcore.NewTee(normalCore, tuiCoreInstance)

	// Create new logger with combined core
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}
	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	newZapLogger := zap.New(combinedCore, opts...)

	// Replace the global component logger
	componentGlobalLogger = &Logger{
		Logger:  newZapLogger,
		config:  cfg,
		closers: closers,
	}

	return nil
}
