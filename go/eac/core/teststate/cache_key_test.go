package teststate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestCacheKey_String(t *testing.T) {
	tests := []struct {
		name     string
		key      TestCacheKey
		expected string
	}{
		{
			name: "unit test key",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "go",
				TestType:  "gotest",
				TestSet:   TestSetUnit,
			},
			expected: "eac-commands:go:gotest:unit",
		},
		{
			name: "integration test key",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "gherkin",
				TestType:  "godog",
				TestSet:   TestSetIntegration,
			},
			expected: "eac-commands:gherkin:godog:integration",
		},
		{
			name: "different module",
			key: TestCacheKey{
				Module:    "r2r-cli",
				Component: "go",
				TestType:  "gotest",
				TestSet:   TestSetUnit,
			},
			expected: "r2r-cli:go:gotest:unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.key.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParseTestCacheKey(t *testing.T) {
	t.Run("valid unit key", func(t *testing.T) {
		key, err := ParseTestCacheKey("eac-commands:go:gotest:unit")
		require.NoError(t, err)
		assert.Equal(t, "eac-commands", key.Module)
		assert.Equal(t, "go", key.Component)
		assert.Equal(t, "gotest", key.TestType)
		assert.Equal(t, TestSetUnit, key.TestSet)
	})

	t.Run("valid integration key", func(t *testing.T) {
		key, err := ParseTestCacheKey("eac-commands:gherkin:godog:integration")
		require.NoError(t, err)
		assert.Equal(t, "eac-commands", key.Module)
		assert.Equal(t, "gherkin", key.Component)
		assert.Equal(t, "godog", key.TestType)
		assert.Equal(t, TestSetIntegration, key.TestSet)
	})

	t.Run("invalid format - too few parts", func(t *testing.T) {
		_, err := ParseTestCacheKey("eac-commands:go:gotest")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cache key format")
	})

	t.Run("invalid format - too many parts", func(t *testing.T) {
		_, err := ParseTestCacheKey("eac-commands:go:gotest:unit:extra")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cache key format")
	})

	t.Run("invalid test set", func(t *testing.T) {
		_, err := ParseTestCacheKey("eac-commands:go:gotest:invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid test set")
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ParseTestCacheKey("")
		require.Error(t, err)
	})
}

func TestTestCacheKey_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		key      TestCacheKey
		expected bool
	}{
		{
			name: "valid unit key",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "go",
				TestType:  "gotest",
				TestSet:   TestSetUnit,
			},
			expected: true,
		},
		{
			name: "valid integration key",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "gherkin",
				TestType:  "godog",
				TestSet:   TestSetIntegration,
			},
			expected: true,
		},
		{
			name: "missing module",
			key: TestCacheKey{
				Module:    "",
				Component: "go",
				TestType:  "gotest",
				TestSet:   TestSetUnit,
			},
			expected: false,
		},
		{
			name: "missing component",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "",
				TestType:  "gotest",
				TestSet:   TestSetUnit,
			},
			expected: false,
		},
		{
			name: "missing test type",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "go",
				TestType:  "",
				TestSet:   TestSetUnit,
			},
			expected: false,
		},
		{
			name: "invalid test set",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "go",
				TestType:  "gotest",
				TestSet:   TestSet("invalid"),
			},
			expected: false,
		},
		{
			name: "empty test set",
			key: TestCacheKey{
				Module:    "eac-commands",
				Component: "go",
				TestType:  "gotest",
				TestSet:   TestSet(""),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.key.IsValid()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestTestCacheKey_RoundTrip(t *testing.T) {
	original := TestCacheKey{
		Module:    "eac-commands",
		Component: "go",
		TestType:  "gotest",
		TestSet:   TestSetUnit,
	}

	// Convert to string and back
	str := original.String()
	parsed, err := ParseTestCacheKey(str)

	require.NoError(t, err)
	assert.Equal(t, original, parsed)
}
