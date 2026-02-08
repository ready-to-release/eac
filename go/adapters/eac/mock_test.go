package eac

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockEAC_RecordsCalls(t *testing.T) {
	mock := NewMock()

	cfg1 := &ExecConfig{WorkspaceRoot: "/ws1"}
	cfg2 := &ExecConfig{WorkspaceRoot: "/ws2"}

	_, _ = mock.Execute(context.Background(), []string{"build", "core"}, cfg1)
	_, _ = mock.Execute(context.Background(), []string{"test", "core"}, cfg2)
	_, _ = mock.Execute(context.Background(), []string{"show", "modules"}, nil)

	assert.Len(t, mock.Calls, 3)
	assert.Equal(t, []string{"build", "core"}, mock.Calls[0].Args)
	assert.Equal(t, cfg1, mock.Calls[0].Config)
	assert.Equal(t, []string{"test", "core"}, mock.Calls[1].Args)
	assert.Nil(t, mock.Calls[2].Config)
}

func TestMockEAC_RoutesResults(t *testing.T) {
	mock := NewMock()
	mock.Results["build"] = &Result{ExitCode: 0, Stdout: []byte("build ok")}
	mock.Results["test"] = &Result{ExitCode: 1, Stderr: []byte("test failed")}

	r1, err := mock.Execute(context.Background(), []string{"build", "core"}, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, r1.ExitCode)
	assert.Equal(t, []byte("build ok"), r1.Stdout)

	r2, err := mock.Execute(context.Background(), []string{"test", "core"}, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, r2.ExitCode)
	assert.Equal(t, []byte("test failed"), r2.Stderr)
}

func TestMockEAC_DefaultResult(t *testing.T) {
	mock := NewMock()

	r, err := mock.Execute(context.Background(), []string{"unknown", "cmd"}, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, r.ExitCode)
}

func TestMockEAC_NilConfig(t *testing.T) {
	mock := NewMock()

	r, err := mock.Execute(context.Background(), []string{"build"}, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, r.ExitCode)
}

func TestMockEAC_EmptyArgs(t *testing.T) {
	mock := NewMock()

	r, err := mock.Execute(context.Background(), []string{}, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, r.ExitCode)
}

func TestMockEAC_StdoutCapture(t *testing.T) {
	mock := NewMock()
	mock.Results["show"] = &Result{ExitCode: 0, Stdout: []byte("module-list")}

	r, err := mock.Execute(context.Background(), []string{"show", "modules"}, nil)
	require.NoError(t, err)
	assert.Equal(t, []byte("module-list"), r.Stdout)
}
