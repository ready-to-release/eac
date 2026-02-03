package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

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
func createTestUnitID(ctx workunit.Context, module, component, tool string) workunit.UnitID {
	return workunit.UnitID{
		Context:   ctx,
		Module:    module,
		Component: component,
		Tool:      tool,
	}
}

// createTestManifest creates a UoWManifest for testing purposes.
func createTestManifest(ctx workunit.Context, module, component, tool string) *UoWManifest {
	return &UoWManifest{
		Context:    ctx,
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
	dirName := manifest.Component + "_" + manifest.Tool
	manifestDir := filepath.Join(workspaceRoot, "out", string(manifest.Context), manifest.Module, dirName)
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
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Expected directory path: out/build/test-module/go_go/
	expectedDir := filepath.Join(tmpDir, "out", "build", "test-module", "go_go")

	// Verify directory does not exist before RecordStart
	_, err := os.Stat(expectedDir)
	assert.True(t, os.IsNotExist(err), "Directory should not exist before RecordStart")

	// TODO: After implementation, call tracker.RecordStart(id) and verify directory is created
	// For now, this test documents the expected behavior
	_ = id
	_ = expectedDir
}

func TestUoWTracker_RecordStart_IdempotentForSameUoW(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Calling RecordStart multiple times for the same UoW should not error
	// TODO: After implementation, verify multiple RecordStart calls succeed
	_ = tmpDir
	_ = id
}

func TestUoWTracker_RecordStart_AllContexts(t *testing.T) {
	contexts := []workunit.Context{
		workunit.ContextBuild,
		workunit.ContextTest,
		workunit.ContextLint,
		workunit.ContextScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			tmpDir := t.TempDir()
			id := createTestUnitID(ctx, "module", "component", "tool")

			// Expected directory path: out/{context}/module/component_tool/
			expectedDir := filepath.Join(tmpDir, "out", string(ctx), "module", "component_tool")

			// TODO: After implementation, verify directory is created for each context
			_ = expectedDir
			_ = id
		})
	}
}

func TestUoWTracker_RecordStart_RecordsStartTime(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// RecordStart should record the start time for duration calculation
	// TODO: After implementation, verify start time is recorded
	_ = tmpDir
	_ = id
}

// =============================================================================
// RecordComplete Tests
// =============================================================================

func TestUoWTracker_RecordComplete_WritesManifestToDisk(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")

	// Expected manifest path: out/build/test-module/go_go/uow.manifest.json
	expectedPath := filepath.Join(tmpDir, "out", "build", "test-module", "go_go", "uow.manifest.json")

	// TODO: After implementation:
	// 1. Call tracker.RecordComplete(id, manifest)
	// 2. Verify file exists at expectedPath
	// 3. Load and verify manifest content matches
	_ = id
	_ = manifest
	_ = expectedPath
}

func TestUoWTracker_RecordComplete_SetsDurationFromStartTime(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Duration = 0 // Duration should be set by RecordComplete

	// TODO: After implementation:
	// 1. Call tracker.RecordStart(id)
	// 2. time.Sleep(100ms) or similar
	// 3. Call tracker.RecordComplete(id, manifest)
	// 4. Verify manifest.Duration is approximately 100ms
	_ = tmpDir
	_ = id
	_ = manifest
}

func TestUoWTracker_RecordComplete_WithoutRecordStart_StillWorks(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Duration = 30 * time.Second // Pre-set duration since no start time

	expectedPath := filepath.Join(tmpDir, "out", "build", "test-module", "go_go", "uow.manifest.json")

	// RecordComplete should work even without RecordStart
	// Duration will use the pre-set value since no start time exists
	// TODO: After implementation, verify this behavior
	_ = id
	_ = manifest
	_ = expectedPath
}

func TestUoWTracker_RecordComplete_CreatesDirectoryIfMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "new-module", "go", "go")
	manifest := createTestManifest(workunit.ContextBuild, "new-module", "go", "go")

	expectedDir := filepath.Join(tmpDir, "out", "build", "new-module", "go_go")

	// Directory should not exist before RecordComplete
	_, err := os.Stat(expectedDir)
	assert.True(t, os.IsNotExist(err))

	// TODO: After implementation:
	// 1. Call tracker.RecordComplete(id, manifest) without RecordStart
	// 2. Verify directory is created
	// 3. Verify manifest file exists
	_ = id
	_ = manifest
}

func TestUoWTracker_RecordComplete_OverwritesExistingManifest(t *testing.T) {
	_ = t.TempDir() // Will be used after implementation
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Create first manifest
	manifest1 := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest1.InputHash = "sha256:first-hash"

	// Create second manifest
	manifest2 := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest2.InputHash = "sha256:second-hash"

	// TODO: After implementation:
	// 1. Call tracker.RecordComplete(id, manifest1)
	// 2. Call tracker.RecordComplete(id, manifest2)
	// 3. Load manifest from disk
	// 4. Verify it contains "sha256:second-hash"
	_ = id
	_ = manifest1
	_ = manifest2
}

func TestUoWTracker_RecordComplete_WritesAllFields(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextTest, "core", "go", "gotest")

	manifest := &UoWManifest{
		Context:    workunit.ContextTest,
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

	expectedPath := filepath.Join(tmpDir, "out", "test", "core", "go_gotest", "uow.manifest.json")

	// TODO: After implementation:
	// 1. Call tracker.RecordComplete(id, manifest)
	// 2. Load manifest from disk
	// 3. Verify all fields match
	_ = id
	_ = manifest
	_ = expectedPath
}

func TestUoWTracker_RecordComplete_AllContexts(t *testing.T) {
	tests := []struct {
		name      string
		context   workunit.Context
		module    string
		component string
		tool      string
	}{
		{
			name:      "build context",
			context:   workunit.ContextBuild,
			module:    "eac-cli",
			component: "go",
			tool:      "go",
		},
		{
			name:      "test context",
			context:   workunit.ContextTest,
			module:    "core",
			component: "go",
			tool:      "gotest",
		},
		{
			name:      "lint context",
			context:   workunit.ContextLint,
			module:    "web-app",
			component: "typescript",
			tool:      "eslint",
		},
		{
			name:      "scan context",
			context:   workunit.ContextScan,
			module:    "eac-cli",
			component: "docker",
			tool:      "trivy-vuln",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			id := createTestUnitID(tt.context, tt.module, tt.component, tt.tool)
			manifest := createTestManifest(tt.context, tt.module, tt.component, tt.tool)

			dirName := tt.component + "_" + tt.tool
			expectedPath := filepath.Join(tmpDir, "out", string(tt.context), tt.module, dirName, "uow.manifest.json")

			// TODO: After implementation, verify each context works correctly
			_ = id
			_ = manifest
			_ = expectedPath
		})
	}
}

func TestUoWTracker_RecordComplete_WithFailedExecution(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextTest, "failing-module", "go", "gotest")

	manifest := createTestManifest(workunit.ContextTest, "failing-module", "go", "gotest")
	manifest.ExitCode = 1 // Non-zero exit code

	// RecordComplete should save manifests for failed executions too
	// TODO: After implementation, verify failed manifests are saved
	_ = tmpDir
	_ = id
	_ = manifest
}

// =============================================================================
// RecordCacheHit Tests
// =============================================================================

func TestUoWTracker_RecordCacheHit_ReturnsExistingManifest(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Create manifest on disk
	originalManifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	originalManifest.InputHash = "sha256:cache-test-input"
	createManifestOnDisk(t, tmpDir, originalManifest)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify returned manifest matches originalManifest
	_ = id
	_ = originalManifest
}

func TestUoWTracker_RecordCacheHit_ReturnsErrorWhenManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "nonexistent-module", "go", "go")

	// No manifest exists on disk
	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify error is returned
	// 3. Verify returned manifest is nil
	_ = tmpDir
	_ = id
}

func TestUoWTracker_RecordCacheHit_ReturnsErrorWhenArtifactsMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Create manifest that references non-existent artifacts
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: "eac-linux-amd64", SHA256: "sha256:nonexistent", Size: 1000, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify error is returned (cache invalid - artifacts missing)
	_ = id
}

func TestUoWTracker_RecordCacheHit_ReturnsErrorWhenArtifactsCorrupt(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

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
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: "eac-linux-amd64", SHA256: "sha256:wrong-hash", Size: int64(len(artifactContent)), Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify error is returned (cache invalid - hash mismatch)
	_ = id
}

func TestUoWTracker_RecordCacheHit_SucceedsWhenArtifactsValid(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

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
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: artifactRelPath, SHA256: hash, Size: size, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify no error is returned
	// 3. Verify returned manifest is valid
	_ = id
}

func TestUoWTracker_RecordCacheHit_ValidatesMultipleArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

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
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = artifacts
	createManifestOnDisk(t, tmpDir, manifest)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify no error is returned
	// 3. Verify returned manifest has all 3 artifacts
	_ = id
}

func TestUoWTracker_RecordCacheHit_FailsIfAnyArtifactMissing(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Create UoW directory
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	// Create only one artifact but reference two in manifest
	artifact1Content := []byte("existing binary content")
	hash1, size1 := createArtifactOnDisk(t, uowDir, "eac-linux-amd64", artifact1Content)

	// Create manifest referencing existing and missing artifacts
	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: hash1, Size: size1, Type: "binary"},
		{ID: "darwin", Path: "eac-darwin-amd64", SHA256: "sha256:missing", Size: 1000, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify error is returned (one artifact missing)
	_ = id
}

func TestUoWTracker_RecordCacheHit_AllContexts(t *testing.T) {
	contexts := []workunit.Context{
		workunit.ContextBuild,
		workunit.ContextTest,
		workunit.ContextLint,
		workunit.ContextScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			tmpDir := t.TempDir()
			id := createTestUnitID(ctx, "module", "component", "tool")

			// Create manifest for this context
			manifest := createTestManifest(ctx, "module", "component", "tool")
			createManifestOnDisk(t, tmpDir, manifest)

			// TODO: After implementation, verify RecordCacheHit works for all contexts
			_ = id
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
		createTestUnitID(workunit.ContextBuild, "module1", "go", "go"),
		createTestUnitID(workunit.ContextBuild, "module2", "go", "go"),
		createTestUnitID(workunit.ContextBuild, "module3", "go", "go"),
		createTestUnitID(workunit.ContextTest, "module1", "go", "gotest"),
		createTestUnitID(workunit.ContextLint, "module1", "go", "golangci-lint"),
	}

	// TODO: After implementation:
	// 1. Create tracker with tmpDir
	// 2. Call RecordStart concurrently for all IDs
	// 3. Verify all succeed without race conditions
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id workunit.UnitID) {
			defer wg.Done()
			// TODO: Call tracker.RecordStart(id)
			_ = id
		}(id)
	}
	wg.Wait()

	_ = tmpDir
}

func TestUoWTracker_ConcurrentRecordComplete(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple UnitIDs for different UoWs
	ids := []workunit.UnitID{
		createTestUnitID(workunit.ContextBuild, "module1", "go", "go"),
		createTestUnitID(workunit.ContextBuild, "module2", "go", "go"),
		createTestUnitID(workunit.ContextBuild, "module3", "go", "go"),
		createTestUnitID(workunit.ContextTest, "module1", "go", "gotest"),
		createTestUnitID(workunit.ContextLint, "module1", "go", "golangci-lint"),
	}

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id workunit.UnitID) {
			defer wg.Done()
			manifest := createTestManifest(id.Context, id.Module, id.Component, id.Tool)
			// TODO: Call tracker.RecordComplete(id, manifest)
			_ = manifest
		}(id)
	}
	wg.Wait()

	// TODO: Verify all manifests were written correctly
	_ = tmpDir
}

func TestUoWTracker_ConcurrentRecordCacheHit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple UnitIDs and their manifests on disk
	ids := []workunit.UnitID{
		createTestUnitID(workunit.ContextBuild, "module1", "go", "go"),
		createTestUnitID(workunit.ContextBuild, "module2", "go", "go"),
		createTestUnitID(workunit.ContextBuild, "module3", "go", "go"),
	}

	// Create manifests on disk for all IDs
	for _, id := range ids {
		manifest := createTestManifest(id.Context, id.Module, id.Component, id.Tool)
		createManifestOnDisk(t, tmpDir, manifest)
	}

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id workunit.UnitID) {
			defer wg.Done()
			// TODO: Call tracker.RecordCacheHit(id)
			// Verify no errors and manifest is returned
		}(id)
	}
	wg.Wait()
}

func TestUoWTracker_ConcurrentMixedOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some manifests on disk for cache hits
	cachedID := createTestUnitID(workunit.ContextBuild, "cached-module", "go", "go")
	cachedManifest := createTestManifest(workunit.ContextBuild, "cached-module", "go", "go")
	createManifestOnDisk(t, tmpDir, cachedManifest)

	// IDs for new work
	newID1 := createTestUnitID(workunit.ContextBuild, "new-module1", "go", "go")
	newID2 := createTestUnitID(workunit.ContextBuild, "new-module2", "go", "go")

	var wg sync.WaitGroup

	// RecordStart for new work
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: tracker.RecordStart(newID1)
		_ = newID1
	}()

	// RecordComplete for new work
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: tracker.RecordComplete(newID2, createTestManifest(...))
		_ = newID2
	}()

	// RecordCacheHit for cached work
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: tracker.RecordCacheHit(cachedID)
		_ = cachedID
	}()

	wg.Wait()

	_ = tmpDir
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestUoWTracker_RecordComplete_WithEmptyArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextLint, "test-module", "go", "golangci-lint")

	manifest := createTestManifest(workunit.ContextLint, "test-module", "go", "golangci-lint")
	manifest.Artifacts = []Artifact{} // No artifacts for lint

	// TODO: After implementation, verify empty artifacts list is saved correctly
	_ = tmpDir
	_ = id
	_ = manifest
}

func TestUoWTracker_RecordComplete_WithNilArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextLint, "test-module", "go", "golangci-lint")

	manifest := createTestManifest(workunit.ContextLint, "test-module", "go", "golangci-lint")
	manifest.Artifacts = nil // Nil artifacts

	// TODO: After implementation, verify nil artifacts is handled correctly
	_ = tmpDir
	_ = id
	_ = manifest
}

func TestUoWTracker_RecordCacheHit_ManifestExistsButInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Create manifest directory and write invalid JSON
	dirName := "go_go"
	manifestDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	err = os.WriteFile(manifestPath, []byte("invalid json {{{"), 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify error is returned (invalid manifest)
	_ = id
}

func TestUoWTracker_RecordCacheHit_ManifestExistsButEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// Create manifest directory and write empty file
	dirName := "go_go"
	manifestDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	err = os.WriteFile(manifestPath, []byte(""), 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify error is returned (empty manifest)
	_ = id
}

func TestUoWTracker_WorksWithSpecialCharactersInNames(t *testing.T) {
	tmpDir := t.TempDir()

	// Module and component names with hyphens (common pattern)
	id := createTestUnitID(workunit.ContextBuild, "my-complex-module", "go-component", "my-tool")
	manifest := createTestManifest(workunit.ContextBuild, "my-complex-module", "go-component", "my-tool")

	// Expected path: out/build/my-complex-module/go-component_my-tool/uow.manifest.json
	expectedDir := filepath.Join(tmpDir, "out", "build", "my-complex-module", "go-component_my-tool")

	// TODO: After implementation, verify special characters in names work correctly
	_ = id
	_ = manifest
	_ = expectedDir
}

func TestUoWTracker_RecordCacheHit_WithEmptyArtifactsList(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextLint, "test-module", "go", "golangci-lint")

	// Create manifest with empty artifacts (valid for lint operations)
	manifest := createTestManifest(workunit.ContextLint, "test-module", "go", "golangci-lint")
	manifest.Artifacts = []Artifact{}
	createManifestOnDisk(t, tmpDir, manifest)

	// TODO: After implementation:
	// 1. Call tracker.RecordCacheHit(id)
	// 2. Verify no error (empty artifacts is valid)
	// 3. Verify manifest is returned correctly
	_ = id
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestUoWTracker_FullWorkflow_Success(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// TODO: After implementation:
	// 1. RecordStart(id)
	// 2. Simulate work (create artifacts)
	// 3. RecordComplete(id, manifest) with artifacts
	// 4. Verify manifest on disk

	// Create UoW directory and artifacts
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	// Create artifact
	artifactContent := []byte("built binary content")
	hash, size := createArtifactOnDisk(t, uowDir, "eac-linux-amd64", artifactContent)

	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: hash, Size: size, Type: "binary"},
	}

	_ = id
	_ = manifest
}

func TestUoWTracker_FullWorkflow_CacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// TODO: After implementation:
	// 1. First run: RecordStart, create artifacts, RecordComplete
	// 2. Second run: RecordCacheHit should succeed

	// Simulate existing build with valid artifacts
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	artifactContent := []byte("cached binary content")
	hash, size := createArtifactOnDisk(t, uowDir, "eac-linux-amd64", artifactContent)

	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: hash, Size: size, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// Now RecordCacheHit should succeed
	_ = id
}

func TestUoWTracker_FullWorkflow_CacheMiss(t *testing.T) {
	tmpDir := t.TempDir()
	id := createTestUnitID(workunit.ContextBuild, "test-module", "go", "go")

	// TODO: After implementation:
	// 1. Create manifest but delete an artifact
	// 2. RecordCacheHit should return error (cache miss)
	// 3. Caller should then RecordStart and rebuild

	// Create manifest with artifact reference
	dirName := "go_go"
	uowDir := filepath.Join(tmpDir, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	manifest := createTestManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "linux", Path: "eac-linux-amd64", SHA256: "sha256:deleted-artifact", Size: 1000, Type: "binary"},
	}
	createManifestOnDisk(t, tmpDir, manifest)

	// Artifact does not exist - cache should be invalid
	_ = id
}

// =============================================================================
// Table-Driven Tests
// =============================================================================

func TestUoWTracker_RecordComplete_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		context   workunit.Context
		module    string
		component string
		tool      string
		exitCode  int
		artifacts []Artifact
	}{
		{
			name:      "successful build with artifacts",
			context:   workunit.ContextBuild,
			module:    "eac-cli",
			component: "go",
			tool:      "go",
			exitCode:  0,
			artifacts: []Artifact{
				{ID: "binary", Path: "eac", SHA256: "sha256:abc", Size: 1000, Type: "binary"},
			},
		},
		{
			name:      "failed test with report",
			context:   workunit.ContextTest,
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
			context:   workunit.ContextLint,
			module:    "web-app",
			component: "typescript",
			tool:      "eslint",
			exitCode:  0,
			artifacts: []Artifact{},
		},
		{
			name:      "scan with multiple artifacts",
			context:   workunit.ContextScan,
			module:    "eac-cli",
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
				Context:    tt.context,
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

			dirName := tt.component + "_" + tt.tool
			expectedPath := filepath.Join(tmpDir, "out", string(tt.context), tt.module, dirName, "uow.manifest.json")

			// TODO: After implementation:
			// 1. Call tracker.RecordComplete(id, manifest)
			// 2. Verify file exists at expectedPath
			// 3. Load and verify content
			_ = id
			_ = manifest
			_ = expectedPath
		})
	}
}

func TestUoWTracker_RecordCacheHit_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		setupManifest   *UoWManifest
		setupArtifacts  bool
		expectError     bool
		errorContains   string
	}{
		{
			name: "valid manifest with valid artifacts",
			setupManifest: &UoWManifest{
				Context:   workunit.ContextBuild,
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
				Context:   workunit.ContextBuild,
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
				Context:   workunit.ContextLint,
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
				id = createTestUnitID(tt.setupManifest.Context, tt.setupManifest.Module, tt.setupManifest.Component, tt.setupManifest.Tool)

				// Set up UoW directory
				dirName := tt.setupManifest.Component + "_" + tt.setupManifest.Tool
				uowDir := filepath.Join(tmpDir, "out", string(tt.setupManifest.Context), tt.setupManifest.Module, dirName)
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
				id = createTestUnitID(workunit.ContextBuild, "nonexistent", "go", "go")
			}

			// TODO: After implementation:
			// 1. Call tracker.RecordCacheHit(id)
			// 2. Check error based on tt.expectError
			// 3. If error expected, check it contains tt.errorContains
			_ = id
		})
	}
}
