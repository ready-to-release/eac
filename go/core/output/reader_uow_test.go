package output

import (
	"os"
	"path/filepath"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GetComponent Tests
// =============================================================================

func TestOutputReader_GetComponent_AggregatesUoWsForComponent(t *testing.T) {
	f := newTestFixture(t)

	// Create multiple UoWs for the same component (different tools)
	f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	f.createUoWManifest(core.ActionBuild, "core", "go", "cgo")

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetComponent(core.ActionBuild, "core", "go")
	require.NoError(t, err)
	assert.Equal(t, "core", view.Module)
	assert.Equal(t, "go", view.Component)
	assert.Len(t, view.UoWs, 2)
}

func TestOutputReader_GetComponent_ComputesStatusFromExitCodes_AllSuccess(t *testing.T) {
	f := newTestFixture(t)

	// All UoWs successful
	f.createUoWManifestWithExitCode(core.ActionBuild, "core", "go", "go", 0)
	f.createUoWManifestWithExitCode(core.ActionBuild, "core", "go", "cgo", 0)

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetComponent(core.ActionBuild, "core", "go")
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, view.Status)
}

func TestOutputReader_GetComponent_ComputesStatusFromExitCodes_OneFailed(t *testing.T) {
	f := newTestFixture(t)

	// One UoW failed
	f.createUoWManifestWithExitCode(core.ActionBuild, "core", "go", "go", 0)
	f.createUoWManifestWithExitCode(core.ActionBuild, "core", "go", "cgo", 1)

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetComponent(core.ActionBuild, "core", "go")
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, view.Status)
}

func TestOutputReader_GetComponent_ComputesStatusFromExitCodes_AllFailed(t *testing.T) {
	f := newTestFixture(t)

	// All UoWs failed
	f.createUoWManifestWithExitCode(core.ActionTest, "core", "go", "gotest", 1)
	f.createUoWManifestWithExitCode(core.ActionTest, "core", "go", "godog", 2)

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetComponent(core.ActionTest, "core", "go")
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, view.Status)
}

func TestOutputReader_GetComponent_ReturnsErrorWhenNoUoWsExist(t *testing.T) {
	f := newTestFixture(t)

	// No UoWs exist for this component - GetComponent returns empty view, no error
	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetComponent(core.ActionBuild, "nonexistent", "go")
	require.NoError(t, err)
	assert.Equal(t, StatusPending, view.Status)
	assert.Empty(t, view.UoWs)
}

func TestOutputReader_GetComponent_SingleUoW(t *testing.T) {
	f := newTestFixture(t)

	// Single UoW for component
	f.createUoWManifest(core.ActionLint, "core", "go", "golangci-lint")

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetComponent(core.ActionLint, "core", "go")
	require.NoError(t, err)
	assert.Len(t, view.UoWs, 1)
}

func TestOutputReader_GetComponent_ComputesTotalSize(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs with artifacts of known sizes
	f.createUoWManifestWithArtifacts(core.ActionBuild, "core", "go", "go", map[string][]byte{
		"binary1": make([]byte, 1000),
		"binary2": make([]byte, 2000),
	})
	f.createUoWManifestWithArtifacts(core.ActionBuild, "core", "go", "cgo", map[string][]byte{
		"binary3": make([]byte, 3000),
	})

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetComponent(core.ActionBuild, "core", "go")
	require.NoError(t, err)
	assert.Equal(t, int64(6000), view.TotalSize)
}

// =============================================================================
// GetModule Tests
// =============================================================================

func TestOutputReader_GetModule_AggregatesComponentsForModule(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs for multiple components in the same module
	f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "go")
	f.createUoWManifest(core.ActionBuild, "eac-cli", "docker", "docker")
	f.createUoWManifest(core.ActionBuild, "eac-cli", "docker", "buildx")

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionBuild, "eac-cli")
	require.NoError(t, err)
	assert.Equal(t, "eac-cli", view.Module)
	assert.Len(t, view.Components, 2) // go, docker
}

func TestOutputReader_GetModule_ComputesStatusFromComponentStatuses_AllCompleted(t *testing.T) {
	f := newTestFixture(t)

	// All components successful
	f.createUoWManifestWithExitCode(core.ActionBuild, "eac-cli", "go", "go", 0)
	f.createUoWManifestWithExitCode(core.ActionBuild, "eac-cli", "docker", "docker", 0)

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionBuild, "eac-cli")
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, view.Status)
}

func TestOutputReader_GetModule_ComputesStatusFromComponentStatuses_OneFailed(t *testing.T) {
	f := newTestFixture(t)

	// One component failed
	f.createUoWManifestWithExitCode(core.ActionBuild, "eac-cli", "go", "go", 0)
	f.createUoWManifestWithExitCode(core.ActionBuild, "eac-cli", "docker", "docker", 1)

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionBuild, "eac-cli")
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, view.Status)
}

func TestOutputReader_GetModule_ReturnsCorrectTotalUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create 3 UoWs across 2 components
	f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "go")
	f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "cgo")
	f.createUoWManifest(core.ActionBuild, "eac-cli", "docker", "docker")

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionBuild, "eac-cli")
	require.NoError(t, err)
	totalUoWs := 0
	for _, comp := range view.Components {
		totalUoWs += len(comp.UoWs)
	}
	assert.Equal(t, 3, totalUoWs)
}

func TestOutputReader_GetModule_ComputesTotalSize(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs with artifacts
	f.createUoWManifestWithArtifacts(core.ActionBuild, "eac-cli", "go", "go", map[string][]byte{
		"binary": make([]byte, 5000),
	})
	f.createUoWManifestWithArtifacts(core.ActionBuild, "eac-cli", "docker", "docker", map[string][]byte{
		"image.tar": make([]byte, 10000),
	})

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionBuild, "eac-cli")
	require.NoError(t, err)
	assert.Equal(t, int64(15000), view.TotalSize)
}

func TestOutputReader_GetModule_ReturnsEmptyWhenNoUoWsExist(t *testing.T) {
	f := newTestFixture(t)

	// No UoWs exist for this module
	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionBuild, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, StatusPending, view.Status)
	assert.Empty(t, view.Components)
	_ = f
}

func TestOutputReader_GetModule_SingleComponent(t *testing.T) {
	f := newTestFixture(t)

	// Single component in module
	f.createUoWManifest(core.ActionLint, "core", "go", "golangci-lint")

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionLint, "core")
	require.NoError(t, err)
	assert.Len(t, view.Components, 1)
}

// =============================================================================
// ListUoWs Tests
// =============================================================================

func TestOutputReader_ListUoWs_ReturnsAllManifestsForModule(t *testing.T) {
	f := newTestFixture(t)

	// Create multiple UoWs
	f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "go")
	f.createUoWManifest(core.ActionBuild, "eac-cli", "go", "cgo")
	f.createUoWManifest(core.ActionBuild, "eac-cli", "docker", "docker")

	reader := NewReader(f.workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionBuild, "eac-cli")
	require.NoError(t, err)
	assert.Len(t, manifests, 3)
}

func TestOutputReader_ListUoWs_ReturnsEmptyListWhenNoManifestsExist(t *testing.T) {
	f := newTestFixture(t)

	// No manifests exist
	reader := NewReader(f.workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionBuild, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, manifests)
	_ = f
}

func TestOutputReader_ListUoWs_FiltersbyContext(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs in different contexts for the same module
	f.createUoWManifest(core.ActionBuild, "core", "go", "go")
	f.createUoWManifest(core.ActionTest, "core", "go", "gotest")
	f.createUoWManifest(core.ActionLint, "core", "go", "golangci-lint")

	reader := NewReader(f.workspaceRoot)

	buildManifests, err := reader.ListUoWs(core.ActionBuild, "core")
	require.NoError(t, err)
	assert.Len(t, buildManifests, 1)

	testManifests, err := reader.ListUoWs(core.ActionTest, "core")
	require.NoError(t, err)
	assert.Len(t, testManifests, 1)
}

func TestOutputReader_ListUoWs_FiltersByModule(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs in different modules
	f.createUoWManifest(core.ActionBuild, "module1", "go", "go")
	f.createUoWManifest(core.ActionBuild, "module2", "go", "go")
	f.createUoWManifest(core.ActionBuild, "module3", "go", "go")

	reader := NewReader(f.workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionBuild, "module1")
	require.NoError(t, err)
	assert.Len(t, manifests, 1)
	assert.Equal(t, "module1", manifests[0].Module)
}

func TestOutputReader_ListUoWs_SkipsInvalidManifests(t *testing.T) {
	f := newTestFixture(t)

	// Create valid manifest
	f.createUoWManifest(core.ActionBuild, "test-module", "go", "go")

	// Create invalid manifest (bad JSON)
	invalidDir := filepath.Join(f.workspaceRoot, "out", "build", "test-module", "docker-docker")
	err := os.MkdirAll(invalidDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(invalidDir, "uow.manifest.json"), []byte("invalid json {{{"), 0644)
	require.NoError(t, err)

	reader := NewReader(f.workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionBuild, "test-module")
	require.NoError(t, err)
	assert.Len(t, manifests, 1) // Only valid manifest returned
}
