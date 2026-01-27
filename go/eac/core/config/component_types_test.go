//go:build L1

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResources_CalculateWeight_NilResources(t *testing.T) {
	var r *Resources
	assert.Equal(t, 1, r.CalculateWeight(), "nil Resources should return weight 1")
}

func TestResources_CalculateWeight_EmptyResources(t *testing.T) {
	r := &Resources{}
	assert.Equal(t, 1, r.CalculateWeight(), "empty Resources should return weight 1")
}

func TestResources_CalculateWeight_RAMOnly(t *testing.T) {
	tests := []struct {
		name     string
		ram      string
		expected int
	}{
		{"1GB", "1GB", 1},
		{"2GB", "2GB", 1},
		{"3GB", "3GB", 2},
		{"4GB", "4GB", 2},
		{"8GB", "8GB", 4},
		{"16GB", "16GB", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resources{RAM: tt.ram}
			assert.Equal(t, tt.expected, r.CalculateWeight())
		})
	}
}

func TestResources_CalculateWeight_CoresOnly(t *testing.T) {
	tests := []struct {
		name     string
		cores    int
		expected int
	}{
		{"1 core", 1, 1},
		{"2 cores", 2, 2},
		{"4 cores", 4, 4},
		{"8 cores", 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resources{Cores: tt.cores}
			assert.Equal(t, tt.expected, r.CalculateWeight())
		})
	}
}

func TestResources_CalculateWeight_RAMAndCores(t *testing.T) {
	tests := []struct {
		name     string
		ram      string
		cores    int
		expected int
	}{
		{"1GB + 1 core", "1GB", 1, 1},        // max(1, 1) = 1
		{"2GB + 4 cores", "2GB", 4, 4},       // max(1, 4) = 4
		{"8GB + 2 cores", "8GB", 2, 4},       // max(4, 2) = 4
		{"8GB + 4 cores", "8GB", 4, 4},       // max(4, 4) = 4
		{"16GB + 4 cores", "16GB", 4, 8},     // max(8, 4) = 8
		{"4GB + 8 cores", "4GB", 8, 8},       // max(2, 8) = 8
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resources{RAM: tt.ram, Cores: tt.cores}
			assert.Equal(t, tt.expected, r.CalculateWeight())
		})
	}
}

func TestParseRAMSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"", 0},
		{"1GB", 1 * 1024 * 1024 * 1024},
		{"2GB", 2 * 1024 * 1024 * 1024},
		{"1gb", 1 * 1024 * 1024 * 1024}, // lowercase
		{"512MB", 512 * 1024 * 1024},
		{"1024KB", 1024 * 1024},
		{"1024B", 1024},
		{"1024", 1024}, // no suffix = bytes
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseRAMSize(tt.input))
		})
	}
}

func TestComponentType_GetBuildWeight_WithResources(t *testing.T) {
	ct := &ComponentType{
		Resources: &Resources{
			RAM:   "8GB",
			Cores: 4,
		},
	}

	// Should use Resources.CalculateWeight() = max(4, 4) = 4
	assert.Equal(t, 4, ct.GetBuildWeight())
}

func TestComponentType_GetBuildWeight_WithoutResources(t *testing.T) {
	ct := &ComponentType{
		BuildWeight: 3,
	}

	// Should use BuildWeight directly
	assert.Equal(t, 3, ct.GetBuildWeight())
}

func TestComponentType_GetBuildWeight_Default(t *testing.T) {
	ct := &ComponentType{}

	// Should default to 1
	assert.Equal(t, 1, ct.GetBuildWeight())
}

func TestComponentType_GetTestWeight_WithResources(t *testing.T) {
	ct := &ComponentType{
		Resources: &Resources{
			RAM:   "4GB",
			Cores: 2,
		},
	}

	// Should use Resources.CalculateWeight() = max(2, 2) = 2
	assert.Equal(t, 2, ct.GetTestWeight())
}

func TestComponentTypesDefaults_BookWeight(t *testing.T) {
	// Load defaults from contracts
	ctConfig, err := LoadComponentTypesDefaults("")
	if err != nil {
		t.Skipf("Could not load component types defaults: %v", err)
	}

	// Verify book component type has weight 4
	bookType := ctConfig.Get("book")
	if bookType == nil {
		t.Fatal("book component type not found in defaults")
	}

	weight := bookType.GetBuildWeight()
	assert.Equal(t, 4, weight, "book component type should have build_weight=4")
}
