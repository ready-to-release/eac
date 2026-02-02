//go:build L1 && ov
// +build L1,ov

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/core/paths"
)

// TestLoadPrompt_ThreeTierPriority tests the three-tier priority system for prompt loading
func TestLoadPrompt_ThreeTierPriority(t *testing.T) {
	// Create temporary workspace
	tmpDir, err := os.MkdirTemp("", "prompt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup directory structure
	systemDefaultDir := filepath.Join(tmpDir, paths.TemplatesDir, "ai", "test-type")
	teamOverrideDir := filepath.Join(tmpDir, paths.R2RDir, paths.EACDir, "templates", "ai", "test-type")

	if err := os.MkdirAll(systemDefaultDir, 0755); err != nil {
		t.Fatalf("Failed to create system default dir: %v", err)
	}
	if err := os.MkdirAll(teamOverrideDir, 0755); err != nil {
		t.Fatalf("Failed to create team override dir: %v", err)
	}

	// Create system default prompt
	systemPromptPath := filepath.Join(systemDefaultDir, "test.md")
	systemPromptContent := "# System Default Prompt\nThis is the system default."
	if err := os.WriteFile(systemPromptPath, []byte(systemPromptContent), 0644); err != nil {
		t.Fatalf("Failed to write system default prompt: %v", err)
	}

	// Create team override prompt
	teamPromptPath := filepath.Join(teamOverrideDir, "test.md")
	teamPromptContent := "# Team Override Prompt\nThis is the team override."
	if err := os.WriteFile(teamPromptPath, []byte(teamPromptContent), 0644); err != nil {
		t.Fatalf("Failed to write team override prompt: %v", err)
	}

	// Create custom prompt
	customPromptPath := filepath.Join(tmpDir, "custom-prompt.md")
	customPromptContent := "# Custom Prompt\nThis is a custom prompt."
	if err := os.WriteFile(customPromptPath, []byte(customPromptContent), 0644); err != nil {
		t.Fatalf("Failed to write custom prompt: %v", err)
	}

	// Test 1: System default (when no team override)
	t.Run("SystemDefault", func(t *testing.T) {
		// Remove team override temporarily
		os.Remove(teamPromptPath)
		defer func() {
			// Restore team override for other tests
			os.WriteFile(teamPromptPath, []byte(teamPromptContent), 0644)
		}()

		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		prompt, source, err := loader.LoadPrompt("test.md", "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "system default" {
			t.Errorf("Expected source 'system default', got '%s'", source)
		}
		if prompt != systemPromptContent {
			t.Errorf("Expected system default content, got: %s", prompt)
		}
	})

	// Test 2: Team override (when present, overrides system default)
	t.Run("TeamOverride", func(t *testing.T) {
		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		prompt, source, err := loader.LoadPrompt("test.md", "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "team override" {
			t.Errorf("Expected source 'team override', got '%s'", source)
		}
		if prompt != teamPromptContent {
			t.Errorf("Expected team override content, got: %s", prompt)
		}
	})

	// Test 3: Command flag with absolute path (highest priority)
	t.Run("CommandFlagAbsolute", func(t *testing.T) {
		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		prompt, source, err := loader.LoadPrompt(customPromptPath, "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "command flag" {
			t.Errorf("Expected source 'command flag', got '%s'", source)
		}
		if prompt != customPromptContent {
			t.Errorf("Expected custom prompt content, got: %s", prompt)
		}
	})

	// Test 4: Fallback (when nothing else found)
	t.Run("Fallback", func(t *testing.T) {
		fallbackContent := "# Fallback Prompt\nThis is the fallback."
		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		prompt, source, err := loader.LoadPrompt("nonexistent.md", fallbackContent)
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "embedded fallback" {
			t.Errorf("Expected source 'embedded fallback', got '%s'", source)
		}
		if prompt != fallbackContent {
			t.Errorf("Expected fallback content, got: %s", prompt)
		}
	})

	// Test 5: Error when no fallback and file not found
	t.Run("ErrorNoFallback", func(t *testing.T) {
		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		_, _, err := loader.LoadPrompt("nonexistent.md", "")
		if err == nil {
			t.Errorf("Expected error when prompt not found and no fallback, got nil")
		}
	})
}

// TestLoadPromptWithPriority_CustomPath tests the explicit priority method
func TestLoadPromptWithPriority_CustomPath(t *testing.T) {
	// Create temporary workspace
	tmpDir, err := os.MkdirTemp("", "prompt-priority-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup directory structure
	systemDefaultDir := filepath.Join(tmpDir, paths.TemplatesDir, "ai", "test-type")
	if err := os.MkdirAll(systemDefaultDir, 0755); err != nil {
		t.Fatalf("Failed to create system default dir: %v", err)
	}

	// Create system default prompt
	systemPromptPath := filepath.Join(systemDefaultDir, "test.md")
	systemPromptContent := "# System Default"
	if err := os.WriteFile(systemPromptPath, []byte(systemPromptContent), 0644); err != nil {
		t.Fatalf("Failed to write system default prompt: %v", err)
	}

	// Create custom prompt
	customPromptPath := filepath.Join(tmpDir, "custom.md")
	customPromptContent := "# Custom"
	if err := os.WriteFile(customPromptPath, []byte(customPromptContent), 0644); err != nil {
		t.Fatalf("Failed to write custom prompt: %v", err)
	}

	// Test 1: With custom path (highest priority)
	t.Run("WithCustomPath", func(t *testing.T) {
		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		prompt, source, err := loader.LoadPromptWithPriority("test.md", customPromptPath)
		if err != nil {
			t.Fatalf("LoadPromptWithPriority failed: %v", err)
		}

		if source != "command flag" {
			t.Errorf("Expected source 'command flag', got '%s'", source)
		}
		if prompt != customPromptContent {
			t.Errorf("Expected custom content, got: %s", prompt)
		}
	})

	// Test 2: Without custom path (falls back to system default)
	t.Run("WithoutCustomPath", func(t *testing.T) {
		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		prompt, source, err := loader.LoadPromptWithPriority("test.md", "")
		if err != nil {
			t.Fatalf("LoadPromptWithPriority failed: %v", err)
		}

		if source != "system default" {
			t.Errorf("Expected source 'system default', got '%s'", source)
		}
		if prompt != systemPromptContent {
			t.Errorf("Expected system default content, got: %s", prompt)
		}
	})

	// Test 3: Relative custom path (converted to absolute)
	t.Run("RelativeCustomPath", func(t *testing.T) {
		// Create a relative path file
		relativePromptContent := "# Relative Custom"
		relativePromptPath := "relative-custom.md"
		fullRelativePath := filepath.Join(tmpDir, relativePromptPath)
		if err := os.WriteFile(fullRelativePath, []byte(relativePromptContent), 0644); err != nil {
			t.Fatalf("Failed to write relative custom prompt: %v", err)
		}

		loader := NewContractLoader(tmpDir, "ai/test-type", "")
		prompt, source, err := loader.LoadPromptWithPriority("test.md", relativePromptPath)
		if err != nil {
			t.Fatalf("LoadPromptWithPriority failed: %v", err)
		}

		if source != "command flag" {
			t.Errorf("Expected source 'command flag', got '%s'", source)
		}
		if prompt != relativePromptContent {
			t.Errorf("Expected relative custom content, got: %s", prompt)
		}
	})
}

// TestLoadPrompt_RealWorldScenarios tests real-world usage patterns
func TestLoadPrompt_RealWorldScenarios(t *testing.T) {
	// Create temporary workspace
	tmpDir, err := os.MkdirTemp("", "prompt-real-world-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Scenario 1: New project, only system defaults exist
	t.Run("NewProject", func(t *testing.T) {
		// Create only system default
		systemDir := filepath.Join(tmpDir, "scenario1", paths.TemplatesDir, "ai", "specs")
		if err := os.MkdirAll(systemDir, 0755); err != nil {
			t.Fatalf("Failed to create system dir: %v", err)
		}

		systemContent := "# Specs Prompt"
		if err := os.WriteFile(filepath.Join(systemDir, "specs.md"), []byte(systemContent), 0644); err != nil {
			t.Fatalf("Failed to write system prompt: %v", err)
		}

		loader := NewContractLoader(filepath.Join(tmpDir, "scenario1"), "ai/specs", "")
		prompt, source, err := loader.LoadPrompt("specs.md", "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "system default" {
			t.Errorf("Expected 'system default', got '%s'", source)
		}
		if prompt != systemContent {
			t.Errorf("Expected system content")
		}
	})

	// Scenario 2: Team has customized prompts
	t.Run("TeamCustomization", func(t *testing.T) {
		baseDir := filepath.Join(tmpDir, "scenario2")

		// Create system default
		systemDir := filepath.Join(baseDir, paths.TemplatesDir, "ai", "commit")
		if err := os.MkdirAll(systemDir, 0755); err != nil {
			t.Fatalf("Failed to create system dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(systemDir, "module.md"), []byte("# System"), 0644); err != nil {
			t.Fatalf("Failed to write system prompt: %v", err)
		}

		// Create team override
		teamDir := filepath.Join(baseDir, paths.R2RDir, paths.EACDir, "templates", "ai", "commit")
		if err := os.MkdirAll(teamDir, 0755); err != nil {
			t.Fatalf("Failed to create team dir: %v", err)
		}
		teamContent := "# Team Custom\nUse JIRA-XXX format"
		if err := os.WriteFile(filepath.Join(teamDir, "module.md"), []byte(teamContent), 0644); err != nil {
			t.Fatalf("Failed to write team prompt: %v", err)
		}

		loader := NewContractLoader(baseDir, "ai/commit", "")
		prompt, source, err := loader.LoadPrompt("module.md", "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "team override" {
			t.Errorf("Expected 'team override', got '%s'", source)
		}
		if prompt != teamContent {
			t.Errorf("Expected team content")
		}
	})

	// Scenario 3: Developer testing new prompt locally
	t.Run("DeveloperExperiment", func(t *testing.T) {
		baseDir := filepath.Join(tmpDir, "scenario3")

		// Create system default
		systemDir := filepath.Join(baseDir, paths.TemplatesDir, "ai", "design")
		if err := os.MkdirAll(systemDir, 0755); err != nil {
			t.Fatalf("Failed to create system dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(systemDir, "design.md"), []byte("# System"), 0644); err != nil {
			t.Fatalf("Failed to write system prompt: %v", err)
		}

		// Create team override
		teamDir := filepath.Join(baseDir, paths.R2RDir, paths.EACDir, "templates", "ai", "design")
		if err := os.MkdirAll(teamDir, 0755); err != nil {
			t.Fatalf("Failed to create team dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(teamDir, "design.md"), []byte("# Team"), 0644); err != nil {
			t.Fatalf("Failed to write team prompt: %v", err)
		}

		// Developer's experimental prompt
		experimentPath := filepath.Join(baseDir, "experiment.md")
		experimentContent := "# Experiment\nTrying new C4 format"
		if err := os.WriteFile(experimentPath, []byte(experimentContent), 0644); err != nil {
			t.Fatalf("Failed to write experiment prompt: %v", err)
		}

		// Use command flag to override everything
		loader := NewContractLoader(baseDir, "ai/design", "")
		prompt, source, err := loader.LoadPrompt(experimentPath, "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "command flag" {
			t.Errorf("Expected 'command flag', got '%s'", source)
		}
		if prompt != experimentContent {
			t.Errorf("Expected experiment content")
		}
	})
}

// TestLoadPrompt_Convention tests the convention-based naming system
func TestLoadPrompt_Convention(t *testing.T) {
	// Create temporary workspace
	tmpDir, err := os.MkdirTemp("", "convention-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test 1: Empty string uses type name
	t.Run("EmptyStringUsesTypeName", func(t *testing.T) {
		// Setup: ai/specs type should use specs.md
		systemDir := filepath.Join(tmpDir, paths.TemplatesDir, "ai", "specs")
		if err := os.MkdirAll(systemDir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}

		content := "# Specs Convention Test"
		if err := os.WriteFile(filepath.Join(systemDir, "specs.md"), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		loader := NewContractLoader(tmpDir, "ai/specs", "")
		prompt, source, err := loader.LoadPrompt("", "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if source != "system default" {
			t.Errorf("Expected 'system default', got '%s'", source)
		}
		if prompt != content {
			t.Errorf("Expected convention content, got: %s", prompt)
		}
	})

	// Test 2: Hyphenated type names
	t.Run("HyphenatedTypeName", func(t *testing.T) {
		// Setup: ai/risk-profile type should use risk-profile.md
		systemDir := filepath.Join(tmpDir, paths.TemplatesDir, "ai", "risk-profile")
		if err := os.MkdirAll(systemDir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}

		content := "# Risk Profile Convention"
		if err := os.WriteFile(filepath.Join(systemDir, "risk-profile.md"), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		loader := NewContractLoader(tmpDir, "ai/risk-profile", "")
		prompt, _, err := loader.LoadPrompt("", "")
		if err != nil {
			t.Fatalf("LoadPrompt failed: %v", err)
		}

		if prompt != content {
			t.Errorf("Expected convention content")
		}
	})

	// Test 3: Extension normalization
	t.Run("ExtensionNormalization", func(t *testing.T) {
		systemDir := filepath.Join(tmpDir, paths.TemplatesDir, "ai", "design")
		if err := os.MkdirAll(systemDir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}

		content := "# Design Prompt"
		if err := os.WriteFile(filepath.Join(systemDir, "design.md"), []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		loader := NewContractLoader(tmpDir, "ai/design", "")

		// Test with no extension (should add .md)
		prompt1, _, err := loader.LoadPrompt("design", "")
		if err != nil {
			t.Fatalf("LoadPrompt without extension failed: %v", err)
		}
		if prompt1 != content {
			t.Errorf("Extension normalization failed")
		}

		// Test with extension (should work)
		prompt2, _, err := loader.LoadPrompt("design.md", "")
		if err != nil {
			t.Fatalf("LoadPrompt with extension failed: %v", err)
		}
		if prompt2 != content {
			t.Errorf("Extension explicit failed")
		}

		// Test empty string (should use convention)
		prompt3, _, err := loader.LoadPrompt("", "")
		if err != nil {
			t.Fatalf("LoadPrompt with empty string failed: %v", err)
		}
		if prompt3 != content {
			t.Errorf("Convention failed")
		}

		// All three should return the same content
		if prompt1 != prompt2 || prompt2 != prompt3 {
			t.Errorf("Inconsistent prompt loading")
		}
	})

	// Test 4: Variant prompts
	t.Run("VariantPrompts", func(t *testing.T) {
		systemDir := filepath.Join(tmpDir, paths.TemplatesDir, "ai", "commit-message")
		if err := os.MkdirAll(systemDir, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}

		// Create variant prompts
		moduleContent := "# Module Prompt"
		squashContent := "# Squash Prompt"
		if err := os.WriteFile(filepath.Join(systemDir, "module.md"), []byte(moduleContent), 0644); err != nil {
			t.Fatalf("Failed to write module.md: %v", err)
		}
		if err := os.WriteFile(filepath.Join(systemDir, "squash.md"), []byte(squashContent), 0644); err != nil {
			t.Fatalf("Failed to write squash.md: %v", err)
		}

		loader := NewContractLoader(tmpDir, "ai/commit-message", "")

		// Load variant without extension
		modulePrompt, _, err := loader.LoadPrompt("module", "")
		if err != nil {
			t.Fatalf("LoadPrompt module failed: %v", err)
		}
		if modulePrompt != moduleContent {
			t.Errorf("Expected module content")
		}

		// Load another variant
		squashPrompt, _, err := loader.LoadPrompt("squash", "")
		if err != nil {
			t.Fatalf("LoadPrompt squash failed: %v", err)
		}
		if squashPrompt != squashContent {
			t.Errorf("Expected squash content")
		}
	})
}

// TestConvention_TypeNameExtraction tests the getDefaultPromptName helper
func TestConvention_TypeNameExtraction(t *testing.T) {
	tests := []struct {
		typeName string
		expected string
	}{
		{"ai/specs", "specs.md"},
		{"ai/design", "design.md"},
		{"ai/commit-message", "commit-message.md"},
		{"ai/risk-profile", "risk-profile.md"},
		{"ai/risk-access", "risk-access.md"},
		{"specs", "specs.md"},     // No "ai/" prefix
		{"commit", "commit.md"},   // Simple name
		{"a/b/c/deep", "deep.md"}, // Deep path
	}

	tmpDir, err := os.MkdirTemp("", "extraction-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			loader := NewContractLoader(tmpDir, tt.typeName, "")
			result := loader.getDefaultPromptName()
			if result != tt.expected {
				t.Errorf("For type '%s': expected '%s', got '%s'", tt.typeName, tt.expected, result)
			}
		})
	}
}

// TestConvention_BackwardCompatibility ensures old explicit names still work
func TestConvention_BackwardCompatibility(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "compat-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup
	systemDir := filepath.Join(tmpDir, paths.TemplatesDir, "ai", "specs")
	if err := os.MkdirAll(systemDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	content := "# Test Content"
	if err := os.WriteFile(filepath.Join(systemDir, "specs.md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	loader := NewContractLoader(tmpDir, "ai/specs", "")

	// Old way: explicit "specs.md"
	prompt1, _, err := loader.LoadPrompt("specs.md", "")
	if err != nil {
		t.Fatalf("Old way failed: %v", err)
	}

	// New way: empty string
	prompt2, _, err := loader.LoadPrompt("", "")
	if err != nil {
		t.Fatalf("New way failed: %v", err)
	}

	// Middle way: no extension
	prompt3, _, err := loader.LoadPrompt("specs", "")
	if err != nil {
		t.Fatalf("Middle way failed: %v", err)
	}

	// All should return the same content
	if prompt1 != content || prompt2 != content || prompt3 != content {
		t.Errorf("Backward compatibility broken: got different content")
	}
}
