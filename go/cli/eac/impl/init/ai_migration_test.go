//go:build L0 && ov
// +build L0,ov

// File: go/cli/eac/impl/init/ai_migration_test.go
package init

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/adapters/ai"
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
      go:
        root: go/myservice
        type: service
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
      go:
        root: services/api
        type: service

  - moniker: frontend
    description: Frontend web application
    depends_on:
      - api
    components:
      typescript:
        root: apps/frontend
        type: app
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
			mockProvider := ai.NewMockProvider(tt.aiResponse)
			executor := ai.NewExecutor(tmpDir)
			executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
				return mockProvider, nil
			})

			// Create mock AI config
			configDir := filepath.Join(tmpDir, ".r2r", "eac")
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

			// Set injected executor for testing
			SetAIExecutor(executor)
			defer ResetAIExecutor()

			// Execute
			got, err := GenerateConfig(tmpDir, tt.scanResult, "mock")

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
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
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
	mockProvider := ai.NewMockProvider(`repository:
  name: minimal-repo
  description: Minimal repository

modules: []
`)
	executor := ai.NewExecutor(tmpDir)
	executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
		return mockProvider, nil
	})
	SetAIExecutor(executor)
	defer ResetAIExecutor()

	// Empty scan results
	scanResult := &ScanResult{
		Modules: []ModuleInfo{},
	}

	got, err := GenerateConfig(tmpDir, scanResult, "mock")

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
	configDir := filepath.Join(tmpDir, ".r2r", "eac")
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

	executor := ai.NewExecutor(tmpDir)
	executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
		return mockProvider, nil
	})
	SetAIExecutor(executor)
	defer ResetAIExecutor()

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

	_, err := GenerateConfig(tmpDir, scanResult, "mock")
	require.NoError(t, err)

	// Verify the captured input contains valid JSON
	assert.Contains(t, capturedInput, "test-module")
	assert.Contains(t, capturedInput, "modules/test")

	// Extract the JSON part and validate it
	scanJSON, _ := json.MarshalIndent(scanResult, "", "  ")
	assert.Contains(t, capturedInput, string(scanJSON))
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
