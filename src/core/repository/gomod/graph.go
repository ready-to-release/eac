package gomod

import (
	"fmt"

	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// GraphBuilder builds dependency graphs from go.mod files
type GraphBuilder struct {
	mapper   *Mapper
	infos    []*GoModInfo
	rootPath string
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder(mapper *Mapper, rootPath string) *GraphBuilder {
	return &GraphBuilder{
		mapper:   mapper,
		infos:    []*GoModInfo{},
		rootPath: rootPath,
	}
}

// AddGoModInfo adds a parsed go.mod info to the builder
func (gb *GraphBuilder) AddGoModInfo(info *GoModInfo) {
	gb.infos = append(gb.infos, info)
}

// Build constructs the dependency graph from all added go.mod files
func (gb *GraphBuilder) Build() (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Modules:      make(map[string]*ModuleNode),
		Dependencies: make(map[string][]string),
	}

	// First pass: Create all module nodes
	for _, info := range gb.infos {
		// Get moniker for this module
		moniker, err := gb.mapper.GetMonikerFromPath(info.ModulePath)
		if err != nil {
			// If we can't find a moniker, try by module directory
			moniker, err = gb.mapper.GetMonikerFromModuleDir(info.ModuleDir)
			if err != nil {
				return nil, fmt.Errorf("failed to find moniker for %s: %w", info.ModulePath, err)
			}
		}

		node := &ModuleNode{
			Moniker:    moniker,
			ModulePath: info.ModulePath,
			SourceRoot: info.ModuleDir,
			GoModPath:  info.FilePath,
			DependsOn:  []string{},
			UsedBy:     []string{},
		}

		graph.Modules[moniker] = node
		graph.Dependencies[moniker] = []string{}
	}

	// Second pass: Build dependencies
	for _, info := range gb.infos {
		// Get moniker for this module
		moniker, err := gb.mapper.GetMonikerFromPath(info.ModulePath)
		if err != nil {
			moniker, err = gb.mapper.GetMonikerFromModuleDir(info.ModuleDir)
			if err != nil {
				continue // Skip if we can't find moniker
			}
		}

		// Filter to only internal dependencies
		internalPaths := FilterInternalDependencies(info.Requires, gb.mapper.GetBaseModulePath())

		// Convert paths to monikers
		depMonikers, errors := gb.mapper.MapInternalDependenciesToMonikers(internalPaths)
		if len(errors) > 0 {
			// Log errors but continue
			for _, err := range errors {
				fmt.Printf("Warning: %v\n", err)
			}
		}

		// Update node and dependencies map
		if node, ok := graph.Modules[moniker]; ok {
			node.DependsOn = depMonikers
			graph.Dependencies[moniker] = depMonikers
		}
	}

	// Third pass: Calculate reverse dependencies (UsedBy)
	gb.calculateReverseDependencies(graph)

	return graph, nil
}

// calculateReverseDependencies populates the UsedBy field for each node
func (gb *GraphBuilder) calculateReverseDependencies(graph *DependencyGraph) {
	// Clear existing UsedBy lists
	for _, node := range graph.Modules {
		node.UsedBy = []string{}
	}

	// Build reverse mappings
	for moniker, deps := range graph.Dependencies {
		for _, dep := range deps {
			if depNode, ok := graph.Modules[dep]; ok {
				depNode.UsedBy = append(depNode.UsedBy, moniker)
			}
		}
	}
}

// BuildFromDirectory scans a directory for go.mod files and builds the graph
func BuildFromDirectory(rootPath string, registry *modules.Registry, baseModulePath string, excludeDirs []string) (*DependencyGraph, error) {
	// Parse all go.mod files
	infos, err := ParseAllGoMods(rootPath, excludeDirs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod files: %w", err)
	}

	// Create mapper
	mapper := NewMapper(registry, baseModulePath)

	// Create graph builder
	builder := NewGraphBuilder(mapper, rootPath)

	// Add all parsed info
	for _, info := range infos {
		builder.AddGoModInfo(info)
	}

	// Build and return graph
	return builder.Build()
}

// GetModuleByMoniker retrieves a module node by its moniker
func (g *DependencyGraph) GetModuleByMoniker(moniker string) (*ModuleNode, bool) {
	node, exists := g.Modules[moniker]
	return node, exists
}

// GetDependencies returns the dependencies for a given moniker
func (g *DependencyGraph) GetDependencies(moniker string) []string {
	if deps, ok := g.Dependencies[moniker]; ok {
		return deps
	}
	return []string{}
}

// AllMonikers returns all module monikers in the graph
func (g *DependencyGraph) AllMonikers() []string {
	monikers := make([]string, 0, len(g.Modules))
	for moniker := range g.Modules {
		monikers = append(monikers, moniker)
	}
	return monikers
}

// ModuleCount returns the total number of modules in the graph
func (g *DependencyGraph) ModuleCount() int {
	return len(g.Modules)
}
