package logging

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestRawEncoder(t *testing.T) {
	encoder := newRawEncoder()

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "Hello World",
	}

	buf, err := encoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Raw encoder should only output message + newline
	expected := "Hello World\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}

	// Should NOT contain timestamp or level
	if strings.Contains(output, "INFO") {
		t.Error("raw encoder should not contain level")
	}
	if strings.Contains(output, ":") {
		t.Error("raw encoder should not contain timestamp separator")
	}
}

func TestTimestampedEncoder(t *testing.T) {
	encoder := newTimestampedEncoder("mymodule")

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Date(2024, 1, 15, 14, 30, 45, 123000000, time.UTC),
		Message: "Test message",
	}

	buf, err := encoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should contain timestamp
	if !strings.Contains(output, "14:30:45.123") {
		t.Errorf("expected timestamp in output, got %q", output)
	}

	// Should contain level
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO level in output, got %q", output)
	}

	// Should contain module
	if !strings.Contains(output, "mymodule:") {
		t.Errorf("expected module in output, got %q", output)
	}

	// Should contain message
	if !strings.Contains(output, "Test message") {
		t.Errorf("expected message in output, got %q", output)
	}
}

func TestTimestampedEncoder_LevelPadding(t *testing.T) {
	encoder := newTimestampedEncoder("test")

	tests := []struct {
		level       zapcore.Level
		expectedStr string
	}{
		{zapcore.DebugLevel, "DEBUG"},
		{zapcore.InfoLevel, "INFO"},
		{zapcore.WarnLevel, "WARN"},
		{zapcore.ErrorLevel, "ERROR"},
	}

	for _, tc := range tests {
		entry := zapcore.Entry{
			Level:   tc.level,
			Time:    time.Now(),
			Message: "msg",
		}

		buf, err := encoder.EncodeEntry(entry, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, tc.expectedStr) {
			t.Errorf("expected %q in output, got %q", tc.expectedStr, output)
		}
	}
}

func TestJSONEncoder(t *testing.T) {
	encoder := newJSONEncoder()

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
		Message: "Test message",
	}

	buf, err := encoder.EncodeEntry(entry, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should be valid JSON with expected fields
	if !strings.Contains(output, `"level":"info"`) {
		t.Errorf("expected level in JSON output, got %q", output)
	}
	if !strings.Contains(output, `"message":"Test message"`) {
		t.Errorf("expected message in JSON output, got %q", output)
	}
	if !strings.Contains(output, `"timestamp"`) {
		t.Errorf("expected timestamp in JSON output, got %q", output)
	}
}

func TestCreateEncoder(t *testing.T) {
	tests := []struct {
		formatter FormatterType
		module    string
	}{
		{FormatterRaw, "test"},
		{FormatterTimestamped, "mymodule"},
		{FormatterJSON, "json-test"},
		{"unknown", "fallback"}, // Should default to raw
	}

	for _, tc := range tests {
		encoder := CreateEncoder(tc.formatter, tc.module)
		if encoder == nil {
			t.Errorf("CreateEncoder(%q, %q) returned nil", tc.formatter, tc.module)
		}
	}
}

func TestRawEncoder_Clone(t *testing.T) {
	encoder := newRawEncoder()
	cloned := encoder.Clone()

	if cloned == nil {
		t.Error("Clone() returned nil")
	}

	// Both should produce same output
	entry := zapcore.Entry{Message: "test"}

	buf1, _ := encoder.EncodeEntry(entry, nil)
	buf2, _ := cloned.EncodeEntry(entry, nil)

	if buf1.String() != buf2.String() {
		t.Errorf("cloned encoder produced different output: %q vs %q", buf1.String(), buf2.String())
	}
}

func TestTimestampedEncoder_Clone(t *testing.T) {
	encoder := newTimestampedEncoder("module1")
	cloned := encoder.Clone()

	if cloned == nil {
		t.Error("Clone() returned nil")
	}

	// Cloned encoder should preserve module
	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "test",
	}

	buf, _ := cloned.EncodeEntry(entry, nil)
	if !strings.Contains(buf.String(), "module1:") {
		t.Errorf("cloned encoder should preserve module, got %q", buf.String())
	}
}
