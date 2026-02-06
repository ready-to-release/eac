package tool

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExecutionContext_StreamingFields verifies streaming fields are available.
func TestExecutionContext_StreamingFields(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	stdinBuf := bytes.NewBufferString("test input")

	ctx := &ExecutionContext{
		WorkspaceRoot: "/workspace",
		ModuleRoot:    "/workspace/module",
		StdoutWriter:  &stdoutBuf,
		StderrWriter:  &stderrBuf,
		StdinReader:   stdinBuf,
		LogWriter:     nil,
	}

	// Verify fields are accessible
	assert.NotNil(t, ctx.StdoutWriter)
	assert.NotNil(t, ctx.StderrWriter)
	assert.NotNil(t, ctx.StdinReader)
	assert.Nil(t, ctx.LogWriter)
}

// TestExecutionContext_StreamingPrecedence documents expected behavior.
func TestExecutionContext_StreamingPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		stdoutW     bool
		stderrW     bool
		logW        bool
		stdinR      bool
		description string
	}{
		{
			name:        "default capture mode",
			stdoutW:     false,
			stderrW:     false,
			logW:        false,
			stdinR:      false,
			description: "output captured to ExecutionResult",
		},
		{
			name:        "log writer only",
			stdoutW:     false,
			stderrW:     false,
			logW:        true,
			stdinR:      false,
			description: "output to LogWriter + captured",
		},
		{
			name:        "separate streaming",
			stdoutW:     true,
			stderrW:     true,
			logW:        false,
			stdinR:      false,
			description: "stdout/stderr to separate writers",
		},
		{
			name:        "full streaming with stdin",
			stdoutW:     true,
			stderrW:     true,
			logW:        true,
			stdinR:      true,
			description: "complete IO streaming setup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx ExecutionContext

			if tt.stdoutW {
				ctx.StdoutWriter = &bytes.Buffer{}
			}
			if tt.stderrW {
				ctx.StderrWriter = &bytes.Buffer{}
			}
			if tt.logW {
				ctx.LogWriter = &bytes.Buffer{}
			}
			if tt.stdinR {
				ctx.StdinReader = bytes.NewBufferString("input")
			}

			assert.Equal(t, tt.stdoutW, ctx.StdoutWriter != nil)
			assert.Equal(t, tt.stderrW, ctx.StderrWriter != nil)
			assert.Equal(t, tt.logW, ctx.LogWriter != nil)
			assert.Equal(t, tt.stdinR, ctx.StdinReader != nil)
		})
	}
}
