package execution

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/workunit"
)

// Helper to create a test UnitSpec
func testUnit(module, component, tool string, dependsOn ...workunit.UnitID) workunit.UnitSpec {
	return workunit.UnitSpec{
		ID: workunit.UnitID{
			Context:   workunit.ContextBuild,
			Module:    module,
			Component: component,
			Tool:      tool,
		},
		ComponentType: component,
		Weight:        1,
		DependsOn:     dependsOn,
	}
}

// Helper to create a UnitID for dependency references
func testDep(module, component, tool string) workunit.UnitID {
	return workunit.UnitID{
		Context:   workunit.ContextBuild,
		Module:    module,
		Component: component,
		Tool:      tool,
	}
}

func TestLayerPlanner_ComputePlan_EmptyInput(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)
	plan, err := planner.ComputePlan(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil {
		t.Fatal("expected non-nil plan")
	}
	if len(plan.ModuleLayers) != 0 {
		t.Errorf("expected 0 module layers, got %d", len(plan.ModuleLayers))
	}
	if plan.Stats.TotalUoWs != 0 {
		t.Errorf("expected 0 UoWs, got %d", plan.Stats.TotalUoWs)
	}
}

func TestLayerPlanner_ComputePlan_SingleModule(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("core", "ts", "tsc"),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.ModuleLayers) != 1 {
		t.Errorf("expected 1 module layer, got %d", len(plan.ModuleLayers))
	}

	if plan.Stats.TotalUoWs != 2 {
		t.Errorf("expected 2 UoWs, got %d", plan.Stats.TotalUoWs)
	}

	// Both units should be in the same module layer
	if len(plan.ModuleLayers[0].Modules) != 1 {
		t.Errorf("expected 1 module in layer 0, got %d", len(plan.ModuleLayers[0].Modules))
	}
	if plan.ModuleLayers[0].Modules[0] != "core" {
		t.Errorf("expected module 'core', got '%s'", plan.ModuleLayers[0].Modules[0])
	}
}

func TestLayerPlanner_ComputePlan_ModuleDependencies(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	// cli depends on core
	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("cli", "go", "go-build", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.ModuleLayers) != 2 {
		t.Errorf("expected 2 module layers, got %d", len(plan.ModuleLayers))
	}

	// Layer 0 should have core
	if len(plan.ModuleLayers[0].Modules) != 1 || plan.ModuleLayers[0].Modules[0] != "core" {
		t.Errorf("expected layer 0 to have ['core'], got %v", plan.ModuleLayers[0].Modules)
	}

	// Layer 1 should have cli
	if len(plan.ModuleLayers[1].Modules) != 1 || plan.ModuleLayers[1].Modules[0] != "cli" {
		t.Errorf("expected layer 1 to have ['cli'], got %v", plan.ModuleLayers[1].Modules)
	}
}

func TestLayerPlanner_ComputePlan_ComponentDependencies(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	// ts depends on go within same module
	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("core", "ts", "tsc", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.ModuleLayers) != 1 {
		t.Errorf("expected 1 module layer, got %d", len(plan.ModuleLayers))
	}

	// Should have 2 component layers within the module layer
	ml := plan.ModuleLayers[0]
	if len(ml.ComponentLayers) != 2 {
		t.Errorf("expected 2 component layers, got %d", len(ml.ComponentLayers))
	}

	// Layer 0 should have go
	if len(ml.ComponentLayers[0].Units) != 1 {
		t.Errorf("expected 1 unit in component layer 0, got %d", len(ml.ComponentLayers[0].Units))
	}
	if ml.ComponentLayers[0].Units[0].ID.Component != "go" {
		t.Errorf("expected 'go' in component layer 0, got '%s'", ml.ComponentLayers[0].Units[0].ID.Component)
	}

	// Layer 1 should have ts
	if len(ml.ComponentLayers[1].Units) != 1 {
		t.Errorf("expected 1 unit in component layer 1, got %d", len(ml.ComponentLayers[1].Units))
	}
	if ml.ComponentLayers[1].Units[0].ID.Component != "ts" {
		t.Errorf("expected 'ts' in component layer 1, got '%s'", ml.ComponentLayers[1].Units[0].ID.Component)
	}
}

func TestLayerPlanner_ComputePlan_CircularModuleDependency(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	// Circular: a -> b -> a
	units := []workunit.UnitSpec{
		testUnit("a", "go", "go-build", testDep("b", "go", "go-build")),
		testUnit("b", "go", "go-build", testDep("a", "go", "go-build")),
	}

	_, err := planner.ComputePlan(units)
	if err == nil {
		t.Error("expected error for circular dependency")
	}
}

func TestLayerPlanner_ComputePlan_ParallelModules(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	// Three independent modules
	units := []workunit.UnitSpec{
		testUnit("a", "go", "go-build"),
		testUnit("b", "go", "go-build"),
		testUnit("c", "go", "go-build"),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All modules should be in the same layer (no dependencies)
	if len(plan.ModuleLayers) != 1 {
		t.Errorf("expected 1 module layer, got %d", len(plan.ModuleLayers))
	}

	if len(plan.ModuleLayers[0].Modules) != 3 {
		t.Errorf("expected 3 modules in layer 0, got %d", len(plan.ModuleLayers[0].Modules))
	}

	// Max parallelism should be 3
	if plan.Stats.MaxParallelism != 3 {
		t.Errorf("expected max parallelism 3, got %d", plan.Stats.MaxParallelism)
	}
}

func TestLayerPlanner_IsReady_Strict_PreviousLayerIncomplete(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("cli", "go", "go-build", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	completed := make(map[string]bool)

	// cli should NOT be ready (core not complete)
	cliID := units[1].ID
	if planner.IsReady(plan, cliID, completed) {
		t.Error("cli should not be ready when core is incomplete")
	}

	// core should be ready (no dependencies)
	coreID := units[0].ID
	if !planner.IsReady(plan, coreID, completed) {
		t.Error("core should be ready")
	}

	// Mark core complete
	completed[coreID.Longname()] = true

	// Now cli should be ready
	if !planner.IsReady(plan, cliID, completed) {
		t.Error("cli should be ready after core completes")
	}
}

func TestLayerPlanner_IsReady_Strict_ComponentLayer(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("core", "ts", "tsc", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	completed := make(map[string]bool)

	// ts should NOT be ready (go not complete)
	tsID := units[1].ID
	if planner.IsReady(plan, tsID, completed) {
		t.Error("ts should not be ready when go is incomplete")
	}

	// go should be ready
	goID := units[0].ID
	if !planner.IsReady(plan, goID, completed) {
		t.Error("go should be ready")
	}

	// Mark go complete
	completed[goID.Longname()] = true

	// Now ts should be ready
	if !planner.IsReady(plan, tsID, completed) {
		t.Error("ts should be ready after go completes")
	}
}

func TestLayerPlanner_IsReady_None_ComponentLayersStillEnforced(t *testing.T) {
	planner := NewLayerPlanner(LayerModeNone)

	// Two components in same module, ts depends on go (via component layer)
	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("core", "ts", "tsc", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	completed := make(map[string]bool)

	// In None mode, module layers are not enforced BUT component layers ARE.
	// ts depends on go via explicit DependsOn, so it should still wait.
	goID := units[0].ID
	tsID := units[1].ID

	if !planner.IsReady(plan, goID, completed) {
		t.Error("go should be ready")
	}

	// ts should NOT be ready because of component layer + explicit DependsOn
	if planner.IsReady(plan, tsID, completed) {
		t.Error("ts should not be ready due to component layer dependency")
	}

	// Mark go complete
	completed[goID.Longname()] = true

	// Now ts should be ready
	if !planner.IsReady(plan, tsID, completed) {
		t.Error("ts should be ready after go completes")
	}
}

func TestLayerPlanner_IsReady_None(t *testing.T) {
	planner := NewLayerPlanner(LayerModeNone)

	// cli depends on core
	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("cli", "go", "go-build", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	completed := make(map[string]bool)

	// In None mode, only explicit DependsOn matters
	coreID := units[0].ID
	cliID := units[1].ID

	// core should be ready
	if !planner.IsReady(plan, coreID, completed) {
		t.Error("core should be ready")
	}

	// cli should NOT be ready (explicit DependsOn on core)
	if planner.IsReady(plan, cliID, completed) {
		t.Error("cli should not be ready due to explicit DependsOn")
	}

	// Mark core complete
	completed[coreID.Longname()] = true

	// Now cli should be ready
	if !planner.IsReady(plan, cliID, completed) {
		t.Error("cli should be ready after core completes")
	}
}

func TestLayerPlanner_GetReadyUnits(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	units := []workunit.UnitSpec{
		{
			ID: workunit.UnitID{
				Context:   workunit.ContextBuild,
				Module:    "core",
				Component: "go",
				Tool:      "go-build",
			},
			Weight: 5,
		},
		{
			ID: workunit.UnitID{
				Context:   workunit.ContextBuild,
				Module:    "core",
				Component: "ts",
				Tool:      "tsc",
			},
			Weight: 3,
		},
		{
			ID: workunit.UnitID{
				Context:   workunit.ContextBuild,
				Module:    "cli",
				Component: "go",
				Tool:      "go-build",
			},
			Weight: 2,
			DependsOn: []workunit.UnitID{
				testDep("core", "go", "go-build"),
			},
		},
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	completed := make(map[string]bool)

	// Initially, only core units should be ready
	ready := planner.GetReadyUnits(plan, completed)
	if len(ready) != 2 {
		t.Errorf("expected 2 ready units, got %d", len(ready))
	}

	// Should be sorted by weight descending (go:5, ts:3)
	if len(ready) >= 2 && ready[0].ID.Component != "go" {
		t.Errorf("expected 'go' first (weight 5), got '%s'", ready[0].ID.Component)
	}
	if len(ready) >= 2 && ready[1].ID.Component != "ts" {
		t.Errorf("expected 'ts' second (weight 3), got '%s'", ready[1].ID.Component)
	}

	// Mark core units complete
	completed[units[0].ID.Longname()] = true
	completed[units[1].ID.Longname()] = true

	// Now cli should be ready
	ready = planner.GetReadyUnits(plan, completed)
	if len(ready) != 1 {
		t.Errorf("expected 1 ready unit, got %d", len(ready))
	}
	if len(ready) >= 1 && ready[0].ID.Module != "cli" {
		t.Errorf("expected 'cli' to be ready, got '%s'", ready[0].ID.Module)
	}
}

func TestLayerPlan_FindUnit(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("cli", "go", "go-build", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find core:go
	coreID := units[0].ID
	mlIdx, clIdx, uIdx := plan.FindUnit(coreID)
	if mlIdx != 0 {
		t.Errorf("expected module layer 0, got %d", mlIdx)
	}
	if clIdx != 0 {
		t.Errorf("expected component layer 0, got %d", clIdx)
	}
	if uIdx != 0 {
		t.Errorf("expected unit index 0, got %d", uIdx)
	}

	// Find cli:go
	cliID := units[1].ID
	mlIdx, clIdx, uIdx = plan.FindUnit(cliID)
	if mlIdx != 1 {
		t.Errorf("expected module layer 1, got %d", mlIdx)
	}

	// Find non-existent unit
	badID := workunit.UnitID{
		Context:   workunit.ContextBuild,
		Module:    "nonexistent",
		Component: "go",
		Tool:      "go-build",
	}
	mlIdx, clIdx, uIdx = plan.FindUnit(badID)
	if mlIdx != -1 || clIdx != -1 || uIdx != -1 {
		t.Errorf("expected (-1,-1,-1) for non-existent unit, got (%d,%d,%d)", mlIdx, clIdx, uIdx)
	}
}

func TestLayerPlan_AllUnits(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("core", "ts", "tsc"),
		testUnit("cli", "go", "go-build", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allUnits := plan.AllUnits()
	if len(allUnits) != 3 {
		t.Errorf("expected 3 units, got %d", len(allUnits))
	}
}

func TestLayerPlan_UnitsInModule(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("core", "ts", "tsc"),
		testUnit("cli", "go", "go-build", testDep("core", "go", "go-build")),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	coreUnits := plan.UnitsInModule("core")
	if len(coreUnits) != 2 {
		t.Errorf("expected 2 units in core, got %d", len(coreUnits))
	}

	cliUnits := plan.UnitsInModule("cli")
	if len(cliUnits) != 1 {
		t.Errorf("expected 1 unit in cli, got %d", len(cliUnits))
	}
}

func TestLayerMode_String(t *testing.T) {
	tests := []struct {
		mode     LayerMode
		expected string
	}{
		{LayerModeStrict, "strict"},
		{LayerModeNone, "none"},
		{LayerMode(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.expected {
			t.Errorf("LayerMode(%d).String() = %q, want %q", tt.mode, got, tt.expected)
		}
	}
}

func TestParseLayerMode(t *testing.T) {
	tests := []struct {
		input    string
		expected LayerMode
	}{
		{"strict", LayerModeStrict},
		{"layered", LayerModeStrict},
		{"none", LayerModeNone},
		{"unlayered", LayerModeNone},
		{"invalid", LayerModeStrict}, // default
		{"", LayerModeStrict},        // default
	}

	for _, tt := range tests {
		if got := ParseLayerMode(tt.input); got != tt.expected {
			t.Errorf("ParseLayerMode(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestSimpleCompletionTracker(t *testing.T) {
	tracker := NewSimpleCompletionTracker()

	id1 := workunit.UnitID{
		Context:   workunit.ContextBuild,
		Module:    "core",
		Component: "go",
		Tool:      "go-build",
	}

	id2 := workunit.UnitID{
		Context:   workunit.ContextBuild,
		Module:    "cli",
		Component: "go",
		Tool:      "go-build",
	}

	// Initially nothing is complete
	if tracker.IsComplete(id1) {
		t.Error("id1 should not be complete initially")
	}

	// Mark id1 complete
	tracker.MarkComplete(id1)
	if !tracker.IsComplete(id1) {
		t.Error("id1 should be complete after MarkComplete")
	}
	if tracker.IsFailed(id1) {
		t.Error("id1 should not be failed")
	}

	// Mark id2 failed
	tracker.MarkFailed(id2)
	if !tracker.IsComplete(id2) {
		t.Error("id2 should be complete (failed is also complete)")
	}
	if !tracker.IsFailed(id2) {
		t.Error("id2 should be failed")
	}

	// Check Completed snapshot
	completed := tracker.Completed()
	if !completed[id1.Longname()] {
		t.Error("id1 should be in completed snapshot")
	}
	if !completed[id2.Longname()] {
		t.Error("id2 should be in completed snapshot")
	}
}

func TestLayerPlanner_ComplexDependencyGraph(t *testing.T) {
	planner := NewLayerPlanner(LayerModeStrict)

	// Complex graph:
	// Layer 0: core
	// Layer 1: lib1, lib2 (both depend on core)
	// Layer 2: cli (depends on lib1 and lib2)
	units := []workunit.UnitSpec{
		testUnit("core", "go", "go-build"),
		testUnit("lib1", "go", "go-build", testDep("core", "go", "go-build")),
		testUnit("lib2", "go", "go-build", testDep("core", "go", "go-build")),
		testUnit("cli", "go", "go-build",
			testDep("lib1", "go", "go-build"),
			testDep("lib2", "go", "go-build"),
		),
	}

	plan, err := planner.ComputePlan(units)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.ModuleLayers) != 3 {
		t.Errorf("expected 3 module layers, got %d", len(plan.ModuleLayers))
	}

	// Layer 0: core
	if len(plan.ModuleLayers[0].Modules) != 1 || plan.ModuleLayers[0].Modules[0] != "core" {
		t.Errorf("expected layer 0 = [core], got %v", plan.ModuleLayers[0].Modules)
	}

	// Layer 1: lib1, lib2 (sorted alphabetically)
	if len(plan.ModuleLayers[1].Modules) != 2 {
		t.Errorf("expected 2 modules in layer 1, got %d", len(plan.ModuleLayers[1].Modules))
	}
	if plan.ModuleLayers[1].Modules[0] != "lib1" || plan.ModuleLayers[1].Modules[1] != "lib2" {
		t.Errorf("expected layer 1 = [lib1, lib2], got %v", plan.ModuleLayers[1].Modules)
	}

	// Layer 2: cli
	if len(plan.ModuleLayers[2].Modules) != 1 || plan.ModuleLayers[2].Modules[0] != "cli" {
		t.Errorf("expected layer 2 = [cli], got %v", plan.ModuleLayers[2].Modules)
	}
}
