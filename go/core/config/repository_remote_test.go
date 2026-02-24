package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestRemoteConfig_OmitemptySupressesEmptyFields verifies that an all-zero RemoteConfig
// does not write individual empty fields when marshalled (Bug 6 regression test).
// We marshal RemoteConfig directly to isolate the RemoteConfig tag behavior.
func TestRemoteConfig_OmitemptySupressesEmptyFields(t *testing.T) {
	remote := RemoteConfig{} // all zero values

	out, err := yaml.Marshal(remote)
	require.NoError(t, err)
	output := string(out)

	// With omitempty, all empty-string fields should be suppressed
	assert.False(t, strings.Contains(output, "type:"), "empty 'type' should be omitted")
	assert.False(t, strings.Contains(output, "owner:"), "empty 'owner' should be omitted")
	assert.False(t, strings.Contains(output, "repo:"), "empty 'repo' should be omitted")
	assert.False(t, strings.Contains(output, "url:"), "empty 'url' should be omitted")
	assert.False(t, strings.Contains(output, "pages_url:"), "empty 'pages_url' should be omitted")
	assert.False(t, strings.Contains(output, "registry_url:"), "empty 'registry_url' should be omitted")
}

// TestRemoteConfig_OmitemptyPreservesNonEmptyFields verifies that set fields are still marshalled.
func TestRemoteConfig_OmitemptyPreservesNonEmptyFields(t *testing.T) {
	remote := RemoteConfig{
		Owner:    "myorg",
		RepoName: "myrepo",
	}

	out, err := yaml.Marshal(remote)
	require.NoError(t, err)
	output := string(out)

	assert.True(t, strings.Contains(output, "owner: myorg"), "non-empty 'owner' should be present")
	assert.True(t, strings.Contains(output, "repo: myrepo"), "non-empty 'repo' should be present")
	// Empty fields should still be omitted
	assert.False(t, strings.Contains(output, "type:"), "empty 'type' should be omitted")
	assert.False(t, strings.Contains(output, "url:"), "empty 'url' should be omitted")
	assert.False(t, strings.Contains(output, "pages_url:"), "empty 'pages_url' should be omitted")
	assert.False(t, strings.Contains(output, "registry_url:"), "empty 'registry_url' should be omitted")
}

// TestLoadRepositoryConfigUnmerged_ReturnsRawUserContent verifies that system defaults
// are NOT injected by LoadRepositoryConfigUnmerged (Bug 2 regression test).
// The key difference vs config.Load(): system defaults (trunk_branch: main, ci: 8, etc.)
// are not applied — only the user's explicit values are returned.
func TestLoadRepositoryConfigUnmerged_ReturnsRawUserContent(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	require.NoError(t, os.MkdirAll(eacDir, 0o755))

	// Minimal user file — no trunk_branch, no parallelism, no conventions
	content := `repository:
  type: poly
modules:
  - moniker: myapp
    description: My application
    components:
      - type: python
        root: .
`
	require.NoError(t, os.WriteFile(filepath.Join(eacDir, "repository.yml"), []byte(content), 0o644))

	raw, err := LoadRepositoryConfigUnmerged(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, raw)

	// Only the explicit type should be set; system defaults must NOT be present.
	// config.Load() would return TrunkBranch="main", Parallelism.CI=8, Conventions.GodogTest="godog_test.go".
	// LoadRepositoryConfigUnmerged must return zero values for these instead.
	assert.Equal(t, "poly", raw.Repository.Type)
	assert.Empty(t, raw.Repository.TrunkBranch, "trunk_branch should be empty (not filled by defaults)")
	assert.Equal(t, 0, raw.Repository.Parallelism.CI, "parallelism.ci should be zero (not filled by defaults)")
	assert.Empty(t, raw.Conventions.GodogTest, "godog_test convention should be empty (not filled by defaults)")

	require.Len(t, raw.Modules, 1)
	assert.Equal(t, "myapp", raw.Modules[0].Moniker)
	assert.Nil(t, raw.Modules[0].DependsOn, "DependsOn should be nil (not filled by applyModuleDefaults)")
}

// TestLoadRepositoryConfigUnmerged_MarshalDoesNotLeakDefaultValues verifies that
// system-injected default VALUES are not written back to YAML after a roundtrip.
// This is distinct from zero-value fields: zero-value fields may still appear (e.g. trunk_branch: "")
// but system-specific values like "main", "squash", "unrestricted" should not be written
// unless the user explicitly set them.
func TestLoadRepositoryConfigUnmerged_MarshalDoesNotLeakDefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".eac")
	require.NoError(t, os.MkdirAll(eacDir, 0o755))

	// Minimal user file — no system-default values set
	content := `repository:
  type: poly
modules:
  - moniker: myapp
    description: My app
    components:
      - type: python
        root: .
`
	require.NoError(t, os.WriteFile(filepath.Join(eacDir, "repository.yml"), []byte(content), 0o644))

	raw, err := LoadRepositoryConfigUnmerged(tmpDir)
	require.NoError(t, err)

	out, err := yaml.Marshal(raw)
	require.NoError(t, err)
	output := string(out)

	// System-specific default VALUES must not appear (they'd only appear if config.Load() injected them).
	assert.False(t, strings.Contains(output, "trunk_branch: main"), "default value 'main' should not be written")
	assert.False(t, strings.Contains(output, "merge_strategy: squash"), "default value 'squash' should not be written")
	assert.False(t, strings.Contains(output, "constraint: unrestricted"), "default versioning constraint should not be written")
	assert.False(t, strings.Contains(output, "specs_root: specs"), "default specs_root value should not be written")
	assert.False(t, strings.Contains(output, "godog_test: godog_test.go"), "default godog_test convention should not be written")
	// Remote config: with omitempty on RemoteConfig fields, empty remote sub-fields should be absent
	assert.False(t, strings.Contains(output, "owner: \"\""), "empty owner field should be omitted by omitempty")
}
