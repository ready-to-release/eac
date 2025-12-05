// Package internal provides shared helpers for godog BDD tests.
package internal

import (
	"sync"

	contractsreports "github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

// TestCache provides cached repository data shared across all test scenarios.
// It delegates file operations to repository.FileCache (single source of truth)
// and adds module contract caching for test performance.
//
// This cache is for the ORIGINAL repository root only, not isolated test repos.
type TestCache struct {
	mu sync.RWMutex

	// repoRoot is the repository root used to populate this cache
	repoRoot string

	// fileCache delegates to core repository.FileCache for git operations
	fileCache *repository.FileCache

	// moduleReport is the cached module contracts
	moduleReport *contractsreports.ModuleContractReport
}

// NewTestCache creates a new empty test cache.
func NewTestCache() *TestCache {
	return &TestCache{}
}

// EnsurePopulated ensures the cache is populated for the given repo root.
// If already populated for the same root, this is a no-op.
// Thread-safe.
func (c *TestCache) EnsurePopulated(repoRoot string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Already populated for this repo
	if c.repoRoot == repoRoot && c.fileCache != nil {
		return nil
	}

	// Reset if repo root changed
	if c.repoRoot != repoRoot {
		c.fileCache = nil
		c.moduleReport = nil
		c.repoRoot = repoRoot
	}

	// Create file cache (delegates to core repository package)
	c.fileCache = repository.NewFileCache(repoRoot)

	// Pre-populate the cache (optional but improves first-access time)
	_, err := c.fileCache.TrackedFiles()
	return err
}

// TrackedFiles returns all git-tracked files (normalized paths).
// Must call EnsurePopulated first.
func (c *TestCache) TrackedFiles() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.fileCache == nil {
		return nil
	}

	files, _ := c.fileCache.TrackedFiles()
	return files
}

// FilesByExtension returns files matching the given extension (e.g., ".md").
// Results are cached for subsequent calls.
func (c *TestCache) FilesByExtension(ext string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.fileCache == nil {
		return nil
	}

	files, _ := c.fileCache.FilesByExtension(ext)
	return files
}

// FilesBySuffix returns files matching the given suffix (e.g., "_test.go").
// Results are cached for subsequent calls.
func (c *TestCache) FilesBySuffix(suffix string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.fileCache == nil {
		return nil
	}

	files, _ := c.fileCache.FilesBySuffix(suffix)
	return files
}

// FilesInDir returns files under the given directory prefix.
// The dir should use forward slashes (e.g., "go/eac/specs").
func (c *TestCache) FilesInDir(dir string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.fileCache == nil {
		return nil
	}

	files, _ := c.fileCache.FilesInDir(dir)
	return files
}

// FilesInDirWithExtension returns files under dir matching extension.
func (c *TestCache) FilesInDirWithExtension(dir, ext string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.fileCache == nil {
		return nil
	}

	files, _ := c.fileCache.FilesInDirWithExtension(dir, ext)
	return files
}

// FilesMatchingAnyExtension returns files matching any of the given extensions.
// Extensions should include the dot (e.g., ".sh", ".ps1").
func (c *TestCache) FilesMatchingAnyExtension(extensions []string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.fileCache == nil {
		return nil
	}

	files, _ := c.fileCache.FilesMatchingAnyExtension(extensions)
	return files
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

	// Load module contracts
	report, err := contractsreports.GetModuleContracts(c.repoRoot)
	if err != nil {
		return nil, err
	}
	c.moduleReport = report
	return report, nil
}

// AbsolutePath returns the absolute path for a relative tracked file.
func (c *TestCache) AbsolutePath(relPath string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.fileCache == nil {
		return relPath
	}

	return c.fileCache.AbsolutePath(relPath)
}

// RepoRoot returns the cached repository root.
func (c *TestCache) RepoRoot() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.repoRoot
}
