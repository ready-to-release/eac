package workunit

// UoWAggregator tracks UoW counts per module for cache aggregation.
// A module is considered cached only if ALL its UoWs are cached.
type UoWAggregator struct {
	moduleUoWCounts    map[string]int
	moduleCachedCounts map[string]int
}

// NewUoWAggregator creates an aggregator from expected UoWs.
func NewUoWAggregator(expectedUoWs []UnitID) *UoWAggregator {
	agg := &UoWAggregator{
		moduleUoWCounts:    make(map[string]int),
		moduleCachedCounts: make(map[string]int),
	}
	for _, id := range expectedUoWs {
		agg.moduleUoWCounts[id.Module]++
	}
	return agg
}

// MarkCached marks a UoW as cached for its module.
func (a *UoWAggregator) MarkCached(id UnitID) {
	a.moduleCachedCounts[id.Module]++
}

// IsModuleCached returns true if all UoWs for the module are cached.
func (a *UoWAggregator) IsModuleCached(module string) bool {
	total := a.moduleUoWCounts[module]
	cached := a.moduleCachedCounts[module]
	return cached == total && cached > 0
}

// GetModuleLists returns changed and up-to-date module lists.
// The order is preserved from expectedUoWs (first occurrence of each module).
func (a *UoWAggregator) GetModuleLists(expectedUoWs []UnitID) (changed, upToDate []string) {
	seen := make(map[string]bool)
	for _, id := range expectedUoWs {
		if seen[id.Module] {
			continue
		}
		seen[id.Module] = true

		if a.IsModuleCached(id.Module) {
			upToDate = append(upToDate, id.Module)
		} else {
			changed = append(changed, id.Module)
		}
	}
	return
}

// Stats returns UoW counts for a module.
func (a *UoWAggregator) Stats(module string) (total, cached int) {
	return a.moduleUoWCounts[module], a.moduleCachedCounts[module]
}
