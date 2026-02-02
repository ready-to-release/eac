package workunit

import "testing"

func TestDisplayNameResolver_Resolve(t *testing.T) {
	tests := []struct {
		name   string
		units  []UnitID
		query  UnitID
		want   string
	}{
		{
			name: "unique component returns just component",
			units: []UnitID{
				{Module: "core", Component: "go"},
				{Module: "eac-cli", Component: "impl"},
			},
			query: UnitID{Module: "core", Component: "go"},
			want:  "go",
		},
		{
			name: "duplicate component returns module:component",
			units: []UnitID{
				{Module: "core", Component: "go"},
				{Module: "eac-cli", Component: "go"},
			},
			query: UnitID{Module: "core", Component: "go"},
			want:  "core:go",
		},
		{
			name: "three modules same component all disambiguated",
			units: []UnitID{
				{Module: "mod-a", Component: "lib"},
				{Module: "mod-b", Component: "lib"},
				{Module: "mod-c", Component: "lib"},
			},
			query: UnitID{Module: "mod-b", Component: "lib"},
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
				{Module: "core", Component: "go"},
			},
			query:    UnitID{Module: "core", Component: "go"},
			maxWidth: 10,
			want:     "go",
		},
		{
			name: "long disambiguated name truncated",
			units: []UnitID{
				{Module: "core", Component: "component"},
				{Module: "eac-cli", Component: "component"},
			},
			query:    UnitID{Module: "core", Component: "component"},
			maxWidth: 10,
			want:     "core:co...",
		},
		{
			name: "very short maxWidth",
			units: []UnitID{
				{Module: "core", Component: "go"},
			},
			query:    UnitID{Module: "core", Component: "go"},
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
		{Module: "core", Component: "go"},
		{Module: "eac-cli", Component: "go"},
		{Module: "eac-specs", Component: "unique"},
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
		{Module: "a", Component: "go"},
		{Module: "b", Component: "go"},
		{Module: "c", Component: "go"},
		{Module: "d", Component: "unique"},
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
				{Module: "eac-cli", Component: "complex-path:godog", Spec: "build-module"},
				{Module: "core", Component: "go"},
			},
			query: UnitID{Module: "eac-cli", Component: "complex-path:godog", Spec: "build-module"},
			want:  "build-module",
		},
		{
			name: "duplicate spec returns module:specname",
			units: []UnitID{
				{Module: "eac-cli", Component: "complex-path1:godog", Spec: "build-module"},
				{Module: "core", Component: "complex-path2:godog", Spec: "build-module"},
			},
			query: UnitID{Module: "eac-cli", Component: "complex-path1:godog", Spec: "build-module"},
			want:  "eac-cli:build-module",
		},
		{
			name: "mixed spec and non-spec units",
			units: []UnitID{
				{Module: "eac-cli", Component: "complex-path:godog", Spec: "build-module"},
				{Module: "core", Component: "go"},
				{Module: "core", Component: "complex-path:godog", Spec: "cache-invalidation"},
			},
			query: UnitID{Module: "core", Component: "complex-path:godog", Spec: "cache-invalidation"},
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
		{Module: "eac-cli", Component: "path1:godog", Spec: "build-module"},
		{Module: "core", Component: "path2:godog", Spec: "build-module"}, // Duplicate spec name
		{Module: "core", Component: "go"},
		{Module: "eac-specs", Component: "path3:godog", Spec: "unique-spec"},
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
