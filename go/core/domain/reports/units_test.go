package reports

import (
	"testing"
	"time"
)

func TestFilterUnits_NoFilters(t *testing.T) {
	units := []*UnitInfo{
		{Module: "mod1", Component: "go", Tool: "go"},
		{Module: "mod2", Component: "ts", Tool: "tsc"},
	}

	filters := UnitFilters{}
	result := FilterUnits(units, filters)

	if len(result) != 2 {
		t.Errorf("expected 2 units, got %d", len(result))
	}
}

func TestFilterUnits_ModuleFilter(t *testing.T) {
	units := []*UnitInfo{
		{Module: "mod1", Component: "go", Tool: "go"},
		{Module: "mod2", Component: "ts", Tool: "tsc"},
	}

	filters := UnitFilters{Module: "mod1"}
	result := FilterUnits(units, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 unit, got %d", len(result))
	}
	if result[0].Module != "mod1" {
		t.Errorf("expected mod1, got %s", result[0].Module)
	}
}

func TestFilterUnits_ComponentFilter(t *testing.T) {
	units := []*UnitInfo{
		{Module: "mod1", Component: "go", Tool: "go"},
		{Module: "mod1", Component: "ts", Tool: "tsc"},
	}

	filters := UnitFilters{Component: "go"}
	result := FilterUnits(units, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 unit, got %d", len(result))
	}
	if result[0].Component != "go" {
		t.Errorf("expected go component, got %s", result[0].Component)
	}
}

func TestFilterUnits_CachedFilter(t *testing.T) {
	units := []*UnitInfo{
		{
			Module:      "mod1",
			Component:   "go",
			CacheStatus: &CacheStatus{State: "up_to_date"},
		},
		{
			Module:      "mod2",
			Component:   "ts",
			CacheStatus: &CacheStatus{State: "stale"},
		},
	}

	filters := UnitFilters{Cached: true}
	result := FilterUnits(units, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 unit, got %d", len(result))
	}
	if result[0].Module != "mod1" {
		t.Errorf("expected mod1, got %s", result[0].Module)
	}
}

func TestFilterUnits_StaleFilter(t *testing.T) {
	units := []*UnitInfo{
		{
			Module:      "mod1",
			Component:   "go",
			CacheStatus: &CacheStatus{State: "up_to_date"},
		},
		{
			Module:      "mod2",
			Component:   "ts",
			CacheStatus: &CacheStatus{State: "stale", Reason: "source changed"},
		},
	}

	filters := UnitFilters{Stale: true}
	result := FilterUnits(units, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 unit, got %d", len(result))
	}
	if result[0].Module != "mod2" {
		t.Errorf("expected mod2, got %s", result[0].Module)
	}
}

func TestFilterUnits_ContainerFilter(t *testing.T) {
	units := []*UnitInfo{
		{Module: "mod1", Component: "go", Container: false, HostInstalled: true},
		{Module: "mod2", Component: "docker", Container: true, HostInstalled: false},
	}

	filters := UnitFilters{Container: true}
	result := FilterUnits(units, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 unit, got %d", len(result))
	}
	if result[0].Component != "docker" {
		t.Errorf("expected docker component, got %s", result[0].Component)
	}
}

func TestFilterUnits_HostFilter(t *testing.T) {
	units := []*UnitInfo{
		{Module: "mod1", Component: "go", Container: false, HostInstalled: true},
		{Module: "mod2", Component: "docker", Container: true, HostInstalled: false},
	}

	filters := UnitFilters{Host: true}
	result := FilterUnits(units, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 unit, got %d", len(result))
	}
	if result[0].Component != "go" {
		t.Errorf("expected go component, got %s", result[0].Component)
	}
}

func TestUnitFilters_HasFilters(t *testing.T) {
	tests := []struct {
		name     string
		filters  UnitFilters
		expected bool
	}{
		{"empty", UnitFilters{}, false},
		{"module only", UnitFilters{Module: "mod1"}, true},
		{"component only", UnitFilters{Component: "go"}, true},
		{"cached only", UnitFilters{Cached: true}, true},
		{"stale only", UnitFilters{Stale: true}, true},
		{"container only", UnitFilters{Container: true}, true},
		{"host only", UnitFilters{Host: true}, true},
		{"combined", UnitFilters{Module: "mod1", Cached: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filters.HasFilters(); got != tt.expected {
				t.Errorf("HasFilters() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGroupUnitsByModule(t *testing.T) {
	units := []*UnitInfo{
		{Module: "mod1", Component: "go"},
		{Module: "mod1", Component: "ts"},
		{Module: "mod2", Component: "go"},
	}

	result := GroupUnitsByModule(units)

	if len(result) != 2 {
		t.Errorf("expected 2 modules, got %d", len(result))
	}
	if len(result["mod1"]) != 2 {
		t.Errorf("expected 2 units in mod1, got %d", len(result["mod1"]))
	}
	if len(result["mod2"]) != 1 {
		t.Errorf("expected 1 unit in mod2, got %d", len(result["mod2"]))
	}
}

func TestGetStaleUnits(t *testing.T) {
	units := []*UnitInfo{
		{Module: "mod1", CacheStatus: &CacheStatus{State: "up_to_date"}},
		{Module: "mod2", CacheStatus: &CacheStatus{State: "stale", Reason: "source changed"}},
		{Module: "mod3", CacheStatus: &CacheStatus{State: "new"}},
		{Module: "mod4", CacheStatus: &CacheStatus{State: "stale", Reason: "previous failure"}},
	}

	result := GetStaleUnits(units)

	if len(result) != 2 {
		t.Errorf("expected 2 stale units, got %d", len(result))
	}
}

func TestBuildUnitSummary(t *testing.T) {
	units := []*UnitInfo{
		{
			Module:        "mod1",
			Tool:          "go",
			Container:     false,
			HostInstalled: true,
			CacheStatus:   &CacheStatus{State: "up_to_date"},
		},
		{
			Module:        "mod1",
			Tool:          "golangci-lint",
			Container:     false,
			HostInstalled: true,
			CacheStatus:   &CacheStatus{State: "stale"},
		},
		{
			Module:        "mod2",
			Tool:          "trivy",
			Container:     true,
			HostInstalled: false,
			CacheStatus:   &CacheStatus{State: "new"},
		},
	}

	summary := buildUnitSummary(units)

	if summary.Total != 3 {
		t.Errorf("Total: expected 3, got %d", summary.Total)
	}
	if summary.Cached != 1 {
		t.Errorf("Cached: expected 1, got %d", summary.Cached)
	}
	if summary.Stale != 1 {
		t.Errorf("Stale: expected 1, got %d", summary.Stale)
	}
	if summary.New != 1 {
		t.Errorf("New: expected 1, got %d", summary.New)
	}
	if summary.Container != 1 {
		t.Errorf("Container: expected 1, got %d", summary.Container)
	}
	if summary.HostInstalled != 2 {
		t.Errorf("HostInstalled: expected 2, got %d", summary.HostInstalled)
	}
	if summary.ByModule["mod1"] != 2 {
		t.Errorf("ByModule[mod1]: expected 2, got %d", summary.ByModule["mod1"])
	}
	if summary.ByTool["go"] != 1 {
		t.Errorf("ByTool[go]: expected 1, got %d", summary.ByTool["go"])
	}
}

func TestCacheStatus_Fields(t *testing.T) {
	now := time.Now()
	status := &CacheStatus{
		State:      "stale",
		SourceHash: "abc123",
		ExecutedAt: now,
		Reason:     "source changed",
	}

	if status.State != "stale" {
		t.Errorf("State: expected stale, got %s", status.State)
	}
	if status.SourceHash != "abc123" {
		t.Errorf("SourceHash: expected abc123, got %s", status.SourceHash)
	}
	if !status.ExecutedAt.Equal(now) {
		t.Errorf("ExecutedAt: expected %v, got %v", now, status.ExecutedAt)
	}
	if status.Reason != "source changed" {
		t.Errorf("Reason: expected 'source changed', got %s", status.Reason)
	}
}
