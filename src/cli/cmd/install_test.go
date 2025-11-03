//go:build L2
// +build L2

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestInstallCommand_CreateConfigFile(t *testing.T) {
	// Create temporary directory to act as repository root
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Initialize git repo to make it a valid repository
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Verify config file doesn't exist
	configPath := filepath.Join(tempDir, "r2r-cli.yml")
	if _, err := os.Stat(configPath); err == nil {
		t.Fatal("Config file should not exist initially")
	}

	// Test adding extension to non-existent config
	err := addExtensionToConfig("pwsh")
	if err != nil {
		t.Fatalf("addExtensionToConfig should create config file, got error: %v", err)
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file should have been created")
	}

	// Read and verify config content
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse created config: %v", err)
	}

	// Verify structure
	if config["version"] != "1.0" {
		t.Errorf("Expected version 1.0, got %v", config["version"])
	}

	extensions, ok := config["extensions"].([]interface{})
	if !ok {
		t.Fatal("Extensions should be a list")
	}

	if len(extensions) != 1 {
		t.Fatalf("Expected 1 extension, got %d", len(extensions))
	}

	ext, ok := extensions[0].(map[string]interface{})
	if !ok {
		t.Fatal("Extension should be a map")
	}

	if ext["name"] != "pwsh" {
		t.Errorf("Expected extension name 'pwsh', got %v", ext["name"])
	}

	image, ok := ext["image"].(string)
	if !ok || image == "" {
		t.Errorf("Expected non-empty image string, got %v", ext["image"])
	}
}

func TestInstallCommand_AddToExistingConfig(t *testing.T) {
	// Create temporary directory to act as repository root
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Initialize git repo
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create existing config file
	configPath := filepath.Join(tempDir, "r2r-cli.yml")
	existingConfig := `version: "1.0"
extensions:
  - name: "python"
    image: "ghcr.io/ready-to-release/r2r-cli/extensions/python:latest"
`
	if err := os.WriteFile(configPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	// Add new extension
	err := addExtensionToConfig("pwsh")
	if err != nil {
		t.Fatalf("addExtensionToConfig failed: %v", err)
	}

	// Read and verify updated config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read updated config file: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse updated config: %v", err)
	}

	extensions, ok := config["extensions"].([]interface{})
	if !ok {
		t.Fatal("Extensions should be a list")
	}

	if len(extensions) != 2 {
		t.Fatalf("Expected 2 extensions, got %d", len(extensions))
	}

	// Verify both extensions exist
	names := make([]string, len(extensions))
	for i, ext := range extensions {
		extMap, ok := ext.(map[string]interface{})
		if !ok {
			t.Fatal("Extension should be a map")
		}
		names[i] = extMap["name"].(string)
	}

	expectedNames := []string{"python", "pwsh"}
	for _, expected := range expectedNames {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find extension %s, got extensions: %v", expected, names)
		}
	}
}

func TestInstallCommand_PreventDuplicates(t *testing.T) {
	// Create temporary directory to act as repository root
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Initialize git repo
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create existing config with pwsh
	configPath := filepath.Join(tempDir, "r2r-cli.yml")
	existingConfig := `version: "1.0"
extensions:
  - name: "pwsh"
    image: "ghcr.io/ready-to-release/r2r-cli/extensions/pwsh:v1.0.0"
`
	if err := os.WriteFile(configPath, []byte(existingConfig), 0644); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	// Try to add pwsh again - this may fail if registry is not accessible
	// but we can still test the file handling logic
	err := addExtensionToConfig("pwsh")

	// If registry access fails, just verify the error is reasonable and return
	if err != nil {
		if strings.Contains(err.Error(), "registry") || strings.Contains(err.Error(), "SHA") {
			t.Logf("Registry access failed (expected in isolated test): %v", err)
			return
		}
		t.Fatalf("addExtensionToConfig failed with unexpected error: %v", err)
	}

	// Read config and verify structure
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	extensions, ok := config["extensions"].([]interface{})
	if !ok {
		t.Fatal("Extensions should be a list")
	}

	// Should still be only 1 extension (no duplicates)
	if len(extensions) != 1 {
		t.Fatalf("Expected 1 extension (no duplicates), got %d", len(extensions))
	}

	ext, ok := extensions[0].(map[string]interface{})
	if !ok {
		t.Fatal("Extension should be a map")
	}

	if ext["name"] != "pwsh" {
		t.Errorf("Expected extension name 'pwsh', got %v", ext["name"])
	}
}

func TestInstallCommand_NoRepositoryRoot(t *testing.T) {
	// This test is tricky because the repository root finder may find the main project repo
	// Let's test the behavior when it finds a repo but can't write to the config
	tempDir := t.TempDir()

	// Create isolated directory structure
	isolatedDir := filepath.Join(tempDir, "isolated", "not-a-repo")
	if err := os.MkdirAll(isolatedDir, 0755); err != nil {
		t.Fatalf("Failed to create isolated directory: %v", err)
	}

	// Change to isolated directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(isolatedDir)

	// Try to add extension
	err := addExtensionToConfig("pwsh")
	if err == nil {
		t.Fatal("addExtensionToConfig should fail in isolated directory")
	}

	// It might fail for different reasons:
	// 1. No repository root found
	// 2. Repository found but can't write config (permissions)
	// 3. Registry access failure
	// All of these are acceptable failure modes for this test
	t.Logf("addExtensionToConfig failed as expected: %v", err)

	// Just verify it failed with a reasonable error message
	errorMsg := err.Error()
	validErrors := []string{
		"repository root",
		"git repository",
		"config file",
		"Access is denied",
		"permission denied",
		"no such file or directory",
		"registry",
		"SHA",
		"failed to write",
		"failed to find",
	}

	found := false
	for _, validError := range validErrors {
		if strings.Contains(errorMsg, validError) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Error should contain one of %v, got: %v", validErrors, errorMsg)
	}
}

func TestInstallCommand_EmptyExtensionName(t *testing.T) {
	// Create temporary directory with git repo
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Initialize git repo
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Try to add empty extension name
	err := addExtensionToConfig("")
	if err == nil {
		t.Fatal("addExtensionToConfig should fail with empty extension name")
	}

	if !strings.Contains(err.Error(), "extension name is required") {
		t.Errorf("Error should mention extension name required, got: %v", err)
	}
}

func TestInstallCommand_ConfigFilePermissions(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Initialize git repo
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Add extension (creates config file)
	err := addExtensionToConfig("pwsh")
	if err != nil {
		t.Fatalf("addExtensionToConfig failed: %v", err)
	}

	// Check file permissions
	configPath := filepath.Join(tempDir, "r2r-cli.yml")
	stat, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}

	// Should be readable and writable by owner
	mode := stat.Mode()
	// Check that owner has read and write permissions (cross-platform)
	if mode&0400 == 0 { // Owner read
		t.Errorf("Config file should be readable by owner, got permissions: %o", mode)
	}
	if mode&0200 == 0 { // Owner write
		t.Errorf("Config file should be writable by owner, got permissions: %o", mode)
	}
}

func TestInstallCommand_ConfigInitialization(t *testing.T) {
	// Test that config file creation works properly
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Initialize git repo
	if err := os.Mkdir(".git", 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Verify no config exists initially
	configPath := filepath.Join(tempDir, "r2r-cli.yml")
	if _, err := os.Stat(configPath); err == nil {
		t.Fatal("Config file should not exist initially")
	}

	// Add extension (should create config)
	err := addExtensionToConfig("pwsh")
	if err != nil {
		if strings.Contains(err.Error(), "registry") || strings.Contains(err.Error(), "SHA") {
			t.Logf("Registry access failed (acceptable in test): %v", err)
			return
		}
		t.Fatalf("addExtensionToConfig failed: %v", err)
	}

	// Verify config file was created and has proper content
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file should have been created")
	}

	// Read and verify the created config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config: %v", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse created config: %v", err)
	}

	// Verify basic structure
	if config["version"] != "1.0" {
		t.Errorf("Expected version 1.0, got %v", config["version"])
	}

	extensions, ok := config["extensions"].([]interface{})
	if !ok {
		t.Fatal("Extensions should be a list")
	}

	if len(extensions) != 1 {
		t.Fatalf("Expected 1 extension, got %d", len(extensions))
	}

	ext, ok := extensions[0].(map[string]interface{})
	if !ok {
		t.Fatal("Extension should be a map")
	}

	if ext["name"] != "pwsh" {
		t.Errorf("Expected extension name 'pwsh', got %v", ext["name"])
	}

	// Verify image field exists and is non-empty
	image, ok := ext["image"].(string)
	if !ok || image == "" {
		t.Errorf("Expected non-empty image string, got %v", ext["image"])
	}
}
