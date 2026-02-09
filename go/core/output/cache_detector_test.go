package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// UoWChangeResult Type Tests
// =============================================================================

func TestUoWChangeResult_FieldsExist(t *testing.T) {
	// Verify UoWChangeResult has required fields
	result := &UoWChangeResult{
		Changed:       []workunit.UnitID{},
		UpToDate:      []workunit.UnitID{},
		ChangeReasons: map[string]string{},
		FreshRun:      false,
		DetectionTime: time.Duration(0),
	}

	assert.NotNil(t, result.Changed)
	assert.NotNil(t, result.UpToDate)
	assert.NotNil(t, result.ChangeReasons)
	assert.False(t, result.FreshRun)
	assert.Zero(t, result.DetectionTime)
}

// =============================================================================
// DetectUoWChanges Tests
// =============================================================================

func TestDetectUoWChanges_FreshRun_NoManifests(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:current-hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.True(t, result.FreshRun, "should detect fresh run when no manifests exist")
	assert.Len(t, result.Changed, 2, "all UoWs should be changed on fresh run")
	assert.Empty(t, result.UpToDate, "no UoWs should be up-to-date on fresh run")

	// Verify change reasons
	for _, id := range expectedUoWs {
		reason, ok := result.ChangeReasons[id.Longname()]
		assert.True(t, ok, "should have reason for %s", id.Longname())
		assert.Contains(t, reason, "fresh run")
	}
}

func TestDetectUoWChanges_AllCached_HashesMatch(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests with specific input hashes
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	m2 := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	m2.InputHash = "sha256:hash-docker"
	f.saveManifest(m2)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	// Return matching hashes
	getInputHash := func(id workunit.UnitID) (string, error) {
		if id.ComponentName == "go" {
			return "sha256:hash-go", nil
		}
		return "sha256:hash-docker", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Empty(t, result.Changed, "no UoWs should be changed when hashes match")
	assert.Len(t, result.UpToDate, 2, "all UoWs should be up-to-date")
}

func TestDetectUoWChanges_PartialCached_SomeHashesDiffer(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests with specific input hashes
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	m2 := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	m2.InputHash = "sha256:hash-docker-old" // Will differ
	f.saveManifest(m2)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	// Go hash matches, docker hash differs
	getInputHash := func(id workunit.UnitID) (string, error) {
		if id.ComponentName == "go" {
			return "sha256:hash-go", nil
		}
		return "sha256:hash-docker-new", nil // Changed!
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun)
	assert.Len(t, result.Changed, 1, "one UoW should be changed")
	assert.Len(t, result.UpToDate, 1, "one UoW should be up-to-date")

	// Docker should be changed
	assert.Equal(t, "docker", result.Changed[0].ComponentName)
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "source changed")

	// Go should be up-to-date
	assert.Equal(t, "go", result.UpToDate[0].ComponentName)
}

func TestDetectUoWChanges_MissingManifest_MarkedAsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create only one manifest
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)
	// docker manifest is missing

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "core", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-go", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.False(t, result.FreshRun, "not a fresh run since go manifest exists")
	assert.Len(t, result.Changed, 1, "missing manifest should be changed")
	assert.Len(t, result.UpToDate, 1, "existing manifest should be up-to-date")

	// Docker should be changed due to missing manifest
	assert.Equal(t, "docker", result.Changed[0].ComponentName)
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "no prior manifest")
}

func TestDetectUoWChanges_PreviousFailure_MarkedAsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifest with exit code indicating failure
	m1 := f.createUoWManifestWithExitCode(core.ActionBuild, "core", "go", "go", 1)
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	// Hash matches but previous run failed
	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-go", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Len(t, result.Changed, 1, "previously failed UoW should be changed")
	assert.Empty(t, result.UpToDate)
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "previous failure")
}

func TestDetectUoWChanges_InvalidArtifacts_MarkedAsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifest with artifacts, but corrupt the artifacts
	m1 := f.createUoWManifestWithArtifacts(core.ActionBuild, "core", "go", "go", map[string][]byte{
		"binary": []byte("original content"),
	})
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	// Now corrupt the artifact
	artifactPath := filepath.Join(f.workspaceRoot, "out", "build", "core", "go-go", "binary")
	err := os.WriteFile(artifactPath, []byte("modified content"), 0644)
	require.NoError(t, err)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	// Hash matches
	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-go", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Len(t, result.Changed, 1, "UoW with corrupt artifacts should be changed")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "artifacts invalid")
}

func TestDetectUoWChanges_EmptyExpectedList_ReturnsEmpty(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	expectedUoWs := []workunit.UnitID{}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Empty(t, result.Changed)
	assert.Empty(t, result.UpToDate)
	assert.False(t, result.FreshRun)
}

func TestDetectUoWChanges_HashProviderError_MarkedAsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	// Hash provider returns error
	getInputHash := func(id workunit.UnitID) (string, error) {
		return "", assert.AnError
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Len(t, result.Changed, 1, "UoW with hash error should be changed")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "hash error")
}

func TestDetectUoWChanges_RecordsDetectionTime(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	// Detection time can be 0 on fast machines - just verify it's non-negative
	assert.GreaterOrEqual(t, result.DetectionTime, time.Duration(0), "detection time should be non-negative")
}

func TestDetectUoWChanges_AllContexts(t *testing.T) {
	contexts := []core.ActionType{
		core.ActionBuild,
		core.ActionTest,
		core.ActionLint,
		core.ActionScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			f := newTestFixture(t)
			reader := NewReader(f.workspaceRoot)

			m1 := f.createUoWManifest(ctx, "core", "go", "tool")
			m1.InputHash = "sha256:hash"
			f.saveManifest(m1)

			expectedUoWs := []workunit.UnitID{
				{Action: ctx, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "tool"},
			}

			getInputHash := func(id workunit.UnitID) (string, error) {
				return "sha256:hash", nil
			}

			result, err := reader.DetectUoWChanges(ctx, expectedUoWs, getInputHash, nil)
			require.NoError(t, err)

			assert.Empty(t, result.Changed)
			assert.Len(t, result.UpToDate, 1)
		})
	}
}

// =============================================================================
// IsModuleChanged Tests
// =============================================================================

func TestIsModuleChanged_NoManifests_ReturnsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	changed, reason, err := reader.IsModuleChanged(core.ActionBuild, "core", getInputHash)
	require.NoError(t, err)

	assert.True(t, changed, "module with no manifests should be changed")
	assert.Contains(t, reason, "no prior manifest")
}

func TestIsModuleChanged_AllUoWsCached_ReturnsNotChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests with matching hashes
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	m2 := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	m2.InputHash = "sha256:hash-docker"
	f.saveManifest(m2)

	getInputHash := func(id workunit.UnitID) (string, error) {
		if id.ComponentName == "go" {
			return "sha256:hash-go", nil
		}
		return "sha256:hash-docker", nil
	}

	changed, reason, err := reader.IsModuleChanged(core.ActionBuild, "core", getInputHash)
	require.NoError(t, err)

	assert.False(t, changed, "module with all cached UoWs should not be changed")
	assert.Empty(t, reason)
}

func TestIsModuleChanged_OneUoWChanged_ReturnsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests - one will have different hash
	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	m2 := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	m2.InputHash = "sha256:hash-docker-old"
	f.saveManifest(m2)

	getInputHash := func(id workunit.UnitID) (string, error) {
		if id.ComponentName == "go" {
			return "sha256:hash-go", nil
		}
		return "sha256:hash-docker-new", nil // Different!
	}

	changed, reason, err := reader.IsModuleChanged(core.ActionBuild, "core", getInputHash)
	require.NoError(t, err)

	assert.True(t, changed, "module with one changed UoW should be changed")
	assert.Contains(t, reason, "source changed")
}

func TestIsModuleChanged_PreviousFailure_ReturnsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifest with failure exit code
	m1 := f.createUoWManifestWithExitCode(core.ActionBuild, "core", "go", "go", 1)
	m1.InputHash = "sha256:hash-go"
	f.saveManifest(m1)

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-go", nil
	}

	changed, reason, err := reader.IsModuleChanged(core.ActionBuild, "core", getInputHash)
	require.NoError(t, err)

	assert.True(t, changed, "module with failed UoW should be changed")
	assert.Contains(t, reason, "previous failure")
}

func TestIsModuleChanged_DifferentContext_IndependentResults(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create build manifest (cached)
	buildManifest := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	buildManifest.InputHash = "sha256:hash"
	f.saveManifest(buildManifest)

	// Test context has no manifests

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	// Build context should be cached
	buildChanged, _, err := reader.IsModuleChanged(core.ActionBuild, "core", getInputHash)
	require.NoError(t, err)
	assert.False(t, buildChanged, "build should be cached")

	// Test context should be changed (no manifests)
	testChanged, _, err := reader.IsModuleChanged(core.ActionTest, "core", getInputHash)
	require.NoError(t, err)
	assert.True(t, testChanged, "test should be changed (no manifests)")
}

// =============================================================================
// Test Fixture Helpers (additional)
// =============================================================================

// saveManifest writes a modified manifest back to disk
// Uses the new flat directory format: component[-extra1][-extra2]
func (f *testFixture) saveManifest(manifest *UoWManifest) {
	f.t.Helper()

	// Use the new flat directory format
	dirName := manifest.DirName()
	manifestDir := filepath.Join(f.workspaceRoot, "out", string(manifest.Action), manifest.Module, dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(f.t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(f.t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(f.t, err)
}

// =============================================================================
// UoWManifest Metadata Tests
// =============================================================================

func TestUoWManifest_MetadataField_Serialization(t *testing.T) {
	manifest := &UoWManifest{
		Action:    core.ActionTest,
		Module:    "core",
		Component: "go",
		Tool:      "gotest",
		ExitCode:  0,
		InputHash: "sha256:hash",
		Metadata: map[string]string{
			"testset":  "unit",
			"build_id": "abc123",
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(manifest)
	require.NoError(t, err)

	// Deserialize
	var loaded UoWManifest
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, "unit", loaded.Metadata["testset"])
	assert.Equal(t, "abc123", loaded.Metadata["build_id"])
}

func TestUoWManifest_MetadataField_OmittedWhenEmpty(t *testing.T) {
	manifest := &UoWManifest{
		Action:    core.ActionBuild,
		Module:    "core",
		Component: "go",
		Tool:      "go",
		ExitCode:  0,
		InputHash: "sha256:hash",
		// No Metadata
	}

	data, err := json.Marshal(manifest)
	require.NoError(t, err)

	// Should not contain "metadata" key when empty
	assert.NotContains(t, string(data), "metadata")
}

func TestUoWManifest_MetadataField_LoadFromDisk(t *testing.T) {
	f := newTestFixture(t)

	// Create manifest with metadata
	manifest := &UoWManifest{
		Action:     core.ActionTest,
		Module:     "core",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   0,
		InputHash:  "sha256:hash",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		OutputHash: "sha256:output",
		Version:    "1.0.0",
		Metadata: map[string]string{
			"testset":  "integration",
			"build_id": "def456",
		},
	}
	f.saveManifest(manifest)

	// Load using reader
	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{
		Action:        core.ActionTest,
		Module:        "core",
		ComponentType: "go",
		ComponentName: "go",
		Tool:          "gotest",
	}
	loaded, err := reader.GetUoW(id)
	require.NoError(t, err)

	assert.Equal(t, "integration", loaded.Metadata["testset"])
	assert.Equal(t, "def456", loaded.Metadata["build_id"])
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestDetectUoWChanges_ConcurrentSafe(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifests
	for i := 0; i < 10; i++ {
		m := f.createUoWManifest(core.ActionBuild, "core", "comp"+string(rune('0'+i)), "tool")
		m.InputHash = "sha256:hash"
		f.saveManifest(m)
	}

	var expectedUoWs []workunit.UnitID
	for i := 0; i < 10; i++ {
		expectedUoWs = append(expectedUoWs, workunit.UnitID{
			Action:        core.ActionBuild,
			Module:        "core",
			ComponentType: "comp" + string(rune('0'+i)),
			ComponentName: "comp" + string(rune('0'+i)),
			Tool:          "tool",
		})
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	// Run detection multiple times concurrently
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
			assert.NoError(t, err)
			assert.Len(t, result.UpToDate, 10)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestDetectUoWChanges_NoArtifacts_StillValid(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create manifest with no artifacts (valid for lint)
	m1 := f.createUoWManifest(core.ActionLint, "core", "go", "golangci-lint")
	m1.InputHash = "sha256:hash"
	m1.Artifacts = []Artifact{} // Empty artifacts
	f.saveManifest(m1)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionLint, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "golangci-lint"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionLint, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Empty(t, result.Changed, "UoW with no artifacts but matching hash should be cached")
	assert.Len(t, result.UpToDate, 1)
}

func TestDetectUoWChanges_NilHashProvider_AllChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	m1 := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	m1.InputHash = "sha256:hash"
	f.saveManifest(m1)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	// nil hash provider - should treat as hash error or always changed
	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, nil, nil)
	require.NoError(t, err)

	// With nil hash provider, we can't verify the hash, so should be changed
	assert.Len(t, result.Changed, 1)
}

// =============================================================================
// Integration with Existing Reader Tests
// =============================================================================

func TestDetectUoWChanges_ConsistentWithValidateUoW(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create valid manifest with artifacts
	m1 := f.createUoWManifestWithArtifacts(core.ActionBuild, "core", "go", "go", map[string][]byte{
		"binary": []byte("content"),
	})
	m1.InputHash = "sha256:hash"
	f.saveManifest(m1)

	// ValidateUoW should return valid
	id := workunit.UnitID{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"}
	validationResult := reader.ValidateUoW(id)
	assert.True(t, validationResult.Valid, "ValidateUoW should return valid")

	// DetectUoWChanges should return up-to-date
	expectedUoWs := []workunit.UnitID{id}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	changeResult, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Empty(t, changeResult.Changed, "DetectUoWChanges should be consistent with ValidateUoW")
	assert.Len(t, changeResult.UpToDate, 1)
}

// =============================================================================
// NoOp Manifest Tests
// =============================================================================

func TestDetectUoWChanges_NoOpManifest_AlwaysUpToDate(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create NoOp manifest (module with no buildable components)
	m1 := f.createUoWManifest(core.ActionBuild, "templates", "none", "")
	m1.NoOp = true
	m1.InputHash = "" // NoOp UoWs have empty input hash
	m1.OutputHash = ""
	m1.Artifacts = nil
	m1.Metadata = map[string]string{"reason": "no buildable components"}
	f.saveManifest(m1)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "templates", ComponentType: "none", ComponentName: "none", Tool: ""},
	}

	// Hash provider shouldn't even be called for NoOp UoWs
	hashCalled := false
	getInputHash := func(id workunit.UnitID) (string, error) {
		hashCalled = true
		return "sha256:should-not-matter", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.False(t, hashCalled, "hash provider should not be called for NoOp UoWs")
	assert.Empty(t, result.Changed, "NoOp UoW should be up-to-date")
	assert.Len(t, result.UpToDate, 1, "NoOp UoW should be in up-to-date list")
}

func TestDetectUoWChanges_NoOpManifest_WithFailure_MarkedAsChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create NoOp manifest with failure exit code
	m1 := f.createUoWManifestWithExitCode(core.ActionBuild, "templates", "none", "", 1)
	m1.NoOp = true
	m1.InputHash = ""
	f.saveManifest(m1)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "templates", ComponentType: "none", ComponentName: "none", Tool: ""},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Len(t, result.Changed, 1, "failed NoOp UoW should be marked as changed")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "previous failure")
}

func TestIsModuleChanged_NoOpManifest_ReturnsNotChanged(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create NoOp manifest
	m1 := f.createUoWManifest(core.ActionBuild, "templates", "none", "")
	m1.NoOp = true
	m1.InputHash = ""
	f.saveManifest(m1)

	// nil hash provider - should still work for NoOp
	changed, reason, err := reader.IsModuleChanged(core.ActionBuild, "templates", nil)
	require.NoError(t, err)

	assert.False(t, changed, "module with NoOp UoW should not be changed")
	assert.Empty(t, reason)
}

func TestDetectUoWChanges_MixedNoOpAndRegular(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Create a NoOp manifest
	m1 := f.createUoWManifest(core.ActionTest, "templates", "none", "none")
	m1.NoOp = true
	m1.InputHash = ""
	m1.Metadata = map[string]string{"reason": "no tests for module"}
	f.saveManifest(m1)

	// Create a regular manifest
	m2 := f.createUoWManifest(core.ActionTest, "core", "go", "gotest")
	m2.InputHash = "sha256:hash-core"
	f.saveManifest(m2)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionTest, Module: "templates", ComponentType: "none", ComponentName: "none", Tool: "none"},
		{Action: core.ActionTest, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "gotest"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		if id.Module == "core" {
			return "sha256:hash-core", nil // Matches
		}
		return "", nil // Shouldn't be called for NoOp
	}

	result, err := reader.DetectUoWChanges(core.ActionTest, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Empty(t, result.Changed, "both UoWs should be up-to-date")
	assert.Len(t, result.UpToDate, 2)
}

func TestDetectUoWChanges_NoOpInAllContexts(t *testing.T) {
	contexts := []struct {
		ctx    core.ActionType
		reason string
	}{
		{core.ActionBuild, "no buildable components"},
		{core.ActionTest, "no tests for module"},
		{core.ActionLint, "no lintable components"},
		{core.ActionScan, "component type not scannable"},
	}

	for _, tc := range contexts {
		t.Run(string(tc.ctx), func(t *testing.T) {
			f := newTestFixture(t)
			reader := NewReader(f.workspaceRoot)

			m1 := f.createUoWManifest(tc.ctx, "placeholder", "none", "none")
			m1.NoOp = true
			m1.InputHash = ""
			m1.Metadata = map[string]string{"reason": tc.reason}
			f.saveManifest(m1)

			expectedUoWs := []workunit.UnitID{
				{Action: tc.ctx, Module: "placeholder", ComponentType: "none", ComponentName: "none", Tool: "none"},
			}

			result, err := reader.DetectUoWChanges(tc.ctx, expectedUoWs, nil, nil)
			require.NoError(t, err)

			assert.Empty(t, result.Changed, "NoOp UoW in %s context should be up-to-date", tc.ctx)
			assert.Len(t, result.UpToDate, 1)
		})
	}
}

// =============================================================================
// Build Invalidation Tests
// =============================================================================

func TestDetectUoWChanges_TestInvalidatedByNewerBuild(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD test manifest (ran 1 hour ago)
	testManifest := f.createUoWManifest(core.ActionTest, "core", "go", "gotest")
	testManifest.InputHash = "sha256:hash"
	testManifest.ExecutedAt = oldTime
	f.saveManifest(testManifest)

	// Create a NEW build manifest (ran just now - simulates build --skip-cache)
	buildManifest := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	buildManifest.InputHash = "sha256:hash"
	buildManifest.ExecutedAt = newTime
	f.saveManifest(buildManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionTest, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "gotest"},
	}

	// Hash matches - normally would be cached
	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionTest, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	// Test should be invalidated because build is newer
	assert.Len(t, result.Changed, 1, "test should be invalidated by newer build")
	assert.Empty(t, result.UpToDate)
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "build invalidated")
}

func TestDetectUoWChanges_TestNotInvalidatedByOlderBuild(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create a NEW test manifest (ran just now)
	testManifest := f.createUoWManifest(core.ActionTest, "core", "go", "gotest")
	testManifest.InputHash = "sha256:hash"
	testManifest.ExecutedAt = newTime
	f.saveManifest(testManifest)

	// Create an OLD build manifest (ran 1 hour ago)
	buildManifest := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	buildManifest.InputHash = "sha256:hash"
	buildManifest.ExecutedAt = oldTime
	f.saveManifest(buildManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionTest, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "gotest"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionTest, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	// Test should NOT be invalidated because build is older
	assert.Empty(t, result.Changed, "test should NOT be invalidated by older build")
	assert.Len(t, result.UpToDate, 1)
}

func TestDetectUoWChanges_LintInvalidatedByNewerBuild(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD lint manifest
	lintManifest := f.createUoWManifest(core.ActionLint, "core", "go", "golangci-lint")
	lintManifest.InputHash = "sha256:hash"
	lintManifest.ExecutedAt = oldTime
	f.saveManifest(lintManifest)

	// Create a NEW build manifest
	buildManifest := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	buildManifest.InputHash = "sha256:hash"
	buildManifest.ExecutedAt = newTime
	f.saveManifest(buildManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionLint, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "golangci-lint"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionLint, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Len(t, result.Changed, 1, "lint should be invalidated by newer build")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "build invalidated")
}

func TestDetectUoWChanges_ScanInvalidatedByNewerBuild(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD scan manifest
	scanManifest := f.createUoWManifest(core.ActionScan, "core", "go", "trivy")
	scanManifest.InputHash = "sha256:hash"
	scanManifest.ExecutedAt = oldTime
	f.saveManifest(scanManifest)

	// Create a NEW build manifest
	buildManifest := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	buildManifest.InputHash = "sha256:hash"
	buildManifest.ExecutedAt = newTime
	f.saveManifest(buildManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionScan, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "trivy"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionScan, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	assert.Len(t, result.Changed, 1, "scan should be invalidated by newer build")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "build invalidated")
}

func TestDetectUoWChanges_BuildNotInvalidatedByBuild(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	// Build context should NOT check for build invalidation (would be circular)
	oldTime := time.Now().Add(-1 * time.Hour)

	// Create build manifest
	buildManifest := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	buildManifest.InputHash = "sha256:hash"
	buildManifest.ExecutedAt = oldTime
	f.saveManifest(buildManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	// Build should be cached (no self-invalidation)
	assert.Empty(t, result.Changed)
	assert.Len(t, result.UpToDate, 1)
}

func TestDetectUoWChanges_TestInvalidatedByAnyBuildComponent(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD test manifest
	testManifest := f.createUoWManifest(core.ActionTest, "core", "go", "gotest")
	testManifest.InputHash = "sha256:hash"
	testManifest.ExecutedAt = oldTime
	f.saveManifest(testManifest)

	// Create an OLD build manifest for go component
	buildGoManifest := f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	buildGoManifest.InputHash = "sha256:hash"
	buildGoManifest.ExecutedAt = oldTime
	f.saveManifest(buildGoManifest)

	// Create a NEW build manifest for docker component (different component, same module)
	buildDockerManifest := f.createUoWManifest(core.ActionBuild, "core", "docker", "docker")
	buildDockerManifest.InputHash = "sha256:hash"
	buildDockerManifest.ExecutedAt = newTime
	f.saveManifest(buildDockerManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionTest, Module: "core", ComponentType: "go", ComponentName: "go", Tool: "gotest"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	result, err := reader.DetectUoWChanges(core.ActionTest, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	// Test should be invalidated because ANY build component in the module is newer
	assert.Len(t, result.Changed, 1, "test should be invalidated by newer build of any component")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "build invalidated")
}

// =============================================================================
// Cross-Module Build Dependency Invalidation Tests
// =============================================================================

func TestDetectUoWChanges_BuildInvalidatedByDependencyModuleBuild(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD build manifest for eac-ext (dependent module)
	extManifest := f.createUoWManifest(core.ActionBuild, "eac-ext", "dockerfile", "buildx")
	extManifest.InputHash = "sha256:hash-ext"
	extManifest.ExecutedAt = oldTime
	f.saveManifest(extManifest)

	// Create a NEW build manifest for eac-cli (dependency module)
	cliManifest := f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "go")
	cliManifest.InputHash = "sha256:hash-cli"
	cliManifest.ExecutedAt = newTime
	f.saveManifest(cliManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac-ext", ComponentType: "dockerfile", ComponentName: "dockerfile", Tool: "buildx"},
	}

	// Hash matches - normally would be cached
	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-ext", nil
	}

	// Dependency resolver: eac-ext depends on eac-cli
	depResolver := func(module string) []string {
		if module == "eac-ext" {
			return []string{"eac-cli"}
		}
		return nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, depResolver)
	require.NoError(t, err)

	// eac-ext should be invalidated because eac-cli was rebuilt more recently
	assert.Len(t, result.Changed, 1, "build should be invalidated by newer dependency build")
	assert.Empty(t, result.UpToDate)
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "dependency build invalidated")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "eac-cli")
}

func TestDetectUoWChanges_BuildNotInvalidatedByOlderDependencyBuild(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create a NEW build manifest for eac-ext (dependent - built after dep)
	extManifest := f.createUoWManifest(core.ActionBuild, "eac-ext", "dockerfile", "buildx")
	extManifest.InputHash = "sha256:hash-ext"
	extManifest.ExecutedAt = newTime
	f.saveManifest(extManifest)

	// Create an OLD build manifest for eac-cli (dependency - built before)
	cliManifest := f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "go")
	cliManifest.InputHash = "sha256:hash-cli"
	cliManifest.ExecutedAt = oldTime
	f.saveManifest(cliManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac-ext", ComponentType: "dockerfile", ComponentName: "dockerfile", Tool: "buildx"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-ext", nil
	}

	depResolver := func(module string) []string {
		if module == "eac-ext" {
			return []string{"eac-cli"}
		}
		return nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, depResolver)
	require.NoError(t, err)

	// eac-ext should NOT be invalidated because eac-cli build is older
	assert.Empty(t, result.Changed, "build should NOT be invalidated by older dependency build")
	assert.Len(t, result.UpToDate, 1)
}

func TestDetectUoWChanges_NilDependencyResolver_NoInvalidation(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD build manifest for eac-ext
	extManifest := f.createUoWManifest(core.ActionBuild, "eac-ext", "dockerfile", "buildx")
	extManifest.InputHash = "sha256:hash-ext"
	extManifest.ExecutedAt = oldTime
	f.saveManifest(extManifest)

	// Create a NEW build manifest for eac-cli
	cliManifest := f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "go")
	cliManifest.InputHash = "sha256:hash-cli"
	cliManifest.ExecutedAt = newTime
	f.saveManifest(cliManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac-ext", ComponentType: "dockerfile", ComponentName: "dockerfile", Tool: "buildx"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-ext", nil
	}

	// nil depResolver - should behave like before (no cross-module check)
	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, nil)
	require.NoError(t, err)

	// Should be cached since nil resolver means no cross-module check
	assert.Empty(t, result.Changed, "nil depResolver should not trigger cross-module invalidation")
	assert.Len(t, result.UpToDate, 1)
}

func TestDetectUoWChanges_MultipleDependencies_AnyNewerInvalidates(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD build manifest for dependent module
	depManifest := f.createUoWManifest(core.ActionBuild, "my-image", "dockerfile", "buildx")
	depManifest.InputHash = "sha256:hash-img"
	depManifest.ExecutedAt = oldTime
	f.saveManifest(depManifest)

	// dep-a is OLD (not a trigger)
	depAManifest := f.createUoWManifest(core.ActionBuild, "dep-a", "go", "go")
	depAManifest.InputHash = "sha256:hash-a"
	depAManifest.ExecutedAt = oldTime
	f.saveManifest(depAManifest)

	// dep-b is NEW (trigger!)
	depBManifest := f.createUoWManifest(core.ActionBuild, "dep-b", "go", "go")
	depBManifest.InputHash = "sha256:hash-b"
	depBManifest.ExecutedAt = newTime
	f.saveManifest(depBManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "my-image", ComponentType: "dockerfile", ComponentName: "dockerfile", Tool: "buildx"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash-img", nil
	}

	depResolver := func(module string) []string {
		if module == "my-image" {
			return []string{"dep-a", "dep-b"}
		}
		return nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, depResolver)
	require.NoError(t, err)

	// Should be invalidated because dep-b is newer
	assert.Len(t, result.Changed, 1, "should be invalidated when any dependency is newer")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "dependency build invalidated")
	assert.Contains(t, result.ChangeReasons[result.Changed[0].Longname()], "dep-b")
}

func TestDetectUoWChanges_DependencyResolver_OnlyAffectsBuildContext(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)
	newTime := time.Now()

	// Create an OLD test manifest
	testManifest := f.createUoWManifest(core.ActionTest, "eac-ext", "go", "gotest")
	testManifest.InputHash = "sha256:hash"
	testManifest.ExecutedAt = oldTime
	f.saveManifest(testManifest)

	// Create a NEW build manifest for dependency module
	cliManifest := f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "go")
	cliManifest.InputHash = "sha256:hash-cli"
	cliManifest.ExecutedAt = newTime
	f.saveManifest(cliManifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionTest, Module: "eac-ext", ComponentType: "go", ComponentName: "go", Tool: "gotest"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	// Dep resolver provided but action is test, not build
	depResolver := func(module string) []string {
		if module == "eac-ext" {
			return []string{"eac-cli"}
		}
		return nil
	}

	result, err := reader.DetectUoWChanges(core.ActionTest, expectedUoWs, getInputHash, depResolver)
	require.NoError(t, err)

	// Test context should NOT use dependency resolver for cross-module check
	// (test/lint/scan use same-module build invalidation instead)
	assert.Empty(t, result.Changed, "dependency resolver should only affect build context")
	assert.Len(t, result.UpToDate, 1)
}

func TestDetectUoWChanges_DependencyWithNoDeps_NotInvalidated(t *testing.T) {
	f := newTestFixture(t)
	reader := NewReader(f.workspaceRoot)

	oldTime := time.Now().Add(-1 * time.Hour)

	// Create a build manifest for a module with no dependencies
	manifest := f.createUoWManifest(core.ActionBuild, "standalone", "go", "go")
	manifest.InputHash = "sha256:hash"
	manifest.ExecutedAt = oldTime
	f.saveManifest(manifest)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "standalone", ComponentType: "go", ComponentName: "go", Tool: "go"},
	}

	getInputHash := func(id workunit.UnitID) (string, error) {
		return "sha256:hash", nil
	}

	// Resolver returns nil for standalone (no dependencies)
	depResolver := func(module string) []string {
		return nil
	}

	result, err := reader.DetectUoWChanges(core.ActionBuild, expectedUoWs, getInputHash, depResolver)
	require.NoError(t, err)

	assert.Empty(t, result.Changed, "module with no dependencies should not be invalidated")
	assert.Len(t, result.UpToDate, 1)
}
