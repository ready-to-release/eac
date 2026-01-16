//go:build L1 && ov
// +build L1,ov

// File: go/eac/commands/impl/init/init_test.go
package init

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateDirectoryStructure(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create .r2r/eac directory successfully",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Test directory structure creation
			err := createDirectoryStructure(tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("createDirectoryStructure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify .r2r/eac directory was created
			eacDir := filepath.Join(tmpDir, ".r2r", "eac")
			if _, err := os.Stat(eacDir); os.IsNotExist(err) {
				t.Errorf(".r2r/eac directory was not created")
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
				"ai:",
				"provider: claude-api",
				"model: claude-3-haiku-20240307",
				"endpoint: https://api.anthropic.com/v1",
				"api_key: ${ANTHROPIC_API_KEY}",
				"git:",
				"token: ${GIT_TOKEN}",
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
				"ai:",
				"provider: openai",
				"model: gpt-4-turbo",
				"api_key: ${OPENAI_API_KEY}",
				"git:",
			},
		},
		{
			name: "write gemini config",
			config: &agentConfig{
				providerName: "gemini",
				envVarName:   "GOOGLE_API_KEY",
				model:        "gemini-1.5-pro",
				endpoint:     "https://generativelanguage.googleapis.com",
			},
			wantErr: false,
			wantContains: []string{
				"ai:",
				"provider: gemini",
				"model: gemini-1.5-pro",
				"api_key: ${GOOGLE_API_KEY}",
				"git:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			// Create .r2r/eac directory
			eacDir := filepath.Join(tmpDir, ".r2r", "eac")
			if err := os.MkdirAll(eacDir, 0755); err != nil {
				t.Fatalf("failed to create .r2r/eac directory: %v", err)
			}

			// Create empty token config since we're just testing team config generation
			tokens := &tokenConfig{}
			configPath, err := writeConfig(tmpDir, tt.config, tokens)
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
				if !strings.Contains(contentStr, want) {
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
		{
			name:     "claude-cli not supported in init (use importer.ps1)",
			provider: "claude-cli",
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

func TestGenerateBooksYML(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".r2r", "eac")
	if err := os.MkdirAll(eacDir, 0755); err != nil {
		t.Fatalf("failed to create .r2r/eac directory: %v", err)
	}

	err := generateBooksYML(eacDir)
	if err != nil {
		t.Errorf("generateBooksYML() error = %v", err)
		return
	}

	// Verify file was created
	booksPath := filepath.Join(eacDir, "books.yml")
	content, err := os.ReadFile(booksPath)
	if err != nil {
		t.Fatalf("failed to read books.yml: %v", err)
	}

	// Verify content
	contentStr := string(content)
	expectedContents := []string{
		"# Documentation Books Configuration",
		"# Generated by: r2r eac init",
		"books: []",
	}
	for _, want := range expectedContents {
		if !strings.Contains(contentStr, want) {
			t.Errorf("books.yml missing expected content: %q\nGot:\n%s", want, contentStr)
		}
	}
}

func TestGenerateEnvironmentsYML(t *testing.T) {
	tmpDir := t.TempDir()
	eacDir := filepath.Join(tmpDir, ".r2r", "eac")
	if err := os.MkdirAll(eacDir, 0755); err != nil {
		t.Fatalf("failed to create .r2r/eac directory: %v", err)
	}

	err := generateEnvironmentsYML(eacDir)
	if err != nil {
		t.Errorf("generateEnvironmentsYML() error = %v", err)
		return
	}

	// Verify file was created
	envPath := filepath.Join(eacDir, "environments.yml")
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("failed to read environments.yml: %v", err)
	}

	// Verify content
	contentStr := string(content)
	expectedContents := []string{
		"# Test Environments Configuration",
		"# Generated by: r2r eac init",
		"environments: []",
		"L0 - Unit tests",
		"L1 - Integration tests",
	}
	for _, want := range expectedContents {
		if !strings.Contains(contentStr, want) {
			t.Errorf("environments.yml missing expected content: %q\nGot:\n%s", want, contentStr)
		}
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test non-existent file
	if fileExists(filepath.Join(tmpDir, "nonexistent.txt")) {
		t.Error("fileExists() should return false for non-existent file")
	}

	// Create a file and test
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if !fileExists(testFile) {
		t.Error("fileExists() should return true for existing file")
	}
}
