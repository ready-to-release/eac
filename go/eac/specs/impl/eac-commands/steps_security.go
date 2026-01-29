// Package srccommands contains godog step implementations for specs/eac-commands.
//
// This file contains security command step definitions.
// Features: specs/eac-commands/security/
//
// This implements step definitions for security scanner features including:
// - Evidence file verification
// - JSON schema validation
// - Log file checks
// - Security scanner output validation
package srccommands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
)

// securityTestState holds state for security tests.
type securityTestState struct {
	lastCheckedDirectory string
}

// registerSecuritySteps registers step definitions for security command features.
func registerSecuritySteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	state := &securityTestState{}

	// Evidence file steps
	sc.Step(`^evidence files should exist in directory "([^"]*)"$`, func(directory string) error {
		return evidenceFilesExistInDirectory(directory, ctx, state)
	})
}

// ============================================================================
// Evidence File Verification Steps
// ============================================================================

// evidenceFilesExistInDirectory checks that evidence files exist in the specified directory.
func evidenceFilesExistInDirectory(directory string, ctx *internal.TestContext, state *securityTestState) error {
	// Save the directory for use in subsequent steps
	state.lastCheckedDirectory = directory

	// Use isolated test directory if available, otherwise use repository root
	workspaceRoot := ctx.IsolatedDir
	if workspaceRoot == "" {
		root, err := repository.GetRepositoryRoot("")
		if err != nil {
			return fmt.Errorf("failed to get workspace root: %w", err)
		}
		workspaceRoot = root
	}

	fullPath := filepath.Join(workspaceRoot, directory)

	// Check if directory exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", fullPath)
	}

	// Check for JSON files in directory
	files, err := filepath.Glob(filepath.Join(fullPath, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no evidence files found in directory: %s", fullPath)
	}

	return nil
}

