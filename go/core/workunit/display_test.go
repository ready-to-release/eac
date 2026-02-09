package workunit

import "testing"

func TestDisplayNameResolver_Resolve(t *testing.T) {
	tests := []struct {
		name  string
		units []UnitID
		query UnitID
		want  string
	}{
		{
			name: "unique component returns just component",
			units: []UnitID{
				{Module: "core", ComponentType: "go", ComponentName: "go"},
				{Module: "eac", ComponentType: "impl", ComponentName: "impl"},
			},
			query: UnitID{Module: "core", ComponentType: "go", ComponentName: "go"},
			want:  "go",
		},
		{
			name: "duplicate component returns module:component",
			units: []UnitID{
				{Module: "core", ComponentType: "go", ComponentName: "go"},
				{Module: "eac", ComponentType: "go", ComponentName: "go"},
			},
			query: UnitID{Module: "core", ComponentType: "go", ComponentName: "go"},
			want:  "core:go",
		},
		{
			name: "three modules same component all disambiguated",
			units: []UnitID{
				{Module: "mod-a", ComponentType: "lib", ComponentName: "lib"},
				{Module: "mod-b", ComponentType: "lib", ComponentName: "lib"},
				{Module: "mod-c", ComponentType: "lib", ComponentName: "lib"},
			},
			query: UnitID{Module: "mod-b", ComponentType: "lib", ComponentName: "lib"},
			want:  "mod-b:lib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDisplayNameResolver(tt.units)
			got := resolver.Resolve(tt.query)
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDisplayNameResolver_ResolveTabLabel(t *testing.T) {
	tests := []struct {
		name     string
		units    []UnitID
		query    UnitID
		maxWidth int
		want     string
	}{
		{
			name: "short unique component fits",
			units: []UnitID{
				{Module: "core", ComponentType: "go", ComponentName: "go"},
			},
			query:    UnitID{Module: "core", ComponentType: "go", ComponentName: "go"},
			maxWidth: 10,
			want:     "go",
		},
		{
			name: "long disambiguated name truncated",
			units: []UnitID{
				{Module: "core", ComponentType: "component", ComponentName: "component"},
				{Module: "eac", ComponentType: "component", ComponentName: "component"},
			},
			query:    UnitID{Module: "core", ComponentType: "component", ComponentName: "component"},
			maxWidth: 10,
			want:     "core:co...",
		},
		{
			name: "very short maxWidth",
			units: []UnitID{
				{Module: "core", ComponentType: "go", ComponentName: "go"},
			},
			query:    UnitID{Module: "core", ComponentType: "go", ComponentName: "go"},
			maxWidth: 2,
			want:     "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDisplayNameResolver(tt.units)
			got := resolver.ResolveTabLabel(tt.query, tt.maxWidth)
			if got != tt.want {
				t.Errorf("ResolveTabLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDisplayNameResolver_NeedsDisambiguation(t *testing.T) {
	units := []UnitID{
		{Module: "core", ComponentType: "go", ComponentName: "go"},
		{Module: "eac", ComponentType: "go", ComponentName: "go"},
		{Module: "eac-specs", ComponentType: "unique", ComponentName: "unique"},
	}
	resolver := NewDisplayNameResolver(units)

	if !resolver.NeedsDisambiguation("go") {
		t.Error("go should need disambiguation")
	}
	if resolver.NeedsDisambiguation("unique") {
		t.Error("unique should not need disambiguation")
	}
	if resolver.NeedsDisambiguation("nonexistent") {
		t.Error("nonexistent should not need disambiguation")
	}
}

func TestDisplayNameResolver_Count(t *testing.T) {
	units := []UnitID{
		{Module: "a", ComponentType: "go", ComponentName: "go"},
		{Module: "b", ComponentType: "go", ComponentName: "go"},
		{Module: "c", ComponentType: "go", ComponentName: "go"},
		{Module: "d", ComponentType: "unique", ComponentName: "unique"},
	}
	resolver := NewDisplayNameResolver(units)

	if got := resolver.Count("go"); got != 3 {
		t.Errorf("Count(go) = %d, want 3", got)
	}
	if got := resolver.Count("unique"); got != 1 {
		t.Errorf("Count(unique) = %d, want 1", got)
	}
	if got := resolver.Count("missing"); got != 0 {
		t.Errorf("Count(missing) = %d, want 0", got)
	}
}

// =============================================================================
// Spec Field Tests (BDD Tests)
// =============================================================================

func TestDisplayNameResolver_Resolve_WithSpec(t *testing.T) {
	tests := []struct {
		name  string
		units []UnitID
		query UnitID
		want  string
	}{
		{
			name: "unique spec returns just spec name",
			units: []UnitID{
				{Action: ActionTest, Module: "eac", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"},
				{Action: ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go"},
			},
			query: UnitID{Action: ActionTest, Module: "eac", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"},
			want:  "build-module",
		},
		{
			name: "duplicate spec returns module:specname",
			units: []UnitID{
				{Action: ActionTest, Module: "eac", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"},
				{Action: ActionTest, Module: "core", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"},
			},
			query: UnitID{Action: ActionTest, Module: "eac", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"},
			want:  "eac:build-module",
		},
		{
			name: "mixed spec and non-spec units",
			units: []UnitID{
				{Action: ActionTest, Module: "eac", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"},
				{Action: ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go"},
				{Action: ActionTest, Module: "core", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "cache-invalidation"},
			},
			query: UnitID{Action: ActionTest, Module: "core", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "cache-invalidation"},
			want:  "cache-invalidation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := NewDisplayNameResolver(tt.units)
			got := resolver.Resolve(tt.query)
			if got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDisplayNameResolver_NeedsDisambiguation_WithSpec(t *testing.T) {
	units := []UnitID{
		{Action: ActionTest, Module: "eac", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"},
		{Action: ActionTest, Module: "core", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "build-module"}, // Duplicate spec name
		{Action: ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go"},
		{Action: ActionTest, Module: "eac-specs", ComponentType: "gherkin", ComponentName: "gherkin", Tool: "godog", Spec: "unique-spec"},
	}
	resolver := NewDisplayNameResolver(units)

	if !resolver.NeedsDisambiguation("build-module") {
		t.Error("build-module should need disambiguation (appears in 2 modules)")
	}
	if resolver.NeedsDisambiguation("unique-spec") {
		t.Error("unique-spec should not need disambiguation")
	}
	if resolver.NeedsDisambiguation("go") {
		t.Error("go should not need disambiguation (only 1 unit)")
	}
}
