// Package internal provides shared helpers for godog BDD tests.
package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	contractsreports "github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/git"
)

// TestCache provides cached repository data shared across all test scenarios.
// Uses direct git operations, avoiding FileCache wrapper for simplicity.
//
// This cache is for the ORIGINAL repository root only, not isolated test repos.
// Thread-safe for concurrent access by parallel test packages.
type TestCache struct {
	mu sync.RWMutex

	// repoRoot is the repository root used to populate this cache
	repoRoot string

	// populated indicates if the cache has been loaded
	populated bool

	// trackedFiles is the cached list from git ls-files
	trackedFiles []string

	// moduleReport is the cached module contracts
	moduleReport *contractsreports.ModuleContractReport
}

// NewTestCache creates a new empty test cache.
func NewTestCache() *TestCache {
	return &TestCache{}
}

// EnsurePopulated ensures the cache is populated for the given repo root.
// If already populated for the same root, this is a no-op.
// Thread-safe with double-checked locking pattern.
func (c *TestCache) EnsurePopulated(repoRoot string) error {
	// Fast path: read lock to check if already populated
	c.mu.RLock()
	if c.populated && c.repoRoot == repoRoot {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	// Slow path: write lock to populate
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine may have populated while we waited
	if c.populated && c.repoRoot == repoRoot {
		return nil
	}

	// Reset if repo root changed
	if c.repoRoot != repoRoot {
		c.trackedFiles = nil
		c.moduleReport = nil
		c.populated = false
		c.repoRoot = repoRoot
	}

	// Load git tracked files - this is the ONLY git operation per process!
	// First check for pre-computed file list (CI optimization via GitHub API)
	// In CI: trigger-ci.yaml pre-computes this using GitHub Tree API (~5s vs 84s for git ls-files)
	// Locally: file doesn't exist, falls through to git ls-files (fast on local storage)
	cachedFilePath := filepath.Join(repoRoot, ".git", "cached-files.txt")
	if cachedData, err := os.ReadFile(cachedFilePath); err == nil && len(cachedData) > 0 {
		c.trackedFiles = strings.Split(strings.TrimSpace(string(cachedData)), "\n")
		c.populated = true
		return nil
	}

	// Fall back to git ls-files with timing
	start := time.Now()
	repo, err := git.Open(repoRoot)
	openDuration := time.Since(start)
	if err != nil {
		return err
	}

	start = time.Now()
	files, err := repo.TrackedFiles()
	lsFilesDuration := time.Since(start)
	if err != nil {
		return err
	}

	// Log timing for investigation
	fmt.Printf("⏱️  Cache populate timing: git.Open=%v, TrackedFiles=%v, total=%v, files=%d\n",
		openDuration, lsFilesDuration, openDuration+lsFilesDuration, len(files))

	// Normalize paths to forward slashes for consistency
	c.trackedFiles = make([]string, len(files))
	for i, f := range files {
		c.trackedFiles[i] = strings.ReplaceAll(f, "\\", "/")
	}

	c.populated = true
	return nil
}

// TrackedFiles returns all git-tracked files (normalized paths).
// Must call EnsurePopulated first.
func (c *TestCache) TrackedFiles() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.trackedFiles
}

// FilesByExtension returns files matching the given extension (e.g., ".md").
func (c *TestCache) FilesByExtension(ext string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var filtered []string
	for _, f := range c.trackedFiles {
		if strings.HasSuffix(f, ext) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilesBySuffix returns files matching the given suffix (e.g., "_test.go").
func (c *TestCache) FilesBySuffix(suffix string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var filtered []string
	for _, f := range c.trackedFiles {
		if strings.HasSuffix(f, suffix) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilesInDir returns files under the given directory prefix.
// The dir should use forward slashes (e.g., "go/eac/specs").
func (c *TestCache) FilesInDir(dir string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure dir ends with /
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	var filtered []string
	for _, f := range c.trackedFiles {
		if strings.HasPrefix(f, dir) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilesInDirWithExtension returns files under dir matching extension.
func (c *TestCache) FilesInDirWithExtension(dir, ext string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Ensure dir ends with /
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	var filtered []string
	for _, f := range c.trackedFiles {
		if strings.HasPrefix(f, dir) && strings.HasSuffix(f, ext) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// FilesMatchingAnyExtension returns files matching any of the given extensions.
// Extensions should include the dot (e.g., ".sh", ".ps1").
func (c *TestCache) FilesMatchingAnyExtension(extensions []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var filtered []string
	for _, f := range c.trackedFiles {
		for _, ext := range extensions {
			if strings.HasSuffix(f, ext) {
				filtered = append(filtered, f)
				break
			}
		}
	}
	return filtered
}

// ModuleReport returns the cached module report, loading it if necessary.
func (c *TestCache) ModuleReport() (*contractsreports.ModuleContractReport, error) {
	c.mu.RLock()
	if c.moduleReport != nil {
		report := c.moduleReport
		c.mu.RUnlock()
		return report, nil
	}
	c.mu.RUnlock()

	// Upgrade to write lock to load
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.moduleReport != nil {
		return c.moduleReport, nil
	}

	// Load module contracts with timing
	start := time.Now()
	report, err := contractsreports.GetModuleContracts(c.repoRoot)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}

	fmt.Printf("⏱️  ModuleReport loading: %v, modules=%d\n", duration, len(report.Modules))
	c.moduleReport = report
	return report, nil
}

// AbsolutePath returns the absolute path for a relative tracked file.
func (c *TestCache) AbsolutePath(relPath string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return filepath.Join(c.repoRoot, relPath)
}

// RepoRoot returns the cached repository root.
func (c *TestCache) RepoRoot() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repoRoot
}
