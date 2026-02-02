package repository

import (
	"testing"
)

func TestCalculateGraphStats(t *testing.T) {
	tests := []struct {
		name         string
		monikers     []string
		dependencies map[string][]string
		dependents   map[string][]string
		wantStats    DependencyGraphStats
	}{
		{
			name:         "empty graph",
			monikers:     []string{},
			dependencies: map[string][]string{},
			dependents:   map[string][]string{},
			wantStats: DependencyGraphStats{
				TotalModules:      0,
				TotalDependencies: 0,
				RootModules:       0,
				LeafModules:       0,
				MaxDependencies:   0,
				MaxDependents:     0,
			},
		},
		{
			name:     "single module - no deps",
			monikers: []string{"mod-a"},
			dependencies: map[string][]string{
				"mod-a": {},
			},
			dependents: map[string][]string{
				"mod-a": {},
			},
			wantStats: DependencyGraphStats{
				TotalModules:      1,
				TotalDependencies: 0,
				RootModules:       1,
				LeafModules:       1,
				MaxDependencies:   0,
				MaxDependents:     0,
			},
		},
		{
			name:     "linear chain A->B->C",
			monikers: []string{"mod-a", "mod-b", "mod-c"},
			dependencies: map[string][]string{
				"mod-a": {"mod-b"},
				"mod-b": {"mod-c"},
				"mod-c": {},
			},
			dependents: map[string][]string{
				"mod-a": {},
				"mod-b": {"mod-a"},
				"mod-c": {"mod-b"},
			},
			wantStats: DependencyGraphStats{
				TotalModules:      3,
				TotalDependencies: 2,
				RootModules:       1, // mod-c has no deps
				LeafModules:       1, // mod-a has no dependents
				MaxDependencies:   1,
				MaxDependents:     1,
			},
		},
		{
			name:     "diamond: A->B,C->D",
			monikers: []string{"mod-a", "mod-b", "mod-c", "mod-d"},
			dependencies: map[string][]string{
				"mod-a": {"mod-b", "mod-c"},
				"mod-b": {"mod-d"},
				"mod-c": {"mod-d"},
				"mod-d": {},
			},
			dependents: map[string][]string{
				"mod-a": {},
				"mod-b": {"mod-a"},
				"mod-c": {"mod-a"},
				"mod-d": {"mod-b", "mod-c"},
			},
			wantStats: DependencyGraphStats{
				TotalModules:      4,
				TotalDependencies: 4,
				RootModules:       1, // mod-d
				LeafModules:       1, // mod-a
				MaxDependencies:   2, // mod-a
				MaxDependents:     2, // mod-d
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateGraphStats(tt.monikers, tt.dependencies, tt.dependents)

			if got.TotalModules != tt.wantStats.TotalModules {
				t.Errorf("TotalModules = %d, want %d", got.TotalModules, tt.wantStats.TotalModules)
			}
			if got.TotalDependencies != tt.wantStats.TotalDependencies {
				t.Errorf("TotalDependencies = %d, want %d", got.TotalDependencies, tt.wantStats.TotalDependencies)
			}
			if got.RootModules != tt.wantStats.RootModules {
				t.Errorf("RootModules = %d, want %d", got.RootModules, tt.wantStats.RootModules)
			}
			if got.LeafModules != tt.wantStats.LeafModules {
				t.Errorf("LeafModules = %d, want %d", got.LeafModules, tt.wantStats.LeafModules)
			}
			if got.MaxDependencies != tt.wantStats.MaxDependencies {
				t.Errorf("MaxDependencies = %d, want %d", got.MaxDependencies, tt.wantStats.MaxDependencies)
			}
			if got.MaxDependents != tt.wantStats.MaxDependents {
				t.Errorf("MaxDependents = %d, want %d", got.MaxDependents, tt.wantStats.MaxDependents)
			}
		})
	}
}

func TestSanitizeMermaidID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"mod-a", "mod_a"},
		{"core", "core"},
		{"r2r-cli", "r2r_cli"},
		{"test_module", "test_module"},
		{"Module123", "Module123"},
		{"a-b-c-d", "a_b_c_d"},
		{"foo.bar", "foo_bar"},
		{"foo/bar", "foo_bar"},
		{"mod@v1", "mod_v1"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeMermaidID(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeMermaidID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAddTransitiveDependents(t *testing.T) {
	// Graph: A->B->C, A->D
	// Dependents: B depends on nothing, A depends on B, etc.
	// Reverse: D is depended on by A, B is depended on by A
	dependentsGraph := map[string][]string{
		"mod-a": {},        // nothing depends on A
		"mod-b": {"mod-a"}, // A depends on B
		"mod-c": {"mod-b"}, // B depends on C
		"mod-d": {"mod-a"}, // A depends on D
	}

	tests := []struct {
		name      string
		module    string
		graph     map[string][]string
		initial   map[string]bool
		wantAdded []string
	}{
		{
			name:      "leaf module - no dependents",
			module:    "mod-a",
			graph:     dependentsGraph,
			initial:   map[string]bool{},
			wantAdded: []string{}, // A has no dependents
		},
		{
			name:      "module with one dependent",
			module:    "mod-d",
			graph:     dependentsGraph,
			initial:   map[string]bool{},
			wantAdded: []string{"mod-a"},
		},
		{
			name:      "transitive dependents",
			module:    "mod-c",
			graph:     dependentsGraph,
			initial:   map[string]bool{},
			wantAdded: []string{"mod-b", "mod-a"}, // C->B->A
		},
		{
			name:      "already in result - skips visited",
			module:    "mod-c",
			graph:     dependentsGraph,
			initial:   map[string]bool{"mod-b": true},
			wantAdded: []string{}, // mod-b already present, doesn't recurse through it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make(map[string]bool)
			for k, v := range tt.initial {
				result[k] = v
			}

			addTransitiveDependents(tt.module, tt.graph, result)

			// Check all expected modules were added
			for _, expected := range tt.wantAdded {
				if !result[expected] {
					t.Errorf("expected %s to be in result", expected)
				}
			}

			// Count new additions (excluding initial)
			newCount := 0
			for k := range result {
				if !tt.initial[k] {
					newCount++
				}
			}
			if newCount != len(tt.wantAdded) {
				t.Errorf("added %d modules, want %d", newCount, len(tt.wantAdded))
			}
		})
	}
}

func TestModuleDependency_Struct(t *testing.T) {
	dep := ModuleDependency{
		From: "r2r-cli",
		To:   "core",
	}

	if dep.From != "r2r-cli" {
		t.Errorf("From = %q, want %q", dep.From, "r2r-cli")
	}
	if dep.To != "core" {
		t.Errorf("To = %q, want %q", dep.To, "core")
	}
}

func TestExecutionPlan_Struct(t *testing.T) {
	plan := ExecutionPlan{
		Layers: [][]string{
			{"mod-c", "mod-d"},
			{"mod-b"},
			{"mod-a"},
		},
		ExecutionOrder: []string{"mod-c", "mod-d", "mod-b", "mod-a"},
		LayerCount:     3,
	}

	if len(plan.Layers) != 3 {
		t.Errorf("len(Layers) = %d, want 3", len(plan.Layers))
	}
	if plan.LayerCount != 3 {
		t.Errorf("LayerCount = %d, want 3", plan.LayerCount)
	}
	if len(plan.ExecutionOrder) != 4 {
		t.Errorf("len(ExecutionOrder) = %d, want 4", len(plan.ExecutionOrder))
	}
}

func TestDependencyGraphStats_Struct(t *testing.T) {
	stats := DependencyGraphStats{
		TotalModules:      10,
		TotalDependencies: 15,
		RootModules:       3,
		LeafModules:       2,
		MaxDependencies:   4,
		MaxDependents:     5,
	}

	if stats.TotalModules != 10 {
		t.Errorf("TotalModules = %d, want 10", stats.TotalModules)
	}
	if stats.TotalDependencies != 15 {
		t.Errorf("TotalDependencies = %d, want 15", stats.TotalDependencies)
	}
	if stats.RootModules != 3 {
		t.Errorf("RootModules = %d, want 3", stats.RootModules)
	}
	if stats.LeafModules != 2 {
		t.Errorf("LeafModules = %d, want 2", stats.LeafModules)
	}
	if stats.MaxDependencies != 4 {
		t.Errorf("MaxDependencies = %d, want 4", stats.MaxDependencies)
	}
	if stats.MaxDependents != 5 {
		t.Errorf("MaxDependents = %d, want 5", stats.MaxDependents)
	}
}
