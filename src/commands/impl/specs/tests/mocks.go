// Package tests provides BDD step definitions for the specs commands.
//
// This file contains mock setup and cleanup functions for isolated testing.
// It follows the same pattern as the commit command's mocks.go.
package tests

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/commands/impl/specs/create"
	"github.com/ready-to-release/eac/src/commands/impl/specs/validate"
	"github.com/ready-to-release/eac/src/core/git"
)

// SetupSpecsMocks sets up git and AI mocks for isolated testing.
// Called automatically when @env:isolated-test-project tag is present.
func SetupSpecsMocks() error {
	if IsolatedTestProjectDir == "" {
		return nil // Not in isolated mode
	}

	// Validate OriginalRepoRoot is set
	if OriginalRepoRoot == "" {
		return fmt.Errorf("OriginalRepoRoot is not set - cannot copy contracts")
	}

	// Set up mock git repository for both commands
	TestMockRepo = git.NewMockRepository(IsolatedTestProjectDir).
		WithCurrentBranch("main").
		WithHeadSHA("abc1234567890")

	// Inject into both create and validate commands
	create.SetGitRepo(TestMockRepo)
	validate.SetGitRepo(TestMockRepo)

	// Load and set mock AI response for create command
	mockResponsePath := filepath.Join(OriginalRepoRoot,
		"src/commands/impl/specs/tests/assets/mock-response.txt")
	mockResponse, err := os.ReadFile(mockResponsePath)
	if err == nil {
		create.SetMockAIResponse(string(mockResponse))
	}

	// Copy contracts to isolated directory so commands can find them
	if err := copyContractsToIsolatedDir(); err != nil {
		return fmt.Errorf("failed to copy contracts from %s to %s: %w",
			filepath.Join(OriginalRepoRoot, "contracts"),
			filepath.Join(IsolatedTestProjectDir, "contracts"),
			err)
	}

	return nil
}

// copyContractsToIsolatedDir copies the contracts directory to the isolated test directory
func copyContractsToIsolatedDir() error {
	srcContracts := filepath.Join(OriginalRepoRoot, "contracts")
	dstContracts := filepath.Join(IsolatedTestProjectDir, "contracts")

	return copyDir(srcContracts, dstContracts)
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CleanupSpecsMocks resets all mocks.
func CleanupSpecsMocks() {
	create.ResetGitRepo()
	create.ResetMockAIResponse()
	validate.ResetGitRepo()
	TestMockRepo = nil
}
