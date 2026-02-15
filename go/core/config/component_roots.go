package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// expandComponentRoots expands component_roots entries into individual ComponentEntry
// entries in the module's Components map. Each root path becomes a separate component
// with a name derived from the kind's name_pattern, the entry's name_pattern override,
// or DefaultDeriveName as fallback.
//
// After expansion, ComponentRoots is cleared to avoid double-processing.
func expandComponentRoots(mod *Module, kinds map[string]*ComponentType) error {
	if len(mod.ComponentRoots) == 0 {
		return nil
	}

	if mod.Components == nil {
		mod.Components = make(ModuleComponents)
	}

	// Sort kind names for deterministic ordering
	kindNames := make([]string, 0, len(mod.ComponentRoots))
	for k := range mod.ComponentRoots {
		kindNames = append(kindNames, k)
	}
	sort.Strings(kindNames)

	for _, kindName := range kindNames {
		entry := mod.ComponentRoots[kindName]
		if entry == nil || len(entry.Roots) == 0 {
			continue
		}

		// Compile entry-level name_pattern once per entry (not per root)
		var entryRe *regexp.Regexp
		if entry.NamePattern != "" {
			re, err := compileNamePattern(entry.NamePattern)
			if err != nil {
				return fmt.Errorf("module %q, component_roots %q: %w",
					mod.Moniker, kindName, err)
			}
			entryRe = re
		}

		kind := kinds[kindName]
		for _, root := range entry.Roots {
			name := deriveName(entryRe, kind, root)

			// Strip moniker prefix (e.g. "adapters-godog" → "godog")
			name = strings.TrimPrefix(name, mod.Moniker+"-")

			// Check for name collision with existing components
			if _, exists := mod.Components[name]; exists {
				return fmt.Errorf("module %q: component_roots derived name %q (from root %q) "+
					"conflicts with existing component", mod.Moniker, name, root)
			}

			mod.Components[name] = &ComponentEntry{
				Type: kindName,
				Root: root,
			}
			mod.ComponentOrder = append(mod.ComponentOrder, name)
		}
	}

	// Clear after expansion
	mod.ComponentRoots = nil
	return nil
}

// deriveName derives a component name from a root path using the priority:
// 1. Pre-compiled entry-level regex override
// 2. Kind's compiled DeriveName (uses kind.namePatternRe)
// 3. DefaultDeriveName (replace "/" with "-")
func deriveName(entryRe *regexp.Regexp, kind *ComponentType, root string) string {
	if entryRe != nil {
		return deriveNameFromRe(entryRe, root)
	}
	if kind != nil {
		return kind.DeriveName(root)
	}
	return DefaultDeriveName(root)
}
