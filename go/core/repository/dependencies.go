package repository

import (
	"fmt"
	"sort"

	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// ModuleDependency represents a single dependency relationship.
type ModuleDependency struct {
	From string `json:"from" yaml:"from"` // Module that depends on another
	To   string `json:"to" yaml:"to"`     // Module that is depended upon
}

// ModuleDependencyGraph represents the full dependency graph.
type ModuleDependencyGraph struct {
	Modules      []string             `json:"modules" yaml:"modules"`           // All module monikers
	Dependencies map[string][]string  `json:"dependencies" yaml:"dependencies"` // Module -> its dependencies
	Dependents   map[string][]string  `json:"dependents" yaml:"dependents"`     // Module -> modules that depend on it
	Edges        []ModuleDependency   `json:"edges" yaml:"edges"`               // All dependency edges for visualization
	Stats        DependencyGraphStats `json:"stats" yaml:"stats"`               // Graph statistics
}

// DependencyGraphStats provides statistics about the dependency graph.
type DependencyGraphStats struct {
	TotalModules      int `json:"total_modules" yaml:"total_modules"`
	TotalDependencies int `json:"total_dependencies" yaml:"total_dependencies"`
	RootModules       int `json:"root_modules" yaml:"root_modules"`         // Modules with no dependencies
	LeafModules       int `json:"leaf_modules" yaml:"leaf_modules"`         // Modules with no dependents
	MaxDependencies   int `json:"max_dependencies" yaml:"max_dependencies"` // Maximum dependencies for any module
	MaxDependents     int `json:"max_dependents" yaml:"max_dependents"`     // Maximum dependents for any module
}

// GetModuleDependencyGraph builds a complete dependency graph for all modules.
func GetModuleDependencyGraph(rootPath string) (*ModuleDependencyGraph, error) {
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	registry, err := modules.LoadFromWorkspace(rootPath)
	if err != nil {
		return nil, NewRepositoryError("dependencies", rootPath, err, "failed to load module contracts")
	}

	dependencies := registry.GetDependencyGraph()
	dependents := registry.GetReverseDependencyGraph()
	monikers := registry.AllMonikers()

	// Build edges list for visualization
	edges := []ModuleDependency{}
	for from, deps := range dependencies {
		for _, to := range deps {
			edges = append(edges, ModuleDependency{
				From: from,
				To:   to,
			})
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

// calculateGraphStats computes statistics about the dependency graph.
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

// GetChangedModules returns modules that own the given changed files.
func GetChangedModules(changedFiles []string, rootPath string) ([]string, error) {
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	registry, err := modules.LoadFromWorkspace(rootPath)
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

// GetModulesRequiringRebuild returns all modules that need to be rebuilt given a set of changed files.
// This includes:
// 1. Modules that directly own the changed files
// 2. All modules that transitively depend on the changed modules (cache invalidation)
//
// For example, if core changes, this returns core plus all modules that depend on it
// (directly or transitively), because their cached builds are now invalid.
func GetModulesRequiringRebuild(changedFiles []string, rootPath string) ([]string, error) {
	if rootPath == "" {
		var err error
		rootPath, err = GetRepositoryRoot("")
		if err != nil {
			return nil, err
		}
	}

	// Get directly changed modules
	directlyChanged, err := GetChangedModules(changedFiles, rootPath)
	if err != nil {
		return nil, err
	}

	if len(directlyChanged) == 0 {
		return []string{}, nil
	}

	// Get the dependency graph to find dependents
	graph, err := GetModuleDependencyGraph(rootPath)
	if err != nil {
		return nil, err
	}

	// Collect all modules requiring rebuild (changed + transitive dependents)
	requiresRebuild := make(map[string]bool)

	// Add directly changed modules
	for _, m := range directlyChanged {
		requiresRebuild[m] = true
	}

	// For each directly changed module, add all transitive dependents
	for _, changedModule := range directlyChanged {
		addTransitiveDependents(changedModule, graph.Dependents, requiresRebuild)
	}

	// Convert to sorted slice for consistent output
	result := make([]string, 0, len(requiresRebuild))
	for moniker := range requiresRebuild {
		result = append(result, moniker)
	}
	sort.Strings(result)

	return result, nil
}

// addTransitiveDependents recursively adds all modules that depend on the given module
// (directly or transitively) to the result set.
func addTransitiveDependents(module string, dependentsGraph map[string][]string, result map[string]bool) {
	dependents := dependentsGraph[module]
	for _, dependent := range dependents {
		if !result[dependent] {
			result[dependent] = true
			// Recursively add dependents of this dependent
			addTransitiveDependents(dependent, dependentsGraph, result)
		}
	}
}

// GetPlantUMLDiagram generates a PlantUML diagram from the dependency graph.
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

// GetMermaidDiagram generates a Mermaid diagram from the dependency graph.
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

// sanitizeMermaidID converts a module moniker to a valid Mermaid node ID.
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
