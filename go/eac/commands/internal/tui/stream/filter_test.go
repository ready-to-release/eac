package stream

import "testing"

func TestFilter_ShouldShow_EmptyLines(t *testing.T) {
	f := NewFilter()

	if f.ShouldShow("") {
		t.Error("empty string should not be shown")
	}
	if f.ShouldShow("   ") {
		t.Error("whitespace-only string should not be shown")
	}
}

func TestFilter_ShouldShow_ImportantPatterns(t *testing.T) {
	f := NewFilter()

	// These should always be shown
	importantLines := []string{
		"error: something went wrong",
		"Error: Something Went Wrong",
		"build failed",
		"FAIL test/package",
		"warning: deprecated function",
		"panic: runtime error",
		"fatal: unable to proceed",
		"undefined: SomeType",
		"cannot find module",
		"package not found",
	}

	for _, line := range importantLines {
		if !f.ShouldShow(line) {
			t.Errorf("important line should be shown: %q", line)
		}
	}
}

func TestFilter_ShouldShow_NoisePatterns(t *testing.T) {
	f := NewFilter()

	// These should be hidden
	noiseLines := []string{
		"go: downloading github.com/some/package v1.0.0",
		"=== RUN   TestSomething",
		"=== PAUSE TestSomething",
		"=== CONT  TestSomething",
		"--- PASS: TestSomething (0.00s)",
		"PASS",
		"ok  \tgithub.com/package\t0.123s",
		"coverage: 80.5% of statements",
		"?   \tgithub.com/pkg\t[no test files]",
	}

	for _, line := range noiseLines {
		if f.ShouldShow(line) {
			t.Errorf("noise line should not be shown: %q", line)
		}
	}
}

func TestFilter_ShouldShow_NormalLines(t *testing.T) {
	f := NewFilter()

	// Normal output lines should be shown
	normalLines := []string{
		"Building module eac-core...",
		"Compiling main.go",
		"Running tests...",
		"Test completed successfully",
	}

	for _, line := range normalLines {
		if !f.ShouldShow(line) {
			t.Errorf("normal line should be shown: %q", line)
		}
	}
}

func TestFilter_Deduplication(t *testing.T) {
	f := NewFilter()

	line := "some repeated line"

	// First few occurrences should be shown
	for i := 0; i < 2; i++ {
		if !f.ShouldShow(line) {
			t.Errorf("line occurrence %d should be shown", i+1)
		}
	}

	// After too many repetitions, should be hidden
	// Push it past the threshold
	for i := 0; i < 5; i++ {
		f.ShouldShow(line)
	}

	if f.ShouldShow(line) {
		t.Error("heavily repeated line should eventually be hidden")
	}
}

func TestFilter_Reset(t *testing.T) {
	f := NewFilter()

	line := "some line"

	// Fill up deduplication
	for i := 0; i < 10; i++ {
		f.ShouldShow(line)
	}

	// Reset
	f.Reset()

	// Should be shown again after reset
	if !f.ShouldShow(line) {
		t.Error("line should be shown after reset")
	}
}
