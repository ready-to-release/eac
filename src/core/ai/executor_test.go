// File: src/core/ai/executor_test.go
package ai_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/core/ai"
	"github.com/ready-to-release/eac/src/core/ai/providers"
)

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		createConfig   bool
		envVars        map[string]string
		input          string
		wantProvider   string // Expected provider name
		wantErr        bool
		wantContains   string // Check response contains this
	}{
		{
			name:         "execute with claude-cli configured",
			createConfig: true,
			configContent: `provider:
  name: claude-cli
  model: sonnet`,
			input:        "test prompt",
			wantProvider: "claude-cli",
			wantErr:      false,
		},
		{
			name:         "execute with claude-api configured (SDK not implemented, should error)",
			createConfig: true,
			configContent: `provider:
  name: claude-api
  model: claude-3-haiku-20240307
  endpoint: https://api.anthropic.com/v1
  api_key: ${ANTHROPIC_API_KEY}`,
			envVars: map[string]string{
				"ANTHROPIC_API_KEY": "test-key",
			},
			input:        "test prompt",
			wantProvider: "claude-api",
			wantErr:      true, // SDK not yet integrated
		},
		{
			name:         "execute with no config falls back to claude-cli",
			createConfig: false,
			input:        "test prompt",
			wantProvider: "claude-cli",
			wantErr:      false,
		},
		{
			name:         "execute with malformed config falls back to claude-cli",
			createConfig: true,
			configContent: `invalid: yaml: content:
  - broken`,
			input:        "test prompt",
			wantProvider: "claude-cli",
			wantErr:      false,
		},
		{
			name:         "execute with missing API key falls back to claude-cli",
			createConfig: true,
			configContent: `provider:
  name: claude-api
  model: claude-3-haiku-20240307
  endpoint: https://api.anthropic.com/v1
  api_key: ${ANTHROPIC_API_KEY}`,
			envVars:      map[string]string{},
			input:        "test prompt",
			wantProvider: "claude-cli",
			wantErr:      false,
		},
		{
			name:         "execute with invalid provider falls back to claude-cli",
			createConfig: true,
			configContent: `provider:
  name: invalid-provider
  model: some-model`,
			input:        "test prompt",
			wantProvider: "claude-cli",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create temporary directory for config
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".r2r", "agent-config.yml")

			// Create config file if needed
			if tt.createConfig {
				if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					t.Fatalf("failed to create config dir: %v", err)
				}
				if err := os.WriteFile(configPath, []byte(tt.configContent), 0644); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}

			// Create executor with test workspace root
			executor := ai.NewExecutor(tmpDir)
			providers.RegisterBuiltIn(executor)

			// Execute
			ctx := context.Background()
			response, err := executor.Execute(ctx, tt.input)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For successful executions, check provider was used correctly
			if !tt.wantErr {
				// Mock provider returns input as output, so we can verify execution happened
				if response == "" {
					t.Errorf("Execute() returned empty response")
				}

				// Verify the correct provider was used
				// We can check this by examining the executor's last used provider
				usedProvider := executor.GetLastUsedProvider()
				if usedProvider == nil {
					t.Errorf("Execute() did not set last used provider")
				} else if usedProvider.Name() != tt.wantProvider {
					t.Errorf("Execute() used provider %v, want %v", usedProvider.Name(), tt.wantProvider)
				}
			}

			// Check response contains expected string if specified
			if tt.wantContains != "" && !strings.Contains(response, tt.wantContains) {
				t.Errorf("Execute() response = %v, want to contain %v", response, tt.wantContains)
			}
		})
	}
}

func TestExecutor_ExecuteWithOptions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config with claude-cli
	configPath := filepath.Join(tmpDir, ".r2r", "agent-config.yml")
	configContent := `provider:
  name: claude-cli
  model: sonnet`

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	executor := ai.NewExecutor(tmpDir)
	providers.RegisterBuiltIn(executor)
	ctx := context.Background()

	// Test with options
	response, err := executor.Execute(ctx, "test prompt",
		ai.WithModel("opus"),
		ai.WithTemperature(0.7),
		ai.WithMaxTokens(1000),
	)

	if err != nil {
		t.Errorf("Execute() with options error = %v, want nil", err)
	}

	if response == "" {
		t.Errorf("Execute() returned empty response")
	}

	// Verify options were passed to provider
	// (This would be verified in integration tests with real provider)
}

func TestExecutor_LoadProvider(t *testing.T) {
	tests := []struct {
		name          string
		config        *ai.Config
		wantProvider  string
		wantFallback  bool
	}{
		{
			name: "load claude-cli provider",
			config: &ai.Config{
				ProviderName: "claude-cli",
				Model:        "sonnet",
			},
			wantProvider: "claude-cli",
			wantFallback: false,
		},
		{
			name: "load claude-api provider with API key",
			config: &ai.Config{
				ProviderName: "claude-api",
				Model:        "claude-3-haiku-20240307",
				Endpoint:     "https://api.anthropic.com/v1",
				APIKey:       "test-key",
			},
			wantProvider: "claude-api",
			wantFallback: false,
		},
		{
			name: "load claude-api without API key falls back",
			config: &ai.Config{
				ProviderName: "claude-api",
				Model:        "claude-3-haiku-20240307",
				Endpoint:     "https://api.anthropic.com/v1",
				APIKey:       "",
			},
			wantProvider: "claude-cli",
			wantFallback: true,
		},
		{
			name: "invalid provider falls back",
			config: &ai.Config{
				ProviderName: "invalid-provider",
				Model:        "some-model",
			},
			wantProvider: "claude-cli",
			wantFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			executor := ai.NewExecutor(tmpDir)
			providers.RegisterBuiltIn(executor)

			provider, didFallback := executor.LoadProvider(tt.config)

			if provider == nil {
				t.Errorf("loadProvider() returned nil provider")
				return
			}

			if provider.Name() != tt.wantProvider {
				t.Errorf("loadProvider() provider = %v, want %v", provider.Name(), tt.wantProvider)
			}

			if didFallback != tt.wantFallback {
				t.Errorf("loadProvider() fallback = %v, want %v", didFallback, tt.wantFallback)
			}
		})
	}
}

func TestExecutor_FallbackChain(t *testing.T) {
	// Test the fallback chain: config → claude-cli
	tmpDir := t.TempDir()

	// No config file exists
	executor := ai.NewExecutor(tmpDir)
	providers.RegisterBuiltIn(executor)
	ctx := context.Background()

	response, err := executor.Execute(ctx, "test prompt")

	if err != nil {
		t.Errorf("Execute() with no config error = %v, want nil (should fallback)", err)
	}

	if response == "" {
		t.Errorf("Execute() returned empty response")
	}

	// Verify fallback to claude-cli
	usedProvider := executor.GetLastUsedProvider()
	if usedProvider == nil {
		t.Errorf("Execute() did not set last used provider")
	} else if usedProvider.Name() != "claude-cli" {
		t.Errorf("Execute() used provider %v, want claude-cli (fallback)", usedProvider.Name())
	}
}

func TestExecutor_WithMockProvider(t *testing.T) {
	// Test executor with mock provider for predictable results
	tmpDir := t.TempDir()

	// Create config with mock provider
	configPath := filepath.Join(tmpDir, ".r2r", "agent-config.yml")
	configContent := `provider:
  name: mock
  model: test-model`

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	executor := ai.NewExecutor(tmpDir)

	// Register mock provider factory
	mockResponse := "mock response"
	executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
		return providers.NewMockProvider(mockResponse), nil
	})

	// Set mock as fallback too
	executor.SetFallbackProvider(func(config *ai.Config) (ai.Provider, error) {
		return providers.NewMockProvider(mockResponse), nil
	})

	ctx := context.Background()
	response, err := executor.Execute(ctx, "test input")

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if response != mockResponse {
		t.Errorf("Execute() response = %v, want %v", response, mockResponse)
	}
}
