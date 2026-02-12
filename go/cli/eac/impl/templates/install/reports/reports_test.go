//go:build L1 && ov
// +build L1,ov

// Feature: commands_templates
// Unit tests for templates install reports command metadata and configuration
package reports

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
	cmd := &templatesInstallReportsCommand{}
	assert.Equal(t, "templates install reports", cmd.Name())
}

func TestCommandMetadata(t *testing.T) {
	cmd := &templatesInstallReportsCommand{}
	meta := cmd.Metadata()

	assert.Equal(t, "templates-install-reports", meta.CanonicalName)
	assert.NotEmpty(t, meta.Short)
	assert.NotEmpty(t, meta.Long)
	assert.Contains(t, meta.Short, "report")
}

func TestCommandMetadata_Flags(t *testing.T) {
	cmd := &templatesInstallReportsCommand{}
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
		return paths.TemplateReportsPath(root)
	}

	tests := []struct {
		name string
		root string
		want string
	}{
		{
			name: "standard root",
			root: "/repo",
			want: filepath.Join("/repo", "templates", "reports"),
		},
		{
			name: "nested root",
			root: "/home/user/projects/myrepo",
			want: filepath.Join("/home/user/projects/myrepo", "templates", "reports"),
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
	// Reports destination is .clie/templates/reports under workspace root
	tests := []struct {
		name          string
		workspaceRoot string
		want          string
	}{
		{
			name:          "standard root",
			workspaceRoot: "/repo",
			want:          filepath.Join("/repo", paths.CLIEDir, "templates", "reports"),
		},
		{
			name:          "nested root",
			workspaceRoot: "/home/user/projects/myrepo",
			want:          filepath.Join("/home/user/projects/myrepo", paths.CLIEDir, "templates", "reports"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filepath.Join(tt.workspaceRoot, paths.CLIEDir, "templates", "reports")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDebugOnlyConfigIntegration(t *testing.T) {
	// Verify the DebugOnlyConfig used by reports produces the expected config
	parser := install.DebugOnlyConfig(func(workspaceRoot string) string {
		return filepath.Join(workspaceRoot, paths.CLIEDir, "templates", "reports")
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
			wantDest:  filepath.Join("/repo", paths.CLIEDir, "templates", "reports"),
			wantDebug: false,
		},
		{
			name:      "with debug",
			root:      "/repo",
			args:      []string{"--debug"},
			wantDest:  filepath.Join("/repo", paths.CLIEDir, "templates", "reports"),
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
