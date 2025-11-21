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

			// Create mock contracts/ai directory structure
			mockRepoRoot := filepath.Join(tmpDir, "mock-repo")
			mockCommitDir := filepath.Join(mockRepoRoot, "contracts", "ai", "commit-message", "0.1.0")
			mockSpecsDir := filepath.Join(mockRepoRoot, "contracts", "ai", "specifications", "0.1.0")
			if err := os.MkdirAll(mockCommitDir, 0755); err != nil {
				t.Fatalf("failed to create mock commit contracts directory: %v", err)
			}
			if err := os.MkdirAll(mockSpecsDir, 0755); err != nil {
				t.Fatalf("failed to create mock specs contracts directory: %v", err)
			}

			// Create mock contract files
			contractContent := "version: 0.1.0\nname: test"
			antiCorruptionContent := "version: 0.1.0\nname: test-ac"
			topLevelPrompt := "# Mock top-level prompt"
			modulePrompt := "# Mock module prompt"
			specPrompt := "# Mock spec prompt"

			// Write commit-message contracts
			if err := os.WriteFile(filepath.Join(mockCommitDir, "contract.yml"), []byte(contractContent), 0644); err != nil {
				t.Fatalf("failed to create contract.yml: %v", err)
			}
			if err := os.WriteFile(filepath.Join(mockCommitDir, "anti-corruption.yml"), []byte(antiCorruptionContent), 0644); err != nil {
				t.Fatalf("failed to create anti-corruption.yml: %v", err)
			}
			if err := os.WriteFile(filepath.Join(mockCommitDir, "top-level.md"), []byte(topLevelPrompt), 0644); err != nil {
				t.Fatalf("failed to create top-level.md: %v", err)
			}
			if err := os.WriteFile(filepath.Join(mockCommitDir, "module.md"), []byte(modulePrompt), 0644); err != nil {
				t.Fatalf("failed to create module.md: %v", err)
			}

			// Write specifications contracts
			if err := os.WriteFile(filepath.Join(mockSpecsDir, "contract.yml"), []byte(contractContent), 0644); err != nil {
				t.Fatalf("failed to create specs contract.yml: %v", err)
			}
			if err := os.WriteFile(filepath.Join(mockSpecsDir, "anti-corruption.yml"), []byte(antiCorruptionContent), 0644); err != nil {
				t.Fatalf("failed to create specs anti-corruption.yml: %v", err)
			}
			if err := os.WriteFile(filepath.Join(mockSpecsDir, "specification.md"), []byte(specPrompt), 0644); err != nil {
				t.Fatalf("failed to create specification.md: %v", err)
			}

			// Create mock source prompts directory for specs (correct path: specs/create/prompts)
			mockSpecsPromptsDir := filepath.Join(mockRepoRoot, "src", "commands", "impl", "specs", "create", "prompts")
			if err := os.MkdirAll(mockSpecsPromptsDir, 0755); err != nil {
				t.Fatalf("failed to create mock specs prompts directory: %v", err)
			}

			// Create mock specs prompt file
			specificationContent := "# Mock specification prompt"
			if err := os.WriteFile(filepath.Join(mockSpecsPromptsDir, "specification.md"), []byte(specificationContent), 0644); err != nil {
				t.Fatalf("failed to create mock specification.md: %v", err)
			}

			// Test directory structure creation
			targetDir := filepath.Join(tmpDir, "target")
			err := createDirectoryStructure(targetDir, mockRepoRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("createDirectoryStructure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify .r2r directory was created
			r2rDir := filepath.Join(targetDir, ".r2r")
			if _, err := os.Stat(r2rDir); os.IsNotExist(err) {
				t.Errorf(".r2r directory was not created")
			}

			// Verify .r2r/contracts/ai/commit-message directory was created
			contractsDir := filepath.Join(targetDir, ".r2r", "contracts", "ai", "commit-message", "0.1.0")
			if _, err := os.Stat(contractsDir); os.IsNotExist(err) {
				t.Errorf(".r2r/contracts/ai/commit-message/0.1.0 directory was not created")
			}

			// Verify contracts were copied
			copiedContract := filepath.Join(contractsDir, "contract.yml")
			if _, err := os.Stat(copiedContract); os.IsNotExist(err) {
				t.Errorf("contract.yml was not copied")
			}
			copiedTopLevel := filepath.Join(contractsDir, "top-level.md")
			if _, err := os.Stat(copiedTopLevel); os.IsNotExist(err) {
				t.Errorf("top-level.md was not copied")
			}
			copiedModule := filepath.Join(contractsDir, "module.md")
			if _, err := os.Stat(copiedModule); os.IsNotExist(err) {
				t.Errorf("module.md was not copied")
			}

			// Verify specs contracts were copied
			specsDir := filepath.Join(targetDir, ".r2r", "contracts", "ai", "specifications", "0.1.0")
			copiedSpecContract := filepath.Join(specsDir, "contract.yml")
			if _, err := os.Stat(copiedSpecContract); os.IsNotExist(err) {
				t.Errorf("specs contract.yml was not copied")
			}
			copiedSpec := filepath.Join(specsDir, "specification.md")
			if _, err := os.Stat(copiedSpec); os.IsNotExist(err) {
				t.Errorf("specification.md was not copied")
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
				model:        "claude-3-haiku-20240307",
				endpoint:     "",
			},
			wantErr: false,
			wantContains: []string{
				"name: claude-cli",
				"model: claude-3-haiku-20240307",
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
			wantModel:        "claude-3-haiku-20240307",
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
