//go:build L0 && ov
// +build L0,ov

// File: go/cli/eac/impl/init/config_generator_test.go
package init

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/ready-to-release/eac/go/core/config"
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

	// Verify structure
	assert.Contains(t, result, "repository:")
	assert.Contains(t, result, "modules:")
	assert.Contains(t, result, "moniker: api-service")
	assert.Contains(t, result, "moniker: frontend")
	// Components must be in list format (- type: <language>)
	assert.Contains(t, result, "- type: go")
	assert.Contains(t, result, "- type: typescript")
	assert.Contains(t, result, "root: services/api")
	assert.Contains(t, result, "root: apps/web")
	// Must NOT contain old mapping format (language as a map key)
	assert.NotContains(t, result, "go:\n")
	assert.NotContains(t, result, "typescript:\n")
}

// TestRuleBasedGenerator_YAMLParseable verifies that generated YAML round-trips
// through ModuleComponents.UnmarshalYAML without error. This test catches format
// regressions — if components are emitted as a map instead of a list, this fails.
func TestRuleBasedGenerator_YAMLParseable(t *testing.T) {
	generator := NewRuleBasedGenerator()

	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{Name: "api-service", Root: "services/api", Language: "go", BuildTool: "go"},
			{Name: "frontend", Root: "apps/web", Language: "typescript", BuildTool: "npm"},
		},
	}

	result, err := generator.Generate("/tmp/test", scanResult)
	require.NoError(t, err)

	var cfg config.RepositoryConfig
	err = yaml.Unmarshal([]byte(result), &cfg)
	require.NoError(t, err, "generated YAML must be parseable as RepositoryConfig (check component list format)")
	assert.Len(t, cfg.Modules, 2)
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

// TestRuleBasedGenerator_LanguageTypes tests that component type matches the detected language
func TestRuleBasedGenerator_LanguageTypes(t *testing.T) {
	languages := []string{"go", "python", "rust", "typescript", "javascript", "dotnet", "java"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			generator := NewRuleBasedGenerator()
			scanResult := &ScanResult{
				Modules: []ModuleInfo{
					{
						Name:      "test-module",
						Root:      "test",
						Language:  lang,
						BuildTool: "test",
					},
				},
			}

			result, err := generator.Generate("/tmp/test", scanResult)
			require.NoError(t, err)
			// Component type should match the language
			assert.Contains(t, result, "- type: "+lang)
		})
	}
}

// TestRuleBasedGenerator_RepoNameFromWorkspaceRoot tests that the repository name
// is derived from the workspace root directory name, not hardcoded as "project"
func TestRuleBasedGenerator_RepoNameFromWorkspaceRoot(t *testing.T) {
	generator := NewRuleBasedGenerator()
	scanResult := &ScanResult{Modules: []ModuleInfo{}}

	result, err := generator.Generate("/home/user/my-project", scanResult)
	require.NoError(t, err)
	assert.Contains(t, result, "name: my-project")
	assert.NotContains(t, result, "name: project")
}

// TestRuleBasedGenerator_MonikerNormalization tests moniker name normalization
func TestRuleBasedGenerator_MonikerNormalization(t *testing.T) {
	tests := []struct {
		inputName     string
		expectedMoniker string
	}{
		{"API_Service", "api-service"},
		{"Web Frontend", "web-frontend"},
		{"@myorg/my-lib", "my-lib"},
		{"my/nested/pkg", "my-nested-pkg"},
	}

	for _, tt := range tests {
		t.Run(tt.inputName, func(t *testing.T) {
			generator := NewRuleBasedGenerator()
			scanResult := &ScanResult{
				Modules: []ModuleInfo{
					{Name: tt.inputName, Root: ".", Language: "go", BuildTool: "go"},
				},
			}
			result, err := generator.Generate("/tmp/test", scanResult)
			require.NoError(t, err)
			assert.Contains(t, result, "moniker: "+tt.expectedMoniker)
		})
	}
}

// TestRuleBasedGenerator_Description tests that descriptions are meaningful
func TestRuleBasedGenerator_Description(t *testing.T) {
	generator := NewRuleBasedGenerator()

	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{Name: "api-service", Root: ".", Language: "go", BuildTool: "go"},
			{Name: "frontend", Root: "apps/web", Language: "typescript", BuildTool: "npm"},
		},
	}

	result, err := generator.Generate("/tmp/test", scanResult)
	require.NoError(t, err)

	// Root module should say "Root <language> module"
	assert.Contains(t, result, "Root go module")
	// Non-root module should mention the name and path
	assert.Contains(t, result, "frontend")
	assert.Contains(t, result, "apps/web")
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

// TestModuleDescription_UsesNormalizedMoniker tests that description uses normalized moniker.
func TestModuleDescription_UsesNormalizedMoniker(t *testing.T) {
	mod := ModuleInfo{Name: "js_example", Language: "javascript", Root: "examples/js"}
	desc := moduleDescription(mod)
	expected := "js-example javascript module at examples/js"
	assert.Equal(t, expected, desc)
	assert.False(t, strings.Contains(desc, "js_example"), "raw name must not appear in description: %q", desc)
}

// TestNormalizeMoniker tests the moniker normalization function directly
func TestNormalizeMoniker(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"API_Service", "api-service"},
		{"Web Frontend", "web-frontend"},
		{"@myorg/my-lib", "my-lib"},
		{"my/nested/pkg", "my-nested-pkg"},
		{"--leading-hyphens--", "leading-hyphens"},
		{"hello world!", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeMoniker(tt.input))
		})
	}
}
