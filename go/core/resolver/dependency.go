package resolver

import (
	"fmt"
	"sort"
)

// DependencyGraph manages component build order within a module.
// It uses build_after relationships from component types to determine
// the correct execution order.
type DependencyGraph struct {
	module string
	// edges maps component name -> components it depends on
	edges map[string][]string
	// nodes contains all component names in the graph
	nodes map[string]bool
}

// NewDependencyGraph creates a dependency graph for a module.
// It extracts build_after relationships from component type configurations.
//
// Parameters:
//   - module: the module moniker
//   - enabledComponents: map of component name -> component type name
//   - getBuildAfter: function that returns build_after for a component type
func NewDependencyGraph(module string, enabledComponents map[string]string, getBuildAfter func(compType string) []string) *DependencyGraph {
	g := &DependencyGraph{
		module: module,
		edges:  make(map[string][]string),
		nodes:  make(map[string]bool),
	}

	// Build set of enabled component types for quick lookup
	enabledTypes := make(map[string]bool)
	for _, compType := range enabledComponents {
		enabledTypes[compType] = true
	}

	// Build reverse map: component type -> component names
	typeToComponents := make(map[string][]string)
	for compName, compType := range enabledComponents {
		g.nodes[compName] = true
		typeToComponents[compType] = append(typeToComponents[compType], compName)
	}

	// Build dependency edges from build_after
	for compName, compType := range enabledComponents {
		buildAfter := getBuildAfter(compType)
		for _, depType := range buildAfter {
			// Only add edge if dependency type is enabled in this module
			if enabledTypes[depType] {
				// Find all components of the dependency type
				for _, depComp := range typeToComponents[depType] {
					g.edges[compName] = append(g.edges[compName], depComp)
				}
			}
		}
	}

	return g
}

// DependsOn returns component names that must complete before the given component.
func (g *DependencyGraph) DependsOn(component string) []string {
	if g == nil {
		return nil
	}
	deps := g.edges[component]
	if len(deps) == 0 {
		return nil
	}
	// Return a copy to avoid mutation
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// HasCycle returns true if the graph contains a cycle.
// This would indicate a circular build_after dependency, which is invalid.
func (g *DependencyGraph) HasCycle() bool {
	if g == nil {
		return false
	}

	// Standard DFS-based cycle detection
	// States: 0 = unvisited, 1 = visiting, 2 = visited
	state := make(map[string]int)

	var visit func(node string) bool
	visit = func(node string) bool {
		if state[node] == 1 {
			return true // Back edge found - cycle!
		}
		if state[node] == 2 {
			return false // Already processed
		}

		state[node] = 1 // Mark as visiting

		for _, dep := range g.edges[node] {
			if visit(dep) {
				return true
			}
		}

		state[node] = 2 // Mark as visited
		return false
	}

	for node := range g.nodes {
		if state[node] == 0 {
			if visit(node) {
				return true
			}
		}
	}

	return false
}

// TopologicalOrder returns components in dependency order (dependencies first).
// Returns an error if the graph contains a cycle.
func (g *DependencyGraph) TopologicalOrder() ([]string, error) {
	if g == nil {
		return nil, nil
	}

	if g.HasCycle() {
		return nil, fmt.Errorf("dependency cycle detected in module %s", g.module)
	}

	// Kahn's algorithm for topological sort
	// Calculate in-degree for each node
	inDegree := make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = 0
	}
	for _, deps := range g.edges {
		for _, dep := range deps {
			inDegree[dep]++ // This is wrong - we need reverse edges
		}
	}

	// Actually, we need to reverse the logic: if A depends on B (A -> B in edges),
	// then B must come before A. So we need to track in-degree of nodes
	// where in-degree means "how many nodes must come AFTER this node"
	// But for topological sort, we want "how many nodes must come BEFORE"

	// Let me recompute: edges[A] = [B, C] means A depends on B and C
	// So B and C must come before A
	// In topological terms, B -> A and C -> A (A has in-degree 2)
	inDegree = make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = len(g.edges[node]) // Number of dependencies
	}

	// Start with nodes that have no dependencies
	var queue []string
	for node := range g.nodes {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}
	// Sort for deterministic order
	sort.Strings(queue)

	var result []string
	for len(queue) > 0 {
		// Pop first node
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// For each node that depends on this one, decrement in-degree
		for dependent, deps := range g.edges {
			for _, dep := range deps {
				if dep == node {
					inDegree[dependent]--
					if inDegree[dependent] == 0 {
						queue = append(queue, dependent)
						sort.Strings(queue) // Keep sorted for determinism
					}
				}
			}
		}
	}

	// If not all nodes processed, there's a cycle (shouldn't happen if HasCycle passed)
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("internal error: topological sort incomplete")
	}

	return result, nil
}

// Nodes returns all component names in the graph.
func (g *DependencyGraph) Nodes() []string {
	if g == nil {
		return nil
	}
	nodes := make([]string, 0, len(g.nodes))
	for node := range g.nodes {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	return nodes
}

// Module returns the module name for this graph.
func (g *DependencyGraph) Module() string {
	if g == nil {
		return ""
	}
	return g.module
}
