package platform

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapCommand_ReturnsNameUnchanged(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		args     []string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "simple command no args",
			cmd:      "echo",
			args:     nil,
			wantCmd:  "echo",
			wantArgs: nil,
		},
		{
			name:     "command with single arg",
			cmd:      "go",
			args:     []string{"test"},
			wantCmd:  "go",
			wantArgs: []string{"test"},
		},
		{
			name:     "command with multiple args",
			cmd:      "docker",
			args:     []string{"run", "--rm", "-it", "alpine"},
			wantCmd:  "docker",
			wantArgs: []string{"run", "--rm", "-it", "alpine"},
		},
		{
			name:     "command with path",
			cmd:      "/usr/bin/env",
			args:     []string{"bash", "-c", "echo hello"},
			wantCmd:  "/usr/bin/env",
			wantArgs: []string{"bash", "-c", "echo hello"},
		},
		{
			name:     "empty args slice",
			cmd:      "ls",
			args:     []string{},
			wantCmd:  "ls",
			wantArgs: []string{},
		},
		{
			name:     "command with special characters in args",
			cmd:      "git",
			args:     []string{"commit", "-m", "fix: handle spaces & symbols"},
			wantCmd:  "git",
			wantArgs: []string{"commit", "-m", "fix: handle spaces & symbols"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := WrapCommand(tt.cmd, tt.args...)
			assert.Equal(t, tt.wantCmd, gotCmd)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

func TestWrapCommand_PassThrough(t *testing.T) {
	// Verify that WrapCommand is purely a pass-through on the current platform
	cmd, args := WrapCommand("npm", "install", "--save")
	assert.Equal(t, "npm", cmd)
	assert.Equal(t, []string{"install", "--save"}, args)
}

func TestLineEnding_PlatformSpecific(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		assert.Equal(t, "\r\n", LineEnding,
			"Windows LineEnding should be \\r\\n")
	default:
		assert.Equal(t, "\n", LineEnding,
			"Unix LineEnding should be \\n")
	}
}

func TestLineEnding_NotEmpty(t *testing.T) {
	assert.NotEmpty(t, LineEnding, "LineEnding should never be empty")
}

func TestLineEnding_EndsWithNewline(t *testing.T) {
	assert.Contains(t, LineEnding, "\n",
		"LineEnding should always contain a newline character")
}

func TestLineEnding_MaxLength(t *testing.T) {
	assert.LessOrEqual(t, len(LineEnding), 2,
		"LineEnding should be at most 2 characters (\\r\\n)")
}
