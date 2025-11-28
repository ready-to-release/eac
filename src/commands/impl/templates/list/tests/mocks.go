package tests

import (
	"path/filepath"

	"github.com/ready-to-release/eac/src/commands/impl/templates/internal"
)

// SetupMocks sets up git mocking for list tests
func SetupMocks() error {
	if listCtx == nil || listCtx.workDir == "" {
		return nil // Not in isolated mode
	}

	// Enable git mocking to prevent actual cloning during tests
	mockTemplatesDir := filepath.Join(listCtx.workDir, "mock-templates")
	internal.EnableMockGit(mockTemplatesDir)

	return nil
}

// CleanupMocks resets all mocks
func CleanupMocks() {
	internal.DisableMockGit()
}
