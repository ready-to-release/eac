package modules

import (
	"fmt"
	"sort"
	"strings"
)

// Registry provides fast access to module contracts
type Registry struct {
	modules            map[string]*ModuleContract // Keyed by moniker
	version            string
	workspaceRoot      string
	modulesByRoot      map[string][]*ModuleContract // Index: root -> modules with that root
	modulesWithRepoPat []*ModuleContract            // Modules that have repo patterns (need checking for all files)
}

// NewRegistry creates a new module registry
func NewRegistry(version, workspaceRoot string) *Registry {
	return &Registry{
		modules:            make(map[string]*ModuleContract),
		version:            version,
		workspaceRoot:      workspaceRoot,
		modulesByRoot:      make(map[string][]*ModuleContract),
		modulesWithRepoPat: make([]*ModuleContract, 0),
	}
}

// Add adds a module contract to the registry
func (r *Registry) Add(module *ModuleContract) error {
	if module.Moniker == "" {
		return fmt.Errorf("cannot add module with empty moniker")
	}

	if _, exists := r.modules[module.Moniker]; exists {
		return fmt.Errorf("module with moniker '%s' already exists in registry", module.Moniker)
	}

	r.modules[module.Moniker] = module

	// Build index by root
	root := normalizeRoot(module.Files.Root)
	r.modulesByRoot[root] = append(r.modulesByRoot[root], module)

	// Track modules with repo patterns
	if len(module.getRepoPatterns()) > 0 {
		r.modulesWithRepoPat = append(r.modulesWithRepoPat, module)
	}

	return nil
}

// normalizeRoot normalizes a root path for indexing
func normalizeRoot(root string) string {
	if root == "" || root == "/" {
		return ""
	}
	// Normalize path separators
	return normalizePathSeparators(root)
}

// Get retrieves a module contract by moniker
func (r *Registry) Get(moniker string) (*ModuleContract, bool) {
	module, exists := r.modules[moniker]
	return module, exists
}

// Has checks if a module exists in the registry
func (r *Registry) Has(moniker string) bool {
	_, exists := r.modules[moniker]
	return exists
}

// All returns all module contracts in the registry
func (r *Registry) All() []*ModuleContract {
	modules := make([]*ModuleContract, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}
	return modules
}

// AllMonikers returns all module monikers sorted alphabetically
func (r *Registry) AllMonikers() []string {
	monikers := make([]string, 0, len(r.modules))
	for moniker := range r.modules {
		monikers = append(monikers, moniker)
	}
	sort.Strings(monikers)
	return monikers
}

// Count returns the number of modules in the registry
func (r *Registry) Count() int {
	return len(r.modules)
}

// Version returns the contract version
func (r *Registry) Version() string {
	return r.version
}

// WorkspaceRoot returns the workspace root path
func (r *Registry) WorkspaceRoot() string {
	return r.workspaceRoot
}

// FilterByType returns all modules of a specific type
func (r *Registry) FilterByType(contractType string) []*ModuleContract {
	var filtered []*ModuleContract
	for _, module := range r.modules {
		if module.Type == contractType {
			filtered = append(filtered, module)
		}
	}
	return filtered
}

// FindByRoot returns modules that match the given root path
func (r *Registry) FindByRoot(rootPath string) []*ModuleContract {
	var matches []*ModuleContract
	for _, module := range r.modules {
		if module.Files.Root == rootPath {
			matches = append(matches, module)
		}
	}
	return matches
}

// GetDependencyGraph returns a map of module dependencies
// Key: module moniker, Value: list of dependency monikers
func (r *Registry) GetDependencyGraph() map[string][]string {
	graph := make(map[string][]string)
	for moniker, module := range r.modules {
		graph[moniker] = module.DependsOn
	}
	return graph
}

// GetReverseDependencyGraph returns a map of reverse dependencies
// Key: module moniker, Value: list of modules that depend on it
func (r *Registry) GetReverseDependencyGraph() map[string][]string {
	graph := make(map[string][]string)

	// Initialize
	for moniker := range r.modules {
		graph[moniker] = []string{}
	}

	// Build reverse graph
	for moniker, module := range r.modules {
		for _, dep := range module.DependsOn {
			graph[dep] = append(graph[dep], moniker)
		}
	}

	return graph
}

// GetUsedBy returns all modules that depend on the given module
// This is computed from depends_on relationships (no longer stored in config)
func (r *Registry) GetUsedBy(moniker string) []string {
	reverseGraph := r.GetReverseDependencyGraph()
	return reverseGraph[moniker]
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

	// 2. Check modules with root="" (repository-wide modules)
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

	// 3. Check modules with repo patterns (they can match files anywhere)
	for _, module := range r.modulesWithRepoPat {
		if !checked[module.Moniker] {
			checked[module.Moniker] = true
			if module.MatchesFile(filePath) {
				matches = append(matches, module)
			}
		}
	}

	return matches
}

// splitPath splits a path into components
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

// joinPath joins path components
func joinPath(parts []string) string {
	return strings.Join(parts, "/")
}
