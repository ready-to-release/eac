package cache

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCICacheChecker_Cached_SHAMatchesHead(t *testing.T) {
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "abc123",
	})
	checker := NewCICacheChecker(querier, nil)

	result := checker.Check("core", "abc123")

	assert.True(t, result.Cached)
	assert.Equal(t, "core", result.Module)
	assert.Equal(t, "valid_ci_at_head", result.Reason)
	assert.Equal(t, "abc123", result.LastRunSHA)
}

func TestCICacheChecker_Uncached_NoRuns(t *testing.T) {
	querier := NewMockCIRunQuerier(nil)
	checker := NewCICacheChecker(querier, nil)

	result := checker.Check("core", "abc123")

	assert.False(t, result.Cached)
	assert.Equal(t, "no_ci_run", result.Reason)
	assert.Empty(t, result.LastRunSHA)
}

func TestCICacheChecker_Uncached_DifferentSHA(t *testing.T) {
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "different123",
	})
	checker := NewCICacheChecker(querier, nil)

	result := checker.Check("core", "abc123")

	assert.False(t, result.Cached)
	assert.Contains(t, result.Reason, "ci_at_different_sha:")
	assert.Equal(t, "different123", result.LastRunSHA)
}

func TestCICacheChecker_Uncached_QuerierError(t *testing.T) {
	querier := &MockCIRunQuerier{
		Runs: nil,
		Err:  fmt.Errorf("API rate limit exceeded"),
	}
	checker := NewCICacheChecker(querier, nil)

	result := checker.Check("core", "abc123")

	assert.False(t, result.Cached, "should fail open and dispatch")
	assert.Contains(t, result.Reason, "query_failed:")
	assert.Empty(t, result.LastRunSHA)
}

func TestCICacheChecker_Bypass_RemoteCISkip(t *testing.T) {
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "abc123", // Would be cached without skip
	})
	cfg := NewConfig()
	cfg.Skip(Spec{Level: LevelRemote, Type: TypeCI})
	checker := NewCICacheChecker(querier, cfg)

	result := checker.Check("core", "abc123")

	assert.False(t, result.Cached, "should bypass when remote:ci is skipped")
	assert.Equal(t, "cache_bypassed", result.Reason)
}

func TestCICacheChecker_Bypass_AllSkip(t *testing.T) {
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "abc123",
	})
	cfg := NewConfig()
	cfg.SkipAll()
	checker := NewCICacheChecker(querier, cfg)

	result := checker.Check("core", "abc123")

	assert.False(t, result.Cached, "should bypass when all is skipped")
	assert.Equal(t, "cache_bypassed", result.Reason)
}

func TestCICacheChecker_NilConfig_CachingEnabled(t *testing.T) {
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "abc123",
	})
	checker := NewCICacheChecker(querier, nil)

	result := checker.Check("core", "abc123")

	assert.True(t, result.Cached, "nil config should mean caching enabled")
}

func TestCICacheChecker_WorkflowNameMapping(t *testing.T) {
	// Verify module name -> workflow name mapping
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-eac-ext.yaml":       "sha1",
		"ci-clie-installer.yaml": "sha2",
		"ci-docs.yaml":          "sha3",
	})
	checker := NewCICacheChecker(querier, nil)

	tests := []struct {
		module  string
		headSHA string
		cached  bool
	}{
		{"eac-ext", "sha1", true},
		{"clie-installer", "sha2", true},
		{"docs", "sha3", true},
		{"docs", "other", false},
		{"nonexistent", "sha1", false},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			result := checker.Check(tt.module, tt.headSHA)
			assert.Equal(t, tt.cached, result.Cached)
			assert.Equal(t, tt.module, result.Module)
		})
	}
}

func TestCICacheChecker_ShortSHATruncation(t *testing.T) {
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "abcdef1234567890",
	})
	checker := NewCICacheChecker(querier, nil)

	result := checker.Check("core", "different")

	assert.Contains(t, result.Reason, "ci_at_different_sha:abcdef1")
}

func TestCICacheChecker_ShortSHA_ShortInput(t *testing.T) {
	// SHA shorter than 7 chars should not panic
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "abc",
	})
	checker := NewCICacheChecker(querier, nil)

	result := checker.Check("core", "different")

	assert.Contains(t, result.Reason, "ci_at_different_sha:abc")
}

func TestCICacheChecker_EmptyConfigNotSkipping(t *testing.T) {
	querier := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "abc123",
	})
	cfg := NewConfig() // Empty config - no skips
	checker := NewCICacheChecker(querier, cfg)

	result := checker.Check("core", "abc123")

	assert.True(t, result.Cached, "empty config should not bypass CI cache")
}

func TestMockCIRunQuerier_BasicOperation(t *testing.T) {
	mock := NewMockCIRunQuerier(map[string]string{
		"ci-core.yaml": "sha123",
	})

	sha, err := mock.LastSuccessfulRunSHA("ci-core.yaml")
	assert.NoError(t, err)
	assert.Equal(t, "sha123", sha)

	sha, err = mock.LastSuccessfulRunSHA("ci-unknown.yaml")
	assert.NoError(t, err)
	assert.Empty(t, sha)
}

func TestMockCIRunQuerier_Error(t *testing.T) {
	mock := &MockCIRunQuerier{Err: fmt.Errorf("network error")}

	_, err := mock.LastSuccessfulRunSHA("ci-core.yaml")
	assert.Error(t, err)
}

func TestNewMockCIRunQuerier_NilRuns(t *testing.T) {
	mock := NewMockCIRunQuerier(nil)
	assert.NotNil(t, mock.Runs)
}
