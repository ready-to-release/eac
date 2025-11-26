package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestConsoleEnablerDefaultMode(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		enabled bool
	}{
		{"Debug hidden by default", DebugLevel, false},
		{"Info shown by default", InfoLevel, true},
		{"Warn shown by default", WarnLevel, true},
		{"Error shown by default", ErrorLevel, true},
	}

	enabler := newConsoleEnabler(false) // debug mode off

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := enabler.Enabled(tt.level); got != tt.enabled {
				t.Errorf("consoleEnabler.Enabled(%v) = %v, want %v", tt.level, got, tt.enabled)
			}
		})
	}
}

func TestConsoleEnablerDebugMode(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		enabled bool
	}{
		{"Debug shown in debug mode", DebugLevel, true},
		{"Info shown in debug mode", InfoLevel, true},
		{"Warn shown in debug mode", WarnLevel, true},
		{"Error shown in debug mode", ErrorLevel, true},
	}

	enabler := newConsoleEnabler(true) // debug mode on

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := enabler.Enabled(tt.level); got != tt.enabled {
				t.Errorf("consoleEnabler.Enabled(%v) = %v, want %v", tt.level, got, tt.enabled)
			}
		})
	}
}

func TestFileEnablerAllLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   Level
		enabled bool
	}{
		{"Debug written to file", DebugLevel, true},
		{"Info written to file", InfoLevel, true},
		{"Warn written to file", WarnLevel, true},
		{"Error written to file", ErrorLevel, true},
	}

	enabler := newFileEnabler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := enabler.Enabled(tt.level); got != tt.enabled {
				t.Errorf("fileEnabler.Enabled(%v) = %v, want %v", tt.level, got, tt.enabled)
			}
		})
	}
}

func TestNoFileLoggingByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig("test-module", tmpDir)

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Sync()

	// Log some messages
	logger.Debug("test debug message")
	logger.Info("test info message")
	logger.Sync()

	// Verify log file was NOT created
	logPath := filepath.Join(tmpDir, "out", "logs", "test-module", "debug.log")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("Log file should NOT be created in default mode, but found at %s", logPath)
	}
}

func TestFileLoggingWhenDebugEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig("test-module", tmpDir).WithDebugMode(true)

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Sync()

	// Log some messages
	logger.Debug("test debug message")
	logger.Info("test info message")
	logger.Sync()

	// Verify log file WAS created
	logPath := filepath.Join(tmpDir, "out", "logs", "test-module", "debug.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file should be created in debug mode at %s", logPath)
	}
}

func TestAllLevelsWriteToFileWhenDebugEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig("test-module", tmpDir).WithDebugMode(true)

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
	logger.Sync()

	logPath := filepath.Join(tmpDir, "out", "logs", "test-module", "debug.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	contentStr := string(content)

	tests := []struct {
		name    string
		message string
	}{
		{"Debug in file", "debug message"},
		{"Info in file", "info message"},
		{"Warn in file", "warn message"},
		{"Error in file", "error message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(contentStr, tt.message) {
				t.Errorf("Log file should contain %q, got: %s", tt.message, contentStr)
			}
		})
	}
}

func TestDebugModeFlag(t *testing.T) {
	t.Run("debug mode disabled by default", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := DefaultConfig("test", tmpDir)
		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer logger.Sync()
		if logger.IsDebugMode() {
			t.Error("Debug mode should be disabled by default")
		}
	})

	t.Run("debug mode enabled with WithDebugMode", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := DefaultConfig("test", tmpDir).WithDebugMode(true)
		logger, err := New(cfg)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer logger.Sync()
		if !logger.IsDebugMode() {
			t.Error("Debug mode should be enabled")
		}
	})

	t.Run("NewWithDebug enables debug mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		logger, err := NewWithDebug("test", tmpDir)
		if err != nil {
			t.Fatalf("NewWithDebug() error = %v", err)
		}
		defer logger.Sync()
		if !logger.IsDebugMode() {
			t.Error("NewWithDebug should enable debug mode")
		}
	})
}

func TestGlobalLogger(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig("global-test", tmpDir)

	if err := Initialize(cfg); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	logger := Get()
	if logger == nil {
		t.Fatal("Get() returned nil")
	}

	// L() should return same instance
	if L() != logger {
		t.Error("L() should return same logger as Get()")
	}
}

func TestChildLoggerWithDebugMode(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig("child-test", tmpDir).WithDebugMode(true)

	logger, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer logger.Sync()

	child := logger.With(zap.String("component", "test"))
	if child == nil {
		t.Fatal("With() returned nil")
	}

	child.Info("child logger message")
	logger.Sync() // Sync parent to flush and close file

	logPath := filepath.Join(tmpDir, "out", "logs", "child-test", "debug.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if !strings.Contains(string(content), "component") {
		t.Errorf("Log should contain component field, got: %s", content)
	}
}
