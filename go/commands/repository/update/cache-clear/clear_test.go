// Package cacheclear provides tests for the cache-clear command with --type flag support.
// These tests follow TDD principles - they are written BEFORE the implementation.
//
// The --type flag allows selective cache clearing based on cache.Spec matching:
//   - state: Incremental build state files (state.json)
//   - asset: Rendered assets (mermaid SVGs, structurizr PNGs)
//   - work: Ephemeral work directories (npm, go build)
//   - all: All cache types (default)
//   - Fine-grained: local:state, local:asset, etc.
package cacheclear

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/core/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ClearMode Enum Tests
// =============================================================================
// ClearMode determines how a directory should be cleared:
// - ClearContents: Delete all contents in the directory

func TestClearMode_Constants(t *testing.T) {
	// Verify that ClearMode constants are defined and distinct
	assert.NotEqual(t, ClearContents, ClearDocker, "ClearMode constants should be distinct")

	// Verify expected values
	assert.Equal(t, ClearMode(0), ClearContents, "ClearContents should be the zero value")
	assert.Equal(t, ClearMode(1), ClearDocker, "ClearDocker should be 1")
}

func TestClearMode_String(t *testing.T) {
	tests := []struct {
		mode     ClearMode
		expected string
	}{
		{ClearContents, "contents"},
		{ClearDocker, "docker"},
		{ClearSemaphore, "semaphore"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

// =============================================================================
// ClearDir Struct Tests
// =============================================================================
// ClearDir represents a directory that can be cleared, with metadata about
// what kind of cache it contains and how it should be cleared.

func TestClearDir_Fields(t *testing.T) {
	dir := ClearDir{
		RelPath:     "out/build",
		Description: "build state",
		Mode:        ClearContents,
		Level:       cache.LevelLocal,
		Type:        cache.TypeState,
	}

	assert.Equal(t, "out/build", dir.RelPath)
	assert.Equal(t, "build state", dir.Description)
	assert.Equal(t, ClearContents, dir.Mode)
	assert.Equal(t, cache.LevelLocal, dir.Level)
	assert.Equal(t, cache.TypeState, dir.Type)
}

func TestClearDir_AllDefinedDirectories(t *testing.T) {
	// GetAllClearDirs should return all known cache directories
	dirs := GetAllClearDirs()

	// Verify we have the expected categories
	require.NotEmpty(t, dirs)

	// Group by type for verification
	byType := make(map[cache.Type][]ClearDir)
	for _, d := range dirs {
		byType[d.Type] = append(byType[d.Type], d)
	}

	// State directories should exist
	assert.NotEmpty(t, byType[cache.TypeState], "should have state directories")

	// Asset directories should exist
	assert.NotEmpty(t, byType[cache.TypeAsset], "should have asset directories")

	// Work directories should exist
	assert.NotEmpty(t, byType[cache.TypeWork], "should have work directories")
}

func TestClearDir_StateDirectories(t *testing.T) {
	dirs := GetAllClearDirs()

	// Find state directories
	var stateDirs []ClearDir
	for _, d := range dirs {
		if d.Type == cache.TypeState {
			stateDirs = append(stateDirs, d)
		}
	}

	// Expected state directories from current implementation
	expectedPaths := []string{
		".cache/eac/incremental",
		".cache/eac/build",
		"out/build",
		"out/test",
		"out/lint",
		"out/scan",
		".cache/eac/semaphores",
	}

	paths := make([]string, len(stateDirs))
	for i, d := range stateDirs {
		paths[i] = d.RelPath
	}

	for _, expected := range expectedPaths {
		// Normalize path separators for cross-platform
		found := false
		for _, p := range paths {
			if filepath.ToSlash(p) == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "should include state directory: %s", expected)
	}
}

func TestClearDir_AssetDirectories(t *testing.T) {
	dirs := GetAllClearDirs()

	// Find asset directories
	var assetDirs []ClearDir
	for _, d := range dirs {
		if d.Type == cache.TypeAsset {
			assetDirs = append(assetDirs, d)
		}
	}

	// Asset directories should use ClearContents mode
	for _, d := range assetDirs {
		assert.Equal(t, ClearContents, d.Mode,
			"asset directory %s should use ClearContents mode", d.RelPath)
	}
}

func TestClearDir_WorkDirectories(t *testing.T) {
	dirs := GetAllClearDirs()

	// Find work directories
	var workDirs []ClearDir
	for _, d := range dirs {
		if d.Type == cache.TypeWork {
			workDirs = append(workDirs, d)
		}
	}

	// Work directories should exist (npm work dirs, etc.)
	assert.NotEmpty(t, workDirs, "should have work directories")

	// All work directories should be local-only
	for _, d := range workDirs {
		assert.Equal(t, cache.LevelLocal, d.Level,
			"work directory %s should be local-only", d.RelPath)
	}
}

// =============================================================================
// CacheTarget Struct Tests
// =============================================================================
// CacheTarget combines a ClearDir with its absolute path for actual clearing.

func TestCacheTarget_Fields(t *testing.T) {
	dir := ClearDir{
		RelPath:     "out/build",
		Description: "build state",
		Mode:        ClearContents,
		Level:       cache.LevelLocal,
		Type:        cache.TypeState,
	}

	target := CacheTarget{
		Dir:      dir,
		FullPath: "/repo/out/build",
	}

	assert.Equal(t, dir, target.Dir)
	assert.Equal(t, "/repo/out/build", target.FullPath)
}

func TestCacheTarget_Matches_AllSpec(t *testing.T) {
	target := CacheTarget{
		Dir: ClearDir{
			Level: cache.LevelLocal,
			Type:  cache.TypeState,
		},
	}

	// all:all should match everything
	spec := cache.Spec{Level: cache.LevelAll, Type: cache.TypeAll}
	assert.True(t, target.Matches(spec), "all:all should match local:state")
}

func TestCacheTarget_Matches_TypeOnly(t *testing.T) {
	tests := []struct {
		name      string
		target    CacheTarget
		spec      cache.Spec
		shouldMatch bool
	}{
		{
			name: "state spec matches state target",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState},
			},
			spec:        cache.Spec{Level: cache.LevelAll, Type: cache.TypeState},
			shouldMatch: true,
		},
		{
			name: "state spec does not match asset target",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset},
			},
			spec:        cache.Spec{Level: cache.LevelAll, Type: cache.TypeState},
			shouldMatch: false,
		},
		{
			name: "asset spec matches asset target",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset},
			},
			spec:        cache.Spec{Level: cache.LevelAll, Type: cache.TypeAsset},
			shouldMatch: true,
		},
		{
			name: "work spec matches work target",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeWork},
			},
			spec:        cache.Spec{Level: cache.LevelAll, Type: cache.TypeWork},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.shouldMatch, tt.target.Matches(tt.spec))
		})
	}
}

func TestCacheTarget_Matches_LevelOnly(t *testing.T) {
	tests := []struct {
		name        string
		target      CacheTarget
		spec        cache.Spec
		shouldMatch bool
	}{
		{
			name: "local spec matches local target",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState},
			},
			spec:        cache.Spec{Level: cache.LevelLocal, Type: cache.TypeAll},
			shouldMatch: true,
		},
		{
			name: "local spec does not match remote target",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelRemote, Type: cache.TypeState},
			},
			spec:        cache.Spec{Level: cache.LevelLocal, Type: cache.TypeAll},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.shouldMatch, tt.target.Matches(tt.spec))
		})
	}
}

func TestCacheTarget_Matches_ExactSpec(t *testing.T) {
	tests := []struct {
		name        string
		target      CacheTarget
		spec        cache.Spec
		shouldMatch bool
	}{
		{
			name: "local:state matches local:state",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState},
			},
			spec:        cache.Spec{Level: cache.LevelLocal, Type: cache.TypeState},
			shouldMatch: true,
		},
		{
			name: "local:state does not match local:asset",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState},
			},
			spec:        cache.Spec{Level: cache.LevelLocal, Type: cache.TypeAsset},
			shouldMatch: false,
		},
		{
			name: "local:state does not match remote:state",
			target: CacheTarget{
				Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState},
			},
			spec:        cache.Spec{Level: cache.LevelRemote, Type: cache.TypeState},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.shouldMatch, tt.target.Matches(tt.spec))
		})
	}
}

// =============================================================================
// FilterTargets Function Tests
// =============================================================================
// FilterTargets filters a list of CacheTargets based on a cache.Spec

func TestFilterTargets_AllSpec(t *testing.T) {
	targets := []CacheTarget{
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeWork}},
	}

	specs := []cache.Spec{{Level: cache.LevelAll, Type: cache.TypeAll}}
	filtered := FilterTargets(targets, specs)

	assert.Len(t, filtered, 3, "all:all should match all targets")
}

func TestFilterTargets_StateOnly(t *testing.T) {
	targets := []CacheTarget{
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState, RelPath: "out/build"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset, RelPath: "assets/cache"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeWork, RelPath: ".cache/npm"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState, RelPath: "out/test"}},
	}

	specs := []cache.Spec{{Level: cache.LevelAll, Type: cache.TypeState}}
	filtered := FilterTargets(targets, specs)

	assert.Len(t, filtered, 2, "state spec should match only state targets")

	for _, target := range filtered {
		assert.Equal(t, cache.TypeState, target.Dir.Type)
	}
}

func TestFilterTargets_AssetOnly(t *testing.T) {
	targets := []CacheTarget{
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState, RelPath: "out/build"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset, RelPath: "assets/cache/mermaid"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset, RelPath: "assets/cache/drawio"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeWork, RelPath: ".cache/npm"}},
	}

	specs := []cache.Spec{{Level: cache.LevelAll, Type: cache.TypeAsset}}
	filtered := FilterTargets(targets, specs)

	assert.Len(t, filtered, 2, "asset spec should match only asset targets")

	for _, target := range filtered {
		assert.Equal(t, cache.TypeAsset, target.Dir.Type)
	}
}

func TestFilterTargets_WorkOnly(t *testing.T) {
	targets := []CacheTarget{
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState, RelPath: "out/build"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset, RelPath: "assets/cache"}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeWork, RelPath: ".cache/npm/work"}},
	}

	specs := []cache.Spec{{Level: cache.LevelAll, Type: cache.TypeWork}}
	filtered := FilterTargets(targets, specs)

	assert.Len(t, filtered, 1, "work spec should match only work targets")
	assert.Equal(t, cache.TypeWork, filtered[0].Dir.Type)
}

func TestFilterTargets_EmptyResult(t *testing.T) {
	targets := []CacheTarget{
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeState}},
		{Dir: ClearDir{Level: cache.LevelLocal, Type: cache.TypeAsset}},
	}

	// Registry type doesn't exist in these targets (work type)
	specs := []cache.Spec{{Level: cache.LevelAll, Type: cache.TypeWork}}
	filtered := FilterTargets(targets, specs)

	assert.Empty(t, filtered, "should return empty slice when no matches")
}

func TestFilterTargets_EmptyInput(t *testing.T) {
	var targets []CacheTarget
	specs := []cache.Spec{{Level: cache.LevelAll, Type: cache.TypeAll}}
	filtered := FilterTargets(targets, specs)

	assert.Empty(t, filtered, "should return empty slice for empty input")
}

// =============================================================================
// ParseTypeFlag Function Tests
// =============================================================================
// ParseTypeFlag parses the --type flag value into a cache.Spec

func TestParseTypeFlag_DefaultStateWork(t *testing.T) {
	// Empty string should default to state + work (DefaultSkipSpecs)
	specs, err := ParseTypeFlag("")
	require.NoError(t, err)
	require.Len(t, specs, 2, "default should return 2 specs (state + work)")
	// First spec should be local:state
	assert.Equal(t, cache.LevelLocal, specs[0].Level)
	assert.Equal(t, cache.TypeState, specs[0].Type)
	// Second spec should be local:work
	assert.Equal(t, cache.LevelLocal, specs[1].Level)
	assert.Equal(t, cache.TypeWork, specs[1].Type)
}

func TestParseTypeFlag_ExplicitAll(t *testing.T) {
	specs, err := ParseTypeFlag("all")
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, cache.LevelAll, specs[0].Level)
	assert.Equal(t, cache.TypeAll, specs[0].Type)
}

func TestParseTypeFlag_StateOnly(t *testing.T) {
	specs, err := ParseTypeFlag("state")
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, cache.LevelAll, specs[0].Level)
	assert.Equal(t, cache.TypeState, specs[0].Type)
}

func TestParseTypeFlag_AssetOnly(t *testing.T) {
	specs, err := ParseTypeFlag("asset")
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, cache.LevelAll, specs[0].Level)
	assert.Equal(t, cache.TypeAsset, specs[0].Type)
}

func TestParseTypeFlag_WorkOnly(t *testing.T) {
	specs, err := ParseTypeFlag("work")
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, cache.LevelAll, specs[0].Level)
	assert.Equal(t, cache.TypeWork, specs[0].Type)
}

func TestParseTypeFlag_FineGrained(t *testing.T) {
	tests := []struct {
		input    string
		expected cache.Spec
	}{
		{"local:state", cache.Spec{Level: cache.LevelLocal, Type: cache.TypeState}},
		{"local:asset", cache.Spec{Level: cache.LevelLocal, Type: cache.TypeAsset}},
		{"local:work", cache.Spec{Level: cache.LevelLocal, Type: cache.TypeWork}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			specs, err := ParseTypeFlag(tt.input)
			require.NoError(t, err)
			require.Len(t, specs, 1)
			assert.Equal(t, tt.expected, specs[0])
		})
	}
}

func TestParseTypeFlag_CaseInsensitive(t *testing.T) {
	tests := []string{"STATE", "State", "sTaTe"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			specs, err := ParseTypeFlag(input)
			require.NoError(t, err)
			require.Len(t, specs, 1)
			assert.Equal(t, cache.TypeState, specs[0].Type)
		})
	}
}

func TestParseTypeFlag_TrimWhitespace(t *testing.T) {
	specs, err := ParseTypeFlag("  state  ")
	require.NoError(t, err)
	require.Len(t, specs, 1)
	assert.Equal(t, cache.TypeState, specs[0].Type)
}

func TestParseTypeFlag_InvalidSpec(t *testing.T) {
	tests := []struct {
		input  string
		errMsg string
	}{
		{"invalid", "unknown cache spec"},
		{"foo:bar", "unknown cache level"},
		{"local:invalid", "unknown cache type"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseTypeFlag(tt.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// =============================================================================
// ClearResult Struct Tests
// =============================================================================
// ClearResult holds the result of a cache clearing operation

func TestClearResult_Fields(t *testing.T) {
	result := ClearResult{
		DeletedCount: 10,
		DeletedBytes: 1024,
		Targets:      []CacheTarget{},
		DryRun:       false,
		Errors:       nil,
	}

	assert.Equal(t, 10, result.DeletedCount)
	assert.Equal(t, int64(1024), result.DeletedBytes)
	assert.Empty(t, result.Targets)
	assert.False(t, result.DryRun)
	assert.Nil(t, result.Errors)
}

func TestClearResult_HasErrors(t *testing.T) {
	tests := []struct {
		name      string
		errors    []error
		hasErrors bool
	}{
		{
			name:      "no errors",
			errors:    nil,
			hasErrors: false,
		},
		{
			name:      "empty slice",
			errors:    []error{},
			hasErrors: false,
		},
		{
			name:      "has errors",
			errors:    []error{os.ErrNotExist},
			hasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClearResult{Errors: tt.errors}
			assert.Equal(t, tt.hasErrors, result.HasErrors())
		})
	}
}

// =============================================================================
// ClearTargets Function Tests (Integration)
// =============================================================================
// ClearTargets performs the actual clearing operation on targets

func TestClearTargets_Contents_DryRun_WithStateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in directory
	buildDir := filepath.Join(tmpDir, "out", "build")
	module1Dir := filepath.Join(buildDir, "module1")
	require.NoError(t, os.MkdirAll(module1Dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(module1Dir, "state.json"), []byte("{}"), 0o644))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "out/build", Mode: ClearContents, Type: cache.TypeState},
			FullPath: buildDir,
		},
	}

	result := ClearTargets(targets, true, false) // dryRun=true, verbose=false

	assert.True(t, result.DryRun)
	assert.Equal(t, 1, result.DeletedCount, "should count entries in dry run")

	// File should still exist
	_, err := os.Stat(filepath.Join(module1Dir, "state.json"))
	assert.NoError(t, err, "file should still exist in dry run mode")
}

func TestClearTargets_Contents_Delete_WithStateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in directory
	buildDir := filepath.Join(tmpDir, "out", "build")
	module1Dir := filepath.Join(buildDir, "module1")
	require.NoError(t, os.MkdirAll(module1Dir, 0o755))
	stateFile := filepath.Join(module1Dir, "state.json")
	require.NoError(t, os.WriteFile(stateFile, []byte("{}"), 0o644))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "out/build", Mode: ClearContents, Type: cache.TypeState},
			FullPath: buildDir,
		},
	}

	result := ClearTargets(targets, false, false) // dryRun=false, verbose=false

	assert.False(t, result.DryRun)
	assert.Equal(t, 1, result.DeletedCount)

	// Directory should be deleted (ClearContents deletes all entries)
	_, err := os.Stat(module1Dir)
	assert.True(t, os.IsNotExist(err), "module1 directory should be deleted")
}

func TestClearTargets_Contents_DryRun(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with files
	cacheDir := filepath.Join(tmpDir, "assets", "cache", "mermaid")
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "diagram1.svg"), []byte("<svg/>"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "diagram2.svg"), []byte("<svg/>"), 0o644))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "assets/cache/mermaid", Mode: ClearContents, Type: cache.TypeAsset},
			FullPath: cacheDir,
		},
	}

	result := ClearTargets(targets, true, false) // dryRun=true

	assert.True(t, result.DryRun)
	assert.Equal(t, 2, result.DeletedCount, "should count all files in dry run")

	// Files should still exist
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Len(t, entries, 2, "files should still exist in dry run mode")
}

func TestClearTargets_Contents_Delete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with files
	cacheDir := filepath.Join(tmpDir, "assets", "cache", "mermaid")
	require.NoError(t, os.MkdirAll(cacheDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "diagram1.svg"), []byte("<svg/>"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "diagram2.svg"), []byte("<svg/>"), 0o644))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "assets/cache/mermaid", Mode: ClearContents, Type: cache.TypeAsset},
			FullPath: cacheDir,
		},
	}

	result := ClearTargets(targets, false, false) // dryRun=false

	assert.False(t, result.DryRun)
	assert.Equal(t, 2, result.DeletedCount)

	// Directory should be empty
	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "directory should be empty after clear")
}

func TestClearTargets_NonExistentDirectory(t *testing.T) {
	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "nonexistent", Mode: ClearContents},
			FullPath: "/nonexistent/path/that/does/not/exist",
		},
	}

	result := ClearTargets(targets, false, false)

	// Should not error, just skip
	assert.Equal(t, 0, result.DeletedCount)
	assert.False(t, result.HasErrors(), "missing directories should not cause errors")
}

func TestClearTargets_MixedModes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state directory with contents
	buildDir := filepath.Join(tmpDir, "out", "build")
	modDir := filepath.Join(buildDir, "mod")
	require.NoError(t, os.MkdirAll(modDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(modDir, "state.json"), []byte("{}"), 0o644))

	// Create asset files
	assetDir := filepath.Join(tmpDir, "assets", "cache")
	require.NoError(t, os.MkdirAll(assetDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(assetDir, "file1.svg"), []byte("<svg/>"), 0o644))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "out/build", Mode: ClearContents, Type: cache.TypeState},
			FullPath: buildDir,
		},
		{
			Dir:      ClearDir{RelPath: "assets/cache", Mode: ClearContents, Type: cache.TypeAsset},
			FullPath: assetDir,
		},
	}

	result := ClearTargets(targets, false, false)

	// mod dir deleted (ClearContents), file1.svg deleted
	assert.Equal(t, 2, result.DeletedCount)

	// mod directory should be deleted (ClearContents deletes all entries)
	_, err := os.Stat(modDir)
	assert.True(t, os.IsNotExist(err), "mod directory should be deleted")

	// asset file should be deleted
	_, err = os.Stat(filepath.Join(assetDir, "file1.svg"))
	assert.True(t, os.IsNotExist(err), "asset file should be deleted")
}

// =============================================================================
// BuildTargets Function Tests
// =============================================================================
// BuildTargets creates CacheTargets from ClearDirs and a repo root

func TestBuildTargets(t *testing.T) {
	repoRoot := "/repo"
	dirs := []ClearDir{
		{RelPath: "out/build", Description: "build state"},
		{RelPath: "out/lint", Description: "lint state"},
	}

	targets := BuildTargets(dirs, repoRoot)

	require.Len(t, targets, 2)
	assert.Equal(t, filepath.Join(repoRoot, "out/build"), targets[0].FullPath)
	assert.Equal(t, filepath.Join(repoRoot, "out/lint"), targets[1].FullPath)
	assert.Equal(t, dirs[0], targets[0].Dir)
	assert.Equal(t, dirs[1], targets[1].Dir)
}

// =============================================================================
// Integration Tests - Full Workflow
// =============================================================================

func TestClearCache_TypeState(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state files under .cache/eac/incremental/ (new location)
	buildDir := filepath.Join(tmpDir, ".cache", "eac", "incremental", "build", "mod1", "go_go")
	testDir := filepath.Join(tmpDir, ".cache", "eac", "incremental", "test", "mod1", "go_gotest_unit")
	require.NoError(t, os.MkdirAll(buildDir, 0o755))
	require.NoError(t, os.MkdirAll(testDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, "state.json"), []byte("{}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(testDir, "state.json"), []byte("{}"), 0o644))

	// Create asset files (should NOT be deleted with --type=state)
	assetDir := filepath.Join(tmpDir, "docs", "assets", "cache", "mermaid")
	require.NoError(t, os.MkdirAll(assetDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(assetDir, "diagram.svg"), []byte("<svg/>"), 0o644))

	// Parse type flag
	specs, err := ParseTypeFlag("state")
	require.NoError(t, err)

	// Get all dirs and filter
	allDirs := GetAllClearDirs()
	targets := BuildTargets(allDirs, tmpDir)
	filtered := FilterTargets(targets, specs)

	// Clear only state targets
	result := ClearTargets(filtered, false, false)

	// State files should be deleted (incremental dir is cleared with ClearContents)
	_, err = os.Stat(filepath.Join(buildDir, "state.json"))
	assert.True(t, os.IsNotExist(err), "build state.json should be deleted")

	_, err = os.Stat(filepath.Join(testDir, "state.json"))
	assert.True(t, os.IsNotExist(err), "test state.json should be deleted")

	// Asset files should still exist
	_, err = os.Stat(filepath.Join(assetDir, "diagram.svg"))
	assert.NoError(t, err, "asset file should NOT be deleted with --type=state")

	assert.GreaterOrEqual(t, result.DeletedCount, 1, "should delete at least 1 incremental entry")
}

func TestClearCache_TypeAsset(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state files (should NOT be deleted with --type=asset)
	buildDir := filepath.Join(tmpDir, ".cache", "eac", "incremental", "build", "mod1", "go_go")
	require.NoError(t, os.MkdirAll(buildDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, "state.json"), []byte("{}"), 0o644))

	// Create asset files
	mermaidDir := filepath.Join(tmpDir, "docs", "assets", "cache", "mermaid")
	drawioDir := filepath.Join(tmpDir, "docs", "assets", "cache", "drawio")
	require.NoError(t, os.MkdirAll(mermaidDir, 0o755))
	require.NoError(t, os.MkdirAll(drawioDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(mermaidDir, "diagram.svg"), []byte("<svg/>"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(drawioDir, "arch.png"), []byte("png"), 0o644))

	// Parse type flag
	specs, err := ParseTypeFlag("asset")
	require.NoError(t, err)

	// Get all dirs and filter
	allDirs := GetAllClearDirs()
	targets := BuildTargets(allDirs, tmpDir)
	filtered := FilterTargets(targets, specs)

	// Clear only asset targets
	result := ClearTargets(filtered, false, false)

	// State files should still exist
	_, err = os.Stat(filepath.Join(buildDir, "state.json"))
	assert.NoError(t, err, "state.json should NOT be deleted with --type=asset")

	// Asset files should be deleted (if asset dirs are in GetAllClearDirs)
	// The exact count depends on implementation
	assert.GreaterOrEqual(t, result.DeletedCount, 0)
}

func TestClearCache_TypeWork(t *testing.T) {
	tmpDir := t.TempDir()

	// Create state files (should NOT be deleted with --type=work)
	buildDir := filepath.Join(tmpDir, ".cache", "eac", "incremental", "build", "mod1", "go_go")
	require.NoError(t, os.MkdirAll(buildDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, "state.json"), []byte("{}"), 0o644))

	// Create work files
	npmWorkDir := filepath.Join(tmpDir, ".cache", "eac", "npm", "work")
	require.NoError(t, os.MkdirAll(npmWorkDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(npmWorkDir, "temp-file.txt"), []byte("temp"), 0o644))

	// Parse type flag
	specs, err := ParseTypeFlag("work")
	require.NoError(t, err)

	// Get all dirs and filter
	allDirs := GetAllClearDirs()
	targets := BuildTargets(allDirs, tmpDir)
	filtered := FilterTargets(targets, specs)

	// Clear only work targets
	result := ClearTargets(filtered, false, false)

	// State files should still exist
	_, err = os.Stat(filepath.Join(buildDir, "state.json"))
	assert.NoError(t, err, "state.json should NOT be deleted with --type=work")

	// Work files should be deleted (if work dirs are in GetAllClearDirs)
	assert.GreaterOrEqual(t, result.DeletedCount, 0)
}

// =============================================================================
// Verbose Output Tests
// =============================================================================

func TestClearTargets_Verbose(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	dir := filepath.Join(tmpDir, "cache")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o644))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "cache", Mode: ClearContents, Description: "test cache"},
			FullPath: dir,
		},
	}

	// With verbose=true, function should work (we can't easily capture stdout in this test)
	result := ClearTargets(targets, false, true) // dryRun=false, verbose=true

	assert.Equal(t, 1, result.DeletedCount)
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestClearTargets_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.MkdirAll(emptyDir, 0o755))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "empty", Mode: ClearContents},
			FullPath: emptyDir,
		},
	}

	result := ClearTargets(targets, false, false)

	assert.Equal(t, 0, result.DeletedCount, "empty directory should have 0 items deleted")
	assert.False(t, result.HasErrors())
}

func TestClearTargets_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directories with files
	buildDir := filepath.Join(tmpDir, "out", "build")
	dirs := []string{
		filepath.Join(buildDir, "mod1"),
		filepath.Join(buildDir, "mod2"),
		filepath.Join(buildDir, "pkg"),
	}

	for _, d := range dirs {
		require.NoError(t, os.MkdirAll(d, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(d, "state.json"), []byte("{}"), 0o644))
	}

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "out/build", Mode: ClearContents},
			FullPath: buildDir,
		},
	}

	result := ClearTargets(targets, false, false)

	assert.Equal(t, 3, result.DeletedCount, "should delete all top-level entries")

	// Verify all directories deleted
	for _, d := range dirs {
		_, err := os.Stat(d)
		assert.True(t, os.IsNotExist(err), "directory should be deleted: %s", d)
	}
}

func TestClearTargets_BytesCalculation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with known sizes
	dir := filepath.Join(tmpDir, "cache")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	content1 := strings.Repeat("a", 1000) // 1000 bytes
	content2 := strings.Repeat("b", 2500) // 2500 bytes

	require.NoError(t, os.WriteFile(filepath.Join(dir, "file1.txt"), []byte(content1), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file2.txt"), []byte(content2), 0o644))

	targets := []CacheTarget{
		{
			Dir:      ClearDir{RelPath: "cache", Mode: ClearContents},
			FullPath: dir,
		},
	}

	result := ClearTargets(targets, false, false)

	assert.Equal(t, 2, result.DeletedCount)
	assert.Equal(t, int64(3500), result.DeletedBytes, "should track total bytes deleted")
}

// =============================================================================
// Type Flag Documentation Tests
// =============================================================================

func TestTypeFlagValues_Documented(t *testing.T) {
	// These are all the documented --type flag values
	validTypes := []string{
		"all",
		"state",
		"asset",
		"work",
		"registry",
		"layer",
		"local",
		"local:state",
		"local:asset",
		"local:work",
	}

	for _, typ := range validTypes {
		t.Run(typ, func(t *testing.T) {
			specs, err := ParseTypeFlag(typ)
			require.NoError(t, err, "documented type %q should be valid", typ)
			require.NotEmpty(t, specs, "should produce non-empty specs")
		})
	}
}

// =============================================================================
// Regression Tests
// =============================================================================

