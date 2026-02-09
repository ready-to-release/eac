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
// OutputReader Interface Definition (for reference)
// =============================================================================

// OutputReader provides aggregation views computed from UoW manifests.
// This interface will be implemented in reader.go.
type OutputReader interface {
	// GetUoW loads a single UoW manifest from disk.
	GetUoW(ctx core.ActionType, module, component, tool string) (*UoWManifest, error)

	// GetComponent computes component view by aggregating its UoWs.
	GetComponent(ctx core.ActionType, module, component string) (*ComponentView, error)

	// GetModule computes module view by aggregating all components.
	GetModule(ctx core.ActionType, module string) (*ModuleView, error)

	// ListUoWs returns all UoW manifests for a module.
	ListUoWs(ctx core.ActionType, module string) ([]*UoWManifest, error)

	// ValidateUoW checks if UoW output is valid.
	ValidateUoW(ctx core.ActionType, module, component, tool string) ValidationResult

	// ValidateModule checks if all UoWs for a module are valid.
	ValidateModule(ctx core.ActionType, module string, expectedUoWs []workunit.UnitID) ValidationResult
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
// Uses the new flat directory format: component[-extra1][-extra2]
func (f *testFixture) createUoWManifest(ctx core.ActionType, module, component, tool string) *UoWManifest {
	f.t.Helper()
	manifest := &UoWManifest{
		Action:     ctx,
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

	// Use the new flat directory format: just component name (no tool)
	dirName := manifest.DirName()
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
// Uses the new flat directory format: component[-extra1][-extra2]
func (f *testFixture) createUoWManifestWithArtifacts(ctx core.ActionType, module, component, tool string, artifactContents map[string][]byte) *UoWManifest {
	f.t.Helper()

	// Create manifest first to get the correct DirName
	manifest := &UoWManifest{
		Action:     ctx,
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   0,
		InputHash:  "sha256:input-" + module + "-" + component + "-" + tool,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		OutputHash: "sha256:output-" + module + "-" + component + "-" + tool,
		Version:    "1.0.0",
	}

	// Use the new flat directory format
	dirName := manifest.DirName()
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

	manifest.Artifacts = artifacts

	manifestPath := filepath.Join(uowDir, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(f.t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(f.t, err)

	return manifest
}

// createUoWManifestWithExitCode creates a UoW manifest with specific exit code.
// Uses the new flat directory format: component[-extra1][-extra2]
func (f *testFixture) createUoWManifestWithExitCode(ctx core.ActionType, module, component, tool string, exitCode int) *UoWManifest {
	f.t.Helper()
	manifest := &UoWManifest{
		Action:     ctx,
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

	// Use the new flat directory format
	dirName := manifest.DirName()
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
	expected := f.createUoWManifest(core.ActionBuild, "test-module", "go", "go")

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "go"}
	manifest, err := reader.GetUoW(id)
	require.NoError(t, err)
	assert.Equal(t, expected.Action, manifest.Action)
	assert.Equal(t, expected.Module, manifest.Module)
	assert.Equal(t, expected.Component, manifest.Component)
	assert.Equal(t, expected.Tool, manifest.Tool)
}

func TestOutputReader_GetUoW_ReturnsErrorWhenManifestMissing(t *testing.T) {
	f := newTestFixture(t)

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "nonexistent", ComponentType: "go", ComponentName: "go", Tool: "go"}
	manifest, err := reader.GetUoW(id)
	assert.Error(t, err)
	assert.Nil(t, manifest)
	_ = f
}

func TestOutputReader_GetUoW_AllContexts(t *testing.T) {
	contexts := []core.ActionType{
		core.ActionBuild,
		core.ActionTest,
		core.ActionLint,
		core.ActionScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			f := newTestFixture(t)

			// Create manifest for this context
			f.createUoWManifest(ctx, "module", "component", "tool")

			reader := NewReader(f.workspaceRoot)
			id := workunit.UnitID{Action: ctx, Module: "module", ComponentType: "component", ComponentName: "component", Tool: "tool"}
			manifest, err := reader.GetUoW(id)
			require.NoError(t, err)
			assert.Equal(t, ctx, manifest.Action)
		})
	}
}

func TestOutputReader_GetUoW_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		action    core.ActionType
		module    string
		component string
		tool      string
	}{
		{
			name:      "build go module",
			action:    core.ActionBuild,
			module:    "core",
			component: "go",
			tool:      "go",
		},
		{
			name:      "test with gotest",
			action:    core.ActionTest,
			module:    "eac",
			component: "go",
			tool:      "gotest",
		},
		{
			name:      "lint with eslint",
			action:    core.ActionLint,
			module:    "web-app",
			component: "typescript",
			tool:      "eslint",
		},
		{
			name:      "scan with trivy",
			action:    core.ActionScan,
			module:    "eac",
			component: "docker",
			tool:      "trivy-vuln",
		},
		{
			name:      "module with hyphens",
			action:    core.ActionBuild,
			module:    "my-complex-module",
			component: "go",
			tool:      "go",
		},
		{
			name:      "tool with hyphens",
			action:    core.ActionLint,
			module:    "core",
			component: "go",
			tool:      "golangci-lint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)

			expected := f.createUoWManifest(tt.action, tt.module, tt.component, tt.tool)

			reader := NewReader(f.workspaceRoot)
			id := workunit.UnitID{Action: tt.action, Module: tt.module, ComponentType: tt.component, ComponentName: tt.component, Tool: tt.tool}
			manifest, err := reader.GetUoW(id)
			require.NoError(t, err)
			assert.Equal(t, expected.Module, manifest.Module)
			assert.Equal(t, expected.Component, manifest.Component)
			assert.Equal(t, expected.Tool, manifest.Tool)
		})
	}
}
