// repository_groups.go contains module group expansion, dependency derivation,
// and component group expansion logic for RepositoryConfig.
package config

import (
	"fmt"
	"slices"
	"sort"
)

// RootDependency is a reserved sentinel value for depends_on.
// Modules that declare depends_on: [root] are root-level baseline tooling.
// When "root" is in depends_on it must be the only entry; it is stripped during
// loading so the module ends up with no actual dependencies (Layer 0).
const RootDependency = "root"

// applyModuleGroupDefaults sets ModuleGroup to the module's Moniker for any module
// that doesn't already have an explicit ModuleGroup. This mirrors applyComponentGroupDefaults
// and ensures every module belongs to a group (at minimum, a group-of-one named after itself).
func (c *RepositoryConfig) applyModuleGroupDefaults() {
	for i := range c.Modules {
		if c.Modules[i].ModuleGroup == "" {
			c.Modules[i].ModuleGroup = c.Modules[i].Moniker
		}
	}
}

// expandModuleGroups resolves group names in depends_on to individual module monikers.
// Called after all modules are loaded but before template expansion.
//
// Resolution rules:
//  1. The reserved sentinel "root" is validated and stripped first
//  2. depends_on entries are resolved in order: exact moniker match first, then group match
//  3. A group name MUST NOT collide with any module moniker
//  4. Group expansion replaces group names with individual module monikers
//  5. Self-references are excluded (a module won't depend on itself via group expansion)
//  6. Duplicates are removed while preserving order
func (c *RepositoryConfig) expandModuleGroups() error {
	// Validate and strip the "root" sentinel before group expansion
	if err := c.validateAndStripRoot(); err != nil {
		return err
	}

	// Build moniker set for collision detection
	monikers := make(map[string]bool, len(c.Modules))
	for _, m := range c.Modules {
		monikers[m.Moniker] = true
	}

	// Build group -> []moniker index (preserving declaration order)
	// Skip self-named groups (where module_group == moniker) — these are
	// synthetic defaults that represent a group-of-one and should not
	// participate in collision detection or group expansion.
	groups := make(map[string][]string)
	for _, m := range c.Modules {
		if m.ModuleGroup != "" && m.ModuleGroup != m.Moniker {
			groups[m.ModuleGroup] = append(groups[m.ModuleGroup], m.Moniker)
		}
	}

	// Validate: group names must not collide with module monikers
	for groupName := range groups {
		if monikers[groupName] {
			return fmt.Errorf("module_group %q collides with module moniker %q", groupName, groupName)
		}
	}

	// Expand depends_on for each module
	for i := range c.Modules {
		if len(c.Modules[i].DependsOn) == 0 {
			continue
		}

		var expanded []string
		seen := make(map[string]bool)
		for _, dep := range c.Modules[i].DependsOn {
			if members, isGroup := groups[dep]; isGroup {
				for _, m := range members {
					// Skip self-references
					if m == c.Modules[i].Moniker {
						continue
					}
					if !seen[m] {
						expanded = append(expanded, m)
						seen[m] = true
					}
				}
			} else if !seen[dep] {
				expanded = append(expanded, dep)
				seen[dep] = true
			}
		}
		c.Modules[i].DependsOn = expanded
	}

	return nil
}

// validateAndStripRoot validates the "root" sentinel in depends_on.
// Rules:
//   - "root" must be the only entry if present (cannot mix with real dependencies)
//   - "root" is not a real module; it is stripped so the module has no dependencies
//   - No module may use "root" as its moniker
func (c *RepositoryConfig) validateAndStripRoot() error {
	// Validate: "root" must not be used as a module moniker
	for _, m := range c.Modules {
		if m.Moniker == RootDependency {
			return fmt.Errorf("module moniker %q is reserved; use it in depends_on to mark root-level modules", RootDependency)
		}
	}

	c.baselineModules = make(map[string]bool)

	for i := range c.Modules {
		deps := c.Modules[i].DependsOn
		if len(deps) == 0 {
			continue
		}

		if !slices.Contains(deps, RootDependency) {
			continue
		}

		// "root" found — it must be the only entry
		if len(deps) > 1 {
			return fmt.Errorf(
				"module %q has depends_on: %v; when %q is specified it must be the only entry",
				c.Modules[i].Moniker, deps, RootDependency,
			)
		}

		// Record as baseline before stripping
		c.baselineModules[c.Modules[i].Moniker] = true

		// Strip "root" — module becomes dependency-free (Layer 0)
		c.Modules[i].DependsOn = []string{}
	}

	return nil
}

// deriveModuleDepsFromComponentDeps iterates all modules and components,
// parses each component_deps entry, and unions the referenced module monikers
// into the module's DependsOn list. This enables component-level inter-module
// dependencies to automatically participate in module-level scheduling.
//
// Validation:
//   - Each entry must parse as 2-4 colon-separated parts
//   - Self-references (dep module == own module) are rejected
//   - A dep module must not already appear in explicit depends_on (overlap rule)
func (c *RepositoryConfig) deriveModuleDepsFromComponentDeps() error {
	for i := range c.Modules {
		m := &c.Modules[i]

		// Build set of existing explicit depends_on for overlap detection
		explicitDeps := make(map[string]bool, len(m.DependsOn))
		for _, dep := range m.DependsOn {
			explicitDeps[dep] = true
		}

		// Collect derived module monikers from component_deps
		derived := make(map[string]bool)
		for compName, entry := range m.Components {
			if entry == nil {
				continue
			}
			for _, dep := range entry.ComponentDeps {
				parsed, err := ParseComponentDep(dep)
				if err != nil {
					return fmt.Errorf("module %q component %q: %w", m.Moniker, compName, err)
				}
				if parsed.Module == m.Moniker {
					return fmt.Errorf("module %q component %q: component_deps entry %q is a self-reference", m.Moniker, compName, dep)
				}
				if explicitDeps[parsed.Module] {
					return fmt.Errorf("module %q component %q: component_deps entry %q targets module %q which is already in depends_on; use one or the other, not both",
						m.Moniker, compName, dep, parsed.Module)
				}
				derived[parsed.Module] = true
			}
		}

		// Union derived monikers into DependsOn
		for depModule := range derived {
			m.DependsOn = append(m.DependsOn, depModule)
		}
	}
	return nil
}

// expandAllComponentGroups expands component groups for all modules.
func (c *RepositoryConfig) expandAllComponentGroups() error {
	for i := range c.Modules {
		c.Modules[i].Components.applyComponentGroupDefaults()
		if err := c.Modules[i].expandComponentGroups(); err != nil {
			return err
		}
	}
	return nil
}

// expandComponentGroups resolves group names in component depends_on to individual component names.
// Mirrors the module_group expansion logic but scoped to a single module.
//
// Resolution rules:
//  1. depends_on entries are resolved in order: exact component name match first, then group match
//  2. A group name MUST NOT collide with any component name
//  3. Group expansion replaces group names with individual component names
//  4. Self-references are excluded (a component won't depend on itself via group expansion)
//  5. Duplicates are removed while preserving order
func (m *Module) expandComponentGroups() error {
	if m.Components == nil {
		return nil
	}

	compNames := make(map[string]bool, len(m.Components))
	for name := range m.Components {
		compNames[name] = true
	}

	// Build group -> []component name index
	// Skip self-named groups (component_group == component name) — synthetic defaults
	groups := make(map[string][]string)
	for name, entry := range m.Components {
		if entry == nil {
			continue
		}
		if entry.ComponentGroup != "" && entry.ComponentGroup != name {
			groups[entry.ComponentGroup] = append(groups[entry.ComponentGroup], name)
		}
	}

	// Sort group members for deterministic expansion order
	for groupName := range groups {
		sort.Strings(groups[groupName])
	}

	// Validate: group names must not collide with component names
	for groupName := range groups {
		if compNames[groupName] {
			return fmt.Errorf("module %q: component_group %q collides with component name %q",
				m.Moniker, groupName, groupName)
		}
	}

	// Expand depends_on for each component
	for compName, entry := range m.Components {
		if entry == nil || len(entry.DependsOn) == 0 {
			continue
		}
		var expanded []string
		seen := make(map[string]bool)
		for _, dep := range entry.DependsOn {
			if members, isGroup := groups[dep]; isGroup {
				for _, member := range members {
					if member == compName {
						continue
					}
					if !seen[member] {
						expanded = append(expanded, member)
						seen[member] = true
					}
				}
			} else if !seen[dep] {
				expanded = append(expanded, dep)
				seen[dep] = true
			}
		}
		entry.DependsOn = expanded
	}
	return nil
}
