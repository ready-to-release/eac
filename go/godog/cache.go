package godog

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/core/domain/modules"
	contractsreports "github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
)

var log = logging.C()

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
		c.populated = false
		c.repoRoot = repoRoot
	}

	// Load git tracked files - this is the ONLY git operation per process!
	// First check for pre-computed file list (CI optimization via GitHub API)
	// In CI: change-trigger.yaml pre-computes this using GitHub Tree API (~5s vs 84s for git ls-files)
	// Locally: file doesn't exist, falls through to git ls-files (fast on local storage)
	cachedFilePath := filepath.Join(repoRoot, ".git", "cached-files.txt")
	if cachedData, err := os.ReadFile(cachedFilePath); err == nil && len(cachedData) > 0 {
		c.trackedFiles = strings.Split(strings.TrimSpace(string(cachedData)), "\n")
		c.populated = true
		log.Debugf("Cache populated from pre-computed file (files=%d, source=.git/cached-files.txt)", len(c.trackedFiles))
		return nil
	}

	// Fall back to git ls-files with timing
	log.Debug("Pre-computed file list not found, using git ls-files")
	start := time.Now()
	gitMgr := git.NewManager(logging.C().Zap())
	repo, err := gitMgr.Open(repoRoot)
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
	totalDuration := openDuration + lsFilesDuration
	log.Debugf("Cache populate timing: git.Open=%v, TrackedFiles=%v, total=%v, files=%d",
		openDuration, lsFilesDuration, totalDuration, len(files))

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

// ModuleReport returns the module contract report as a port interface.
// Caching is handled internally by GetModuleContracts - consumers don't need to know.
func (c *TestCache) ModuleReport() (interfaces.ModuleReportPort, error) {
	c.mu.RLock()
	repoRoot := c.repoRoot
	c.mu.RUnlock()

	// GetModuleContracts has internal global caching - just call it
	report, err := contractsreports.GetModuleContracts(repoRoot)
	if err != nil {
		return nil, err
	}
	return newModuleReportAdapter(report), nil
}

// ModuleReportConcrete returns the concrete module contract report.
// Use this when you need access to the concrete type (e.g., during migration).
func (c *TestCache) ModuleReportConcrete() (*contractsreports.ModuleContractReport, error) {
	c.mu.RLock()
	repoRoot := c.repoRoot
	c.mu.RUnlock()

	return contractsreports.GetModuleContracts(repoRoot)
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

// Interface compliance check
var _ interfaces.TestCachePort = (*TestCache)(nil)

// ============================================================================
// Adapters for concrete types to port interfaces
// ============================================================================

// moduleReportAdapter wraps *contractsreports.ModuleContractReport to implement interfaces.ModuleReportPort.
type moduleReportAdapter struct {
	report *contractsreports.ModuleContractReport
}

func newModuleReportAdapter(report *contractsreports.ModuleContractReport) *moduleReportAdapter {
	return &moduleReportAdapter{report: report}
}

// Registry implements interfaces.ModuleReportPort.
func (a *moduleReportAdapter) Registry() interfaces.ModuleRegistryPort {
	if a.report.Registry == nil {
		return nil
	}
	return newModuleRegistryAdapter(a.report.Registry)
}

// Errors implements interfaces.ModuleReportPort.
func (a *moduleReportAdapter) Errors() []error {
	return nil // ModuleContractReport doesn't track errors
}

var _ interfaces.ModuleReportPort = (*moduleReportAdapter)(nil)

// moduleRegistryAdapter wraps *modules.Registry to implement interfaces.ModuleRegistryPort.
type moduleRegistryAdapter struct {
	registry *modules.Registry
}

func newModuleRegistryAdapter(registry *modules.Registry) *moduleRegistryAdapter {
	return &moduleRegistryAdapter{registry: registry}
}

// Get implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) Get(moniker string) (interfaces.ModuleContractPort, bool) {
	mod, ok := a.registry.Get(moniker)
	if !ok || mod == nil {
		return nil, false
	}
	return newModuleContractAdapter(mod), true
}

// Has implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) Has(moniker string) bool {
	_, ok := a.registry.Get(moniker)
	return ok
}

// All implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) All() []interfaces.ModuleContractPort {
	mods := a.registry.All()
	result := make([]interfaces.ModuleContractPort, len(mods))
	for i, mod := range mods {
		result[i] = newModuleContractAdapter(mod)
	}
	return result
}

// AllMonikers implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) AllMonikers() []string {
	mods := a.registry.All()
	result := make([]string, len(mods))
	for i, mod := range mods {
		result[i] = mod.Moniker
	}
	sort.Strings(result)
	return result
}

// Count implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) Count() int {
	return len(a.registry.All())
}

// FilterByComponent implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) FilterByComponent(componentType string) []interfaces.ModuleContractPort {
	mods := a.registry.All()
	var result []interfaces.ModuleContractPort
	for _, mod := range mods {
		if mod.HasComponent(componentType) {
			result = append(result, newModuleContractAdapter(mod))
		}
	}
	return result
}

// FindModulesForFile implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) FindModulesForFile(filePath string) []interfaces.ModuleContractPort {
	mods := a.registry.FindModulesForFile(filePath)
	result := make([]interfaces.ModuleContractPort, len(mods))
	for i, mod := range mods {
		result[i] = newModuleContractAdapter(mod)
	}
	return result
}

// WorkspaceRoot implements interfaces.ModuleRegistryPort.
func (a *moduleRegistryAdapter) WorkspaceRoot() string {
	return a.registry.WorkspaceRoot()
}

var _ interfaces.ModuleRegistryPort = (*moduleRegistryAdapter)(nil)

// moduleContractAdapter wraps *modules.ModuleContract to implement interfaces.ModuleContractPort.
type moduleContractAdapter struct {
	module *modules.ModuleContract
}

func newModuleContractAdapter(module *modules.ModuleContract) *moduleContractAdapter {
	return &moduleContractAdapter{module: module}
}

// GetMoniker implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetMoniker() string {
	return a.module.Moniker
}

// GetName implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetName() string {
	return a.module.GetName()
}

// GetDescription implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetDescription() string {
	return a.module.GetDescription()
}

// HasComponent implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) HasComponent(componentType string) bool {
	return a.module.HasComponent(componentType)
}

// GetComponentRoot implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentRoot(componentType string) string {
	return a.module.GetComponentRoot(componentType)
}

// GetComponentRoots implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentRoots() map[string]string {
	return a.module.GetComponentRoots()
}

// GetComponentTypesDisplay implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentTypesDisplay() string {
	return a.module.GetComponentTypesDisplay()
}

// GetComponentAmp implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetComponentAmp(componentName, operation string) float64 {
	// Access amp from component config if defined
	if comp, ok := a.module.Components[componentName]; ok && comp != nil {
		return comp.Amp.GetAmp(operation)
	}
	return 1.0 // Default amplifier
}

// GetDependsOn implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetDependsOn() []string {
	return a.module.DependsOn
}

// GetVersioningScheme implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetVersioningScheme() string {
	if a.module.Versioning != nil {
		return a.module.Versioning.Scheme
	}
	return ""
}

// GetReleaseType implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetReleaseType() string {
	if a.module.Versioning != nil {
		return a.module.Versioning.ReleaseType
	}
	return ""
}

// GetChangelog implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetChangelog() string {
	return a.module.GetChangelog()
}

// HasVersioning implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) HasVersioning() bool {
	return a.module.Versioning != nil && a.module.Versioning.Scheme != ""
}

// GetMetadata implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetMetadata() map[string]interface{} {
	// Convert map[string]string to map[string]interface{}
	if a.module.Metadata == nil {
		return nil
	}
	result := make(map[string]interface{}, len(a.module.Metadata))
	for k, v := range a.module.Metadata {
		result[k] = v
	}
	return result
}

// GetContentHash implements interfaces.ModuleContractPort.
func (a *moduleContractAdapter) GetContentHash() (string, error) {
	return a.module.GetContentHash()
}

var _ interfaces.ModuleContractPort = (*moduleContractAdapter)(nil)
