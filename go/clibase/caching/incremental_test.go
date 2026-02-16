package caching

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/cmdframework"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test helpers ---

// makeUnitID creates a UnitID for testing with the given action, module,
// component type, component name, and tool.
func makeUnitID(action core.ActionType, module, compType, compName, tool string) workunit.UnitID {
	return workunit.UnitID{
		Action:        action,
		Module:        module,
		ComponentType: compType,
		ComponentName: compName,
		Tool:          tool,
	}
}

// makeUnitSpec creates a UnitSpec wrapping a UnitID.
func makeUnitSpec(id workunit.UnitID) workunit.UnitSpec {
	return workunit.UnitSpec{ID: id}
}

// writeSourceFile creates a file in the workspace directory for glob matching.
func writeSourceFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	absPath := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))
	require.NoError(t, os.WriteFile(absPath, []byte(content), 0o644))
}

// saveManifest creates a UoW manifest on disk at the expected path.
func saveManifest(t *testing.T, workspaceRoot string, m *coreoutput.UoWManifest) {
	t.Helper()
	require.NoError(t, m.Save(workspaceRoot))
}

// createModuleWithPatterns creates a ModuleContract with the given source patterns
// and registers it in the registry.
func createModuleWithPatterns(t *testing.T, registry *modules.Registry, moniker, root string, sourcePatterns []string) {
	t.Helper()
	base := domain.BaseContract{
		Moniker: moniker,
		Components: config.ModuleComponents{
			"go": &config.ComponentEntry{
				Root: root,
				Patterns: &config.ComponentPatterns{
					Source: sourcePatterns,
				},
			},
		},
	}
	mc := modules.NewModuleContract(base, "")
	require.NoError(t, registry.Add(mc))
}

// newTestContext creates a minimal ExecutionContext for testing.
// If initSummary is true, attaches an InitSummary.
func newTestContext(workspaceRoot string, registry *modules.Registry, withInitSummary bool) *cmdframework.ExecutionContext {
	ctx := &cmdframework.ExecutionContext{
		Config:         &cmdframework.CommandConfig{},
		WorkspaceRoot:  workspaceRoot,
		ModuleRegistry: registry,
	}
	if withInitSummary {
		ctx.InitSummary = &initsummary.Summary{}
	}
	return ctx
}

// --- tests ---

func TestDetectIncrementalChanges_EmptySpecs(t *testing.T) {
	tmpDir := t.TempDir()
	registry := modules.NewRegistry("0.1.0", tmpDir)
	ctx := newTestContext(tmpDir, registry, false)

	result := DetectIncrementalChanges(ctx, core.ActionTest, nil, "test")

	assert.Nil(t, result, "nil specs should return nil result")
}

func TestDetectIncrementalChanges_EmptySpecSlice(t *testing.T) {
	tmpDir := t.TempDir()
	registry := modules.NewRegistry("0.1.0", tmpDir)
	ctx := newTestContext(tmpDir, registry, false)

	result := DetectIncrementalChanges(ctx, core.ActionTest, []workunit.UnitSpec{}, "test")

	assert.Nil(t, result, "empty spec slice should return nil result")
}

func TestDetectIncrementalChanges_FreshRun_NoManifests(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files that match the glob patterns
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, true)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.True(t, result.FreshRun, "should be a fresh run when no manifests exist")
	assert.NotEmpty(t, result.ModuleInputHashes, "should have pre-computed module input hashes")
	assert.Contains(t, result.ModuleInputHashes, "mod-a")

	// InitSummary should reflect fresh run
	require.NotNil(t, ctx.InitSummary.Incremental)
	assert.True(t, ctx.InitSummary.Incremental.FreshBuild)
	assert.True(t, ctx.InitSummary.Incremental.Enabled)
}

func TestDetectIncrementalChanges_AllCached(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, true)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	// First pass: get the hash so we can create a matching manifest
	firstResult := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	require.NotNil(t, firstResult)
	require.True(t, firstResult.FreshRun)
	inputHash := firstResult.ModuleInputHashes["mod-a"]
	require.NotEmpty(t, inputHash, "first pass should compute an input hash")

	// Write a manifest with matching input hash
	manifest := &coreoutput.UoWManifest{
		Action:     core.ActionTest,
		Module:     "mod-a",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   0,
		InputHash:  inputHash,
		ExecutedAt: time.Now().Add(-1 * time.Hour),
	}
	saveManifest(t, tmpDir, manifest)

	// Reset InitSummary for second pass
	ctx.InitSummary = &initsummary.Summary{}

	// Second pass: should detect cache hit
	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.False(t, result.FreshRun)
	assert.Contains(t, result.CachedUoWs, id.Longname())
	assert.True(t, result.CachedUoWs[id.Longname()])
	assert.Contains(t, result.CachedModules, "mod-a")
	assert.True(t, result.CachedModules["mod-a"])
	assert.Contains(t, result.UpToDateModules, "mod-a")
	assert.Empty(t, result.ChangedModules)

	// InitSummary should reflect incremental with up-to-date modules
	require.NotNil(t, ctx.InitSummary.Incremental)
	assert.False(t, ctx.InitSummary.Incremental.FreshBuild)
	assert.Contains(t, ctx.InitSummary.Incremental.UpToDate, "mod-a")
}

func TestDetectIncrementalChanges_SourceChanged(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main // v1")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, true)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	// Write a manifest with an OLD (non-matching) input hash
	manifest := &coreoutput.UoWManifest{
		Action:     core.ActionTest,
		Module:     "mod-a",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   0,
		InputHash:  "stale-hash-that-wont-match",
		ExecutedAt: time.Now().Add(-1 * time.Hour),
	}
	saveManifest(t, tmpDir, manifest)

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.False(t, result.FreshRun)
	assert.NotContains(t, result.CachedUoWs, id.Longname())
	assert.Contains(t, result.ChangedModules, "mod-a")
	assert.Empty(t, result.UpToDateModules)

	// InitSummary should list mod-a as changed
	require.NotNil(t, ctx.InitSummary.Incremental)
	assert.Contains(t, ctx.InitSummary.Incremental.Changed, "mod-a")
}

func TestDetectIncrementalChanges_PreviousFailure(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	// First, get the correct hash
	firstResult := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	require.NotNil(t, firstResult)
	inputHash := firstResult.ModuleInputHashes["mod-a"]

	// Write a manifest with ExitCode=1 (failure) but matching hash
	manifest := &coreoutput.UoWManifest{
		Action:     core.ActionTest,
		Module:     "mod-a",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   1, // Previous failure
		InputHash:  inputHash,
		ExecutedAt: time.Now().Add(-1 * time.Hour),
	}
	saveManifest(t, tmpDir, manifest)

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.False(t, result.FreshRun)
	// Previous failure should cause re-execution
	assert.Contains(t, result.ChangedModules, "mod-a")
	assert.Empty(t, result.UpToDateModules)
}

func TestDetectIncrementalChanges_MultipleModules_MixedState(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files for two modules
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main // mod-a")
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-b", "main.go"), "package main // mod-b")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})
	createModuleWithPatterns(t, registry, "mod-b", "src/mod-b", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, true)

	idA := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	idB := makeUnitID(core.ActionTest, "mod-b", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(idA), makeUnitSpec(idB)}

	// First pass to get hashes
	firstResult := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	require.NotNil(t, firstResult)
	hashA := firstResult.ModuleInputHashes["mod-a"]
	require.NotEmpty(t, hashA)

	// Only create a matching manifest for mod-a (cached); mod-b gets stale hash
	manifestA := &coreoutput.UoWManifest{
		Action:     core.ActionTest,
		Module:     "mod-a",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   0,
		InputHash:  hashA,
		ExecutedAt: time.Now().Add(-1 * time.Hour),
	}
	saveManifest(t, tmpDir, manifestA)

	manifestB := &coreoutput.UoWManifest{
		Action:     core.ActionTest,
		Module:     "mod-b",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   0,
		InputHash:  "stale-hash-for-mod-b",
		ExecutedAt: time.Now().Add(-1 * time.Hour),
	}
	saveManifest(t, tmpDir, manifestB)

	// Reset summary
	ctx.InitSummary = &initsummary.Summary{}

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.False(t, result.FreshRun)

	// mod-a should be cached, mod-b should be changed
	assert.Contains(t, result.CachedModules, "mod-a")
	assert.Contains(t, result.UpToDateModules, "mod-a")
	assert.Contains(t, result.ChangedModules, "mod-b")
	assert.NotContains(t, result.CachedModules, "mod-b")

	// Both modules should have input hashes
	assert.Contains(t, result.ModuleInputHashes, "mod-a")
	assert.Contains(t, result.ModuleInputHashes, "mod-b")

	// InitSummary
	require.NotNil(t, ctx.InitSummary.Incremental)
	assert.Contains(t, ctx.InitSummary.Incremental.UpToDate, "mod-a")
	assert.Contains(t, ctx.InitSummary.Incremental.Changed, "mod-b")
}

func TestDetectIncrementalChanges_ModuleNotInRegistry(t *testing.T) {
	tmpDir := t.TempDir()

	// Do NOT register the module in the registry
	registry := modules.NewRegistry("0.1.0", tmpDir)

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "unknown-module", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	// Should still return a result (fresh run since no manifests exist)
	require.NotNil(t, result)
	assert.True(t, result.FreshRun)

	// No input hash should be computed for a module not in registry
	assert.Empty(t, result.ModuleInputHashes)
}

func TestDetectIncrementalChanges_NoGlobMatches(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure but no matching files
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "src", "mod-a"), 0o755))

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	// Should return a result; fresh run since no manifests
	require.NotNil(t, result)
	assert.True(t, result.FreshRun)

	// No files matched, so no input hash (hash.Files on empty list may
	// return an error or empty hash)
	// The important thing is it does not panic
}

func TestDetectIncrementalChanges_NilInitSummary(t *testing.T) {
	tmpDir := t.TempDir()

	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	// Explicitly nil InitSummary
	ctx := newTestContext(tmpDir, registry, false)
	assert.Nil(t, ctx.InitSummary)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	// Should not panic when InitSummary is nil
	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.True(t, result.FreshRun)
}

func TestDetectIncrementalChanges_ModuleInputHashConsistency(t *testing.T) {
	tmpDir := t.TempDir()

	// Create deterministic source files
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main\nfunc main() {}\n")
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "util.go"), "package main\nfunc util() {}\n")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	// Run detection twice - hashes should be identical
	result1 := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	result2 := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result1)
	require.NotNil(t, result2)
	assert.Equal(t, result1.ModuleInputHashes["mod-a"], result2.ModuleInputHashes["mod-a"],
		"same source files should produce identical input hashes across runs")
}

func TestDetectIncrementalChanges_HashChangesWhenSourceChanges(t *testing.T) {
	tmpDir := t.TempDir()

	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main // v1")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	result1 := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	hash1 := result1.ModuleInputHashes["mod-a"]

	// Modify the source file
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main // v2 - modified")

	result2 := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	hash2 := result2.ModuleInputHashes["mod-a"]

	assert.NotEqual(t, hash1, hash2,
		"input hash should change when source file content changes")
}

func TestDetectIncrementalChanges_DeduplicatesModuleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source files for one module
	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	// Two UoWs for the same module (e.g., unit tests and integration tests)
	id1 := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	id1.Extra = map[string]string{"testname": "unit"}
	id2 := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	id2.Extra = map[string]string{"testname": "integration"}
	specs := []workunit.UnitSpec{makeUnitSpec(id1), makeUnitSpec(id2)}

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	// Module input hash should exist exactly once (deduplicated)
	assert.Contains(t, result.ModuleInputHashes, "mod-a")
	assert.NotEmpty(t, result.ModuleInputHashes["mod-a"])
}

func TestDetectIncrementalChanges_MultipleUoWsPerModule_AllCached(t *testing.T) {
	tmpDir := t.TempDir()

	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, true)

	// Two UoWs for the same module
	id1 := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	id1.Extra = map[string]string{"testname": "unit"}
	id2 := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	id2.Extra = map[string]string{"testname": "integration"}
	specs := []workunit.UnitSpec{makeUnitSpec(id1), makeUnitSpec(id2)}

	// Get hash for creating manifests
	firstResult := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	require.NotNil(t, firstResult)
	inputHash := firstResult.ModuleInputHashes["mod-a"]
	require.NotEmpty(t, inputHash)

	// Create manifests for both UoWs
	for _, id := range []workunit.UnitID{id1, id2} {
		m := &coreoutput.UoWManifest{
			Action:     id.Action,
			Module:     id.Module,
			Component:  id.ComponentName,
			Tool:       id.Tool,
			Extra:      id.Extra,
			ExitCode:   0,
			InputHash:  inputHash,
			ExecutedAt: time.Now().Add(-1 * time.Hour),
		}
		saveManifest(t, tmpDir, m)
	}

	ctx.InitSummary = &initsummary.Summary{}
	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.False(t, result.FreshRun)

	// Both UoWs should be cached
	assert.True(t, result.CachedUoWs[id1.Longname()], "unit testset UoW should be cached")
	assert.True(t, result.CachedUoWs[id2.Longname()], "integration testset UoW should be cached")

	// Module-level: all UoWs cached means module is cached
	assert.Contains(t, result.CachedModules, "mod-a")
	assert.Contains(t, result.UpToDateModules, "mod-a")
	assert.Empty(t, result.ChangedModules)
}

func TestDetectIncrementalChanges_MultipleUoWsPerModule_PartiallyCached(t *testing.T) {
	tmpDir := t.TempDir()

	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, true)

	// Two UoWs for the same module
	id1 := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	id1.Extra = map[string]string{"testname": "unit"}
	id2 := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	id2.Extra = map[string]string{"testname": "integration"}
	specs := []workunit.UnitSpec{makeUnitSpec(id1), makeUnitSpec(id2)}

	// Get hash
	firstResult := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	require.NotNil(t, firstResult)
	inputHash := firstResult.ModuleInputHashes["mod-a"]

	// Only create a manifest for id1 (unit tests); id2 has no manifest
	m := &coreoutput.UoWManifest{
		Action:     id1.Action,
		Module:     id1.Module,
		Component:  id1.ComponentName,
		Tool:       id1.Tool,
		Extra:      id1.Extra,
		ExitCode:   0,
		InputHash:  inputHash,
		ExecutedAt: time.Now().Add(-1 * time.Hour),
	}
	saveManifest(t, tmpDir, m)

	// id2 gets no manifest -> changed

	ctx.InitSummary = &initsummary.Summary{}
	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.False(t, result.FreshRun)

	// id1 should be cached, id2 should not
	assert.True(t, result.CachedUoWs[id1.Longname()], "unit testset should be cached")
	assert.False(t, result.CachedUoWs[id2.Longname()], "integration testset should not be cached")

	// Module not fully cached (only 1/2 UoWs cached)
	assert.NotContains(t, result.CachedModules, "mod-a")
	assert.Contains(t, result.ChangedModules, "mod-a")
}

func TestDetectIncrementalChanges_SetsChangeDetectionTiming(t *testing.T) {
	tmpDir := t.TempDir()

	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	// SetChangeDetectionTiming is called internally via defer.
	// We verify it doesn't panic and the function completes successfully.
	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	require.NotNil(t, result)
}

func TestDetectIncrementalChanges_DifferentActionTypes(t *testing.T) {
	tests := []struct {
		name   string
		action core.ActionType
	}{
		{name: "test", action: core.ActionTest},
		{name: "lint", action: core.ActionLint},
		{name: "scan", action: core.ActionScan},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

			registry := modules.NewRegistry("0.1.0", tmpDir)
			createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

			ctx := newTestContext(tmpDir, registry, false)

			id := makeUnitID(tt.action, "mod-a", "go", "go", "tool")
			specs := []workunit.UnitSpec{makeUnitSpec(id)}

			result := DetectIncrementalChanges(ctx, tt.action, specs, tt.name)

			require.NotNil(t, result)
			assert.True(t, result.FreshRun, "fresh run expected for action type %s", tt.action)
			assert.Contains(t, result.ModuleInputHashes, "mod-a")
		})
	}
}

func TestDetectIncrementalChanges_CacheTimesPopulated(t *testing.T) {
	tmpDir := t.TempDir()

	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	// Get hash for manifest
	firstResult := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")
	inputHash := firstResult.ModuleInputHashes["mod-a"]

	executedAt := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	manifest := &coreoutput.UoWManifest{
		Action:     core.ActionTest,
		Module:     "mod-a",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   0,
		InputHash:  inputHash,
		ExecutedAt: executedAt,
	}
	saveManifest(t, tmpDir, manifest)

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.False(t, result.FreshRun)

	// UoW cache time should be populated
	uowTime, ok := result.UoWCacheTimes[id.Longname()]
	assert.True(t, ok, "UoW cache time should be present")
	assert.False(t, uowTime.IsZero(), "UoW cache time should not be zero")

	// Module cache time should be populated
	modTime, ok := result.ModuleCacheTimes["mod-a"]
	assert.True(t, ok, "module cache time should be present")
	assert.False(t, modTime.IsZero(), "module cache time should not be zero")
}

func TestDetectIncrementalChanges_IncrementalResultFields(t *testing.T) {
	// Verify that IncrementalResult has the expected zero values
	result := &IncrementalResult{}

	assert.Nil(t, result.CachedUoWs)
	assert.Nil(t, result.UoWCacheTimes)
	assert.Nil(t, result.CachedModules)
	assert.Nil(t, result.ModuleCacheTimes)
	assert.Nil(t, result.ModuleInputHashes)
	assert.Nil(t, result.ChangedModules)
	assert.Nil(t, result.UpToDateModules)
	assert.False(t, result.FreshRun)
}

func TestDetectIncrementalChanges_FreshRunResult_Structure(t *testing.T) {
	tmpDir := t.TempDir()

	writeSourceFile(t, tmpDir, filepath.Join("src", "mod-a", "main.go"), "package main")

	registry := modules.NewRegistry("0.1.0", tmpDir)
	createModuleWithPatterns(t, registry, "mod-a", "src/mod-a", []string{"*.go"})

	ctx := newTestContext(tmpDir, registry, false)

	id := makeUnitID(core.ActionTest, "mod-a", "go", "go", "gotest")
	specs := []workunit.UnitSpec{makeUnitSpec(id)}

	result := DetectIncrementalChanges(ctx, core.ActionTest, specs, "test")

	require.NotNil(t, result)
	assert.True(t, result.FreshRun)

	// Fresh run result should have ModuleInputHashes but no cached state
	assert.NotEmpty(t, result.ModuleInputHashes)
	assert.Nil(t, result.CachedUoWs, "fresh run should not have CachedUoWs")
	assert.Nil(t, result.UoWCacheTimes, "fresh run should not have UoWCacheTimes")
	assert.Nil(t, result.CachedModules, "fresh run should not have CachedModules")
	assert.Nil(t, result.ModuleCacheTimes, "fresh run should not have ModuleCacheTimes")
	assert.Nil(t, result.ChangedModules, "fresh run should not populate ChangedModules")
	assert.Nil(t, result.UpToDateModules, "fresh run should not populate UpToDateModules")
}
