package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanPhaseConfig_GetToolForCategory(t *testing.T) {
	cfg := &ScanPhaseConfig{
		SBOM:    "trivy-sbom",
		Vuln:    "trivy-vuln",
		Secrets: "trivy-secrets",
		SAST:    "semgrep",
	}

	tests := []struct {
		category ScanCategory
		expected string
	}{
		{ScanCategorySBOM, "trivy-sbom"},
		{ScanCategoryVuln, "trivy-vuln"},
		{ScanCategorySecrets, "trivy-secrets"},
		{ScanCategorySAST, "semgrep"},
		{ScanCategoryIAC, ""},
		{ScanCategoryCompliance, ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			result := cfg.GetToolForCategory(tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScanPhaseConfig_GetToolForCategory_Nil(t *testing.T) {
	var cfg *ScanPhaseConfig
	assert.Empty(t, cfg.GetToolForCategory(ScanCategorySBOM))
}

func TestScanPhaseConfig_GetDefaultCategories(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ScanPhaseConfig
		expected []ScanCategory
	}{
		{
			name:     "nil config uses global defaults",
			cfg:      nil,
			expected: DefaultScanCategories(),
		},
		{
			name:     "empty default uses global defaults",
			cfg:      &ScanPhaseConfig{},
			expected: DefaultScanCategories(),
		},
		{
			name: "custom defaults",
			cfg: &ScanPhaseConfig{
				Default: []ScanCategory{ScanCategorySBOM, ScanCategorySAST},
			},
			expected: []ScanCategory{ScanCategorySBOM, ScanCategorySAST},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.GetDefaultCategories()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComponentTools_BuildField(t *testing.T) {
	tests := []struct {
		name     string
		ct       *ComponentTools
		expected string
	}{
		{
			name: "build field set",
			ct: &ComponentTools{
				Build: "go",
			},
			expected: "go",
		},
		{
			name:     "no build tool",
			ct:       &ComponentTools{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ct.Build)
		})
	}
}

func TestComponentTools_LintField(t *testing.T) {
	tests := []struct {
		name     string
		ct       *ComponentTools
		expected string
	}{
		{
			name: "lint field set",
			ct: &ComponentTools{
				Lint: "golangci-lint",
			},
			expected: "golangci-lint",
		},
		{
			name:     "no lint tool",
			ct:       &ComponentTools{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ct.Lint)
		})
	}
}

func TestScanToolMapping_GetToolForCategory(t *testing.T) {
	mapping := &ScanToolMapping{
		SBOM:    "custom-sbom",
		Vuln:    "custom-vuln",
		Secrets: "custom-secrets",
	}

	assert.Equal(t, "custom-sbom", mapping.GetToolForCategory(ScanCategorySBOM))
	assert.Equal(t, "custom-vuln", mapping.GetToolForCategory(ScanCategoryVuln))
	assert.Equal(t, "custom-secrets", mapping.GetToolForCategory(ScanCategorySecrets))
	assert.Empty(t, mapping.GetToolForCategory(ScanCategorySAST))

	var nilMapping *ScanToolMapping
	assert.Empty(t, nilMapping.GetToolForCategory(ScanCategorySBOM))
}

func TestExtendedComponentType_GetBuilders(t *testing.T) {
	tests := []struct {
		name     string
		ect      *ExtendedComponentType
		expected []string
	}{
		{
			name:     "nil config",
			ect:      nil,
			expected: nil,
		},
		{
			name: "explicit builders field",
			ect: &ExtendedComponentType{
				Builders: []string{"go"},
			},
			expected: []string{"go"},
		},
		{
			name: "tools.build fallback",
			ect: &ExtendedComponentType{
				Tools: &ToolsConfig{
					Build: &PhaseConfig{Default: "go"},
				},
			},
			expected: []string{"go"},
		},
		{
			name: "builders takes precedence over tools",
			ect: &ExtendedComponentType{
				Builders: []string{"new-go"},
				Tools: &ToolsConfig{
					Build: &PhaseConfig{Default: "old-go"},
				},
			},
			expected: []string{"new-go"},
		},
		{
			name:     "empty returns nil",
			ect:      &ExtendedComponentType{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ect.GetBuilders()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtendedComponentType_GetScanCategories(t *testing.T) {
	tests := []struct {
		name     string
		ect      *ExtendedComponentType
		expected []ScanCategory
	}{
		{
			name:     "nil config",
			ect:      nil,
			expected: nil,
		},
		{
			name: "new tools.scan.default",
			ect: &ExtendedComponentType{
				Scanners: []string{"sbom", "vuln"},
				Tools: &ToolsConfig{
					Scan: &ScanPhaseConfig{
						Default: []ScanCategory{ScanCategorySAST},
					},
				},
			},
			expected: []ScanCategory{ScanCategorySAST},
		},
		{
			name: "legacy scanners field",
			ect: &ExtendedComponentType{
				Scanners: []string{"sbom", "vuln", "secrets"},
			},
			expected: []ScanCategory{ScanCategorySBOM, ScanCategoryVuln, ScanCategorySecrets},
		},
		{
			name:     "no scanners",
			ect:      &ExtendedComponentType{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ect.GetScanCategories()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtendedComponentType_IsBuildable(t *testing.T) {
	assert.True(t, (&ExtendedComponentType{Builders: []string{"go"}}).IsBuildable())
	assert.True(t, (&ExtendedComponentType{
		Tools: &ToolsConfig{Build: &PhaseConfig{Default: "go"}},
	}).IsBuildable())
	assert.False(t, (&ExtendedComponentType{}).IsBuildable())
	assert.False(t, (*ExtendedComponentType)(nil).IsBuildable())
}

func TestExtendedComponentType_IsScannable(t *testing.T) {
	assert.True(t, (&ExtendedComponentType{Scanners: []string{"sbom"}}).IsScannable())
	assert.True(t, (&ExtendedComponentType{
		Tools: &ToolsConfig{Scan: &ScanPhaseConfig{Default: []ScanCategory{ScanCategorySBOM}}},
	}).IsScannable())
	assert.False(t, (&ExtendedComponentType{}).IsScannable())
	assert.False(t, (*ExtendedComponentType)(nil).IsScannable())
}

func TestAllScanCategories(t *testing.T) {
	cats := AllScanCategories()
	assert.Len(t, cats, 7)
	assert.Contains(t, cats, ScanCategorySBOM)
	assert.Contains(t, cats, ScanCategoryVuln)
	assert.Contains(t, cats, ScanCategorySecrets)
	assert.Contains(t, cats, ScanCategorySAST)
	assert.Contains(t, cats, ScanCategoryIAC)
	assert.Contains(t, cats, ScanCategoryCompliance)
	assert.Contains(t, cats, ScanCategoryZAP)
}

func TestDefaultScanCategories(t *testing.T) {
	cats := DefaultScanCategories()
	assert.Len(t, cats, 3)
	assert.Contains(t, cats, ScanCategorySBOM)
	assert.Contains(t, cats, ScanCategoryVuln)
	assert.Contains(t, cats, ScanCategorySecrets)
}
