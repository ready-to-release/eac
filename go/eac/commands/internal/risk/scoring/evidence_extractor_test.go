package scoring

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/evidence"
	"github.com/ready-to-release/eac/go/eac/core/domain"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
	"github.com/stretchr/testify/assert"
)

func TestExtractVulnerabilityFindings_NilInput(t *testing.T) {
	findings := ExtractVulnerabilityFindings(nil)
	assert.Empty(t, findings)
}

func TestExtractVulnerabilityFindings_NoSecurityResults(t *testing.T) {
	ec := &evidence.EvidenceCollection{
		Module:          "test-module",
		SecurityResults: nil,
	}
	findings := ExtractVulnerabilityFindings(ec)
	assert.Empty(t, findings)
}

func TestExtractVulnerabilityFindings_EmptySecurityResults(t *testing.T) {
	ec := &evidence.EvidenceCollection{
		Module: "test-module",
		SecurityResults: &evidence.SecurityResults{
			ModuleName: "test-module",
		},
	}
	findings := ExtractVulnerabilityFindings(ec)
	assert.Empty(t, findings)
}

func TestBuildModuleContext_NilRegistry(t *testing.T) {
	ctx := BuildModuleContext("test-module", nil, []string{"ac-2", "ac-3"})

	assert.Equal(t, "service", ctx.ModuleType)
	assert.Equal(t, "medium", ctx.Criticality)
	assert.Equal(t, []string{"ac-2", "ac-3"}, ctx.ExistingControls)
}

func TestBuildModuleContext_ModuleNotFound(t *testing.T) {
	registry := modules.NewRegistry("0.1.0", "/test")

	ctx := BuildModuleContext("nonexistent", registry, []string{"ac-2"})

	assert.Equal(t, "service", ctx.ModuleType)
	assert.Equal(t, "medium", ctx.Criticality)
	assert.Equal(t, []string{"ac-2"}, ctx.ExistingControls)
}

func TestBuildModuleContext_WithModule(t *testing.T) {
	registry := modules.NewRegistry("0.1.0", "/test")

	// Create a module contract with "go" component
	module := modules.NewModuleContract(domain.BaseContract{
		Moniker: "api-gateway",
		Components: domain.ModuleComponents{
			"go": &domain.ComponentEntry{Root: "go/test"},
		},
	}, "/test")

	err := registry.Add(module)
	assert.NoError(t, err)

	ctx := BuildModuleContext("api-gateway", registry, []string{"ia-2", "ia-5"})

	// Module type is determined by component type (go), criticality defaults to medium
	assert.Equal(t, "go", ctx.ModuleType)
	assert.Equal(t, "medium", ctx.Criticality)
	assert.Equal(t, []string{"ia-2", "ia-5"}, ctx.ExistingControls)
}

func TestDetermineCriticality(t *testing.T) {
	tests := []struct {
		name       string
		moduleType string
		expected   string
	}{
		{
			name:       "API is high criticality",
			moduleType: "api",
			expected:   "high",
		},
		{
			name:       "Gateway is high criticality",
			moduleType: "gateway",
			expected:   "high",
		},
		{
			name:       "Service is high criticality",
			moduleType: "service",
			expected:   "high",
		},
		{
			name:       "Core is medium criticality",
			moduleType: "core",
			expected:   "medium",
		},
		{
			name:       "Library is medium criticality",
			moduleType: "library",
			expected:   "medium",
		},
		{
			name:       "CLI is low criticality",
			moduleType: "cli",
			expected:   "low",
		},
		{
			name:       "Tool is low criticality",
			moduleType: "tool",
			expected:   "low",
		},
		{
			name:       "Unknown defaults to medium",
			moduleType: "unknown",
			expected:   "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineCriticality(tt.moduleType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapGoSecSeverity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"HIGH maps to HIGH", "HIGH", "HIGH"},
		{"high maps to HIGH", "high", "HIGH"},
		{"MEDIUM maps to MEDIUM", "MEDIUM", "MEDIUM"},
		{"medium maps to MEDIUM", "medium", "MEDIUM"},
		{"LOW maps to LOW", "LOW", "LOW"},
		{"low maps to LOW", "low", "LOW"},
		{"Unknown defaults to MEDIUM", "unknown", "MEDIUM"},
		{"Empty defaults to MEDIUM", "", "MEDIUM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapGoSecSeverity(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapZAPRisk(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"High maps to HIGH", "High", "HIGH"},
		{"HIGH maps to HIGH", "HIGH", "HIGH"},
		{"Medium maps to MEDIUM", "Medium", "MEDIUM"},
		{"MEDIUM maps to MEDIUM", "MEDIUM", "MEDIUM"},
		{"Low maps to LOW", "Low", "LOW"},
		{"LOW maps to LOW", "LOW", "LOW"},
		{"Informational maps to LOW", "Informational", "LOW"},
		{"INFORMATIONAL maps to LOW", "INFORMATIONAL", "LOW"},
		{"Unknown defaults to MEDIUM", "unknown", "MEDIUM"},
		{"Empty defaults to MEDIUM", "", "MEDIUM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapZAPRisk(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "Returns string value",
			m:        map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "Returns empty for missing key",
			m:        map[string]interface{}{"other": "value"},
			key:      "key",
			expected: "",
		},
		{
			name:     "Returns empty for non-string value",
			m:        map[string]interface{}{"key": 123},
			key:      "key",
			expected: "",
		},
		{
			name:     "Returns empty for nil map",
			m:        nil,
			key:      "key",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
