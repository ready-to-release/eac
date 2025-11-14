// File: src/core/ai/config_loader_test.go
package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		envVars     map[string]string
		want        *Config
		wantErr     bool
	}{
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
			name: "claude-cli provider without API key",
			configYAML: `provider:
  name: claude-cli
  model: sonnet`,
			envVars: map[string]string{},
			want: &Config{
				ProviderName: "claude-cli",
				Model:        "sonnet",
				APIKey:       "",
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
			name:        "malformed YAML returns error",
			configYAML:  "invalid: yaml: content:",
			wantErr:     true,
		},
		{
			name: "missing provider name returns error",
			configYAML: `provider:
  model: some-model`,
			wantErr: true,
		},
		{
			name: "empty config file returns error",
			configYAML: "",
			wantErr: true,
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
