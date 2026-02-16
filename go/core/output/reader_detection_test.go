package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ValidateUoW Tests
// =============================================================================

func TestOutputReader_ValidateUoW_ReturnsValidWhenManifestAndArtifactsExist(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW with valid artifacts
	f.createUoWManifestWithArtifacts(core.ActionBuild, "test-module", "go", "go", map[string][]byte{
		"binary": []byte("binary content"),
	})

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "go"}
	result := reader.ValidateUoW(id)
	assert.True(t, result.Valid)
	assert.True(t, result.ManifestExists)
	assert.True(t, result.ManifestValid)
	assert.True(t, result.ArtifactsValid)
}

func TestOutputReader_ValidateUoW_ReturnsInvalidWithMissingArtifacts(t *testing.T) {
	f := newTestFixture(t)

	// Create manifest referencing non-existent artifacts
	manifest := f.createUoWManifest(core.ActionBuild, "test-module", "go", "go")
	manifest.Artifacts = []Artifact{
		{ID: "binary", Path: "missing-binary", SHA256: "sha256:missing", Size: 1000, Type: "binary"},
	}

	// Overwrite manifest with artifact references
	dirName := manifest.DirName()
	manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "test-module", dirName, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "go"}
	result := reader.ValidateUoW(id)
	assert.False(t, result.Valid)
	assert.True(t, result.ManifestExists)
	assert.True(t, result.ManifestValid)
	assert.False(t, result.ArtifactsValid)
	assert.Contains(t, result.MissingArtifacts, "missing-binary")
}

func TestOutputReader_ValidateUoW_ReturnsInvalidWhenManifestMissing(t *testing.T) {
	f := newTestFixture(t)

	// No manifest exists
	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "nonexistent", ComponentType: "go", ComponentName: "go", Tool: "go"}
	result := reader.ValidateUoW(id)
	assert.False(t, result.Valid)
	assert.False(t, result.ManifestExists)
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
		Action:    core.ActionBuild,
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

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "go"}
	result := reader.ValidateUoW(id)
	assert.False(t, result.Valid)
	assert.False(t, result.ArtifactsValid)
	assert.Contains(t, result.CorruptArtifacts, "binary")
}

func TestOutputReader_ValidateUoW_ReturnsValidForEmptyArtifacts(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW with no artifacts (valid for lint)
	f.createUoWManifest(core.ActionLint, "test-module", "go", "golangci-lint")

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionLint, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "golangci-lint"}
	result := reader.ValidateUoW(id)
	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

func TestOutputReader_ValidateUoW_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		setupManifest   bool
		artifacts       map[string][]byte
		wrongHash       bool
		expectValid     bool
		expectManifest  bool
		expectArtifacts bool
		expectMissing   int
		expectCorrupt   int
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
			name:            "valid with no artifacts",
			setupManifest:   true,
			artifacts:       map[string][]byte{},
			expectValid:     true,
			expectManifest:  true,
			expectArtifacts: true,
		},
		{
			name:            "no manifest",
			setupManifest:   false,
			expectValid:     false,
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
					manifest := f.createUoWManifestWithArtifacts(core.ActionBuild, "test", "go", "go", tt.artifacts)

					if tt.wrongHash {
						// Corrupt the hash
						manifest.Artifacts[0].SHA256 = "sha256:wrong"
						dirName := manifest.DirName()
						manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "test", dirName, "uow.manifest.json")
						data, err := json.MarshalIndent(manifest, "", "  ")
						require.NoError(t, err)
						err = os.WriteFile(manifestPath, data, 0644)
						require.NoError(t, err)
					}
				} else {
					f.createUoWManifest(core.ActionBuild, "test", "go", "go")
				}
			}

			reader := NewReader(f.workspaceRoot)
			id := workunit.UnitID{Action: core.ActionBuild, Module: "test", ComponentType: "go", ComponentName: "go", Tool: "go"}
			result := reader.ValidateUoW(id)
			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Equal(t, tt.expectManifest, result.ManifestExists)
			assert.Equal(t, tt.expectArtifacts, result.ArtifactsValid)
			assert.Len(t, result.MissingArtifacts, tt.expectMissing)
			assert.Len(t, result.CorruptArtifacts, tt.expectCorrupt)
		})
	}
}

// =============================================================================
// ValidateModule Tests
// =============================================================================

func TestOutputReader_ValidateModule_ChecksAllExpectedUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create expected UoWs
	f.createUoWManifestWithArtifacts(core.ActionBuild, "eac", "go", "go", map[string][]byte{
		"binary": []byte("binary content"),
	})
	f.createUoWManifest(core.ActionBuild, "eac", "docker", "docker")

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "eac", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	reader := NewReader(f.workspaceRoot)
	result := reader.ValidateModule(core.ActionBuild, "eac", expectedUoWs)
	assert.True(t, result.Valid)
}

func TestOutputReader_ValidateModule_ReturnsInvalidWhenUoWMissing(t *testing.T) {
	f := newTestFixture(t)

	// Only create one of two expected UoWs
	f.createUoWManifest(core.ActionBuild, "eac", "go", "go")
	// Missing: docker

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "eac", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	reader := NewReader(f.workspaceRoot)
	result := reader.ValidateModule(core.ActionBuild, "eac", expectedUoWs)
	assert.False(t, result.Valid)
}

func TestOutputReader_ValidateModule_ReturnsInvalidWhenAnyUoWInvalid(t *testing.T) {
	f := newTestFixture(t)

	// Create one valid UoW
	f.createUoWManifestWithArtifacts(core.ActionBuild, "eac", "go", "go", map[string][]byte{
		"binary": []byte("binary content"),
	})

	// Create one UoW with missing artifacts
	manifest := f.createUoWManifest(core.ActionBuild, "eac", "docker", "docker")
	manifest.Artifacts = []Artifact{
		{ID: "image", Path: "missing-image.tar", SHA256: "sha256:missing", Size: 1000, Type: "docker-image"},
	}
	dirName := manifest.DirName()
	manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "eac", dirName, "uow.manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "eac", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	reader := NewReader(f.workspaceRoot)
	result := reader.ValidateModule(core.ActionBuild, "eac", expectedUoWs)
	assert.False(t, result.Valid)
}

func TestOutputReader_ValidateModule_SucceedsWithEmptyExpectedList(t *testing.T) {
	f := newTestFixture(t)

	expectedUoWs := []workunit.UnitID{}

	reader := NewReader(f.workspaceRoot)
	result := reader.ValidateModule(core.ActionBuild, "eac", expectedUoWs)
	assert.True(t, result.Valid) // No expected UoWs, so validation passes
	_ = f
}

func TestOutputReader_ValidateModule_IgnoresUnexpectedUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create more UoWs than expected
	f.createUoWManifest(core.ActionBuild, "eac", "go", "go")
	f.createUoWManifest(core.ActionBuild, "eac", "docker", "docker")
	f.createUoWManifest(core.ActionBuild, "eac", "extra", "extra") // Unexpected

	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "eac", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}

	reader := NewReader(f.workspaceRoot)
	result := reader.ValidateModule(core.ActionBuild, "eac", expectedUoWs)
	assert.True(t, result.Valid) // Extra UoWs should not cause failure
}

func TestOutputReader_ValidateModule_AllContexts(t *testing.T) {
	contexts := []core.ActionType{
		core.ActionBuild,
		core.ActionTest,
		core.ActionLint,
		core.ActionScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			f := newTestFixture(t)

			f.createUoWManifest(ctx, "module", "component", "tool")

			expectedUoWs := []workunit.UnitID{
				{Action: ctx, Module: "module", ComponentType: "component", ComponentName: "component", Tool: "tool"},
			}

			reader := NewReader(f.workspaceRoot)
			result := reader.ValidateModule(ctx, "module", expectedUoWs)
			assert.True(t, result.Valid)
		})
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestOutputReader_ConcurrentGetUoW(t *testing.T) {
	f := newTestFixture(t)

	// Create multiple UoWs
	f.createUoWManifest(core.ActionBuild, "module1", "go", "go")
	f.createUoWManifest(core.ActionBuild, "module2", "go", "go")
	f.createUoWManifest(core.ActionBuild, "module3", "go", "go")

	reader := NewReader(f.workspaceRoot)

	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(moduleNum int) {
			defer wg.Done()
			module := "module" + string(rune('0'+moduleNum))
			id := workunit.UnitID{Action: core.ActionBuild, Module: module, ComponentType: "go", ComponentName: "go", Tool: "go"}
			manifest, err := reader.GetUoW(id)
			assert.NoError(t, err)
			assert.Equal(t, module, manifest.Module)
		}(i)
	}
	wg.Wait()
}

func TestOutputReader_ConcurrentListUoWs(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs in the module
	f.createUoWManifest(core.ActionBuild, "module", "comp1", "tool1")
	f.createUoWManifest(core.ActionBuild, "module", "comp2", "tool2")

	reader := NewReader(f.workspaceRoot)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manifests, err := reader.ListUoWs(core.ActionBuild, "module")
			assert.NoError(t, err)
			assert.Len(t, manifests, 2)
		}()
	}
	wg.Wait()
}

func TestOutputReader_ConcurrentValidateUoW(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs with artifacts
	f.createUoWManifestWithArtifacts(core.ActionBuild, "module1", "go", "go", map[string][]byte{
		"binary": []byte("content1"),
	})
	f.createUoWManifestWithArtifacts(core.ActionBuild, "module2", "go", "go", map[string][]byte{
		"binary": []byte("content2"),
	})

	reader := NewReader(f.workspaceRoot)

	var wg sync.WaitGroup
	for i := 1; i <= 2; i++ {
		for j := 0; j < 5; j++ {
			wg.Add(1)
			go func(moduleNum int) {
				defer wg.Done()
				module := "module" + string(rune('0'+moduleNum))
				id := workunit.UnitID{Action: core.ActionBuild, Module: module, ComponentType: "go", ComponentName: "go", Tool: "go"}
				result := reader.ValidateUoW(id)
				assert.True(t, result.Valid)
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

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "go"}
	manifest, err := reader.GetUoW(id)
	assert.Error(t, err)
	assert.Nil(t, manifest)
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

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "go"}
	_, err = reader.GetUoW(id)
	assert.Error(t, err)
}

func TestOutputReader_ListUoWs_HandlesEmptyModule(t *testing.T) {
	f := newTestFixture(t)

	// Create module directory but no UoW directories
	moduleDir := filepath.Join(f.workspaceRoot, "out", "build", "empty-module")
	err := os.MkdirAll(moduleDir, 0755)
	require.NoError(t, err)

	reader := NewReader(f.workspaceRoot)
	manifests, err := reader.ListUoWs(core.ActionBuild, "empty-module")
	require.NoError(t, err)
	assert.Empty(t, manifests)
}

func TestOutputReader_GetModule_HandlesMixedValidAndInvalidManifests(t *testing.T) {
	f := newTestFixture(t)

	// Create valid manifest
	f.createUoWManifest(core.ActionBuild, "mixed-module", "go", "go")

	// Create invalid manifest
	invalidDir := filepath.Join(f.workspaceRoot, "out", "build", "mixed-module", "docker_docker")
	err := os.MkdirAll(invalidDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(invalidDir, "uow.manifest.json"), []byte("invalid"), 0644)
	require.NoError(t, err)

	reader := NewReader(f.workspaceRoot)
	view, err := reader.GetModule(core.ActionBuild, "mixed-module")
	require.NoError(t, err)
	// Should return view with only valid UoWs
	assert.Len(t, view.Components, 1)
}

func TestOutputReader_WorksWithSpecialCharactersInNames(t *testing.T) {
	f := newTestFixture(t)

	// Module and component names with hyphens
	f.createUoWManifest(core.ActionBuild, "my-complex-module", "go-component", "my-tool")

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "my-complex-module", ComponentType: "go-component", ComponentName: "go-component", Tool: "my-tool"}
	manifest, err := reader.GetUoW(id)
	require.NoError(t, err)
	assert.Equal(t, "my-complex-module", manifest.Module)
}

func TestOutputReader_ValidateUoW_MultipleArtifacts(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW with multiple artifacts
	f.createUoWManifestWithArtifacts(core.ActionBuild, "test-module", "go", "go", map[string][]byte{
		"eac-linux-amd64":       make([]byte, 1000),
		"eac-darwin-amd64":      make([]byte, 1100),
		"eac-windows-amd64.exe": make([]byte, 1200),
	})

	reader := NewReader(f.workspaceRoot)
	id := workunit.UnitID{Action: core.ActionBuild, Module: "test-module", ComponentType: "go", ComponentName: "go", Tool: "go"}
	result := reader.ValidateUoW(id)
	assert.True(t, result.Valid)
	assert.True(t, result.ArtifactsValid)
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestOutputReader_FullWorkflow_BuildModule(t *testing.T) {
	f := newTestFixture(t)

	// Simulate a full build output structure
	// eac module with go and docker components

	// Go component
	f.createUoWManifestWithArtifacts(core.ActionBuild, "eac", "go", "go", map[string][]byte{
		"eac-linux-amd64":       make([]byte, 10000),
		"eac-darwin-amd64":      make([]byte, 11000),
		"eac-windows-amd64.exe": make([]byte, 12000),
	})

	// Docker component
	f.createUoWManifestWithArtifacts(core.ActionBuild, "eac", "docker", "docker", map[string][]byte{
		"image.tar": make([]byte, 50000),
	})

	reader := NewReader(f.workspaceRoot)

	// 1. List all UoWs
	manifests, err := reader.ListUoWs(core.ActionBuild, "eac")
	require.NoError(t, err)
	assert.Len(t, manifests, 2)

	// 2. Get module view
	moduleView, err := reader.GetModule(core.ActionBuild, "eac")
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, moduleView.Status)
	assert.Len(t, moduleView.Components, 2)

	// 3. Validate module
	expectedUoWs := []workunit.UnitID{
		{Action: core.ActionBuild, Module: "eac", ComponentType: "go", ComponentName: "go", Tool: "go"},
		{Action: core.ActionBuild, Module: "eac", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
	}
	result := reader.ValidateModule(core.ActionBuild, "eac", expectedUoWs)
	assert.True(t, result.Valid)
}

func TestOutputReader_FullWorkflow_TestModule(t *testing.T) {
	f := newTestFixture(t)

	// Simulate test output structure
	f.createUoWManifestWithExitCode(core.ActionTest, "core", "go", "gotest", 0)
	f.createUoWManifestWithExitCode(core.ActionTest, "core", "gherkin", "godog", 0)

	reader := NewReader(f.workspaceRoot)

	moduleView, err := reader.GetModule(core.ActionTest, "core")
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, moduleView.Status)
}

func TestOutputReader_FullWorkflow_PartialFailure(t *testing.T) {
	f := newTestFixture(t)

	// Simulate partial test failure
	f.createUoWManifestWithExitCode(core.ActionTest, "core", "go", "gotest", 0)     // Pass
	f.createUoWManifestWithExitCode(core.ActionTest, "core", "gherkin", "godog", 1) // Fail

	reader := NewReader(f.workspaceRoot)

	moduleView, err := reader.GetModule(core.ActionTest, "core")
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, moduleView.Status)
}

func TestOutputReader_FullWorkflow_MultipleModules(t *testing.T) {
	f := newTestFixture(t)

	// Create UoWs for multiple modules
	f.createUoWManifest(core.ActionBuild, "module1", "go", "go")
	f.createUoWManifest(core.ActionBuild, "module2", "go", "go")
	f.createUoWManifest(core.ActionBuild, "module3", "go", "go")

	reader := NewReader(f.workspaceRoot)

	for i := 1; i <= 3; i++ {
		module := "module" + string(rune('0'+i))
		view, err := reader.GetModule(core.ActionBuild, module)
		require.NoError(t, err)
		assert.Equal(t, module, view.Module)
	}
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
				f.createUoWManifestWithExitCode(core.ActionBuild, "module", component, tool, exitCode)
			}

			reader := NewReader(f.workspaceRoot)
			view, err := reader.GetModule(core.ActionBuild, "module")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, view.Status)
		})
	}
}

func TestOutputReader_ValidateModule_Comprehensive(t *testing.T) {
	tests := []struct {
		name      string
		setupUoWs []struct {
			component string
			tool      string
			artifacts map[string][]byte
			wrongHash bool
		}
		expectedUoWs []workunit.UnitID
		expectValid  bool
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
				{Action: core.ActionBuild, Module: "module", ComponentType: "go", ComponentName: "go", Tool: "go"},
				{Action: core.ActionBuild, Module: "module", ComponentType: "docker", ComponentName: "docker", Tool: "docker"},
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
				{Action: core.ActionBuild, Module: "module", ComponentType: "go", ComponentName: "go", Tool: "go"},
				{Action: core.ActionBuild, Module: "module", ComponentType: "docker", ComponentName: "docker", Tool: "docker"}, // Missing
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
				{Action: core.ActionBuild, Module: "module", ComponentType: "go", ComponentName: "go", Tool: "go"},
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTestFixture(t)

			// Setup UoWs
			for _, uow := range tt.setupUoWs {
				manifest := f.createUoWManifestWithArtifacts(core.ActionBuild, "module", uow.component, uow.tool, uow.artifacts)

				if uow.wrongHash && len(manifest.Artifacts) > 0 {
					manifest.Artifacts[0].SHA256 = "sha256:wrong"
					dirName := manifest.DirName()
					manifestPath := filepath.Join(f.workspaceRoot, "out", "build", "module", dirName, "uow.manifest.json")
					data, err := json.MarshalIndent(manifest, "", "  ")
					require.NoError(t, err)
					err = os.WriteFile(manifestPath, data, 0644)
					require.NoError(t, err)
				}
			}

			reader := NewReader(f.workspaceRoot)
			result := reader.ValidateModule(core.ActionBuild, "module", tt.expectedUoWs)
			assert.Equal(t, tt.expectValid, result.Valid)
		})
	}
}
