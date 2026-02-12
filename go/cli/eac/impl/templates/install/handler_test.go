//go:build L1 && ov
// +build L1,ov

// Feature: commands_templates
// Unit tests for the shared install handler configuration logic
package install

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugOnlyConfig(t *testing.T) {
	destinationFn := func(workspaceRoot string) string {
		return filepath.Join(workspaceRoot, "output", "templates")
	}

	tests := []struct {
		name          string
		workspaceRoot string
		args          []string
		wantDest      string
		wantDebug     bool
	}{
		{
			name:          "no args produces default config",
			workspaceRoot: "/repo",
			args:          []string{},
			wantDest:      filepath.Join("/repo", "output", "templates"),
			wantDebug:     false,
		},
		{
			name:          "debug long flag",
			workspaceRoot: "/repo",
			args:          []string{"--debug"},
			wantDest:      filepath.Join("/repo", "output", "templates"),
			wantDebug:     true,
		},
		{
			name:          "debug short flag",
			workspaceRoot: "/repo",
			args:          []string{"-d"},
			wantDest:      filepath.Join("/repo", "output", "templates"),
			wantDebug:     true,
		},
		{
			name:          "debug flag among other args",
			workspaceRoot: "/repo",
			args:          []string{"--something", "--debug", "--other"},
			wantDest:      filepath.Join("/repo", "output", "templates"),
			wantDebug:     true,
		},
		{
			name:          "no debug flag with other args",
			workspaceRoot: "/repo",
			args:          []string{"--something", "--other"},
			wantDest:      filepath.Join("/repo", "output", "templates"),
			wantDebug:     false,
		},
		{
			name:          "workspace root passed to destination function",
			workspaceRoot: "/custom/workspace",
			args:          []string{},
			wantDest:      filepath.Join("/custom/workspace", "output", "templates"),
			wantDebug:     false,
		},
		{
			name:          "nil args treated as empty",
			workspaceRoot: "/repo",
			args:          nil,
			wantDest:      filepath.Join("/repo", "output", "templates"),
			wantDebug:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := DebugOnlyConfig(destinationFn)
			cfg, err := parser(tt.workspaceRoot, tt.args)

			require.NoError(t, err)
			assert.Equal(t, tt.wantDest, cfg.Destination)
			assert.Equal(t, tt.workspaceRoot, cfg.WorkspaceRoot)
			assert.Equal(t, tt.wantDebug, cfg.Debug)
		})
	}
}

func TestDebugOnlyConfig_DestinationFunction(t *testing.T) {
	// Verify that different destination functions produce different paths
	tests := []struct {
		name          string
		destinationFn func(string) string
		workspaceRoot string
		wantDest      string
	}{
		{
			name: "simple subdirectory",
			destinationFn: func(root string) string {
				return filepath.Join(root, "docs")
			},
			workspaceRoot: "/repo",
			wantDest:      filepath.Join("/repo", "docs"),
		},
		{
			name: "nested path",
			destinationFn: func(root string) string {
				return filepath.Join(root, ".eac", "templates", "ai")
			},
			workspaceRoot: "/repo",
			wantDest:      filepath.Join("/repo", ".eac", "templates", "ai"),
		},
		{
			name: "single directory",
			destinationFn: func(root string) string {
				return filepath.Join(root, ".claude")
			},
			workspaceRoot: "/repo",
			wantDest:      filepath.Join("/repo", ".claude"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := DebugOnlyConfig(tt.destinationFn)
			cfg, err := parser(tt.workspaceRoot, nil)

			require.NoError(t, err)
			assert.Equal(t, tt.wantDest, cfg.Destination)
		})
	}
}

func TestConfig_Fields(t *testing.T) {
	// Verify Config struct holds all expected fields correctly
	cfg := &Config{
		Destination:   "/output/path",
		WorkspaceRoot: "/workspace",
		Debug:         true,
	}

	assert.Equal(t, "/output/path", cfg.Destination)
	assert.Equal(t, "/workspace", cfg.WorkspaceRoot)
	assert.True(t, cfg.Debug)
}

func TestParams_Fields(t *testing.T) {
	// Verify Params struct holds all expected fields correctly
	resolverCalled := false
	parserCalled := false

	p := Params{
		TemplateName: "Test",
		TemplateDir: func(root string) string {
			resolverCalled = true
			return filepath.Join(root, "templates", "test")
		},
		ParseConfig: func(workspaceRoot string, args []string) (*Config, error) {
			parserCalled = true
			return &Config{}, nil
		},
	}

	assert.Equal(t, "Test", p.TemplateName)

	// Verify TemplateDir resolver works
	result := p.TemplateDir("/repo")
	assert.True(t, resolverCalled)
	assert.Equal(t, filepath.Join("/repo", "templates", "test"), result)

	// Verify ParseConfig parser works
	_, err := p.ParseConfig("/repo", nil)
	require.NoError(t, err)
	assert.True(t, parserCalled)
}

func TestWriteDebugFile_SkipsWhenDebugDisabled(t *testing.T) {
	// writeDebugFile should be a no-op when Debug is false
	// This tests the guard clause without needing real filesystem writes
	cfg := &Config{
		Debug:         false,
		WorkspaceRoot: "/nonexistent",
		Destination:   "/nonexistent",
	}

	// Should not panic or error - just returns silently
	writeDebugFile(cfg, "test.txt", "content")
}
