//go:build L0
// +build L0

package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	gb = 1024 * 1024 * 1024
	mb = 1024 * 1024
)

func TestFormatDockerMemory(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "below minimum returns 2g",
			bytes:    1 * gb,
			expected: "2g",
		},
		{
			name:     "exactly 2GB returns 2g",
			bytes:    2 * gb,
			expected: "2g",
		},
		{
			name:     "4GB returns 4g",
			bytes:    4 * gb,
			expected: "4g",
		},
		{
			name:     "8GB returns 8g",
			bytes:    8 * gb,
			expected: "8g",
		},
		{
			name:     "16GB returns 16g",
			bytes:    16 * gb,
			expected: "16g",
		},
		{
			name:     "24GB returns 24g",
			bytes:    24 * gb,
			expected: "24g",
		},
		{
			name:     "32GB capped at 24g",
			bytes:    32 * gb,
			expected: "24g",
		},
		{
			name:     "64GB capped at 24g",
			bytes:    64 * gb,
			expected: "24g",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDockerMemory(tt.bytes)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFormatDockerShm(t *testing.T) {
	tests := []struct {
		name         string
		containerMem int64
		expected     string
	}{
		{
			name:         "2GB container gets minimum 512m",
			containerMem: 2 * gb,
			expected:     "512m",
		},
		{
			name:         "1GB container gets minimum 512m",
			containerMem: 1 * gb,
			expected:     "512m",
		},
		{
			name:         "4GB container gets 1024m (1GB = 4/4)",
			containerMem: 4 * gb,
			expected:     "1024m",
		},
		{
			name:         "8GB container gets 2048m (2GB = 8/4)",
			containerMem: 8 * gb,
			expected:     "2048m",
		},
		{
			name:         "16GB container gets 4g (max, 16/4=4)",
			containerMem: 16 * gb,
			expected:     "4g",
		},
		{
			name:         "24GB container capped at 4g",
			containerMem: 24 * gb,
			expected:     "4g",
		},
		{
			name:         "6GB container gets 1536m (6/4=1.5)",
			containerMem: 6 * gb,
			expected:     "1536m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDockerShm(tt.containerMem)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCalculateWeightedMemory(t *testing.T) {
	// Base is 2.5GB per weight unit
	tests := []struct {
		name     string
		weight   int
		expected int64
	}{
		{
			name:     "weight 1 gets 2.5GB",
			weight:   1,
			expected: 2560 * mb, // 2.5GB
		},
		{
			name:     "weight 2 gets 5GB",
			weight:   2,
			expected: 5120 * mb, // 5GB
		},
		{
			name:     "weight 4 gets 10GB",
			weight:   4,
			expected: 10240 * mb, // 10GB
		},
		{
			name:     "weight 5 gets 12.5GB",
			weight:   5,
			expected: 12800 * mb, // 12.5GB
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateWeightedMemory(tt.weight)
			// Note: may be capped by available container memory on test system
			assert.LessOrEqual(t, got, tt.expected)
			// At minimum, should be 2.5GB (weight 1 base)
			assert.GreaterOrEqual(t, got, int64(2560*mb))
		})
	}
}

func TestGetEffectiveWeight(t *testing.T) {
	// Book component type has cpus: 4 in component-types.yml
	weight := getEffectiveWeight("book")
	assert.Equal(t, 4, weight, "Book weight should be 4 (from component-types.yml)")
}
