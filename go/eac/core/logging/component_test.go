package logging

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
)

// TestMain sets up the test environment with timestamped formatter.
func TestMain(m *testing.M) {
	consoleConfig = &SinkConfig{
		Levels:    []string{"debug", "info", "warn", "error"},
		Formatter: FormatterTimestamped,
	}
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
	// The inferred name should contain "core/logging" since we're in that package
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
	// Package path: github.com/ready-to-release/eac/go/eac/core/logging
	// Inferred: core/logging
	if log.component == "" || log.component == "unknown" {
		t.Errorf("inferComponent returned empty or unknown, got '%s'", log.component)
	}
}

func TestComponentLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	log := C("git")

	// Test when disabled
	DisableDebug()
	log.Debug("should not appear")
	if buf.Len() > 0 {
		t.Error("expected no output when debug is disabled")
	}

	// Test when enabled
	EnableDebug()
	log.Debug("test message")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("expected output to contain 'DEBUG', got: %s", output)
	}
	if !strings.Contains(output, "git:test message") {
		t.Errorf("expected output to contain 'git:test message', got: %s", output)
	}
}

func TestComponentLoggerDebugf(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	EnableDebug()
	log := C("repository")
	log.Debugf("processing %d files", 42)

	output := buf.String()
	if !strings.Contains(output, "repository:processing 42 files") {
		t.Errorf("expected formatted message, got: %s", output)
	}
}

func TestComponentLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := stdOutput
	stdOutput = &buf
	defer func() { stdOutput = originalOutput }()

	DisableDebug() // Info should still work even when debug is disabled
	log := C("test")
	log.Info("info message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected output to contain 'INFO', got: %s", output)
	}
	if !strings.Contains(output, "test:info message") {
		t.Errorf("expected output to contain 'test:info message', got: %s", output)
	}
}

func TestComponentLoggerInfof(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := stdOutput
	stdOutput = &buf
	defer func() { stdOutput = originalOutput }()

	log := C("test")
	log.Infof("count: %d", 5)

	output := buf.String()
	if !strings.Contains(output, "test:count: 5") {
		t.Errorf("expected formatted message, got: %s", output)
	}
}

func TestComponentLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	log := C("test")
	log.Warn("warning message")

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("expected output to contain 'WARN', got: %s", output)
	}
	if !strings.Contains(output, "test:warning message") {
		t.Errorf("expected output to contain 'test:warning message', got: %s", output)
	}
}

func TestComponentLoggerWarnf(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	log := C("test")
	log.Warnf("warning %s", "here")

	output := buf.String()
	if !strings.Contains(output, "test:warning here") {
		t.Errorf("expected formatted message, got: %s", output)
	}
}

func TestComponentLoggerError(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	log := C("test")
	log.Error("error message")

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected output to contain 'ERROR', got: %s", output)
	}
	if !strings.Contains(output, "test:error message") {
		t.Errorf("expected output to contain 'test:error message', got: %s", output)
	}
}

func TestComponentLoggerErrorf(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	log := C("test")
	log.Errorf("error code: %d", 500)

	output := buf.String()
	if !strings.Contains(output, "test:error code: 500") {
		t.Errorf("expected formatted message, got: %s", output)
	}
}

func TestComponentLoggerWithSuffix(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	EnableDebug()
	log := C("git")
	funcLog := log.WithSuffix("StagedFiles")
	funcLog.Debug("start")

	output := buf.String()
	if !strings.Contains(output, "git.StagedFiles:start") {
		t.Errorf("expected 'git.StagedFiles:start', got: %s", output)
	}
}

func TestComponentLoggerOutputFormat(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	EnableDebug()
	log := C("mymodule")
	log.Debug("test")

	output := strings.TrimSpace(buf.String())

	// Format should be: "HH:MM:SS.mmm  DEBUG  module:message"
	// Using double spaces as separators
	parts := strings.Split(output, "  ")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d: %v", len(parts), parts)
	}

	// Part 0: timestamp
	if len(parts[0]) != 12 { // "15:04:05.000"
		t.Errorf("expected timestamp length 12, got %d: %s", len(parts[0]), parts[0])
	}

	// Part 1: level
	if parts[1] != "DEBUG" {
		t.Errorf("expected 'DEBUG', got '%s'", parts[1])
	}

	// Part 2: module:message
	if parts[2] != "mymodule:test" {
		t.Errorf("expected 'mymodule:test', got '%s'", parts[2])
	}
}

func TestComponentLoggerLevelFormatAlignment(t *testing.T) {
	// Capture both stdout (Info) and stderr (Debug/Warn/Error)
	var stdBuf bytes.Buffer
	var debugBuf bytes.Buffer
	originalStdOutput := stdOutput
	originalDebugOutput := debugOutput
	stdOutput = &stdBuf
	debugOutput = &debugBuf
	defer func() {
		stdOutput = originalStdOutput
		debugOutput = originalDebugOutput
	}()

	EnableDebug()
	log := C("test")

	log.Debug("d")
	log.Info("i")
	log.Warn("w")
	log.Error("e")

	// Combine outputs: Debug/Warn/Error go to debugBuf, Info goes to stdBuf
	allOutput := debugBuf.String() + stdBuf.String()
	lines := strings.Split(strings.TrimSpace(allOutput), "\n")

	// Should have 4 lines total (Debug, Info, Warn, Error)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d: %v", len(lines), lines)
	}

	// All level strings should be present (order may vary due to different buffers)
	expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for _, level := range expectedLevels {
		found := false
		for _, line := range lines {
			if strings.Contains(line, level) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find '%s' in output, got: %v", level, lines)
		}
	}
}

func TestComponentLoggerConcurrency(t *testing.T) {
	buf := &syncBuffer{}
	originalOutput := debugOutput
	debugOutput = buf
	defer func() { debugOutput = originalOutput }()

	EnableDebug()
	log := C("concurrent")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			log.Debugf("message %d", n)
		}(i)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 50 {
		t.Errorf("expected 50 lines, got %d", len(lines))
	}
}

func TestComponentLoggerPackageLevelUsage(t *testing.T) {
	// Simulate package-level logger pattern
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	EnableDebug()

	// This is how it would be used in a real package
	pkgLog := C("eac-core-git")
	pkgLog.Debug("package initialized")

	output := buf.String()
	if !strings.Contains(output, "eac-core-git:package initialized") {
		t.Errorf("expected module name in output, got: %s", output)
	}
}
