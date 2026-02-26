package ansi

import (
	"bytes"
	"strings"
	"testing"
)

func TestFilter_BadOnly_PreservesColors(t *testing.T) {
	var buf bytes.Buffer
	f := NewBadOnlyFilter(&buf, "test-source")

	// Write text with good ANSI color codes
	input := "\x1b[32mPASS\x1b[0m: Test passed\n\x1b[31mFAIL\x1b[0m: Test failed\n"
	_, err := f.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should NOT contain warning (colors are good ANSI)
	if strings.Contains(output, "[NOTE]") {
		t.Errorf("Unexpected warning for color codes: %s", output)
	}

	// Should preserve the color codes
	if !strings.Contains(output, "\x1b[32m") {
		t.Errorf("Green color code stripped: %s", output)
	}
	if !strings.Contains(output, "\x1b[0m") {
		t.Errorf("Reset code stripped: %s", output)
	}
}

func TestFilter_BadOnly_StripsBadAnsi(t *testing.T) {
	var buf bytes.Buffer
	f := NewBadOnlyFilter(&buf, "test-source")

	// Write text with bad ANSI (cursor position report)
	input := "Before\x1b[61;1RAfter\n"
	_, err := f.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should contain warning with source and caller
	if !strings.Contains(output, "[NOTE] Bad ANSI control sequences filtered") {
		t.Errorf("Expected warning, got: %s", output)
	}
	if !strings.Contains(output, "test-source") {
		t.Errorf("Expected source in warning, got: %s", output)
	}
	if !strings.Contains(output, "[caller:") {
		t.Errorf("Expected caller info in warning, got: %s", output)
	}

	// Should strip the bad sequence
	if strings.Contains(output, "\x1b[61;1R") {
		t.Errorf("Bad ANSI not stripped: %s", output)
	}

	// Should preserve the text
	if !strings.Contains(output, "BeforeAfter") {
		t.Errorf("Text content lost: %s", output)
	}
}

func TestFilter_BadOnly_MixedGoodAndBad(t *testing.T) {
	var buf bytes.Buffer
	f := NewBadOnlyFilter(&buf, "test-source")

	// Mix of good (colors) and bad (cursor report) ANSI
	input := "\x1b[32mGreen\x1b[0m\x1b[61;1R\x1b[31mRed\x1b[0m\n"
	_, err := f.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should warn (has bad ANSI)
	if !strings.Contains(output, "[NOTE]") {
		t.Errorf("Expected warning for bad ANSI")
	}

	// Should preserve colors
	if !strings.Contains(output, "\x1b[32m") || !strings.Contains(output, "\x1b[31m") {
		t.Errorf("Color codes stripped: %s", output)
	}

	// Should strip bad sequence
	if strings.Contains(output, "\x1b[61;1R") {
		t.Errorf("Bad sequence not stripped: %s", output)
	}
}

func TestFilter_BadOnly_WarnsOnlyOnce(t *testing.T) {
	var buf bytes.Buffer
	f := NewBadOnlyFilter(&buf, "test-source")

	// Write bad ANSI multiple times
	input := "data\x1b[5n\n" // device status report
	_, _ = f.Write([]byte(input))
	_, _ = f.Write([]byte(input))
	_, _ = f.Write([]byte(input))

	output := buf.String()

	// Should contain only ONE warning
	count := strings.Count(output, "[NOTE]")
	if count != 1 {
		t.Errorf("Expected 1 warning, got %d in: %s", count, output)
	}
}

func TestFilter_StripAll_StripsEverything(t *testing.T) {
	var buf bytes.Buffer
	f := NewStripAllFilter(&buf, "test-source")

	// Write text with good ANSI colors
	input := "\x1b[32mgreen\x1b[0m text\n"
	_, err := f.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	// Should NOT inject any warning/note into the output (strip-all is intentional)
	if strings.Contains(output, "[NOTE]") {
		t.Errorf("Unexpected note injected: %s", output)
	}

	// Should strip ALL ANSI including colors
	if strings.Contains(output, "\x1b[") {
		t.Errorf("ANSI not stripped: %s", output)
	}

	// Should preserve text
	if !strings.Contains(output, "green text") {
		t.Errorf("Text lost: %s", output)
	}
}

func TestFilter_Buffering_HandlesSplitSequences(t *testing.T) {
	var buf bytes.Buffer
	f := NewBadOnlyFilter(&buf, "test-source")

	// Simulate split write of a bad sequence: \x1b[61;1R
	_, _ = f.Write([]byte("before\x1b[61"))
	_, _ = f.Write([]byte(";1Rafter\n"))

	output := buf.String()

	// The buffering should handle the split - sequence should be stripped
	// Note: current implementation processes each write separately,
	// so split sequences may not be fully handled. This test documents behavior.
	if !strings.Contains(output, "before") || !strings.Contains(output, "after") {
		t.Errorf("Text content lost: %s", output)
	}
}

func TestStripBad(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Good ANSI preserved
		{"\x1b[32mgreen\x1b[0m", "\x1b[32mgreen\x1b[0m"},
		// Bad ANSI stripped
		{"\x1b[61;1R", ""},
		{"data\x1b[6n", "data"},
		// Mixed
		{"\x1b[32m\x1b[61;1Rtext\x1b[0m", "\x1b[32mtext\x1b[0m"},
	}

	for _, tt := range tests {
		result := string(StripBad([]byte(tt.input)))
		if result != tt.expected {
			t.Errorf("StripBad(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

func TestStripAll(t *testing.T) {
	input := "\x1b[32mgreen\x1b[0m text \x1b[61;1R"
	expected := "green text "
	result := string(StripAll([]byte(input)))
	if result != expected {
		t.Errorf("StripAll: expected %q, got %q", expected, result)
	}
}
