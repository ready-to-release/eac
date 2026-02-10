// Package nuget provides isolated NuGet environments for .NET test and build execution.
//
// This module prevents:
//   - NuGet cache corruption from parallel dotnet restore operations
//   - Windows EPERM errors when dotnet tries to delete locked bin/obj files
//   - Interference between parallel test runs sharing the same bin/obj directories
//
// Solution: Copy project files to .cache/eac/nuget/work/{key}/ and run dotnet restore there.
package nuget

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/core/fileutil"
	"github.com/ready-to-release/eac/go/core/paths"
)

// NuGetRestoreMu serializes all dotnet restore calls to prevent
// concurrent NuGet cache contention (Windows file locking on shared packages).
// Test execution itself remains fully parallel - only the restore step is serialized.
var NuGetRestoreMu sync.Mutex

// NuGetIsolation manages isolated .NET environments for test and build execution.
type NuGetIsolation struct {
	workspaceRoot string
	workRoot      string // .cache/eac/nuget/work (isolated work directories)
	nugetCache    string // .cache/eac/nuget/packages (NuGet package cache)
}

// IsolatedEnv represents a prepared isolated .NET environment.
type IsolatedEnv struct {
	WorkDir    string   // Isolated directory for dotnet operations
	Env        []string // Environment with NUGET_PACKAGES set
	SourceRoot string   // Original module root (for logging)
	SlnFile    string   // Solution file path (if discovered), relative to WorkDir
}

// NewNuGetIsolation creates a new NuGetIsolation instance.
func NewNuGetIsolation(workspaceRoot string) *NuGetIsolation {
	return &NuGetIsolation{
		workspaceRoot: workspaceRoot,
		workRoot:      paths.NuGetWorkCachePath(workspaceRoot),
		nugetCache:    paths.NuGetPackageCachePath(workspaceRoot),
	}
}

// PrepareIsolatedEnv creates an isolated environment by:
// 1. Creating .cache/nuget/work/{key}/ where key is derived from outputDir
// 2. Copying: *.csproj, *.sln, *.props, *.targets, nuget.config, Directory.Build.props
// 3. Syncing directories: src/, test/, tests/
// 4. Setting NUGET_PACKAGES environment variable for shared package cache
//
// outputDir is the unique per-UoW output directory. Its basename is used as the isolation key.
func (n *NuGetIsolation) PrepareIsolatedEnv(moduleRoot, outputDir string) (*IsolatedEnv, error) {
	key := filepath.Base(outputDir)
	if key == "" || key == "." {
		key = filepath.Base(moduleRoot)
	}
	workDir := filepath.Join(n.workRoot, key)

	// Check if project files changed - if so, full reset needed
	needsReset := projectFilesChanged(moduleRoot, workDir)
	if needsReset {
		if err := fileutil.RemoveAllWithRetry(workDir); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to reset work directory: %w", err)
		}
	}

	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}

	// Copy project files (*.csproj, *.sln, Directory.Build.props, etc.)
	if err := syncDotnetProjectFiles(moduleRoot, workDir); err != nil {
		return nil, fmt.Errorf("failed to sync project files: %w", err)
	}

	// Sync source directories
	for _, dir := range []string{"src", "test", "tests"} {
		srcDir := filepath.Join(moduleRoot, dir)
		dstDir := filepath.Join(workDir, dir)
		_ = syncDirectory(srcDir, dstDir) // Non-fatal: directory may not exist
	}

	// Discover solution file
	slnFile := discoverSlnFile(workDir)

	// Build environment with shared NuGet package cache
	env := os.Environ()
	nugetSet := false
	for i, e := range env {
		if strings.HasPrefix(e, "NUGET_PACKAGES=") {
			env[i] = "NUGET_PACKAGES=" + n.nugetCache
			nugetSet = true
			break
		}
	}
	if !nugetSet {
		env = append(env, "NUGET_PACKAGES="+n.nugetCache)
	}

	return &IsolatedEnv{
		WorkDir:    workDir,
		Env:        env,
		SourceRoot: moduleRoot,
		SlnFile:    slnFile,
	}, nil
}

// syncDotnetProjectFiles copies all .NET project configuration files.
func syncDotnetProjectFiles(srcRoot, dstRoot string) error {
	patterns := []string{
		"*.csproj", "*.fsproj", "*.vbproj", "*.sln",
		"Directory.Build.props", "Directory.Build.targets",
		"Directory.Packages.props", "nuget.config", "NuGet.Config",
		"global.json",
	}
	return filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		for _, pattern := range patterns {
			if matched, _ := filepath.Match(pattern, info.Name()); matched {
				relPath, _ := filepath.Rel(srcRoot, path)
				dst := filepath.Join(dstRoot, relPath)
				if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
					return err
				}
				_, copyErr := copyFileIfChanged(path, dst)
				return copyErr
			}
		}
		return nil
	})
}

// discoverSlnFile finds the first .sln file in the work directory.
func discoverSlnFile(workDir string) string {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sln") {
			return e.Name()
		}
	}
	return ""
}

// projectFilesChanged checks if any .csproj or .sln files have changed.
func projectFilesChanged(moduleRoot, workDir string) bool {
	patterns := []string{"*.csproj", "*.sln", "Directory.Build.props"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(moduleRoot, pattern))
		for _, src := range matches {
			dst := filepath.Join(workDir, filepath.Base(src))
			srcInfo, err := os.Stat(src)
			if err != nil {
				continue
			}
			dstInfo, err := os.Stat(dst)
			if err != nil {
				return true
			}
			if srcInfo.Size() != dstInfo.Size() || !srcInfo.ModTime().Equal(dstInfo.ModTime()) {
				return true
			}
		}
	}
	return false
}

// copyFileIfChanged copies src to dst only if src has changed (based on mtime and size).
// Returns (true, nil) if file was copied, (false, nil) if no update needed.
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

	dstFiles := make(map[string]bool)
	_ = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(dstDir, path)
		dstFiles[relPath] = true
		return nil
	})

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

	for relPath := range dstFiles {
		_ = os.Remove(filepath.Join(dstDir, relPath))
	}

	return err
}
