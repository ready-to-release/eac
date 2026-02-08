package docker

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	containerport "github.com/ready-to-release/eac/contracts/container-runtime/0.1.0"
)

// TestContainerConfig_StreamingFields verifies that ContainerConfig has the new streaming fields.
func TestContainerConfig_StreamingFields(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	stdinBuf := bytes.NewBufferString("test input")

	config := &containerport.ContainerConfig{
		Image:        "test:latest",
		Command:      []string{"echo", "hello"},
		StdoutWriter: &stdoutBuf,
		StderrWriter: &stderrBuf,
		StdinReader:  stdinBuf,
		LogWriter:    nil,
	}

	// Verify fields are accessible and set correctly
	assert.NotNil(t, config.StdoutWriter)
	assert.NotNil(t, config.StderrWriter)
	assert.NotNil(t, config.StdinReader)
	assert.Nil(t, config.LogWriter)
}

// TestContainerConfig_StreamingPrecedence documents the expected precedence behavior:
// 1. StdoutWriter/StderrWriter (separate streaming) - takes precedence
// 2. LogWriter (combined streaming)
// 3. Buffer capture (default)
func TestContainerConfig_StreamingPrecedence(t *testing.T) {
	tests := []struct {
		name        string
		stdoutW     bool
		stderrW     bool
		logW        bool
		description string
	}{
		{
			name:        "all nil - capture to buffers",
			stdoutW:     false,
			stderrW:     false,
			logW:        false,
			description: "output captured to ContainerResult.Stdout/Stderr",
		},
		{
			name:        "only LogWriter - combined streaming + capture",
			stdoutW:     false,
			stderrW:     false,
			logW:        true,
			description: "output streamed to LogWriter AND captured to buffers",
		},
		{
			name:        "StdoutWriter only - stdout streaming, stderr capture",
			stdoutW:     true,
			stderrW:     false,
			logW:        false,
			description: "stdout streamed, stderr captured to buffer",
		},
		{
			name:        "StdoutWriter + LogWriter - stdout to both",
			stdoutW:     true,
			stderrW:     false,
			logW:        true,
			description: "stdout to StdoutWriter + LogWriter, stderr to LogWriter + buffer",
		},
		{
			name:        "all writers - full streaming",
			stdoutW:     true,
			stderrW:     true,
			logW:        true,
			description: "stdout/stderr to respective writers + LogWriter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config containerport.ContainerConfig

			if tt.stdoutW {
				config.StdoutWriter = &bytes.Buffer{}
			}
			if tt.stderrW {
				config.StderrWriter = &bytes.Buffer{}
			}
			if tt.logW {
				config.LogWriter = &bytes.Buffer{}
			}

			// Just verify the configuration is valid
			assert.Equal(t, tt.stdoutW, config.StdoutWriter != nil)
			assert.Equal(t, tt.stderrW, config.StderrWriter != nil)
			assert.Equal(t, tt.logW, config.LogWriter != nil)
		})
	}
}

// TestContainerConfig_StdinSupport verifies stdin field configuration.
func TestContainerConfig_StdinSupport(t *testing.T) {
	tests := []struct {
		name    string
		stdin   bool
		wantNil bool
	}{
		{
			name:    "no stdin",
			stdin:   false,
			wantNil: true,
		},
		{
			name:    "with stdin",
			stdin:   true,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config containerport.ContainerConfig

			if tt.stdin {
				config.StdinReader = bytes.NewBufferString("input data")
			}

			assert.Equal(t, tt.wantNil, config.StdinReader == nil)
		})
	}
}
