// Package workspace provides unified workspace detection for the EAC CLI.
//
// This package consolidates all workspace/repository root detection logic
// into a single, well-tested implementation with clear precedence rules.
//
// # Detection Precedence (highest to lowest)
//
//  1. R2R_REPO_ROOT - Explicit override (test isolation, CI, scripts)
//  2. R2R_DOCKER_MODE=true - Container mode -> /var/task
//  3. Git detection - Walk up from cwd looking for .git
//  4. Fallback error - No valid workspace found
//
// # Basic Usage
//
//	ws, err := workspace.Detect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Workspace root:", ws.Root)
//	fmt.Println("Detected via:", ws.Source)
//
// # Test Isolation
//
// For tests that need to operate in an isolated directory:
//
//	func TestSomething(t *testing.T) {
//	    tempDir := t.TempDir()
//	    createTestFiles(t, tempDir)
//
//	    cleanup := workspace.ForTesting(t, tempDir)
//	    defer cleanup()
//
//	    // workspace.Detect() now returns tempDir
//	}
//
// # Explicit Mode
//
// For tests that require strict isolation (should fail if env var not set):
//
//	func TestIntegration(t *testing.T) {
//	    workspace.RequireIsolation(t)
//	    // Test will fail immediately if R2R_REPO_ROOT not set
//	}
//
// # Environment Variables
//
//   - R2R_REPO_ROOT: Explicit workspace root override
//   - R2R_PWD: Working directory override (distinct from repo root for worktrees)
//   - R2R_DOCKER_MODE: Container detection flag ("true" when in container)
//   - R2R_CONTAINER_ROOT: Distribution root for tool binaries in containers
package workspace
