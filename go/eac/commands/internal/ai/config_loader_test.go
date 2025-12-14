//go:build L1 && ov
// +build L1,ov

// File: go/eac/commands/internal/ai/config_loader_test.go
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
			configYAML: `ai:
  provider: claude-cli
  model: claude-3-haiku-20240307
git:
  token: ""`,
			envVars: map[string]string{},
			want: &Config{
				AI: AIConfig{
					Provider: "claude-cli",
					Model:    "claude-3-haiku-20240307",
					APIKey:   "",
				},
				Git: GitConfig{
					Token: "",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with env var substitution",
			configYAML: `ai:
  provider: claude-api
  model: claude-3-haiku-20240307
  endpoint: https://api.anthropic.com/v1
  api_key: ${ANTHROPIC_API_KEY}
git:
  token: ${GIT_TOKEN}`,
			envVars: map[string]string{"ANTHROPIC_API_KEY": "sk-ant-test", "GIT_TOKEN": "ghp_test"},
			want: &Config{
				AI: AIConfig{
					Provider: "claude-api",
					Model:    "claude-3-haiku-20240307",
					Endpoint: "https://api.anthropic.com/v1",
					APIKey:   "sk-ant-test",
				},
				Git: GitConfig{
					Token: "ghp_test",
				},
			},
			wantErr: false,
		},
		{
			name: "missing env var returns error with instructions",
			configYAML: `ai:
  provider: openai
  model: gpt-4-turbo
  api_key: ${MISSING_VAR}
git:
  token: ""`,
			envVars:     map[string]string{},
			wantErr:     true,
			errContains: "missing environment variable",
		},
		{
			name:        "malformed YAML returns error with init instructions",
			configYAML:  "invalid: yaml: content:",
			wantErr:     true,
			errContains: "run: eac init",
		},
		{
			name: "missing provider name returns ErrAIProviderNotConfigured",
			configYAML: `ai:
  model: some-model
git:
  token: ""`,
			wantErr:     true,
			errContains: "ai provider not configured",
		},
		{
			name:        "empty config file returns ErrAIProviderNotConfigured",
			configYAML:  "",
			wantErr:     true,
			errContains: "ai provider not configured",
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
			configPath := filepath.Join(tmpDir, "ai-provider.yml")
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
			if got.AI.Provider != tt.want.AI.Provider {
				t.Errorf("AI.Provider = %v, want %v", got.AI.Provider, tt.want.AI.Provider)
			}
			if got.AI.Model != tt.want.AI.Model {
				t.Errorf("AI.Model = %v, want %v", got.AI.Model, tt.want.AI.Model)
			}
			if got.AI.APIKey != tt.want.AI.APIKey {
				t.Errorf("AI.APIKey = %v, want %v", got.AI.APIKey, tt.want.AI.APIKey)
			}
			if got.Git.Token != tt.want.Git.Token {
				t.Errorf("Git.Token = %v, want %v", got.Git.Token, tt.want.Git.Token)
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

	// Verify error is ErrAIProviderNotConfigured
	if err != ErrAIProviderNotConfigured {
		t.Errorf("LoadConfig() error = %v, want ErrAIProviderNotConfigured", err)
	}

	// Verify error message contains proper instructions
	if !strings.Contains(err.Error(), "ai provider not configured") {
		t.Errorf("LoadConfig() error = %v, want error containing 'ai provider not configured'", err)
	}
	if !strings.Contains(err.Error(), "eac init --ai-provider") {
		t.Errorf("LoadConfig() error = %v, want error containing 'eac init --ai-provider'", err)
	}
}

func TestLoadConfigFromRepoRoot(t *testing.T) {
	// Test that config is loaded from repository root
	tmpDir := t.TempDir()

	// Create ai-provider.yml in repo root (tmpDir simulates repo root)
	configPath := filepath.Join(tmpDir, "ai-provider.yml")
	configContent := `ai:
  provider: claude-cli
  model: claude-3-haiku-20240307
git:
  token: ""`

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
	if got.AI.Provider != "claude-cli" {
		t.Errorf("AI.Provider = %v, want claude-cli", got.AI.Provider)
	}
	if got.AI.Model != "claude-3-haiku-20240307" {
		t.Errorf("AI.Model = %v, want claude-3-haiku-20240307", got.AI.Model)
	}
}

func TestSubstituteEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		envVars     map[string]string
		want        string
		wantMissing []string
	}{
		{
			name:        "single env var",
			input:       "${API_KEY}",
			envVars:     map[string]string{"API_KEY": "secret"},
			want:        "secret",
			wantMissing: nil,
		},
		{
			name:        "multiple env vars",
			input:       "${VAR1}-${VAR2}",
			envVars:     map[string]string{"VAR1": "foo", "VAR2": "bar"},
			want:        "foo-bar",
			wantMissing: nil,
		},
		{
			name:        "missing env var",
			input:       "${MISSING}",
			envVars:     map[string]string{},
			want:        "",
			wantMissing: []string{"MISSING"},
		},
		{
			name:        "no env vars to substitute",
			input:       "literal-string",
			envVars:     map[string]string{},
			want:        "literal-string",
			wantMissing: nil,
		},
		{
			name:        "env var in middle of string",
			input:       "prefix-${VAR}-suffix",
			envVars:     map[string]string{"VAR": "middle"},
			want:        "prefix-middle-suffix",
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got, missing := substituteEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("substituteEnvVars() result = %v, want %v", got, tt.want)
			}
			if len(missing) != len(tt.wantMissing) {
				t.Errorf("substituteEnvVars() missing = %v, want %v", missing, tt.wantMissing)
			} else {
				for i, m := range missing {
					if m != tt.wantMissing[i] {
						t.Errorf("substituteEnvVars() missing[%d] = %v, want %v", i, m, tt.wantMissing[i])
					}
				}
			}
		})
	}
}
