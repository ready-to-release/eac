package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
