//go:build L1 && ov
// +build L1,ov

// File: src/core/ai/config_loader_test.go
package ai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		envVars     map[string]string
		want        *Config
		wantErr     bool
		errContains string // Expected error message content
	}{
		{
			name: "valid config with claude-cli provider",
			configYAML: `provider:
  name: claude-cli
  model: claude-3-haiku-20240307`,
			envVars: map[string]string{},
			want: &Config{
				ProviderName: "claude-cli",
				Model:        "claude-3-haiku-20240307",
				APIKey:       "",
			},
			wantErr: false,
		},
		{
			name: "valid config with env var substitution",
			configYAML: `provider:
  name: claude-api
  model: claude-3-haiku-20240307
  endpoint: https://api.anthropic.com/v1
  api_key: ${ANTHROPIC_API_KEY}`,
			envVars: map[string]string{"ANTHROPIC_API_KEY": "sk-ant-test"},
			want: &Config{
				ProviderName: "claude-api",
				Model:        "claude-3-haiku-20240307",
				Endpoint:     "https://api.anthropic.com/v1",
				APIKey:       "sk-ant-test",
			},
			wantErr: false,
		},
		{
			name: "missing env var results in empty string",
			configYAML: `provider:
  name: openai
  model: gpt-4-turbo
  api_key: ${MISSING_VAR}`,
			envVars: map[string]string{},
			want: &Config{
				ProviderName: "openai",
				Model:        "gpt-4-turbo",
				APIKey:       "",
			},
			wantErr: false,
		},
		{
			name:        "malformed YAML returns error with init instructions",
			configYAML:  "invalid: yaml: content:",
			wantErr:     true,
			errContains: "run: r2r init",
		},
		{
			name: "missing provider name returns error with init instructions",
			configYAML: `provider:
  model: some-model`,
			wantErr:     true,
			errContains: "provider name is required",
		},
		{
			name:        "empty config file returns error with init instructions",
			configYAML:  "",
			wantErr:     true,
			errContains: "run: r2r init",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "agent-config.yml")
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			// Test LoadConfig
			got, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// Verify error message contains expected instructions
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("LoadConfig() error = %v, want error containing %q", err, tt.errContains)
				}
				return // Skip comparison if we expected an error
			}

			// Compare results
			if got.ProviderName != tt.want.ProviderName {
				t.Errorf("ProviderName = %v, want %v", got.ProviderName, tt.want.ProviderName)
			}
			if got.Model != tt.want.Model {
				t.Errorf("Model = %v, want %v", got.Model, tt.want.Model)
			}
			if got.APIKey != tt.want.APIKey {
				t.Errorf("APIKey = %v, want %v", got.APIKey, tt.want.APIKey)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.yml")

	_, err := LoadConfig(nonExistentPath)
	if err == nil {
		t.Error("LoadConfig() expected error for non-existent file, got nil")
	}

	// Verify error message contains init instructions
	if !strings.Contains(err.Error(), ".r2r/eac/agent-config.yml not found") {
		t.Errorf("LoadConfig() error = %v, want error containing '.r2r/eac/agent-config.yml not found'", err)
	}
	if !strings.Contains(err.Error(), "run: r2r init") {
		t.Errorf("LoadConfig() error = %v, want error containing 'run: r2r init'", err)
	}
}

func TestLoadConfigFromRepoRoot(t *testing.T) {
	// Test that config is loaded from repository root
	tmpDir := t.TempDir()

	// Create agent-config.yml in repo root (tmpDir simulates repo root)
	configPath := filepath.Join(tmpDir, "agent-config.yml")
	configContent := `provider:
  name: claude-cli
  model: claude-3-haiku-20240307`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Load config
	got, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("LoadConfig() error = %v, want nil", err)
		return
	}

	// Verify config was loaded correctly
	if got.ProviderName != "claude-cli" {
		t.Errorf("ProviderName = %v, want claude-cli", got.ProviderName)
	}
	if got.Model != "claude-3-haiku-20240307" {
		t.Errorf("Model = %v, want claude-3-haiku-20240307", got.Model)
	}
}

func TestSubstituteEnvVars(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		envVars map[string]string
		want    string
	}{
		{
			name:    "single env var",
			input:   "${API_KEY}",
			envVars: map[string]string{"API_KEY": "secret"},
			want:    "secret",
		},
		{
			name:    "multiple env vars",
			input:   "${VAR1}-${VAR2}",
			envVars: map[string]string{"VAR1": "foo", "VAR2": "bar"},
			want:    "foo-bar",
		},
		{
			name:    "missing env var",
			input:   "${MISSING}",
			envVars: map[string]string{},
			want:    "",
		},
		{
			name:    "no env vars to substitute",
			input:   "literal-string",
			envVars: map[string]string{},
			want:    "literal-string",
		},
		{
			name:    "env var in middle of string",
			input:   "prefix-${VAR}-suffix",
			envVars: map[string]string{"VAR": "middle"},
			want:    "prefix-middle-suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got := substituteEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("substituteEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}
