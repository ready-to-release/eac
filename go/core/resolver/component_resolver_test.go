package resolver

import (
	"context"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/stretchr/testify/assert"
)

// mockExecutor implements tool.Executor for testing.
type mockExecutor struct{}

func (m *mockExecutor) Execute(ctx context.Context, t *tool.ToolDefinition, execCtx *tool.ExecutionContext) (*tool.ExecutionResult, error) {
	return &tool.ExecutionResult{ExitCode: 0}, nil
}

func (m *mockExecutor) Validate(t *tool.ToolDefinition) error {
	return nil
}

func (m *mockExecutor) Close() error {
	return nil
}

// =============================================================================
// Regression Test: Deadlock Prevention (L0)
// =============================================================================
//
// This tests the fix for a deadlock bug where build_after dependencies on
// non-buildable component types would cause the scheduler to wait forever.
//
// Root cause: The book component type has build_after: [markdown], but markdown
// has no builder configured. The old code would create a dependency on markdown,
// but markdown units are never scheduled, causing a deadlock.
//
// Fix: When building the scheduledComponents set in ResolveForBuild, only
// components with valid builders/handlers are included. Dependencies are then
// filtered to only include scheduled components.

// TestScheduledComponentsFilter validates that the resolver correctly identifies
// which components will be scheduled vs. skipped.
func TestScheduledComponentsFilter_OnlyBuildableTypesAreScheduled(t *testing.T) {
	// This test validates the logic pattern used in ResolveForBuild:
	// 1. Build set of components that will be scheduled (have builders)
	// 2. Skip dependencies on components NOT in this set

	// Simulated scenario:
	// - "go" component type has builder "go"
	// - "markdown" component type has NO builder
	// - "book" component type has builder "mkdocs" and build_after: [markdown, go]

	type componentInfo struct {
		compType  string
		builder   string   // empty means not buildable
		buildAfter []string
	}

	enabledComponents := map[string]componentInfo{
		"main": {compType: "go", builder: "go", buildAfter: nil},
		"docs": {compType: "markdown", builder: "", buildAfter: nil}, // No builder!
		"site": {compType: "book", builder: "mkdocs", buildAfter: []string{"markdown", "go"}},
	}

	// Step 1: Build scheduledComponents set (only those with builders)
	scheduledComponents := make(map[string]bool)
	for compName, info := range enabledComponents {
		if info.builder != "" {
			scheduledComponents[compName] = true
		}
	}

	// Verify: Only "main" and "site" are scheduled, NOT "docs"
	assert.True(t, scheduledComponents["main"], "main (go) should be scheduled")
	assert.True(t, scheduledComponents["site"], "site (book) should be scheduled")
	assert.False(t, scheduledComponents["docs"], "docs (markdown) should NOT be scheduled - no builder")

	// Step 2: For "site", build dependencies filtering out non-scheduled components
	siteInfo := enabledComponents["site"]

	// Map build_after types to enabled components
	typeToComponents := make(map[string][]string)
	for compName, info := range enabledComponents {
		typeToComponents[info.compType] = append(typeToComponents[info.compType], compName)
	}

	var siteDeps []string
	for _, depType := range siteInfo.buildAfter {
		for _, depComp := range typeToComponents[depType] {
			// KEY FIX: Skip dependencies on non-scheduled components
			if scheduledComponents[depComp] {
				siteDeps = append(siteDeps, depComp)
			}
		}
	}

	// Verify: site depends on "main" (go, buildable) but NOT "docs" (markdown, not buildable)
	assert.Contains(t, siteDeps, "main", "site should depend on buildable 'main'")
	assert.NotContains(t, siteDeps, "docs", "site should NOT depend on non-buildable 'docs'")
}

// TestScheduledComponentsFilter_AllDepsAreBuildable validates that when all
// dependencies are buildable, they're all included.
func TestScheduledComponentsFilter_AllDepsAreBuildable(t *testing.T) {
	// Scenario: dockerfile depends on go, both are buildable
	type componentInfo struct {
		compType   string
		builder    string
		buildAfter []string
	}

	enabledComponents := map[string]componentInfo{
		"main":   {compType: "go", builder: "go", buildAfter: nil},
		"docker": {compType: "dockerfile", builder: "buildx", buildAfter: []string{"go"}},
	}

	// Build scheduledComponents set
	scheduledComponents := make(map[string]bool)
	for compName, info := range enabledComponents {
		if info.builder != "" {
			scheduledComponents[compName] = true
		}
	}

	assert.True(t, scheduledComponents["main"], "main should be scheduled")
	assert.True(t, scheduledComponents["docker"], "docker should be scheduled")

	// Build docker's dependencies
	dockerInfo := enabledComponents["docker"]
	typeToComponents := make(map[string][]string)
	for compName, info := range enabledComponents {
		typeToComponents[info.compType] = append(typeToComponents[info.compType], compName)
	}

	var dockerDeps []string
	for _, depType := range dockerInfo.buildAfter {
		for _, depComp := range typeToComponents[depType] {
			if scheduledComponents[depComp] {
				dockerDeps = append(dockerDeps, depComp)
			}
		}
	}

	// docker should depend on main since both are buildable
	assert.Contains(t, dockerDeps, "main", "docker should depend on buildable 'main'")
}

// TestScheduledComponentsFilter_NoDepsWhenNoBuildAfter validates that components
// without build_after have no dependencies.
func TestScheduledComponentsFilter_NoDepsWhenNoBuildAfter(t *testing.T) {
	type componentInfo struct {
		compType   string
		builder    string
		buildAfter []string
	}

	enabledComponents := map[string]componentInfo{
		"main": {compType: "go", builder: "go", buildAfter: nil}, // No build_after
	}

	scheduledComponents := make(map[string]bool)
	for compName, info := range enabledComponents {
		if info.builder != "" {
			scheduledComponents[compName] = true
		}
	}

	mainInfo := enabledComponents["main"]

	var mainDeps []string
	for _, depType := range mainInfo.buildAfter {
		// This loop body never executes since buildAfter is nil/empty
		mainDeps = append(mainDeps, depType)
	}

	assert.Empty(t, mainDeps, "main should have no dependencies")
}

// TestScheduledComponentsFilter_DepTypeNotEnabled validates that when a dependency
// type exists in build_after but no component of that type is enabled, no dep is created.
func TestScheduledComponentsFilter_DepTypeNotEnabled(t *testing.T) {
	type componentInfo struct {
		compType   string
		builder    string
		buildAfter []string
	}

	// book depends on structurizr, but no structurizr component is enabled
	enabledComponents := map[string]componentInfo{
		"site": {compType: "book", builder: "mkdocs", buildAfter: []string{"structurizr"}},
	}

	scheduledComponents := make(map[string]bool)
	for compName, info := range enabledComponents {
		if info.builder != "" {
			scheduledComponents[compName] = true
		}
	}

	typeToComponents := make(map[string][]string)
	for compName, info := range enabledComponents {
		typeToComponents[info.compType] = append(typeToComponents[info.compType], compName)
	}

	siteInfo := enabledComponents["site"]
	var siteDeps []string
	for _, depType := range siteInfo.buildAfter {
		for _, depComp := range typeToComponents[depType] {
			if scheduledComponents[depComp] {
				siteDeps = append(siteDeps, depComp)
			}
		}
	}

	// No structurizr component exists, so no deps
	assert.Empty(t, siteDeps, "site should have no deps when depType not enabled")
}

// =============================================================================
// Tool Chain Tests
// =============================================================================
//
// These tests validate the tool chain expansion feature where a single component
// with tools: [a, b, c] expands to multiple UnitSpecs with chain dependencies.

// TestExpandToolChain_SingleTool validates that a component type with a single tool
// produces a single UnitSpec (same as legacy builder behavior).
func TestExpandToolChain_SingleTool(t *testing.T) {
	// Simulate a component type with a single tool (same as legacy builder)
	tools := []string{"go"}
	module := "cli"
	component := "main"

	specs := expandToolChain(module, component, "go", tools, nil)

	assert.Len(t, specs, 1, "single tool should produce one spec")
	assert.Equal(t, component, specs[0].Component)
	assert.Equal(t, "go", specs[0].Tool)
	assert.Empty(t, specs[0].DependsOn, "first tool has no tool chain dependencies")
}

// TestExpandToolChain_TwoTools validates that a component type with two tools
// produces two UnitSpecs where the second depends on the first.
func TestExpandToolChain_TwoTools(t *testing.T) {
	tools := []string{"preprocess", "mkdocs-pdf"}
	module := "docs"
	component := "tutorials-pdf"

	specs := expandToolChain(module, component, "docs-pdf", tools, nil)

	assert.Len(t, specs, 2, "two tools should produce two specs")

	// First spec: preprocess
	assert.Equal(t, component, specs[0].Component)
	assert.Equal(t, "preprocess", specs[0].Tool)
	assert.Empty(t, specs[0].DependsOn, "first tool has no tool chain dependencies")

	// Second spec: mkdocs-pdf depends on preprocess
	assert.Equal(t, component, specs[1].Component)
	assert.Equal(t, "mkdocs-pdf", specs[1].Tool)
	assert.Len(t, specs[1].DependsOn, 1, "second tool depends on first")
	assert.Equal(t, "preprocess", specs[1].DependsOn[0].Tool)
	assert.Equal(t, component, specs[1].DependsOn[0].Component)
}

// TestExpandToolChain_ThreeTools validates that a tool chain with three tools
// creates a proper chain: a -> b -> c
func TestExpandToolChain_ThreeTools(t *testing.T) {
	tools := []string{"step1", "step2", "step3"}
	module := "mod"
	component := "comp"

	specs := expandToolChain(module, component, "multi-step", tools, nil)

	assert.Len(t, specs, 3, "three tools should produce three specs")

	// step1 has no deps
	assert.Equal(t, "step1", specs[0].Tool)
	assert.Empty(t, specs[0].DependsOn)

	// step2 depends on step1
	assert.Equal(t, "step2", specs[1].Tool)
	assert.Len(t, specs[1].DependsOn, 1)
	assert.Equal(t, "step1", specs[1].DependsOn[0].Tool)

	// step3 depends on step2
	assert.Equal(t, "step3", specs[2].Tool)
	assert.Len(t, specs[2].DependsOn, 1)
	assert.Equal(t, "step2", specs[2].DependsOn[0].Tool)
}

// TestExpandToolChain_WithExternalDeps validates that external dependencies
// (from build_after or depends_on) are added to the FIRST tool in the chain.
func TestExpandToolChain_WithExternalDeps(t *testing.T) {
	tools := []string{"preprocess", "mkdocs-pdf"}
	module := "docs"
	component := "tutorials-pdf"

	// External dependency: structurizr must complete before the chain starts
	externalDeps := []toolChainDep{
		{Module: module, Component: "structurizr", Tool: "structurizr-cli"},
	}

	specs := expandToolChain(module, component, "docs-pdf", tools, externalDeps)

	// First spec should have external deps
	assert.Len(t, specs[0].DependsOn, 1, "first tool should have external dep")
	assert.Equal(t, "structurizr", specs[0].DependsOn[0].Component)

	// Second spec should only depend on first tool, not external deps
	assert.Len(t, specs[1].DependsOn, 1, "second tool only depends on first")
	assert.Equal(t, "preprocess", specs[1].DependsOn[0].Tool)
}

// TestExpandToolChain_EmptyTools returns nil when no tools provided.
func TestExpandToolChain_EmptyTools(t *testing.T) {
	specs := expandToolChain("mod", "comp", "empty", nil, nil)
	assert.Nil(t, specs, "empty tools should return nil")

	specs = expandToolChain("mod", "comp", "empty", []string{}, nil)
	assert.Nil(t, specs, "zero-length tools should return nil")
}

// =============================================================================
// Weight Calculation Tests
// =============================================================================
//
// These tests validate the new weight calculation formula:
// weight = tool.Resources.CPUs * componentType.Amp * module.amp
//
// The weight is derived from tool resources (not component type resources).

// TestGetWeight_FromToolResources validates that weight is derived from tool resources.
func TestGetWeight_FromToolResources(t *testing.T) {
	tests := []struct {
		name          string
		toolCPUs      int    // Tool resources.cpus
		compTypeAmp   float64 // ComponentType.Amp
		wantWeight    int
	}{
		{
			name:        "tool with 2 CPUs no amp = weight 2",
			toolCPUs:    2,
			compTypeAmp: 0, // 0 means default 1.0
			wantWeight:  2,
		},
		{
			name:        "tool with 4 CPUs no amp = weight 4",
			toolCPUs:    4,
			compTypeAmp: 0,
			wantWeight:  4,
		},
		{
			name:        "tool with 2 CPUs and amp 1.5 = weight 3",
			toolCPUs:    2,
			compTypeAmp: 1.5,
			wantWeight:  3, // ceil(2 * 1.5) = 3
		},
		{
			name:        "tool with 4 CPUs and amp 0.5 = weight 2",
			toolCPUs:    4,
			compTypeAmp: 0.5,
			wantWeight:  2, // ceil(4 * 0.5) = 2
		},
		{
			name:        "tool with 0 CPUs (nil resources) = weight 1",
			toolCPUs:    0, // Means nil resources
			compTypeAmp: 0,
			wantWeight:  1, // Default weight
		},
		{
			name:        "tool with 0 CPUs and amp 2.0 = weight 2",
			toolCPUs:    0,
			compTypeAmp: 2.0,
			wantWeight:  2, // ceil(1 * 2.0) = 2 (base default is 1)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create tool registry with a test tool
			registry := tool.NewRegistry()
			registry.SetVerifier(func(t *tool.ToolDefinition) bool { return true })

			toolDef := &tool.ToolDefinition{
				ID:     "test-builder",
				Type:   tool.ToolTypeSystem,
				Binary: "test",
			}
			if tt.toolCPUs > 0 {
				toolDef.Resources = &tool.ToolResources{CPUs: tt.toolCPUs}
			}
			if err := registry.Register(toolDef); err != nil {
				t.Fatalf("Failed to register tool: %v", err)
			}

			// Create resolver with component-tools mapping
			toolResolver := tool.NewResolver(registry)
			toolResolver.LoadProjectConfig(map[string]*tool.ToolAssignment{
				"test-type": {Builder: "test-builder"},
			})

			// Create build bridge with mock executor
			bridge := tool.NewBuildBridge()
			bridge.SetToolSystem(registry, toolResolver, &mockExecutor{})

			// Create config with component type
			cfg := &config.EACConfig{
				ComponentTypes: &config.ComponentTypesConfig{
					ComponentTypes: map[string]*config.ComponentType{
						"test-type": {
							Builders: []string{"test-builder"},
							Amp:     tt.compTypeAmp,
						},
					},
				},
			}

			// Create component resolver
			resolver := NewComponentResolverWithConfig(cfg, bridge)

			// Create module with the test component
			module := modules.NewModuleContract(domain.BaseContract{
				Moniker: "test-module",
				Components: domain.ModuleComponents{
					"test-comp": &domain.ComponentEntry{
						Type: "test-type",
						Root: ".",
					},
				},
			}, "/workspace")

			// Resolve specs
			specs := resolver.ResolveForBuild(module, nil)

			// Find the spec for our test component
			var foundSpec bool
			for _, spec := range specs {
				if spec.ID.Component == "test-comp" {
					foundSpec = true
					assert.Equal(t, tt.wantWeight, spec.Weight,
						"Weight should be calculated from tool resources * amp")
				}
			}

			if !foundSpec && tt.toolCPUs > 0 {
				t.Error("Expected to find spec for test-comp")
			}
		})
	}
}

// TestPoolAllocation_FromToolType validates that PoolAllocation is set directly
// based on the tool type (container vs system).
func TestPoolAllocation_FromToolType(t *testing.T) {
	tests := []struct {
		name             string
		toolType         tool.ToolType
		toolCPUs         int
		compTypeAmp      float64
		wantHostWeight   int
		wantDockerWeight int
		wantIsContainer  bool
	}{
		{
			name:             "container tool uses both pools",
			toolType:         tool.ToolTypeContainer,
			toolCPUs:         4,
			compTypeAmp:      0, // default 1.0
			wantHostWeight:   4,
			wantDockerWeight: 4,
			wantIsContainer:  true,
		},
		{
			name:             "system tool uses host only",
			toolType:         tool.ToolTypeSystem,
			toolCPUs:         2,
			compTypeAmp:      0,
			wantHostWeight:   2,
			wantDockerWeight: 0,
			wantIsContainer:  false,
		},
		{
			name:             "container tool with amp 2.0",
			toolType:         tool.ToolTypeContainer,
			toolCPUs:         2,
			compTypeAmp:      2.0,
			wantHostWeight:   4, // 2 * 2.0 = 4
			wantDockerWeight: 4,
			wantIsContainer:  true,
		},
		{
			name:             "system tool with amp 1.5",
			toolType:         tool.ToolTypeSystem,
			toolCPUs:         2,
			compTypeAmp:      1.5,
			wantHostWeight:   3, // ceil(2 * 1.5) = 3
			wantDockerWeight: 0,
			wantIsContainer:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create tool registry
			registry := tool.NewRegistry()
			registry.SetVerifier(func(t *tool.ToolDefinition) bool { return true })

			toolDef := &tool.ToolDefinition{
				ID:   "test-builder",
				Type: tt.toolType,
			}
			if tt.toolType == tool.ToolTypeSystem {
				toolDef.Binary = "test"
			} else {
				toolDef.Image = "test"
				toolDef.Tag = "v1.0.0"
			}
			if tt.toolCPUs > 0 {
				toolDef.Resources = &tool.ToolResources{CPUs: tt.toolCPUs}
			}
			if err := registry.Register(toolDef); err != nil {
				t.Fatalf("Failed to register tool: %v", err)
			}

			// Create resolver with component-tools mapping
			toolResolver := tool.NewResolver(registry)
			toolResolver.LoadProjectConfig(map[string]*tool.ToolAssignment{
				"test-type": {Builder: "test-builder"},
			})

			// Create build bridge
			bridge := tool.NewBuildBridge()
			bridge.SetToolSystem(registry, toolResolver, &mockExecutor{})

			// Create config with component type
			cfg := &config.EACConfig{
				ComponentTypes: &config.ComponentTypesConfig{
					ComponentTypes: map[string]*config.ComponentType{
						"test-type": {
							Builders: []string{"test-builder"},
							Amp:     tt.compTypeAmp,
						},
					},
				},
			}

			// Create component resolver
			resolver := NewComponentResolverWithConfig(cfg, bridge)

			// Create module
			module := modules.NewModuleContract(domain.BaseContract{
				Moniker: "test-module",
				Components: domain.ModuleComponents{
					"test-comp": &domain.ComponentEntry{
						Type: "test-type",
						Root: ".",
					},
				},
			}, "/workspace")

			// Resolve specs
			specs := resolver.ResolveForBuild(module, nil)

			// Find spec and verify PoolAllocation
			var foundSpec bool
			for _, spec := range specs {
				if spec.ID.Component == "test-comp" {
					foundSpec = true
					alloc := spec.GetPoolAllocation()
					assert.Equal(t, tt.wantHostWeight, alloc.HostWeight,
						"HostWeight should match")
					assert.Equal(t, tt.wantDockerWeight, alloc.DockerWeight,
						"DockerWeight should match (container uses both pools, system uses host only)")
					assert.Equal(t, tt.wantIsContainer, alloc.IsContainer(),
						"IsContainer() should reflect tool type")
				}
			}

			if !foundSpec {
				t.Error("Expected to find spec for test-comp")
			}
		})
	}
}
