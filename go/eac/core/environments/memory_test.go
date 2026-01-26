//go:build L0
// +build L0

package environments

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPDFExportConcurrency(t *testing.T) {
	tests := []struct {
		name           string
		memoryBytes    uint64
		expectedResult int
	}{
		{
			name:           "less than 8GB returns 1",
			memoryBytes:    7 * 1024 * 1024 * 1024, // 7GB
			expectedResult: 1,
		},
		{
			name:           "exactly 8GB returns 2",
			memoryBytes:    8 * 1024 * 1024 * 1024, // 8GB
			expectedResult: 2,
		},
		{
			name:           "12GB returns 2",
			memoryBytes:    12 * 1024 * 1024 * 1024, // 12GB
			expectedResult: 2,
		},
		{
			name:           "exactly 16GB returns 3",
			memoryBytes:    16 * 1024 * 1024 * 1024, // 16GB
			expectedResult: 3,
		},
		{
			name:           "32GB returns 3",
			memoryBytes:    32 * 1024 * 1024 * 1024, // 32GB
			expectedResult: 3,
		},
		{
			name:           "4GB returns 1",
			memoryBytes:    4 * 1024 * 1024 * 1024, // 4GB
			expectedResult: 1,
		},
		{
			name:           "zero memory returns 1",
			memoryBytes:    0,
			expectedResult: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePDFConcurrency(tt.memoryBytes)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestGetSystemMemoryBytes(t *testing.T) {
	// This test verifies that GetSystemMemoryBytes returns a reasonable value
	// on the current system (must be > 0 for any real system)
	mem := GetSystemMemoryBytes()
	assert.Greater(t, mem, uint64(0), "System memory should be greater than 0")

	// Sanity check: should be at least 512MB for any modern system
	minExpected := uint64(512 * 1024 * 1024)
	assert.Greater(t, mem, minExpected, "System memory should be at least 512MB")

	// Log detected values for diagnostics
	t.Logf("Detected system memory: %.1f GB", float64(mem)/(1024*1024*1024))
	t.Logf("PDF export concurrency: %d", GetPDFExportConcurrency())
}

func TestGetPDFExportConcurrencyDefault(t *testing.T) {
	// GetPDFExportConcurrency should return at least 1 on any system
	result := GetPDFExportConcurrency()
	assert.GreaterOrEqual(t, result, 1, "PDF concurrency should be at least 1")
	assert.LessOrEqual(t, result, 3, "PDF concurrency should not exceed 3")
}
