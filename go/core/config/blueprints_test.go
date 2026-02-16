//go:build L0 && ov

package config

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeBlueprintsConfig_BothNil(t *testing.T) {
	result := MergeBlueprintsConfig(nil, nil)
	require.NotNil(t, result)
	assert.Empty(t, result.ArtifactMatrices)
}

func TestMergeBlueprintsConfig_BaseNil(t *testing.T) {
	override := &BlueprintsConfig{
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "linux-amd64"}}},
		},
	}
	result := MergeBlueprintsConfig(nil, override)
	assert.Equal(t, override, result)
}

func TestMergeBlueprintsConfig_OverrideNil(t *testing.T) {
	base := &BlueprintsConfig{
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "linux-amd64"}}},
		},
	}
	result := MergeBlueprintsConfig(base, nil)
	assert.Equal(t, base, result)
}

func TestMergeBlueprintsConfig_OverrideKeys(t *testing.T) {
	base := &BlueprintsConfig{
		ComponentKinds: map[string]*ComponentType{
			"go": {},
		},
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "linux-amd64"}}},
		},
	}

	override := &BlueprintsConfig{
		ComponentKinds: map[string]*ComponentType{
			"container": {},
		},
		ArtifactMatrices: map[string]*ArtifactMatrix{
			"cross-platform": {Entries: []ArtifactMatrixEntry{{ID: "override"}}},
		},
	}

	result := MergeBlueprintsConfig(base, override)

	// Base component kind preserved
	assert.Contains(t, result.ComponentKinds, "go")
	// New component kind added
	assert.Contains(t, result.ComponentKinds, "container")
	// Overridden matrix
	assert.Equal(t, "override", result.ArtifactMatrices["cross-platform"].Entries[0].ID)
}

func TestMergeBlueprintsConfig_EmptyMaps(t *testing.T) {
	base := &BlueprintsConfig{
		ComponentKinds: map[string]*ComponentType{
			"go": {},
		},
	}
	override := &BlueprintsConfig{} // All maps nil

	result := MergeBlueprintsConfig(base, override)

	// Base should be preserved
	assert.Contains(t, result.ComponentKinds, "go")
}

// TestLoadBlueprintsDefaults_ArtifactMatrices verifies that artifact matrices are loaded.
func TestLoadBlueprintsDefaults_ArtifactMatrices(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadBlueprintsDefaults(repoRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify artifact matrices are loaded
	require.NotNil(t, cfg.ArtifactMatrices)
	assert.Contains(t, cfg.ArtifactMatrices, "cross-platform")
	assert.Contains(t, cfg.ArtifactMatrices, "cross-platform-upx")
	assert.Contains(t, cfg.ArtifactMatrices, "single-platform")

	// Verify cross-platform matrix has expected entries
	cp := cfg.ArtifactMatrices["cross-platform"]
	require.NotNil(t, cp)
	assert.Len(t, cp.Entries, 5)
}

// TestLoadBlueprintsDefaults_ComponentKinds verifies component kinds are loaded.
func TestLoadBlueprintsDefaults_ComponentKinds(t *testing.T) {
	repoRoot, err := workspace.Root()
	require.NoError(t, err)

	cfg, err := LoadBlueprintsDefaults(repoRoot)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify expected component kinds exist
	require.NotNil(t, cfg.ComponentKinds)
	assert.Contains(t, cfg.ComponentKinds, "go")
	assert.Contains(t, cfg.ComponentKinds, "container")
	assert.Contains(t, cfg.ComponentKinds, "typescript")

	// Verify container kind has docker_build_defaults
	containerKind := cfg.ComponentKinds["container"]
	require.NotNil(t, containerKind)
	require.NotNil(t, containerKind.DockerBuildDefaults)
	assert.Equal(t, "{moniker}", containerKind.DockerBuildDefaults.Container)
	assert.Equal(t, "ghcr.io", containerKind.DockerBuildDefaults.Registry)

	// Verify godog kind has runner_search_dirs for BDD test runner discovery
	godogKind := cfg.ComponentKinds["godog"]
	require.NotNil(t, godogKind)
	assert.Equal(t, []string{".", "specs"}, godogKind.RunnerSearchDirs)
}
