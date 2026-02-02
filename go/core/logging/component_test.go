package logging

import (
	"os"
	"strings"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestMain sets up the test environment.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestComponentLoggerCreation(t *testing.T) {
	// Test explicit component name
	log := C("mymodule")
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	if log.component != "mymodule" {
		t.Errorf("expected component 'mymodule', got '%s'", log.component)
	}

	// Test long form with explicit name
	log2 := Component("another")
	if log2.component != "another" {
		t.Errorf("expected component 'another', got '%s'", log2.component)
	}
}

func TestComponentLoggerInferment(t *testing.T) {
	// Test inference from call site (no argument)
	log := C()
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	// The inferred name should contain "logging" since we're in that package
	if !strings.Contains(log.component, "logging") {
		t.Errorf("expected component to contain 'logging', got '%s'", log.component)
	}

	// Test long form inference
	log2 := Component()
	if !strings.Contains(log2.component, "logging") {
		t.Errorf("expected component to contain 'logging', got '%s'", log2.component)
	}
}

func TestInferComponentFunction(t *testing.T) {
	// Test the inferComponent function directly by wrapping it
	// We need to call it through C() to test the skip count
	log := C()

	// Should have inferred from this test file's package
	// Package path: github.com/ready-to-release/eac/go/core/logging
	// Inferred: core/logging
	if log.component == "" || log.component == "unknown" {
		t.Errorf("inferComponent returned empty or unknown, got '%s'", log.component)
	}
}

// setupTestLogger creates a component logger with an observer core for testing.
// Returns the logger, observed logs, and a cleanup function that must be called
// when the test is done (use defer).
func setupTestLogger(component string) (*ComponentLogger, *observer.ObservedLogs, func()) {
	// Create observer core to capture log entries
	core, logs := observer.New(zapcore.DebugLevel)

	// Create a zap logger with the observer core
	zapLogger := zap.New(core)

	// Inject this logger into the global state
	cleanup := SetTestLogger(zapLogger)

	// Create component logger (will use the injected test logger)
	return &ComponentLogger{
		component: component,
	}, logs, cleanup
}

func TestComponentLoggerDebug(t *testing.T) {
	log, logs, cleanup := setupTestLogger("test")
	defer cleanup()

	// Log a debug message
	log.Debug("test message")

	// Verify message was logged
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zapcore.DebugLevel {
		t.Errorf("expected DEBUG level, got %v", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("expected message 'test message', got '%s'", entry.Message)
	}

	// Note: component field is NOT added to log context to avoid polluting console output.
	// The component name is stored in ComponentLogger.component for internal tracking only.
}

func TestComponentLoggerDebugf(t *testing.T) {
	log, logs, cleanup := setupTestLogger("repository")
	defer cleanup()

	// Log formatted debug message
	log.Debugf("processing %d files", 42)

	// Verify message was logged
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if !strings.Contains(entry.Message, "processing 42 files") {
		t.Errorf("expected formatted message, got: %s", entry.Message)
	}
}

func TestComponentLoggerInfo(t *testing.T) {
	log, logs, cleanup := setupTestLogger("test")
	defer cleanup()

	log.Info("info message")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zapcore.InfoLevel {
		t.Errorf("expected INFO level, got %v", entry.Level)
	}
	if entry.Message != "info message" {
		t.Errorf("expected message 'info message', got '%s'", entry.Message)
	}
}

func TestComponentLoggerInfof(t *testing.T) {
	log, logs, cleanup := setupTestLogger("test")
	defer cleanup()

	log.Infof("count: %d", 5)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if !strings.Contains(entry.Message, "count: 5") {
		t.Errorf("expected formatted message, got: %s", entry.Message)
	}
}

func TestComponentLoggerWarn(t *testing.T) {
	log, logs, cleanup := setupTestLogger("test")
	defer cleanup()

	log.Warn("warning message")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zapcore.WarnLevel {
		t.Errorf("expected WARN level, got %v", entry.Level)
	}
	if entry.Message != "warning message" {
		t.Errorf("expected message 'warning message', got '%s'", entry.Message)
	}
}

func TestComponentLoggerWarnf(t *testing.T) {
	log, logs, cleanup := setupTestLogger("test")
	defer cleanup()

	log.Warnf("warning %s", "here")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if !strings.Contains(entry.Message, "warning here") {
		t.Errorf("expected formatted message, got: %s", entry.Message)
	}
}

func TestComponentLoggerError(t *testing.T) {
	log, logs, cleanup := setupTestLogger("test")
	defer cleanup()

	log.Error("error message")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zapcore.ErrorLevel {
		t.Errorf("expected ERROR level, got %v", entry.Level)
	}
	if entry.Message != "error message" {
		t.Errorf("expected message 'error message', got '%s'", entry.Message)
	}
}

func TestComponentLoggerErrorf(t *testing.T) {
	log, logs, cleanup := setupTestLogger("test")
	defer cleanup()

	log.Errorf("error code: %d", 500)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if !strings.Contains(entry.Message, "error code: 500") {
		t.Errorf("expected formatted message, got: %s", entry.Message)
	}
}

func TestComponentLoggerWithSuffix(t *testing.T) {
	// Create base logger with observer
	baseLog, logs, cleanup := setupTestLogger("git")
	defer cleanup()

	// Create logger with suffix
	funcLog := baseLog.WithSuffix("StagedFiles")

	// Verify component name includes suffix
	if funcLog.component != "git.StagedFiles" {
		t.Errorf("expected component 'git.StagedFiles', got '%s'", funcLog.component)
	}

	// Log a message
	funcLog.Debug("start")

	// Verify the log was captured
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Message != "start" {
		t.Errorf("expected message 'start', got '%s'", entry.Message)
	}

	// Note: component field is NOT added to log context to avoid polluting console output.
	// The suffix is tracked internally in ComponentLogger.component only.
}

func TestComponentLoggerLevelFiltering(t *testing.T) {
	// Create logger with INFO level (filters out DEBUG)
	core, logs := observer.New(zapcore.InfoLevel)
	zapLogger := zap.New(core)
	cleanup := SetTestLogger(zapLogger)
	defer cleanup()

	log := &ComponentLogger{
		component: "test",
	}

	// Log at different levels
	log.Debug("debug - should be filtered")
	log.Info("info - should appear")
	log.Warn("warn - should appear")
	log.Error("error - should appear")

	entries := logs.All()
	// Should only have 3 entries (Info, Warn, Error - Debug filtered)
	if len(entries) != 3 {
		t.Fatalf("expected 3 log entries (Info, Warn, Error), got %d", len(entries))
	}

	// Verify levels
	levels := []zapcore.Level{zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel}
	for i, entry := range entries {
		if entry.Level != levels[i] {
			t.Errorf("entry %d: expected level %v, got %v", i, levels[i], entry.Level)
		}
	}
}

func TestComponentLoggerConcurrency(t *testing.T) {
	log, logs, cleanup := setupTestLogger("concurrent")
	defer cleanup()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			log.Debugf("message %d", n)
		}(i)
	}
	wg.Wait()

	entries := logs.All()
	if len(entries) != 50 {
		t.Errorf("expected 50 log entries, got %d", len(entries))
	}
}

func TestComponentLoggerPackageLevelUsage(t *testing.T) {
	log, logs, cleanup := setupTestLogger("core-git")
	defer cleanup()

	log.Debug("package initialized")

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Message != "package initialized" {
		t.Errorf("expected message 'package initialized', got '%s'", entry.Message)
	}

	// Verify the component is stored internally (not in log context)
	if log.component != "core-git" {
		t.Errorf("expected component 'core-git', got '%s'", log.component)
	}
}
