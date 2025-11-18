package contract

import (
	"strings"
	"testing"
)

func TestBuildPromptWithTemplate(t *testing.T) {
	// Setup test data
	contract := &SpecContract{
		Version: "0.1.0",
		Name:    "Test Contract",
		RawData: map[string]interface{}{
			"version": "0.1.0",
			"name":    "Test Contract",
			"rules": []string{
				"Rule 1",
				"Rule 2",
			},
		},
	}

	antiCorruption := &AntiCorruptionRules{
		Version: "0.1.0",
		Name:    "Test Rules",
		RawData: map[string]interface{}{
			"version": "0.1.0",
			"forbidden_prefixes": []string{
				"Here is",
				"Let me",
			},
		},
	}

	customData := map[string]string{
		"UserInput": "Create a feature for user authentication",
		"TagsSpec":  "- @ov\n- @iv\n- @pv",
	}

	// Test template
	template := `# Generate Output

## Contract
{{.Contract}}

## Custom Input
{{.Custom.UserInput}}

## Tags
{{.Custom.TagsSpec}}

## Version
Contract Version: {{.ContractRaw.Version}}
`

	// Execute
	result, err := BuildPromptWithTemplate(template, contract, antiCorruption, customData)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate failed: %v", err)
	}

	// Verify output
	if !strings.Contains(result, "Contract Version: 0.1.0") {
		t.Error("Template did not render ContractRaw.Version")
	}

	if !strings.Contains(result, "Create a feature for user authentication") {
		t.Error("Template did not render Custom.UserInput")
	}

	if !strings.Contains(result, "- @ov") {
		t.Error("Template did not render Custom.TagsSpec")
	}

	if !strings.Contains(result, "name: Test Contract") {
		t.Error("Template did not render Contract YAML")
	}
}

func TestBuildPromptWithTemplate_MissingFields(t *testing.T) {
	// Test with nil contract and anti-corruption
	template := `Contract: {{.Contract}}
Custom: {{.Custom.Key1}}`

	customData := map[string]string{
		"Key1": "Value1",
	}

	result, err := BuildPromptWithTemplate(template, nil, nil, customData)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate failed: %v", err)
	}

	if !strings.Contains(result, "Value1") {
		t.Error("Template did not render custom data")
	}

	// Contract should be empty string when nil
	if strings.Contains(result, "Contract: version:") {
		t.Error("Template rendered contract when it should be empty")
	}
}

func TestBuildPromptWithTemplate_ConditionalRendering(t *testing.T) {
	template := `{{if .Custom.Module}}
Module: {{.Custom.Module}}
{{else}}
No module specified
{{end}}`

	// Test with module
	customData := map[string]string{
		"Module": "src-cli",
	}

	result, err := BuildPromptWithTemplate(template, nil, nil, customData)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate failed: %v", err)
	}

	if !strings.Contains(result, "Module: src-cli") {
		t.Error("Conditional rendering failed with module")
	}

	// Test without module
	result, err = BuildPromptWithTemplate(template, nil, nil, nil)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate failed: %v", err)
	}

	if !strings.Contains(result, "No module specified") {
		t.Error("Conditional rendering failed without module")
	}
}

func TestBuildPromptWithTemplate_InvalidTemplate(t *testing.T) {
	// Invalid template syntax
	template := `{{.Invalid syntax}}`

	_, err := BuildPromptWithTemplate(template, nil, nil, nil)
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}

func TestBuildPromptWithTemplate_TemplateInjection(t *testing.T) {
	// Test that user input containing template syntax doesn't execute
	contract := &SpecContract{
		Version: "1.0.0",
		Name:    "Test",
		RawData: map[string]interface{}{
			"version": "1.0.0",
		},
	}

	// Malicious custom data with template injection attempt
	customData := map[string]string{
		"UserInput": "{{.ContractRaw.Version}}", // Attempt to inject template
		"Evil":      "{{.Custom.Secret}}",
	}

	template := `User said: {{.Custom.UserInput}}`

	result, err := BuildPromptWithTemplate(template, contract, nil, customData)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate failed: %v", err)
	}

	// The injected template syntax should be rendered as literal text
	if !strings.Contains(result, "{{.ContractRaw.Version}}") {
		t.Error("Template injection was executed! Should be literal text.")
	}

	// Should NOT contain the actual version value
	// (If it does, the injection worked)
	expectedInjected := "User said: 1.0.0"
	if result == expectedInjected {
		t.Error("Template injection succeeded - security issue!")
	}
}

func TestBuildPromptWithTemplate_VeryLargeTemplate(t *testing.T) {
	// Test with a very large template (1MB+)
	largeTemplate := "# Header\n" + strings.Repeat("{{.Custom.Data}}\n", 50000)

	customData := map[string]string{
		"Data": "x",
	}

	result, err := BuildPromptWithTemplate(largeTemplate, nil, nil, customData)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate failed with large template: %v", err)
	}

	// Should successfully handle large template
	if !strings.Contains(result, "# Header") {
		t.Error("Large template processing failed")
	}

	// Verify it actually expanded the template
	expectedOccurrences := 50000
	actualOccurrences := strings.Count(result, "x")
	if actualOccurrences != expectedOccurrences {
		t.Errorf("Expected %d occurrences of 'x', got %d", expectedOccurrences, actualOccurrences)
	}
}

func TestBuildPromptWithTemplate_MissingCustomData(t *testing.T) {
	// Template references custom data that doesn't exist
	template := `Contract: {{.Contract}}
Custom: {{.Custom.NonExistent}}
More: {{.Custom.AlsoMissing}}`

	// No custom data provided
	result, err := BuildPromptWithTemplate(template, nil, nil, nil)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate should handle missing custom data gracefully: %v", err)
	}

	// When Custom is nil (no custom data), Go templates render map access as "<no value>"
	// This is expected behavior - templates handle nil maps gracefully
	if !strings.Contains(result, "Contract:") {
		t.Error("Template should render with empty/no value for missing data")
	}

	// The result should still be a valid string (no panic, no critical error)
	if result == "" {
		t.Error("Template should produce output even with missing custom data")
	}
}

func TestBuildPromptWithTemplate_SpecialCharactersInCustomData(t *testing.T) {
	customData := map[string]string{
		"Special": "Line1\nLine2\tTabbed\r\nWindows\x00Null",
		"Quotes":  `"double" and 'single' quotes`,
		"Unicode": "Emoji: 🚀 Chinese: 中文 Arabic: العربية",
	}

	template := `Special: {{.Custom.Special}}
Quotes: {{.Custom.Quotes}}
Unicode: {{.Custom.Unicode}}`

	result, err := BuildPromptWithTemplate(template, nil, nil, customData)
	if err != nil {
		t.Fatalf("BuildPromptWithTemplate failed: %v", err)
	}

	// Should preserve special characters
	if !strings.Contains(result, "Line1\nLine2") {
		t.Error("Newlines should be preserved")
	}

	if !strings.Contains(result, `"double" and 'single'`) {
		t.Error("Quotes should be preserved")
	}

	if !strings.Contains(result, "🚀") {
		t.Error("Unicode should be preserved")
	}
}
