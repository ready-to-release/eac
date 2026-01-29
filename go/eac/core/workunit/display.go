package workunit

// DisplayNameResolver computes shortest unique names within a set.
// When multiple units have the same component name, it includes the module
// for disambiguation. When a component name is unique, just the component is returned.
type DisplayNameResolver struct {
	componentCounts map[string]int // component -> count of units with this component
}

// NewDisplayNameResolver creates a resolver for the given set of units.
// It counts how many times each component name appears to determine
// which names need disambiguation.
func NewDisplayNameResolver(units []UnitID) *DisplayNameResolver {
	r := &DisplayNameResolver{componentCounts: make(map[string]int)}
	for _, u := range units {
		r.componentCounts[u.Component]++
	}
	return r
}

// Resolve returns shortest unique name for unit.
// If the component name is unique in the set, returns just the component (e.g., "go").
// If the component name is ambiguous, returns module:component (e.g., "eac-core:go").
func (r *DisplayNameResolver) Resolve(u UnitID) string {
	if r.componentCounts[u.Component] == 1 {
		return u.Component // Unique: just "go"
	}
	return u.Module + ":" + u.Component // Ambiguous: "eac-core:go"
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

// NeedsDisambiguation returns true if the component name appears more than once.
func (r *DisplayNameResolver) NeedsDisambiguation(component string) bool {
	return r.componentCounts[component] > 1
}

// Count returns how many units have the given component name.
func (r *DisplayNameResolver) Count(component string) int {
	return r.componentCounts[component]
}
