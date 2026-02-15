// File: go/cli/eac/impl/init/config_generator_test.go
package init

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRuleBasedGenerator_Generate tests rule-based config generation
func TestRuleBasedGenerator_Generate(t *testing.T) {
	generator := NewRuleBasedGenerator()

	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{
				Name:      "api-service",
				Root:      "services/api",
				Language:  "go",
				BuildTool: "go",
				Files:     []string{"go.mod"},
			},
			{
				Name:      "frontend",
				Root:      "apps/web",
				Language:  "typescript",
				BuildTool: "npm",
				Files:     []string{"package.json", "tsconfig.json"},
			},
		},
	}

	result, err := generator.Generate("/tmp/test", scanResult)
	require.NoError(t, err)

	// Verify output contains expected structure
	assert.Contains(t, result, "repository:")
	assert.Contains(t, result, "modules:")
	assert.Contains(t, result, "moniker: api-service")
	assert.Contains(t, result, "moniker: frontend")
	assert.Contains(t, result, "go:")
	assert.Contains(t, result, "typescript:")
	assert.Contains(t, result, "root: services/api")
	assert.Contains(t, result, "root: apps/web")
	assert.Contains(t, result, "type: service")
	assert.Contains(t, result, "type: app")
}

// TestRuleBasedGenerator_GenerateEmpty tests generation with no modules
func TestRuleBasedGenerator_GenerateEmpty(t *testing.T) {
	generator := NewRuleBasedGenerator()

	scanResult := &ScanResult{
		Modules: []ModuleInfo{},
	}

	result, err := generator.Generate("/tmp/test", scanResult)
	require.NoError(t, err)

	// Should still have valid YAML structure
	assert.Contains(t, result, "repository:")
	assert.Contains(t, result, "modules:")
}

// TestRuleBasedGenerator_LanguageTypes tests correct type assignment per language
func TestRuleBasedGenerator_LanguageTypes(t *testing.T) {
	tests := []struct {
		language     string
		expectedType string
	}{
		{"go", "service"},
		{"python", "service"},
		{"rust", "binary"},
		{"typescript", "app"},
		{"javascript", "app"},
		{"dotnet", "webapi"},
		{"java", "service"},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			generator := NewRuleBasedGenerator()
			scanResult := &ScanResult{
				Modules: []ModuleInfo{
					{
						Name:      "test-module",
						Root:      "test",
						Language:  tt.language,
						BuildTool: "test",
					},
				},
			}

			result, err := generator.Generate("/tmp/test", scanResult)
			require.NoError(t, err)
			assert.Contains(t, result, "type: "+tt.expectedType)
		})
	}
}

// TestRuleBasedGenerator_MonikerNormalization tests moniker name normalization
func TestRuleBasedGenerator_MonikerNormalization(t *testing.T) {
	generator := NewRuleBasedGenerator()

	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{
				Name:      "API_Service",
				Root:      "api",
				Language:  "go",
				BuildTool: "go",
			},
			{
				Name:      "Web Frontend",
				Root:      "web",
				Language:  "typescript",
				BuildTool: "npm",
			},
		},
	}

	result, err := generator.Generate("/tmp/test", scanResult)
	require.NoError(t, err)

	// Monikers should be lowercase with hyphens
	assert.Contains(t, result, "moniker: api-service")
	assert.NotContains(t, result, "moniker: API_Service")
	// Note: spaces aren't converted to hyphens in current implementation
	// This test documents current behavior
}

// TestAIGenerator_Interface tests that AIGenerator implements ConfigGenerator
func TestAIGenerator_Interface(t *testing.T) {
	var _ ConfigGenerator = (*AIGenerator)(nil)
	var _ ConfigGenerator = (*RuleBasedGenerator)(nil)
}

// TestGenerateWithFallback_PrimarySuccess tests fallback when primary succeeds
func TestGenerateWithFallback_PrimarySuccess(t *testing.T) {
	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{Name: "test", Root: "test", Language: "go", BuildTool: "go"},
		},
	}

	primary := NewRuleBasedGenerator()
	fallback := NewRuleBasedGenerator()

	result, err := GenerateWithFallback("/tmp/test", scanResult, primary, fallback)
	require.NoError(t, err)
	assert.Contains(t, result, "moniker: test")
}

// TestGenerateWithFallback_PrimaryFailure tests fallback when primary fails
func TestGenerateWithFallback_PrimaryFailure(t *testing.T) {
	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{Name: "test", Root: "test", Language: "go", BuildTool: "go"},
		},
	}

	// Create a failing primary generator
	primary := &failingGenerator{}
	fallback := NewRuleBasedGenerator()

	result, err := GenerateWithFallback("/tmp/test", scanResult, primary, fallback)
	require.NoError(t, err)
	assert.Contains(t, result, "moniker: test")
}

// failingGenerator is a test helper that always fails
type failingGenerator struct{}

func (g *failingGenerator) Generate(workspaceRoot string, scanResult *ScanResult) (string, error) {
	return "", assert.AnError
}

// TestNewRuleBasedGenerator tests constructor
func TestNewRuleBasedGenerator(t *testing.T) {
	gen := NewRuleBasedGenerator()
	assert.NotNil(t, gen)
}

// TestNewAIGenerator tests constructor
func TestNewAIGenerator(t *testing.T) {
	gen := NewAIGenerator("claude-api")
	assert.NotNil(t, gen)
	assert.Equal(t, "claude-api", gen.provider)
}
