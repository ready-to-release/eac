package npm

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --------------------------------------------------------------------------
// NewNpmIsolation
// --------------------------------------------------------------------------

func TestNewNpmIsolation(t *testing.T) {
	root := t.TempDir()
	iso := NewNpmIsolation(root)

	if iso.workspaceRoot != root {
		t.Errorf("workspaceRoot = %q, want %q", iso.workspaceRoot, root)
	}
	if iso.workRoot == "" {
		t.Error("workRoot should not be empty")
	}
	if iso.npmCacheDir == "" {
		t.Error("npmCacheDir should not be empty")
	}
	// Both paths should live under the workspace root's cache
	if !filepath.IsAbs(iso.workRoot) {
		// The paths module may return relative or absolute; just check non-empty
		t.Logf("workRoot is relative: %s", iso.workRoot)
	}
}

// --------------------------------------------------------------------------
// copyFileIfChanged
// --------------------------------------------------------------------------

func TestCopyFileIfChanged(t *testing.T) {
	t.Run("copies file when destination does not exist", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		src := filepath.Join(srcDir, "file.txt")
		dst := filepath.Join(dstDir, "file.txt")

		writeFile(t, src, "hello world")

		copied, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !copied {
			t.Error("expected file to be copied, got false")
		}

		assertFileContent(t, dst, "hello world")
	})

	t.Run("skips copy when destination is up to date", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		src := filepath.Join(srcDir, "file.txt")
		dst := filepath.Join(dstDir, "file.txt")

		writeFile(t, src, "hello world")

		// First copy
		_, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("first copy failed: %v", err)
		}

		// Second copy should be a no-op
		copied, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("second copy failed: %v", err)
		}
		if copied {
			t.Error("expected skip (no change), but file was copied")
		}
	})

	t.Run("re-copies when source content changes", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		src := filepath.Join(srcDir, "file.txt")
		dst := filepath.Join(dstDir, "file.txt")

		writeFile(t, src, "version 1")

		// First copy
		_, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("first copy failed: %v", err)
		}

		// Modify source -- different content AND different size triggers re-copy
		writeFile(t, src, "version 2 with more content")
		// Ensure mtime is different
		future := time.Now().Add(2 * time.Second)
		if err := os.Chtimes(src, future, future); err != nil {
			t.Fatalf("chtimes failed: %v", err)
		}

		copied, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("second copy failed: %v", err)
		}
		if !copied {
			t.Error("expected file to be re-copied after source change")
		}

		assertFileContent(t, dst, "version 2 with more content")
	})

	t.Run("re-copies when mtime changes even if size is same", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		src := filepath.Join(srcDir, "file.txt")
		dst := filepath.Join(dstDir, "file.txt")

		writeFile(t, src, "AAAA")

		// First copy
		_, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("first copy failed: %v", err)
		}

		// Same size, different mtime
		writeFile(t, src, "BBBB")
		future := time.Now().Add(5 * time.Second)
		if err := os.Chtimes(src, future, future); err != nil {
			t.Fatalf("chtimes failed: %v", err)
		}

		copied, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("second copy failed: %v", err)
		}
		if !copied {
			t.Error("expected file to be re-copied when mtime differs")
		}
		assertFileContent(t, dst, "BBBB")
	})

	t.Run("returns error when source does not exist", func(t *testing.T) {
		dstDir := t.TempDir()
		src := filepath.Join(t.TempDir(), "nonexistent.txt")
		dst := filepath.Join(dstDir, "file.txt")

		_, err := copyFileIfChanged(src, dst)
		if err == nil {
			t.Error("expected error for missing source, got nil")
		}
	})

	t.Run("preserves modification time on destination", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		src := filepath.Join(srcDir, "file.txt")
		dst := filepath.Join(dstDir, "file.txt")

		writeFile(t, src, "preserve mtime")
		mtime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
		if err := os.Chtimes(src, mtime, mtime); err != nil {
			t.Fatalf("chtimes failed: %v", err)
		}

		_, err := copyFileIfChanged(src, dst)
		if err != nil {
			t.Fatalf("copy failed: %v", err)
		}

		dstInfo, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("stat dst failed: %v", err)
		}
		if !dstInfo.ModTime().Equal(mtime) {
			t.Errorf("dst mtime = %v, want %v", dstInfo.ModTime(), mtime)
		}
	})
}

// --------------------------------------------------------------------------
// packageFilesChanged
// --------------------------------------------------------------------------

func TestPackageFilesChanged(t *testing.T) {
	t.Run("returns true when destination does not exist", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := filepath.Join(t.TempDir(), "nonexistent")

		writeFile(t, filepath.Join(srcDir, "package.json"), `{"name":"test"}`)

		if !packageFilesChanged(srcDir, dstDir) {
			t.Error("expected true when destination dir does not exist")
		}
	})

	t.Run("returns false when both files match", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		content := `{"name":"test","version":"1.0.0"}`
		mtime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		for _, dir := range []string{srcDir, dstDir} {
			writeFile(t, filepath.Join(dir, "package.json"), content)
			if err := os.Chtimes(filepath.Join(dir, "package.json"), mtime, mtime); err != nil {
				t.Fatal(err)
			}
			writeFile(t, filepath.Join(dir, "package-lock.json"), content)
			if err := os.Chtimes(filepath.Join(dir, "package-lock.json"), mtime, mtime); err != nil {
				t.Fatal(err)
			}
		}

		if packageFilesChanged(srcDir, dstDir) {
			t.Error("expected false when files match exactly")
		}
	})

	t.Run("returns true when package.json size differs", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()
		mtime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		writeFile(t, filepath.Join(srcDir, "package.json"), `{"name":"updated","version":"2.0.0"}`)
		if err := os.Chtimes(filepath.Join(srcDir, "package.json"), mtime, mtime); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(dstDir, "package.json"), `{"name":"old"}`)
		if err := os.Chtimes(filepath.Join(dstDir, "package.json"), mtime, mtime); err != nil {
			t.Fatal(err)
		}

		if !packageFilesChanged(srcDir, dstDir) {
			t.Error("expected true when package.json size differs")
		}
	})

	t.Run("returns true when package-lock.json mtime differs", func(t *testing.T) {
		srcDir := t.TempDir()
		dstDir := t.TempDir()

		content := `{"lockfileVersion": 2}`
		writeFile(t, filepath.Join(srcDir, "package-lock.json"), content)
		writeFile(t, filepath.Join(dstDir, "package-lock.json"), content)

		// Set same mtime on package.json to pass that check
		pkgContent := `{"name":"test"}`
		mtime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		writeFile(t, filepath.Join(srcDir, "package.json"), pkgContent)
		writeFile(t, filepath.Join(dstDir, "package.json"), pkgContent)
		setMtime(t, filepath.Join(srcDir, "package.json"), mtime)
		setMtime(t, filepath.Join(dstDir, "package.json"), mtime)

		// Different mtime on lock file
		setMtime(t, filepath.Join(srcDir, "package-lock.json"), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))
		setMtime(t, filepath.Join(dstDir, "package-lock.json"), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

		if !packageFilesChanged(srcDir, dstDir) {
			t.Error("expected true when package-lock.json mtime differs")
		}
	})

	t.Run("returns false when neither file exists in source", func(t *testing.T) {
		srcDir := t.TempDir() // empty
		dstDir := t.TempDir() // empty

		if packageFilesChanged(srcDir, dstDir) {
			t.Error("expected false when source has no package files")
		}
	})
}

// --------------------------------------------------------------------------
// syncDirectory
// --------------------------------------------------------------------------

func TestSyncDirectory(t *testing.T) {
	t.Run("copies all files from source to destination", func(t *testing.T) {
		srcDir := filepath.Join(t.TempDir(), "src")
		dstDir := filepath.Join(t.TempDir(), "dst")
		mkdirAll(t, srcDir)

		writeFile(t, filepath.Join(srcDir, "a.ts"), "const a = 1;")
		writeFile(t, filepath.Join(srcDir, "b.ts"), "const b = 2;")

		if err := syncDirectory(srcDir, dstDir); err != nil {
			t.Fatalf("syncDirectory failed: %v", err)
		}

		assertFileContent(t, filepath.Join(dstDir, "a.ts"), "const a = 1;")
		assertFileContent(t, filepath.Join(dstDir, "b.ts"), "const b = 2;")
	})

	t.Run("copies nested subdirectories", func(t *testing.T) {
		srcDir := filepath.Join(t.TempDir(), "src")
		dstDir := filepath.Join(t.TempDir(), "dst")
		mkdirAll(t, filepath.Join(srcDir, "sub", "deep"))

		writeFile(t, filepath.Join(srcDir, "top.ts"), "top")
		writeFile(t, filepath.Join(srcDir, "sub", "mid.ts"), "mid")
		writeFile(t, filepath.Join(srcDir, "sub", "deep", "bottom.ts"), "bottom")

		if err := syncDirectory(srcDir, dstDir); err != nil {
			t.Fatalf("syncDirectory failed: %v", err)
		}

		assertFileContent(t, filepath.Join(dstDir, "top.ts"), "top")
		assertFileContent(t, filepath.Join(dstDir, "sub", "mid.ts"), "mid")
		assertFileContent(t, filepath.Join(dstDir, "sub", "deep", "bottom.ts"), "bottom")
	})

	t.Run("removes files from destination that are no longer in source", func(t *testing.T) {
		srcDir := filepath.Join(t.TempDir(), "src")
		dstDir := filepath.Join(t.TempDir(), "dst")
		mkdirAll(t, srcDir)
		mkdirAll(t, dstDir)

		// Initial sync with two files
		writeFile(t, filepath.Join(srcDir, "keep.ts"), "keep")
		writeFile(t, filepath.Join(srcDir, "remove.ts"), "remove")

		if err := syncDirectory(srcDir, dstDir); err != nil {
			t.Fatalf("first sync failed: %v", err)
		}

		// Remove one file from source
		if err := os.Remove(filepath.Join(srcDir, "remove.ts")); err != nil {
			t.Fatalf("remove failed: %v", err)
		}

		// Sync again
		if err := syncDirectory(srcDir, dstDir); err != nil {
			t.Fatalf("second sync failed: %v", err)
		}

		// "keep.ts" should still exist
		assertFileContent(t, filepath.Join(dstDir, "keep.ts"), "keep")

		// "remove.ts" should be gone
		if _, err := os.Stat(filepath.Join(dstDir, "remove.ts")); !os.IsNotExist(err) {
			t.Error("expected remove.ts to be deleted from destination")
		}
	})

	t.Run("returns error when source does not exist", func(t *testing.T) {
		srcDir := filepath.Join(t.TempDir(), "nonexistent")
		dstDir := filepath.Join(t.TempDir(), "dst")

		err := syncDirectory(srcDir, dstDir)
		if err == nil {
			t.Error("expected error for nonexistent source, got nil")
		}
	})

	t.Run("returns error when source is a file not a directory", func(t *testing.T) {
		srcFile := filepath.Join(t.TempDir(), "file.txt")
		dstDir := filepath.Join(t.TempDir(), "dst")

		writeFile(t, srcFile, "not a directory")

		err := syncDirectory(srcFile, dstDir)
		if err == nil {
			t.Error("expected error when source is a file, got nil")
		}
	})

	t.Run("handles empty source directory", func(t *testing.T) {
		srcDir := filepath.Join(t.TempDir(), "src")
		dstDir := filepath.Join(t.TempDir(), "dst")
		mkdirAll(t, srcDir)

		if err := syncDirectory(srcDir, dstDir); err != nil {
			t.Fatalf("syncDirectory failed on empty source: %v", err)
		}

		// Destination directory should exist but be empty
		entries, err := os.ReadDir(dstDir)
		if err != nil {
			t.Fatalf("ReadDir failed: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty destination, got %d entries", len(entries))
		}
	})
}

// --------------------------------------------------------------------------
// PrepareIsolatedEnv
// --------------------------------------------------------------------------

func TestPrepareIsolatedEnv(t *testing.T) {
	t.Run("derives isolation key from outputDir basename", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "my-module")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)

		env, err := iso.PrepareIsolatedEnv(moduleRoot, "/some/path/unique-key")
		if err != nil {
			t.Fatalf("PrepareIsolatedEnv failed: %v", err)
		}

		// WorkDir should end with the isolation key
		if filepath.Base(env.WorkDir) != "unique-key" {
			t.Errorf("WorkDir basename = %q, want %q", filepath.Base(env.WorkDir), "unique-key")
		}
	})

	t.Run("falls back to moduleRoot basename when outputDir is empty", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "my-module")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)

		env, err := iso.PrepareIsolatedEnv(moduleRoot, "")
		if err != nil {
			t.Fatalf("PrepareIsolatedEnv failed: %v", err)
		}

		if filepath.Base(env.WorkDir) != "my-module" {
			t.Errorf("WorkDir basename = %q, want %q", filepath.Base(env.WorkDir), "my-module")
		}
	})

	t.Run("falls back to moduleRoot basename when outputDir is dot", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "fallback-module")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)

		env, err := iso.PrepareIsolatedEnv(moduleRoot, ".")
		if err != nil {
			t.Fatalf("PrepareIsolatedEnv failed: %v", err)
		}

		if filepath.Base(env.WorkDir) != "fallback-module" {
			t.Errorf("WorkDir basename = %q, want %q", filepath.Base(env.WorkDir), "fallback-module")
		}
	})

	t.Run("copies required and optional files", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)

		// Required
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)
		writeFile(t, filepath.Join(moduleRoot, "package-lock.json"), `{"lockfileVersion":2}`)
		// Optional
		writeFile(t, filepath.Join(moduleRoot, "tsconfig.json"), `{"compilerOptions":{}}`)
		writeFile(t, filepath.Join(moduleRoot, ".mocharc.json"), `{"spec":"test/**/*.ts"}`)
		writeFile(t, filepath.Join(moduleRoot, "cucumber.js"), `module.exports = {}`)

		env, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/test/mod/mydir")
		if err != nil {
			t.Fatalf("PrepareIsolatedEnv failed: %v", err)
		}

		// Required files must be present
		assertFileContent(t, filepath.Join(env.WorkDir, "package.json"), `{"name":"test"}`)
		assertFileContent(t, filepath.Join(env.WorkDir, "package-lock.json"), `{"lockfileVersion":2}`)

		// Optional files should be present
		assertFileContent(t, filepath.Join(env.WorkDir, "tsconfig.json"), `{"compilerOptions":{}}`)
		assertFileContent(t, filepath.Join(env.WorkDir, ".mocharc.json"), `{"spec":"test/**/*.ts"}`)
		assertFileContent(t, filepath.Join(env.WorkDir, "cucumber.js"), `module.exports = {}`)
	})

	t.Run("succeeds when package-lock.json is missing", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)
		// No package-lock.json

		_, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/key")
		if err != nil {
			t.Fatalf("expected success without package-lock.json, got: %v", err)
		}
	})

	t.Run("fails when package.json is missing", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "empty-mod")
		mkdirAll(t, moduleRoot)
		// No package.json at all

		_, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/key")
		if err == nil {
			t.Error("expected error when package.json is missing, got nil")
		}
	})

	t.Run("syncs source directories", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)

		// Create directories that should be synced
		mkdirAll(t, filepath.Join(moduleRoot, "src"))
		mkdirAll(t, filepath.Join(moduleRoot, "test"))
		mkdirAll(t, filepath.Join(moduleRoot, "features"))
		mkdirAll(t, filepath.Join(moduleRoot, "steps"))

		writeFile(t, filepath.Join(moduleRoot, "src", "index.ts"), "export const x = 1;")
		writeFile(t, filepath.Join(moduleRoot, "test", "index.test.ts"), "describe('x', () => {});")
		writeFile(t, filepath.Join(moduleRoot, "features", "login.feature"), "Feature: Login")
		writeFile(t, filepath.Join(moduleRoot, "steps", "login.steps.ts"), "Given('I login', () => {});")

		env, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/test/mod/key")
		if err != nil {
			t.Fatalf("PrepareIsolatedEnv failed: %v", err)
		}

		assertFileContent(t, filepath.Join(env.WorkDir, "src", "index.ts"), "export const x = 1;")
		assertFileContent(t, filepath.Join(env.WorkDir, "test", "index.test.ts"), "describe('x', () => {});")
		assertFileContent(t, filepath.Join(env.WorkDir, "features", "login.feature"), "Feature: Login")
		assertFileContent(t, filepath.Join(env.WorkDir, "steps", "login.steps.ts"), "Given('I login', () => {});")
	})

	t.Run("sets NPM_CONFIG_CACHE in environment", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)

		env, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/key")
		if err != nil {
			t.Fatalf("PrepareIsolatedEnv failed: %v", err)
		}

		found := false
		for _, e := range env.Env {
			if len(e) > len("NPM_CONFIG_CACHE=") && e[:len("NPM_CONFIG_CACHE=")] == "NPM_CONFIG_CACHE=" {
				found = true
				val := e[len("NPM_CONFIG_CACHE="):]
				if val != iso.npmCacheDir {
					t.Errorf("NPM_CONFIG_CACHE = %q, want %q", val, iso.npmCacheDir)
				}
				break
			}
		}
		if !found {
			t.Error("NPM_CONFIG_CACHE not found in environment")
		}
	})

	t.Run("sets SourceRoot to moduleRoot", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)

		env, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/key")
		if err != nil {
			t.Fatalf("PrepareIsolatedEnv failed: %v", err)
		}

		if env.SourceRoot != moduleRoot {
			t.Errorf("SourceRoot = %q, want %q", env.SourceRoot, moduleRoot)
		}
	})

	t.Run("resets work directory when package.json changes", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"v1"}`)

		// First prepare
		env, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/reset-key")
		if err != nil {
			t.Fatalf("first prepare failed: %v", err)
		}

		// Write a marker file inside the work directory
		markerPath := filepath.Join(env.WorkDir, "node_modules", "marker.txt")
		mkdirAll(t, filepath.Join(env.WorkDir, "node_modules"))
		writeFile(t, markerPath, "should be removed on reset")

		// Modify package.json (different size triggers reset)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"v2","version":"2.0.0"}`)
		future := time.Now().Add(5 * time.Second)
		if err := os.Chtimes(filepath.Join(moduleRoot, "package.json"), future, future); err != nil {
			t.Fatal(err)
		}

		// Second prepare should trigger reset
		_, err = iso.PrepareIsolatedEnv(moduleRoot, "/out/reset-key")
		if err != nil {
			t.Fatalf("second prepare failed: %v", err)
		}

		// Marker file should be gone (directory was reset)
		if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
			t.Error("expected marker file to be removed after reset, but it still exists")
		}
	})

	t.Run("skips missing optional directories without error", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)
		// No src/, test/, features/, steps/ directories

		_, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/key")
		if err != nil {
			t.Fatalf("expected success with missing optional dirs, got: %v", err)
		}
	})

	t.Run("skips missing optional config files without error", func(t *testing.T) {
		root := t.TempDir()
		iso := NewNpmIsolation(root)

		moduleRoot := filepath.Join(t.TempDir(), "mod")
		mkdirAll(t, moduleRoot)
		writeFile(t, filepath.Join(moduleRoot, "package.json"), `{"name":"test"}`)
		// No tsconfig.json, .mocharc.json, cucumber.js

		env, err := iso.PrepareIsolatedEnv(moduleRoot, "/out/key")
		if err != nil {
			t.Fatalf("expected success with missing optional files, got: %v", err)
		}

		// Optional files should not exist in work dir
		for _, f := range []string{"tsconfig.json", ".mocharc.json", "cucumber.js"} {
			if _, err := os.Stat(filepath.Join(env.WorkDir, f)); err == nil {
				t.Errorf("optional file %s should not exist in work dir", f)
			}
		}
	})
}

// --------------------------------------------------------------------------
// Test helpers
// --------------------------------------------------------------------------

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) failed: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) failed: %v", path, err)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) failed: %v", path, err)
	}
}

func setMtime(t *testing.T, path string, mtime time.Time) {
	t.Helper()
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("Chtimes(%s) failed: %v", path, err)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) failed: %v", path, err)
	}
	if string(data) != want {
		t.Errorf("file %s content = %q, want %q", path, string(data), want)
	}
}
