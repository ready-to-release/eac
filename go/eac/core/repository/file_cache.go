// Package repository provides repository-level operations.
package repository

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/github"
)

// FileCache provides cached access to repository files.
// It caches the output of `git ls-files` (or GitHub Trees API in CI) and provides efficient filtered views.
//
// The cache is thread-safe and lazily populated on first access.
// All paths are normalized to forward slashes for cross-platform consistency.
//
// Example:
//
//	cache := repository.NewFileCache(repoRoot, false)  // use git ls-files
//	cache := repository.NewFileCache(repoRoot, true)   // use GitHub API in CI
//	mdFiles := cache.FilesByExtension(".md")
//	goFiles := cache.FilesInDirWithExtension("go/", ".go")
type FileCache struct {
	mu sync.RWMutex

	repoRoot      string
	gitRepo       *git.Repository
	useGitHubInCI bool // Use GitHub Trees API in CI (faster than git ls-files)

	// trackedFiles is the cached output of `git ls-files` or GitHub Trees API
	// All paths are normalized to forward slashes
	trackedFiles []string
	populated    bool

	// Cached filtered views (computed on demand)
	byExtension map[string][]string
	bySuffix    map[string][]string

	// testGitRepo is used for testing to inject a mock git repository
	testGitRepo git.GitRepository
}

// NewFileCache creates a new file cache for the given repository root.
// The cache is lazily populated - no git operations are performed until
// files are requested.
//
// If useGitHubInCI is true and running in CI (detected via environments.IsCI()),
// the cache will use the GitHub Trees API (via GITHUB_SHA env var) instead of
// git ls-files. This is significantly faster in CI environments.
func NewFileCache(repoRoot string, useGitHubInCI bool) *FileCache {
	return &FileCache{
		repoRoot:      repoRoot,
		useGitHubInCI: useGitHubInCI,
		byExtension:   make(map[string][]string),
		bySuffix:      make(map[string][]string),
	}
}

// ensurePopulated ensures the cache is populated with file data.
// In CI with useGitHubInCI enabled, uses GitHub Trees API (fast).
// Otherwise uses git ls-files (local).
// Thread-safe and idempotent.
func (c *FileCache) ensurePopulated() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.populated {
		return nil
	}

	var files []string
	var err error

	// In CI with optimization enabled, try GitHub Trees API first
	// CI detection: check CI or GITHUB_ACTIONS env vars (same as environments.IsCI())
	isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
	if c.useGitHubInCI && isCI {
		files, err = c.populateFromGitHub()
		if err != nil {
			// Fall back to git ls-files on GitHub API failure
			log.Debugf("GitHub Trees API failed, falling back to git ls-files: %v", err)
			files, err = c.populateFromGit()
		}
	} else {
		// DevBox or optimization disabled: use git ls-files
		files, err = c.populateFromGit()
	}

	if err != nil {
		return err
	}

	// Normalize paths to forward slashes
	c.trackedFiles = make([]string, 0, len(files))
	for _, f := range files {
		if f != "" {
			c.trackedFiles = append(c.trackedFiles, strings.ReplaceAll(f, "\\", "/"))
		}
	}

	c.populated = true
	return nil
}

// populateFromGitHub fetches file list using GitHub Trees API.
// Requires GITHUB_SHA env var and github.Global() to be set.
func (c *FileCache) populateFromGitHub() ([]string, error) {
	sha := os.Getenv("GITHUB_SHA")
	if sha == "" {
		return nil, &RepositoryError{Op: "github-trees", Message: "GITHUB_SHA not set"}
	}

	api := github.Global()
	if api == nil {
		return nil, &RepositoryError{Op: "github-trees", Message: "GitHub API not initialized"}
	}

	log.Debugf("Using GitHub Trees API with SHA: %s", sha)
	return api.GetTreeFiles(sha)
}

// populateFromGit fetches file list using git ls-files.
func (c *FileCache) populateFromGit() ([]string, error) {
	// Use injected test repo if available (for unit testing)
	if c.testGitRepo != nil {
		return c.testGitRepo.TrackedFiles()
	}

	repo, err := git.Open(c.repoRoot)
	if err != nil {
		return nil, err
	}
	c.gitRepo = repo

	return repo.TrackedFiles()
}

// TrackedFiles returns all git-tracked files (normalized to forward slashes).
// The result is cached after the first call.
func (c *FileCache) TrackedFiles() ([]string, error) {
	if err := c.ensurePopulated(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.trackedFiles, nil
}

// FilesByExtension returns files matching the given extension (e.g., ".md").
// Results are cached for subsequent calls with the same extension.
func (c *FileCache) FilesByExtension(ext string) ([]string, error) {
	if err := c.ensurePopulated(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	if files, ok := c.byExtension[ext]; ok {
		c.mu.RUnlock()
		return files, nil
	}
	c.mu.RUnlock()

	// Upgrade to write lock to compute
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if files, ok := c.byExtension[ext]; ok {
		return files, nil
	}

	// Filter files by extension
	var filtered []string
	for _, f := range c.trackedFiles {
		if strings.HasSuffix(f, ext) {
			filtered = append(filtered, f)
		}
	}
	c.byExtension[ext] = filtered
	return filtered, nil
}

// FilesBySuffix returns files matching the given suffix (e.g., "_test.go").
// Results are cached for subsequent calls with the same suffix.
func (c *FileCache) FilesBySuffix(suffix string) ([]string, error) {
	if err := c.ensurePopulated(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	if files, ok := c.bySuffix[suffix]; ok {
		c.mu.RUnlock()
		return files, nil
	}
	c.mu.RUnlock()

	// Upgrade to write lock to compute
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if files, ok := c.bySuffix[suffix]; ok {
		return files, nil
	}

	// Filter files by suffix
	var filtered []string
	for _, f := range c.trackedFiles {
		if strings.HasSuffix(f, suffix) {
			filtered = append(filtered, f)
		}
	}
	c.bySuffix[suffix] = filtered
	return filtered, nil
}

// FilesInDir returns files under the given directory prefix.
// The dir should use forward slashes (e.g., "go/eac/specs").
// Results are NOT cached as directory queries vary widely.
func (c *FileCache) FilesInDir(dir string) ([]string, error) {
	if err := c.ensurePopulated(); err != nil {
		return nil, err
	}

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
	return filtered, nil
}

// FilesInDirWithExtension returns files under dir matching extension.
// Results are NOT cached as these queries vary widely.
func (c *FileCache) FilesInDirWithExtension(dir, ext string) ([]string, error) {
	if err := c.ensurePopulated(); err != nil {
		return nil, err
	}

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
	return filtered, nil
}

// FilesMatchingAnyExtension returns files matching any of the given extensions.
// Extensions should include the dot (e.g., ".sh", ".ps1").
// Results are NOT cached as extension combinations vary widely.
func (c *FileCache) FilesMatchingAnyExtension(extensions []string) ([]string, error) {
	if err := c.ensurePopulated(); err != nil {
		return nil, err
	}

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
	return filtered, nil
}

// AbsolutePath returns the absolute path for a relative tracked file.
func (c *FileCache) AbsolutePath(relPath string) string {
	return filepath.Join(c.repoRoot, relPath)
}

// RepoRoot returns the repository root path.
func (c *FileCache) RepoRoot() string {
	return c.repoRoot
}

// Invalidate clears the cache, forcing a refresh on next access.
// Use this if repository files have changed since the cache was populated.
func (c *FileCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.trackedFiles = nil
	c.populated = false
	c.byExtension = make(map[string][]string)
	c.bySuffix = make(map[string][]string)
}
