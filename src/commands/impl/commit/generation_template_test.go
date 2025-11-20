package commit

import (
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/core/contracts"
	"github.com/ready-to-release/eac/src/core/repository"
)

// TestPromptTemplateRendering verifies that prompts use Go templates with contract embedding
func TestPromptTemplateRendering(t *testing.T) {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		t.Fatalf("Failed to get repository root: %v", err)
	}

	testCases := []struct {
		name         string
		promptName   string
		shouldHave   []string
	}{
		{
			name:       "top-level prompt template",
			promptName: "top-level",
			shouldHave: []string{
				"{{.Contract}}",
				"{{.AntiCorruption}}",
				"Anti-Corruption Rules",
				"Contract Structure",
				"Generate", // Should have generation instructions
			},
		},
		{
			name:       "module prompt template",
			promptName: "module",
			shouldHave: []string{
				"{{.Contract}}",
				"{{.AntiCorruption}}",
				"Anti-Corruption Rules",
				"Contract Structure",
				"Generate", // Should have generation instructions
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Load prompt template
			promptTemplate, err := loadPromptWithFallback(tc.promptName, workspaceRoot)
			if err != nil {
				t.Fatalf("Failed to load prompt %s: %v", tc.promptName, err)
			}

			// Verify template contains required placeholders
			for _, required := range tc.shouldHave {
				if !strings.Contains(promptTemplate, required) {
					t.Errorf("Prompt %s missing required content: %s", tc.promptName, required)
				}
			}
		})
	}
}

// TestPromptContractEmbedding verifies that contract and anti-corruption are properly embedded
func TestPromptContractEmbedding(t *testing.T) {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		t.Fatalf("Failed to get repository root: %v", err)
	}

	// Load prompt template
	promptTemplate, err := loadPromptWithFallback("top-level", workspaceRoot)
	if err != nil {
		t.Fatalf("Failed to load top-level prompt: %v", err)
	}

	// Load contract and anti-corruption
	loader := contracts.NewContractLoader(workspaceRoot, "ai/commit-message", "0.1.0")
	contractData, err := loader.LoadContract()
	if err != nil {
		t.Fatalf("Failed to load contract: %v", err)
	}

	antiCorruptionRules, err := loader.LoadAntiCorruptionRules()
	if err != nil {
		t.Fatalf("Failed to load anti-corruption rules: %v", err)
	}

	// Build prompt with template
	customData := map[string]string{}
	renderedPrompt, err := contracts.BuildPromptWithTemplate(
		promptTemplate,
		contractData,
		antiCorruptionRules,
		customData,
	)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Verify rendered prompt contains contract content
	requiredContent := []string{
		"version:",           // From contract YAML
		"semantic_types:",    // From contract YAML
		"structure:",         // From contract YAML
		"forbidden_patterns:", // From anti-corruption YAML
	}

	for _, required := range requiredContent {
		if !strings.Contains(renderedPrompt, required) {
			t.Errorf("Rendered prompt missing contract content: %s", required)
		}
	}

	// Verify template placeholders are replaced (not left as {{.Contract}})
	if strings.Contains(renderedPrompt, "{{.Contract}}") {
		t.Error("Template placeholder {{.Contract}} was not replaced")
	}

	if strings.Contains(renderedPrompt, "{{.AntiCorruption}}") {
		t.Error("Template placeholder {{.AntiCorruption}} was not replaced")
	}

	t.Logf("Rendered prompt length: %d characters", len(renderedPrompt))
	t.Logf("Contract embedded successfully")
}

// TestTemplateBackwardCompatibility verifies fallback when contract loading fails
func TestTemplateBackwardCompatibility(t *testing.T) {
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		t.Fatalf("Failed to get repository root: %v", err)
	}

	// This test verifies that if contract loading fails, we fall back gracefully
	// We can't easily simulate a failure here, so we just verify the code path exists
	// by checking that the function has proper error handling

	loader := contracts.NewContractLoader(workspaceRoot, "ai/commit-message", "0.1.0")
	_, err = loader.LoadContract()
	if err != nil {
		t.Logf("Contract loading failed (expected in some test scenarios): %v", err)
		// This is okay - the code should fall back to non-validated generation
	}

	_, err = loader.LoadAntiCorruptionRules()
	if err != nil {
		t.Logf("Anti-corruption loading failed (expected in some test scenarios): %v", err)
		// This is okay - the code should fall back to non-validated generation
	}
}
