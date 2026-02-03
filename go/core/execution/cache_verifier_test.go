package execution_test

import (
	"context"
	"testing"
	"time"

	"github.com/ready-to-release/eac/go/core/execution"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
)

func TestCacheVerifierFunc(t *testing.T) {
	// Test that CacheVerifierFunc satisfies the CacheVerifier interface
	cacheTime := time.Now()

	var verifier execution.CacheVerifier = execution.CacheVerifierFunc(
		func(ctx context.Context, unit workunit.UnitSpec) (execution.CacheResult, error) {
			return execution.CacheResult{
				Cached:    true,
				CacheTime: cacheTime,
			}, nil
		},
	)

	result, err := verifier.Verify(context.Background(), workunit.UnitSpec{
		ID: workunit.UnitID{Module: "mod1", Component: "comp1"},
	})

	assert.NoError(t, err)
	assert.True(t, result.Cached)
	assert.Equal(t, cacheTime, result.CacheTime)
}

func TestCacheResult_ZeroValue(t *testing.T) {
	// Test zero value represents "not cached"
	var result execution.CacheResult
	assert.False(t, result.Cached)
	assert.True(t, result.CacheTime.IsZero())
}
