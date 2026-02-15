package build

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCommandsBinaryNeedsRebuild(t *testing.T) {
	// Base time: all "old" files use this timestamp.
	// "New" files are set 1 second after the binary.
	baseTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	binaryTime := baseTime
	olderTime := baseTime.Add(-10 * time.Second)
	newerTime := baseTime.Add(10 * time.Second)

	// helper: create a file and set its mtime
	touch := func(t *testing.T, path string, mtime time.Time) {
		t.Helper()
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", dir, err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", path, err)
		}
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatalf("Chtimes %s: %v", path, err)
		}
	}

	// setupWorkspace creates the standard sentinel files in tmp dirs and
	// returns (workspaceRoot, cmdDir, binaryPath). All sentinels default
	// to olderTime so the binary looks up-to-date unless a test overrides.
	setupWorkspace := func(t *testing.T) (string, string, string) {
		t.Helper()
		root := t.TempDir()
		cmdDir := filepath.Join(root, "go", "cli", "eac")
		toolsDir := filepath.Join(root, "out", "tools")

		binaryPath := filepath.Join(toolsDir, "eac")
		touch(t, binaryPath, binaryTime)

		// Sentinel files in cmdDir
		touch(t, filepath.Join(cmdDir, "go.mod"), olderTime)
		touch(t, filepath.Join(cmdDir, "go.sum"), olderTime)

		// Sentinel files in workspace root
		touch(t, filepath.Join(root, "go.work"), olderTime)
		touch(t, filepath.Join(root, "go.work.sum"), olderTime)

		return root, cmdDir, binaryPath
	}

	tests := []struct {
		name       string
		setup      func(t *testing.T) (workspaceRoot, cmdDir, binaryPath string)
		wantRebuild bool
		wantReason  string
	}{
		{
			name: "binary missing returns true",
			setup: func(t *testing.T) (string, string, string) {
				t.Helper()
				root := t.TempDir()
				cmdDir := filepath.Join(root, "go", "cli", "eac")
				touch(t, filepath.Join(cmdDir, "go.mod"), olderTime)
				// binary does not exist
				binaryPath := filepath.Join(root, "out", "tools", "eac")
				return root, cmdDir, binaryPath
			},
			wantRebuild: true,
			wantReason:  "binary missing",
		},
		{
			name: "all sentinels older than binary returns false",
			setup: func(t *testing.T) (string, string, string) {
				t.Helper()
				return setupWorkspace(t)
			},
			wantRebuild: false,
			wantReason:  "",
		},
		{
			name: "go.mod newer than binary returns true",
			setup: func(t *testing.T) (string, string, string) {
				t.Helper()
				root, cmdDir, binaryPath := setupWorkspace(t)
				// Make go.mod newer than the binary
				touch(t, filepath.Join(cmdDir, "go.mod"), newerTime)
				return root, cmdDir, binaryPath
			},
			wantRebuild: true,
			wantReason:  "go.mod changed",
		},
		{
			name: "go.work.sum newer than binary returns true",
			setup: func(t *testing.T) (string, string, string) {
				t.Helper()
				root, cmdDir, binaryPath := setupWorkspace(t)
				// Make go.work.sum newer than the binary
				touch(t, filepath.Join(root, "go.work.sum"), newerTime)
				return root, cmdDir, binaryPath
			},
			wantRebuild: true,
			wantReason:  "go.work.sum changed",
		},
		{
			name: "go.sum newer than binary returns true",
			setup: func(t *testing.T) (string, string, string) {
				t.Helper()
				root, cmdDir, binaryPath := setupWorkspace(t)
				touch(t, filepath.Join(cmdDir, "go.sum"), newerTime)
				return root, cmdDir, binaryPath
			},
			wantRebuild: true,
			wantReason:  "go.sum changed",
		},
		{
			name: "go.work newer than binary returns true",
			setup: func(t *testing.T) (string, string, string) {
				t.Helper()
				root, cmdDir, binaryPath := setupWorkspace(t)
				touch(t, filepath.Join(root, "go.work"), newerTime)
				return root, cmdDir, binaryPath
			},
			wantRebuild: true,
			wantReason:  "go.work changed",
		},
		{
			name: "missing sentinel file does not trigger rebuild",
			setup: func(t *testing.T) (string, string, string) {
				t.Helper()
				root := t.TempDir()
				cmdDir := filepath.Join(root, "go", "cli", "eac")
				toolsDir := filepath.Join(root, "out", "tools")

				binaryPath := filepath.Join(toolsDir, "eac")
				touch(t, binaryPath, binaryTime)

				// Only create go.mod (older), skip the rest entirely
				touch(t, filepath.Join(cmdDir, "go.mod"), olderTime)
				return root, cmdDir, binaryPath
			},
			wantRebuild: false,
			wantReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspaceRoot, cmdDir, binaryPath := tt.setup(t)

			gotRebuild, gotReason := commandsBinaryNeedsRebuild(workspaceRoot, cmdDir, binaryPath)

			if gotRebuild != tt.wantRebuild {
				t.Errorf("commandsBinaryNeedsRebuild() rebuild = %v, want %v", gotRebuild, tt.wantRebuild)
			}
			if gotReason != tt.wantReason {
				t.Errorf("commandsBinaryNeedsRebuild() reason = %q, want %q", gotReason, tt.wantReason)
			}
		})
	}
}
