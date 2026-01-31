package ansi

import (
	"bytes"
	"strings"
	"testing"
)

func TestStrippingWriter_StripsCodes(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStrippingWriter(&buf, "test-source")

	// Write text with ANSI codes
	input := "\x1b[32mPASS\x1b[0m: Test passed\n\x1b[31mFAIL\x1b[0m: Test failed\n"
	_, err := sw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should contain warning
	if !strings.Contains(output, "[WARNING] ANSI escape codes detected") {
		t.Errorf("Expected warning, got: %s", output)
	}

	// Should contain source
	if !strings.Contains(output, "test-source") {
		t.Errorf("Expected source in warning, got: %s", output)
	}

	// Should have stripped ANSI codes
	if strings.Contains(output, "\x1b[") {
		t.Errorf("ANSI codes not stripped: %s", output)
	}

	// Should contain the actual text
	if !strings.Contains(output, "PASS: Test passed") {
		t.Errorf("Expected 'PASS: Test passed', got: %s", output)
	}
	if !strings.Contains(output, "FAIL: Test failed") {
		t.Errorf("Expected 'FAIL: Test failed', got: %s", output)
	}
}

func TestStrippingWriter_NoWarningForCleanInput(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStrippingWriter(&buf, "test-source")

	// Write clean text (no ANSI)
	input := "PASS: Test passed\nFAIL: Test failed\n"
	_, err := sw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should NOT contain warning
	if strings.Contains(output, "[WARNING]") {
		t.Errorf("Unexpected warning for clean input: %s", output)
	}

	// Should be identical to input
	if output != input {
		t.Errorf("Expected %q, got %q", input, output)
	}
}

func TestStrippingWriter_WarnsOnlyOnce(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStrippingWriter(&buf, "test-source")

	// Write ANSI codes multiple times
	input := "\x1b[32mGreen\x1b[0m\n"
	_, _ = sw.Write([]byte(input))
	_, _ = sw.Write([]byte(input))
	_, _ = sw.Write([]byte(input))

	output := buf.String()

	// Should contain only ONE warning
	count := strings.Count(output, "[WARNING]")
	if count != 1 {
		t.Errorf("Expected 1 warning, got %d in: %s", count, output)
	}
}

func TestStrip(t *testing.T) {
	input := []byte("\x1b[1;31mERROR\x1b[0m: Something failed")
	expected := []byte("ERROR: Something failed")

	result := Strip(input)
	if !bytes.Equal(result, expected) {
		t.Errorf("Strip: expected %q, got %q", expected, result)
	}
}

func TestStripString(t *testing.T) {
	input := "\x1b[32m✓\x1b[0m PASS"
	expected := "✓ PASS"

	result := StripString(input)
	if result != expected {
		t.Errorf("StripString: expected %q, got %q", expected, result)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		input    []byte
		expected bool
	}{
		{[]byte("\x1b[32mgreen\x1b[0m"), true},
		{[]byte("plain text"), false},
		{[]byte("\x1b[1;31;40mcomplex\x1b[0m"), true},
		{[]byte(""), false},
	}

	for _, tt := range tests {
		result := Contains(tt.input)
		if result != tt.expected {
			t.Errorf("Contains(%q): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}
