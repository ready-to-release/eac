//go:build L1 && ov
// +build L1,ov

// File: go/eac/commands/internal/ai/executor_adapter_test.go
package ai_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/ai/providers"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

func TestExecutorAdapter_ImplementsAIExecutorInterface(t *testing.T) {
	// Verify ExecutorAdapter implements contracts.AIExecutor interface
	tmpDir := t.TempDir()
	executor := ai.NewExecutor(tmpDir)
	adapter := ai.NewExecutorAdapter(executor)

	// This will fail to compile if ExecutorAdapter doesn't implement AIExecutor
	var _ contracts.AIExecutor = adapter
}

func TestExecutorAdapter_ImplementsAIExecutorWithProviderInfo(t *testing.T) {
	// Verify ExecutorAdapter implements contracts.AIExecutorWithProviderInfo interface
	tmpDir := t.TempDir()
	executor := ai.NewExecutor(tmpDir)
	adapter := ai.NewExecutorAdapter(executor)

	// This will fail to compile if ExecutorAdapter doesn't implement AIExecutorWithProviderInfo
	var _ contracts.AIExecutorWithProviderInfo = adapter
}

func TestExecutorAdapter_Execute(t *testing.T) {
	tests := []struct {
		name          string
		configContent string
		createConfig  bool
		input         string
		modelOverride string
		wantErr       bool
		errContains   string
		wantResponse  string
	}{
		{
			name:         "execute with valid config",
			createConfig: true,
			configContent: fmt.Sprintf(`ai:
  provider: claude-cli
  model: %s
git:
  token: ""`, providers.DefaultClaudeCLIModel),
			input:        "test prompt",
			wantErr:      false,
			wantResponse: "mock response from claude-cli",
		},
		{
			name:         "execute with model override",
			createConfig: true,
			configContent: fmt.Sprintf(`ai:
  provider: mock
  model: default-model
git:
  token: ""`),
			input:         "test prompt",
			modelOverride: "override-model",
			wantErr:       false,
			wantResponse:  "mock response",
		},
		{
			name:         "execute with no config returns error",
			createConfig: false,
			input:        "test prompt",
			wantErr:      true,
			errContains:  "ai provider not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create config if needed
			if tt.createConfig {
				configPath := filepath.Join(tmpDir, ".r2r", "eac", "ai-provider.yml")
				if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					t.Fatalf("failed to create .r2r dir: %v", err)
				}
				if err := os.WriteFile(configPath, []byte(tt.configContent), 0644); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}

			// Create executor and adapter
			executor := ai.NewExecutor(tmpDir)
			executor.RegisterProvider("claude-cli", func(config *ai.Config) (ai.Provider, error) {
				return &testMockProviderWithName{name: "claude-cli", response: "mock response from claude-cli"}, nil
			})
			executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
				return &testMockProviderWithName{name: "mock", response: "mock response"}, nil
			})

			var adapter contracts.AIExecutor
			if tt.modelOverride != "" {
				adapter = ai.NewExecutorAdapterWithModel(executor, tt.modelOverride)
			} else {
				adapter = ai.NewExecutorAdapter(executor)
			}

			// Execute via interface
			ctx := context.Background()
			response, err := adapter.Execute(ctx, tt.input)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil {
					t.Errorf("Execute() error = nil, want error containing %q", tt.errContains)
				} else if err.Error() == "" {
					t.Errorf("Execute() error message empty, want containing %q", tt.errContains)
				}
				return
			}

			// Check response
			if !tt.wantErr && response != tt.wantResponse {
				t.Errorf("Execute() response = %v, want %v", response, tt.wantResponse)
			}
		})
	}
}

func TestExecutorAdapter_GetProviderName(t *testing.T) {
	tests := []struct {
		name             string
		configContent    string
		createConfig     bool
		executeFirst     bool
		wantProviderName string
	}{
		{
			name:         "get provider name after successful execution",
			createConfig: true,
			configContent: fmt.Sprintf(`ai:
  provider: claude-cli
  model: %s
git:
  token: ""`, providers.DefaultClaudeCLIModel),
			executeFirst:     true,
			wantProviderName: "claude-cli",
		},
		{
			name:         "get provider name without execution returns empty",
			createConfig: true,
			configContent: fmt.Sprintf(`ai:
  provider: mock
  model: test-model
git:
  token: ""`),
			executeFirst:     false,
			wantProviderName: "",
		},
		{
			name:             "get provider name with no config returns empty",
			createConfig:     false,
			executeFirst:     false,
			wantProviderName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create config if needed
			if tt.createConfig {
				configPath := filepath.Join(tmpDir, ".r2r", "eac", "ai-provider.yml")
				if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					t.Fatalf("failed to create .r2r dir: %v", err)
				}
				if err := os.WriteFile(configPath, []byte(tt.configContent), 0644); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}

			// Create executor and adapter
			executor := ai.NewExecutor(tmpDir)
			executor.RegisterProvider("claude-cli", func(config *ai.Config) (ai.Provider, error) {
				return &testMockProviderWithName{name: "claude-cli", response: "mock response from claude-cli"}, nil
			})
			executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
				return &testMockProviderWithName{name: "mock", response: "mock response"}, nil
			})

			adapter := ai.NewExecutorAdapter(executor)

			// Execute first if needed
			if tt.executeFirst {
				ctx := context.Background()
				_, err := adapter.Execute(ctx, "test prompt")
				if err != nil {
					t.Fatalf("Execute() failed: %v", err)
				}
			}

			// Call GetProviderName() directly (method exists on *ExecutorAdapter)
			providerName := adapter.GetProviderName()

			if providerName != tt.wantProviderName {
				t.Errorf("GetProviderName() = %v, want %v", providerName, tt.wantProviderName)
			}
		})
	}
}

func TestExecutorAdapter_ContextHandling(t *testing.T) {
	tests := []struct {
		name         string
		ctx          interface{}
		wantPanic    bool
		wantResponse string
	}{
		{
			name:         "execute with valid context.Context",
			ctx:          context.Background(),
			wantPanic:    false,
			wantResponse: "mock response",
		},
		{
			name:         "execute with nil context uses background",
			ctx:          nil,
			wantPanic:    false,
			wantResponse: "mock response",
		},
		{
			name:         "execute with invalid context type uses background",
			ctx:          "invalid context",
			wantPanic:    false,
			wantResponse: "mock response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create config
			configPath := filepath.Join(tmpDir, ".r2r", "eac", "ai-provider.yml")
			configContent := `ai:
  provider: mock
  model: test-model
git:
  token: ""`
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				t.Fatalf("failed to create .r2r dir: %v", err)
			}
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			executor := ai.NewExecutor(tmpDir)
			executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
				return &testMockProviderWithName{name: "mock", response: "mock response"}, nil
			})

			adapter := ai.NewExecutorAdapter(executor)

			// Execute with various context types
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("Execute() panicked unexpectedly: %v", r)
					}
				}
			}()

			response, err := adapter.Execute(tt.ctx, "test prompt")

			if err != nil {
				t.Errorf("Execute() error = %v, want nil", err)
				return
			}

			if response != tt.wantResponse {
				t.Errorf("Execute() response = %v, want %v", response, tt.wantResponse)
			}
		})
	}
}

func TestExecutorAdapter_OptionsHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config
	configPath := filepath.Join(tmpDir, ".r2r", "eac", "ai-provider.yml")
	configContent := `ai:
  provider: mock
  model: test-model
git:
  token: ""`
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("failed to create .r2r dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	executor := ai.NewExecutor(tmpDir)
	executor.RegisterProvider("mock", func(config *ai.Config) (ai.Provider, error) {
		return providers.NewMockProvider("mock response"), nil
	})

	adapter := ai.NewExecutorAdapter(executor)

	ctx := context.Background()

	// Test with ai.Option types (should be accepted)
	response, err := adapter.Execute(ctx, "test prompt",
		ai.WithModel("custom-model"),
		ai.WithTemperature(0.7),
	)

	if err != nil {
		t.Errorf("Execute() with valid options error = %v, want nil", err)
	}

	if response != "mock response" {
		t.Errorf("Execute() response = %v, want %v", response, "mock response")
	}

	// Test with invalid option types (should be ignored without panic)
	response, err = adapter.Execute(ctx, "test prompt", "invalid option", 123)

	if err != nil {
		t.Errorf("Execute() with invalid options error = %v, want nil", err)
	}

	if response != "mock response" {
		t.Errorf("Execute() response = %v, want %v", response, "mock response")
	}
}

// testMockProviderWithName is a test helper that allows configuring the provider name
// This is needed for testing GetProviderName() functionality
type testMockProviderWithName struct {
	name     string
	response string
}

func (p *testMockProviderWithName) Name() string {
	return p.name
}

func (p *testMockProviderWithName) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	return p.response, nil
}
