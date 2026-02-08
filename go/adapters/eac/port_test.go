package eac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecConfig_Defaults(t *testing.T) {
	cfg := ExecConfig{}
	assert.Empty(t, cfg.WorkspaceRoot)
	assert.Empty(t, cfg.ModuleRoot)
	assert.Empty(t, cfg.OutputDir)
	assert.Nil(t, cfg.StdoutWriter)
	assert.Nil(t, cfg.StderrWriter)
	assert.Nil(t, cfg.StdinReader)
	assert.Nil(t, cfg.Env)
	assert.Nil(t, cfg.FullEnv)
}

func TestResult_Success_Zero(t *testing.T) {
	r := &Result{ExitCode: 0}
	assert.True(t, r.Success())
}

func TestResult_Success_NonZero(t *testing.T) {
	r := &Result{ExitCode: 1}
	assert.False(t, r.Success())
}

func TestResult_Success_Negative(t *testing.T) {
	r := &Result{ExitCode: -1}
	assert.False(t, r.Success())
}
