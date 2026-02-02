package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDebugDisabledByDefault(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Debug("should not appear")
	Debugf("should not appear: %s", "test")

	if stderr.Len() > 0 {
		t.Errorf("Debug output should be empty when disabled, got: %s", stderr.String())
	}
}

func TestDebugEnabledOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)
	EnableDebug()

	Debug("debug message")

	output := stderr.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Expected debug message in output, got: %s", output)
	}
	// Note: With raw formatter (default), level is not included in output
}

func TestDebugfEnabledOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)
	EnableDebug()

	Debugf("debug with value: %d", 42)

	output := stderr.String()
	if !strings.Contains(output, "debug with value: 42") {
		t.Errorf("Expected formatted debug message, got: %s", output)
	}
}

func TestEnableDisableDebug(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	if IsDebugEnabled() {
		t.Error("Debug should be disabled by default")
	}

	EnableDebug()
	if !IsDebugEnabled() {
		t.Error("Debug should be enabled after EnableDebug()")
	}

	DisableDebug()
	if IsDebugEnabled() {
		t.Error("Debug should be disabled after DisableDebug()")
	}
}

func TestInfoOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Info("info message")

	output := stdout.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("Expected info message in stdout, got: %s", output)
	}
	if stderr.Len() > 0 {
		t.Errorf("Info should not write to stderr, got: %s", stderr.String())
	}
}

func TestInfofOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Infof("info with value: %s", "test")

	output := stdout.String()
	if !strings.Contains(output, "info with value: test") {
		t.Errorf("Expected formatted info message, got: %s", output)
	}
}

func TestWarnOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Warn("warn message")

	output := stderr.String()
	if !strings.Contains(output, "warn message") {
		t.Errorf("Expected warn message in stderr, got: %s", output)
	}
	// Note: With raw formatter (default), level is not included in output
}

func TestWarnfOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Warnf("warn with value: %d", 123)

	output := stderr.String()
	if !strings.Contains(output, "warn with value: 123") {
		t.Errorf("Expected formatted warn message, got: %s", output)
	}
}

func TestErrorOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Error("error message")

	output := stderr.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("Expected error message in stderr, got: %s", output)
	}
	// Note: With raw formatter (default), level is not included in output
}

func TestDebugWithTimestampedFormatter(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	Initialize(LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"debug", "info", "warn", "error"},
			Formatter: FormatterTimestamped,
		},
	})
	EnableDebug()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Debug("debug message")

	output := stderr.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Expected debug message in output, got: %s", output)
	}
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("Expected DEBUG level in timestamped output, got: %s", output)
	}
}

func TestWarnWithTimestampedFormatter(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	Initialize(LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"info", "warn", "error"},
			Formatter: FormatterTimestamped,
		},
	})

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Warn("warn message")

	output := stderr.String()
	if !strings.Contains(output, "warn message") {
		t.Errorf("Expected warn message in output, got: %s", output)
	}
	if !strings.Contains(output, "WARN") {
		t.Errorf("Expected WARN level in timestamped output, got: %s", output)
	}
}

func TestErrorWithTimestampedFormatter(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	Initialize(LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"info", "warn", "error"},
			Formatter: FormatterTimestamped,
		},
	})

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Error("error message")

	output := stderr.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("Expected error message in output, got: %s", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected ERROR level in timestamped output, got: %s", output)
	}
}

func TestErrorfOutput(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Errorf("error with value: %v", "details")

	output := stderr.String()
	if !strings.Contains(output, "error with value: details") {
		t.Errorf("Expected formatted error message, got: %s", output)
	}
}

func TestSetLevel(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	tests := []struct {
		level       string
		expectDebug bool
		expectError bool
	}{
		{"debug", true, false},
		{"info", false, false},
		{"warn", false, false},
		{"warning", false, false},
		{"error", false, false},
		{"invalid", false, true},
	}

	for _, tt := range tests {
		ResetForTesting()
		err := SetLevel(tt.level)

		if tt.expectError && err == nil {
			t.Errorf("SetLevel(%q) expected error, got nil", tt.level)
		}
		if !tt.expectError && err != nil {
			t.Errorf("SetLevel(%q) unexpected error: %v", tt.level, err)
		}
		if !tt.expectError && IsDebugEnabled() != tt.expectDebug {
			t.Errorf("SetLevel(%q) debug enabled = %v, want %v", tt.level, IsDebugEnabled(), tt.expectDebug)
		}
	}
}

func TestGetLevel(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	if GetLevel() != "info" {
		t.Errorf("Default level should be 'info', got %q", GetLevel())
	}

	EnableDebug()
	if GetLevel() != "debug" {
		t.Errorf("Level should be 'debug' after EnableDebug(), got %q", GetLevel())
	}

	DisableDebug()
	if GetLevel() != "info" {
		t.Errorf("Level should be 'info' after DisableDebug(), got %q", GetLevel())
	}
}

func TestFormatterRaw(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	// Initialize with raw formatter
	Initialize(LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"info", "warn", "error"},
			Formatter: FormatterRaw,
		},
	})

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Info("raw message")

	output := stdout.String()
	// Raw format should just have the message
	if !strings.Contains(output, "raw message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
}

func TestFormatterTimestamped(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	Initialize(LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"info", "warn", "error"},
			Formatter: FormatterTimestamped,
		},
	})

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Info("timestamped message")

	output := stdout.String()
	// Timestamped format should have HH:MM:SS pattern and INFO level
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected INFO in timestamped output, got: %s", output)
	}
	if !strings.Contains(output, "timestamped message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
	// Check for time pattern (at least contains colons for HH:MM:SS)
	if !strings.Contains(output, ":") {
		t.Errorf("Expected timestamp with colons in output, got: %s", output)
	}
}

func TestFormatterJSON(t *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	Initialize(LoggingConfig{
		Console: SinkConfig{
			Levels:    []string{"info", "warn", "error"},
			Formatter: FormatterJSON,
		},
	})

	var stdout, stderr bytes.Buffer
	SetOutput(&stdout, &stderr)

	Info("json message")

	output := strings.TrimSpace(stdout.String())

	var data map[string]string
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Errorf("Expected valid JSON output, got error: %v, output: %s", err, output)
	}

	if data["level"] != "INFO" {
		t.Errorf("Expected level 'INFO', got %q", data["level"])
	}
	if data["message"] != "json message" {
		t.Errorf("Expected message 'json message', got %q", data["message"])
	}
	if data["time"] == "" {
		t.Error("Expected time field in JSON output")
	}
}

func TestResetOutput(_ *testing.T) {
	ResetForTesting()
	defer ResetForTesting()

	var buf bytes.Buffer
	SetOutput(&buf, &buf)

	// After reset, should use os.Stdout/os.Stderr (we can't easily verify this,
	// but we can verify the function doesn't panic)
	ResetOutput()
}

func TestResetForTesting(t *testing.T) {
	// Enable debug and set custom output
	EnableDebug()
	var buf bytes.Buffer
	SetOutput(&buf, &buf)

	// Reset should clear everything
	ResetForTesting()

	if IsDebugEnabled() {
		t.Error("Debug should be disabled after ResetForTesting()")
	}
}
