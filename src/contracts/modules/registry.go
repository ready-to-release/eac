package modules

import (
	"fmt"
	"sort"
)

// Registry provides fast access to module contracts
type Registry struct {
	modules       map[string]*ModuleContract // Keyed by moniker
	version       string
	workspaceRoot string
}

// NewRegistry creates a new module registry
func NewRegistry(version, workspaceRoot string) *Registry {
	return &Registry{
		modules:       make(map[string]*ModuleContract),
		version:       version,
		workspaceRoot: workspaceRoot,
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
	return nil
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
		if module.Source.Root == rootPath {
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
