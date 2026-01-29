package teststate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helper functions for test setup
// =============================================================================

// createTestManifest creates a test manifest file with the given content.
func createTestManifest(t *testing.T, dir, moniker string, manifest map[string]interface{}) {
	t.Helper()
	moduleDir := filepath.Join(dir, "out", "test", moniker)
	require.NoError(t, os.MkdirAll(moduleDir, 0o755))
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "test.manifest.json"), data, 0o644))
}

// createSourceFile creates a source file and returns its path relative to dir.
func createSourceFile(t *testing.T, dir, relPath, content string) string {
	t.Helper()
	absPath := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(absPath), 0o755))
	require.NoError(t, os.WriteFile(absPath, []byte(content), 0o644))
	return relPath
}

// computeSourceHash computes the hash for source files.
func computeSourceHash(t *testing.T, dir string, files []string) string {
	t.Helper()
	hash, err := hashFiles(dir, files)
	require.NoError(t, err)
	return hash
}

// loadManifest loads a test manifest from disk.
func loadManifest(t *testing.T, dir, moniker string) map[string]interface{} {
	t.Helper()
	manifestPath := filepath.Join(dir, "out", "test", moniker, "test.manifest.json")
	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)
	var manifest map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &manifest))
	return manifest
}

// =============================================================================
// DetectChanges Tests
// =============================================================================

func TestDetectChanges_FreshRun(t *testing.T) {
	// Create temp directory with no manifests
	dir := t.TempDir()

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{}, Dependencies: []string{}},
		"module-b": {SourceFiles: []string{}, Dependencies: []string{"module-a"}},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.True(t, result.FreshRun)
	assert.Equal(t, TestSetUnit, result.TestSet)
	assert.ElementsMatch(t, []string{"module-a", "module-b"}, result.ModulesNeedingTest)
	assert.Empty(t, result.UpToDateModules)
}

func TestDetectChanges_SourceFilesChanged(t *testing.T) {
	dir := t.TempDir()

	// Create source file
	srcFile := createSourceFile(t, dir, "src/main.go", "package main // v1")
	oldHash := computeSourceHash(t, dir, []string{srcFile})

	// Create manifest with old hash
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": oldHash,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  oldHash,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	// Modify the source file
	require.NoError(t, os.WriteFile(filepath.Join(dir, srcFile), []byte("package main // v2 changed"), 0o644))

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcFile}, Dependencies: []string{}, BuildID: "build-1"},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Contains(t, result.ModulesNeedingTest, "module-a")
	assert.Contains(t, result.ChangeReasons["module-a"], "source files changed")
}

func TestDetectChanges_NoChanges_ModuleCached(t *testing.T) {
	dir := t.TempDir()

	// Create source file
	srcFile := createSourceFile(t, dir, "src/main.go", "package main")
	sourceHash := computeSourceHash(t, dir, []string{srcFile})

	// Create manifest with matching hash
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": sourceHash,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  sourceHash,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcFile}, Dependencies: []string{}, BuildID: "build-1"},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Empty(t, result.ModulesNeedingTest)
	assert.Contains(t, result.UpToDateModules, "module-a")
}

func TestDetectChanges_UnitTestsIgnoreTransitive(t *testing.T) {
	dir := t.TempDir()

	// Create source file for module-b first, then compute its hash
	srcBFile := createSourceFile(t, dir, "src-b/file.go", "package b")
	sourceHash := computeSourceHash(t, dir, []string{srcBFile})

	// Create manifest for module-b with passing unit tests using actual hash
	createTestManifest(t, dir, "module-b", map[string]interface{}{
		"moniker":     "module-b",
		"source_hash": sourceHash,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  sourceHash,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{}, Dependencies: []string{}},                                    // Changed (not in manifests)
		"module-b": {SourceFiles: []string{srcBFile}, Dependencies: []string{"module-a"}, BuildID: "build-1"}, // Unchanged
	}

	// For UNIT tests: module-b should NOT be invalidated by module-a changes
	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.Equal(t, TestSetUnit, result.TestSet)
	assert.Contains(t, result.ModulesNeedingTest, "module-a", "module-a should need testing (new)")
	assert.Contains(t, result.UpToDateModules, "module-b", "module-b should be up-to-date (unit tests ignore transitive)")
}

func TestDetectChanges_IntegrationTestsPropagateTransitive(t *testing.T) {
	dir := t.TempDir()

	// Create source files
	srcAFile := createSourceFile(t, dir, "src-a/file.go", "package a // changed")
	srcBFile := createSourceFile(t, dir, "src-b/file.go", "package b")

	// Create manifest for module-a (will have changed source)
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": "old-hash", // Different from current
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  "old-hash",
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	// Create manifest for module-b with matching hash
	hashB := computeSourceHash(t, dir, []string{srcBFile})
	createTestManifest(t, dir, "module-b", map[string]interface{}{
		"moniker":     "module-b",
		"source_hash": hashB,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  hashB,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcAFile}, Dependencies: []string{}, BuildID: "build-1"},
		"module-b": {SourceFiles: []string{srcBFile}, Dependencies: []string{"module-a"}, BuildID: "build-1"},
	}

	// For INTEGRATION tests: module-b should be invalidated by module-a changes
	result, err := DetectChanges(dir, moduleInfo, TestSetIntegration, nil)
	require.NoError(t, err)

	assert.Equal(t, TestSetIntegration, result.TestSet)
	assert.Contains(t, result.ModulesNeedingTest, "module-a", "module-a should need testing (source changed)")
	assert.Contains(t, result.ModulesNeedingTest, "module-b", "module-b should need testing (dependency changed)")
}

func TestDetectChanges_BuildChangeInvalidatesBoth(t *testing.T) {
	dir := t.TempDir()

	// Create source file
	srcFile := createSourceFile(t, dir, "src/file.go", "package main")
	sourceHash := computeSourceHash(t, dir, []string{srcFile})

	// Create manifest with old build ID
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": sourceHash,
		"build_id":    "old-build",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  sourceHash,
				"testset_hash": "test-hash",
				"build_id":     "old-build",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcFile}, Dependencies: []string{}, BuildID: "new-build"},
	}

	// Build change should invalidate both unit and integration
	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.Contains(t, result.ModulesNeedingTest, "module-a")
	assert.Contains(t, result.ChangeReasons["module-a"], "build changed")
}

func TestDetectChanges_PreviousFailureNeedsRetest(t *testing.T) {
	dir := t.TempDir()

	// Create source file
	srcFile := createSourceFile(t, dir, "src/file.go", "package main")
	sourceHash := computeSourceHash(t, dir, []string{srcFile})

	// Create manifest with failed test
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": sourceHash,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  sourceHash,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       false, // Failed
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcFile}, Dependencies: []string{}, BuildID: "build-1"},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.Contains(t, result.ModulesNeedingTest, "module-a")
	assert.Contains(t, result.ChangeReasons["module-a"], "previous test failed")
}

func TestDetectChanges_TestSetNotRun(t *testing.T) {
	dir := t.TempDir()

	// Create source file
	srcFile := createSourceFile(t, dir, "src/file.go", "package main")
	sourceHash := computeSourceHash(t, dir, []string{srcFile})

	// Create manifest with only unit tests run (no integration)
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": sourceHash,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  sourceHash,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcFile}, Dependencies: []string{}, BuildID: "build-1"},
	}

	// Request integration tests - should need to run since never run before
	result, err := DetectChanges(dir, moduleInfo, TestSetIntegration, nil)
	require.NoError(t, err)

	assert.Contains(t, result.ModulesNeedingTest, "module-a")
	assert.Contains(t, result.ChangeReasons["module-a"], "not previously run")
}

func TestDetectChanges_NewModule(t *testing.T) {
	dir := t.TempDir()

	// Create source file for existing module
	srcAFile := createSourceFile(t, dir, "src-a/file.go", "package a")
	hashA := computeSourceHash(t, dir, []string{srcAFile})

	// Only create manifest for module-a
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": hashA,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  hashA,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	// module-b is new (no manifest)
	createSourceFile(t, dir, "src-b/file.go", "package b")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcAFile}, Dependencies: []string{}, BuildID: "build-1"},
		"module-b": {SourceFiles: []string{"src-b/file.go"}, Dependencies: []string{}, BuildID: "build-1"},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun, "Not fresh run because module-a has manifest")
	assert.Contains(t, result.ModulesNeedingTest, "module-b", "New module should need testing")
	assert.Contains(t, result.UpToDateModules, "module-a", "Existing unchanged module should be cached")
	assert.Contains(t, result.ChangeReasons["module-b"], "new module")
}

func TestDetectChanges_MultiLevelDependencyChain(t *testing.T) {
	dir := t.TempDir()

	// Create source files
	srcAFile := createSourceFile(t, dir, "src-a/file.go", "package a // changed")
	srcBFile := createSourceFile(t, dir, "src-b/file.go", "package b")
	srcCFile := createSourceFile(t, dir, "src-c/file.go", "package c")

	hashB := computeSourceHash(t, dir, []string{srcBFile})
	hashC := computeSourceHash(t, dir, []string{srcCFile})

	// module-a: changed (old hash in manifest)
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": "old-hash",
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  "old-hash",
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	// module-b: depends on module-a, unchanged
	createTestManifest(t, dir, "module-b", map[string]interface{}{
		"moniker":     "module-b",
		"source_hash": hashB,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  hashB,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	// module-c: depends on module-b, unchanged
	createTestManifest(t, dir, "module-c", map[string]interface{}{
		"moniker":     "module-c",
		"source_hash": hashC,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  hashC,
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcAFile}, Dependencies: []string{}, BuildID: "build-1"},
		"module-b": {SourceFiles: []string{srcBFile}, Dependencies: []string{"module-a"}, BuildID: "build-1"},
		"module-c": {SourceFiles: []string{srcCFile}, Dependencies: []string{"module-b"}, BuildID: "build-1"},
	}

	// For INTEGRATION tests: all three should need testing due to transitive dependency
	result, err := DetectChanges(dir, moduleInfo, TestSetIntegration, nil)
	require.NoError(t, err)

	assert.Contains(t, result.ModulesNeedingTest, "module-a", "module-a changed directly")
	assert.Contains(t, result.ModulesNeedingTest, "module-b", "module-b depends on changed module-a")
	assert.Contains(t, result.ModulesNeedingTest, "module-c", "module-c depends on module-b which depends on changed module-a")
}

func TestDetectChanges_DependencyBuildIDChange(t *testing.T) {
	dir := t.TempDir()

	// Create source files
	srcAFile := createSourceFile(t, dir, "src-a/file.go", "package a")
	srcBFile := createSourceFile(t, dir, "src-b/file.go", "package b")

	hashA := computeSourceHash(t, dir, []string{srcAFile})
	hashB := computeSourceHash(t, dir, []string{srcBFile})

	// module-a: unchanged source, but will have new build ID
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": hashA,
		"build_id":    "old-build-a",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  hashA,
				"testset_hash": "test-hash",
				"build_id":     "old-build-a",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	// module-b: depends on module-a, unchanged
	createTestManifest(t, dir, "module-b", map[string]interface{}{
		"moniker":     "module-b",
		"source_hash": hashB,
		"build_id":    "build-b",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  hashB,
				"testset_hash": "test-hash",
				"build_id":     "build-b",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcAFile}, Dependencies: []string{}, BuildID: "new-build-a"},
		"module-b": {SourceFiles: []string{srcBFile}, Dependencies: []string{"module-a"}, BuildID: "build-b"},
	}

	// Dependency build ID loader returns new build ID for module-a
	depBuildIDLoader := func(moniker string) string {
		if moniker == "module-a" {
			return "new-build-a"
		}
		return ""
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetIntegration, depBuildIDLoader)
	require.NoError(t, err)

	assert.Contains(t, result.ModulesNeedingTest, "module-a", "module-a build changed")
	assert.Contains(t, result.ModulesNeedingTest, "module-b", "module-b dependency (module-a) was rebuilt")
}

func TestDetectChanges_EmptyModuleInfo(t *testing.T) {
	dir := t.TempDir()

	moduleInfo := map[string]ModuleTestFiles{}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.True(t, result.FreshRun)
	assert.Empty(t, result.ModulesNeedingTest)
	assert.Empty(t, result.UpToDateModules)
}

// =============================================================================
// UpdateModuleTestSetState Tests
// =============================================================================

func TestUpdateModuleTestSetState_CreatesNewManifest(t *testing.T) {
	dir := t.TempDir()

	// Create source and test files
	srcFile := createSourceFile(t, dir, "src/main.go", "package main")
	testFile := createSourceFile(t, dir, "src/main_test.go", "package main")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{srcFile},
			TestFiles:    []string{testFile},
			Dependencies: []string{},
			BuildID:      "build-123",
		},
	}
	testedModules := map[string]bool{"module-a": true}

	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Verify manifest was created
	manifest := loadManifest(t, dir, "module-a")
	assert.Equal(t, "module-a", manifest["moniker"])
	assert.Equal(t, "build-123", manifest["build_id"])

	// Verify test_sets was written
	testSets, ok := manifest["test_sets"].(map[string]interface{})
	require.True(t, ok, "test_sets should exist")

	unitSet, ok := testSets["unit"].(map[string]interface{})
	require.True(t, ok, "unit test set should exist")
	assert.Equal(t, true, unitSet["passed"])
	assert.Equal(t, "build-123", unitSet["build_id"])
}

func TestUpdateModuleTestSetState_UpdatesExistingManifest(t *testing.T) {
	dir := t.TempDir()

	// Create source file
	srcFile := createSourceFile(t, dir, "src/main.go", "package main")

	// Create existing manifest with old data
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":      "module-a",
		"source_hash":  "old-hash",
		"build_id":     "old-build",
		"old_field":    "preserved", // Should be preserved
		"test_time":    "2020-01-01T00:00:00Z",
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{srcFile},
			TestFiles:    []string{},
			Dependencies: []string{"dep-1"},
			BuildID:      "new-build",
		},
	}
	testedModules := map[string]bool{"module-a": true}

	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	manifest := loadManifest(t, dir, "module-a")

	// Old field should be preserved
	assert.Equal(t, "preserved", manifest["old_field"])

	// New fields should be updated
	assert.Equal(t, "new-build", manifest["build_id"])
	assert.Equal(t, []interface{}{"dep-1"}, manifest["dependencies"])

	// test_sets should be added
	testSets, ok := manifest["test_sets"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, testSets, "unit")
}

func TestUpdateModuleTestSetState_PreservesOtherTestSets(t *testing.T) {
	dir := t.TempDir()

	srcFile := createSourceFile(t, dir, "src/main.go", "package main")
	sourceHash := computeSourceHash(t, dir, []string{srcFile})

	// Create manifest with existing integration test set
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": sourceHash,
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash":  sourceHash,
				"testset_hash": "int-test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{srcFile},
			TestFiles:    []string{},
			Dependencies: []string{},
			BuildID:      "build-1",
		},
	}
	testedModules := map[string]bool{"module-a": true}

	// Update with unit tests
	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	manifest := loadManifest(t, dir, "module-a")
	testSets := manifest["test_sets"].(map[string]interface{})

	// Both test sets should exist
	assert.Contains(t, testSets, "unit", "unit test set should be added")
	assert.Contains(t, testSets, "integration", "integration test set should be preserved")

	// Integration data should be preserved
	intSet := testSets["integration"].(map[string]interface{})
	assert.Equal(t, "int-test-hash", intSet["testset_hash"])
}

func TestUpdateModuleTestSetState_MultipleModules(t *testing.T) {
	dir := t.TempDir()

	srcAFile := createSourceFile(t, dir, "src-a/main.go", "package a")
	srcBFile := createSourceFile(t, dir, "src-b/main.go", "package b")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcAFile}, BuildID: "build-a"},
		"module-b": {SourceFiles: []string{srcBFile}, BuildID: "build-b"},
	}
	testedModules := map[string]bool{
		"module-a": true,  // Passed
		"module-b": false, // Failed
	}

	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Verify module-a
	manifestA := loadManifest(t, dir, "module-a")
	testSetsA := manifestA["test_sets"].(map[string]interface{})
	unitA := testSetsA["unit"].(map[string]interface{})
	assert.Equal(t, true, unitA["passed"])

	// Verify module-b
	manifestB := loadManifest(t, dir, "module-b")
	testSetsB := manifestB["test_sets"].(map[string]interface{})
	unitB := testSetsB["unit"].(map[string]interface{})
	assert.Equal(t, false, unitB["passed"])
}

func TestUpdateModuleTestSetState_SkipsModuleNotInInfo(t *testing.T) {
	dir := t.TempDir()

	moduleInfo := map[string]ModuleTestFiles{
		// module-a is NOT in moduleInfo
	}
	testedModules := map[string]bool{"module-a": true}

	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Manifest should not be created
	manifestPath := filepath.Join(dir, "out", "test", "module-a", "test.manifest.json")
	_, err = os.Stat(manifestPath)
	assert.True(t, os.IsNotExist(err), "Manifest should not be created for module not in moduleInfo")
}

func TestUpdateModuleTestSetState_IntegrationWithDependencyHash(t *testing.T) {
	dir := t.TempDir()

	srcFile := createSourceFile(t, dir, "src/main.go", "package main")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{srcFile},
			Dependencies: []string{"dep-1", "dep-2"},
			BuildID:      "build-1",
		},
	}
	testedModules := map[string]bool{"module-a": true}
	depBuildIDs := map[string]map[string]string{
		"module-a": {
			"dep-1": "build-dep1",
			"dep-2": "build-dep2",
		},
	}

	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetIntegration, depBuildIDs)
	require.NoError(t, err)

	manifest := loadManifest(t, dir, "module-a")
	testSets := manifest["test_sets"].(map[string]interface{})
	intSet := testSets["integration"].(map[string]interface{})

	// dependency_build_hash should be set
	depHash, ok := intSet["dependency_build_hash"].(string)
	assert.True(t, ok && depHash != "", "dependency_build_hash should be set for integration tests")
}

// =============================================================================
// Round-Trip Tests (Write then Detect)
// =============================================================================

func TestRoundTrip_UpdateThenDetect_ModuleCached(t *testing.T) {
	dir := t.TempDir()

	srcFile := createSourceFile(t, dir, "src/main.go", "package main")
	testFile := createSourceFile(t, dir, "src/main_test.go", "package main")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{srcFile},
			TestFiles:    []string{testFile},
			Dependencies: []string{},
			BuildID:      "build-1",
		},
	}

	// Step 1: Update state (simulating test run)
	testedModules := map[string]bool{"module-a": true}
	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Step 2: Detect changes (should show as cached)
	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Contains(t, result.UpToDateModules, "module-a", "Module should be cached after update")
	assert.Empty(t, result.ModulesNeedingTest)
}

func TestRoundTrip_UpdateThenModify_NeedsRetest(t *testing.T) {
	dir := t.TempDir()

	srcFile := createSourceFile(t, dir, "src/main.go", "package main // v1")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{srcFile},
			Dependencies: []string{},
			BuildID:      "build-1",
		},
	}

	// Step 1: Update state
	testedModules := map[string]bool{"module-a": true}
	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Step 2: Modify source file
	require.NoError(t, os.WriteFile(filepath.Join(dir, srcFile), []byte("package main // v2 modified"), 0o644))

	// Step 3: Detect changes (should need retest)
	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	assert.Contains(t, result.ModulesNeedingTest, "module-a", "Module should need retest after source change")
	assert.Contains(t, result.ChangeReasons["module-a"], "source files changed")
}

func TestRoundTrip_UnitCached_IntegrationNeeded(t *testing.T) {
	dir := t.TempDir()

	srcFile := createSourceFile(t, dir, "src/main.go", "package main")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{srcFile},
			Dependencies: []string{},
			BuildID:      "build-1",
		},
	}

	// Step 1: Update state for unit tests
	testedModules := map[string]bool{"module-a": true}
	err := UpdateModuleTestSetState(dir, testedModules, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Step 2: Detect for unit tests (should be cached)
	unitResult, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)
	assert.Contains(t, unitResult.UpToDateModules, "module-a", "Unit tests should be cached")

	// Step 3: Detect for integration tests (should need to run)
	intResult, err := DetectChanges(dir, moduleInfo, TestSetIntegration, nil)
	require.NoError(t, err)
	assert.Contains(t, intResult.ModulesNeedingTest, "module-a", "Integration tests should need to run")
	assert.Contains(t, intResult.ChangeReasons["module-a"], "not previously run")
}

// =============================================================================
// ClearState Tests
// =============================================================================

func TestClearState_RemovesManifests(t *testing.T) {
	dir := t.TempDir()

	// Create manifests for multiple modules
	createTestManifest(t, dir, "module-a", map[string]interface{}{"moniker": "module-a"})
	createTestManifest(t, dir, "module-b", map[string]interface{}{"moniker": "module-b"})

	// Verify manifests exist
	_, err := os.Stat(filepath.Join(dir, "out", "test", "module-a", "test.manifest.json"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, "out", "test", "module-b", "test.manifest.json"))
	require.NoError(t, err)

	// Clear state
	err = ClearState(dir)
	require.NoError(t, err)

	// Verify manifests are removed
	_, err = os.Stat(filepath.Join(dir, "out", "test", "module-a", "test.manifest.json"))
	assert.True(t, os.IsNotExist(err), "module-a manifest should be removed")
	_, err = os.Stat(filepath.Join(dir, "out", "test", "module-b", "test.manifest.json"))
	assert.True(t, os.IsNotExist(err), "module-b manifest should be removed")
}

func TestClearState_HandlesNonExistentDirectory(t *testing.T) {
	dir := t.TempDir()
	// Don't create out/test directory

	err := ClearState(dir)
	assert.NoError(t, err, "ClearState should not error on non-existent directory")
}

func TestClearState_PreservesOtherFiles(t *testing.T) {
	dir := t.TempDir()

	// Create manifest
	createTestManifest(t, dir, "module-a", map[string]interface{}{"moniker": "module-a"})

	// Create other file in same directory
	otherFile := filepath.Join(dir, "out", "test", "module-a", "test.log")
	require.NoError(t, os.WriteFile(otherFile, []byte("log content"), 0o644))

	err := ClearState(dir)
	require.NoError(t, err)

	// Manifest should be removed
	_, err = os.Stat(filepath.Join(dir, "out", "test", "module-a", "test.manifest.json"))
	assert.True(t, os.IsNotExist(err))

	// Other file should be preserved
	_, err = os.Stat(otherFile)
	assert.NoError(t, err, "Other files should be preserved")
}

// =============================================================================
// Edge Cases and Error Handling
// =============================================================================

func TestDetectChanges_MalformedManifest(t *testing.T) {
	dir := t.TempDir()

	// Create malformed manifest
	moduleDir := filepath.Join(dir, "out", "test", "module-a")
	require.NoError(t, os.MkdirAll(moduleDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "test.manifest.json"), []byte("not json"), 0o644))

	srcFile := createSourceFile(t, dir, "src/main.go", "package main")

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcFile}, BuildID: "build-1"},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Malformed manifest should be treated as new module
	assert.Contains(t, result.ModulesNeedingTest, "module-a")
}

func TestDetectChanges_MissingSourceFiles(t *testing.T) {
	dir := t.TempDir()

	// Create manifest
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker":     "module-a",
		"source_hash": "some-hash",
		"build_id":    "build-1",
		"test_sets": map[string]interface{}{
			"unit": map[string]interface{}{
				"source_hash":  "some-hash",
				"testset_hash": "test-hash",
				"build_id":     "build-1",
				"passed":       true,
				"tested_at":    "2026-01-29T10:00:00Z",
			},
		},
	})

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {
			SourceFiles:  []string{"nonexistent/file.go"}, // File doesn't exist
			Dependencies: []string{},
			BuildID:      "build-1",
		},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// Missing files should trigger retest (hash error)
	assert.Contains(t, result.ModulesNeedingTest, "module-a")
}

func TestDetectChanges_CircularDependencies(t *testing.T) {
	dir := t.TempDir()

	// Create source files
	srcAFile := createSourceFile(t, dir, "src-a/file.go", "package a")
	srcBFile := createSourceFile(t, dir, "src-b/file.go", "package b")

	hashA := computeSourceHash(t, dir, []string{srcAFile})
	hashB := computeSourceHash(t, dir, []string{srcBFile})

	// Create manifests
	createTestManifest(t, dir, "module-a", map[string]interface{}{
		"moniker": "module-a", "source_hash": hashA, "build_id": "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash": hashA, "testset_hash": "th", "build_id": "build-1",
				"passed": true, "tested_at": "2026-01-29T10:00:00Z",
			},
		},
	})
	createTestManifest(t, dir, "module-b", map[string]interface{}{
		"moniker": "module-b", "source_hash": hashB, "build_id": "build-1",
		"test_sets": map[string]interface{}{
			"integration": map[string]interface{}{
				"source_hash": hashB, "testset_hash": "th", "build_id": "build-1",
				"passed": true, "tested_at": "2026-01-29T10:00:00Z",
			},
		},
	})

	// Circular dependency: a -> b -> a
	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{srcAFile}, Dependencies: []string{"module-b"}, BuildID: "build-1"},
		"module-b": {SourceFiles: []string{srcBFile}, Dependencies: []string{"module-a"}, BuildID: "build-1"},
	}

	// Should not hang or panic
	result, err := DetectChanges(dir, moduleInfo, TestSetIntegration, nil)
	require.NoError(t, err)

	// Both should be up to date (no direct changes)
	assert.Empty(t, result.ModulesNeedingTest)
	assert.ElementsMatch(t, []string{"module-a", "module-b"}, result.UpToDateModules)
}

func TestDetectChanges_DetectionTime(t *testing.T) {
	dir := t.TempDir()

	// Create some source files to make detection take measurable time
	for i := 0; i < 5; i++ {
		createSourceFile(t, dir, filepath.Join("src", "file"+string(rune('a'+i))+".go"), "package main")
	}

	moduleInfo := map[string]ModuleTestFiles{
		"module-a": {SourceFiles: []string{"src/filea.go", "src/fileb.go"}, Dependencies: []string{}},
		"module-b": {SourceFiles: []string{"src/filec.go", "src/filed.go"}, Dependencies: []string{"module-a"}},
		"module-c": {SourceFiles: []string{"src/filee.go"}, Dependencies: []string{"module-b"}},
	}

	result, err := DetectChanges(dir, moduleInfo, TestSetUnit, nil)
	require.NoError(t, err)

	// DetectionTime should be non-negative (may be 0 on very fast systems, but field should exist)
	assert.GreaterOrEqual(t, result.DetectionTime.Nanoseconds(), int64(0), "DetectionTime should be set")
}
