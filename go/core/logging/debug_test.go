package logging

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/ready-to-release/eac/go/core/environments"
)

func TestEnableDisableDebug(t *testing.T) {
	// Reset state
	DisableDebug()

	if IsDebugEnabled() {
		t.Error("expected debug to be disabled initially")
	}

	EnableDebug()
	if !IsDebugEnabled() {
		t.Error("expected debug to be enabled after EnableDebug()")
	}

	DisableDebug()
	if IsDebugEnabled() {
		t.Error("expected debug to be disabled after DisableDebug()")
	}
}

func TestInitFromEnv(t *testing.T) {
	// Save and restore original state
	originalValue := os.Getenv(environments.EnvEACDebug)
	defer func() {
		if originalValue == "" {
			os.Unsetenv(environments.EnvEACDebug)
		} else {
			os.Setenv(environments.EnvEACDebug, originalValue)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"empty env", "", false},
		{"set to 1", "1", true},
		{"set to true", "true", true},
		{"set to any value", "yes", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DisableDebug() // Reset state

			if tt.envValue == "" {
				os.Unsetenv(environments.EnvEACDebug)
			} else {
				os.Setenv(environments.EnvEACDebug, tt.envValue)
			}

			InitFromEnv()

			if IsDebugEnabled() != tt.expected {
				t.Errorf("expected IsDebugEnabled()=%v, got %v", tt.expected, IsDebugEnabled())
			}
		})
	}
}

func TestDebugDirect(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	// Test when disabled
	DisableDebug()
	DebugDirect("test", "should not appear")
	if buf.Len() > 0 {
		t.Error("expected no output when debug is disabled")
	}

	// Test when enabled
	EnableDebug()
	DebugDirect("mymodule", "test message")

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("expected output to contain 'DEBUG', got: %s", output)
	}
	if !strings.Contains(output, "mymodule:test message") {
		t.Errorf("expected output to contain 'mymodule:test message', got: %s", output)
	}

	// Verify format: "HH:MM:SS.mmm  DEBUG  module:message"
	parts := strings.Split(strings.TrimSpace(output), "  ")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts separated by double space, got %d: %v", len(parts), parts)
	}
}

func TestDebugDirectf(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	EnableDebug()
	DebugDirectf("test", "value is %d", 42)

	output := buf.String()
	if !strings.Contains(output, "value is 42") {
		t.Errorf("expected formatted message, got: %s", output)
	}
}

func TestDebugTime(t *testing.T) {
	ts := debugTime()

	// Should be in format "15:04:05.000"
	if len(ts) != 12 {
		t.Errorf("expected timestamp length 12, got %d: %s", len(ts), ts)
	}

	// Should contain colons for time separation
	if strings.Count(ts, ":") != 2 {
		t.Errorf("expected 2 colons in timestamp, got: %s", ts)
	}

	// Should contain dot for milliseconds
	if !strings.Contains(ts, ".") {
		t.Errorf("expected dot for milliseconds in timestamp, got: %s", ts)
	}
}

func TestDebugConcurrency(t *testing.T) {
	buf := &syncBuffer{}
	originalOutput := debugOutput
	debugOutput = buf
	defer func() { debugOutput = originalOutput }()

	EnableDebug()

	// Run concurrent debug calls
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			DebugDirectf("concurrent", "message %d", n)
		}(i)
	}
	wg.Wait()

	// Count output lines
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 100 {
		t.Errorf("expected 100 lines, got %d", len(lines))
	}
}

func TestDebugZeroCostWhenDisabled(t *testing.T) {
	DisableDebug()

	// This should be essentially a no-op
	// We can't directly test performance, but we can verify no output
	var buf bytes.Buffer
	originalOutput := debugOutput
	debugOutput = &buf
	defer func() { debugOutput = originalOutput }()

	for i := 0; i < 1000; i++ {
		DebugDirect("test", "message")
		DebugDirectf("test", "message %d", i)
	}

	if buf.Len() > 0 {
		t.Error("expected no output when debug is disabled")
	}
}
