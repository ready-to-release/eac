package repository

import (
	"fmt"
	"sort"

	"github.com/ready-to-release/eac/src/internal/contracts/modules"
)

// ModuleDependency represents a single dependency relationship
type ModuleDependency struct {
	From string `json:"from" yaml:"from"` // Module that depends on another
	To   string `json:"to" yaml:"to"`     // Module that is depended upon
}

// ModuleDependencyGraph represents the full dependency graph
type ModuleDependencyGraph struct {
	Modules      []string                  `json:"modules" yaml:"modules"`           // All module monikers
	Dependencies map[string][]string       `json:"dependencies" yaml:"dependencies"` // Module -> its dependencies
	Dependents   map[string][]string       `json:"dependents" yaml:"dependents"`     // Module -> modules that depend on it
	Edges        []ModuleDependency        `json:"edges" yaml:"edges"`               // All dependency edges for visualization
	Stats        DependencyGraphStats      `json:"stats" yaml:"stats"`               // Graph statistics
}

// DependencyGraphStats provides statistics about the dependency graph
type DependencyGraphStats struct {
	TotalModules      int `json:"total_modules" yaml:"total_modules"`
	TotalDependencies int `json:"total_dependencies" yaml:"total_dependencies"`
	RootModules       int `json:"root_modules" yaml:"root_modules"`         // Modules with no dependencies
	LeafModules       int `json:"leaf_modules" yaml:"leaf_modules"`         // Modules with no dependents
	MaxDependencies   int `json:"max_dependencies" yaml:"max_dependencies"` // Maximum dependencies for any module
	MaxDependents     int `json:"max_dependents" yaml:"max_dependents"`     // Maximum dependents for any module
}

// ExecutionPlan represents a layered execution plan for modules
type ExecutionPlan struct {
	Layers         [][]string `json:"layers" yaml:"layers"`                   // Modules grouped by dependency layer
	ExecutionOrder []string   `json:"execution_order" yaml:"execution_order"` // Flattened order of all modules
	LayerCount     int        `json:"layer_count" yaml:"layer_count"`         // Number of layers
}

// GetModuleDependencyGraph builds a complete dependency graph for all modules
func GetModuleDependencyGraph(rootPath string, version string) (*ModuleDependencyGraph, error) {
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	registry, err := modules.LoadFromWorkspace(rootPath, version)
	if err != nil {
		return nil, NewRepositoryError("dependencies", rootPath, err, "failed to load module contracts")
	}

	dependencies := registry.GetDependencyGraph()
	dependents := registry.GetReverseDependencyGraph()
	monikers := registry.AllMonikers()

	// Build edges list for visualization
	edges := []ModuleDependency{}
	totalDeps := 0
	for from, deps := range dependencies {
		for _, to := range deps {
			edges = append(edges, ModuleDependency{
				From: from,
				To:   to,
			})
			totalDeps++
		}
	}

	// Calculate statistics
	stats := calculateGraphStats(monikers, dependencies, dependents)

	return &ModuleDependencyGraph{
		Modules:      monikers,
		Dependencies: dependencies,
		Dependents:   dependents,
		Edges:        edges,
		Stats:        stats,
	}, nil
}

// calculateGraphStats computes statistics about the dependency graph
func calculateGraphStats(monikers []string, dependencies, dependents map[string][]string) DependencyGraphStats {
	stats := DependencyGraphStats{
		TotalModules: len(monikers),
	}

	rootCount := 0
	leafCount := 0
	maxDeps := 0
	maxDependents := 0

	for _, moniker := range monikers {
		deps := dependencies[moniker]
		depts := dependents[moniker]

		stats.TotalDependencies += len(deps)

		if len(deps) == 0 {
			rootCount++
		}
		if len(depts) == 0 {
			leafCount++
		}
		if len(deps) > maxDeps {
			maxDeps = len(deps)
		}
		if len(depts) > maxDependents {
			maxDependents = len(depts)
		}
	}

	stats.RootModules = rootCount
	stats.LeafModules = leafCount
	stats.MaxDependencies = maxDeps
	stats.MaxDependents = maxDependents

	return stats
}

// CalculateExecutionOrder performs topological sort on module dependencies
// Returns an ExecutionPlan with modules grouped into layers for parallel execution
//
// Algorithm: Kahn's topological sort
// - Layer 0: Modules with no dependencies (can run in parallel)
// - Layer N: Modules that depend only on layers 0..N-1 (can run in parallel within layer)
//
// If monikers is empty or nil, calculates order for all modules
// Returns error if circular dependencies detected
func CalculateExecutionOrder(monikers []string, rootPath string, version string) (*ExecutionPlan, error) {
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	registry, err := modules.LoadFromWorkspace(rootPath, version)
	if err != nil {
		return nil, NewRepositoryError("execution-order", rootPath, err, "failed to load module contracts")
	}

	// If no monikers specified, use all modules
	if len(monikers) == 0 {
		monikers = registry.AllMonikers()
	}

	// Build set of all modules to include (input modules + their dependencies)
	allModules := make(map[string]bool)
	for _, moniker := range monikers {
		allModules[moniker] = true
		if err := addDependenciesRecursive(moniker, registry, allModules); err != nil {
			return nil, err
		}
	}

	// Build dependency graph for included modules only
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for moniker := range allModules {
		module, exists := registry.Get(moniker)
		if !exists {
			return nil, NewRepositoryError("execution-order", rootPath, nil,
				fmt.Sprintf("module '%s' not found in registry", moniker))
		}

		graph[moniker] = []string{}
		inDegree[moniker] = 0

		// Only include dependencies that are in our set
		for _, dep := range module.DependsOn {
			if allModules[dep] {
				graph[moniker] = append(graph[moniker], dep)
				inDegree[moniker]++
			}
		}
	}

	// Kahn's algorithm for topological sort
	layers := [][]string{}
	processed := make(map[string]bool)

	for len(processed) < len(allModules) {
		layer := []string{}

		// Find all modules with in-degree 0 (no unprocessed dependencies)
		for moniker := range allModules {
			if !processed[moniker] && inDegree[moniker] == 0 {
				layer = append(layer, moniker)
			}
		}

		if len(layer) == 0 {
			// Circular dependency detected
			remaining := []string{}
			for moniker := range allModules {
				if !processed[moniker] {
					remaining = append(remaining, moniker)
				}
			}
			return nil, NewRepositoryError("execution-order", rootPath, nil,
				fmt.Sprintf("circular dependency detected among: %v", remaining))
		}

		// Sort layer for consistent output
		sort.Strings(layer)
		layers = append(layers, layer)

		// Mark layer as processed and update in-degrees
		for _, moniker := range layer {
			processed[moniker] = true

			// Decrease in-degree for all dependents
			for dependent := range allModules {
				if processed[dependent] {
					continue
				}
				for _, dep := range graph[dependent] {
					if dep == moniker {
						inDegree[dependent]--
					}
				}
			}
		}
	}

	// Flatten layers to execution order
	executionOrder := []string{}
	for _, layer := range layers {
		executionOrder = append(executionOrder, layer...)
	}

	return &ExecutionPlan{
		Layers:         layers,
		ExecutionOrder: executionOrder,
		LayerCount:     len(layers),
	}, nil
}

// addDependenciesRecursive recursively adds all dependencies of a module
// Returns error listing all missing dependencies found
func addDependenciesRecursive(moniker string, registry *modules.Registry, result map[string]bool) error {
	module, exists := registry.Get(moniker)
	if !exists {
		return fmt.Errorf("module '%s' not found in registry", moniker)
	}

	var missingDeps []string

	for _, dep := range module.DependsOn {
		// Check if dependency exists in registry
		if _, exists := registry.Get(dep); !exists {
			missingDeps = append(missingDeps, fmt.Sprintf("%s->%s", moniker, dep))
			continue
		}

		if !result[dep] {
			result[dep] = true
			if err := addDependenciesRecursive(dep, registry, result); err != nil {
				return err
			}
		}
	}

	if len(missingDeps) > 0 {
		return fmt.Errorf("missing dependency contracts: %v", missingDeps)
	}

	return nil
}

// GetChangedModules returns modules that own the given changed files
func GetChangedModules(changedFiles []string, rootPath string, version string) ([]string, error) {
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	registry, err := modules.LoadFromWorkspace(rootPath, version)
	if err != nil {
		return nil, NewRepositoryError("changed-modules", rootPath, err, "failed to load module contracts")
	}

	changedModules := make(map[string]bool)

	for _, filePath := range changedFiles {
		if filePath == "" {
			continue
		}
		matchingModules := registry.FindModulesForFile(filePath)
		for _, module := range matchingModules {
			changedModules[module.Moniker] = true
		}
	}

	// Convert to sorted slice for consistent output
	result := []string{}
	for moniker := range changedModules {
		result = append(result, moniker)
	}
	sort.Strings(result)

	return result, nil
}

// GetPlantUMLDiagram generates a PlantUML diagram from the dependency graph
func GetPlantUMLDiagram(graph *ModuleDependencyGraph) string {
	output := "@startuml\n"
	output += "!theme plain\n"
	output += "skinparam componentStyle rectangle\n\n"
	output += "title Module Dependency Graph\n\n"

	// Add all modules as components
	for _, moniker := range graph.Modules {
		output += fmt.Sprintf("component [%s]\n", moniker)
	}

	output += "\n"

	// Add dependency edges
	for _, edge := range graph.Edges {
		output += fmt.Sprintf("[%s] --> [%s]\n", edge.From, edge.To)
	}

	output += "\n@enduml\n"
	return output
}

// GetMermaidDiagram generates a Mermaid diagram from the dependency graph
func GetMermaidDiagram(graph *ModuleDependencyGraph) string {
	output := "```mermaid\n"
	output += "graph TD\n"

	// Add dependency edges (nodes are created automatically)
	for _, edge := range graph.Edges {
		// Replace hyphens with underscores for valid Mermaid IDs
		fromID := sanitizeMermaidID(edge.From)
		toID := sanitizeMermaidID(edge.To)
		output += fmt.Sprintf("    %s[\"%s\"] --> %s[\"%s\"]\n", fromID, edge.From, toID, edge.To)
	}

	output += "```\n"
	return output
}

// sanitizeMermaidID converts a module moniker to a valid Mermaid node ID
func sanitizeMermaidID(moniker string) string {
	// Replace hyphens and other special characters with underscores
	result := ""
	for _, ch := range moniker {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			result += string(ch)
		} else {
			result += "_"
		}
	}
	return result
}
