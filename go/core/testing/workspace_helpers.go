// Package testing provides workspace test helpers alongside TestIsolation.
//
// This file adds lightweight helpers for unit tests that need workspace isolation
// without the full BDD/Godog infrastructure of TestIsolation.
//
// Helper Selection Guide:
//   - SetupWorkspaceIsolation: Unit tests reading real repo files (read-only)
//   - SetupTempWorkspaceIsolation: Unit tests creating/modifying files
//   - CopyToTempWorkspace: Unit tests modifying copies of real files
//   - TestIsolation: BDD/Godog tests with full git repo
package testing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/workspace"
)

// SetupWorkspaceIsolation sets up test isolation pointing to the real repo.
// This is the standard pattern for tests that read (but don't modify) repo files.
//
// Replaces the common 8-line boilerplate pattern with 1 line.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    repoRoot := coretesting.SetupWorkspaceIsolation(t)
//	    // Test runs against real repository files (read-only)
//	}
func SetupWorkspaceIsolation(t testing.TB) string {
	t.Helper()

	ws, err := workspace.DetectWithOptions(workspace.Options{
		Mode: workspace.ModeGitOnly,
	})
	if err != nil {
		t.Fatalf("failed to detect workspace: %v", err)
	}

	workspace.ForTesting(t, ws.Root)
	workspace.RequireIsolation(t)

	return ws.Root
}

// SetupTempWorkspaceIsolation sets up test isolation with a temporary directory.
// Use for tests that create/modify files to avoid polluting the real repo.
//
// The temporary directory is automatically cleaned up when the test completes.
//
// Usage:
//
//	func TestModifyFiles(t *testing.T) {
//	    tempRoot := coretesting.SetupTempWorkspaceIsolation(t)
//	    // Create/modify test files in tempRoot
//	}
func SetupTempWorkspaceIsolation(t testing.TB) string {
	t.Helper()

	tempDir := t.TempDir() // Automatically cleaned up

	// Create minimal workspace markers
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	workspace.ForTesting(t, tempDir)
	workspace.RequireIsolation(t)

	return tempDir
}

// RealRepoRoot returns the actual repository root using git detection.
// Useful when you need the real root before setting up isolation.
//
// This function does NOT set up isolation - it only returns the path.
// Use SetupWorkspaceIsolation if you need both detection and isolation.
func RealRepoRoot(t testing.TB) string {
	t.Helper()

	ws, err := workspace.DetectWithOptions(workspace.Options{
		Mode: workspace.ModeGitOnly,
	})
	if err != nil {
		t.Fatalf("failed to detect real repo root: %v", err)
	}
	return ws.Root
}

// CopyToTempWorkspace copies files from real repo to an isolated temp workspace.
// Useful for tests that need to modify copies of real files.
//
// The temporary directory is automatically cleaned up when the test completes.
//
// Usage:
//
//	func TestModifyChangelog(t *testing.T) {
//	    tempRoot := coretesting.CopyToTempWorkspace(t, []string{
//	        "release/ext-eac/CHANGELOG.md",
//	        ".eac/repository.yml",
//	    })
//	    // Modify files in tempRoot safely
//	}
func CopyToTempWorkspace(t testing.TB, relativePaths []string) string {
	t.Helper()

	realRoot := RealRepoRoot(t)
	tempRoot := t.TempDir()

	// Create .git marker
	if err := os.MkdirAll(filepath.Join(tempRoot, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	// Copy specified files
	for _, relPath := range relativePaths {
		srcPath := filepath.Join(realRoot, relPath)
		dstPath := filepath.Join(tempRoot, relPath)

		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", relPath, err)
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			t.Fatalf("failed to copy %s: %v", relPath, err)
		}
	}

	workspace.ForTesting(t, tempRoot)
	workspace.RequireIsolation(t)

	return tempRoot
}
