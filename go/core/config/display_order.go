package config

import (
	"slices"
	"sort"
)

// DisplayOrder holds precomputed display ordering for modules and components.
// Computed once during config loading, reused by all display consumers.
type DisplayOrder struct {
	// Modules in display order (baseline first, then by depth/group/config-order).
	Modules []string

	// Depth per module: -1 = baseline tooling, 0 = no deps, N = max dep depth.
	Depth map[string]int

	// IsBaseline tracks modules that declared depends_on: [root].
	IsBaseline map[string]bool

	// Components maps module moniker to ordered component names.
	// Respects intra-module component dependencies (build_after, depends_on)
	// and falls back to ComponentOrder (YAML declaration order).
	Components map[string][]string
}

// computeDisplayOrder builds the display ordering from module dependencies,
// baseline tracking, and component-level dependencies.
// Must be called after expandModuleGroups (module deps resolved) and
// LoadComponentTypes (build_after available).
func (c *RepositoryConfig) computeDisplayOrder(compTypes *ComponentTypesConfig) {
	order := &DisplayOrder{
		Depth:      make(map[string]int),
		IsBaseline: c.baselineModules,
		Components: make(map[string][]string),
	}
	if order.IsBaseline == nil {
		order.IsBaseline = make(map[string]bool)
	}

	// Build indexes for stable sorting.
	declOrder := make(map[string]int, len(c.Modules))
	groupOf := make(map[string]string, len(c.Modules))
	monikers := make([]string, 0, len(c.Modules))
	deps := make(map[string][]string, len(c.Modules))
	for i, m := range c.Modules {
		declOrder[m.Moniker] = i
		groupOf[m.Moniker] = m.ModuleGroup
		monikers = append(monikers, m.Moniker)
		deps[m.Moniker] = m.DependsOn
	}

	// Compute depth via fixed-point iteration.
	// Initialize all modules to depth 0 so the ok-check finds them.
	for _, m := range monikers {
		order.Depth[m] = 0
	}
	changed := true
	for changed {
		changed = false
		for _, m := range monikers {
			for _, dep := range deps[m] {
				if d, ok := order.Depth[dep]; ok && d+1 > order.Depth[m] {
					order.Depth[m] = d + 1
					changed = true
				}
			}
		}
	}

	// Override baseline modules to depth -1.
	for m := range order.IsBaseline {
		order.Depth[m] = -1
	}

	// Sort modules: (depth, grouped-before-ungrouped, group-name, declaration-order).
	sorted := make([]string, len(monikers))
	copy(sorted, monikers)
	sort.SliceStable(sorted, func(i, j int) bool {
		di, dj := order.Depth[sorted[i]], order.Depth[sorted[j]]
		if di != dj {
			return di < dj
		}
		gi, gj := groupOf[sorted[i]], groupOf[sorted[j]]
		if gi != gj {
			if gi == "" || gj == "" {
				return gj == ""
			}
			return gi < gj
		}
		return declOrder[sorted[i]] < declOrder[sorted[j]]
	})
	order.Modules = sorted

	// Compute component ordering for each module.
	for i := range c.Modules {
		order.Components[c.Modules[i].Moniker] = computeComponentOrder(&c.Modules[i], compTypes)
	}

	c.DisplayOrder = order
}

// computeComponentOrder returns components in dependency-respecting order.
// Priority: DependsOn > build_after > ComponentOrder (YAML declaration order).
func computeComponentOrder(m *Module, compTypes *ComponentTypesConfig) []string {
	if len(m.Components) == 0 {
		return nil
	}

	// Build dependency graph: compDeps[A] = [B, C] means A depends on B and C.
	compDeps := make(map[string][]string, len(m.Components))
	for name, entry := range m.Components {
		var deps []string

		// 1. Explicit DependsOn from ComponentEntry.
		if entry != nil {
			for _, dep := range entry.DependsOn {
				if _, ok := m.Components[dep]; ok {
					deps = append(deps, dep)
				}
			}
		}

		// 2. build_after from ComponentTypes.
		if compTypes != nil && entry != nil {
			compType := name
			if entry.Type != "" {
				compType = entry.Type
			}
			ct := compTypes.Get(compType)
			if ct != nil {
				for _, afterType := range ct.BuildAfter {
					for otherName, otherEntry := range m.Components {
						if otherName == name {
							continue
						}
						otherType := otherName
						if otherEntry != nil && otherEntry.Type != "" {
							otherType = otherEntry.Type
						}
						if otherType == afterType && !slices.Contains(deps, otherName) {
							deps = append(deps, otherName)
						}
					}
				}
			}
		}

		compDeps[name] = deps
	}

	// Build declaration order index from ComponentOrder.
	declOrder := make(map[string]int, len(m.Components))
	for i, name := range m.ComponentOrder {
		declOrder[name] = i
	}
	nextIdx := len(m.ComponentOrder)
	for name := range m.Components {
		if _, ok := declOrder[name]; !ok {
			declOrder[name] = nextIdx
			nextIdx++
		}
	}

	// Topological sort (Kahn's algorithm) with declaration order as tiebreaker.
	dependents := make(map[string][]string, len(m.Components))
	inDegree := make(map[string]int, len(m.Components))
	for name := range m.Components {
		inDegree[name] = len(compDeps[name])
		for _, dep := range compDeps[name] {
			dependents[dep] = append(dependents[dep], name)
		}
	}

	var result []string
	processed := make(map[string]bool, len(m.Components))

	for len(processed) < len(m.Components) {
		var ready []string
		for name := range m.Components {
			if !processed[name] && inDegree[name] == 0 {
				ready = append(ready, name)
			}
		}

		if len(ready) == 0 {
			// Cycle — append remaining in declaration order.
			var remaining []string
			for name := range m.Components {
				if !processed[name] {
					remaining = append(remaining, name)
				}
			}
			sort.Slice(remaining, func(i, j int) bool {
				return declOrder[remaining[i]] < declOrder[remaining[j]]
			})
			result = append(result, remaining...)
			break
		}

		sort.Slice(ready, func(i, j int) bool {
			return declOrder[ready[i]] < declOrder[ready[j]]
		})

		for _, name := range ready {
			result = append(result, name)
			processed[name] = true
			for _, dep := range dependents[name] {
				if !processed[dep] {
					inDegree[dep]--
				}
			}
		}
	}

	return result
}
