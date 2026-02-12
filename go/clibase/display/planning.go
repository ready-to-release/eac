package display

import "time"

// PlannedTool represents a tool that will be used during execution.
type PlannedTool struct {
	Name        string
	IsContainer bool
}

// PlannedWorkItem describes predicted work from module/component discovery.
type PlannedWorkItem struct {
	ID            string
	DisplayName   string
	Weight        int
	Module        string
	Component     string
	ComponentType string
}

// UoWEnrichment carries incremental enrichment data for an existing planned tab.
type UoWEnrichment struct {
	ID          string
	Tool        string
	Container   bool
	Weight      int
	CacheStatus CacheHit
	CacheTime   time.Time
	DependsOn   []string
}

// CacheHit describes the cache status of a work unit.
type CacheHit int

const (
	CacheUnknown  CacheHit = iota
	CacheMiss
	CacheHitFresh
)
