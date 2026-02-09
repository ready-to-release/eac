package output

import (
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// AggregatedChangeResult Type Tests
// =============================================================================

func TestAggregatedChangeResult_FieldsExist(t *testing.T) {
	result := &AggregatedChangeResult{
		UoWResult: &UoWChangeResult{
			Changed:  []workunit.UnitID{},
			UpToDate: []workunit.UnitID{},
		},
		CachedUoWs:     make(map[string]bool),
		UoWCacheTimes:  make(map[string]time.Time),
		CachedModules:  make(map[string]bool),
		ModuleCacheTimes: make(map[string]time.Time),
		ChangedModules:  []string{},
		UpToDateModules: []string{},
	}

	assert.NotNil(t, result.UoWResult)
	assert.NotNil(t, result.CachedUoWs)
	assert.NotNil(t, result.CachedModules)
}

// =============================================================================
// AggregateUoWChanges Tests
// =============================================================================

func TestAggregateUoWChanges_AllCached_ModuleCached(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests for two UoWs in same module
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash"
	m1.ExecutedAt = time.Now().Add(-1 * time.Hour)
	f.saveManifest(m1)

	m2 := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	m2.InputHash = "sha256:hash"
	m2.ExecutedAt = time.Now().Add(-2 * time.Hour)
	f.saveManifest(m2)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := AggregateUoWChanges(reader, core.ActionBuild, expectedUoWs, getInputHash)
	require.NoError(t, err)

	// Module should be cached since all UoWs are cached
	assert.True(t, result.CachedModules["core"], "module should be cached when all UoWs cached")
	assert.Contains(t, result.UpToDateModules, "core")
	assert.Empty(t, result.ChangedModules)

	// Both UoWs should be marked cached
	assert.True(t, result.CachedUoWs["build:core:go:go:go"])
	assert.True(t, result.CachedUoWs["build:core:docker:docker:docker"])
}

func TestAggregateUoWChanges_PartialCached_ModuleChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests - only one UoW cached
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash"
	f.saveManifest(m1)

	m2 := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	m2.InputHash = "sha256:old-hash" // Different hash
	f.saveManifest(m2)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil // Docker will differ
	}

	result, err := AggregateUoWChanges(reader, core.ActionBuild, expectedUoWs, getInputHash)
	require.NoError(t, err)

	// Module should be changed because docker UoW changed
	assert.False(t, result.CachedModules["core"], "module should not be cached when some UoWs changed")
	assert.Contains(t, result.ChangedModules, "core")
	assert.Empty(t, result.UpToDateModules)

	// Only go UoW should be cached
	assert.True(t, result.CachedUoWs["build:core:go:go:go"])
	assert.False(t, result.CachedUoWs["build:core:docker:docker:docker"])
}

func TestAggregateUoWChanges_MultipleModules(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// core module: all cached
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash-core"
	f.saveManifest(m1)

	// cli module: not cached (no manifest)
	// api module: cached
	m2 := f.createUoWManifest(core.ActionBuild, "api", "go", "go")
	m2.InputHash = "sha256:hash-api"
	f.saveManifest(m2)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "cli", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "api", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		if id.Module == "core" {
			return "sha256:hash-core", nil
		}
		if id.Module == "api" {
			return "sha256:hash-api", nil
		}
		return "sha256:hash-cli", nil
	}

	result, err := AggregateUoWChanges(reader, core.ActionBuild, expectedUoWs, getInputHash)
	require.NoError(t, err)

	// core and api should be cached, cli should be changed
	assert.True(t, result.CachedModules["core"])
	assert.True(t, result.CachedModules["api"])
	assert.False(t, result.CachedModules["cli"])

	assert.ElementsMatch(t, []string{"core", "api"}, result.UpToDateModules)
	assert.ElementsMatch(t, []string{"cli"}, result.ChangedModules)
}

func TestAggregateUoWChanges_FreshRun(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// No manifests - fresh run
	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := AggregateUoWChanges(reader, core.ActionBuild, expectedUoWs, getInputHash)
	require.NoError(t, err)

	assert.True(t, result.UoWResult.FreshRun)
	assert.Empty(t, result.CachedModules)
	assert.Empty(t, result.UpToDateModules)
	// On fresh run, ChangedModules may or may not be populated depending on implementation
}

func TestAggregateUoWChanges_EmptyInput(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	result, err := AggregateUoWChanges(reader, core.ActionBuild, []workunit.UnitID{}, nil)
	require.NoError(t, err)

	assert.Empty(t, result.CachedModules)
	assert.Empty(t, result.CachedUoWs)
	assert.Empty(t, result.ChangedModules)
	assert.Empty(t, result.UpToDateModules)
}

func TestAggregateUoWChanges_RecordsCacheTimes(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests with specific times
	time1 := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	time2 := time.Now().Add(-1 * time.Hour).Truncate(time.Second)

	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash"
	m1.ExecutedAt = time1
	f.saveManifest(m1)

	m2 := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	m2.InputHash = "sha256:hash"
	m2.ExecutedAt = time2
	f.saveManifest(m2)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := AggregateUoWChanges(reader, core.ActionBuild, expectedUoWs, getInputHash)
	require.NoError(t, err)

	// UoW cache times should be recorded
	assert.Equal(t, time1, result.UoWCacheTimes["build:core:go:go:go"])
	assert.Equal(t, time2, result.UoWCacheTimes["build:core:docker:docker:docker"])

	// Module cache time should be the earliest (oldest) of its UoWs
	assert.Equal(t, time1, result.ModuleCacheTimes["core"])
}

func TestAggregateUoWChanges_PreservesUoWResult(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash"
	f.saveManifest(m1)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"}, // No manifest
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := AggregateUoWChanges(reader, core.ActionBuild, expectedUoWs, getInputHash)
	require.NoError(t, err)

	// UoWResult should have the underlying detection result
	assert.NotNil(t, result.UoWResult)
	assert.Len(t, result.UoWResult.Changed, 1)
	assert.Len(t, result.UoWResult.UpToDate, 1)
	assert.Contains(t, result.UoWResult.ChangeReasons, "build:core:docker:docker:docker")
}
