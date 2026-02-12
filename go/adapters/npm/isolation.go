// Package npm provides isolated npm environments for TypeScript test execution.
//
// This module prevents:
//   - Windows EPERM errors when npm tries to delete locked node_modules files
//   - Interference between parallel test runs
//   - File lock conflicts when switching branches
//
// Solution: Copy source files to .cache/eac/npm/work/{module-moniker}/ and run npm ci there.
package npm

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/core/fileutil"
	"github.com/ready-to-release/eac/go/core/paths"
)

// npmInstallMu serializes all npm ci/install calls to prevent
// concurrent npm cache contention (Windows EBUSY on shared esbuild.exe etc).
// Test execution itself remains fully parallel - only the install step is serialized.
var npmInstallMu sync.Mutex

// WithInstallLock executes fn while holding the npm install mutex.
// This serializes npm ci/install calls to prevent concurrent npm cache contention.
func WithInstallLock(fn func() error) error {
	npmInstallMu.Lock()
	defer npmInstallMu.Unlock()
	return fn()
}

// NpmIsolation manages isolated npm environments for TypeScript test execution.
type NpmIsolation struct {
	workspaceRoot string
	workRoot      string // .cache/eac/npm/work (isolated work directories)
	npmCacheDir   string // .cache/eac/npm/cache (npm download cache)
}

// IsolatedEnv represents a prepared isolated npm environment.
type IsolatedEnv struct {
	WorkDir      string   // Isolated directory for npm ci
	Env          []string // Environment with NPM_CONFIG_CACHE set
	SourceRoot   string   // Original module root (for logging)
	DepsUpToDate bool     // True if npm ci/install can be skipped (lockfile unchanged)
}

// NewNpmIsolation creates a new NpmIsolation instance.
func NewNpmIsolation(workspaceRoot string) *NpmIsolation {
	return &NpmIsolation{
		workspaceRoot: workspaceRoot,
		workRoot:      paths.NpmWorkCachePath(workspaceRoot),
		npmCacheDir:   paths.NpmDownloadCachePath(workspaceRoot),
	}
}

// PrepareIsolatedEnv creates an isolated environment by:
// 1. Creating .cache/npm/work/{key}/ where key is derived from outputDir
// 2. Copying: package.json, package-lock.json, tsconfig*.json, .mocharc.json, cucumber.js
// 3. Syncing directories: src/, test/, features/, steps/
// 4. Setting NPM_CONFIG_CACHE environment variable
//
// outputDir is the unique per-UoW output directory (e.g., out/test/<module>/<dirname>).
// Its basename is used as the isolation key. If empty, falls back to moduleRoot basename.
func (n *NpmIsolation) PrepareIsolatedEnv(moduleRoot, outputDir string) (*IsolatedEnv, error) {
	key := filepath.Base(outputDir)
	if key == "" || key == "." {
		// Derive a stable name from the module root to prevent all modules
		// sharing a single work directory (which corrupts node_modules).
		key = filepath.Base(moduleRoot)
	}
	workDir := filepath.Join(n.workRoot, key)

	// Check if package.json changed - if so, we need a full reset
	// to avoid stale files from previous configurations
	needsReset := packageFilesChanged(moduleRoot, workDir)

	if needsReset {
		// Full reset: remove entire work directory and recreate
		if err := fileutil.RemoveAllWithRetry(workDir); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to reset work directory: %w", err)
		}
	}

	// Create isolated work directory
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Required files to copy
	requiredFiles := []string{"package.json", "package-lock.json"}
	for _, file := range requiredFiles {
		src := filepath.Join(moduleRoot, file)
		dst := filepath.Join(workDir, file)
		if _, err := copyFileIfChanged(src, dst); err != nil {
			if os.IsNotExist(err) && file == "package-lock.json" {
				continue // package-lock.json is optional
			}
			return nil, fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Optional files to copy (if they exist)
	optionalFiles := []string{
		"tsconfig.json",
		"tsconfig.build.json",
		".mocharc.json",
		"cucumber.js",
	}
	for _, file := range optionalFiles {
		src := filepath.Join(moduleRoot, file)
		dst := filepath.Join(workDir, file)
		_, _ = copyFileIfChanged(src, dst) // Ignore errors for optional files
	}

	// Directories to sync (with incremental updates)
	syncDirs := []string{"src", "test", "features", "steps"}
	for _, dir := range syncDirs {
		srcDir := filepath.Join(moduleRoot, dir)
		dstDir := filepath.Join(workDir, dir)
		if err := syncDirectory(srcDir, dstDir); err != nil {
			// Non-fatal: directory may not exist
			continue
		}
	}

	// Build environment with NPM_CONFIG_CACHE pointing to .cache/npm
	// This shares the npm download cache across all modules
	env := os.Environ()
	npmCacheSet := false
	for i, e := range env {
		if strings.HasPrefix(e, "NPM_CONFIG_CACHE=") {
			env[i] = "NPM_CONFIG_CACHE=" + n.npmCacheDir
			npmCacheSet = true
			break
		}
	}
	if !npmCacheSet {
		env = append(env, "NPM_CONFIG_CACHE="+n.npmCacheDir)
	}

	return &IsolatedEnv{
		WorkDir:      workDir,
		Env:          env,
		SourceRoot:   moduleRoot,
		DepsUpToDate: !needsReset && npmDepsMarkerValid(workDir, moduleRoot),
	}, nil
}

// MarkNpmDepsInstalled writes a marker file after a successful npm ci/install.
// The marker records the mtime of package-lock.json (or package.json as fallback)
// so subsequent runs can skip reinstallation.
func MarkNpmDepsInstalled(workDir, moduleRoot string) {
	markerPath := filepath.Join(workDir, ".eac-npm-deps-installed")
	lockFile := filepath.Join(moduleRoot, "package-lock.json")
	info, err := os.Stat(lockFile)
	if err != nil {
		// Fallback to package.json
		lockFile = filepath.Join(moduleRoot, "package.json")
		info, err = os.Stat(lockFile)
		if err != nil {
			return
		}
	}
	_ = os.WriteFile(markerPath, []byte(info.ModTime().Format(time.RFC3339Nano)), 0o644)
}

// npmDepsMarkerValid returns true if the npm deps marker matches the current lockfile mtime.
func npmDepsMarkerValid(workDir, moduleRoot string) bool {
	markerPath := filepath.Join(workDir, ".eac-npm-deps-installed")
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return false
	}
	// Also check that node_modules exists
	if _, err := os.Stat(filepath.Join(workDir, "node_modules")); os.IsNotExist(err) {
		return false
	}
	lockFile := filepath.Join(moduleRoot, "package-lock.json")
	info, err := os.Stat(lockFile)
	if err != nil {
		lockFile = filepath.Join(moduleRoot, "package.json")
		info, err = os.Stat(lockFile)
		if err != nil {
			return false
		}
	}
	recorded, err := time.Parse(time.RFC3339Nano, string(data))
	if err != nil {
		return false
	}
	return info.ModTime().Equal(recorded)
}

// packageFilesChanged checks if package.json or package-lock.json have changed.
// Note: mtime+size comparison can miss content changes if both are identical.
// This is an acceptable trade-off for sync performance; content hashing would be slower.
func packageFilesChanged(moduleRoot, workDir string) bool {
	files := []string{"package.json", "package-lock.json"}
	for _, file := range files {
		src := filepath.Join(moduleRoot, file)
		dst := filepath.Join(workDir, file)

		srcInfo, err := os.Stat(src)
		if err != nil {
			continue // Source doesn't exist
		}

		dstInfo, err := os.Stat(dst)
		if err != nil {
			return true // Destination doesn't exist, needs sync
		}

		// Check if changed
		if srcInfo.Size() != dstInfo.Size() || !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
			return true
		}
	}
	return false
}

// copyFileIfChanged copies src to dst only if src has changed (based on mtime and size).
// Returns (true, nil) if file was copied, (false, nil) if no update needed.
// Note: mtime+size comparison can miss content changes if both are identical.
// This is an acceptable trade-off for sync performance; content hashing would be slower.
func copyFileIfChanged(src, dst string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	// Check if destination exists and is up to date
	dstInfo, err := os.Stat(dst)
	if err == nil {
		// File exists - check if it needs updating
		if srcInfo.Size() == dstInfo.Size() && srcInfo.ModTime().Equal(dstInfo.ModTime()) {
			return false, nil // No update needed
		}
	}

	// Copy the file
	srcFile, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return false, err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return false, err
	}

	// Preserve modification time for incremental detection
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return true, err
	}
	return true, nil
}

// syncDirectory synchronizes src directory to dst with incremental updates.
// Only copies files that have changed based on mtime and size.
// Uses a single pass: walks the source to copy changed files while building
// a set of source-relative paths, then removes destination files not in that set.
func syncDirectory(srcDir, dstDir string) error {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err // Source doesn't exist
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", srcDir)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	// Single pass: walk source, copy changed files, track source-relative paths
	srcFiles := make(map[string]bool)
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		// Track this source file
		srcFiles[relPath] = true

		// Copy if changed
		_, copyErr := copyFileIfChanged(path, dstPath)
		return copyErr
	})
	if err != nil {
		return err
	}

	// Walk destination to remove files that no longer exist in source
	_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(dstDir, path)
		if !srcFiles[relPath] {
			_ = os.Remove(path)
		}
		return nil
	})

	return nil
}
