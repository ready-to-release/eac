package gitexec_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/clibase/gitexec"
)

func TestRun_Success(t *testing.T) {
	// Create temp directory with git repo
	tempDir := t.TempDir()

	// Initialize git repo
	_, err := gitexec.Run(tempDir, "init")
	if err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Configure git user for the test repo
	_, _ = gitexec.Run(tempDir, "config", "user.email", "test@example.com")
	_, _ = gitexec.Run(tempDir, "config", "user.name", "Test User")

	// Create and commit a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = gitexec.Run(tempDir, "add", "test.txt")
	if err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	_, err = gitexec.Run(tempDir, "commit", "-m", "initial commit")
	if err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Now test rev-parse HEAD
	output, err := gitexec.Run(tempDir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse HEAD failed: %v", err)
	}

	sha := strings.TrimSpace(string(output))
	if len(sha) != 40 {
		t.Errorf("expected 40 character SHA, got %d: %s", len(sha), sha)
	}
}

func TestRun_Error(t *testing.T) {
	// Run in non-git directory
	tempDir := t.TempDir()

	_, err := gitexec.Run(tempDir, "rev-parse", "HEAD")
	if err == nil {
		t.Error("expected error for rev-parse in non-git directory")
	}
	if !strings.Contains(err.Error(), "not a git repository") &&
		!strings.Contains(err.Error(), "exit") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunCombined_NonZeroExit(t *testing.T) {
	// Create temp directory with git repo
	tempDir := t.TempDir()

	_, err := gitexec.Run(tempDir, "init")
	if err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Try to verify a nonexistent ref - should return non-zero exit without error
	_, exitCode, err := gitexec.RunCombined(context.Background(), tempDir,
		"show-ref", "--verify", "refs/heads/nonexistent")

	if err != nil {
		t.Fatalf("RunCombined should not return execution error: %v", err)
	}
	if exitCode == 0 {
		t.Error("expected non-zero exit code for nonexistent ref")
	}
}

func TestRunSilent(t *testing.T) {
	// Create temp directory with git repo
	tempDir := t.TempDir()

	// Initialize should succeed silently
	err := gitexec.RunSilent(context.Background(), tempDir, "init")
	if err != nil {
		t.Fatalf("git init via RunSilent failed: %v", err)
	}

	// Verify repo was created
	if _, err := os.Stat(filepath.Join(tempDir, ".git")); os.IsNotExist(err) {
		t.Error("expected .git directory to exist")
	}
}

func TestRunContext_Cancellation(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize git repo first
	_, err := gitexec.Run(tempDir, "init")
	if err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Running with cancelled context should fail
	_, err = gitexec.RunContext(ctx, tempDir, "status")
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
