//go:build L1 && ov
// +build L1,ov

// Feature: commands_templates
// Unit tests for templates install specs command metadata and configuration
package specs

import (
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/cli/eac/impl/templates/install"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandsReturnsOneCommand(t *testing.T) {
	cmds := Commands()
	require.Len(t, cmds, 1)
}

func TestCommandName(t *testing.T) {
	cmd := &templatesInstallSpecsCommand{}
	assert.Equal(t, "templates install specs", cmd.Name())
}

func TestCommandMetadata(t *testing.T) {
	cmd := &templatesInstallSpecsCommand{}
	meta := cmd.Metadata()

	assert.Equal(t, "templates-install-specs", meta.CanonicalName)
	assert.NotEmpty(t, meta.Short)
	assert.NotEmpty(t, meta.Long)
	assert.Contains(t, meta.Short, "specification")
}

func TestCommandMetadata_Flags(t *testing.T) {
	cmd := &templatesInstallSpecsCommand{}
	meta := cmd.Metadata()

	require.Len(t, meta.Flags, 1, "should have exactly one flag (debug)")

	debugFlag := meta.Flags[0]
	assert.Equal(t, "debug", debugFlag.Name)
	assert.Equal(t, "d", debugFlag.Shorthand)
	assert.Equal(t, "bool", debugFlag.Type)
	assert.Equal(t, "false", debugFlag.DefaultValue)
}

func TestTemplateDirResolver(t *testing.T) {
	resolver := func(root string) string {
		return paths.TemplateSpecsPath(root)
	}

	tests := []struct {
		name string
		root string
		want string
	}{
		{
			name: "standard root",
			root: "/repo",
			want: filepath.Join("/repo", "templates", "specs"),
		},
		{
			name: "nested root",
			root: "/home/user/projects/myrepo",
			want: filepath.Join("/home/user/projects/myrepo", "templates", "specs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolver(tt.root)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDestinationPath(t *testing.T) {
	// Specs destination is specs/.risk-controls under workspace root
	tests := []struct {
		name          string
		workspaceRoot string
		want          string
	}{
		{
			name:          "standard root",
			workspaceRoot: "/repo",
			want:          filepath.Join("/repo", paths.SpecsDir, paths.RiskControlsDir),
		},
		{
			name:          "nested root",
			workspaceRoot: "/home/user/projects/myrepo",
			want:          filepath.Join("/home/user/projects/myrepo", paths.SpecsDir, paths.RiskControlsDir),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := paths.RiskControlsPath(tt.workspaceRoot)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDebugOnlyConfigIntegration(t *testing.T) {
	// Verify the DebugOnlyConfig used by specs produces the expected config
	parser := install.DebugOnlyConfig(func(workspaceRoot string) string {
		return paths.RiskControlsPath(workspaceRoot)
	})

	tests := []struct {
		name      string
		root      string
		args      []string
		wantDest  string
		wantDebug bool
	}{
		{
			name:      "default config",
			root:      "/repo",
			args:      []string{},
			wantDest:  filepath.Join("/repo", paths.SpecsDir, paths.RiskControlsDir),
			wantDebug: false,
		},
		{
			name:      "with debug long flag",
			root:      "/repo",
			args:      []string{"--debug"},
			wantDest:  filepath.Join("/repo", paths.SpecsDir, paths.RiskControlsDir),
			wantDebug: true,
		},
		{
			name:      "with debug short flag",
			root:      "/workspace",
			args:      []string{"-d"},
			wantDest:  filepath.Join("/workspace", paths.SpecsDir, paths.RiskControlsDir),
			wantDebug: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parser(tt.root, tt.args)

			require.NoError(t, err)
			assert.Equal(t, tt.wantDest, cfg.Destination)
			assert.Equal(t, tt.wantDebug, cfg.Debug)
			assert.Equal(t, tt.root, cfg.WorkspaceRoot)
		})
	}
}
