//go:build L0 && ov

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituteParams(t *testing.T) {
	params := map[string]string{
		"moniker": "my-app",
		"owner":   "myorg",
		"go_root": "go/my-app",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"containers/{moniker}", "containers/my-app"},
		{"{owner}/{moniker}", "myorg/my-app"},
		{"{go_root}/internal", "go/my-app/internal"},
		{"no-placeholders", "no-placeholders"},
		{"", ""},
		{"{unknown}", "{unknown}"}, // Unknown placeholders left as-is
	}

	for _, tc := range tests {
		result := substituteParams(tc.input, params)
		assert.Equal(t, tc.expected, result, "input: %s", tc.input)
	}
}

func TestBuildModuleParams(t *testing.T) {
	mod := &Module{
		Moniker: "my-module",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/my-module"},
		},
	}

	discoveryVars := map[string]string{
		"owner":   "myorg",
		"go_root": "go/my-module",
	}
	params := buildModuleParams(mod, discoveryVars)

	assert.Equal(t, "my-module", params["moniker"])
	assert.Equal(t, "myorg", params["owner"])
	assert.Equal(t, "go/my-module", params["go_root"]) // From discovery vars
}

func TestSubstituteModuleParams_DockerBuild(t *testing.T) {
	mod := &Module{
		Moniker: "my-app",
		Components: ModuleComponents{
			"container": &ComponentEntry{
				Root: "containers/{moniker}",
				DockerBuild: &DockerBuildConfig{
					Container:  "{moniker}",
					Context:    "containers/{moniker}",
					Dockerfile: "containers/{moniker}/Dockerfile",
					Tags: []string{
						"ghcr.io/{owner}/{moniker}:latest",
						"ghcr.io/{owner}/{moniker}:sha-{short_sha}",
					},
				},
			},
		},
	}

	vars := map[string]string{
		"owner": "ready-to-release",
	}
	SubstituteModuleParams(mod, vars)

	comp := mod.Components["container"]
	require.NotNil(t, comp)
	assert.Equal(t, "containers/my-app", comp.Root)

	dbc := comp.DockerBuild
	require.NotNil(t, dbc)
	assert.Equal(t, "my-app", dbc.Container)
	assert.Equal(t, "containers/my-app", dbc.Context)
	assert.Equal(t, "containers/my-app/Dockerfile", dbc.Dockerfile)
	assert.Contains(t, dbc.Tags, "ghcr.io/ready-to-release/my-app:latest")
	// short_sha is not provided, so placeholder remains
	assert.Contains(t, dbc.Tags, "ghcr.io/ready-to-release/my-app:sha-{short_sha}")
}

func TestSubstituteModuleParams_Versioning(t *testing.T) {
	mod := &Module{
		Moniker: "my-tool",
		Versioning: &ModuleVersioning{
			Changelog: "release/{moniker}/CHANGELOG.md",
		},
		Components: ModuleComponents{},
	}

	SubstituteModuleParams(mod, nil)

	assert.Equal(t, "release/my-tool/CHANGELOG.md", mod.Versioning.Changelog)
}

func TestSubstituteModuleParams_NoTemplate(t *testing.T) {
	// Module without placeholders should pass through unchanged
	mod := &Module{
		Moniker: "explicit-mod",
		Components: ModuleComponents{
			"go": &ComponentEntry{Root: "go/explicit"},
		},
	}

	SubstituteModuleParams(mod, nil)

	assert.Equal(t, "go/explicit", mod.Components["go"].Root)
}
