package workunit

// DisplayNameResolver computes shortest unique names within a set.
// When multiple units have the same display key, it includes the module
// for disambiguation. When a display key is unique, just the key is returned.
// Uses DisplayKey() for uniqueness checking which considers context:
// - For BDD tests: spec name
// - For unit tests: testname (or component if no testname)
// - For other contexts: component name
type DisplayNameResolver struct {
	keyCounts map[string]int // DisplayKey() -> count of units with this key
}

// NewDisplayNameResolver creates a resolver for the given set of units.
// It counts how many times each display key appears to determine
// which names need disambiguation.
func NewDisplayNameResolver(units []UnitID) *DisplayNameResolver {
	r := &DisplayNameResolver{keyCounts: make(map[string]int)}
	for _, u := range units {
		r.keyCounts[u.DisplayKey()]++
	}
	return r
}

// Resolve returns shortest unique name for unit.
// For BDD tests, uses spec name (e.g., "build-module" or "eac-cli:build-module").
// If the component name is unique in the set, returns just the component (e.g., "go").
// If the component name is ambiguous, returns module:component (e.g., "core:go").
//
// Note: This method uses the legacy display format. For context-aware display,
// use ResolveShort() instead.
func (r *DisplayNameResolver) Resolve(u UnitID) string {
	key := u.DisplayKey()
	if r.keyCounts[key] == 1 {
		return key // Unique: just "go" or "build-module"
	}
	return u.Module + ":" + key // Ambiguous: "core:go" or "eac-cli:build-module"
}

// ResolveTabLabel returns tab-constrained name.
// First resolves the shortest unique name, then truncates if needed.
func (r *DisplayNameResolver) ResolveTabLabel(u UnitID, maxWidth int) string {
	name := r.Resolve(u)
	if len(name) <= maxWidth {
		return name
	}
	if maxWidth > 3 {
		return name[:maxWidth-3] + "..."
	}
	return name[:maxWidth]
}

// NeedsDisambiguation returns true if the display key appears more than once.
// Implements DisplayResolver interface.
func (r *DisplayNameResolver) NeedsDisambiguation(key string) bool {
	return r.keyCounts[key] > 1
}

// Count returns how many units have the given display key.
func (r *DisplayNameResolver) Count(key string) int {
	return r.keyCounts[key]
}
