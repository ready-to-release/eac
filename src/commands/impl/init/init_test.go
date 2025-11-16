// File: src/commands/impl/init/init_test.go
package init

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirectoryStructure(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create directory structure successfully",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			err := createDirectoryStructure(tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("createDirectoryStructure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify .r2r directory was created
			r2rDir := filepath.Join(tmpDir, ".r2r")
			if _, err := os.Stat(r2rDir); os.IsNotExist(err) {
				t.Errorf(".r2r directory was not created")
			}

			// Verify .r2r/logs directory was created
			logsDir := filepath.Join(tmpDir, ".r2r", "logs")
			if _, err := os.Stat(logsDir); os.IsNotExist(err) {
				t.Errorf(".r2r/logs directory was not created")
			}
		})
	}
}

func TestWriteAgentConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       *agentConfig
		wantErr      bool
		wantContains []string
	}{
		{
			name: "write claude-api config",
			config: &agentConfig{
				providerName: "claude-api",
				envVarName:   "ANTHROPIC_API_KEY",
				model:        "claude-3-haiku-20240307",
				endpoint:     "https://api.anthropic.com/v1",
			},
			wantErr: false,
			wantContains: []string{
				"name: claude-api",
				"model: claude-3-haiku-20240307",
				"endpoint: https://api.anthropic.com/v1",
				"api_key: ${ANTHROPIC_API_KEY}",
			},
		},
		{
			name: "write claude-cli config",
			config: &agentConfig{
				providerName: "claude-cli",
				envVarName:   "",
				model:        "sonnet",
				endpoint:     "",
			},
			wantErr: false,
			wantContains: []string{
				"name: claude-cli",
				"model: sonnet",
			},
		},
		{
			name: "write openai config",
			config: &agentConfig{
				providerName: "openai",
				envVarName:   "OPENAI_API_KEY",
				model:        "gpt-4-turbo",
				endpoint:     "https://api.openai.com/v1",
			},
			wantErr: false,
			wantContains: []string{
				"name: openai",
				"model: gpt-4-turbo",
				"api_key: ${OPENAI_API_KEY}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "agent-config.yml")

			err := writeAgentConfig(configPath, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeAgentConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Read the file and verify contents
			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("failed to read config file: %v", err)
			}

			contentStr := string(content)
			for _, want := range tt.wantContains {
				if !contains(contentStr, want) {
					t.Errorf("config file missing expected content: %q\nGot:\n%s", want, contentStr)
				}
			}
		})
	}
}

func TestConfigureProvider(t *testing.T) {
	tests := []struct {
		name             string
		provider         string
		wantProviderName string
		wantModel        string
		wantEnvVar       string
		wantErr          bool
	}{
		{
			name:             "configure claude-api",
			provider:         "claude-api",
			wantProviderName: "claude-api",
			wantModel:        "claude-3-haiku-20240307",
			wantEnvVar:       "ANTHROPIC_API_KEY",
			wantErr:          false,
		},
		{
			name:             "configure claude-cli",
			provider:         "claude-cli",
			wantProviderName: "claude-cli",
			wantModel:        "sonnet",
			wantEnvVar:       "",
			wantErr:          false,
		},
		{
			name:             "configure openai",
			provider:         "openai",
			wantProviderName: "openai",
			wantModel:        "gpt-4-turbo",
			wantEnvVar:       "OPENAI_API_KEY",
			wantErr:          false,
		},
		{
			name:             "configure gemini",
			provider:         "gemini",
			wantProviderName: "gemini",
			wantModel:        "gemini-1.5-pro",
			wantEnvVar:       "GOOGLE_API_KEY",
			wantErr:          false,
		},
		{
			name:     "invalid provider returns error",
			provider: "invalid-provider",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &agentConfig{}
			err := configureProvider(config, tt.provider)

			if (err != nil) != tt.wantErr {
				t.Errorf("configureProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if config.providerName != tt.wantProviderName {
				t.Errorf("providerName = %v, want %v", config.providerName, tt.wantProviderName)
			}
			if config.model != tt.wantModel {
				t.Errorf("model = %v, want %v", config.model, tt.wantModel)
			}
			if config.envVarName != tt.wantEnvVar {
				t.Errorf("envVarName = %v, want %v", config.envVarName, tt.wantEnvVar)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
