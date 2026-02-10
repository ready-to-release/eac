// Package pip provides isolated Python virtual environments for test execution.
//
// This module prevents:
//   - Interference between parallel test runs sharing the same venv
//   - File lock conflicts on Windows when multiple tests install packages
//   - Stale venvs when requirements change
//
// Solution: Create venvs in .cache/eac/python/venv/{moniker}/ and copy source
// files to .cache/eac/python/work/{moniker}/ for isolated execution.
package pip

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/core/fileutil"
	"github.com/ready-to-release/eac/go/core/paths"
)

// PipInstallMu serializes all pip install calls to prevent
// concurrent pip cache contention. Test execution itself remains
// fully parallel - only the install step is serialized.
var PipInstallMu sync.Mutex

// PipIsolation manages isolated Python environments for test execution.
type PipIsolation struct {
	workspaceRoot string
	workRoot      string // .cache/eac/python/work (isolated work directories)
	venvRoot      string // .cache/eac/python/venv (virtual environments)
	pipCacheDir   string // .cache/eac/python/pip-cache (pip download cache)
}

// IsolatedEnv represents a prepared isolated Python environment.
type IsolatedEnv struct {
	WorkDir    string   // Isolated directory with copied source files
	VenvDir    string   // Path to the virtual environment
	PythonBin  string   // Path to python binary in venv
	Env        []string // Environment with VIRTUAL_ENV, PATH, PIP_CACHE_DIR
	SourceRoot string   // Original module root (for logging)
}

// NewPipIsolation creates a new PipIsolation instance.
func NewPipIsolation(workspaceRoot string) *PipIsolation {
	return &PipIsolation{
		workspaceRoot: workspaceRoot,
		workRoot:      paths.PipWorkCachePath(workspaceRoot),
		venvRoot:      paths.PipVenvCachePath(workspaceRoot),
		pipCacheDir:   paths.PipDownloadCachePath(workspaceRoot),
	}
}

// PrepareIsolatedEnv creates an isolated environment by:
// 1. Creating .cache/python/work/{key}/ where key is derived from outputDir
// 2. Copying: pyproject.toml, setup.py, setup.cfg, requirements.txt, requirements-dev.txt
// 3. Syncing directories: src/, tests/, features/, steps/
// 4. Creating/reusing a venv at .cache/python/venv/{key}/
// 5. Setting VIRTUAL_ENV, PATH, PIP_CACHE_DIR environment variables
//
// outputDir is the unique per-UoW output directory (e.g., out/test/<module>/<dirname>).
// Its basename is used as the isolation key. If empty, falls back to moduleRoot basename.
func (p *PipIsolation) PrepareIsolatedEnv(moduleRoot, outputDir string) (*IsolatedEnv, error) {
	key := filepath.Base(outputDir)
	if key == "" || key == "." {
		key = filepath.Base(moduleRoot)
	}
	workDir := filepath.Join(p.workRoot, key)
	venvDir := filepath.Join(p.venvRoot, key)

	// Check if requirements files changed - if so, full reset
	needsReset := requirementsChanged(moduleRoot, workDir)

	if needsReset {
		if err := fileutil.RemoveAllWithRetry(workDir); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to reset work directory: %w", err)
		}
	}

	// Create isolated work directory
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Required files to copy
	requiredFiles := []string{"pyproject.toml"}
	for _, file := range requiredFiles {
		src := filepath.Join(moduleRoot, file)
		dst := filepath.Join(workDir, file)
		if _, err := copyFileIfChanged(src, dst); err != nil {
			if os.IsNotExist(err) {
				continue // pyproject.toml may not exist for simple setups
			}
			return nil, fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	// Optional files to copy (if they exist)
	optionalFiles := []string{
		"setup.py",
		"setup.cfg",
		"requirements.txt",
		"requirements-dev.txt",
		"pytest.ini",
		"conftest.py",
		"behave.ini",
	}
	for _, file := range optionalFiles {
		src := filepath.Join(moduleRoot, file)
		dst := filepath.Join(workDir, file)
		_, _ = copyFileIfChanged(src, dst) // Ignore errors for optional files
	}

	// Directories to sync (with incremental updates)
	syncDirs := []string{"src", "tests", "features", "steps"}
	for _, dir := range syncDirs {
		srcDir := filepath.Join(moduleRoot, dir)
		dstDir := filepath.Join(workDir, dir)
		if err := syncDirectory(srcDir, dstDir); err != nil {
			continue // Non-fatal: directory may not exist
		}
	}

	// Ensure venv exists
	pythonBin := venvPythonPath(venvDir)

	// Build environment
	env := buildPythonEnv(venvDir, p.pipCacheDir)

	return &IsolatedEnv{
		WorkDir:    workDir,
		VenvDir:    venvDir,
		PythonBin:  pythonBin,
		Env:        env,
		SourceRoot: moduleRoot,
	}, nil
}

// NeedsVenvCreate checks whether the venv needs to be created.
func (p *PipIsolation) NeedsVenvCreate(key string) bool {
	venvDir := filepath.Join(p.venvRoot, key)
	pythonBin := venvPythonPath(venvDir)
	_, err := os.Stat(pythonBin)
	return os.IsNotExist(err)
}

// venvPythonPath returns the platform-specific path to the Python binary in a venv.
func venvPythonPath(venvDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvDir, "Scripts", "python.exe")
	}
	return filepath.Join(venvDir, "bin", "python")
}

// VenvPipPath returns the platform-specific path to the pip binary in a venv.
func VenvPipPath(venvDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvDir, "Scripts", "pip.exe")
	}
	return filepath.Join(venvDir, "bin", "pip")
}

// buildPythonEnv builds an environment slice with VIRTUAL_ENV, updated PATH, and PIP_CACHE_DIR.
func buildPythonEnv(venvDir, pipCacheDir string) []string {
	env := os.Environ()

	var binDir string
	if runtime.GOOS == "windows" {
		binDir = filepath.Join(venvDir, "Scripts")
	} else {
		binDir = filepath.Join(venvDir, "bin")
	}

	result := make([]string, 0, len(env)+3)
	virtualEnvSet := false
	pipCacheSet := false

	for _, e := range env {
		if strings.HasPrefix(e, "VIRTUAL_ENV=") {
			result = append(result, "VIRTUAL_ENV="+venvDir)
			virtualEnvSet = true
		} else if strings.HasPrefix(e, "PIP_CACHE_DIR=") {
			result = append(result, "PIP_CACHE_DIR="+pipCacheDir)
			pipCacheSet = true
		} else if strings.HasPrefix(e, "PATH=") {
			// Prepend venv bin to PATH
			currentPath := strings.TrimPrefix(e, "PATH=")
			result = append(result, "PATH="+binDir+string(os.PathListSeparator)+currentPath)
		} else {
			result = append(result, e)
		}
	}

	if !virtualEnvSet {
		result = append(result, "VIRTUAL_ENV="+venvDir)
	}
	if !pipCacheSet {
		result = append(result, "PIP_CACHE_DIR="+pipCacheDir)
	}

	return result
}

// requirementsChanged checks if pyproject.toml or requirements*.txt have changed.
func requirementsChanged(moduleRoot, workDir string) bool {
	files := []string{"pyproject.toml", "requirements.txt", "requirements-dev.txt", "setup.py"}
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

		if srcInfo.Size() != dstInfo.Size() || !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
			return true
		}
	}
	return false
}

// copyFileIfChanged copies src to dst only if src has changed (based on mtime and size).
func copyFileIfChanged(src, dst string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	dstInfo, err := os.Stat(dst)
	if err == nil {
		if srcInfo.Size() == dstInfo.Size() && srcInfo.ModTime().Equal(dstInfo.ModTime()) {
			return false, nil
		}
	}

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

	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return true, err
	}
	return true, nil
}

// syncDirectory synchronizes src directory to dst with incremental updates.
func syncDirectory(srcDir, dstDir string) error {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", srcDir)
	}

	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}

	// Track files in destination for cleanup
	dstFiles := make(map[string]bool)
	_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(dstDir, path)
		dstFiles[relPath] = true
		return nil
	})

	// Walk source and copy changed files
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return nil
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		delete(dstFiles, relPath)
		_, copyErr := copyFileIfChanged(path, dstPath)
		return copyErr
	})

	// Remove files that no longer exist in source
	for relPath := range dstFiles {
		_ = os.Remove(filepath.Join(dstDir, relPath))
	}

	return err
}
