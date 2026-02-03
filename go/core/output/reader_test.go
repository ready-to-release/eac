package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/require"
	// assert import will be used after implementation
)

// =============================================================================
// OutputReader Interface Definition (for reference)
// =============================================================================

// OutputReader provides aggregation views computed from UoW manifests.
// This interface will be implemented in reader.go.
type OutputReader interface {
	// GetUoW loads a single UoW manifest from disk.
	GetUoW(ctx workunit.Context, module, component, tool string) (*UoWManifest, error)

	// GetComponent computes component view by aggregating its UoWs.
	GetComponent(ctx workunit.Context, module, component string) (*ComponentView, error)

	// GetModule computes module view by aggregating all components.
	GetModule(ctx workunit.Context, module string) (*ModuleView, error)

	// ListUoWs returns all UoW manifests for a module.
	ListUoWs(ctx workunit.Context, module string) ([]*UoWManifest, error)

	// ValidateUoW checks if UoW output is valid.
	ValidateUoW(ctx workunit.Context, module, component, tool string) ValidationResult

	// ValidateModule checks if all UoWs for a module are valid.
	ValidateModule(ctx workunit.Context, module string, expectedUoWs []workunit.UnitID) ValidationResult
}

// =============================================================================
// Test Fixture Helpers
// =============================================================================

// testFixture provides helpers for setting up test directory structures.
type testFixture struct {
	t             *testing.T
	workspaceRoot string
}

// newTestFixture creates a new test fixture with a temporary directory.
func newTestFixture(t *testing.T) *testFixture {
	return &testFixture{
		t:             t,
		workspaceRoot: t.TempDir(),
	}
}

// createUoWManifest creates a UoW manifest file on disk.
func (f *testFixture) createUoWManifest(ctx workunit.Context, module, component, tool string) *UoWManifest {
	f.t.Helper()
	manifest := &UoWManifest{
		Context:    ctx,
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   0,
		InputHash:  "sha256:input-" + module + "-" + component + "-" + tool,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:output-" + module + "-" + component + "-" + tool,
		Version:    "1.0.0",
	}

	dirName := component + "_" + tool
	manifestDir := filepath.Join(f.workspaceRoot, "out", string(ctx), module, dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(f.t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(f.t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(f.t, err)

	return manifest
}

// createUoWManifestWithArtifacts creates a UoW manifest with artifacts on disk.
func (f *testFixture) createUoWManifestWithArtifacts(ctx workunit.Context, module, component, tool string, artifactContents map[string][]byte) *UoWManifest {
	f.t.Helper()

	dirName := component + "_" + tool
	uowDir := filepath.Join(f.workspaceRoot, "out", string(ctx), module, dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(f.t, err)

	// Create artifacts and compute hashes
	var artifacts []Artifact
	for name, content := range artifactContents {
		artifactPath := filepath.Join(uowDir, name)
		err := os.WriteFile(artifactPath, content, 0644)
		require.NoError(f.t, err)

		_, hash, err := HashFile(artifactPath)
		require.NoError(f.t, err)

		artifacts = append(artifacts, Artifact{
			ID:     name,
			Path:   name,
			SHA256: hash,
			Size:   int64(len(content)),
			Type:   "binary",
		})
	}

	manifest := &UoWManifest{
		Context:    ctx,
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   0,
		InputHash:  "sha256:input-" + module + "-" + component + "-" + tool,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts:  artifacts,
		OutputHash: "sha256:output-" + module + "-" + component + "-" + tool,
		Version:    "1.0.0",
	}

	manifestPath := filepath.Join(uowDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(f.t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(f.t, err)

	return manifest
}

// createUoWManifestWithExitCode creates a UoW manifest with specific exit code.
func (f *testFixture) createUoWManifestWithExitCode(ctx workunit.Context, module, component, tool string, exitCode int) *UoWManifest {
	f.t.Helper()
	manifest := &UoWManifest{
		Context:    ctx,
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   exitCode,
		InputHash:  "sha256:input-" + module + "-" + component + "-" + tool,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:output-" + module + "-" + component + "-" + tool,
		Version:    "1.0.0",
	}

	dirName := component + "_" + tool
	manifestDir := filepath.Join(f.workspaceRoot, "out", string(ctx), module, dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(f.t, err)

	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(f.t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(f.t, err)

	return manifest
}

// =============================================================================
// GetUoW Tests
// =============================================================================

func TestOutputReader_GetUoW_LoadsManifestFromCorrectPath(t *testing.T) {
	f := newTestFixture(t)

	// Create manifest at expected path
	expected := f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")

	// Expected path: {workspace}/out/build/test-module/go_go/uow.manifest.json

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifest, err := reader.GetUoW(workunit.ContextBuild, "test-module", "go", "go")
	// require.NoError(t, err)
	// assert.Equal(t, expected.Context, manifest.Context)
	// assert.Equal(t, expected.Module, manifest.Module)
	// assert.Equal(t, expected.Component, manifest.Component)
	// assert.Equal(t, expected.Tool, manifest.Tool)
	_ = expected
}

func TestOutputReader_GetUoW_ReturnsErrorWhenManifestMissing(t *testing.T) {
	f := newTestFixture(t)

	// No manifest exists

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifest, err := reader.GetUoW(workunit.ContextBuild, "nonexistent", "go", "go")
	// assert.Error(t, err)
	// assert.Nil(t, manifest)
	_ = f
}

func TestOutputReader_GetUoW_AllContexts(t *testing.T) {
	contexts := []workunit.Context{
		workunit.ContextBuild,
		workunit.ContextTest,
		workunit.ContextLint,
		workunit.ContextScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			f := newTestFixture(t)

			// Create manifest for this context
			expected := f.createUoWManifest(ctx, "module", "component", "tool")

			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// manifest, err := reader.GetUoW(ctx, "module", "component", "tool")
			// require.NoError(t, err)
			// assert.Equal(t, ctx, manifest.Context)
			_ = expected
		})
	}
}

func TestOutputReader_GetUoW_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		context   workunit.Context
		module    string
		component string
		tool      string
	}{
		{
			name:      "build go module",
			context:   workunit.ContextBuild,
			module:    "core",
			component: "go",
			tool:      "go",
		},
		{
			name:      "test with gotest",
			context:   workunit.ContextTest,
			module:    "eac-cli",
			component: "go",
			tool:      "gotest",
		},
		{
			name:      "lint with eslint",
			context:   workunit.ContextLint,
			module:    "web-app",
			component: "typescript",
			tool:      "eslint",
		},
		{
			name:      "scan with trivy",
			context:   workunit.ContextScan,
			module:    "eac-cli",
			component: "docker",
			tool:      "trivy-vuln",
		},
		{
			name:      "module with hyphens",
			context:   workunit.ContextBuild,
			module:    "my-complex-module",
			component: "go",
			tool:      "go",
		},
		{
			name:      "tool with hyphens",
			context:   workunit.ContextLint,
			module:    "core",
			component: "go",
			tool:      "golangci-lint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)

			expected := f.createUoWManifest(tt.context, tt.module, tt.component, tt.tool)

			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// manifest, err := reader.GetUoW(tt.context, tt.module, tt.component, tt.tool)
			// require.NoError(t, err)
			// assert.Equal(t, expected.Module, manifest.Module)
			// assert.Equal(t, expected.Component, manifest.Component)
			// assert.Equal(t, expected.Tool, manifest.Tool)
			_ = expected
		})
	}
}

// =============================================================================
// GetComponent Tests
// =============================================================================

func TestOutputReader_GetComponent_AggregatesUoWsForComponent(t *testing.T) {
	f := newTestFixture(t)

	// Create multiple UoWs for the same component (different tools)
	f.createUoWManifest(workunit.ContextBuild, "core", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "core", "go", "cgo")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetComponent(workunit.ContextBuild, "core", "go")
	// require.NoError(t, err)
	// assert.Equal(t, "core", view.Module)
	// assert.Equal(t, "go", view.Component)
	// assert.Len(t, view.UoWs, 2)
}

func TestOutputReader_GetComponent_ComputesStatusFromExitCodes_AllSuccess(t *testing.T) {
	f := newTestFixture(t)

	// All UoWs successful
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "core", "go", "go", 0)
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "core", "go", "cgo", 0)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetComponent(workunit.ContextBuild, "core", "go")
	// require.NoError(t, err)
	// assert.Equal(t, StatusCompleted, view.Status)
}

func TestOutputReader_GetComponent_ComputesStatusFromExitCodes_OneFailed(t *testing.T) {
	f := newTestFixture(t)

	// One UoW failed
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "core", "go", "go", 0)
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "core", "go", "cgo", 1)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetComponent(workunit.ContextBuild, "core", "go")
	// require.NoError(t, err)
	// assert.Equal(t, StatusFailed, view.Status)
}

func TestOutputReader_GetComponent_ComputesStatusFromExitCodes_AllFailed(t *testing.T) {
	f := newTestFixture(t)

	// All UoWs failed
	f.createUoWManifestWithExitCode(workunit.ContextTest, "core", "go", "gotest", 1)
	f.createUoWManifestWithExitCode(workunit.ContextTest, "core", "go", "godog", 2)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetComponent(workunit.ContextTest, "core", "go")
	// require.NoError(t, err)
	// assert.Equal(t, StatusFailed, view.Status)
}

func TestOutputReader_GetComponent_ReturnsErrorWhenNoUoWsExist(t *testing.T) {
	f := newTestFixture(t)

	// No UoWs exist for this component

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetComponent(workunit.ContextBuild, "nonexistent", "go")
	// This could either return an error or an empty component view
	// The design should clarify which behavior is expected
	_ = f
}

func TestOutputReader_GetComponent_SingleUoW(t *testing.T) {
	f := newTestFixture(t)

	// Single UoW for component
	f.createUoWManifest(workunit.ContextLint, "core", "go", "golangci-lint")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetComponent(workunit.ContextLint, "core", "go")
	// require.NoError(t, err)
	// assert.Len(t, view.UoWs, 1)
}

func TestOutputReader_GetComponent_ComputesTotalSize(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs with artifacts of known sizes
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "core", "go", "go", map[string][]byte{
		"binary1": make([]byte, 1000),
		"binary2": make([]byte, 2000),
	})
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "core", "go", "cgo", map[string][]byte{
		"binary3": make([]byte, 3000),
	})

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetComponent(workunit.ContextBuild, "core", "go")
	// require.NoError(t, err)
	// assert.Equal(t, int64(6000), view.TotalSize)
}

// =============================================================================
// GetModule Tests
// =============================================================================

func TestOutputReader_GetModule_AggregatesComponentsForModule(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs for multiple components in the same module
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "docker", "docker")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "docker", "buildx")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// assert.Equal(t, "eac-cli", view.Module)
	// assert.Len(t, view.Components, 2) // go, docker
}

func TestOutputReader_GetModule_ComputesStatusFromComponentStatuses_AllCompleted(t *testing.T) {
	f := newTestFixture(t)

	// All components successful
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "eac-cli", "go", "go", 0)
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "eac-cli", "docker", "docker", 0)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// assert.Equal(t, StatusCompleted, view.Status)
}

func TestOutputReader_GetModule_ComputesStatusFromComponentStatuses_OneFailed(t *testing.T) {
	f := newTestFixture(t)

	// One component failed
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "eac-cli", "go", "go", 0)
	f.createUoWManifestWithExitCode(workunit.ContextBuild, "eac-cli", "docker", "docker", 1)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// assert.Equal(t, StatusFailed, view.Status)
}

func TestOutputReader_GetModule_ReturnsCorrectTotalUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create 3 UoWs across 2 components
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "go", "cgo")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "docker", "docker")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// totalUoWs := 0
	// for _, comp := range view.Components {
	//     totalUoWs += len(comp.UoWs)
	// }
	// assert.Equal(t, 3, totalUoWs)
}

func TestOutputReader_GetModule_ComputesTotalSize(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs with artifacts
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "eac-cli", "go", "go", map[string][]byte{
		"binary": make([]byte, 5000),
	})
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "eac-cli", "docker", "docker", map[string][]byte{
		"image.tar": make([]byte, 10000),
	})

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// assert.Equal(t, int64(15000), view.TotalSize)
}

func TestOutputReader_GetModule_ReturnsEmptyWhenNoUoWsExist(t *testing.T) {
	f := newTestFixture(t)

	// No UoWs exist for this module

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextBuild, "nonexistent")
	// This could either return an error or an empty module view
	_ = f
}

func TestOutputReader_GetModule_SingleComponent(t *testing.T) {
	f := newTestFixture(t)

	// Single component in module
	f.createUoWManifest(workunit.ContextLint, "core", "go", "golangci-lint")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextLint, "core")
	// require.NoError(t, err)
	// assert.Len(t, view.Components, 1)
}

// =============================================================================
// ListUoWs Tests
// =============================================================================

func TestOutputReader_ListUoWs_ReturnsAllManifestsForModule(t *testing.T) {
	f := newTestFixture(t)

	// Create multiple UoWs
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "go", "cgo")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "docker", "docker")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifests, err := reader.ListUoWs(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// assert.Len(t, manifests, 3)
}

func TestOutputReader_ListUoWs_ReturnsEmptyListWhenNoManifestsExist(t *testing.T) {
	f := newTestFixture(t)

	// No manifests exist

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifests, err := reader.ListUoWs(workunit.ContextBuild, "nonexistent")
	// require.NoError(t, err)
	// assert.Empty(t, manifests)
	_ = f
}

func TestOutputReader_ListUoWs_FiltersbyContext(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs in different contexts for the same module
	f.createUoWManifest(workunit.ContextBuild, "core", "go", "go")
	f.createUoWManifest(workunit.ContextTest, "core", "go", "gotest")
	f.createUoWManifest(workunit.ContextLint, "core", "go", "golangci-lint")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	//
	// buildManifests, err := reader.ListUoWs(workunit.ContextBuild, "core")
	// require.NoError(t, err)
	// assert.Len(t, buildManifests, 1)
	//
	// testManifests, err := reader.ListUoWs(workunit.ContextTest, "core")
	// require.NoError(t, err)
	// assert.Len(t, testManifests, 1)
}

func TestOutputReader_ListUoWs_FiltersByModule(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs in different modules
	f.createUoWManifest(workunit.ContextBuild, "module1", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "module2", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "module3", "go", "go")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifests, err := reader.ListUoWs(workunit.ContextBuild, "module1")
	// require.NoError(t, err)
	// assert.Len(t, manifests, 1)
	// assert.Equal(t, "module1", manifests[0].Module)
}

func TestOutputReader_ListUoWs_SkipsInvalidManifests(t *testing.T) {
	f := newTestFixture(t)

	// Create valid manifest
	f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")

	// Create invalid manifest (bad JSON)
	invalidDir := filepath.Join(f.workspaceRoot, "out", "build", "test-module", "docker_docker")
	err := os.MkdirAll(invalidDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(invalidDir, "uow.manifest.json"), []byte("invalid json {{{"), 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifests, err := reader.ListUoWs(workunit.ContextBuild, "test-module")
	// require.NoError(t, err)
	// assert.Len(t, manifests, 1) // Only valid manifest returned
}

// =============================================================================
// ValidateUoW Tests
// =============================================================================

func TestOutputReader_ValidateUoW_ReturnsValidWhenManifestAndArtifactsExist(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW with valid artifacts
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "test-module", "go", "go", map[string][]byte{
		"binary": []byte("binary content"),
	})

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateUoW(workunit.ContextBuild, "test-module", "go", "go")
	// assert.True(t, result.Valid)
	// assert.True(t, result.ManifestExists)
	// assert.True(t, result.ManifestValid)
	// assert.True(t, result.ArtifactsValid)
}

func TestOutputReader_ValidateUoW_ReturnsInvalidWithMissingArtifacts(t *testing.T) {
	f := newTestFixture(t)

	// Create manifest referencing non-existent artifacts
	manifest := f.createUoWManifest(workunit.ContextBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: "missing-binary", SHA256: "sha256:missing", Size: 1000, Type: "binary"},
	}

	// Overwrite manifest with artifact references
	dirName := "go_go"
	manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "test-module", dirName, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateUoW(workunit.ContextBuild, "test-module", "go", "go")
	// assert.False(t, result.Valid)
	// assert.True(t, result.ManifestExists)
	// assert.True(t, result.ManifestValid)
	// assert.False(t, result.ArtifactsValid)
	// assert.Contains(t, result.MissingArtifacts, "missing-binary")
}

func TestOutputReader_ValidateUoW_ReturnsInvalidWhenManifestMissing(t *testing.T) {
	f := newTestFixture(t)

	// No manifest exists

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateUoW(workunit.ContextBuild, "nonexistent", "go", "go")
	// assert.False(t, result.Valid)
	// assert.False(t, result.ManifestExists)
	_ = f
}

func TestOutputReader_ValidateUoW_ReturnsInvalidWhenArtifactsCorrupt(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW directory
	dirName := "go_go"
	uowDir := filepath.Join(f.workspaceRoot, "out", "build", "test-module", dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(t, err)

	// Create artifact with content
	artifactContent := []byte("actual content")
	artifactPath := filepath.Join(uowDir, "binary")
	err = os.WriteFile(artifactPath, artifactContent, 0644)
	require.NoError(t, err)

	// Create manifest with wrong hash
	manifest := &UoWManifest{
		Context:   workunit.ContextBuild,
		Module:    "test-module",
		Component: "go",
		Tool:      "go",
		Artifacts: []Artifact{
			{ID: "binary", Path: "binary", SHA256: "sha256:wrong-hash", Size: int64(len(artifactContent)), Type: "binary"},
		},
	}
	manifestPath := filepath.Join(uowDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateUoW(workunit.ContextBuild, "test-module", "go", "go")
	// assert.False(t, result.Valid)
	// assert.False(t, result.ArtifactsValid)
	// assert.Contains(t, result.CorruptArtifacts, "binary")
}

func TestOutputReader_ValidateUoW_ReturnsValidForEmptyArtifacts(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW with no artifacts (valid for lint)
	f.createUoWManifest(workunit.ContextLint, "test-module", "go", "golangci-lint")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateUoW(workunit.ContextLint, "test-module", "go", "golangci-lint")
	// assert.True(t, result.Valid)
	// assert.True(t, result.ArtifactsValid)
}

func TestOutputReader_ValidateUoW_TableDriven(t *testing.T) {
	tests := []struct {
		name               string
		setupManifest      bool
		artifacts          map[string][]byte
		wrongHash          bool
		expectValid        bool
		expectManifest     bool
		expectArtifacts    bool
		expectMissing      int
		expectCorrupt      int
	}{
		{
			name:          "valid with artifacts",
			setupManifest: true,
			artifacts: map[string][]byte{
				"binary": []byte("content"),
			},
			expectValid:     true,
			expectManifest:  true,
			expectArtifacts: true,
		},
		{
			name:          "valid with no artifacts",
			setupManifest: true,
			artifacts:     map[string][]byte{},
			expectValid:   true,
			expectManifest:  true,
			expectArtifacts: true,
		},
		{
			name:          "no manifest",
			setupManifest: false,
			expectValid:   false,
			expectManifest:  false,
			expectArtifacts: false,
		},
		{
			name:          "manifest with wrong hash",
			setupManifest: true,
			artifacts: map[string][]byte{
				"binary": []byte("content"),
			},
			wrongHash:       true,
			expectValid:     false,
			expectManifest:  true,
			expectArtifacts: false,
			expectCorrupt:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)

			if tt.setupManifest {
				if len(tt.artifacts) > 0 {
					manifest := f.createUoWManifestWithArtifacts(workunit.ContextBuild, "test", "go", "go", tt.artifacts)

					if tt.wrongHash {
						// Corrupt the hash
						manifest.Artifacts[0].SHA256 = "sha256:wrong"
						dirName := "go_go"
						manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "test", dirName, "uow.manifest.json")
						data, err := json.MarshalIndent(manifest, "", "  ")
						require.NoError(t, err)
						err = os.WriteFile(manifestPath, data, 0644)
						require.NoError(t, err)
					}
				} else {
					f.createUoWManifest(workunit.ContextBuild, "test", "go", "go")
				}
			}

			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// result := reader.ValidateUoW(workunit.ContextBuild, "test", "go", "go")
			// assert.Equal(t, tt.expectValid, result.Valid)
			// assert.Equal(t, tt.expectManifest, result.ManifestExists)
			// assert.Equal(t, tt.expectArtifacts, result.ArtifactsValid)
			// assert.Len(t, result.MissingArtifacts, tt.expectMissing)
			// assert.Len(t, result.CorruptArtifacts, tt.expectCorrupt)
		})
	}
}

// =============================================================================
// ValidateModule Tests
// =============================================================================

func TestOutputReader_ValidateModule_ChecksAllExpectedUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create expected UoWs
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "eac-cli", "go", "go", map[string][]byte{
		"binary": []byte("binary content"),
	})
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "docker", "docker")

	expectedUoWs := []workunit.UnitID{
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "go", Tool: "go"},
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "docker", Tool: "docker"},
	}

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateModule(workunit.ContextBuild, "eac-cli", expectedUoWs)
	// assert.True(t, result.Valid)
	_ = expectedUoWs
}

func TestOutputReader_ValidateModule_ReturnsInvalidWhenUoWMissing(t *testing.T) {
	f := newTestFixture(t)

	// Only create one of two expected UoWs
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "go", "go")
	// Missing: docker

	expectedUoWs := []workunit.UnitID{
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "go", Tool: "go"},
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "docker", Tool: "docker"},
	}

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateModule(workunit.ContextBuild, "eac-cli", expectedUoWs)
	// assert.False(t, result.Valid)
	// The error should indicate which UoW is missing
	_ = expectedUoWs
}

func TestOutputReader_ValidateModule_ReturnsInvalidWhenAnyUoWInvalid(t *testing.T) {
	f := newTestFixture(t)

	// Create one valid UoW
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "eac-cli", "go", "go", map[string][]byte{
		"binary": []byte("binary content"),
	})

	// Create one UoW with missing artifacts
	manifest := f.createUoWManifest(workunit.ContextBuild, "eac-cli", "docker", "docker")
	manifest.Artifacts = []Artifact{
		{ID: "image", Path: "missing-image.tar", SHA256: "sha256:missing", Size: 1000, Type: "docker-image"},
	}
	dirName := "docker_docker"
	manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "eac-cli", dirName, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	expectedUoWs := []workunit.UnitID{
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "go", Tool: "go"},
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "docker", Tool: "docker"},
	}

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateModule(workunit.ContextBuild, "eac-cli", expectedUoWs)
	// assert.False(t, result.Valid)
	_ = expectedUoWs
}

func TestOutputReader_ValidateModule_SucceedsWithEmptyExpectedList(t *testing.T) {
	f := newTestFixture(t)

	expectedUoWs := []workunit.UnitID{}

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateModule(workunit.ContextBuild, "eac-cli", expectedUoWs)
	// assert.True(t, result.Valid) // No expected UoWs, so validation passes
	_ = f
	_ = expectedUoWs
}

func TestOutputReader_ValidateModule_IgnoresUnexpectedUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create more UoWs than expected
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "docker", "docker")
	f.createUoWManifest(workunit.ContextBuild, "eac-cli", "extra", "extra") // Unexpected

	expectedUoWs := []workunit.UnitID{
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "go", Tool: "go"},
		{Context: workunit.ContextBuild, Module: "eac-cli", Component: "docker", Tool: "docker"},
	}

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateModule(workunit.ContextBuild, "eac-cli", expectedUoWs)
	// assert.True(t, result.Valid) // Extra UoWs should not cause failure
	_ = expectedUoWs
}

func TestOutputReader_ValidateModule_AllContexts(t *testing.T) {
	contexts := []workunit.Context{
		workunit.ContextBuild,
		workunit.ContextTest,
		workunit.ContextLint,
		workunit.ContextScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			f := newTestFixture(t)

			f.createUoWManifest(ctx, "module", "component", "tool")

			expectedUoWs := []workunit.UnitID{
				{Context: ctx, Module: "module", Component: "component", Tool: "tool"},
			}

			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// result := reader.ValidateModule(ctx, "module", expectedUoWs)
			// assert.True(t, result.Valid)
			_ = expectedUoWs
		})
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestOutputReader_ConcurrentGetUoW(t *testing.T) {
	f := newTestFixture(t)

	// Create multiple UoWs
	f.createUoWManifest(workunit.ContextBuild, "module1", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "module2", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "module3", "go", "go")

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(moduleNum int) {
			defer wg.Done()
			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// module := fmt.Sprintf("module%d", moduleNum)
			// manifest, err := reader.GetUoW(workunit.ContextBuild, module, "go", "go")
			// require.NoError(t, err)
			// assert.Equal(t, module, manifest.Module)
			_ = moduleNum
		}(i)
	}
	wg.Wait()
}

func TestOutputReader_ConcurrentListUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs in multiple modules
	for i := 1; i <= 3; i++ {
		f.createUoWManifest(workunit.ContextBuild, "module", "comp1", "tool1")
		f.createUoWManifest(workunit.ContextBuild, "module", "comp2", "tool2")
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// manifests, err := reader.ListUoWs(workunit.ContextBuild, "module")
			// require.NoError(t, err)
			// assert.Len(t, manifests, 2)
		}()
	}
	wg.Wait()
}

func TestOutputReader_ConcurrentValidateUoW(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs with artifacts
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "module1", "go", "go", map[string][]byte{
		"binary": []byte("content1"),
	})
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "module2", "go", "go", map[string][]byte{
		"binary": []byte("content2"),
	})

	var wg sync.WaitGroup
	for i := 1; i <= 2; i++ {
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func(moduleNum int) {
				defer wg.Done()
				// TODO: After implementation:
				// reader := NewReader(f.workspaceRoot)
				// module := fmt.Sprintf("module%d", moduleNum)
				// result := reader.ValidateUoW(workunit.ContextBuild, module, "go", "go")
				// assert.True(t, result.Valid)
				_ = moduleNum
			}(i)
		}
	}
	wg.Wait()
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestOutputReader_GetUoW_HandlesInvalidManifestJSON(t *testing.T) {
	f := newTestFixture(t)

	// Create directory with invalid manifest
	dirName := "go_go"
	manifestDir := filepath.Join(f.workspaceRoot, "out", "build", "test-module", dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(manifestDir, "uow.manifest.json"), []byte("invalid json"), 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifest, err := reader.GetUoW(workunit.ContextBuild, "test-module", "go", "go")
	// assert.Error(t, err)
	// assert.Nil(t, manifest)
}

func TestOutputReader_GetUoW_HandlesEmptyManifest(t *testing.T) {
	f := newTestFixture(t)

	// Create directory with empty manifest
	dirName := "go_go"
	manifestDir := filepath.Join(f.workspaceRoot, "out", "build", "test-module", dirName)
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(manifestDir, "uow.manifest.json"), []byte(""), 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifest, err := reader.GetUoW(workunit.ContextBuild, "test-module", "go", "go")
	// assert.Error(t, err)
}

func TestOutputReader_ListUoWs_HandlesEmptyModule(t *testing.T) {
	f := newTestFixture(t)

	// Create module directory but no UoW directories
	moduleDir := filepath.Join(f.workspaceRoot, "out", "build", "empty-module")
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifests, err := reader.ListUoWs(workunit.ContextBuild, "empty-module")
	// require.NoError(t, err)
	// assert.Empty(t, manifests)
}

func TestOutputReader_GetModule_HandlesMixedValidAndInvalidManifests(t *testing.T) {
	f := newTestFixture(t)

	// Create valid manifest
	f.createUoWManifest(workunit.ContextBuild, "mixed-module", "go", "go")

	// Create invalid manifest
	invalidDir := filepath.Join(f.workspaceRoot, "out", "build", "mixed-module", "docker_docker")
	err := os.MkdirAll(invalidDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(invalidDir, "uow.manifest.json"), []byte("invalid"), 0644)
	require.NoError(t, err)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// view, err := reader.GetModule(workunit.ContextBuild, "mixed-module")
	// require.NoError(t, err)
	// Should return view with only valid UoWs
	// assert.Len(t, view.Components, 1)
}

func TestOutputReader_WorksWithSpecialCharactersInNames(t *testing.T) {
	f := newTestFixture(t)

	// Module and component names with hyphens
	f.createUoWManifest(workunit.ContextBuild, "my-complex-module", "go-component", "my-tool")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// manifest, err := reader.GetUoW(workunit.ContextBuild, "my-complex-module", "go-component", "my-tool")
	// require.NoError(t, err)
	// assert.Equal(t, "my-complex-module", manifest.Module)
}

func TestOutputReader_ValidateUoW_MultipleArtifacts(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW with multiple artifacts
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "test-module", "go", "go", map[string][]byte{
		"eac-linux-amd64":       make([]byte, 1000),
		"eac-darwin-amd64":      make([]byte, 1100),
		"eac-windows-amd64.exe": make([]byte, 1200),
	})

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	// result := reader.ValidateUoW(workunit.ContextBuild, "test-module", "go", "go")
	// assert.True(t, result.Valid)
	// assert.True(t, result.ArtifactsValid)
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestOutputReader_FullWorkflow_BuildModule(t *testing.T) {
	f := newTestFixture(t)

	// Simulate a full build output structure
	// eac-cli module with go and docker components

	// Go component
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "eac-cli", "go", "go", map[string][]byte{
		"eac-linux-amd64":       make([]byte, 10000),
		"eac-darwin-amd64":      make([]byte, 11000),
		"eac-windows-amd64.exe": make([]byte, 12000),
	})

	// Docker component
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "eac-cli", "docker", "docker", map[string][]byte{
		"image.tar": make([]byte, 50000),
	})

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	//
	// 1. List all UoWs
	// manifests, err := reader.ListUoWs(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// assert.Len(t, manifests, 2)
	//
	// 2. Get module view
	// moduleView, err := reader.GetModule(workunit.ContextBuild, "eac-cli")
	// require.NoError(t, err)
	// assert.Equal(t, StatusCompleted, moduleView.Status)
	// assert.Len(t, moduleView.Components, 2)
	//
	// 3. Validate module
	// expectedUoWs := []workunit.UnitID{
	//     {Context: workunit.ContextBuild, Module: "eac-cli", Component: "go", Tool: "go"},
	//     {Context: workunit.ContextBuild, Module: "eac-cli", Component: "docker", Tool: "docker"},
	// }
	// result := reader.ValidateModule(workunit.ContextBuild, "eac-cli", expectedUoWs)
	// assert.True(t, result.Valid)
}

func TestOutputReader_FullWorkflow_TestModule(t *testing.T) {
	f := newTestFixture(t)

	// Simulate test output structure
	f.createUoWManifestWithExitCode(workunit.ContextTest, "core", "go", "gotest", 0)
	f.createUoWManifestWithExitCode(workunit.ContextTest, "core", "gherkin", "godog", 0)

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	//
	// moduleView, err := reader.GetModule(workunit.ContextTest, "core")
	// require.NoError(t, err)
	// assert.Equal(t, StatusCompleted, moduleView.Status)
}

func TestOutputReader_FullWorkflow_PartialFailure(t *testing.T) {
	f := newTestFixture(t)

	// Simulate partial test failure
	f.createUoWManifestWithExitCode(workunit.ContextTest, "core", "go", "gotest", 0)  // Pass
	f.createUoWManifestWithExitCode(workunit.ContextTest, "core", "gherkin", "godog", 1) // Fail

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	//
	// moduleView, err := reader.GetModule(workunit.ContextTest, "core")
	// require.NoError(t, err)
	// assert.Equal(t, StatusFailed, moduleView.Status)
}

func TestOutputReader_FullWorkflow_MultipleModules(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs for multiple modules
	f.createUoWManifest(workunit.ContextBuild, "module1", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "module2", "go", "go")
	f.createUoWManifest(workunit.ContextBuild, "module3", "go", "go")

	// TODO: After implementation:
	// reader := NewReader(f.workspaceRoot)
	//
	// for i := 1; i <= 3; i++ {
	//     module := fmt.Sprintf("module%d", i)
	//     view, err := reader.GetModule(workunit.ContextBuild, module)
	//     require.NoError(t, err)
	//     assert.Equal(t, module, view.Module)
	// }
}

// =============================================================================
// Table-Driven Comprehensive Tests
// =============================================================================

func TestOutputReader_GetModule_StatusComputation(t *testing.T) {
	tests := []struct {
		name           string
		uowExitCodes   []int
		expectedStatus Status
	}{
		{
			name:           "all success",
			uowExitCodes:   []int{0, 0, 0},
			expectedStatus: StatusCompleted,
		},
		{
			name:           "one failure",
			uowExitCodes:   []int{0, 1, 0},
			expectedStatus: StatusFailed,
		},
		{
			name:           "all failure",
			uowExitCodes:   []int{1, 2, 1},
			expectedStatus: StatusFailed,
		},
		{
			name:           "single success",
			uowExitCodes:   []int{0},
			expectedStatus: StatusCompleted,
		},
		{
			name:           "single failure",
			uowExitCodes:   []int{1},
			expectedStatus: StatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)

			// Create UoWs with specified exit codes
			for i, exitCode := range tt.uowExitCodes {
				component := "component"
				tool := "tool" + string(rune('0'+i))
				f.createUoWManifestWithExitCode(workunit.ContextBuild, "module", component, tool, exitCode)
			}

			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// view, err := reader.GetModule(workunit.ContextBuild, "module")
			// require.NoError(t, err)
			// assert.Equal(t, tt.expectedStatus, view.Status)
		})
	}
}

func TestOutputReader_ValidateModule_Comprehensive(t *testing.T) {
	tests := []struct {
		name          string
		setupUoWs     []struct {
			component   string
			tool        string
			artifacts   map[string][]byte
			wrongHash   bool
		}
		expectedUoWs  []workunit.UnitID
		expectValid   bool
	}{
		{
			name: "all UoWs valid",
			setupUoWs: []struct {
				component string
				tool      string
				artifacts map[string][]byte
				wrongHash bool
			}{
				{component: "go", tool: "go", artifacts: map[string][]byte{"binary": []byte("content")}},
				{component: "docker", tool: "docker", artifacts: map[string][]byte{"image": []byte("image")}},
			},
			expectedUoWs: []workunit.UnitID{
				{Context: workunit.ContextBuild, Module: "module", Component: "go", Tool: "go"},
				{Context: workunit.ContextBuild, Module: "module", Component: "docker", Tool: "docker"},
			},
			expectValid: true,
		},
		{
			name: "missing expected UoW",
			setupUoWs: []struct {
				component string
				tool      string
				artifacts map[string][]byte
				wrongHash bool
			}{
				{component: "go", tool: "go", artifacts: map[string][]byte{"binary": []byte("content")}},
			},
			expectedUoWs: []workunit.UnitID{
				{Context: workunit.ContextBuild, Module: "module", Component: "go", Tool: "go"},
				{Context: workunit.ContextBuild, Module: "module", Component: "docker", Tool: "docker"}, // Missing
			},
			expectValid: false,
		},
		{
			name: "corrupt artifact",
			setupUoWs: []struct {
				component string
				tool      string
				artifacts map[string][]byte
				wrongHash bool
			}{
				{component: "go", tool: "go", artifacts: map[string][]byte{"binary": []byte("content")}, wrongHash: true},
			},
			expectedUoWs: []workunit.UnitID{
				{Context: workunit.ContextBuild, Module: "module", Component: "go", Tool: "go"},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)

			// Setup UoWs
			for _, uow := range tt.setupUoWs {
				manifest := f.createUoWManifestWithArtifacts(workunit.ContextBuild, "module", uow.component, uow.tool, uow.artifacts)

				if uow.wrongHash && len(manifest.Artifacts) > 0 {
					manifest.Artifacts[0].SHA256 = "sha256:wrong"
					dirName := uow.component + "_" + uow.tool
					manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "module", dirName, "uow.manifest.json")
					data, err := json.MarshalIndent(manifest, "", "  ")
					require.NoError(t, err)
					err = os.WriteFile(manifestPath, data, 0644)
					require.NoError(t, err)
				}
			}

			// TODO: After implementation:
			// reader := NewReader(f.workspaceRoot)
			// result := reader.ValidateModule(workunit.ContextBuild, "module", tt.expectedUoWs)
			// assert.Equal(t, tt.expectValid, result.Valid)
			_ = tt.expectedUoWs
			_ = tt.expectValid
		})
	}
}
