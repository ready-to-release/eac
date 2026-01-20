package locking

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireAndRelease(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "locking-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := Config{
		BaseDir:      "locks",
		Identifier:   "test-module",
		ResourceType: "module",
		ActionVerb:   "already being built",
	}

	// First acquisition should succeed
	lock, err := Acquire(tempDir, cfg)
	if err != nil {
		t.Fatalf("first acquisition failed: %v", err)
	}
	if lock == nil {
		t.Fatal("lock should not be nil after successful acquisition")
	}

	// Verify lock file exists
	lockPath := filepath.Join(tempDir, cfg.BaseDir, ".lock-"+cfg.Identifier)
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("lock file should exist after acquisition")
	}

	// Second acquisition should fail
	_, err = Acquire(tempDir, cfg)
	if err == nil {
		t.Error("second acquisition should fail while lock is held")
	}

	// Release the lock
	Release(lock)

	// Lock file should be removed after release
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("lock file should be removed after release")
	}

	// Third acquisition should succeed after release
	lock2, err := Acquire(tempDir, cfg)
	if err != nil {
		t.Fatalf("third acquisition failed after release: %v", err)
	}
	Release(lock2)
}

func TestBuildConfig(t *testing.T) {
	cfg := BuildConfig("eac-core", "out/build")

	if cfg.Identifier != "eac-core" {
		t.Errorf("expected identifier 'eac-core', got '%s'", cfg.Identifier)
	}
	if cfg.BaseDir != "out/build" {
		t.Errorf("expected baseDir 'out/build', got '%s'", cfg.BaseDir)
	}
	if cfg.ResourceType != "module" {
		t.Errorf("expected resourceType 'module', got '%s'", cfg.ResourceType)
	}
}

func TestTestConfig(t *testing.T) {
	cfg := TestConfig("unit", "out/test")

	if cfg.Identifier != "unit" {
		t.Errorf("expected identifier 'unit', got '%s'", cfg.Identifier)
	}
	if cfg.BaseDir != "out/test" {
		t.Errorf("expected baseDir 'out/test', got '%s'", cfg.BaseDir)
	}
	if cfg.ResourceType != "test suite" {
		t.Errorf("expected resourceType 'test suite', got '%s'", cfg.ResourceType)
	}
}

func TestReleaseNil(t *testing.T) {
	// Should not panic
	Release(nil)
}

func TestManualTestFileConfig(t *testing.T) {
	testCases := []struct {
		name       string
		filePath   string
		baseDir    string
		wantIdent  string
		wantBase   string
	}{
		{
			name:      "simple filename",
			filePath:  "results.json",
			baseDir:   "out/test",
			wantIdent: "results.json",
			wantBase:  "out/test",
		},
		{
			name:      "full path",
			filePath:  "out/test/manual-results.json",
			baseDir:   "out/test",
			wantIdent: "manual-results.json",
			wantBase:  "out/test",
		},
		{
			name:      "nested path",
			filePath:  "/path/to/test/results/manual-test-1.json",
			baseDir:   "out/test",
			wantIdent: "manual-test-1.json",
			wantBase:  "out/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ManualTestFileConfig(tc.filePath, tc.baseDir)

			if cfg.Identifier != tc.wantIdent {
				t.Errorf("identifier: got %q, want %q", cfg.Identifier, tc.wantIdent)
			}
			if cfg.BaseDir != tc.wantBase {
				t.Errorf("baseDir: got %q, want %q", cfg.BaseDir, tc.wantBase)
			}
			if cfg.ResourceType != "manual test file" {
				t.Errorf("resourceType: got %q, want %q", cfg.ResourceType, "manual test file")
			}
			if cfg.ActionVerb != "already being written" {
				t.Errorf("actionVerb: got %q, want %q", cfg.ActionVerb, "already being written")
			}
		})
	}
}
