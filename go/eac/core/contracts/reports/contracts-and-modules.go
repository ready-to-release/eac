package reports

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// Global process-level cache for module contracts
// This prevents repeated parsing of module.contract.yaml files across parallel test packages.
var (
	globalModuleContractCache     *ModuleContractCache
	globalModuleContractCacheOnce sync.Once
)

// ModuleContractCache provides cached module contract data.
// Thread-safe for concurrent access by parallel test packages.
// No cache invalidation - data persists for process lifetime.
type ModuleContractCache struct {
	mu sync.RWMutex

	// workspaceRoot is the workspace root used to populate this cache
	workspaceRoot string

	// populated indicates if the cache has been loaded and validated
	populated bool

	// report is the cached module contract report (validated)
	report *ModuleContractReport
}

// NewModuleContractCache creates a new empty cache.
func NewModuleContractCache() *ModuleContractCache {
	return &ModuleContractCache{}
}

// EnsurePopulated ensures the cache is populated for the given workspace root.
// If already populated for the same root, this is a no-op.
// Thread-safe with double-checked locking pattern.
//
// Returns error if:
//   - Module loading fails
//   - Module validation fails
//   - Report building fails
func (c *ModuleContractCache) EnsurePopulated(workspaceRoot string) error {
	// Fast path: read lock to check if already populated
	c.mu.RLock()
	if c.populated && c.workspaceRoot == workspaceRoot {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	// Slow path: write lock to populate
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: another goroutine may have populated while we waited
	if c.populated && c.workspaceRoot == workspaceRoot {
		return nil
	}

	// Reset if workspace root changed
	if c.workspaceRoot != workspaceRoot {
		c.report = nil
		c.populated = false
		c.workspaceRoot = workspaceRoot
	}

	// Load and validate module contracts
	// (VALIDATION happens inside LoadFromWorkspace)
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to load module contracts: %w", err)
	}

	// Build report from validated registry
	report, err := buildModuleContractReport(registry)
	if err != nil {
		return fmt.Errorf("failed to build module contract report: %w", err)
	}

	// Cache the validated report
	c.report = report
	c.populated = true
	return nil
}

// GetReport returns the cached module contract report.
// MUST call EnsurePopulated first, otherwise panics.
// This is a programmer error - all callers must ensure population.
func (c *ModuleContractCache) GetReport() *ModuleContractReport {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.populated {
		panic("ModuleContractCache not populated - call EnsurePopulated() first")
	}

	return c.report
}

// ModuleContractReport contains information about loaded module contracts.
type ModuleContractReport struct {
	TotalModules int
	Modules      []*modules.ModuleContract
	Registry     *modules.Registry
}

// GetModuleContracts loads and reports on all module contracts.
// Uses global in-memory cache to avoid repeated file I/O and parsing.
//
// BACKWARD COMPATIBLE: API unchanged, caching is internal optimization.
//
// Thread-safe: Multiple goroutines can call concurrently.
// First caller loads and validates, subsequent callers use cached data.
//
// Parameters:
//   - workspaceRoot: Repository root (if empty, will be detected automatically)
//
// Returns:
//   - ModuleContractReport containing all loaded contracts and metadata
//   - Error if module loading fails
//
// Example:
//
//	report, err := reports.GetModuleContracts("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total modules: %d\n", report.TotalModules)
func GetModuleContracts(workspaceRoot string) (*ModuleContractReport, error) {
	// Initialize global cache once per process
	globalModuleContractCacheOnce.Do(func() {
		globalModuleContractCache = NewModuleContractCache()
	})

	// Ensure cache is populated (no-op if already loaded)
	if err := globalModuleContractCache.EnsurePopulated(workspaceRoot); err != nil {
		return nil, err
	}

	// Return cached validated report
	return globalModuleContractCache.GetReport(), nil
}

// buildModuleContractReport builds a report from a validated registry.
// Extracted from GetModuleContracts for better separation of concerns.
// PRIVATE helper function.
func buildModuleContractReport(registry *modules.Registry) (*ModuleContractReport, error) {
	// Get all modules sorted by moniker
	allModules := registry.All()
	sort.Slice(allModules, func(i, j int) bool {
		return allModules[i].Moniker < allModules[j].Moniker
	})

	report := &ModuleContractReport{
		TotalModules: len(allModules),
		Modules:      allModules,
		Registry:     registry,
	}

	return report, nil
}

// FormatReport returns a formatted string representation of the module contracts.
func (r *ModuleContractReport) FormatReport() string {
	var sb strings.Builder

	sb.WriteString("=== Module Contracts Report ===\n\n")
	sb.WriteString(fmt.Sprintf("✅ Loaded %d module contracts (version: %s)\n\n", r.TotalModules, r.Registry.Version()))

	// List all modules
	sb.WriteString("=== Modules ===\n")
	for i, module := range r.Modules {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, module.Moniker))
		sb.WriteString(fmt.Sprintf("   Name: %s\n", module.Name))
		sb.WriteString(fmt.Sprintf("   Packages: %s\n", module.GetComponentTypesDisplay()))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", module.Description))

		// Packages
		roots := module.GetComponentRoots()
		if len(roots) > 0 {
			sb.WriteString("   Packages:\n")
			for pkgName, root := range roots {
				sb.WriteString(fmt.Sprintf("     - %s: %s\n", pkgName, root))
			}
		}

		// Dependencies
		if len(module.DependsOn) > 0 {
			sb.WriteString(fmt.Sprintf("   Depends on: %v\n", module.DependsOn))
		}

		// Used by (computed from registry)
		usedBy := r.Registry.GetUsedBy(module.Moniker)
		if len(usedBy) > 0 {
			sb.WriteString(fmt.Sprintf("   Used by: %v\n", usedBy))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatCompact returns a compact one-line-per-module format.
func (r *ModuleContractReport) FormatCompact() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Module Contracts (%d modules) ===\n\n", r.TotalModules))

	for _, module := range r.Modules {
		pkgList := module.GetComponentTypesDisplay()
		sb.WriteString(fmt.Sprintf("%-30s %s\n", module.Moniker, pkgList))
	}

	return sb.String()
}

// GetModuleByMoniker returns a specific module contract by moniker.
func (r *ModuleContractReport) GetModuleByMoniker(moniker string) (*modules.ModuleContract, bool) {
	return r.Registry.Get(moniker)
}

// GetModulesByComponent returns all modules that have a specific component type.
func (r *ModuleContractReport) GetModulesByComponent(compType string) []*modules.ModuleContract {
	return r.Registry.FilterByComponent(compType)
}

// GetModulesByRoot returns modules that have a component at the given root path.
func (r *ModuleContractReport) GetModulesByRoot(root string) []*modules.ModuleContract {
	return r.Registry.FindByRoot(root)
}

// GetDependencyGraph returns the dependency relationships.
func (r *ModuleContractReport) GetDependencyGraph() map[string][]string {
	return r.Registry.GetDependencyGraph()
}

// GetReverseDependencyGraph returns the reverse dependency relationships.
func (r *ModuleContractReport) GetReverseDependencyGraph() map[string][]string {
	return r.Registry.GetReverseDependencyGraph()
}

// GetModulesWithPattern returns modules that use a specific glob pattern.
func (r *ModuleContractReport) GetModulesWithPattern(pattern string) []*modules.ModuleContract {
	var result []*modules.ModuleContract
	for _, module := range r.Modules {
		// Check patterns in all components
		for _, pkg := range module.Components {
			if pkg == nil || pkg.Patterns == nil {
				continue
			}
			allPatterns := append(pkg.Patterns.Source, pkg.Patterns.Config...)
			allPatterns = append(allPatterns, pkg.Patterns.Tests...)
			for _, p := range allPatterns {
				if p == pattern {
					result = append(result, module)
					break
				}
			}
		}
	}
	return result
}

// PrintSummary prints a concise summary of the loaded contracts.
func (r *ModuleContractReport) PrintSummary() {
	// Count modules by package type
	pkgCount := make(map[string]int)
	for _, module := range r.Modules {
		for _, pkg := range module.GetEnabledComponents() {
			pkgCount[pkg]++
		}
	}
}
