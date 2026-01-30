package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/paths"
)

// Environment variable names (duplicated to avoid import cycle with environments package).
const (
	envR2RRepoRoot      = "R2R_REPO_ROOT"
	envR2RContainerRoot = "R2R_CONTAINER_ROOT"
	envR2RPWD           = "R2R_PWD"
	envR2RDockerMode    = "R2R_DOCKER_MODE"
)

// Cache for workspace detection results.
// The cache is bypassed when R2R_REPO_ROOT is set (test isolation).
var (
	cachedWorkspace *Workspace
	cacheMu         sync.RWMutex
)

// Workspace holds resolved workspace information.
type Workspace struct {
	// Root is the absolute path to the workspace root.
	Root string

	// Source indicates how the workspace was detected.
	Source string // "env:R2R_REPO_ROOT", "env:R2R_DOCKER_MODE", "git"

	// IsContainer is true when running inside a container.
	IsContainer bool

	// DistRoot is the distribution root (container root or repo root).
	// Use for loading tool assets (templates, schemas).
	DistRoot string
}

// Detect finds the workspace root using the full detection chain.
// This is the primary entry point for most use cases.
//
// Results are cached for performance. The cache is bypassed when
// R2R_REPO_ROOT is set to ensure test isolation works correctly.
func Detect() (*Workspace, error) {
	// IMPORTANT: Check for test isolation override FIRST, before using cached value.
	// This ensures isolated tests use their temporary directory instead of the cached real repo root.
	if envRoot := os.Getenv(envR2RRepoRoot); envRoot != "" {
		// Don't use cache when env override is set
		return DetectWithOptions(DefaultOptions())
	}

	// Check cache (read lock)
	cacheMu.RLock()
	if cachedWorkspace != nil {
		ws := cachedWorkspace
		cacheMu.RUnlock()
		return ws, nil
	}
	cacheMu.RUnlock()

	// Cache miss - detect and cache (write lock)
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-check after acquiring write lock
	if cachedWorkspace != nil {
		return cachedWorkspace, nil
	}

	ws, err := DetectWithOptions(DefaultOptions())
	if err != nil {
		return nil, err
	}

	// Only cache git-detected workspaces (not env overrides or docker mode)
	// because those are stable for the process lifetime
	if ws.Source == "git" {
		cachedWorkspace = ws
	}

	return ws, nil
}

// DetectWithOptions finds the workspace root with custom options.
func DetectWithOptions(opts Options) (*Workspace, error) {
	w := &Workspace{}

	// Set container flag early (used throughout)
	w.IsContainer = os.Getenv("R2R_DOCKER_MODE") == "true"

	// Step 1: Check explicit override (unless ModeGitOnly)
	if opts.Mode != ModeGitOnly {
		if root := os.Getenv(envR2RRepoRoot); root != "" {
			w.Root = filepath.Clean(root)
			w.Source = "env:R2R_REPO_ROOT"
			if err := validateIfRequired(w.Root, opts); err != nil {
				return nil, err
			}
			w.DistRoot = resolveDistRoot(w.Root, w.IsContainer)
			return w, nil
		}
	}

	// Step 2: ModeExplicit fails if no env var set
	if opts.Mode == ModeExplicit {
		return nil, &DetectionError{
			Op:      "detect",
			Source:  "env:R2R_REPO_ROOT",
			Message: "R2R_REPO_ROOT not set but ModeExplicit requested",
			Err:     ErrNotFound,
		}
	}

	// Step 3: Docker mode fallback (unless ModeGitOnly)
	if opts.Mode != ModeGitOnly && w.IsContainer {
		w.Root = paths.ContainerRepoRoot
		w.Source = "env:R2R_DOCKER_MODE"
		if err := validateIfRequired(w.Root, opts); err != nil {
			return nil, err
		}
		w.DistRoot = resolveDistRoot(w.Root, w.IsContainer)
		return w, nil
	}

	// Step 4: Git detection
	startPath := opts.StartPath
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return nil, &DetectionError{
				Op:      "detect",
				Source:  "cwd",
				Message: "failed to get current directory",
				Err:     err,
			}
		}
	}

	gitRoot, err := findGitRoot(startPath)
	if err != nil {
		return nil, &DetectionError{
			Op:      "detect",
			Path:    startPath,
			Source:  "git",
			Message: "not a git repository (or any parent up to filesystem root)",
			Err:     ErrNotFound,
		}
	}

	w.Root = gitRoot
	w.Source = "git"
	if err := validateIfRequired(w.Root, opts); err != nil {
		return nil, err
	}
	w.DistRoot = resolveDistRoot(w.Root, w.IsContainer)
	return w, nil
}

// MustDetect is like Detect but panics on error.
// Use only in init() or where workspace is absolutely required.
func MustDetect() *Workspace {
	w, err := Detect()
	if err != nil {
		panic(fmt.Sprintf("workspace detection failed: %v", err))
	}
	return w
}

// Root returns just the workspace root path.
// Shorthand for Detect().Root.
func Root() (string, error) {
	w, err := Detect()
	if err != nil {
		return "", err
	}
	return w.Root, nil
}

// RootOrPanic returns the workspace root or panics.
// Use only where workspace is absolutely required.
func RootOrPanic() string {
	return MustDetect().Root
}

// IsInContainer returns true if running inside a container.
func IsInContainer() bool {
	return os.Getenv("R2R_DOCKER_MODE") == "true"
}

// DistRoot returns the distribution root for loading tool assets.
// In containers, this is R2R_CONTAINER_ROOT; otherwise, the workspace root.
func DistRoot() string {
	if root := os.Getenv(envR2RContainerRoot); root != "" {
		return root
	}
	r, err := Root()
	if err != nil {
		return ""
	}
	return r
}

// WorkingDir returns the effective working directory.
// Checks R2R_PWD first, then falls back to os.Getwd().
func WorkingDir() (string, error) {
	if pwd := os.Getenv(envR2RPWD); pwd != "" {
		return filepath.Clean(pwd), nil
	}
	return os.Getwd()
}

// ForTesting creates a workspace configuration for test environments.
// Sets R2R_REPO_ROOT to the given path and returns a cleanup function.
//
// Usage:
//
//	cleanup := workspace.ForTesting(t, tempDir)
//	defer cleanup()
func ForTesting(t interface{ Helper(); Cleanup(func()) }, root string) func() {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}

	old := os.Getenv(envR2RRepoRoot)
	os.Setenv(envR2RRepoRoot, root)

	cleanup := func() {
		if old == "" {
			os.Unsetenv(envR2RRepoRoot)
		} else {
			os.Setenv(envR2RRepoRoot, old)
		}
	}

	if c, ok := t.(interface{ Cleanup(func()) }); ok {
		c.Cleanup(cleanup)
	}

	return cleanup
}

// RequireIsolation fails the test immediately if not running in isolation.
// Use at the start of tests that modify workspace state.
func RequireIsolation(t interface{ Helper(); Fatalf(string, ...any) }) {
	if h, ok := t.(interface{ Helper() }); ok {
		h.Helper()
	}
	if os.Getenv(envR2RRepoRoot) == "" {
		t.Fatalf("test requires isolation: R2R_REPO_ROOT not set")
	}
}

// ClearCache clears the cached workspace detection result.
// This is primarily for testing to ensure clean state between tests.
func ClearCache() {
	cacheMu.Lock()
	cachedWorkspace = nil
	cacheMu.Unlock()
}

// findGitRoot walks up from startPath looking for a .git directory.
func findGitRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	current := absPath
	for {
		gitPath := filepath.Join(current, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			// Found .git - check if it's a directory or file (submodule/worktree)
			if info.IsDir() || info.Mode().IsRegular() {
				return current, nil
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			return "", ErrNotFound
		}
		current = parent
	}
}

// validateIfRequired validates the workspace path if validation is enabled.
func validateIfRequired(root string, opts Options) error {
	if !opts.Validate {
		return nil
	}

	// Check directory exists
	info, err := os.Stat(root)
	if os.IsNotExist(err) {
		return &DetectionError{
			Op:      "validate",
			Path:    root,
			Source:  "filesystem",
			Message: "path does not exist",
			Err:     ErrInvalidPath,
		}
	}
	if err != nil {
		return &DetectionError{
			Op:      "validate",
			Path:    root,
			Source:  "filesystem",
			Message: "cannot stat path",
			Err:     err,
		}
	}
	if !info.IsDir() {
		return &DetectionError{
			Op:      "validate",
			Path:    root,
			Source:  "filesystem",
			Message: "path is not a directory",
			Err:     ErrInvalidPath,
		}
	}

	// Check for .git or .r2r/eac/repository.yml
	gitPath := filepath.Join(root, ".git")
	eacPath := filepath.Join(root, paths.EACConfigRelPath, "repository.yml")

	hasGit := false
	if _, err := os.Stat(gitPath); err == nil {
		hasGit = true
	}

	hasEac := false
	if _, err := os.Stat(eacPath); err == nil {
		hasEac = true
	}

	if opts.RequireGit && !hasGit {
		return &DetectionError{
			Op:      "validate",
			Path:    root,
			Source:  ".git",
			Message: ".git directory not found",
			Err:     ErrInvalidPath,
		}
	}

	if !hasGit && !hasEac {
		return &DetectionError{
			Op:      "validate",
			Path:    root,
			Source:  "markers",
			Message: "neither .git nor .r2r/eac/repository.yml found",
			Err:     ErrInvalidPath,
		}
	}

	return nil
}

// resolveDistRoot determines the distribution root for tool assets.
func resolveDistRoot(workspaceRoot string, _ bool) string {
	if containerRoot := os.Getenv(envR2RContainerRoot); containerRoot != "" {
		return containerRoot
	}
	return workspaceRoot
}
