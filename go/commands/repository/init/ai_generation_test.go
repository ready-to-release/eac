//go:build L0 && ov
// +build L0,ov

// File: go/cli/eac/impl/init/ai_generation_test.go
package init

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/aiproviders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateConfig_Success tests successful AI-powered config generation
func TestGenerateConfig_Success(t *testing.T) {
	tests := []struct {
		name         string
		scanResult   *ScanResult
		aiResponse   string
		wantContains []string
		wantErr      bool
	}{
		{
			name: "single Go module",
			scanResult: &ScanResult{
				Modules: []ModuleInfo{
					{
						Name:      "myservice",
						Root:      "go/myservice",
						Language:  "go",
						BuildTool: "go",
						Files:     []string{"go.mod"},
					},
				},
			},
			aiResponse: `repository:
  name: test-repo
  description: Test repository

modules:
  - moniker: myservice
    description: Backend service
    components:
      - type: go
        name: go
        root: go/myservice
`,
			wantContains: []string{
				"repository:",
				"name: test-repo",
				"myservice",
				"Backend service",
			},
			wantErr: false,
		},
		{
			name: "multi-module with dependencies",
			scanResult: &ScanResult{
				Modules: []ModuleInfo{
					{
						Name:      "api",
						Root:      "services/api",
						Language:  "go",
						BuildTool: "go",
					},
					{
						Name:      "frontend",
						Root:      "apps/frontend",
						Language:  "typescript",
						BuildTool: "npm",
					},
				},
			},
			aiResponse: `repository:
  name: full-stack-app
  description: Full-stack application

modules:
  - moniker: api
    description: Backend API service
    components:
      - type: go
        name: go
        root: services/api

  - moniker: frontend
    description: Frontend web application
    depends_on:
      - api
    components:
      - type: typescript
        name: typescript
        root: apps/frontend
`,
			wantContains: []string{
				"full-stack-app",
				"api",
				"frontend",
				"depends_on:",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary workspace
			tmpDir := t.TempDir()

			// Create templates directory and AI prompt template
			templatesDir := filepath.Join(tmpDir, "templates", "ai", "init")
			require.NoError(t, os.MkdirAll(templatesDir, 0755))
			templateContent := `# Test Template
Scan Results: {{.Custom.ScanResults}}
`
			require.NoError(t, os.WriteFile(
				filepath.Join(templatesDir, "scan-repository.md"),
				[]byte(templateContent),
				0644,
			))

			// Create mock AI executor
			mockProv := &simpleMockProvider{response: tt.aiResponse}
			registry := aiproviders.NewRegistry()
			registry.Register("mock", func(config *ai.ProviderConfig) (ai.Provider, error) {
				return mockProv, nil
			})
			executor := aiproviders.NewExecutor(tmpDir, registry)

			// Create mock AI config
			configDir := filepath.Join(tmpDir, ".eac")
			require.NoError(t, os.MkdirAll(configDir, 0755))
			configContent := `ai:
  provider: mock
  model: test-model
  api_key: test-key
git:
  token: ""
`
			require.NoError(t, os.WriteFile(
				filepath.Join(configDir, "ai-provider.yml"),
				[]byte(configContent),
				0644,
			))

			// Create deps with injected executor
			deps := &Deps{
				GetAIExecutor: func(root string) *aiproviders.Executor { return executor },
				GetGitRepo:    defaultDeps().GetGitRepo,
			}

			// Execute
			got, err := generateConfig(deps, tmpDir, tt.scanResult, "mock")

			// Verify
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, got)

			for _, want := range tt.wantContains {
				assert.Contains(t, got, want, "Response should contain: %s", want)
			}
		})
	}
}

// TestGenerateConfig_EmptyScanResults tests handling of empty scan results
func TestGenerateConfig_EmptyScanResults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create templates directory
	templatesDir := filepath.Join(tmpDir, "templates", "ai", "init")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(templatesDir, "scan-repository.md"),
		[]byte("Test template"),
		0644,
	))

	// Create mock config
	configDir := filepath.Join(tmpDir, ".eac")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	configContent := `ai:
  provider: mock
  model: test-model
git:
  token: ""
`
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "ai-provider.yml"),
		[]byte(configContent),
		0644,
	))

	// Create mock AI executor
	mockProv := &simpleMockProvider{response: `repository:
  name: minimal-repo
  description: Minimal repository

modules: []
`}
	registry := aiproviders.NewRegistry()
	registry.Register("mock", func(config *ai.ProviderConfig) (ai.Provider, error) {
		return mockProv, nil
	})
	executor := aiproviders.NewExecutor(tmpDir, registry)
	// Create deps with injected executor
	deps := &Deps{
		GetAIExecutor: func(root string) *aiproviders.Executor { return executor },
		GetGitRepo:    defaultDeps().GetGitRepo,
	}

	// Empty scan results
	scanResult := &ScanResult{
		Modules: []ModuleInfo{},
	}

	got, err := generateConfig(deps, tmpDir, scanResult, "mock")

	require.NoError(t, err)
	assert.Contains(t, got, "modules: []")
}

// TestGenerateConfig_AIProviderError tests handling of AI provider errors
func TestGenerateConfig_AIProviderError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create templates directory
	templatesDir := filepath.Join(tmpDir, "templates", "ai", "init")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(templatesDir, "scan-repository.md"),
		[]byte("Test template"),
		0644,
	))

	// Don't create AI config - this should cause an error

	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{Name: "test", Root: ".", Language: "go", BuildTool: "go"},
		},
	}

	_, err := GenerateConfig(tmpDir, scanResult, "invalid-provider")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "AI provider")
}

// TestGenerateConfig_MissingTemplate tests handling of missing template file
func TestGenerateConfig_MissingTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create templates directory - template loading should fail

	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{Name: "test", Root: ".", Language: "go", BuildTool: "go"},
		},
	}

	_, err := GenerateConfig(tmpDir, scanResult, "mock")

	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "template")
}

// TestGenerateConfig_ValidJSON tests that scan results are properly serialized
func TestGenerateConfig_ValidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create templates directory with a template that echoes the custom data
	templatesDir := filepath.Join(tmpDir, "templates", "ai", "init")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	templateContent := `Scan Results: {{.Custom.ScanResults}}`
	require.NoError(t, os.WriteFile(
		filepath.Join(templatesDir, "scan-repository.md"),
		[]byte(templateContent),
		0644,
	))

	// Create mock config
	configDir := filepath.Join(tmpDir, ".eac")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	configContent := `ai:
  provider: mock
  model: test-model
git:
  token: ""
`
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "ai-provider.yml"),
		[]byte(configContent),
		0644,
	))

	// Create a mock provider that captures the input
	var capturedInput string
	mockProvider := &mockProviderWithCapture{
		response: "repository:\n  name: test\nmodules: []",
		onExecute: func(ctx context.Context, input string) {
			capturedInput = input
		},
	}

	registry := aiproviders.NewRegistry()
	registry.Register("mock", func(config *ai.ProviderConfig) (ai.Provider, error) {
		return mockProvider, nil
	})
	executor := aiproviders.NewExecutor(tmpDir, registry)
	// Create deps with injected executor
	deps := &Deps{
		GetAIExecutor: func(root string) *aiproviders.Executor { return executor },
		GetGitRepo:    defaultDeps().GetGitRepo,
	}

	scanResult := &ScanResult{
		Modules: []ModuleInfo{
			{
				Name:      "test-module",
				Root:      "modules/test",
				Language:  "go",
				BuildTool: "go",
				Files:     []string{"go.mod"},
			},
		},
	}

	_, err := generateConfig(deps, tmpDir, scanResult, "mock")
	require.NoError(t, err)

	// Verify the captured input contains valid JSON
	assert.Contains(t, capturedInput, "test-module")
	assert.Contains(t, capturedInput, "modules/test")

	// Extract the JSON part and validate it
	scanJSON, _ := json.MarshalIndent(scanResult, "", "  ")
	assert.Contains(t, capturedInput, string(scanJSON))
}

// simpleMockProvider is a test provider that returns a configured response.
type simpleMockProvider struct {
	response string
}

func (p *simpleMockProvider) Name() string {
	return "mock"
}

func (p *simpleMockProvider) Execute(_ context.Context, _ string, _ ...ai.Option) (string, error) {
	return p.response, nil
}

// mockProviderWithCapture is a mock provider that captures input for verification
type mockProviderWithCapture struct {
	response  string
	onExecute func(ctx context.Context, input string)
}

func (p *mockProviderWithCapture) Name() string {
	return "mock-capture"
}

func (p *mockProviderWithCapture) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	if p.onExecute != nil {
		p.onExecute(ctx, input)
	}
	return p.response, nil
}

// TestLoadRepositoryContext_WithREADME tests loading context with a README file
func TestLoadRepositoryContext_WithREADME(t *testing.T) {
	tmpDir := t.TempDir()

	readmeContent := "# Test Project\n\nThis is a test project description.\nIt has multiple lines."
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "README.md"),
		[]byte(readmeContent),
		0644,
	))

	context := loadRepositoryContext(tmpDir)

	assert.Contains(t, context, "## README")
	assert.Contains(t, context, "Test Project")
	assert.Contains(t, context, "test project description")
}

// TestLoadRepositoryContext_NoREADME tests loading context without README
func TestLoadRepositoryContext_NoREADME(t *testing.T) {
	tmpDir := t.TempDir()

	context := loadRepositoryContext(tmpDir)

	assert.Contains(t, context, "(No README found)")
}

// TestLoadRepositoryContext_READMETruncation tests that large READMEs are truncated
func TestLoadRepositoryContext_READMETruncation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a README larger than 2000 characters
	largeContent := strings.Repeat("This is a long README. ", 100) // ~2300 chars
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "README.md"),
		[]byte(largeContent),
		0644,
	))

	context := loadRepositoryContext(tmpDir)

	assert.Contains(t, context, "## README")
	assert.Contains(t, context, "...\n(truncated)")
	// Should be around 2000 chars + header + truncation message
	assert.Less(t, len(context), 2100)
}

// TestLoadRepositoryContext_PriorityOrder tests README file priority
func TestLoadRepositoryContext_PriorityOrder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple README files
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "README.md"),
		[]byte("# Markdown README"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "README.txt"),
		[]byte("Text README"),
		0644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "README"),
		[]byte("Plain README"),
		0644,
	))

	context := loadRepositoryContext(tmpDir)

	// Should prefer README.md first
	assert.Contains(t, context, "Markdown README")
	assert.NotContains(t, context, "Text README")
	assert.NotContains(t, context, "Plain README")
}

// TestLoadRepositoryContext_READMETxtFallback tests fallback to README.txt
func TestLoadRepositoryContext_READMETxtFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Only create README.txt (no README.md)
	require.NoError(t, os.WriteFile(
		filepath.Join(tmpDir, "README.txt"),
		[]byte("Text README content"),
		0644,
	))

	context := loadRepositoryContext(tmpDir)

	assert.Contains(t, context, "## README")
	assert.Contains(t, context, "Text README content")
}
