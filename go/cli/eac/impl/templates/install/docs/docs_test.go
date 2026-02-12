//go:build L1 && ov
// +build L1,ov

// Feature: commands_templates
// Unit tests for templates install docs command metadata and configuration
package docs

import (
	"path/filepath"
	"runtime"
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
	cmd := &templatesInstallDocsCommand{}
	assert.Equal(t, "templates install docs", cmd.Name())
}

func TestCommandMetadata(t *testing.T) {
	cmd := &templatesInstallDocsCommand{}
	meta := cmd.Metadata()

	assert.Equal(t, "templates-install-docs", meta.CanonicalName)
	assert.NotEmpty(t, meta.Short)
	assert.NotEmpty(t, meta.Long)
	assert.Contains(t, meta.Short, "documentation")
}

func TestCommandMetadata_Flags(t *testing.T) {
	cmd := &templatesInstallDocsCommand{}
	meta := cmd.Metadata()

	require.Len(t, meta.Flags, 2, "should have two flags (destination and debug)")

	// Find flags by name for order-independent assertions
	flagMap := make(map[string]int)
	for i, f := range meta.Flags {
		flagMap[f.Name] = i
	}

	destIdx, hasDest := flagMap["destination"]
	assert.True(t, hasDest, "should have a destination flag")
	if hasDest {
		assert.Equal(t, "string", meta.Flags[destIdx].Type)
	}

	debugIdx, hasDebug := flagMap["debug"]
	assert.True(t, hasDebug, "should have a debug flag")
	if hasDebug {
		assert.Equal(t, "d", meta.Flags[debugIdx].Shorthand)
		assert.Equal(t, "bool", meta.Flags[debugIdx].Type)
		assert.Equal(t, "false", meta.Flags[debugIdx].DefaultValue)
	}
}

func TestParseDocsConfig_DefaultDestination(t *testing.T) {
	cfg, err := parseDocsConfig("/repo", []string{})

	require.NoError(t, err)
	assert.Equal(t, "/repo", cfg.WorkspaceRoot)
	assert.False(t, cfg.Debug)
	// Default destination should be <workspace>/docs/reference
	expected := filepath.Join("/repo", paths.DocsDir, "reference")
	assert.Equal(t, expected, cfg.Destination)
}

func TestParseDocsConfig_CustomDestination(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		args     []string
		wantDest string
	}{
		{
			name:     "relative destination path",
			root:     "/repo",
			args:     []string{"--destination", "custom-docs"},
			wantDest: filepath.Join("/repo", "custom-docs"),
		},
		{
			name:     "nested relative destination",
			root:     "/repo",
			args:     []string{"--destination", "docs/api"},
			wantDest: filepath.Join("/repo", "docs", "api"),
		},
	}

	// Absolute path behavior differs by OS: on Windows only C:\-style paths are absolute,
	// while on Unix /path is absolute. Add platform-specific absolute path test case.
	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name     string
			root     string
			args     []string
			wantDest string
		}{
			name:     "absolute destination path",
			root:     "C:\\repo",
			args:     []string{"--destination", "C:\\absolute\\docs"},
			wantDest: "C:\\absolute\\docs",
		})
	} else {
		tests = append(tests, struct {
			name     string
			root     string
			args     []string
			wantDest string
		}{
			name:     "absolute destination path",
			root:     "/repo",
			args:     []string{"--destination", "/absolute/docs"},
			wantDest: "/absolute/docs",
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseDocsConfig(tt.root, tt.args)

			require.NoError(t, err)
			assert.Equal(t, tt.wantDest, cfg.Destination)
		})
	}
}

func TestParseDocsConfig_DebugFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantDebug bool
	}{
		{
			name:      "no debug flag",
			args:      []string{},
			wantDebug: false,
		},
		{
			name:      "long debug flag",
			args:      []string{"--debug"},
			wantDebug: true,
		},
		{
			name:      "short debug flag",
			args:      []string{"-d"},
			wantDebug: true,
		},
		{
			name:      "debug with destination",
			args:      []string{"--destination", "out", "--debug"},
			wantDebug: true,
		},
		{
			name:      "destination only",
			args:      []string{"--destination", "out"},
			wantDebug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseDocsConfig("/repo", tt.args)

			require.NoError(t, err)
			assert.Equal(t, tt.wantDebug, cfg.Debug)
		})
	}
}

func TestParseDocsConfig_DestinationAtEndOfArgs(t *testing.T) {
	// When --destination is the last arg with no value following,
	// the default destination should be used
	cfg, err := parseDocsConfig("/repo", []string{"--destination"})
	require.NoError(t, err)

	// With no value after --destination, the default should remain
	expected := filepath.Join("/repo", paths.DocsDir, "reference")
	assert.Equal(t, expected, cfg.Destination)
}

func TestParseDocsConfig_CombinedFlags(t *testing.T) {
	cfg, err := parseDocsConfig("/workspace", []string{"-d", "--destination", "api-docs"})

	require.NoError(t, err)
	assert.True(t, cfg.Debug)
	assert.Equal(t, filepath.Join("/workspace", "api-docs"), cfg.Destination)
	assert.Equal(t, "/workspace", cfg.WorkspaceRoot)
}

func TestTemplateDirResolver(t *testing.T) {
	resolver := func(root string) string {
		return paths.TemplatePath(root, "docs")
	}

	tests := []struct {
		name string
		root string
		want string
	}{
		{
			name: "standard root",
			root: "/repo",
			want: filepath.Join("/repo", "templates", "docs"),
		},
		{
			name: "nested root",
			root: "/home/user/projects/myrepo",
			want: filepath.Join("/home/user/projects/myrepo", "templates", "docs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolver(tt.root)
			assert.Equal(t, tt.want, got)
		})
	}
}
