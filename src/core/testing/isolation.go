// Package testing provides test utilities including isolated test environments.
//
// The TestIsolation type provides a unified way to create isolated test environments
// for BDD/Godog tests that need to operate on a "fake" repository without affecting
// the real repository.
//
// Key features:
//   - Creates temporary directory for test isolation
//   - Sets R2R_REPO_ROOT and R2R_PWD environment variables (no physical .git needed)
//   - Provides hooks for mock injection (git repository, AI responses)
//   - Handles cleanup automatically
//
// Usage:
//
//	isolation := testing.NewTestIsolation().
//	    WithOriginalRepoRoot(originalRoot).
//	    WithCopyContracts(true)
//
//	if err := isolation.Setup(); err != nil {
//	    return err
//	}
//	defer isolation.Cleanup()
//
//	// Access the isolated directory
//	testDir := isolation.IsolatedDir()
//
//	// Get environment variables to pass to subprocesses
//	env := isolation.Environment()
package testing

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// TestIsolation provides an isolated test environment for BDD tests.
// It creates a temporary directory and sets environment variables so that
// repository discovery returns the isolated directory instead of the real repo.
type TestIsolation struct {
	// Configuration (set before Setup)
	originalRepoRoot string
	copyContracts    bool
	copySpecs        bool

	// State (set by Setup)
	isolatedDir string
	cleanedUp   bool
}

// NewTestIsolation creates a new test isolation builder.
func NewTestIsolation() *TestIsolation {
	return &TestIsolation{}
}

// WithOriginalRepoRoot sets the original repository root path.
// This is needed for copying contracts and other resources to the isolated dir.
func (t *TestIsolation) WithOriginalRepoRoot(root string) *TestIsolation {
	t.originalRepoRoot = root
	return t
}

// WithCopyContracts enables copying the contracts directory to the isolated dir.
// This is needed for commands that validate against contracts.
func (t *TestIsolation) WithCopyContracts(copy bool) *TestIsolation {
	t.copyContracts = copy
	return t
}

// WithCopySpecs enables copying the specs directory to the isolated dir.
func (t *TestIsolation) WithCopySpecs(copy bool) *TestIsolation {
	t.copySpecs = copy
	return t
}

// Setup creates the isolated test environment.
// Call Cleanup() when done to remove the temporary directory.
func (t *TestIsolation) Setup() error {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "isolated-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	t.isolatedDir = tmpDir

	// Copy contracts if requested
	if t.copyContracts && t.originalRepoRoot != "" {
		srcContracts := filepath.Join(t.originalRepoRoot, "contracts")
		dstContracts := filepath.Join(t.isolatedDir, "contracts")
		if err := copyDir(srcContracts, dstContracts); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to copy contracts: %w", err)
		}
	}

	// Copy specs if requested
	if t.copySpecs && t.originalRepoRoot != "" {
		srcSpecs := filepath.Join(t.originalRepoRoot, "specs")
		dstSpecs := filepath.Join(t.isolatedDir, "specs")
		if err := copyDir(srcSpecs, dstSpecs); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to copy specs: %w", err)
		}
	}

	return nil
}

// Cleanup removes the temporary directory.
// Safe to call multiple times.
func (t *TestIsolation) Cleanup() {
	if t.cleanedUp || t.isolatedDir == "" {
		return
	}
	os.RemoveAll(t.isolatedDir)
	t.cleanedUp = true
	t.isolatedDir = ""
}

// IsolatedDir returns the path to the isolated test directory.
// Returns empty string if Setup hasn't been called.
func (t *TestIsolation) IsolatedDir() string {
	return t.isolatedDir
}

// Environment returns environment variables that should be set for subprocesses
// to use the isolated directory as the repository root.
func (t *TestIsolation) Environment() []string {
	if t.isolatedDir == "" {
		return nil
	}
	return []string{
		fmt.Sprintf("R2R_PWD=%s", t.isolatedDir),
		fmt.Sprintf("R2R_REPO_ROOT=%s", t.isolatedDir),
	}
}

// AppendToEnvironment appends the isolation environment variables to an existing
// environment slice (typically from os.Environ()).
func (t *TestIsolation) AppendToEnvironment(env []string) []string {
	if t.isolatedDir == "" {
		return env
	}
	return append(env, t.Environment()...)
}

// copyDir recursively copies a directory.
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

// copyFile copies a single file.
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
