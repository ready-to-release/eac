package workunit

// DisplayNameResolver computes shortest unique names within a set.
// When multiple units have the same component name, it includes the module
// for disambiguation. When a component name is unique, just the component is returned.
// For BDD tests (with Spec set), uses the spec name for uniqueness checking.
type DisplayNameResolver struct {
	componentCounts map[string]int // component/spec -> count of units with this name
}

// NewDisplayNameResolver creates a resolver for the given set of units.
// It counts how many times each component/spec name appears to determine
// which names need disambiguation.
func NewDisplayNameResolver(units []UnitID) *DisplayNameResolver {
	r := &DisplayNameResolver{componentCounts: make(map[string]int)}
	for _, u := range units {
		key := u.Component
		if u.Spec != "" {
			key = u.Spec // Use spec name for BDD tests
		}
		r.componentCounts[key]++
	}
	return r
}

// Resolve returns shortest unique name for unit.
// For BDD tests, uses spec name (e.g., "build-module" or "eac-commands:build-module").
// If the component name is unique in the set, returns just the component (e.g., "go").
// If the component name is ambiguous, returns module:component (e.g., "eac-core:go").
func (r *DisplayNameResolver) Resolve(u UnitID) string {
	key := u.Component
	if u.Spec != "" {
		key = u.Spec
	}
	if r.componentCounts[key] == 1 {
		return key // Unique: just "go" or "build-module"
	}
	return u.Module + ":" + key // Ambiguous: "eac-core:go" or "eac-commands:build-module"
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

// NeedsDisambiguation returns true if the component/spec name appears more than once.
func (r *DisplayNameResolver) NeedsDisambiguation(component string) bool {
	return r.componentCounts[component] > 1
}

// Count returns how many units have the given component name.
func (r *DisplayNameResolver) Count(component string) int {
	return r.componentCounts[component]
}
