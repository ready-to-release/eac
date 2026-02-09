// Package show contains godog step implementations for eac.
//
// This file contains show-workspaces command step definitions.
package show

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// registerSteps registers step definitions for show-workspaces command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Setup steps
	sc.Step(`^I have multiple worktrees:$`, func(table *godog.Table) error {
		return setupMultipleWorktrees(ctx, table)
	})
	sc.Step(`^I have multiple worktrees$`, func() error {
		// Default worktrees for verbose/debug tests
		return setupDefaultWorktrees(ctx)
	})
	sc.Step(`^I have a worktree at "([^"]*)" with no uncommitted changes$`, func(path string) error {
		return setupCleanWorktree(ctx, path)
	})
	sc.Step(`^I have a worktree at "([^"]*)" with uncommitted changes$`, func(path string) error {
		return setupDirtyWorktree(ctx, path)
	})
	sc.Step(`^I have only the main worktree$`, func() error {
		// No additional setup needed - isolated git repo starts with just main
		return ensureGitRepository(ctx)
	})
	sc.Step(`^I am in a git repository with no additional worktrees$`, func() error {
		// Same as "I have only the main worktree"
		return ensureGitRepository(ctx)
	})
	// Note: "I am in a git repository" step is already registered in eacgodog/steps.go
	// Note: "I am not in a git repository" step is already registered in eacgodog/steps.go
	// Don't register them again to avoid ambiguous step definitions

	// Verification steps
	sc.Step(`^I see a table with headers "([^"]*)", "([^"]*)", "([^"]*)"$`, func(h1, h2, h3 string) error {
		return verifyTableHeaders(ctx, h1, h2, h3)
	})
	// Note: "I see X worktrees total" is handled by the generic ^I see "([^"]*)"$ step in steps_work.go
	sc.Step(`^I see "([^"]*)" with status "([^"]*)"$`, func(branch, status string) error {
		return verifyBranchStatus(ctx, branch, status)
	})
	sc.Step(`^I see commit SHA for each worktree$`, func() error {
		return verifyCommitSHAPresent(ctx)
	})
	sc.Step(`^debug logs are written to "([^"]*)"$`, func(logPath string) error {
		return verifyDebugLogs(ctx, logPath)
	})
	sc.Step(`^debug logs contain worktree information$`, func() error {
		return verifyDebugLogsContainWorktreeInfo(ctx)
	})
	sc.Step(`^I see the main worktree listed$`, func() error {
		return verifyMainWorktreeListed(ctx)
	})
}

// setupMultipleWorktrees creates multiple worktrees from a table.
// Table columns: path, branch, clean (true/false).
func setupMultipleWorktrees(ctx *eacgodog.TestContext, table *godog.Table) error {
	ctx.MustBeIsolated()

	// First, ensure we have a clean main worktree
	if err := ensureGitRepository(ctx); err != nil {
		return err
	}

	// Create an initial commit if needed
	if err := ensureInitialCommit(ctx); err != nil {
		return err
	}

	// Process each row in the table (skip header)
	baseName := filepath.Base(ctx.IsolatedDir)
	worktreeIndex := 0
	for i, row := range table.Rows {
		if i == 0 {
			continue // Skip header row
		}

		if len(row.Cells) < 3 {
			return fmt.Errorf("table row %d has %d cells, expected 3 (path, branch, clean)", i, len(row.Cells))
		}

		branch := row.Cells[1].Value
		cleanStr := row.Cells[2].Value

		// Determine if this is the main worktree (first data row)
		isMain := (i == 1)
		if isMain {
			// For main worktree, just verify it matches - no need to create
			continue
		}

		// Create worktree with unique path based on isolated dir
		worktreeIndex++
		worktreePath := filepath.Join(filepath.Dir(ctx.IsolatedDir), fmt.Sprintf("%s-wt%d", baseName, worktreeIndex))

		// Create worktree
		if err := createWorktree(ctx, worktreePath, branch); err != nil {
			return fmt.Errorf("failed to create worktree %s: %w", worktreePath, err)
		}

		// Make dirty if needed
		if cleanStr == "false" {
			if err := makeDirtyWorktree(ctx, worktreePath); err != nil {
				return fmt.Errorf("failed to make worktree dirty: %w", err)
			}
		}
	}

	return nil
}

// setupDefaultWorktrees creates a default set of worktrees for testing.
func setupDefaultWorktrees(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()

	if err := ensureGitRepository(ctx); err != nil {
		return err
	}

	if err := ensureInitialCommit(ctx); err != nil {
		return err
	}

	// Create two additional worktrees with unique paths
	// Use the isolated dir name to make paths unique across scenarios
	baseName := filepath.Base(ctx.IsolatedDir)
	featurePath := filepath.Join(filepath.Dir(ctx.IsolatedDir), baseName+"-feature")
	if err := createWorktree(ctx, featurePath, "feature/test"); err != nil {
		return err
	}

	bugfixPath := filepath.Join(filepath.Dir(ctx.IsolatedDir), baseName+"-bugfix")
	if err := createWorktree(ctx, bugfixPath, "bugfix/test"); err != nil {
		return err
	}

	return nil
}

// setupCleanWorktree creates a worktree with no uncommitted changes.
func setupCleanWorktree(ctx *eacgodog.TestContext, relPath string) error {
	ctx.MustBeIsolated()

	if err := ensureGitRepository(ctx); err != nil {
		return err
	}

	if err := ensureInitialCommit(ctx); err != nil {
		return err
	}

	// Convert relative path to absolute with unique name
	baseName := filepath.Base(ctx.IsolatedDir)
	absPath := filepath.Join(filepath.Dir(ctx.IsolatedDir), baseName+"-clean-auth")
	return createWorktree(ctx, absPath, "feature/auth")
}

// setupDirtyWorktree creates a worktree with uncommitted changes.
func setupDirtyWorktree(ctx *eacgodog.TestContext, relPath string) error {
	ctx.MustBeIsolated()

	if err := ensureGitRepository(ctx); err != nil {
		return err
	}

	if err := ensureInitialCommit(ctx); err != nil {
		return err
	}

	// Convert relative path to absolute with unique name
	baseName := filepath.Base(ctx.IsolatedDir)
	absPath := filepath.Join(filepath.Dir(ctx.IsolatedDir), baseName+"-dirty-auth")

	if err := createWorktree(ctx, absPath, "feature/auth"); err != nil {
		return err
	}

	// Add uncommitted changes
	return makeDirtyWorktree(ctx, absPath)
}

// ensureGitRepository verifies we're in a git repository.
func ensureGitRepository(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()

	// Verify .git exists
	gitDir := filepath.Join(ctx.IsolatedDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("git repository not initialized")
	}

	return nil
}

// ensureInitialCommit ensures the repository has at least one commit.
func ensureInitialCommit(ctx *eacgodog.TestContext) error {
	// Check if we have any commits
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = ctx.IsolatedDir
	if err := cmd.Run(); err != nil {
		// No commits yet, create initial commit
		testFile := filepath.Join(ctx.IsolatedDir, "README.md")
		if err := os.WriteFile(testFile, []byte("# Test Repository\n"), 0o644); err != nil {
			return fmt.Errorf("failed to create test file: %w", err)
		}

		if err := runGitCommand(ctx.IsolatedDir, "add", "README.md"); err != nil {
			return err
		}

		if err := runGitCommand(ctx.IsolatedDir, "commit", "-m", "Initial commit"); err != nil {
			return err
		}
	}

	return nil
}

// createWorktree creates a git worktree at the specified path.
func createWorktree(ctx *eacgodog.TestContext, path, branch string) error {
	// Create branch if it doesn't exist
	cmd := exec.Command("git", "branch", branch)
	cmd.Dir = ctx.IsolatedDir
	_ = cmd.Run() //nolint:errcheck // Ignore error - branch might already exist

	// Create worktree directory
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("failed to create worktree parent directory: %w", err)
	}

	// Add worktree
	if err := runGitCommand(ctx.IsolatedDir, "worktree", "add", path, branch); err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}

	return nil
}

// makeDirtyWorktree adds uncommitted changes to a worktree.
func makeDirtyWorktree(ctx *eacgodog.TestContext, path string) error {
	testFile := filepath.Join(path, "uncommitted.txt")
	return os.WriteFile(testFile, []byte("uncommitted change\n"), 0o644)
}

// verifyTableHeaders checks that the output contains the expected table headers.
func verifyTableHeaders(ctx *eacgodog.TestContext, h1, h2, h3 string) error {
	for _, header := range []string{h1, h2, h3} {
		if !strings.Contains(ctx.CommandOutput, header) {
			return fmt.Errorf("expected header %q not found in output", header)
		}
	}
	return nil
}

// verifyBranchStatus checks that a branch appears with the expected status.
func verifyBranchStatus(ctx *eacgodog.TestContext, branch, status string) error {
	output := ctx.CommandOutput

	// Look for lines containing both the branch and status
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, branch) && strings.Contains(line, status) {
			return nil
		}
	}

	return fmt.Errorf("expected to find branch %q with status %q in output:\n%s", branch, status, output)
}

// verifyCommitSHAPresent checks that commit SHAs are shown in the output.
func verifyCommitSHAPresent(ctx *eacgodog.TestContext) error {
	output := ctx.CommandOutput

	if !strings.Contains(output, "SHA") {
		return fmt.Errorf("expected SHA column in output, got:\n%s", output)
	}

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "---") || strings.Contains(line, "SHA") {
			continue
		}

		for _, field := range strings.Fields(line) {
			field = strings.Trim(field, "|")
			field = strings.TrimSpace(field)
			if isHexLike(field) {
				return nil
			}
		}
	}

	return fmt.Errorf("expected to find commit SHA in output, got:\n%s", output)
}

// isHexLike checks if a string looks like a hex SHA.
func isHexLike(s string) bool {
	if len(s) < 7 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// verifyDebugLogs checks that debug logs were written to the expected file.
func verifyDebugLogs(ctx *eacgodog.TestContext, logPath string) error {
	ctx.MustBeIsolated()

	// Construct absolute log path
	absLogFile := filepath.Join(ctx.IsolatedDir, logPath)

	// Check if log file exists
	fileInfo, err := os.Stat(absLogFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("debug log file %q does not exist", absLogFile)
	}
	if err != nil {
		return fmt.Errorf("failed to stat debug log file: %w", err)
	}
	if fileInfo.Size() == 0 {
		return fmt.Errorf("debug log file %q is empty", absLogFile)
	}

	return nil
}

// verifyDebugLogsContainWorktreeInfo checks that debug logs contain worktree information.
func verifyDebugLogsContainWorktreeInfo(ctx *eacgodog.TestContext) error {
	ctx.MustBeIsolated()

	logFile := filepath.Join(ctx.IsolatedDir, "out", "commands.log")

	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("failed to read debug log: %w", err)
	}

	logContent := string(content)

	// Check for expected log messages
	if !strings.Contains(logContent, "worktree") && !strings.Contains(logContent, "Worktree") {
		return fmt.Errorf("debug log does not contain worktree information:\n%s", logContent)
	}

	return nil
}

// verifyMainWorktreeListed checks that the main worktree is listed in output.
func verifyMainWorktreeListed(ctx *eacgodog.TestContext) error {
	output := ctx.CommandOutput

	// Look for the main branch in output
	if !strings.Contains(output, "main") {
		return fmt.Errorf("expected main branch in output, got:\n%s", output)
	}

	return nil
}

// runGitCommand runs a git command in the specified directory.
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v failed: %w\nOutput: %s", args, err, string(output))
	}
	return nil
}
