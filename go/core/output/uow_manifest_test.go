package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ManifestPath() Tests
// =============================================================================

func TestUoWManifest_ManifestPath_BuildContext(t *testing.T) {
	manifest := UoWManifest{
		Action:    core.ActionBuild,
		Module:    "core",
		Component: "go",
		Tool:      "go",
	}

	workspaceRoot := "/workspace"
	path := manifest.ManifestPath(workspaceRoot)

	expected := filepath.Join(workspaceRoot, "out", "build", "core", "go_go", "uow.manifest.json")
	assert.Equal(t, expected, path)
}

func TestUoWManifest_ManifestPath_TestContext(t *testing.T) {
	manifest := UoWManifest{
		Action:    core.ActionTest,
		Module:    "core",
		Component: "go",
		Tool:      "gotest",
	}

	workspaceRoot := "/workspace"
	path := manifest.ManifestPath(workspaceRoot)

	expected := filepath.Join(workspaceRoot, "out", "test", "core", "go_gotest", "uow.manifest.json")
	assert.Equal(t, expected, path)
}

func TestUoWManifest_ManifestPath_LintContext(t *testing.T) {
	manifest := UoWManifest{
		Action:    core.ActionLint,
		Module:    "web-app",
		Component: "typescript",
		Tool:      "eslint",
	}

	workspaceRoot := "/workspace"
	path := manifest.ManifestPath(workspaceRoot)

	expected := filepath.Join(workspaceRoot, "out", "lint", "web-app", "typescript_eslint", "uow.manifest.json")
	assert.Equal(t, expected, path)
}

func TestUoWManifest_ManifestPath_ScanContext(t *testing.T) {
	manifest := UoWManifest{
		Action:    core.ActionScan,
		Module:    "eac",
		Component: "docker",
		Tool:      "trivy-vuln",
	}

	workspaceRoot := "/workspace"
	path := manifest.ManifestPath(workspaceRoot)

	expected := filepath.Join(workspaceRoot, "out", "scan", "eac", "docker_trivy-vuln", "uow.manifest.json")
	assert.Equal(t, expected, path)
}

func TestUoWManifest_ManifestPath_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		context       core.ActionType
		module        string
		component     string
		tool          string
		workspaceRoot string
		expectedPath  string
	}{
		{
			name:          "build go module",
			context:       core.ActionBuild,
			module:        "core",
			component:     "go",
			tool:          "go",
			workspaceRoot: "/project",
			expectedPath:  filepath.Join("/project", "out", "build", "core", "go_go", "uow.manifest.json"),
		},
		{
			name:          "test gherkin with godog",
			context:       core.ActionTest,
			module:        "eac",
			component:     "gherkin",
			tool:          "godog",
			workspaceRoot: "/project",
			expectedPath:  filepath.Join("/project", "out", "test", "eac", "gherkin_godog", "uow.manifest.json"),
		},
		{
			name:          "lint assets files",
			context:       core.ActionLint,
			module:        "configs",
			component:     "assets",
			tool:          "yamllint",
			workspaceRoot: "/workspace/repo",
			expectedPath:  filepath.Join("/workspace/repo", "out", "lint", "configs", "assets_yamllint", "uow.manifest.json"),
		},
		{
			name:          "scan for secrets",
			context:       core.ActionScan,
			module:        "eac-web",
			component:     "source",
			tool:          "trivy-secret",
			workspaceRoot: "/home/user/project",
			expectedPath:  filepath.Join("/home/user/project", "out", "scan", "eac-web", "source_trivy-secret", "uow.manifest.json"),
		},
		{
			name:          "module with hyphens",
			context:       core.ActionBuild,
			module:        "my-complex-module",
			component:     "go",
			tool:          "go",
			workspaceRoot: "/proj",
			expectedPath:  filepath.Join("/proj", "out", "build", "my-complex-module", "go_go", "uow.manifest.json"),
		},
		{
			name:          "tool with hyphens",
			context:       core.ActionLint,
			module:        "core",
			component:     "go",
			tool:          "golangci-lint",
			workspaceRoot: "/proj",
			expectedPath:  filepath.Join("/proj", "out", "lint", "core", "go_golangci-lint", "uow.manifest.json"),
		},
		{
			name:          "empty workspace root",
			context:       core.ActionBuild,
			module:        "mod",
			component:     "comp",
			tool:          "tool",
			workspaceRoot: "",
			expectedPath:  filepath.Join("out", "build", "mod", "comp_tool", "uow.manifest.json"),
		},
		{
			name:          "relative workspace root",
			context:       core.ActionBuild,
			module:        "mod",
			component:     "comp",
			tool:          "tool",
			workspaceRoot: ".",
			expectedPath:  filepath.Join(".", "out", "build", "mod", "comp_tool", "uow.manifest.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := UoWManifest{
				Action:    tt.context,
				Module:    tt.module,
				Component: tt.component,
				Tool:      tt.tool,
			}

			path := manifest.ManifestPath(tt.workspaceRoot)
			assert.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestUoWManifest_ManifestPath_EndsWithManifestJSON(t *testing.T) {
	contexts := []core.ActionType{
		core.ActionBuild,
		core.ActionTest,
		core.ActionLint,
		core.ActionScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			manifest := UoWManifest{
				Action:    ctx,
				Module:    "module",
				Component: "component",
				Tool:      "tool",
			}

			path := manifest.ManifestPath("/workspace")
			assert.True(t, filepath.Base(path) == "uow.manifest.json",
				"Path should end with uow.manifest.json, got: %s", filepath.Base(path))
		})
	}
}

// =============================================================================
// Save() Tests
// =============================================================================

func TestUoWManifest_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "test-module",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:input123",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:output456",
		Version:    "1.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err)

	// Verify directory was created
	manifestPath := manifest.ManifestPath(tmpDir)
	manifestDir := filepath.Dir(manifestPath)
	info, err := os.Stat(manifestDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestUoWManifest_Save_WritesValidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionTest,
		Module:     "core",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   0,
		InputHash:  "sha256:test-input",
		ExecutedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Duration:   60 * time.Second,
		Artifacts: []Artifact{
			{ID: "junit", Path: "junit.xml", SHA256: "sha256:junit", Size: 1000, Type: "report"},
		},
		OutputHash: "sha256:test-output",
		Version:    "1.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err)

	// Read and verify JSON
	manifestPath := manifest.ManifestPath(tmpDir)
	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var loaded map[string]any
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err, "Saved file should be valid JSON")
}

func TestUoWManifest_Save_ContainsAllFields(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "my-module",
		Component:  "my-component",
		Tool:       "my-tool",
		ExitCode:   42,
		InputHash:  "sha256:my-input-hash",
		ExecutedAt: time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC),
		Duration:   123 * time.Second,
		Artifacts: []Artifact{
			{ID: "art1", Path: "path1", SHA256: "sha256:art1", Size: 100, Type: "binary"},
			{ID: "art2", Path: "path2", SHA256: "sha256:art2", Size: 200, Type: "report"},
		},
		OutputHash: "sha256:my-output-hash",
		Version:    "2.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err)

	manifestPath := manifest.ManifestPath(tmpDir)
	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "my-module")
	assert.Contains(t, jsonStr, "my-component")
	assert.Contains(t, jsonStr, "my-tool")
	assert.Contains(t, jsonStr, "my-input-hash")
	assert.Contains(t, jsonStr, "my-output-hash")
	assert.Contains(t, jsonStr, "art1")
	assert.Contains(t, jsonStr, "art2")
	assert.Contains(t, jsonStr, "2.0.0")
}

func TestUoWManifest_Save_OverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()

	manifest1 := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "mod",
		Component:  "comp",
		Tool:       "tool",
		ExitCode:   0,
		InputHash:  "sha256:first-hash",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   10 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:first-output",
		Version:    "1.0.0",
	}

	err := manifest1.Save(tmpDir)
	require.NoError(t, err)

	manifest2 := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "mod",
		Component:  "comp",
		Tool:       "tool",
		ExitCode:   1,
		InputHash:  "sha256:second-hash",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   20 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:second-output",
		Version:    "1.0.0",
	}

	err = manifest2.Save(tmpDir)
	require.NoError(t, err)

	// Load and verify it's the second manifest
	manifestPath := manifest2.ManifestPath(tmpDir)
	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var loaded UoWManifest
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, "sha256:second-hash", loaded.InputHash)
	assert.Equal(t, "sha256:second-output", loaded.OutputHash)
	assert.Equal(t, 1, loaded.ExitCode)
}

func TestUoWManifest_Save_WithEmptyArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionLint,
		Module:     "mod",
		Component:  "go",
		Tool:       "golangci-lint",
		ExitCode:   0,
		InputHash:  "sha256:lint-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   5 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:lint-output",
		Version:    "1.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err)

	manifestPath := manifest.ManifestPath(tmpDir)
	_, err = os.Stat(manifestPath)
	require.NoError(t, err, "Manifest should be saved even with empty artifacts")
}

func TestUoWManifest_Save_WithNilArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionLint,
		Module:     "mod",
		Component:  "go",
		Tool:       "golangci-lint",
		ExitCode:   0,
		InputHash:  "sha256:lint-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   5 * time.Second,
		Artifacts:  nil, // nil artifacts
		OutputHash: "sha256:lint-output",
		Version:    "1.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err)

	manifestPath := manifest.ManifestPath(tmpDir)
	_, err = os.Stat(manifestPath)
	require.NoError(t, err, "Manifest should be saved even with nil artifacts")
}

func TestUoWManifest_Save_MultipleArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "eac",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:multi-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   45 * time.Second,
		Artifacts: []Artifact{
			{ID: "eac-linux-amd64", Path: "out/build/eac/go/eac-linux-amd64", SHA256: "sha256:linux", Size: 10000000, Type: "binary"},
			{ID: "eac-darwin-amd64", Path: "out/build/eac/go/eac-darwin-amd64", SHA256: "sha256:darwin", Size: 11000000, Type: "binary"},
			{ID: "eac-windows-amd64", Path: "out/build/eac/go/eac-windows-amd64.exe", SHA256: "sha256:windows", Size: 12000000, Type: "binary"},
		},
		OutputHash: "sha256:multi-output",
		Version:    "1.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err)

	// Load and verify artifacts
	manifestPath := manifest.ManifestPath(tmpDir)
	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var loaded UoWManifest
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Len(t, loaded.Artifacts, 3)
}

func TestUoWManifest_Save_FailedExecution(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionTest,
		Module:     "failing-module",
		Component:  "go",
		Tool:       "gotest",
		ExitCode:   1, // Non-zero exit code
		InputHash:  "sha256:failed-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts: []Artifact{
			{ID: "junit", Path: "junit.xml", SHA256: "sha256:junit", Size: 2000, Type: "report"},
		},
		OutputHash: "sha256:failed-output",
		Version:    "1.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err, "Should save manifest even for failed executions")

	manifestPath := manifest.ManifestPath(tmpDir)
	data, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	var loaded UoWManifest
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, 1, loaded.ExitCode)
}

// =============================================================================
// Load() Tests
// =============================================================================

func TestLoad_ReadsValidManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest directory
	manifestDir := filepath.Join(tmpDir, "out", "build", "core", "go_go")
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	// Write manifest file
	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	manifestData := `{
		"context": "build",
		"module": "core",
		"component": "go",
		"tool": "go",
		"exit_code": 0,
		"input_hash": "sha256:input",
		"executed_at": "2024-01-15T10:30:00Z",
		"duration": 30000000000,
		"artifacts": [
			{"id": "binary", "path": "out/binary", "sha256": "sha256:bin", "size": 1000, "type": "binary"}
		],
		"output_hash": "sha256:output",
		"version": "1.0.0"
	}`
	err = os.WriteFile(manifestPath, []byte(manifestData), 0644)
	require.NoError(t, err)

	// Load manifest
	loaded, err := Load(manifestPath)
	require.NoError(t, err)

	assert.Equal(t, core.ActionBuild, loaded.Action)
	assert.Equal(t, "core", loaded.Module)
	assert.Equal(t, "go", loaded.Component)
	assert.Equal(t, "go", loaded.Tool)
	assert.Equal(t, 0, loaded.ExitCode)
	assert.Equal(t, "sha256:input", loaded.InputHash)
	assert.Equal(t, 30*time.Second, loaded.Duration)
	assert.Len(t, loaded.Artifacts, 1)
	assert.Equal(t, "sha256:output", loaded.OutputHash)
	assert.Equal(t, "1.0.0", loaded.Version)
}

func TestLoad_ReturnsErrorIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "out", "build", "nonexistent", "go_go", "uow.manifest.json")

	_, err := Load(nonExistentPath)
	assert.Error(t, err, "Load should return error when file doesn't exist")
}

func TestLoad_ReturnsErrorForInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest directory
	manifestDir := filepath.Join(tmpDir, "out", "build", "invalid", "go_go")
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	// Write invalid JSON
	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	err = os.WriteFile(manifestPath, []byte("not valid json {{{"), 0644)
	require.NoError(t, err)

	_, err = Load(manifestPath)
	assert.Error(t, err, "Load should return error for invalid JSON")
}

func TestLoad_ReturnsErrorForEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest directory
	manifestDir := filepath.Join(tmpDir, "out", "build", "empty", "go_go")
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	// Write empty file
	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	err = os.WriteFile(manifestPath, []byte(""), 0644)
	require.NoError(t, err)

	_, err = Load(manifestPath)
	assert.Error(t, err, "Load should return error for empty file")
}

func TestLoad_HandlesPartialJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest directory
	manifestDir := filepath.Join(tmpDir, "out", "build", "partial", "go_go")
	err := os.MkdirAll(manifestDir, 0755)
	require.NoError(t, err)

	// Write partial/minimal JSON
	manifestPath := filepath.Join(manifestDir, "uow.manifest.json")
	manifestData := `{
		"context": "build",
		"module": "partial",
		"component": "go",
		"tool": "go"
	}`
	err = os.WriteFile(manifestPath, []byte(manifestData), 0644)
	require.NoError(t, err)

	loaded, err := Load(manifestPath)
	require.NoError(t, err, "Load should succeed with partial JSON")

	assert.Equal(t, core.ActionBuild, loaded.Action)
	assert.Equal(t, "partial", loaded.Module)
	assert.Zero(t, loaded.ExitCode, "Missing fields should have zero value")
	assert.Empty(t, loaded.InputHash)
	assert.Nil(t, loaded.Artifacts)
}

// =============================================================================
// Save/Load Round-Trip Tests
// =============================================================================

func TestUoWManifest_SaveLoadRoundTrip_PreservesAllFields(t *testing.T) {
	tmpDir := t.TempDir()

	original := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "roundtrip-module",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:roundtrip-input-hash",
		ExecutedAt: time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC),
		Duration:   123 * time.Second,
		Artifacts: []Artifact{
			{ID: "bin1", Path: "out/bin1", SHA256: "sha256:bin1hash", Size: 1000, Type: "binary"},
			{ID: "bin2", Path: "out/bin2", SHA256: "sha256:bin2hash", Size: 2000, Type: "binary"},
		},
		OutputHash: "sha256:roundtrip-output-hash",
		Version:    "1.2.3",
	}

	// Save
	err := original.Save(tmpDir)
	require.NoError(t, err)

	// Load
	manifestPath := original.ManifestPath(tmpDir)
	loaded, err := Load(manifestPath)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, original.Action, loaded.Action)
	assert.Equal(t, original.Module, loaded.Module)
	assert.Equal(t, original.Component, loaded.Component)
	assert.Equal(t, original.Tool, loaded.Tool)
	assert.Equal(t, original.ExitCode, loaded.ExitCode)
	assert.Equal(t, original.InputHash, loaded.InputHash)
	assert.True(t, original.ExecutedAt.Equal(loaded.ExecutedAt), "ExecutedAt should be preserved")
	assert.Equal(t, original.Duration, loaded.Duration)
	assert.Equal(t, original.Artifacts, loaded.Artifacts)
	assert.Equal(t, original.OutputHash, loaded.OutputHash)
	assert.Equal(t, original.Version, loaded.Version)
}

func TestUoWManifest_SaveLoadRoundTrip_AllContexts(t *testing.T) {
	contexts := []core.ActionType{
		core.ActionBuild,
		core.ActionTest,
		core.ActionLint,
		core.ActionScan,
	}

	for _, ctx := range contexts {
		t.Run(string(ctx), func(t *testing.T) {
			tmpDir := t.TempDir()

			original := UoWManifest{
				Action:     ctx,
				Module:     "ctx-test-module",
				Component:  "go",
				Tool:       "tool",
				ExitCode:   0,
				InputHash:  "sha256:" + string(ctx) + "-input",
				ExecutedAt: time.Now().UTC().Truncate(time.Second),
				Duration:   10 * time.Second,
				Artifacts:  []Artifact{},
				OutputHash: "sha256:" + string(ctx) + "-output",
				Version:    "1.0.0",
			}

			err := original.Save(tmpDir)
			require.NoError(t, err)

			manifestPath := original.ManifestPath(tmpDir)
			loaded, err := Load(manifestPath)
			require.NoError(t, err)

			assert.Equal(t, ctx, loaded.Action)
		})
	}
}

func TestUoWManifest_SaveLoadRoundTrip_LargeArtifactList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manifest with many artifacts
	artifacts := make([]Artifact, 100)
	for i := range artifacts {
		artifacts[i] = Artifact{
			ID:     "artifact-" + string(rune('0'+i%10)) + string(rune('0'+i/10)),
			Path:   "out/artifacts/file" + string(rune('0'+i%10)) + string(rune('0'+i/10)),
			SHA256: "sha256:hash" + string(rune('0'+i%10)) + string(rune('0'+i/10)),
			Size:   int64(i * 1000),
			Type:   "binary",
		}
	}

	original := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "large-module",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:large-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   300 * time.Second,
		Artifacts:  artifacts,
		OutputHash: "sha256:large-output",
		Version:    "1.0.0",
	}

	err := original.Save(tmpDir)
	require.NoError(t, err)

	manifestPath := original.ManifestPath(tmpDir)
	loaded, err := Load(manifestPath)
	require.NoError(t, err)

	assert.Len(t, loaded.Artifacts, 100)
}

func TestUoWManifest_SaveLoadRoundTrip_ZeroDuration(t *testing.T) {
	tmpDir := t.TempDir()

	original := UoWManifest{
		Action:     core.ActionLint,
		Module:     "fast-module",
		Component:  "assets",
		Tool:       "yamllint",
		ExitCode:   0,
		InputHash:  "sha256:fast-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   0, // Zero duration
		Artifacts:  []Artifact{},
		OutputHash: "sha256:fast-output",
		Version:    "1.0.0",
	}

	err := original.Save(tmpDir)
	require.NoError(t, err)

	manifestPath := original.ManifestPath(tmpDir)
	loaded, err := Load(manifestPath)
	require.NoError(t, err)

	assert.Zero(t, loaded.Duration)
}

func TestUoWManifest_SaveLoadRoundTrip_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()

	original := UoWManifest{
		Action:     core.ActionBuild,
		Module:     "module-with-dashes",
		Component:  "go",
		Tool:       "go",
		ExitCode:   0,
		InputHash:  "sha256:abc123/def456+ghi789=",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   10 * time.Second,
		Artifacts: []Artifact{
			{
				ID:     "file with spaces",
				Path:   "out/path with spaces/file.bin",
				SHA256: "sha256:special",
				Size:   100,
				Type:   "binary",
			},
		},
		OutputHash: "sha256:xyz789",
		Version:    "1.0.0-beta+build.123",
	}

	err := original.Save(tmpDir)
	require.NoError(t, err)

	manifestPath := original.ManifestPath(tmpDir)
	loaded, err := Load(manifestPath)
	require.NoError(t, err)

	assert.Equal(t, original.InputHash, loaded.InputHash)
	assert.Equal(t, original.Artifacts[0].ID, loaded.Artifacts[0].ID)
	assert.Equal(t, original.Artifacts[0].Path, loaded.Artifacts[0].Path)
	assert.Equal(t, original.Version, loaded.Version)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestUoWManifest_ManifestPath_WithDots(t *testing.T) {
	manifest := UoWManifest{
		Action:    core.ActionBuild,
		Module:    "v1.2.3",
		Component: "go",
		Tool:      "go",
	}

	path := manifest.ManifestPath("/workspace")
	assert.Contains(t, path, "v1.2.3")
}

func TestUoWManifest_Save_CreatesNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := UoWManifest{
		Action:     core.ActionScan,
		Module:     "deeply-nested",
		Component:  "docker",
		Tool:       "trivy-vuln",
		ExitCode:   0,
		InputHash:  "sha256:nested",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   60 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:nested-out",
		Version:    "1.0.0",
	}

	err := manifest.Save(tmpDir)
	require.NoError(t, err)

	// Verify the full directory path exists
	manifestPath := manifest.ManifestPath(tmpDir)
	manifestDir := filepath.Dir(manifestPath)
	info, err := os.Stat(manifestDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestUoWManifest_IsolatesBetweenTools(t *testing.T) {
	tmpDir := t.TempDir()

	manifest1 := UoWManifest{
		Action:     core.ActionScan,
		Module:     "core",
		Component:  "go",
		Tool:       "trivy-vuln",
		ExitCode:   0,
		InputHash:  "sha256:trivy-vuln-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:trivy-vuln-output",
		Version:    "1.0.0",
	}

	manifest2 := UoWManifest{
		Action:     core.ActionScan,
		Module:     "core",
		Component:  "go",
		Tool:       "trivy-secret",
		ExitCode:   0,
		InputHash:  "sha256:trivy-secret-input",
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   20 * time.Second,
		Artifacts:  []Artifact{},
		OutputHash: "sha256:trivy-secret-output",
		Version:    "1.0.0",
	}

	err := manifest1.Save(tmpDir)
	require.NoError(t, err)

	err = manifest2.Save(tmpDir)
	require.NoError(t, err)

	// Verify different paths
	path1 := manifest1.ManifestPath(tmpDir)
	path2 := manifest2.ManifestPath(tmpDir)
	assert.NotEqual(t, path1, path2)

	// Verify both can be loaded independently
	loaded1, err := Load(path1)
	require.NoError(t, err)
	assert.Equal(t, "trivy-vuln", loaded1.Tool)

	loaded2, err := Load(path2)
	require.NoError(t, err)
	assert.Equal(t, "trivy-secret", loaded2.Tool)
}
