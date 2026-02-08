//go:build L1

package testing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestIsolation_Setup(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TestIsolation)
		wantErr bool
	}{
		{
			name:    "basic setup creates temp directory",
			setup:   func(ti *TestIsolation) {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := NewTestIsolation()
			tt.setup(ti)

			err := ti.Setup()
			if (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer ti.Cleanup()

				// Verify directory was created
				if ti.IsolatedDir() == "" {
					t.Error("IsolatedDir() returned empty string after Setup()")
				}
				if _, err := os.Stat(ti.IsolatedDir()); os.IsNotExist(err) {
					t.Error("Isolated directory does not exist")
				}
			}
		})
	}
}

func TestTestIsolation_GitDirectory(t *testing.T) {
	ti := NewTestIsolation()
	if err := ti.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	defer ti.Cleanup()

	// Verify .git directory IS created - required for tests that run git commands
	gitDir := filepath.Join(ti.IsolatedDir(), ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error(".git directory should be created for isolated test environments")
	}
}

func TestTestIsolation_Environment(t *testing.T) {
	ti := NewTestIsolation()
	if err := ti.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	defer ti.Cleanup()

	env := ti.Environment()
	if len(env) != 2 {
		t.Errorf("Environment() returned %d items, want 2", len(env))
	}

	var hasPWD, hasRepoRoot bool
	for _, e := range env {
		if strings.HasPrefix(e, "CLIE_PWD=") {
			hasPWD = true
			if !strings.Contains(e, ti.IsolatedDir()) {
				t.Error("CLIE_PWD does not contain isolated dir")
			}
		}
		if strings.HasPrefix(e, "CLIE_REPO_ROOT=") {
			hasRepoRoot = true
			if !strings.Contains(e, ti.IsolatedDir()) {
				t.Error("CLIE_REPO_ROOT does not contain isolated dir")
			}
		}
	}

	if !hasPWD {
		t.Error("Environment() missing CLIE_PWD")
	}
	if !hasRepoRoot {
		t.Error("Environment() missing CLIE_REPO_ROOT")
	}
}

func TestTestIsolation_AppendToEnvironment(t *testing.T) {
	ti := NewTestIsolation()
	if err := ti.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	defer ti.Cleanup()

	existingEnv := []string{"PATH=/usr/bin", "HOME=/home/test"}
	newEnv := ti.AppendToEnvironment(existingEnv)

	if len(newEnv) != 4 {
		t.Errorf("AppendToEnvironment() returned %d items, want 4", len(newEnv))
	}

	// Original env should be preserved
	if newEnv[0] != "PATH=/usr/bin" || newEnv[1] != "HOME=/home/test" {
		t.Error("AppendToEnvironment() did not preserve existing environment")
	}
}

func TestTestIsolation_Cleanup(t *testing.T) {
	ti := NewTestIsolation()
	if err := ti.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}

	dir := ti.IsolatedDir()

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatal("Directory should exist before cleanup")
	}

	ti.Cleanup()

	// Verify directory was removed
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("Directory should not exist after cleanup")
	}

	// Verify IsolatedDir returns empty after cleanup
	if ti.IsolatedDir() != "" {
		t.Error("IsolatedDir() should return empty string after cleanup")
	}

	// Verify cleanup is idempotent (safe to call multiple times)
	ti.Cleanup() // Should not panic
}

func TestTestIsolation_CleanupWithoutSetup(t *testing.T) {
	ti := NewTestIsolation()
	// Calling Cleanup without Setup should not panic
	ti.Cleanup()
}

func TestTestIsolation_CopyContracts(t *testing.T) {
	// Create a temp "original repo" with contracts
	origRepo, err := os.MkdirTemp("", "orig-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(origRepo)

	// Create repository config directory with a test file
	repoConfigDir := filepath.Join(origRepo, ".eac", "repository")
	if err := os.MkdirAll(repoConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create repository config dir: %v", err)
	}
	testFile := filepath.Join(repoConfigDir, "test.yml")
	if err := os.WriteFile(testFile, []byte("test: true"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create isolation with contract copying
	ti := NewTestIsolation().
		WithOriginalRepoRoot(origRepo).
		WithCopyContracts(true)

	if err := ti.Setup(); err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	defer ti.Cleanup()

	// Verify repository config was copied
	copiedFile := filepath.Join(ti.IsolatedDir(), ".eac", "repository", "test.yml")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Error("Contracts file was not copied to isolated directory")
	}
}

func TestTestIsolation_EnvironmentBeforeSetup(t *testing.T) {
	ti := NewTestIsolation()
	// Before Setup, Environment should return nil
	env := ti.Environment()
	if env != nil {
		t.Errorf("Environment() before Setup() should return nil, got %v", env)
	}
}
