package modules

import (
	"fmt"
	"sort"
	"strings"
)

// Registry provides fast access to module domain.
type Registry struct {
	modules        map[string]*ModuleContract   // Keyed by moniker
	version        string
	workspaceRoot  string
	modulesByRoot  map[string][]*ModuleContract // Index: package root -> modules with packages at that root
	reverseDepGraph map[string][]string          // Pre-computed reverse dependency graph: dep -> dependents
}

// NewRegistry creates a new module registry.
func NewRegistry(version, workspaceRoot string) *Registry {
	return &Registry{
		modules:         make(map[string]*ModuleContract),
		version:         version,
		workspaceRoot:   workspaceRoot,
		modulesByRoot:   make(map[string][]*ModuleContract),
		reverseDepGraph: make(map[string][]string),
	}
}

// Add adds a module contract to the registry.
func (r *Registry) Add(module *ModuleContract) error {
	if module.Moniker == "" {
		return fmt.Errorf("cannot add module with empty moniker")
	}

	if _, exists := r.modules[module.Moniker]; exists {
		return fmt.Errorf("module with moniker '%s' already exists in registry", module.Moniker)
	}

	r.modules[module.Moniker] = module

	// Build index by each component root
	for _, root := range module.GetComponentRoots() {
		normalizedRoot := normalizeRoot(root)
		r.modulesByRoot[normalizedRoot] = append(r.modulesByRoot[normalizedRoot], module)
	}

	// Update reverse dependency graph: for each dependency of this module,
	// record that this module depends on it
	for _, dep := range module.DependsOn {
		r.reverseDepGraph[dep] = append(r.reverseDepGraph[dep], module.Moniker)
	}

	return nil
}

// normalizeRoot normalizes a root path for indexing.
func normalizeRoot(root string) string {
	if root == "" || root == "/" {
		return ""
	}
	// Normalize path separators
	return normalizePathSeparators(root)
}

// Get retrieves a module contract by moniker.
func (r *Registry) Get(moniker string) (*ModuleContract, bool) {
	module, exists := r.modules[moniker]
	return module, exists
}

// Has checks if a module exists in the registry.
func (r *Registry) Has(moniker string) bool {
	_, exists := r.modules[moniker]
	return exists
}

// All returns all module contracts in the registry.
func (r *Registry) All() []*ModuleContract {
	modules := make([]*ModuleContract, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}
	return modules
}

// AllMonikers returns all module monikers sorted alphabetically.
func (r *Registry) AllMonikers() []string {
	monikers := make([]string, 0, len(r.modules))
	for moniker := range r.modules {
		monikers = append(monikers, moniker)
	}
	sort.Strings(monikers)
	return monikers
}

// Count returns the number of modules in the registry.
func (r *Registry) Count() int {
	return len(r.modules)
}

// Version returns the contract version.
func (r *Registry) Version() string {
	return r.version
}

// WorkspaceRoot returns the workspace root path.
func (r *Registry) WorkspaceRoot() string {
	return r.workspaceRoot
}

// FilterByComponent returns all modules that have the specified component type enabled.
// This is the preferred way to filter modules in the multi-component architecture.
func (r *Registry) FilterByComponent(componentType string) []*ModuleContract {
	var filtered []*ModuleContract
	for _, module := range r.modules {
		if module.HasComponent(componentType) {
			filtered = append(filtered, module)
		}
	}
	return filtered
}

// FilterByComponents returns all modules that have ALL of the specified component types enabled.
func (r *Registry) FilterByComponents(componentTypes ...string) []*ModuleContract {
	var filtered []*ModuleContract
	for _, module := range r.modules {
		hasAll := true
		for _, compType := range componentTypes {
			if !module.HasComponent(compType) {
				hasAll = false
				break
			}
		}
		if hasAll {
			filtered = append(filtered, module)
		}
	}
	return filtered
}

// FilterByAnyComponent returns all modules that have ANY of the specified component types enabled.
func (r *Registry) FilterByAnyComponent(componentTypes ...string) []*ModuleContract {
	var filtered []*ModuleContract
	for _, module := range r.modules {
		for _, compType := range componentTypes {
			if module.HasComponent(compType) {
				filtered = append(filtered, module)
				break
			}
		}
	}
	return filtered
}

// FindByRoot returns modules that have a component at the given root path.
func (r *Registry) FindByRoot(rootPath string) []*ModuleContract {
	normalizedPath := normalizeRoot(rootPath)
	if modules, ok := r.modulesByRoot[normalizedPath]; ok {
		return modules
	}
	return nil
}

// GetDependencyGraph returns a map of module dependencies
// Key: module moniker, Value: list of dependency monikers.
func (r *Registry) GetDependencyGraph() map[string][]string {
	graph := make(map[string][]string)
	for moniker, module := range r.modules {
		graph[moniker] = module.DependsOn
	}
	return graph
}

// GetReverseDependencyGraph returns a map of reverse dependencies
// Key: module moniker, Value: list of modules that depend on it.
// Returns a shallow copy so callers cannot mutate internal state.
func (r *Registry) GetReverseDependencyGraph() map[string][]string {
	graph := make(map[string][]string, len(r.modules))

	// Ensure every registered module has an entry (even if no dependents)
	for moniker := range r.modules {
		graph[moniker] = []string{}
	}

	// Copy pre-computed reverse dependencies
	for dep, dependents := range r.reverseDepGraph {
		graph[dep] = append(graph[dep], dependents...)
	}

	return graph
}

// GetUsedBy returns all modules that depend on the given module.
// This uses the pre-computed reverse dependency graph built during Add.
func (r *Registry) GetUsedBy(moniker string) []string {
	return r.reverseDepGraph[moniker]
}

// FindModulesForFile returns all modules that match a given file path.
// Returns empty slice if no modules match (orphan file).
func (r *Registry) FindModulesForFile(filePath string) []*ModuleContract {
	var matches []*ModuleContract
	checked := make(map[string]bool) // Track which modules we've already checked

	path := normalizePathSeparators(filePath)

	// 1. Find modules by matching root prefix (fast path)
	// Try progressively shorter prefixes of the file path
	parts := splitPath(path)
	for i := len(parts); i > 0; i-- {
		prefix := joinPath(parts[:i])
		if modules, ok := r.modulesByRoot[prefix]; ok {
			for _, module := range modules {
				if !checked[module.Moniker] {
					checked[module.Moniker] = true
					if module.MatchesFile(filePath) {
						matches = append(matches, module)
					}
				}
			}
		}
	}

	// 2. Check modules with root="" (repository-wide packages)
	if modules, ok := r.modulesByRoot[""]; ok {
		for _, module := range modules {
			if !checked[module.Moniker] {
				checked[module.Moniker] = true
				if module.MatchesFile(filePath) {
					matches = append(matches, module)
				}
			}
		}
	}

	return matches
}

// splitPath splits a path into components.
func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// joinPath joins path components.
func joinPath(parts []string) string {
	return strings.Join(parts, "/")
}
