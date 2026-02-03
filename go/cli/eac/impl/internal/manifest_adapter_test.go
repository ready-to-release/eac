package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	coreoutput "github.com/ready-to-release/eac/go/core/output"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Fixtures
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
func (f *testFixture) createUoWManifest(ctx workunit.Context, module, component, tool string) *coreoutput.UoWManifest {
	f.t.Helper()
	manifest := &coreoutput.UoWManifest{
		Context:    ctx,
		Module:     module,
		Component:  component,
		Tool:       tool,
		ExitCode:   0,
		InputHash:  "sha256:input-" + module + "-" + component + "-" + tool,
		ExecutedAt: time.Now().UTC().Truncate(time.Second),
		Duration:   30 * time.Second,
		Artifacts:  []coreoutput.Artifact{},
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
func (f *testFixture) createUoWManifestWithArtifacts(ctx workunit.Context, module, component, tool string, artifacts []coreoutput.Artifact) *coreoutput.UoWManifest {
	f.t.Helper()

	dirName := component + "_" + tool
	uowDir := filepath.Join(f.workspaceRoot, "out", string(ctx), module, dirName)
	err := os.MkdirAll(uowDir, 0755)
	require.NoError(f.t, err)

	manifest := &coreoutput.UoWManifest{
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

// =============================================================================
// ConvertModuleViewToManifest Tests
// =============================================================================

func TestConvertModuleViewToManifest_BasicConversion(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	view := &coreoutput.ModuleView{
		Module:    "test-module",
		Status:    coreoutput.StatusCompleted,
		TotalSize: 5000,
		Components: []coreoutput.ComponentView{
			{
				Module:    "test-module",
				Component: "go",
				Status:    coreoutput.StatusCompleted,
				TotalSize: 5000,
				UoWs: []coreoutput.UoWManifest{
					{
						Context:    workunit.ContextBuild,
						Module:     "test-module",
						Component:  "go",
						Tool:       "go",
						ExitCode:   0,
						InputHash:  "sha256:input123",
						ExecutedAt: now,
						Duration:   30 * time.Second,
						Artifacts: []coreoutput.Artifact{
							{ID: "binary", Path: "eac-linux-amd64", SHA256: "sha256:abc", Size: 5000, Type: "binary"},
						},
					},
				},
			},
		},
	}

	manifest := ConvertModuleViewToManifest(view, "cli-app")

	require.NotNil(t, manifest)
	assert.Equal(t, "test-module", manifest.Moniker)
	assert.Equal(t, "cli-app", manifest.Type)
	assert.Equal(t, "sha256:input123", manifest.InputHash)
	assert.Equal(t, float64(30), manifest.DurationSeconds)
	assert.Len(t, manifest.Artifacts, 1)
	assert.Equal(t, "binary", manifest.Artifacts[0].ID)
	assert.Equal(t, "eac-linux-amd64", manifest.Artifacts[0].Path)
	assert.Equal(t, "linux-amd64", manifest.Artifacts[0].Platform)
}

func TestConvertModuleViewToManifest_MultipleComponents(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	view := &coreoutput.ModuleView{
		Module:    "test-module",
		Status:    coreoutput.StatusCompleted,
		TotalSize: 15000,
		Components: []coreoutput.ComponentView{
			{
				Module:    "test-module",
				Component: "go",
				Status:    coreoutput.StatusCompleted,
				TotalSize: 5000,
				UoWs: []coreoutput.UoWManifest{
					{
						Context:    workunit.ContextBuild,
						Module:     "test-module",
						Component:  "go",
						Tool:       "go",
						ExitCode:   0,
						ExecutedAt: now,
						Duration:   20 * time.Second,
						Artifacts: []coreoutput.Artifact{
							{ID: "go-binary", Path: "app", Size: 5000, Type: "binary"},
						},
					},
				},
			},
			{
				Module:    "test-module",
				Component: "docker",
				Status:    coreoutput.StatusCompleted,
				TotalSize: 10000,
				UoWs: []coreoutput.UoWManifest{
					{
						Context:    workunit.ContextBuild,
						Module:     "test-module",
						Component:  "docker",
						Tool:       "docker",
						ExitCode:   0,
						ExecutedAt: now.Add(20 * time.Second),
						Duration:   40 * time.Second,
						Artifacts: []coreoutput.Artifact{
							{ID: "image", Path: "image.tar", Size: 10000, Type: "docker-image"},
						},
					},
				},
			},
		},
	}

	manifest := ConvertModuleViewToManifest(view, "cli-app")

	require.NotNil(t, manifest)
	assert.Equal(t, float64(60), manifest.DurationSeconds) // 20 + 40 seconds
	assert.Len(t, manifest.Artifacts, 2)
}

func TestConvertModuleViewToManifest_EmptyView(t *testing.T) {
	view := &coreoutput.ModuleView{
		Module:     "empty-module",
		Status:     coreoutput.StatusPending,
		Components: []coreoutput.ComponentView{},
	}

	manifest := ConvertModuleViewToManifest(view, "library")

	require.NotNil(t, manifest)
	assert.Equal(t, "empty-module", manifest.Moniker)
	assert.Empty(t, manifest.Artifacts)
}

func TestConvertModuleViewToManifest_NilView(t *testing.T) {
	manifest := ConvertModuleViewToManifest(nil, "library")
	assert.Nil(t, manifest)
}

func TestConvertModuleViewToManifest_PlatformExtraction(t *testing.T) {
	view := &coreoutput.ModuleView{
		Module: "test-module",
		Components: []coreoutput.ComponentView{
			{
				UoWs: []coreoutput.UoWManifest{
					{
						Artifacts: []coreoutput.Artifact{
							{ID: "linux-bin", Path: "eac-linux-amd64", Type: "binary"},
							{ID: "darwin-bin", Path: "eac-darwin-arm64", Type: "binary"},
							{ID: "windows-bin", Path: "eac-windows-amd64.exe", Type: "binary"},
						},
					},
				},
			},
		},
	}

	manifest := ConvertModuleViewToManifest(view, "cli-app")

	require.NotNil(t, manifest)
	// Should have extracted 3 unique platforms
	assert.Len(t, manifest.Platforms, 3)

	// Check each platform
	platformMap := make(map[string]bool)
	for _, p := range manifest.Platforms {
		platformMap[p.OS+"-"+p.Arch] = true
	}
	assert.True(t, platformMap["linux-amd64"])
	assert.True(t, platformMap["darwin-arm64"])
	assert.True(t, platformMap["windows-amd64"])
}

// =============================================================================
// GetModuleManifestFromUoWs Tests
// =============================================================================

func TestGetModuleManifestFromUoWs_LoadsFromDisk(t *testing.T) {
	f := newTestFixture(t)

	// Create UoW manifests on disk
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "test-module", "go", "go", []coreoutput.Artifact{
		{ID: "binary", Path: "app", SHA256: "sha256:abc", Size: 1000, Type: "binary"},
	})

	manifest, err := GetModuleManifestFromUoWs(f.workspaceRoot, workunit.ContextBuild, "test-module", "cli-app")

	require.NoError(t, err)
	require.NotNil(t, manifest)
	assert.Equal(t, "test-module", manifest.Moniker)
	assert.Equal(t, "cli-app", manifest.Type)
	assert.Len(t, manifest.Artifacts, 1)
}

func TestGetModuleManifestFromUoWs_MultipleComponents(t *testing.T) {
	f := newTestFixture(t)

	// Create multiple UoW manifests
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "test-module", "go", "go", []coreoutput.Artifact{
		{ID: "go-binary", Path: "app", Type: "binary"},
	})
	f.createUoWManifestWithArtifacts(workunit.ContextBuild, "test-module", "docker", "docker", []coreoutput.Artifact{
		{ID: "image", Path: "image.tar", Type: "docker-image"},
	})

	manifest, err := GetModuleManifestFromUoWs(f.workspaceRoot, workunit.ContextBuild, "test-module", "cli-app")

	require.NoError(t, err)
	require.NotNil(t, manifest)
	assert.Len(t, manifest.Artifacts, 2)
}

func TestGetModuleManifestFromUoWs_NoUoWs(t *testing.T) {
	f := newTestFixture(t)

	// No UoWs created for module

	manifest, err := GetModuleManifestFromUoWs(f.workspaceRoot, workunit.ContextBuild, "nonexistent", "library")

	require.NoError(t, err)
	assert.Nil(t, manifest)
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestInferPlatformFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"eac-linux-amd64", "linux-amd64"},
		{"eac-linux-arm64", "linux-arm64"},
		{"eac-darwin-amd64", "darwin-amd64"},
		{"eac-darwin-arm64", "darwin-arm64"},
		{"eac-windows-amd64.exe", "windows-amd64"},
		{"bin/myapp-linux-amd64", "linux-amd64"},
		{"linux/amd64/binary", "linux-amd64"},
		{"just-a-file.txt", ""},
		{"no-platform", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := inferPlatformFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePlatformString(t *testing.T) {
	tests := []struct {
		platform     string
		expectedOS   string
		expectedArch string
	}{
		{"linux-amd64", "linux", "amd64"},
		{"darwin-arm64", "darwin", "arm64"},
		{"windows-amd64", "windows", "amd64"},
		{"linux/amd64", "linux", "amd64"},
		{"invalid", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			os, arch := parsePlatformString(tt.platform)
			assert.Equal(t, tt.expectedOS, os)
			assert.Equal(t, tt.expectedArch, arch)
		})
	}
}

func TestExtractPlatformsFromArtifacts(t *testing.T) {
	artifacts := []ArtifactInfo{
		{ID: "a", Platform: "linux-amd64"},
		{ID: "b", Platform: "linux-amd64"}, // Duplicate
		{ID: "c", Platform: "darwin-arm64"},
		{ID: "d", Platform: ""},            // No platform
	}

	platforms := extractPlatformsFromArtifacts(artifacts)

	assert.Len(t, platforms, 2)
}
