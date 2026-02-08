package eac

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandExecutorAdapter_Run_Success(t *testing.T) {
	mock := NewMock()
	mock.Results["show"] = &Result{ExitCode: 0, Stdout: []byte("output")}

	adapter := &CommandExecutorAdapter{Port: mock}
	out, err := adapter.Run(context.Background(), "/workspace", []string{"show", "modules"})

	require.NoError(t, err)
	assert.Equal(t, "output", out)
}

func TestCommandExecutorAdapter_Run_Failure(t *testing.T) {
	mock := NewMock()
	mock.Results["build"] = &Result{ExitCode: 1, Stderr: []byte("compile error")}

	adapter := &CommandExecutorAdapter{Port: mock}
	_, err := adapter.Run(context.Background(), "/workspace", []string{"build", "core"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "compile error")
}

func TestCommandExecutorAdapter_Run_ExecError(t *testing.T) {
	mock := NewMock()
	mock.ErrorFunc = func(args []string) error {
		return fmt.Errorf("execution failed: binary not found")
	}

	adapter := &CommandExecutorAdapter{Port: mock}
	_, err := adapter.Run(context.Background(), "/workspace", []string{"build"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "binary not found")
}

func TestRunCommand_Success(t *testing.T) {
	mock := NewMock()
	mock.Results["get"] = &Result{ExitCode: 0, Stdout: []byte("data")}

	out, err := RunCommand(context.Background(), mock, "/workspace", []string{"get", "modules"})

	require.NoError(t, err)
	assert.Equal(t, "data", out)
}

func TestRunCommand_Failure(t *testing.T) {
	mock := NewMock()
	mock.Results["get"] = &Result{ExitCode: 2, Stderr: []byte("not found")}

	_, err := RunCommand(context.Background(), mock, "/workspace", []string{"get", "modules"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunCommand_WorkspacePassthrough(t *testing.T) {
	mock := NewMock()

	_, _ = RunCommand(context.Background(), mock, "/my/workspace", []string{"show"})

	require.Len(t, mock.Calls, 1)
	assert.Equal(t, "/my/workspace", mock.Calls[0].Config.WorkspaceRoot)
}
