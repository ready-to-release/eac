package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// UoWTracker Interface Definition (for reference)
// =============================================================================

// UoWTracker tracks UoW execution and persists manifests.
// This interface will be implemented in tracker.go.
type UoWTracker interface {
	// RecordStart marks a UoW as started.
	RecordStart(id workunit.UnitID) error

	// RecordComplete records completion and persists UoW manifest.
	RecordComplete(id workunit.UnitID, manifest *UoWManifest) error

	// RecordCacheHit records a cache hit (loads existing manifest).
	// Returns error if manifest missing/invalid (cache miss).
	RecordCacheHit(id workunit.UnitID) (*UoWManifest, error)
}

// =============================================================================
// Test Helpers
// =============================================================================

// createTestUnitID creates a UnitID for testing purposes.
func createTestUnitID(ctx core.ActionType, module, component, tool string) workunit.UnitID {
	return workunit.UnitID{
		Action:        ctx,
		Module:        module,
		ComponentType: component,
		ComponentName: component,
		Tool:          tool,
	}
}

// createTestManifest creates a UoWManifest for testing purposes.
func createTestManifest(ctx core.ActionType, module, component, tool string) *UoWManifest {
	return &UoWManifest{
		Action:     ctx,
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   0,
		InputHash:  "sha256:test-input-hash",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:test-output-hash",
		Version:    "1.0.0",
	}
}

// createManifestOnDisk creates a manifest file at the expected location.
func createManifestOnDisk(t *testing.T, workspaceRoot string, manifest *UoWManifest) string {
	t.Helper()
	dirName := manifest.DirName()
	manifestDir := filepath.Join(workspaceRoot, "out", string(manifest.Action), manifest.Module, dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	return manifestPath
}

// createArtifactOnDisk creates an artifact file and returns its hash.
func createArtifactOnDisk(t *testing.T, basePath, relativePath string, content []byte) (string, int64) {
	t.Helper()
	fullPath := filepath.Join(basePath, relativePath)
	dir := filepath.Dir(fullPath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(fullPath, content, 0644)
	require.NoError(t, err)

	_, hash, err := HashFile(fullPath)
	require.NoError(t, err)
	return hash, int64(len(content))
}

// =============================================================================
// RecordStart Tests
// =============================================================================

func TestUoWTracker_RecordStart_CreatesOutputDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)
	err := tracker.RecordStart(id)
	require.NoError(t, err)

	// Expected directory path: out/build/test-module/go_go/
	expectedDir := filepath.Join(tmpDir, "out", "build", "test-module", "go_go")
	info, err := os.Stat(expectedDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "RecordStart should create the output directory")
}

func TestUoWTracker_RecordStart_IdempotentForSameUoW(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)

	// Calling RecordStart multiple times for the same UoW should not error
	err := tracker.RecordStart(id)
	require.NoError(t, err)

	err = tracker.RecordStart(id)
	require.NoError(t, err)

	err = tracker.RecordStart(id)
	require.NoError(t, err)
}

func TestUoWTracker_RecordStart_AllContexts(t *testing.T) {
	contexts := []core.ActionType{
		core.ActionBuild,
		core.ActionTest,
		core.ActionLint,
		core.ActionScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			tmpDir := t.TempDir()
			id := createTestUnitID(ctx, "module", "component", "tool")

			tracker := NewTracker(tmpDir, ctx)
			err := tracker.RecordStart(id)
			require.NoError(t, err)

			// Expected directory path: out/{context}/module/component_tool/
			expectedDir := filepath.Join(tmpDir, "out", string(ctx), "module", "component_tool")
			info, err := os.Stat(expectedDir)
			require.NoError(t, err)
			assert.True(t, info.IsDir(), "RecordStart should create directory for context %s", ctx)
		})
	}
}

func TestUoWTracker_RecordStart_RecordsStartTime(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)

	before := time.Now()
	err := tracker.RecordStart(id)
	require.NoError(t, err)

	// Verify start time is recorded by completing with Duration=0 and checking it gets set
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Duration = 0

	time.Sleep(10 * time.Millisecond)
	err = tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// Duration should be non-zero since RecordStart recorded a start time
	assert.True(t, manifest.Duration > 0, "Duration should be set from start time")
	assert.True(t, manifest.Duration >= 10*time.Millisecond, "Duration should be at least 10ms")
	_ = before
}

// =============================================================================
// RecordComplete Tests
// =============================================================================

func TestUoWTracker_RecordComplete_WritesManifestToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)
	err := tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// Expected manifest path: out/build/test-module/go_go/uow.manifest.json
	expectedPath := filepath.Join(tmpDir, "out", "build", "test-module", "go_go", "uow.manifest.json")
	_, err = os.Stat(expectedPath)
	require.NoError(t, err, "Manifest file should exist at expected path")

	// Load and verify content matches
	loaded, err := Load(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, manifest.Module, loaded.Module)
	assert.Equal(t, manifest.Component, loaded.Component)
	assert.Equal(t, manifest.Tool, loaded.Tool)
	assert.Equal(t, manifest.InputHash, loaded.InputHash)
}

func TestUoWTracker_RecordComplete_SetsDurationFromStartTime(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Duration = 0 // Duration should be set by RecordComplete

	tracker := NewTracker(tmpDir, core.ActionBuild)
	err := tracker.RecordStart(id)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	err = tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	assert.True(t, manifest.Duration >= 50*time.Millisecond, "Duration should be at least 50ms, got %v", manifest.Duration)
}

func TestUoWTracker_RecordComplete_WithoutRecordStart_StillWorks(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Duration = 30 * time.Second // Pre-set duration since no start time

	tracker := NewTracker(tmpDir, core.ActionBuild)

	// RecordComplete should work even without RecordStart
	err := tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	expectedPath := filepath.Join(tmpDir, "out", "build", "test-module", "go_go", "uow.manifest.json")
	_, err = os.Stat(expectedPath)
	require.NoError(t, err, "Manifest should be written even without RecordStart")

	// Duration should be preserved since no start time exists
	assert.Equal(t, 30*time.Second, manifest.Duration)
}

func TestUoWTracker_RecordComplete_CreatesDirectoryIfMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "new-module", "go", "go")
	manifest := createTestManifest(core.ActionBuild, "new-module", "go", "go")

	expectedDir := filepath.Join(tmpDir, "out", "build", "new-module", "go_go")

	// Directory should not exist before RecordComplete
	_, err := os.Stat(expectedDir)
	assert.True(t, os.IsNotExist(err))

	tracker := NewTracker(tmpDir, core.ActionBuild)
	err = tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// Verify directory is created
	info, err := os.Stat(expectedDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify manifest file exists
	_, err = os.Stat(filepath.Join(expectedDir, "uow.manifest.json"))
	require.NoError(t, err)
}

func TestUoWTracker_RecordComplete_OverwritesExistingManifest(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)

	// Create first manifest
	manifest1 := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest1.InputHash = "sha256:first-hash"
	err := tracker.RecordComplete(id, manifest1)
	require.NoError(t, err)

	// Create second manifest
	manifest2 := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest2.InputHash = "sha256:second-hash"
	err = tracker.RecordComplete(id, manifest2)
	require.NoError(t, err)

	// Load manifest from disk and verify it contains "sha256:second-hash"
	manifestPath := filepath.Join(tmpDir, "out", "build", "test-module", "go_go", "uow.manifest.json")
	loaded, err := Load(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, "sha256:second-hash", loaded.InputHash)
}

func TestUoWTracker_RecordComplete_WritesAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionTest, "core", "go", "gotest")

	manifest := &UoWManifest{
		Action:     core.ActionTest,
		Module:     "core",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   1,
		InputHash:  "sha256:specific-input-hash",
		ExecutedAt: time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
		Duration:   123 * time.Second,
		Artifacts: []Artifact{
			{ID: "junit", Path: "junit.xml", SHA256: "sha256:junit-hash", Size: 1000, Type: "report"},
			{ID: "coverage", Path: "coverage.out", SHA256: "sha256:cov-hash", Size: 500, Type: "coverage"},
		},
		OutputHash: "sha256:specific-output-hash",
		Version:    "2.0.0",
	}

	tracker := NewTracker(tmpDir, core.ActionTest)
	err := tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	expectedPath := filepath.Join(tmpDir, "out", "test", "core", "go_gotest", "uow.manifest.json")
	loaded, err := Load(expectedPath)
	require.NoError(t, err)

	assert.Equal(t, core.ActionTest, loaded.Action)
	assert.Equal(t, "core", loaded.Module)
	assert.Equal(t, "go", loaded.Component)
	assert.Equal(t, "gotest", loaded.Tool)
	assert.Equal(t, 1, loaded.ExitCode)
	assert.Equal(t, "sha256:specific-input-hash", loaded.InputHash)
	assert.Equal(t, time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC), loaded.ExecutedAt)
	assert.Equal(t, 123*time.Second, loaded.Duration)
	assert.Len(t, loaded.Artifacts, 2)
	assert.Equal(t, "sha256:specific-output-hash", loaded.OutputHash)
	assert.Equal(t, "2.0.0", loaded.Version)
}

func TestUoWTracker_RecordComplete_AllContexts(t *testing.T) {
	tests := []struct {
		name      string
		context   core.ActionType
		module    string
		component string
		tool      string
	}{
		{
			name:      "build context",
			context:   core.ActionBuild,
			module:    "eac",
			component: "go",
			tool:      "go",
		},
		{
			name:      "test context",
			context:   core.ActionTest,
			module:    "core",
			component: "go",
			tool:      "gotest",
		},
		{
			name:      "lint context",
			context:   core.ActionLint,
			module:    "web-app",
			component: "typescript",
			tool:      "eslint",
		},
		{
			name:      "scan context",
			context:   core.ActionScan,
			module:    "eac",
			component: "docker",
			tool:      "trivy-vuln",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			id := createTestUnitID(tt.context, tt.module, tt.component, tt.tool)
			manifest := createTestManifest(tt.context, tt.module, tt.component, tt.tool)

			tracker := NewTracker(tmpDir, tt.context)
			err := tracker.RecordComplete(id, manifest)
			require.NoError(t, err)

			dirName := tt.component + "_" + tt.tool
			expectedPath := filepath.Join(tmpDir, "out", string(tt.context), tt.module, dirName, "uow.manifest.json")
			loaded, err := Load(expectedPath)
			require.NoError(t, err)
			assert.Equal(t, tt.context, loaded.Action)
			assert.Equal(t, tt.module, loaded.Module)
			assert.Equal(t, tt.component, loaded.Component)
			assert.Equal(t, tt.tool, loaded.Tool)
		})
	}
}

func TestUoWTracker_RecordComplete_WithFailedExecution(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionTest, "failing-module", "go", "gotest")

	manifest := createTestManifest(core.ActionTest, "failing-module", "go", "gotest")
	manifest.ExitCode = 1 // Non-zero exit code

	tracker := NewTracker(tmpDir, core.ActionTest)
	err := tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// RecordComplete should save manifests for failed executions too
	manifestPath := filepath.Join(tmpDir, "out", "test", "failing-module", "go_gotest", "uow.manifest.json")
	loaded, err := Load(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, 1, loaded.ExitCode)
}

// =============================================================================
// RecordCacheHit Tests
// =============================================================================

func TestUoWTracker_RecordCacheHit_ReturnsExistingManifest(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create manifest on disk
	originalManifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	originalManifest.InputHash = "sha256:cache-test-input"
	createManifestOnDisk(t, tmpDir, originalManifest)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	loaded, err := tracker.RecordCacheHit(id)
	require.NoError(t, err)
	assert.Equal(t, "sha256:cache-test-input", loaded.InputHash)
	assert.Equal(t, originalManifest.Module, loaded.Module)
	assert.Equal(t, originalManifest.Component, loaded.Component)
	assert.Equal(t, originalManifest.Tool, loaded.Tool)
}

func TestUoWTracker_RecordCacheHit_ReturnsErrorWhenManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "nonexistent-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)
	manifest, err := tracker.RecordCacheHit(id)
	assert.Error(t, err)
	assert.Nil(t, manifest)
}

func TestUoWTracker_RecordCacheHit_ReturnsErrorWhenArtifactsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create manifest that references non-existent artifacts
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: "eac-linux-amd64", SHA256: "sha256:nonexistent", Size: 1000, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	result, err := tracker.RecordCacheHit(id)
	assert.Error(t, err, "Should return error for missing artifacts")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "missing")
}

func TestUoWTracker_RecordCacheHit_ReturnsErrorWhenArtifactsCorrupt(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create manifest and artifact, but with mismatched hash
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)

	// Create artifact with actual content
	artifactContent := []byte("actual binary content")
	artifactPath := filepath.Join(uowDir, "eac-linux-amd64")
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(artifactPath, artifactContent, 0644)
	require.NoError(t, err)

	// Create manifest with wrong hash
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: "eac-linux-amd64", SHA256: "sha256:wrong-hash", Size: int64(len(artifactContent)), Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	result, err := tracker.RecordCacheHit(id)
	assert.Error(t, err, "Should return error for corrupt artifacts")
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "corrupt")
}

func TestUoWTracker_RecordCacheHit_SucceedsWhenArtifactsValid(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create UoW directory
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	// Create artifact with actual content and compute hash
	artifactContent := []byte("valid binary content")
	artifactRelPath := "eac-linux-amd64"
	hash, size := createArtifactOnDisk(t, uowDir, artifactRelPath, artifactContent)

	// Create manifest with correct hash
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: artifactRelPath, SHA256: hash, Size: size, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	loaded, err := tracker.RecordCacheHit(id)
	require.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, manifest.Module, loaded.Module)
}

func TestUoWTracker_RecordCacheHit_ValidatesMultipleArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create UoW directory
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	// Create multiple artifacts
	artifacts := []Artifact{}

	artifact1Content := []byte("linux binary content")
	hash1, size1 := createArtifactOnDisk(t, uowDir, "eac-linux-amd64", artifact1Content)
	artifacts = append(artifacts, Artifact{ID: "linux", Path: "eac-linux-amd64", SHA256: hash1, Size: size1, Type: "binary"})

	artifact2Content := []byte("darwin binary content")
	hash2, size2 := createArtifactOnDisk(t, uowDir, "eac-darwin-amd64", artifact2Content)
	artifacts = append(artifacts, Artifact{ID: "darwin", Path: "eac-darwin-amd64", SHA256: hash2, Size: size2, Type: "binary"})

	artifact3Content := []byte("windows binary content")
	hash3, size3 := createArtifactOnDisk(t, uowDir, "eac-windows-amd64.exe", artifact3Content)
	artifacts = append(artifacts, Artifact{ID: "windows", Path: "eac-windows-amd64.exe", SHA256: hash3, Size: size3, Type: "binary"})

	// Create manifest with all artifacts
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = artifacts
	createManifestOnDisk(t, tmpDir, manifest)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	loaded, err := tracker.RecordCacheHit(id)
	require.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Len(t, loaded.Artifacts, 3)
}

func TestUoWTracker_RecordCacheHit_FailsIfAnyArtifactMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create UoW directory
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	// Create only one artifact but reference two in manifest
	artifact1Content := []byte("existing binary content")
	hash1, size1 := createArtifactOnDisk(t, uowDir, "eac-linux-amd64", artifact1Content)

	// Create manifest referencing existing and missing artifacts
	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: hash1, Size: size1, Type: "binary"},
		{ID: "darwin", Path: "eac-darwin-amd64", SHA256: "sha256:missing", Size: 1000, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	result, err := tracker.RecordCacheHit(id)
	assert.Error(t, err, "Should fail when any artifact is missing")
	assert.Nil(t, result)
}

func TestUoWTracker_RecordCacheHit_AllContexts(t *testing.T) {
	contexts := []core.ActionType{
		core.ActionBuild,
		core.ActionTest,
		core.ActionLint,
		core.ActionScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			tmpDir := t.TempDir()
			id := createTestUnitID(ctx, "module", "component", "tool")

			// Create manifest for this context
			manifest := createTestManifest(ctx, "module", "component", "tool")
			createManifestOnDisk(t, tmpDir, manifest)

			tracker := NewTracker(tmpDir, ctx)
			loaded, err := tracker.RecordCacheHit(id)
			require.NoError(t, err)
			assert.NotNil(t, loaded)
			assert.Equal(t, ctx, loaded.Action)
		})
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestUoWTracker_ConcurrentRecordStart(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple UnitIDs for different UoWs
	ids := []workunit.UnitID{
		createTestUnitID(core.ActionBuild, "module1", "go", "go"),
		createTestUnitID(core.ActionBuild, "module2", "go", "go"),
		createTestUnitID(core.ActionBuild, "module3", "go", "go"),
		createTestUnitID(core.ActionTest, "module1", "go", "gotest"),
		createTestUnitID(core.ActionLint, "module1", "go", "golangci-lint"),
	}

	tracker := NewTracker(tmpDir, core.ActionBuild)

	var wg sync.WaitGroup
	errors := make([]error, len(ids))
	for i, id := range ids {
		wg.Add(1)
		go func(idx int, id workunit.UnitID) {
			defer wg.Done()
			errors[idx] = tracker.RecordStart(id)
		}(i, id)
	}
	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "RecordStart failed for id %d", i)
	}
}

func TestUoWTracker_ConcurrentRecordComplete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple UnitIDs for different UoWs
	ids := []workunit.UnitID{
		createTestUnitID(core.ActionBuild, "module1", "go", "go"),
		createTestUnitID(core.ActionBuild, "module2", "go", "go"),
		createTestUnitID(core.ActionBuild, "module3", "go", "go"),
		createTestUnitID(core.ActionTest, "module1", "go", "gotest"),
		createTestUnitID(core.ActionLint, "module1", "go", "golangci-lint"),
	}

	tracker := NewTracker(tmpDir, core.ActionBuild)

	var wg sync.WaitGroup
	errors := make([]error, len(ids))
	for i, id := range ids {
		wg.Add(1)
		go func(idx int, id workunit.UnitID) {
			defer wg.Done()
			manifest := createTestManifest(id.Action, id.Module, id.ComponentName, id.Tool)
			errors[idx] = tracker.RecordComplete(id, manifest)
		}(i, id)
	}
	wg.Wait()

	// Verify all manifests were written correctly
	for i, err := range errors {
		assert.NoError(t, err, "RecordComplete failed for id %d", i)
	}
	for _, id := range ids {
		manifestPath := filepath.Join(tmpDir, "out", string(id.Action), id.Module, id.DirName(), "uow.manifest.json")
		_, err := os.Stat(manifestPath)
		assert.NoError(t, err, "Manifest should exist for %s", id.Longname())
	}
}

func TestUoWTracker_ConcurrentRecordCacheHit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple UnitIDs and their manifests on disk
	ids := []workunit.UnitID{
		createTestUnitID(core.ActionBuild, "module1", "go", "go"),
		createTestUnitID(core.ActionBuild, "module2", "go", "go"),
		createTestUnitID(core.ActionBuild, "module3", "go", "go"),
	}

	// Create manifests on disk for all IDs
	for _, id := range ids {
		manifest := createTestManifest(id.Action, id.Module, id.ComponentName, id.Tool)
		createManifestOnDisk(t, tmpDir, manifest)
	}

	tracker := NewTracker(tmpDir, core.ActionBuild)

	var wg sync.WaitGroup
	errors := make([]error, len(ids))
	for i, id := range ids {
		wg.Add(1)
		go func(idx int, id workunit.UnitID) {
			defer wg.Done()
			_, errors[idx] = tracker.RecordCacheHit(id)
		}(i, id)
	}
	wg.Wait()

	for i, err := range errors {
		assert.NoError(t, err, "RecordCacheHit failed for id %d", i)
	}
}

func TestUoWTracker_ConcurrentMixedOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some manifests on disk for cache hits
	cachedID := createTestUnitID(core.ActionBuild, "cached-module", "go", "go")
	cachedManifest := createTestManifest(core.ActionBuild, "cached-module", "go", "go")
	createManifestOnDisk(t, tmpDir, cachedManifest)

	// IDs for new work
	newID1 := createTestUnitID(core.ActionBuild, "new-module1", "go", "go")
	newID2 := createTestUnitID(core.ActionBuild, "new-module2", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)

	var wg sync.WaitGroup
	var startErr, completeErr, cacheErr error

	// RecordStart for new work
	wg.Add(1)
	go func() {
		defer wg.Done()
		startErr = tracker.RecordStart(newID1)
	}()

	// RecordComplete for new work
	wg.Add(1)
	go func() {
		defer wg.Done()
		completeErr = tracker.RecordComplete(newID2, createTestManifest(core.ActionBuild, "new-module2", "go", "go"))
	}()

	// RecordCacheHit for cached work
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, cacheErr = tracker.RecordCacheHit(cachedID)
	}()

	wg.Wait()

	assert.NoError(t, startErr, "RecordStart should succeed")
	assert.NoError(t, completeErr, "RecordComplete should succeed")
	assert.NoError(t, cacheErr, "RecordCacheHit should succeed")
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestUoWTracker_RecordComplete_WithEmptyArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionLint, "test-module", "go", "golangci-lint")

	manifest := createTestManifest(core.ActionLint, "test-module", "go", "golangci-lint")
	manifest.Artifacts = []Artifact{} // No artifacts for lint

	tracker := NewTracker(tmpDir, core.ActionLint)
	err := tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// Verify manifest was saved with empty artifacts
	manifestPath := filepath.Join(tmpDir, "out", "lint", "test-module", "go_golangci-lint", "uow.manifest.json")
	loaded, err := Load(manifestPath)
	require.NoError(t, err)
	assert.Empty(t, loaded.Artifacts)
}

func TestUoWTracker_RecordComplete_WithNilArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionLint, "test-module", "go", "golangci-lint")

	manifest := createTestManifest(core.ActionLint, "test-module", "go", "golangci-lint")
	manifest.Artifacts = nil // Nil artifacts

	tracker := NewTracker(tmpDir, core.ActionLint)
	err := tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// Verify manifest was saved (nil artifacts handled gracefully)
	manifestPath := filepath.Join(tmpDir, "out", "lint", "test-module", "go_golangci-lint", "uow.manifest.json")
	_, err = os.Stat(manifestPath)
	require.NoError(t, err)
}

func TestUoWTracker_RecordCacheHit_ManifestExistsButInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create manifest directory and write invalid JSON
	dirName := "go_go"
	manifestDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	err = os.WriteFile(manifestPath, []byte("invalid json {{{"), 0644)
	require.NoError(t, err)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	result, err := tracker.RecordCacheHit(id)
	assert.Error(t, err, "Should return error for invalid manifest JSON")
	assert.Nil(t, result)
}

func TestUoWTracker_RecordCacheHit_ManifestExistsButEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create manifest directory and write empty file
	dirName := "go_go"
	manifestDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	err = os.WriteFile(manifestPath, []byte(""), 0644)
	require.NoError(t, err)

	tracker := NewTracker(tmpDir, core.ActionBuild)
	result, err := tracker.RecordCacheHit(id)
	assert.Error(t, err, "Should return error for empty manifest")
	assert.Nil(t, result)
}

func TestUoWTracker_WorksWithSpecialCharactersInNames(t *testing.T) {
	tmpDir := t.TempDir()

	// Module and component names with hyphens (common pattern)
	id := createTestUnitID(core.ActionBuild, "my-complex-module", "go-component", "my-tool")
	manifest := createTestManifest(core.ActionBuild, "my-complex-module", "go-component", "my-tool")

	tracker := NewTracker(tmpDir, core.ActionBuild)
	err := tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// Expected path: out/build/my-complex-module/go-component_my-tool/uow.manifest.json
	expectedDir := filepath.Join(tmpDir, "out", "build", "my-complex-module", "go-component_my-tool")
	info, err := os.Stat(expectedDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestUoWTracker_RecordCacheHit_WithEmptyArtifactsList(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionLint, "test-module", "go", "golangci-lint")

	// Create manifest with empty artifacts (valid for lint operations)
	manifest := createTestManifest(core.ActionLint, "test-module", "go", "golangci-lint")
	manifest.Artifacts = []Artifact{}
	createManifestOnDisk(t, tmpDir, manifest)

	tracker := NewTracker(tmpDir, core.ActionLint)
	loaded, err := tracker.RecordCacheHit(id)
	require.NoError(t, err, "Empty artifacts list should be valid")
	assert.NotNil(t, loaded)
	assert.Empty(t, loaded.Artifacts)
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestUoWTracker_FullWorkflow_Success(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)

	// 1. RecordStart
	err := tracker.RecordStart(id)
	require.NoError(t, err)

	// 2. Create UoW directory and artifacts
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)

	// Create artifact
	artifactContent := []byte("built binary content")
	hash, size := createArtifactOnDisk(t, uowDir, "eac-linux-amd64", artifactContent)

	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Duration = 0 // Let tracker compute from start time
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: hash, Size: size, Type: "binary"},
	}

	// 3. RecordComplete
	err = tracker.RecordComplete(id, manifest)
	require.NoError(t, err)

	// 4. Verify manifest on disk
	manifestPath := filepath.Join(uowDir, "uow.manifest.json")
	loaded, err := Load(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, "test-module", loaded.Module)
	assert.Len(t, loaded.Artifacts, 1)
	assert.True(t, loaded.Duration > 0, "Duration should be set from start time")
}

func TestUoWTracker_FullWorkflow_CacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	tracker := NewTracker(tmpDir, core.ActionBuild)

	// Simulate existing build with valid artifacts
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	artifactContent := []byte("cached binary content")
	hash, size := createArtifactOnDisk(t, uowDir, "eac-linux-amd64", artifactContent)

	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: hash, Size: size, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// Now RecordCacheHit should succeed
	loaded, err := tracker.RecordCacheHit(id)
	require.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Equal(t, "test-module", loaded.Module)
}

func TestUoWTracker_FullWorkflow_CacheMiss(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(core.ActionBuild, "test-module", "go", "go")

	// Create manifest with artifact reference but no actual artifact file
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	manifest := createTestManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: "sha256:deleted-artifact", Size: 1000, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// Artifact does not exist - cache should be invalid
	tracker := NewTracker(tmpDir, core.ActionBuild)
	result, err := tracker.RecordCacheHit(id)
	assert.Error(t, err, "Cache should be invalid when artifacts are missing")
	assert.Nil(t, result)
}

// =============================================================================
// Table-Driven Tests
// =============================================================================

func TestUoWTracker_RecordComplete_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		context   core.ActionType
		module    string
		component string
		tool      string
		exitCode  int
		artifacts []Artifact
	}{
		{
			name:      "successful build with artifacts",
			context:   core.ActionBuild,
			module:    "eac",
			component: "go",
			tool:      "go",
			exitCode:  0,
			artifacts: []Artifact{
				{ID: "binary", Path: "eac", SHA256: "sha256:abc", Size: 1000, Type: "binary"},
			},
		},
		{
			name:      "failed test with report",
			context:   core.ActionTest,
			module:    "core",
			component: "go",
			tool:      "gotest",
			exitCode:  1,
			artifacts: []Artifact{
				{ID: "junit", Path: "junit.xml", SHA256: "sha256:def", Size: 500, Type: "report"},
			},
		},
		{
			name:      "lint with no artifacts",
			context:   core.ActionLint,
			module:    "web-app",
			component: "typescript",
			tool:      "eslint",
			exitCode:  0,
			artifacts: []Artifact{},
		},
		{
			name:      "scan with multiple artifacts",
			context:   core.ActionScan,
			module:    "eac",
			component: "docker",
			tool:      "trivy-vuln",
			exitCode:  0,
			artifacts: []Artifact{
				{ID: "vulns", Path: "vulns.json", SHA256: "sha256:ghi", Size: 2000, Type: "report"},
				{ID: "sbom", Path: "sbom.json", SHA256: "sha256:jkl", Size: 3000, Type: "sbom"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			id := createTestUnitID(tt.context, tt.module, tt.component, tt.tool)

			manifest := &UoWManifest{
				Action:     tt.context,
				Module:     tt.module,
				Component:  tt.component,
				Tool:       tt.tool,
				ExitCode:   tt.exitCode,
				InputHash:  "sha256:test-input",
				ExecutedAt: time.Now().UTC().Truncate(time.Second),
				Duration:   30 * time.Second,
				Artifacts:  tt.artifacts,
				OutputHash: "sha256:test-output",
				Version:    "1.0.0",
			}

			tracker := NewTracker(tmpDir, tt.context)
			err := tracker.RecordComplete(id, manifest)
			require.NoError(t, err)

			dirName := tt.component + "_" + tt.tool
			expectedPath := filepath.Join(tmpDir, "out", string(tt.context), tt.module, dirName, "uow.manifest.json")

			// Verify file exists at expectedPath
			_, err = os.Stat(expectedPath)
			require.NoError(t, err, "Manifest should exist at %s", expectedPath)

			// Load and verify content
			loaded, err := Load(expectedPath)
			require.NoError(t, err)
			assert.Equal(t, tt.exitCode, loaded.ExitCode)
			assert.Equal(t, tt.module, loaded.Module)
			assert.Equal(t, tt.component, loaded.Component)
			assert.Equal(t, tt.tool, loaded.Tool)
			assert.Len(t, loaded.Artifacts, len(tt.artifacts))
		})
	}
}

func TestUoWTracker_RecordCacheHit_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		setupManifest  *UoWManifest
		setupArtifacts bool
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid manifest with valid artifacts",
			setupManifest: &UoWManifest{
				Action:    core.ActionBuild,
				Module:    "test-module",
				Component: "go",
				Tool:      "go",
				Artifacts: []Artifact{
					{ID: "binary", Path: "eac", SHA256: "", Size: 0, Type: "binary"}, // Hash will be computed
				},
			},
			setupArtifacts: true,
			expectError:    false,
		},
		{
			name:          "no manifest exists",
			setupManifest: nil,
			expectError:   true,
			errorContains: "manifest",
		},
		{
			name: "manifest with missing artifacts",
			setupManifest: &UoWManifest{
				Action:    core.ActionBuild,
				Module:    "test-module",
				Component: "go",
				Tool:      "go",
				Artifacts: []Artifact{
					{ID: "binary", Path: "eac-missing", SHA256: "sha256:missing", Size: 1000, Type: "binary"},
				},
			},
			setupArtifacts: false,
			expectError:    true,
			errorContains:  "missing",
		},
		{
			name: "manifest with no artifacts (valid for lint)",
			setupManifest: &UoWManifest{
				Action:    core.ActionLint,
				Module:    "test-module",
				Component: "go",
				Tool:      "golangci-lint",
				Artifacts: []Artifact{},
			},
			setupArtifacts: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			var id workunit.UnitID
			if tt.setupManifest != nil {
				id = createTestUnitID(tt.setupManifest.Action, tt.setupManifest.Module, tt.setupManifest.Component, tt.setupManifest.Tool)

				// Set up UoW directory
				dirName := id.DirName()
				uowDir := filepath.Join(tmpDir, "out", string(tt.setupManifest.Action), tt.setupManifest.Module, dirName)
				err := os.MkdirAll(uowDir, 0755)
				require.NoError(t, err)

				// Create artifacts if requested
				if tt.setupArtifacts && len(tt.setupManifest.Artifacts) > 0 {
					for i := range tt.setupManifest.Artifacts {
						content := []byte("artifact content for " + tt.setupManifest.Artifacts[i].ID)
						hash, size := createArtifactOnDisk(t, uowDir, tt.setupManifest.Artifacts[i].Path, content)
						tt.setupManifest.Artifacts[i].SHA256 = hash
						tt.setupManifest.Artifacts[i].Size = size
					}
				}

				// Create manifest on disk
				createManifestOnDisk(t, tmpDir, tt.setupManifest)
			} else {
				id = createTestUnitID(core.ActionBuild, "nonexistent", "go", "go")
			}

			tracker := NewTracker(tmpDir, id.Action)
			result, err := tracker.RecordCacheHit(id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
