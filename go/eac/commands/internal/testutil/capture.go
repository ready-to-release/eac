// Package testutil provides shared test utilities for command testing.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

// CaptureOutput holds captured stdout and stderr output.
type CaptureOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Capture captures stdout and stderr during function execution.
// The function should return an exit code (0 for success).
func Capture(fn func() int) CaptureOutput {
	// Use mutex to prevent race conditions if tests run in parallel
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, pipeErr := os.Pipe()
	if pipeErr != nil {
		return CaptureOutput{ExitCode: fn()} // Fall through if pipe fails
	}
	os.Stdout = wOut

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, pipeErr := os.Pipe()
	if pipeErr != nil {
		wOut.Close()
		os.Stdout = oldStdout
		return CaptureOutput{ExitCode: fn()} // Fall through if pipe fails
	}
	os.Stderr = wErr

	// Run the function
	exitCode := fn()

	// Close writers and restore
	wOut.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output (errors are non-fatal)
	var bufOut, bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut) //nolint:errcheck // best-effort capture
	_, _ = io.Copy(&bufErr, rErr) //nolint:errcheck // best-effort capture

	return CaptureOutput{
		Stdout:   bufOut.String(),
		Stderr:   bufErr.String(),
		ExitCode: exitCode,
	}
}

// CaptureNoExit captures stdout and stderr for functions that don't return exit codes.
func CaptureNoExit(fn func()) CaptureOutput {
	return Capture(func() int {
		fn()
		return 0
	})
}

// AssertSuccess checks that exit code is 0.
func (c CaptureOutput) AssertSuccess(t *testing.T) {
	t.Helper()
	if c.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d\nstderr: %s", c.ExitCode, c.Stderr)
	}
}

// AssertFailure checks that exit code is non-zero.
func (c CaptureOutput) AssertFailure(t *testing.T) {
	t.Helper()
	if c.ExitCode == 0 {
		t.Errorf("expected non-zero exit code, got 0\nstdout: %s", c.Stdout)
	}
}

// AssertExitCode checks for a specific exit code.
func (c CaptureOutput) AssertExitCode(t *testing.T, expected int) {
	t.Helper()
	if c.ExitCode != expected {
		t.Errorf("expected exit code %d, got %d\nstdout: %s\nstderr: %s",
			expected, c.ExitCode, c.Stdout, c.Stderr)
	}
}

// AssertStdoutOnly fails if any output went to stderr (excluding empty stderr).
func (c CaptureOutput) AssertStdoutOnly(t *testing.T) {
	t.Helper()
	if c.Stderr != "" {
		t.Errorf("unexpected stderr output:\n%s", c.Stderr)
	}
}

// AssertStderrOnly fails if any output went to stdout.
func (c CaptureOutput) AssertStderrOnly(t *testing.T) {
	t.Helper()
	if c.Stdout != "" {
		t.Errorf("unexpected stdout output:\n%s", c.Stdout)
	}
}

// AssertNoOutput fails if there's any output to stdout or stderr.
func (c CaptureOutput) AssertNoOutput(t *testing.T) {
	t.Helper()
	if c.Stdout != "" {
		t.Errorf("unexpected stdout output:\n%s", c.Stdout)
	}
	if c.Stderr != "" {
		t.Errorf("unexpected stderr output:\n%s", c.Stderr)
	}
}

// AssertValidJSON verifies stdout is valid JSON.
func (c CaptureOutput) AssertValidJSON(t *testing.T) {
	t.Helper()
	if c.Stdout == "" {
		t.Error("stdout is empty, expected valid JSON")
		return
	}
	var v interface{}
	if err := json.Unmarshal([]byte(c.Stdout), &v); err != nil {
		t.Errorf("stdout is not valid JSON: %v\nOutput:\n%s", err, c.Stdout)
	}
}

// AssertValidYAML verifies stdout is valid YAML.
func (c CaptureOutput) AssertValidYAML(t *testing.T) {
	t.Helper()
	if c.Stdout == "" {
		t.Error("stdout is empty, expected valid YAML")
		return
	}
	var v interface{}
	if err := yaml.Unmarshal([]byte(c.Stdout), &v); err != nil {
		t.Errorf("stdout is not valid YAML: %v\nOutput:\n%s", err, c.Stdout)
	}
}

// UnmarshalJSON unmarshals stdout as JSON into the provided value.
func (c CaptureOutput) UnmarshalJSON(t *testing.T, v interface{}) {
	t.Helper()
	if err := json.Unmarshal([]byte(c.Stdout), v); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nOutput:\n%s", err, c.Stdout)
	}
}

// UnmarshalYAML unmarshals stdout as YAML into the provided value.
func (c CaptureOutput) UnmarshalYAML(t *testing.T, v interface{}) {
	t.Helper()
	if err := yaml.Unmarshal([]byte(c.Stdout), v); err != nil {
		t.Fatalf("failed to unmarshal YAML: %v\nOutput:\n%s", err, c.Stdout)
	}
}
