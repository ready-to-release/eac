package cmdframework

import (
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/stretchr/testify/assert"
)

func TestAsyncDepsCheck_CollectsResult(t *testing.T) {
	// Set up a verifier that simulates Docker check latency
	expectedStatus := &initsummary.DepsStatus{
		Verified: true,
		Required: []string{"docker"},
		Available: []initsummary.DepsResult{
			{Name: "docker", Available: true, Version: "24.0.0"},
		},
	}

	original := depsVerifier
	defer func() { depsVerifier = original }()

	depsVerifier = func(ctx *ExecutionContext) *initsummary.DepsStatus {
		time.Sleep(10 * time.Millisecond) // Simulate some latency
		return expectedStatus
	}

	ctx := &ExecutionContext{
		Config: &CommandConfig{SkipDeps: false},
	}

	check := startAsyncDepsCheck(ctx)
	assert.NotNil(t, check, "async check should be started")

	result := check.wait()
	assert.Equal(t, expectedStatus, result)
}

func TestAsyncDepsCheck_SkippedWhenFlagSet(t *testing.T) {
	ctx := &ExecutionContext{
		Config: &CommandConfig{SkipDeps: true},
	}

	check := startAsyncDepsCheck(ctx)
	assert.Nil(t, check, "async check should not start when SkipDeps is true")
}

func TestAsyncDepsCheck_NilWhenNoVerifier(t *testing.T) {
	original := depsVerifier
	defer func() { depsVerifier = original }()

	depsVerifier = nil

	ctx := &ExecutionContext{
		Config: &CommandConfig{SkipDeps: false},
	}

	check := startAsyncDepsCheck(ctx)
	assert.Nil(t, check, "async check should not start when no verifier is registered")
}
