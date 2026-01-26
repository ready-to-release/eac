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
			name:           "exactly 8GB returns 1 (low RAM penalty)",
			memoryBytes:    8 * 1024 * 1024 * 1024, // 8GB
			expectedResult: 1,                      // was 2, now 1 due to low RAM penalty
		},
		{
			name:           "12GB returns 1 (low RAM penalty)",
			memoryBytes:    12 * 1024 * 1024 * 1024, // 12GB
			expectedResult: 1,                       // was 2, now 1 due to low RAM penalty
		},
		{
			name:           "exactly 16GB returns 2 (low RAM penalty)",
			memoryBytes:    16 * 1024 * 1024 * 1024, // 16GB
			expectedResult: 2,                       // was 3, now 2 due to low RAM penalty
		},
		{
			name:           "17GB returns 3 (high RAM, no penalty)",
			memoryBytes:    17 * 1024 * 1024 * 1024, // 17GB
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

func TestCalculatePDFConcurrencyWithTurbo(t *testing.T) {
	tests := []struct {
		name           string
		memoryBytes    uint64
		turbo          bool
		expectedResult int
	}{
		// Low RAM (≤16GB) - turbo ignored
		{
			name:           "8GB with turbo still returns 1 (turbo ignored)",
			memoryBytes:    8 * 1024 * 1024 * 1024,
			turbo:          true,
			expectedResult: 1,
		},
		{
			name:           "16GB with turbo still returns 2 (turbo ignored)",
			memoryBytes:    16 * 1024 * 1024 * 1024,
			turbo:          true,
			expectedResult: 2,
		},
		// High RAM (>16GB) - turbo applies
		{
			name:           "17GB without turbo returns 3",
			memoryBytes:    17 * 1024 * 1024 * 1024,
			turbo:          false,
			expectedResult: 3,
		},
		{
			name:           "17GB with turbo returns 4",
			memoryBytes:    17 * 1024 * 1024 * 1024,
			turbo:          true,
			expectedResult: 4,
		},
		{
			name:           "32GB with turbo returns 4",
			memoryBytes:    32 * 1024 * 1024 * 1024,
			turbo:          true,
			expectedResult: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePDFConcurrencyWithTurbo(tt.memoryBytes, tt.turbo)
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
	assert.LessOrEqual(t, result, 4, "PDF concurrency should not exceed 4 (with turbo)")
}

func TestGetPDFExportConcurrencyCI(t *testing.T) {
	// CI environment should always return 2
	t.Setenv("CI", "true")

	result := GetPDFExportConcurrency()
	assert.Equal(t, 2, result, "CI should always return 2")

	// Turbo should be ignored in CI
	resultWithTurbo := GetPDFExportConcurrencyWithTurbo(true)
	assert.Equal(t, 2, resultWithTurbo, "CI should return 2 even with turbo")
}
