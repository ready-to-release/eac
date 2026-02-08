// Package stream provides output streaming utilities for the TUI console.
package stream

import (
	"testing"

	tuicontract "github.com/ready-to-release/eac/contracts/tui/0.1.0"
)

func TestClassifyLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected tuicontract.Level
	}{
		// Should be Info (summary lines - not errors)
		{"summary with 0 errors", "Tests: 10 passed, 0 errors", tuicontract.LevelInfo},
		{"summary with 0 failed", "Tests: 10 passed, 0 failed", tuicontract.LevelInfo},
		{"build completed", "Build completed successfully", tuicontract.LevelInfo},
		{"packages passed", "Packages: 5 passed", tuicontract.LevelInfo},
		{"go test run marker", "=== RUN   TestSomething", tuicontract.LevelInfo},
		{"go test pass marker", "--- PASS: TestSomething (0.00s)", tuicontract.LevelInfo},
		{"no errors found", "Validation: no errors found", tuicontract.LevelInfo},

		// Should be Info (test names containing "error" - false positive prevention)
		{"pass with error in name", "--- PASS: TestIsSummaryLine/error:_something (0.00s)", tuicontract.LevelInfo},
		{"pass with Error in name", "--- PASS: TestValidateError (0.00s)", tuicontract.LevelInfo},
		{"run with error in name", "=== RUN   TestErrorHandling/error:_case", tuicontract.LevelInfo},

		// Should be Error (actual failures)
		{"go test fail", "--- FAIL: TestSomething (0.00s)", tuicontract.LevelError},
		{"package fail", "FAIL\tgithub.com/example/pkg", tuicontract.LevelError},
		{"panic", "panic: runtime error", tuicontract.LevelError},
		{"fatal error", "fatal error: out of memory", tuicontract.LevelError},
		{"compilation failed", "compilation failed: syntax error", tuicontract.LevelError},
		{"undefined symbol", "undefined: someFunction", tuicontract.LevelError},
		{"build failed", "build failed: missing dependency", tuicontract.LevelError},
		{"error with colon", "main.go:10: error: undefined variable", tuicontract.LevelError},
		{"standalone error", "error: something went wrong", tuicontract.LevelError},

		// Should be Warn (test output containing error messages - not failures)
		{"test output with Error", "scenarios_test.go:1280: Error: extensions.0.env.0.name: invalid", tuicontract.LevelWarn},
		{"test output with error colon", "validation_test.go:50: error: missing field", tuicontract.LevelWarn},

		// Should be Warn (warnings)
		{"go test skip", "--- SKIP: TestSomething (0.00s)", tuicontract.LevelWarn},
		{"warning message", "warning: deprecated function", tuicontract.LevelWarn},
		{"deprecated usage", "this API is deprecated", tuicontract.LevelWarn},

		// Regular info lines
		{"normal output", "Running tests...", tuicontract.LevelInfo},
		{"ok package", "ok  \tgithub.com/example/pkg\t0.123s", tuicontract.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyLine(tt.input)
			if got != tt.expected {
				t.Errorf("classifyLine(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsSummaryLine(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Summary lines
		{"tests: 10 passed, 0 failed", true},
		{"packages: 5 passed", true},
		{"total: 100 tests", true},
		{"0 errors found", true},
		{"0 failed tests", true},
		{"no errors", true},
		{"=== run   testsomething", true},
		{"=== test package", true},
		{"build completed", true},
		{"tests succeeded", true},

		// Not summary lines
		{"error: something went wrong", false},
		{"--- fail: testexample", false},
		{"panic occurred", false},
		{"running test...", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isSummaryLine(tt.input)
			if got != tt.expected {
				t.Errorf("isSummaryLine(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
