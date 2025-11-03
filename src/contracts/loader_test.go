package contracts

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	root := "/test/root"
	loader := NewLoader(root)

	if loader.GetWorkspaceRoot() != root {
		t.Errorf("Expected workspace root %s, got %s", root, loader.GetWorkspaceRoot())
	}
}

func TestLoader_LoadYAML(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "contracts-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test YAML file
	testYAML := `
moniker: "test-module"
name: "Test Module"
type: "test-type"
description: "Test description"
root: "test/root"
`
	testFile := filepath.Join(tmpDir, "test.yml")
	if err := ioutil.WriteFile(testFile, []byte(testYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create loader
	loader := NewLoader(tmpDir)

	// Load YAML
	var result BaseContract
	err = loader.LoadYAML("test.yml", &result)
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}

	// Verify fields
	if result.Moniker != "test-module" {
		t.Errorf("Expected moniker 'test-module', got '%s'", result.Moniker)
	}
	if result.Name != "Test Module" {
		t.Errorf("Expected name 'Test Module', got '%s'", result.Name)
	}
	if result.Type != "test-type" {
		t.Errorf("Expected type 'test-type', got '%s'", result.Type)
	}
}

func TestLoader_LoadYAML_NotFound(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "contracts-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	loader := NewLoader(tmpDir)

	var result BaseContract
	err = loader.LoadYAML("nonexistent.yml", &result)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	if !IsNotFound(err) {
		t.Error("Expected IsNotFound to return true")
	}
}

func TestLoader_LoadYAMLPattern(t *testing.T) {
	// Create temporary directory
	tmpDir, err := ioutil.TempDir("", "contracts-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple test files
	files := map[string]string{
		"test1.yml": `moniker: "test1"`,
		"test2.yml": `moniker: "test2"`,
		"test3.yml": `moniker: "test3"`,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}

	// Create loader
	loader := NewLoader(tmpDir)

	// Load all YAML files matching pattern
	loadedFiles := []string{}
	err = loader.LoadYAMLPattern("*.yml", func(relPath string) error {
		loadedFiles = append(loadedFiles, relPath)
		return nil
	})

	if err != nil {
		t.Fatalf("LoadYAMLPattern failed: %v", err)
	}

	if len(loadedFiles) != 3 {
		t.Errorf("Expected 3 files loaded, got %d", len(loadedFiles))
	}
}

func TestLoader_FileExists(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "contracts-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := ioutil.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	loader := NewLoader(tmpDir)

	// Test existing file
	if !loader.FileExists("exists.txt") {
		t.Error("Expected FileExists to return true for existing file")
	}

	// Test non-existing file
	if loader.FileExists("nonexistent.txt") {
		t.Error("Expected FileExists to return false for nonexistent file")
	}
}
