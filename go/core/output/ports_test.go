package output

import (
	"testing"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// OutputReaderAdapter Tests
// =============================================================================

func TestOutputReaderAdapter_GetUoW(t *testing.T) {
	f := newTestFixture(t)
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")

	reader := NewReader(f.workspaceRoot)
	adapter := NewOutputReaderAdapter(reader)

	// Use string context via the adapter
	manifest, err := adapter.GetUoW("build", "test-module", "go", "go")
	require.NoError(t, err)
	require.NotNil(t, manifest)

	assert.Equal(t, "build", manifest.GetContext())
	assert.Equal(t, "test-module", manifest.GetModule())
	assert.Equal(t, "go", manifest.GetComponent())
	assert.Equal(t, "go", manifest.GetTool())
}

func TestOutputReaderAdapter_GetModule(t *testing.T) {
	f := newTestFixture(t)
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "test-module", "docker", "docker")

	reader := NewReader(f.workspaceRoot)
	adapter := NewOutputReaderAdapter(reader)

	view, err := adapter.GetModule("build", "test-module")
	require.NoError(t, err)
	require.NotNil(t, view)

	assert.Equal(t, "test-module", view.GetModule())
	assert.Len(t, view.GetComponents(), 2)
}

func TestOutputReaderAdapter_ListUoWs(t *testing.T) {
	f := newTestFixture(t)
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "test-module", "docker", "docker")
	f.createUoWManifest(workunit.ContextBuild, "test-module", "docker", "buildx")

	reader := NewReader(f.workspaceRoot)
	adapter := NewOutputReaderAdapter(reader)

	manifests, err := adapter.ListUoWs("build", "test-module")
	require.NoError(t, err)
	assert.Len(t, manifests, 3)
}

func TestOutputReaderAdapter_ValidateUoW(t *testing.T) {
	f := newTestFixture(t)
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")

	reader := NewReader(f.workspaceRoot)
	adapter := NewOutputReaderAdapter(reader)

	result := adapter.ValidateUoW("build", "test-module", "go", "go")
	assert.True(t, result.IsValid())
	assert.True(t, result.HasManifest())
	assert.True(t, result.IsManifestValid())
}

func TestOutputReaderAdapter_ValidateUoW_Missing(t *testing.T) {
	f := newTestFixture(t)

	reader := NewReader(f.workspaceRoot)
	adapter := NewOutputReaderAdapter(reader)

	result := adapter.ValidateUoW("build", "nonexistent", "go", "go")
	assert.False(t, result.IsValid())
	assert.False(t, result.HasManifest())
}

func TestOutputReaderAdapter_ValidateModule(t *testing.T) {
	f := newTestFixture(t)
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "test-module", "docker", "docker")

	reader := NewReader(f.workspaceRoot)
	adapter := NewOutputReaderAdapter(reader)

	expectedUoWs := []interfaces.UnitIDPort{
		&testUnitID{context: "build", module: "test-module", component: "go", tool: "go"},
		&testUnitID{context: "build", module: "test-module", component: "docker", tool: "docker"},
	}

	result := adapter.ValidateModule("build", "test-module", expectedUoWs)
	assert.True(t, result.IsValid())
}

func TestOutputReaderAdapter_ValidateModule_Missing(t *testing.T) {
	f := newTestFixture(t)
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")
	// docker is missing

	reader := NewReader(f.workspaceRoot)
	adapter := NewOutputReaderAdapter(reader)

	expectedUoWs := []interfaces.UnitIDPort{
		&testUnitID{context: "build", module: "test-module", component: "go", tool: "go"},
		&testUnitID{context: "build", module: "test-module", component: "docker", tool: "docker"},
	}

	result := adapter.ValidateModule("build", "test-module", expectedUoWs)
	assert.False(t, result.IsValid())
}

// =============================================================================
// UoWTrackerAdapter Tests
// =============================================================================

func TestUoWTrackerAdapter_RecordStart(t *testing.T) {
	f := newTestFixture(t)

	tracker := NewTracker(f.workspaceRoot, workunit.ContextBuild)
	adapter := NewUoWTrackerAdapter(tracker)

	unitID := &testUnitID{context: "build", module: "test-module", component: "go", tool: "go"}

	err := adapter.RecordStart(unitID)
	require.NoError(t, err)
}

func TestUoWTrackerAdapter_RecordComplete(t *testing.T) {
	f := newTestFixture(t)

	tracker := NewTracker(f.workspaceRoot, workunit.ContextBuild)
	adapter := NewUoWTrackerAdapter(tracker)

	unitID := &testUnitID{context: "build", module: "test-module", component: "go", tool: "go"}

	err := adapter.RecordStart(unitID)
	require.NoError(t, err)

	manifest := &testManifest{
		context:    "build",
		module:     "test-module",
		component:  "go",
		tool:       "go",
		exitCode:   0,
		inputHash:  "sha256:input",
		executedAt: time.Now().UTC(),
		duration:   10 * time.Second,
		outputHash: "sha256:output",
	}

	err = adapter.RecordComplete(unitID, manifest)
	require.NoError(t, err)

	// Verify manifest was written
	reader := NewReader(f.workspaceRoot)
	loaded, err := reader.GetUoW(workunit.ContextBuild, "test-module", "go", "go")
	require.NoError(t, err)
	assert.Equal(t, "test-module", loaded.Module)
}

func TestUoWTrackerAdapter_RecordCacheHit(t *testing.T) {
	f := newTestFixture(t)

	// First create a manifest
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")

	tracker := NewTracker(f.workspaceRoot, workunit.ContextBuild)
	adapter := NewUoWTrackerAdapter(tracker)

	unitID := &testUnitID{context: "build", module: "test-module", component: "go", tool: "go"}

	manifest, err := adapter.RecordCacheHit(unitID)
	require.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.Equal(t, "test-module", manifest.GetModule())
}

func TestUoWTrackerAdapter_RecordCacheHit_Missing(t *testing.T) {
	f := newTestFixture(t)

	tracker := NewTracker(f.workspaceRoot, workunit.ContextBuild)
	adapter := NewUoWTrackerAdapter(tracker)

	unitID := &testUnitID{context: "build", module: "nonexistent", component: "go", tool: "go"}

	manifest, err := adapter.RecordCacheHit(unitID)
	assert.Error(t, err)
	assert.Nil(t, manifest)
}

// =============================================================================
// UoWManifest Port Implementation Tests
// =============================================================================

func TestUoWManifest_PortMethods(t *testing.T) {
	manifest := &UoWManifest{
		Context:    workunit.ContextBuild,
		Module:     "test-module",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:input",
		ExecutedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Duration:   30 * time.Second,
		Artifacts: []Artifact{
			{ID: "binary", Path: "bin/test", SHA256: "sha256:abc", Size: 1000, Type: "executable"},
		},
		OutputHash: "sha256:output",
	}

	// Test all port methods
	assert.Equal(t, "build", manifest.GetContext())
	assert.Equal(t, "test-module", manifest.GetModule())
	assert.Equal(t, "go", manifest.GetComponent())
	assert.Equal(t, "go", manifest.GetTool())
	assert.Equal(t, 0, manifest.GetExitCode())
	assert.Equal(t, "sha256:input", manifest.GetInputHash())
	assert.Equal(t, time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), manifest.GetExecutedAt())
	assert.Equal(t, 30*time.Second, manifest.GetDuration())
	assert.Len(t, manifest.GetArtifacts(), 1)
	assert.Equal(t, "sha256:output", manifest.GetOutputHash())
}

func TestArtifact_PortMethods(t *testing.T) {
	artifact := &Artifact{
		ID:     "binary",
		Path:   "bin/test",
		SHA256: "sha256:abc123",
		Size:   1024,
		Type:   "executable",
	}

	assert.Equal(t, "binary", artifact.GetID())
	assert.Equal(t, "bin/test", artifact.GetPath())
	assert.Equal(t, "sha256:abc123", artifact.GetSHA256())
	assert.Equal(t, int64(1024), artifact.GetSize())
	assert.Equal(t, "executable", artifact.GetType())
}

func TestModuleView_PortMethods(t *testing.T) {
	view := &ModuleView{
		Module: "test-module",
		Status: StatusCompleted,
		Components: []ComponentView{
			{Module: "test-module", Component: "go", Status: StatusCompleted},
		},
		TotalSize: 5000,
	}

	assert.Equal(t, "test-module", view.GetModule())
	assert.Equal(t, "completed", view.GetStatus())
	assert.Len(t, view.GetComponents(), 1)
	assert.Equal(t, int64(5000), view.GetTotalSize())
}

func TestComponentView_PortMethods(t *testing.T) {
	view := &ComponentView{
		Module:    "test-module",
		Component: "go",
		Status:    StatusCompleted,
		UoWs: []UoWManifest{
			{Context: workunit.ContextBuild, Module: "test-module", Component: "go", Tool: "go"},
		},
		TotalSize: 2000,
	}

	assert.Equal(t, "test-module", view.GetModule())
	assert.Equal(t, "go", view.GetComponent())
	assert.Equal(t, "completed", view.GetStatus())
	assert.Len(t, view.GetUoWs(), 1)
	assert.Equal(t, int64(2000), view.GetTotalSize())
}

func TestValidationResult_PortMethods(t *testing.T) {
	result := &ValidationResult{
		Valid:            true,
		ManifestExists:   true,
		ManifestValid:    true,
		ArtifactsValid:   true,
		MissingArtifacts: []string{},
		CorruptArtifacts: []string{},
		Error:            nil,
	}

	assert.True(t, result.IsValid())
	assert.True(t, result.HasManifest())
	assert.True(t, result.IsManifestValid())
	assert.True(t, result.AreArtifactsValid())
	assert.Empty(t, result.GetMissingArtifacts())
	assert.Empty(t, result.GetCorruptArtifacts())
	assert.Nil(t, result.GetError())
}

func TestValidationResult_PortMethods_Invalid(t *testing.T) {
	result := &ValidationResult{
		Valid:            false,
		ManifestExists:   false,
		ManifestValid:    false,
		ArtifactsValid:   false,
		MissingArtifacts: []string{"file1", "file2"},
		CorruptArtifacts: []string{"file3"},
		Error:            assert.AnError,
	}

	assert.False(t, result.IsValid())
	assert.False(t, result.HasManifest())
	assert.False(t, result.IsManifestValid())
	assert.False(t, result.AreArtifactsValid())
	assert.Equal(t, []string{"file1", "file2"}, result.GetMissingArtifacts())
	assert.Equal(t, []string{"file3"}, result.GetCorruptArtifacts())
	assert.NotNil(t, result.GetError())
}

// =============================================================================
// Test Helpers
// =============================================================================

// testUnitID implements UnitIDPort for testing.
type testUnitID struct {
	context   string
	module    string
	component string
	tool      string
}

func (t *testUnitID) GetContext() string   { return t.context }
func (t *testUnitID) GetModule() string    { return t.module }
func (t *testUnitID) GetComponent() string { return t.component }
func (t *testUnitID) GetTool() string      { return t.tool }
func (t *testUnitID) GetSpec() string      { return "" }
func (t *testUnitID) Shortname() string    { return t.module + ":" + t.component }
func (t *testUnitID) Longname() string {
	return t.context + ":" + t.module + ":" + t.component + ":" + t.tool
}
func (t *testUnitID) String() string { return t.Longname() }
func (t *testUnitID) OutDir() string { return "" }

// testManifest implements UoWManifestPort for testing.
type testManifest struct {
	context    string
	module     string
	component  string
	tool       string
	exitCode   int
	inputHash  string
	executedAt time.Time
	duration   time.Duration
	outputHash string
	artifacts  []interfaces.OutputArtifactPort
}

func (m *testManifest) GetContext() string                        { return m.context }
func (m *testManifest) GetModule() string                         { return m.module }
func (m *testManifest) GetComponent() string                      { return m.component }
func (m *testManifest) GetTool() string                           { return m.tool }
func (m *testManifest) GetExitCode() int                          { return m.exitCode }
func (m *testManifest) GetInputHash() string                      { return m.inputHash }
func (m *testManifest) GetExecutedAt() time.Time                  { return m.executedAt }
func (m *testManifest) GetDuration() time.Duration                { return m.duration }
func (m *testManifest) GetArtifacts() []interfaces.OutputArtifactPort { return m.artifacts }
func (m *testManifest) GetOutputHash() string                     { return m.outputHash }
