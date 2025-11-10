package modules

import (
	"fmt"
	"strings"
)

// BuildParentChain builds the full parent chain from root to module.
// Returns a slice starting with "." (root) and ending with the module's moniker.
//
// Example:
//   claude-agents (parent: "claude") → [".", "claude", "claude-agents"]
//   repository (parent: ".") → [".", "repository"]
//
// Returns an error if:
//   - A circular parent reference is detected
//   - A parent module doesn't exist in the registry
func BuildParentChain(module *ModuleContract, registry *Registry) ([]string, error) {
	chain := []string{module.Moniker}
	visited := make(map[string]bool)
	visited[module.Moniker] = true

	current := module.Parent

	// Walk up the parent chain
	for current != "." {
		// Detect circular reference
		if visited[current] {
			return nil, fmt.Errorf("circular parent chain detected: %s → %s",
				strings.Join(chain, " → "), current)
		}

		visited[current] = true
		chain = append(chain, current)

		// Get parent module
		parent, exists := registry.Get(current)
		if !exists {
			return nil, fmt.Errorf("parent module '%s' not found in registry (referenced by '%s')",
				current, module.Moniker)
		}

		current = parent.Parent
	}

	// Add root
	chain = append(chain, ".")

	// Reverse to get root-first order: [".", ..., "module"]
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

// GetDepth returns the depth of a module in the parent hierarchy.
// Depth is the number of steps from root (.).
//
// Depth calculation:
//   - Root level (parent = ".") → depth 1
//   - One parent (e.g., claude) → depth 2
//   - Two parents (e.g., sub-module of sub-module) → depth 3
//
// Example:
//   repository (parent: ".") → depth 1
//   claude (parent: ".") → depth 1
//   claude-agents (parent: "claude") → depth 2
//
// Returns an error if the parent chain is invalid.
func GetDepth(module *ModuleContract, registry *Registry) (int, error) {
	chain, err := BuildParentChain(module, registry)
	if err != nil {
		return 0, err
	}

	// Depth = chain length - 1
	// [".", "module"] → depth 1
	// [".", "parent", "module"] → depth 2
	return len(chain) - 1, nil
}

// ValidateParentChain validates that a module's parent chain is valid.
// Returns an error if:
//   - A circular parent reference exists
//   - A parent module doesn't exist in the registry
//   - The chain doesn't terminate at "." (root)
//
// This should be called after all modules are loaded to ensure
// the registry is complete.
func ValidateParentChain(module *ModuleContract, registry *Registry) error {
	_, err := BuildParentChain(module, registry)
	return err
}
