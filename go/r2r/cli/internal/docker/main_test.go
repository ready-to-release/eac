//go:build L0
// +build L0

package docker

import (
	"io"
	"os"
	"testing"

	"github.com/ready-to-release/eac/go/r2r/cli/internal/logging"
)

// TestMain sets up and tears down test environment for all docker package tests.
// It suppresses production logging to prevent log output from being
// misinterpreted as test errors by CI test reporters.
func TestMain(m *testing.M) {
	// Suppress all logging output during tests
	// This prevents production log statements from appearing in test output
	// which can confuse CI test reporters
	logging.SetOutput(io.Discard, io.Discard)

	// Run all tests
	code := m.Run()

	// Reset logging for any cleanup that might need it
	logging.ResetOutput()

	os.Exit(code)
}
