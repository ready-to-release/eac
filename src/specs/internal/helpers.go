// Package internal provides shared helpers for godog BDD tests.
//
// This file contains helper functions that can be used by step implementations
// across different feature domains (verbs). These are NOT step registrations -
// they are composable building blocks that step implementations can call.
//
// Design Principles:
// 1. Helpers are pure functions that take context and return errors
// 2. Step registrations happen in each verb's steps.go, not here
// 3. Two verbs can share a helper without creating step conflicts
// 4. Helpers operate on TestContext to access shared state
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ready-to-release/eac/src/core/repository"
)

// ============================================================================
// Asset Loading Helpers
// ============================================================================

// LoadAsset loads content from an asset file in src/specs/impl/src-commands/assets/.
// The assetPath is relative to the assets directory (e.g., "specs/valid-spec.txt").
// Uses container root if in container, otherwise repo root.
func LoadAsset(ctx *TestContext, assetPath string) (string, error) {
	assetsRoot := repository.GetEffectiveRoot(ctx.OriginalRepoRoot)
	fullPath := filepath.Join(assetsRoot, "src", "specs", "impl", "src-commands", "assets", assetPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to load asset %s: %w", assetPath, err)
	}
	return string(data), nil
}

// MustLoadAsset loads an asset and panics on failure (for use in test setup).
func MustLoadAsset(ctx *TestContext, assetPath string) string {
	content, err := LoadAsset(ctx, assetPath)
	if err != nil {
		panic(err)
	}
	return content
}

// ============================================================================
// Output Verification Helpers
// ============================================================================

// OutputContains checks if the command output contains the given text.
func OutputContains(ctx *TestContext, text string) error {
	if !strings.Contains(ctx.CommandOutput, text) {
		return fmt.Errorf("expected output to contain '%s', got:\n%s", text, ctx.CommandOutput)
	}
	return nil
}

// OutputDoesNotContain checks that the command output does not contain the given text.
func OutputDoesNotContain(ctx *TestContext, text string) error {
	if strings.Contains(ctx.CommandOutput, text) {
		return fmt.Errorf("expected output not to contain '%s', got:\n%s", text, ctx.CommandOutput)
	}
	return nil
}

// OutputMatches checks if the command output matches the given regex pattern.
func OutputMatches(ctx *TestContext, pattern string) error {
	matched, err := regexp.MatchString(pattern, ctx.CommandOutput)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}
	if !matched {
		return fmt.Errorf("output does not match pattern '%s', got:\n%s", pattern, ctx.CommandOutput)
	}
	return nil
}

// OutputContainsAny checks if the command output contains any of the given texts.
func OutputContainsAny(ctx *TestContext, texts ...string) error {
	for _, text := range texts {
		if strings.Contains(ctx.CommandOutput, text) {
			return nil
		}
	}
	return fmt.Errorf("expected output to contain one of %v, got:\n%s", texts, ctx.CommandOutput)
}

// ============================================================================
// Exit Code Helpers
// ============================================================================

// ExitCodeIs checks if the exit code matches the expected value.
func ExitCodeIs(ctx *TestContext, expected int) error {
	if ctx.ExitCode != expected {
		return fmt.Errorf("expected exit code %d, got %d. Output:\n%s",
			expected, ctx.ExitCode, ctx.CommandOutput)
	}
	return nil
}

// ExitCodeIsOneOf checks if the exit code matches any of the expected values.
func ExitCodeIsOneOf(ctx *TestContext, codes ...int) error {
	for _, code := range codes {
		if ctx.ExitCode == code {
			return nil
		}
	}
	return fmt.Errorf("expected exit code to be one of %v, got %d. Output:\n%s",
		codes, ctx.ExitCode, ctx.CommandOutput)
}

// CommandSucceeded checks if the command exited with code 0.
func CommandSucceeded(ctx *TestContext) error {
	return ExitCodeIs(ctx, 0)
}

// CommandFailed checks if the command exited with a non-zero code.
func CommandFailed(ctx *TestContext) error {
	if ctx.ExitCode == 0 {
		return fmt.Errorf("expected command to fail, got exit code 0. Output:\n%s",
			ctx.CommandOutput)
	}
	return nil
}

// ============================================================================
// File Verification Helpers
// ============================================================================

// FileExists checks if a file exists, respecting isolation context.
func FileExists(ctx *TestContext, path string) error {
	fullPath := ResolvePath(ctx, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fullPath)
	}
	return nil
}

// FileContains checks if a file contains the given text.
func FileContains(ctx *TestContext, path, content string) error {
	fullPath := ResolvePath(ctx, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}
	if !strings.Contains(string(data), content) {
		return fmt.Errorf("file %s does not contain '%s'. Content:\n%s", path, content, string(data))
	}
	return nil
}

// DirectoryHasFiles checks if a directory contains at least one file.
func DirectoryHasFiles(ctx *TestContext, dir string) error {
	fullPath := ResolvePath(ctx, dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", fullPath, err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no files found in %s", fullPath)
	}
	return nil
}

// ============================================================================
// File Creation Helpers
// ============================================================================

// CreateFile creates a file with the given content, respecting isolation context.
func CreateFile(ctx *TestContext, path, content string) error {
	fullPath := ResolvePath(ctx, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// CreateDirectory creates a directory, respecting isolation context.
func CreateDirectory(ctx *TestContext, dir string) error {
	fullPath := ResolvePath(ctx, dir)
	return os.MkdirAll(fullPath, 0755)
}

// RemoveAll removes a file or directory.
// SAFETY: This ONLY works in isolated test directories to prevent accidental data loss.
func RemoveAll(ctx *TestContext, path string) error {
	// CRITICAL: Only allow this in isolated test environments
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("SAFETY: RemoveAll can only be used in isolated test environments")
	}

	fullPath := ResolvePath(ctx, path)

	// Verify the path is within the isolated directory
	if !strings.HasPrefix(fullPath, ctx.IsolatedDir) {
		return fmt.Errorf("SAFETY: path %q is not within isolated dir %q", fullPath, ctx.IsolatedDir)
	}

	// Don't allow removing the isolated dir itself
	if fullPath == ctx.IsolatedDir {
		return fmt.Errorf("SAFETY: cannot remove the isolated directory itself")
	}

	return os.RemoveAll(fullPath)
}

// ============================================================================
// Path Helpers
// ============================================================================

// ResolvePath resolves a path relative to the isolated directory if set,
// otherwise returns the path as-is.
func ResolvePath(ctx *TestContext, path string) string {
	if ctx.IsolatedDir != "" && !filepath.IsAbs(path) {
		return filepath.Join(ctx.IsolatedDir, path)
	}
	return path
}

// ============================================================================
// Custom Prompt/Template Helpers
// ============================================================================

// CustomPromptWasUsed checks for evidence of custom prompt usage in debug logs.
// This is a shared helper that both specs and risks can use.
func CustomPromptWasUsed(ctx *TestContext, debugSubdir string) error {
	if ctx.IsolatedDir == "" {
		// Can't verify in non-isolated mode
		return nil
	}
	debugDir := filepath.Join(ctx.IsolatedDir, "out", "logs", debugSubdir)
	if _, err := os.Stat(debugDir); os.IsNotExist(err) {
		return fmt.Errorf("no debug logs found at %s to verify custom prompt usage", debugDir)
	}
	return nil
}

// ============================================================================
// Git Helpers
// ============================================================================

// IsGitRepository ensures we're operating in a git repository context.
// When in isolation, this creates a minimal .git directory to satisfy commands
// that check for git repository presence.
// When not in isolation, checks OriginalRepoRoot (the actual repo).
func IsGitRepository(ctx *TestContext) error {
	var gitDir string
	if ctx.IsolatedDir != "" {
		// In isolation - create .git if it doesn't exist
		gitDir = filepath.Join(ctx.IsolatedDir, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			// Create minimal .git directory for isolated tests
			if err := os.MkdirAll(gitDir, 0755); err != nil {
				return fmt.Errorf("failed to create .git directory in isolation: %w", err)
			}
			// Create minimal git config file
			configPath := filepath.Join(gitDir, "config")
			configContent := "[core]\n\trepositoryformatversion = 0\n\tbare = false\n"
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				return fmt.Errorf("failed to create .git/config: %w", err)
			}
			// Create HEAD file pointing to main
			headPath := filepath.Join(gitDir, "HEAD")
			if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
				return fmt.Errorf("failed to create .git/HEAD: %w", err)
			}
		}
	} else {
		// Not in isolation - check the original repo root
		gitDir = filepath.Join(ctx.OriginalRepoRoot, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("not a git repository: .git directory not found")
		}
	}
	return nil
}

// EnsureNotGitRepository ensures we're not in a git repository.
// SAFETY: This ONLY works in isolated test directories and only removes
// minimal .git dirs created by IsGitRepository (with only config/HEAD).
func EnsureNotGitRepository(ctx *TestContext) error {
	// CRITICAL: Only allow this in isolated test environments
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("SAFETY: EnsureNotGitRepository can only be used in isolated test environments")
	}

	// Double-check we're actually in the isolated directory
	resolvedPath := ResolvePath(ctx, ".")
	if resolvedPath != ctx.IsolatedDir {
		return fmt.Errorf("SAFETY: resolved path %q does not match isolated dir %q", resolvedPath, ctx.IsolatedDir)
	}

	gitDir := filepath.Join(ctx.IsolatedDir, ".git")
	info, err := os.Stat(gitDir)
	if os.IsNotExist(err) {
		return nil // Already not a git repo
	}
	if err != nil {
		return err
	}

	// Only remove if it's a minimal test .git directory
	if info.IsDir() {
		entries, err := os.ReadDir(gitDir)
		if err != nil {
			return err
		}
		// Allow removal if empty OR only has config/HEAD (test-created minimal .git)
		if len(entries) == 0 {
			return os.Remove(gitDir)
		}
		// Check if only config and HEAD exist (test-created)
		isTestGit := len(entries) <= 2
		for _, e := range entries {
			if e.Name() != "config" && e.Name() != "HEAD" {
				isTestGit = false
				break
			}
		}
		if isTestGit {
			return os.RemoveAll(gitDir) // Safe to remove test-created .git
		}
		return fmt.Errorf("SAFETY: .git directory has %d entries and appears to be a real git repo - refusing to remove", len(entries))
	}

	return fmt.Errorf("SAFETY: .git is not a directory")
}
