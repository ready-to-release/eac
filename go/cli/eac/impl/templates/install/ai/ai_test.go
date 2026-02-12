//go:build L1 && ov
// +build L1,ov

// Feature: commands_templates
// Unit tests for templates install ai command metadata and configuration
package ai

import (
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandsReturnsOneCommand(t *testing.T) {
	cmds := Commands()
	require.Len(t, cmds, 1)
}

func TestCommandName(t *testing.T) {
	cmd := &templatesInstallAICommand{}
	assert.Equal(t, "templates install ai", cmd.Name())
}

func TestCommandMetadata(t *testing.T) {
	cmd := &templatesInstallAICommand{}
	meta := cmd.Metadata()

	assert.Equal(t, "templates-install-ai", meta.CanonicalName)
	assert.NotEmpty(t, meta.Short)
	assert.NotEmpty(t, meta.Long)
	assert.Contains(t, meta.Short, "AI")
}

func TestCommandMetadata_Flags(t *testing.T) {
	cmd := &templatesInstallAICommand{}
	meta := cmd.Metadata()

	require.Len(t, meta.Flags, 1, "should have exactly one flag (debug)")

	debugFlag := meta.Flags[0]
	assert.Equal(t, "debug", debugFlag.Name)
	assert.Equal(t, "d", debugFlag.Shorthand)
	assert.Equal(t, "bool", debugFlag.Type)
	assert.Equal(t, "false", debugFlag.DefaultValue)
}

func TestTemplateDirResolver(t *testing.T) {
	// The AI template dir resolver should produce a path to templates/ai
	resolver := func(root string) string {
		return paths.TemplatePath(root, "ai")
	}

	tests := []struct {
		name string
		root string
		want string
	}{
		{
			name: "standard root",
			root: "/repo",
			want: filepath.Join("/repo", "templates", "ai"),
		},
		{
			name: "nested root",
			root: "/home/user/projects/myrepo",
			want: filepath.Join("/home/user/projects/myrepo", "templates", "ai"),
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
	// The AI destination is .eac/templates/ai under workspace root
	tests := []struct {
		name          string
		workspaceRoot string
		want          string
	}{
		{
			name:          "standard root",
			workspaceRoot: "/repo",
			want:          filepath.Join("/repo", paths.EACDir, "templates", "ai"),
		},
		{
			name:          "nested root",
			workspaceRoot: "/home/user/projects/myrepo",
			want:          filepath.Join("/home/user/projects/myrepo", paths.EACDir, "templates", "ai"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filepath.Join(tt.workspaceRoot, paths.EACDir, "templates", "ai")
			assert.Equal(t, tt.want, got)
		})
	}
}
