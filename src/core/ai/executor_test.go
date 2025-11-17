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
		name          string
		configContent string
		createConfig  bool
		envVars       map[string]string
		input         string
		wantProvider  string // Expected provider name
		wantErr       bool
		errContains   string // Check error contains this
		wantContains  string // Check response contains this
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
			name:         "execute with no config defaults to claude-cli",
			createConfig: false,
			input:        "test prompt",
			wantProvider: "claude-cli",
			wantErr:      false,
		},
		{
			name:         "execute with malformed config returns error",
			createConfig: true,
			configContent: `invalid: yaml: content:
  - broken`,
			input:       "test prompt",
			wantErr:     true,
			errContains: "failed to parse agent-config.yml",
		},
		{
			name:         "execute with malformed config suggests r2r init",
			createConfig: true,
			configContent: `invalid: yaml: content:
  - broken`,
			input:       "test prompt",
			wantErr:     true,
			errContains: "run: r2r agent init",
		},
		{
			name:         "execute with invalid provider returns error",
			createConfig: true,
			configContent: `provider:
  name: invalid-provider
  model: some-model`,
			input:       "test prompt",
			wantErr:     true,
			errContains: "unknown provider: invalid-provider",
		},
		{
			name:         "execute with invalid provider suggests r2r init",
			createConfig: true,
			configContent: `provider:
  name: invalid-provider
  model: some-model`,
			input:       "test prompt",
			wantErr:     true,
			errContains: "run: r2r agent init",
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
				t.Fatalf("failed to create .r2r dir: %v", err)
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

			// Verify error message if error expected
			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Execute() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			// For successful executions, check provider was used correctly
			if !tt.wantErr {
				// Mock provider returns input as output, so we can verify execution happened
				if response == "" {
					t.Errorf("Execute() returned empty response")
				}

				// Verify the correct provider was used
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

func TestExecutor_ExecuteWithDebug(t *testing.T) {
	tests := []struct {
		name            string
		debug           bool
		wantDebugInResp bool
	}{
		{
			name:            "execute with debug enabled includes logs in output",
			debug:           true,
			wantDebugInResp: true,
		},
		{
			name:            "execute with debug disabled excludes logs",
			debug:           false,
			wantDebugInResp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create config with claude-cli
			configPath := filepath.Join(tmpDir, ".r2r", "agent-config.yml")
			configContent := `provider:
  name: claude-cli
  model: sonnet`

			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatalf("failed to create .r2r dir: %v", err)
			}
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			executor := ai.NewExecutor(tmpDir)
			providers.RegisterBuiltIn(executor)

			ctx := context.Background()
			response, err := executor.Execute(ctx, "test prompt", ai.WithDebug(tt.debug))

			if err != nil {
				t.Errorf("Execute() error = %v, want nil", err)
				return
			}

			// Check if debug info is included in response
			hasDebugInfo := strings.Contains(response, "provider") ||
				strings.Contains(response, "timestamp") ||
				strings.Contains(response, "success")

			if tt.wantDebugInResp && !hasDebugInfo {
				t.Errorf("Execute() with debug=%v: response should include debug info, got %v", tt.debug, response)
			}

			if !tt.wantDebugInResp && hasDebugInfo {
				t.Errorf("Execute() with debug=%v: response should not include debug info, got %v", tt.debug, response)
			}
		})
	}
}

func TestExecutor_ExecuteWithDebugDefault(t *testing.T) {
	// Verify debug is false by default
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, ".r2r", "agent-config.yml")
	configContent := `provider:
  name: claude-cli
  model: sonnet`

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create .r2r dir: %v", err)
	}
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatalf("failed to create .r2r dir: %v", err)
			}
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	executor := ai.NewExecutor(tmpDir)
	providers.RegisterBuiltIn(executor)

	ctx := context.Background()
	response, err := executor.Execute(ctx, "test prompt")

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
		return
	}

	// Should not include debug info by default
	hasDebugInfo := strings.Contains(response, "provider") ||
		strings.Contains(response, "timestamp")

	if hasDebugInfo {
		t.Errorf("Execute() without debug option should not include debug info, got %v", response)
	}
}

func TestExecutor_NoLogFilesCreated(t *testing.T) {
	// Verify that NO log files are created in .r2r directory
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, ".r2r", "agent-config.yml")
	configContent := `provider:
  name: claude-cli
  model: sonnet`

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create .r2r dir: %v", err)
	}
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatalf("failed to create .r2r dir: %v", err)
			}
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	executor := ai.NewExecutor(tmpDir)
	providers.RegisterBuiltIn(executor)

	ctx := context.Background()

	// Execute without debug
	_, err := executor.Execute(ctx, "test prompt")
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Verify no .r2r/logs directory was created
	logsPath := filepath.Join(tmpDir, ".r2r", "logs")
	if _, err := os.Stat(logsPath); !os.IsNotExist(err) {
		t.Errorf("Execute() created .r2r/logs directory, but should not create any log files")
	}

	// Execute with debug enabled
	_, err = executor.Execute(ctx, "test prompt", ai.WithDebug(true))
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Still no .r2r/logs directory should exist
	if _, err := os.Stat(logsPath); !os.IsNotExist(err) {
		t.Errorf("Execute() with debug=true created .r2r/logs directory, but should not create any log files")
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
		t.Fatalf("failed to create .r2r dir: %v", err)
	}
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatalf("failed to create .r2r dir: %v", err)
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
		name         string
		config       *ai.Config
		wantProvider string
		wantErr      bool
		errContains  string
	}{
		{
			name: "load claude-cli provider",
			config: &ai.Config{
				ProviderName: "claude-cli",
				Model:        "sonnet",
			},
			wantProvider: "claude-cli",
			wantErr:      false,
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
			wantErr:      false,
		},
		{
			name: "load claude-api without API key returns error",
			config: &ai.Config{
				ProviderName: "claude-api",
				Model:        "claude-3-haiku-20240307",
				Endpoint:     "https://api.anthropic.com/v1",
				APIKey:       "",
			},
			wantErr:     true,
			errContains: "ANTHROPIC_API_KEY is required",
		},
		{
			name: "invalid provider returns error",
			config: &ai.Config{
				ProviderName: "invalid-provider",
				Model:        "some-model",
			},
			wantErr:     true,
			errContains: "unknown provider: invalid-provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			executor := ai.NewExecutor(tmpDir)
			providers.RegisterBuiltIn(executor)

			provider, err := executor.LoadProvider(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("LoadProvider() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if provider == nil {
				t.Errorf("LoadProvider() returned nil provider")
				return
			}

			if provider.Name() != tt.wantProvider {
				t.Errorf("LoadProvider() provider = %v, want %v", provider.Name(), tt.wantProvider)
			}
		})
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
		t.Fatalf("failed to create .r2r dir: %v", err)
	}
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatalf("failed to create .r2r dir: %v", err)
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

	ctx := context.Background()
	response, err := executor.Execute(ctx, "test input")

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if response != mockResponse {
		t.Errorf("Execute() response = %v, want %v", response, mockResponse)
	}
}
