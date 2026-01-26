//go:build L1 && ov
// +build L1,ov

package orchestrator

import (
	"testing"
	"time"
)

// TestModuleStatus_String tests the String() method of ModuleStatus.
func TestModuleStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status ModuleStatus
		want   string
	}{
		{
			name:   "pending status",
			status: ModuleStatusPending,
			want:   "pending",
		},
		{
			name:   "running status",
			status: ModuleStatusRunning,
			want:   "running",
		},
		{
			name:   "success status",
			status: ModuleStatusSuccess,
			want:   "success",
		},
		{
			name:   "failed status",
			status: ModuleStatusFailed,
			want:   "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("ModuleStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestComponentResultSet_DeriveStatus tests the DeriveStatus() method.
func TestComponentResultSet_DeriveStatus(t *testing.T) {
	tests := []struct {
		name       string
		components []ComponentResult
		want       ModuleStatus
	}{
		{
			name:       "empty components returns pending",
			components: []ComponentResult{},
			want:       ModuleStatusPending,
		},
		{
			name:       "nil components returns pending",
			components: nil,
			want:       ModuleStatusPending,
		},
		{
			name: "all success returns success",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "book", ExitCode: 0},
				{Module: "mod1", Component: "typescript", ExitCode: 0},
			},
			want: ModuleStatusSuccess,
		},
		{
			name: "single success returns success",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 0},
			},
			want: ModuleStatusSuccess,
		},
		{
			name: "one failure returns failed",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "book", ExitCode: 1},
				{Module: "mod1", Component: "typescript", ExitCode: 0},
			},
			want: ModuleStatusFailed,
		},
		{
			name: "first component failed returns failed",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 1},
				{Module: "mod1", Component: "book", ExitCode: 0},
			},
			want: ModuleStatusFailed,
		},
		{
			name: "last component failed returns failed",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "book", ExitCode: 1},
			},
			want: ModuleStatusFailed,
		},
		{
			name: "all components failed returns failed",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 1},
				{Module: "mod1", Component: "book", ExitCode: 2},
				{Module: "mod1", Component: "typescript", ExitCode: 1},
			},
			want: ModuleStatusFailed,
		},
		{
			name: "negative exit code (pending/running) returns running",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "book", ExitCode: -1},
			},
			want: ModuleStatusRunning,
		},
		{
			name: "all negative exit codes returns running",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: -1},
				{Module: "mod1", Component: "book", ExitCode: -1},
			},
			want: ModuleStatusRunning,
		},
		{
			name: "mixed negative and positive failure - failure takes precedence",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: -1},
				{Module: "mod1", Component: "book", ExitCode: 1},
			},
			want: ModuleStatusFailed,
		},
		{
			name: "high exit code returns failed",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "book", ExitCode: 127},
			},
			want: ModuleStatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := ComponentResultSet{
				Module:     "test-module",
				Components: tt.components,
			}

			got := rs.DeriveStatus()
			if got != tt.want {
				t.Errorf("ComponentResultSet.DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComponentResultSet_GetSortedComponents tests the GetSortedComponents() method.
func TestComponentResultSet_GetSortedComponents(t *testing.T) {
	tests := []struct {
		name       string
		components []ComponentResult
		wantOrder  []string // expected component names in order
	}{
		{
			name:       "empty components returns empty slice",
			components: []ComponentResult{},
			wantOrder:  []string{},
		},
		{
			name:       "nil components returns empty slice",
			components: nil,
			wantOrder:  []string{},
		},
		{
			name: "single component returns single item",
			components: []ComponentResult{
				{Module: "mod1", Component: "go", ExitCode: 0},
			},
			wantOrder: []string{"go"},
		},
		{
			name: "already sorted components remain sorted",
			components: []ComponentResult{
				{Module: "mod1", Component: "book", ExitCode: 0},
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "typescript", ExitCode: 0},
			},
			wantOrder: []string{"book", "go", "typescript"},
		},
		{
			name: "unsorted components are sorted alphabetically",
			components: []ComponentResult{
				{Module: "mod1", Component: "typescript", ExitCode: 0},
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "book", ExitCode: 0},
			},
			wantOrder: []string{"book", "go", "typescript"},
		},
		{
			name: "reverse order is sorted correctly",
			components: []ComponentResult{
				{Module: "mod1", Component: "zeta", ExitCode: 0},
				{Module: "mod1", Component: "beta", ExitCode: 0},
				{Module: "mod1", Component: "alpha", ExitCode: 0},
			},
			wantOrder: []string{"alpha", "beta", "zeta"},
		},
		{
			name: "components with same prefix are sorted correctly",
			components: []ComponentResult{
				{Module: "mod1", Component: "go-test", ExitCode: 0},
				{Module: "mod1", Component: "go", ExitCode: 0},
				{Module: "mod1", Component: "go-build", ExitCode: 0},
			},
			wantOrder: []string{"go", "go-build", "go-test"},
		},
		{
			name: "case-sensitive sorting",
			components: []ComponentResult{
				{Module: "mod1", Component: "Book", ExitCode: 0},
				{Module: "mod1", Component: "alpha", ExitCode: 0},
				{Module: "mod1", Component: "Zebra", ExitCode: 0},
			},
			wantOrder: []string{"Book", "Zebra", "alpha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := ComponentResultSet{
				Module:     "test-module",
				Components: tt.components,
			}

			got := rs.GetSortedComponents()

			// Verify length
			if len(got) != len(tt.wantOrder) {
				t.Fatalf("GetSortedComponents() returned %d items, want %d", len(got), len(tt.wantOrder))
			}

			// Verify order
			for i, wantName := range tt.wantOrder {
				if got[i].Component != wantName {
					t.Errorf("GetSortedComponents()[%d].Component = %q, want %q", i, got[i].Component, wantName)
				}
			}
		})
	}
}

// TestComponentResultSet_GetSortedComponents_DoesNotModifyOriginal verifies that
// GetSortedComponents returns a new slice and does not modify the original.
func TestComponentResultSet_GetSortedComponents_DoesNotModifyOriginal(t *testing.T) {
	original := []ComponentResult{
		{Module: "mod1", Component: "zeta", ExitCode: 0},
		{Module: "mod1", Component: "alpha", ExitCode: 0},
		{Module: "mod1", Component: "beta", ExitCode: 0},
	}

	rs := ComponentResultSet{
		Module:     "test-module",
		Components: original,
	}

	// Call GetSortedComponents
	sorted := rs.GetSortedComponents()

	// Verify sorted is correct
	if sorted[0].Component != "alpha" {
		t.Errorf("sorted[0].Component = %q, want %q", sorted[0].Component, "alpha")
	}

	// Verify original was NOT modified
	if original[0].Component != "zeta" {
		t.Errorf("Original slice was modified: original[0].Component = %q, want %q", original[0].Component, "zeta")
	}

	// Verify rs.Components was NOT modified
	if rs.Components[0].Component != "zeta" {
		t.Errorf("ComponentResultSet.Components was modified: [0].Component = %q, want %q", rs.Components[0].Component, "zeta")
	}
}

// TestComponentResultSet_GetSortedComponents_PreservesAllFields verifies that
// GetSortedComponents preserves all fields of ComponentResult, not just Component name.
func TestComponentResultSet_GetSortedComponents_PreservesAllFields(t *testing.T) {
	components := []ComponentResult{
		{
			Module:    "mod1",
			Component: "zeta",
			ExitCode:  1,
			Duration:  5 * time.Second,
			Errors:    []string{"error1", "error2"},
			Warnings:  []string{"warning1"},
			LogPath:   "/path/to/zeta.log",
		},
		{
			Module:    "mod1",
			Component: "alpha",
			ExitCode:  0,
			Duration:  2 * time.Second,
			Errors:    nil,
			Warnings:  nil,
			LogPath:   "/path/to/alpha.log",
		},
	}

	rs := ComponentResultSet{
		Module:     "test-module",
		Components: components,
	}

	sorted := rs.GetSortedComponents()

	// alpha should be first after sorting
	alpha := sorted[0]
	if alpha.Component != "alpha" {
		t.Fatalf("Expected alpha first, got %q", alpha.Component)
	}
	if alpha.Module != "mod1" {
		t.Errorf("alpha.Module = %q, want %q", alpha.Module, "mod1")
	}
	if alpha.ExitCode != 0 {
		t.Errorf("alpha.ExitCode = %d, want %d", alpha.ExitCode, 0)
	}
	if alpha.Duration != 2*time.Second {
		t.Errorf("alpha.Duration = %v, want %v", alpha.Duration, 2*time.Second)
	}
	if alpha.LogPath != "/path/to/alpha.log" {
		t.Errorf("alpha.LogPath = %q, want %q", alpha.LogPath, "/path/to/alpha.log")
	}

	// zeta should be second after sorting
	zeta := sorted[1]
	if zeta.Component != "zeta" {
		t.Fatalf("Expected zeta second, got %q", zeta.Component)
	}
	if zeta.ExitCode != 1 {
		t.Errorf("zeta.ExitCode = %d, want %d", zeta.ExitCode, 1)
	}
	if len(zeta.Errors) != 2 {
		t.Errorf("zeta.Errors length = %d, want %d", len(zeta.Errors), 2)
	}
	if len(zeta.Warnings) != 1 {
		t.Errorf("zeta.Warnings length = %d, want %d", len(zeta.Warnings), 1)
	}
}

// TestComponentResultSet_Fields tests that ComponentResultSet has the expected fields.
func TestComponentResultSet_Fields(t *testing.T) {
	rs := ComponentResultSet{
		Module: "my-module",
		Components: []ComponentResult{
			{Module: "my-module", Component: "go", ExitCode: 0},
		},
		Status:   ModuleStatusSuccess,
		Duration: 10 * time.Second,
	}

	if rs.Module != "my-module" {
		t.Errorf("Module = %q, want %q", rs.Module, "my-module")
	}
	if len(rs.Components) != 1 {
		t.Errorf("Components length = %d, want %d", len(rs.Components), 1)
	}
	if rs.Status != ModuleStatusSuccess {
		t.Errorf("Status = %v, want %v", rs.Status, ModuleStatusSuccess)
	}
	if rs.Duration != 10*time.Second {
		t.Errorf("Duration = %v, want %v", rs.Duration, 10*time.Second)
	}
}

// TestModuleStatus_Values tests that the ModuleStatus constants have distinct values.
func TestModuleStatus_Values(t *testing.T) {
	statuses := []ModuleStatus{
		ModuleStatusPending,
		ModuleStatusRunning,
		ModuleStatusSuccess,
		ModuleStatusFailed,
	}

	// Check all statuses are distinct
	seen := make(map[ModuleStatus]string)
	for _, s := range statuses {
		name := s.String()
		if existing, exists := seen[s]; exists {
			t.Errorf("ModuleStatus %q has same value as %q", name, existing)
		}
		seen[s] = name
	}

	// Check we have exactly 4 distinct statuses
	if len(seen) != 4 {
		t.Errorf("Expected 4 distinct ModuleStatus values, got %d", len(seen))
	}
}

// TestAggregateToComponentResultSets tests the AggregateToComponentResultSets function.
func TestAggregateToComponentResultSets(t *testing.T) {
	tests := []struct {
		name           string
		results        []ComponentResult
		wantModules    []string   // expected module names in order
		wantComponents [][]string // expected component names per module in order
		wantStatuses   []ModuleStatus
		wantDurations  []time.Duration
	}{
		{
			name:           "empty input returns empty slice",
			results:        []ComponentResult{},
			wantModules:    []string{},
			wantComponents: [][]string{},
			wantStatuses:   []ModuleStatus{},
			wantDurations:  []time.Duration{},
		},
		{
			name:    "nil input returns empty slice",
			results: nil,
			wantModules:    []string{},
			wantComponents: [][]string{},
			wantStatuses:   []ModuleStatus{},
			wantDurations:  []time.Duration{},
		},
		{
			name: "single component creates single ComponentResultSet",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: 0, Duration: 5 * time.Second},
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"go"}},
			wantStatuses:   []ModuleStatus{ModuleStatusSuccess},
			wantDurations:  []time.Duration{5 * time.Second},
		},
		{
			name: "multiple components from same module grouped together",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: 0, Duration: 3 * time.Second},
				{Module: "mod-a", Component: "typescript", ExitCode: 0, Duration: 2 * time.Second},
				{Module: "mod-a", Component: "book", ExitCode: 0, Duration: 4 * time.Second},
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"book", "go", "typescript"}}, // sorted alphabetically
			wantStatuses:   []ModuleStatus{ModuleStatusSuccess},
			wantDurations:  []time.Duration{4 * time.Second}, // max duration
		},
		{
			name: "multiple modules sorted alphabetically",
			results: []ComponentResult{
				{Module: "zebra-mod", Component: "go", ExitCode: 0, Duration: 1 * time.Second},
				{Module: "alpha-mod", Component: "go", ExitCode: 0, Duration: 2 * time.Second},
				{Module: "beta-mod", Component: "go", ExitCode: 0, Duration: 3 * time.Second},
			},
			wantModules:    []string{"alpha-mod", "beta-mod", "zebra-mod"},
			wantComponents: [][]string{{"go"}, {"go"}, {"go"}},
			wantStatuses:   []ModuleStatus{ModuleStatusSuccess, ModuleStatusSuccess, ModuleStatusSuccess},
			wantDurations:  []time.Duration{2 * time.Second, 3 * time.Second, 1 * time.Second},
		},
		{
			name: "components within module sorted by name",
			results: []ComponentResult{
				{Module: "mod-a", Component: "zeta", ExitCode: 0, Duration: 1 * time.Second},
				{Module: "mod-a", Component: "alpha", ExitCode: 0, Duration: 2 * time.Second},
				{Module: "mod-a", Component: "beta", ExitCode: 0, Duration: 3 * time.Second},
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"alpha", "beta", "zeta"}},
			wantStatuses:   []ModuleStatus{ModuleStatusSuccess},
			wantDurations:  []time.Duration{3 * time.Second},
		},
		{
			name: "status derived correctly - all success",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: 0},
				{Module: "mod-a", Component: "book", ExitCode: 0},
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"book", "go"}},
			wantStatuses:   []ModuleStatus{ModuleStatusSuccess},
			wantDurations:  []time.Duration{0},
		},
		{
			name: "status derived correctly - one failure",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: 0},
				{Module: "mod-a", Component: "book", ExitCode: 1},
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"book", "go"}},
			wantStatuses:   []ModuleStatus{ModuleStatusFailed},
			wantDurations:  []time.Duration{0},
		},
		{
			name: "status derived correctly - mixed modules",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: 0},
				{Module: "mod-a", Component: "book", ExitCode: 0},
				{Module: "mod-b", Component: "go", ExitCode: 1},
				{Module: "mod-c", Component: "typescript", ExitCode: 0},
			},
			wantModules:    []string{"mod-a", "mod-b", "mod-c"},
			wantComponents: [][]string{{"book", "go"}, {"go"}, {"typescript"}},
			wantStatuses:   []ModuleStatus{ModuleStatusSuccess, ModuleStatusFailed, ModuleStatusSuccess},
			wantDurations:  []time.Duration{0, 0, 0},
		},
		{
			name: "duration is max of component durations",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: 0, Duration: 10 * time.Second},
				{Module: "mod-a", Component: "book", ExitCode: 0, Duration: 5 * time.Second},
				{Module: "mod-a", Component: "typescript", ExitCode: 0, Duration: 15 * time.Second},
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"book", "go", "typescript"}},
			wantStatuses:   []ModuleStatus{ModuleStatusSuccess},
			wantDurations:  []time.Duration{15 * time.Second},
		},
		{
			name: "complex scenario with multiple modules and components",
			results: []ComponentResult{
				{Module: "eac-core", Component: "go", ExitCode: 0, Duration: 5 * time.Second},
				{Module: "eac-commands", Component: "go", ExitCode: 0, Duration: 10 * time.Second},
				{Module: "eac-core", Component: "typescript", ExitCode: 0, Duration: 3 * time.Second},
				{Module: "eac-commands", Component: "book", ExitCode: 1, Duration: 2 * time.Second},
				{Module: "eac-core", Component: "book", ExitCode: 0, Duration: 8 * time.Second},
			},
			wantModules:    []string{"eac-commands", "eac-core"},
			wantComponents: [][]string{{"book", "go"}, {"book", "go", "typescript"}},
			wantStatuses:   []ModuleStatus{ModuleStatusFailed, ModuleStatusSuccess},
			wantDurations:  []time.Duration{10 * time.Second, 8 * time.Second},
		},
		{
			name: "status with running components (negative exit code)",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: 0},
				{Module: "mod-a", Component: "book", ExitCode: -1}, // running
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"book", "go"}},
			wantStatuses:   []ModuleStatus{ModuleStatusRunning},
			wantDurations:  []time.Duration{0},
		},
		{
			name: "failure takes precedence over running",
			results: []ComponentResult{
				{Module: "mod-a", Component: "go", ExitCode: -1},  // running
				{Module: "mod-a", Component: "book", ExitCode: 1}, // failed
			},
			wantModules:    []string{"mod-a"},
			wantComponents: [][]string{{"book", "go"}},
			wantStatuses:   []ModuleStatus{ModuleStatusFailed},
			wantDurations:  []time.Duration{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateToComponentResultSets(tt.results)

			// Verify number of result sets
			if len(got) != len(tt.wantModules) {
				t.Fatalf("AggregateToComponentResultSets() returned %d sets, want %d", len(got), len(tt.wantModules))
			}

			// Verify each result set
			for i, rs := range got {
				// Check module name
				if rs.Module != tt.wantModules[i] {
					t.Errorf("ResultSet[%d].Module = %q, want %q", i, rs.Module, tt.wantModules[i])
				}

				// Check component count
				if len(rs.Components) != len(tt.wantComponents[i]) {
					t.Errorf("ResultSet[%d].Components length = %d, want %d", i, len(rs.Components), len(tt.wantComponents[i]))
					continue
				}

				// Check component names (should be sorted)
				for j, comp := range rs.Components {
					if comp.Component != tt.wantComponents[i][j] {
						t.Errorf("ResultSet[%d].Components[%d].Component = %q, want %q", i, j, comp.Component, tt.wantComponents[i][j])
					}
				}

				// Check status
				if rs.Status != tt.wantStatuses[i] {
					t.Errorf("ResultSet[%d].Status = %v, want %v", i, rs.Status, tt.wantStatuses[i])
				}

				// Check duration
				if rs.Duration != tt.wantDurations[i] {
					t.Errorf("ResultSet[%d].Duration = %v, want %v", i, rs.Duration, tt.wantDurations[i])
				}
			}
		})
	}
}

// TestAggregateToComponentResultSets_PreservesAllFields verifies that all ComponentResult fields
// are preserved in the aggregated output.
func TestAggregateToComponentResultSets_PreservesAllFields(t *testing.T) {
	input := []ComponentResult{
		{
			Module:    "mod-a",
			Component: "go",
			ExitCode:  1,
			Duration:  5 * time.Second,
			Errors:    []string{"error1", "error2"},
			Warnings:  []string{"warning1"},
			LogPath:   "/path/to/go.log",
		},
		{
			Module:    "mod-a",
			Component: "alpha", // will be sorted first
			ExitCode:  0,
			Duration:  2 * time.Second,
			Errors:    nil,
			Warnings:  []string{"warn-alpha"},
			LogPath:   "/path/to/alpha.log",
		},
	}

	got := AggregateToComponentResultSets(input)

	if len(got) != 1 {
		t.Fatalf("Expected 1 result set, got %d", len(got))
	}

	rs := got[0]
	if len(rs.Components) != 2 {
		t.Fatalf("Expected 2 components, got %d", len(rs.Components))
	}

	// alpha should be first (sorted)
	alpha := rs.Components[0]
	if alpha.Component != "alpha" {
		t.Fatalf("Expected alpha first, got %q", alpha.Component)
	}
	if alpha.Module != "mod-a" {
		t.Errorf("alpha.Module = %q, want %q", alpha.Module, "mod-a")
	}
	if alpha.ExitCode != 0 {
		t.Errorf("alpha.ExitCode = %d, want %d", alpha.ExitCode, 0)
	}
	if alpha.Duration != 2*time.Second {
		t.Errorf("alpha.Duration = %v, want %v", alpha.Duration, 2*time.Second)
	}
	if len(alpha.Warnings) != 1 || alpha.Warnings[0] != "warn-alpha" {
		t.Errorf("alpha.Warnings = %v, want [warn-alpha]", alpha.Warnings)
	}
	if alpha.LogPath != "/path/to/alpha.log" {
		t.Errorf("alpha.LogPath = %q, want %q", alpha.LogPath, "/path/to/alpha.log")
	}

	// go should be second
	goComp := rs.Components[1]
	if goComp.Component != "go" {
		t.Fatalf("Expected go second, got %q", goComp.Component)
	}
	if goComp.ExitCode != 1 {
		t.Errorf("go.ExitCode = %d, want %d", goComp.ExitCode, 1)
	}
	if len(goComp.Errors) != 2 {
		t.Errorf("go.Errors length = %d, want %d", len(goComp.Errors), 2)
	}
	if len(goComp.Warnings) != 1 {
		t.Errorf("go.Warnings length = %d, want %d", len(goComp.Warnings), 1)
	}
	if goComp.LogPath != "/path/to/go.log" {
		t.Errorf("go.LogPath = %q, want %q", goComp.LogPath, "/path/to/go.log")
	}
}

// TestAggregateToComponentResultSets_DoesNotModifyInput verifies that the input slice
// is not modified by the aggregation.
func TestAggregateToComponentResultSets_DoesNotModifyInput(t *testing.T) {
	input := []ComponentResult{
		{Module: "mod-b", Component: "zeta", ExitCode: 0, Duration: 1 * time.Second},
		{Module: "mod-a", Component: "beta", ExitCode: 0, Duration: 2 * time.Second},
		{Module: "mod-b", Component: "alpha", ExitCode: 0, Duration: 3 * time.Second},
	}

	// Save original values
	originalOrder := make([]string, len(input))
	for i, r := range input {
		originalOrder[i] = r.Module + ":" + r.Component
	}

	// Call the function
	_ = AggregateToComponentResultSets(input)

	// Verify input was not modified
	for i, r := range input {
		currentOrder := r.Module + ":" + r.Component
		if currentOrder != originalOrder[i] {
			t.Errorf("Input[%d] was modified: got %q, originally %q", i, currentOrder, originalOrder[i])
		}
	}
}

// TestAggregateToComponentResultSets_ModuleWithSamePrefix tests that modules with similar prefixes
// are sorted correctly.
func TestAggregateToComponentResultSets_ModuleWithSamePrefix(t *testing.T) {
	input := []ComponentResult{
		{Module: "eac-commands-extra", Component: "go", ExitCode: 0},
		{Module: "eac", Component: "go", ExitCode: 0},
		{Module: "eac-commands", Component: "go", ExitCode: 0},
	}

	got := AggregateToComponentResultSets(input)

	if len(got) != 3 {
		t.Fatalf("Expected 3 result sets, got %d", len(got))
	}

	expectedOrder := []string{"eac", "eac-commands", "eac-commands-extra"}
	for i, rs := range got {
		if rs.Module != expectedOrder[i] {
			t.Errorf("ResultSet[%d].Module = %q, want %q", i, rs.Module, expectedOrder[i])
		}
	}
}

// TestComponentResultSet_DeriveStatus_Priority tests that failure status takes
// precedence over running status when both conditions exist.
func TestComponentResultSet_DeriveStatus_Priority(t *testing.T) {
	// When both failed (ExitCode > 0) and running (ExitCode < 0) exist,
	// failed should take precedence
	tests := []struct {
		name       string
		components []ComponentResult
		want       ModuleStatus
	}{
		{
			name: "failure first, then running",
			components: []ComponentResult{
				{Component: "a", ExitCode: 1},
				{Component: "b", ExitCode: -1},
				{Component: "c", ExitCode: 0},
			},
			want: ModuleStatusFailed,
		},
		{
			name: "running first, then failure",
			components: []ComponentResult{
				{Component: "a", ExitCode: -1},
				{Component: "b", ExitCode: 1},
				{Component: "c", ExitCode: 0},
			},
			want: ModuleStatusFailed,
		},
		{
			name: "success and running only",
			components: []ComponentResult{
				{Component: "a", ExitCode: 0},
				{Component: "b", ExitCode: -1},
				{Component: "c", ExitCode: 0},
			},
			want: ModuleStatusRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := ComponentResultSet{
				Module:     "test",
				Components: tt.components,
			}
			got := rs.DeriveStatus()
			if got != tt.want {
				t.Errorf("DeriveStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
