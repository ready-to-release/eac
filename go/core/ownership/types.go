// Package ownership provides the simplified file ownership model.
// Ownership is determined by directory roots (most-specific-root wins),
// not by glob patterns. When multiple components share the same root,
// file extensions from component-kinds determine which component owns which file.
package ownership

// Owner identifies the module and component that owns a file.
type Owner struct {
	Module    string // Module moniker
	Component string // Component name within the module
	Root      string // The root that matched
}

// FileOwnership tracks a file and its resolved owner.
type FileOwnership struct {
	FilePath string // Relative path from repo root (forward slashes)
	Owner    *Owner // Resolved owner (nil if unowned)
}

// ModuleDefinition is the minimal data needed for ownership resolution.
type ModuleDefinition struct {
	Moniker    string
	Components map[string]*ComponentOwnership
}

// ComponentOwnership is the ownership-relevant subset of a component entry.
type ComponentOwnership struct {
	Root          string   // Directory this component owns
	File          string   // Single file this component owns (mutually exclusive with Root)
	ComponentType string   // Component type name (for extension-based tie-breaking)
	Extensions    []string // File extensions from component type (for same-root tie-breaking)
}

// ValidationError describes a file ownership violation.
type ValidationError struct {
	FilePath string  // File with the issue
	Issue    string  // "unowned" or "multi-owned"
	Owners   []Owner // For multi-owned files: the conflicting owners
}
