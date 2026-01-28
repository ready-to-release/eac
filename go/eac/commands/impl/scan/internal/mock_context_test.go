//go:build L0 && ov

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockContext_SetAllMocks tests that SetAllMocks correctly sets all global mocks.
func TestMockContext_SetAllMocks(t *testing.T) {
	// Ensure clean state
	defer ResetAllMocks()

	trivyData := map[string]interface{}{"trivy": "mock"}
	semgrepData := map[string]interface{}{"semgrep": "mock"}
	zapData := map[string]interface{}{"zap": "mock"}

	ctx := &MockContext{
		TrivyOutput:   trivyData,
		SemgrepOutput: semgrepData,
		ZAPOutput:     zapData,
	}

	SetAllMocks(ctx)

	// Verify global mocks are set
	assert.Equal(t, trivyData, mockTrivyOutput, "Trivy mock should be set")
	assert.Equal(t, semgrepData, mockSemgrepOutput, "Semgrep mock should be set")
	assert.Equal(t, zapData, mockZAPOutput, "ZAP mock should be set")
}

// TestMockContext_SetAllMocks_Nil tests that SetAllMocks handles nil context gracefully.
func TestMockContext_SetAllMocks_Nil(t *testing.T) {
	defer ResetAllMocks()

	// Should not panic with nil context
	SetAllMocks(nil)

	// Mocks should remain nil
	assert.Nil(t, mockTrivyOutput, "Trivy mock should be nil")
	assert.Nil(t, mockSemgrepOutput, "Semgrep mock should be nil")
	assert.Nil(t, mockZAPOutput, "ZAP mock should be nil")
}

// TestMockContext_SetAllMocks_Partial tests that SetAllMocks handles partial mocks.
func TestMockContext_SetAllMocks_Partial(t *testing.T) {
	defer ResetAllMocks()

	trivyData := map[string]interface{}{"trivy": "mock"}

	ctx := &MockContext{
		TrivyOutput: trivyData,
		// SemgrepOutput and ZAPOutput left nil
	}

	SetAllMocks(ctx)

	// Verify only Trivy mock is set
	assert.Equal(t, trivyData, mockTrivyOutput, "Trivy mock should be set")
	assert.Nil(t, mockSemgrepOutput, "Semgrep mock should be nil")
	assert.Nil(t, mockZAPOutput, "ZAP mock should be nil")
}

// TestMockContext_ResetAllMocks tests that ResetAllMocks clears all global mocks.
func TestMockContext_ResetAllMocks(t *testing.T) {
	// Set all mocks
	trivyData := map[string]interface{}{"trivy": "mock"}
	semgrepData := map[string]interface{}{"semgrep": "mock"}
	zapData := map[string]interface{}{"zap": "mock"}

	SetMockTrivyOutput(trivyData)
	SetMockSemgrepOutput(semgrepData)
	SetMockZAPOutput(zapData)

	// Verify mocks are set
	require.NotNil(t, mockTrivyOutput)
	require.NotNil(t, mockSemgrepOutput)
	require.NotNil(t, mockZAPOutput)

	// Reset all mocks
	ResetAllMocks()

	// Verify all mocks are cleared
	assert.Nil(t, mockTrivyOutput, "Trivy mock should be cleared")
	assert.Nil(t, mockSemgrepOutput, "Semgrep mock should be cleared")
	assert.Nil(t, mockZAPOutput, "ZAP mock should be cleared")
}

// TestMockContext_HasMocksSet tests the HasMocksSet helper function.
func TestMockContext_HasMocksSet(t *testing.T) {
	defer ResetAllMocks()

	// Initially no mocks
	assert.False(t, HasMocksSet(), "Should have no mocks initially")

	// Set one mock
	SetMockTrivyOutput(map[string]interface{}{"test": "data"})
	assert.True(t, HasMocksSet(), "Should detect Trivy mock")

	// Reset and verify
	ResetAllMocks()
	assert.False(t, HasMocksSet(), "Should have no mocks after reset")

	// Set different mock
	SetMockSemgrepOutput(map[string]interface{}{"test": "data"})
	assert.True(t, HasMocksSet(), "Should detect Semgrep mock")

	ResetAllMocks()
	assert.False(t, HasMocksSet(), "Should have no mocks after second reset")
}

// TestMockContext_TestIsolation tests that mocks don't leak between tests.
func TestMockContext_TestIsolation(t *testing.T) {
	t.Run("first test sets mocks", func(t *testing.T) {
		defer ResetAllMocks()

		SetAllMocks(&MockContext{
			TrivyOutput: map[string]interface{}{"test": "1"},
		})

		assert.True(t, HasMocksSet(), "First test should have mocks")
	})

	t.Run("second test has clean state", func(t *testing.T) {
		// This test should start with no mocks if first test cleaned up properly
		assert.False(t, HasMocksSet(), "Second test should start with no mocks")
	})
}

// TestMockContext_RealWorldUsage demonstrates typical usage in tests.
func TestMockContext_RealWorldUsage(t *testing.T) {
	// This is how tests should use MockContext
	mockCtx := &MockContext{
		TrivyOutput: map[string]interface{}{
			"SchemaVersion": 2,
			"Results": []map[string]interface{}{
				{"Target": "test-module", "Vulnerabilities": []interface{}{}},
			},
		},
		SemgrepOutput: map[string]interface{}{
			"results": []interface{}{},
		},
	}

	// Set mocks at the beginning of test
	SetAllMocks(mockCtx)
	// Ensure cleanup happens even if test panics
	defer ResetAllMocks()

	// Test code would go here...
	// The security scanners will use the mock data

	// Verify mocks are actually set
	assert.True(t, HasMocksSet(), "Mocks should be set for the test")

	// Cleanup happens via defer
}
