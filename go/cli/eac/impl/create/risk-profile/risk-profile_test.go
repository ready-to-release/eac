package riskprofile

import (
	"testing"
)

func TestExtractControlIDs_JSONArray(t *testing.T) {
	response := `["ac-2", "ia-5", "si-10"]`

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"ac-2", "ia-5", "si-10"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d control IDs, got %d", len(expected), len(result))
	}

	for i, id := range expected {
		if result[i] != id {
			t.Errorf("Expected control ID[%d] = %s, got %s", i, id, result[i])
		}
	}
}

func TestExtractControlIDs_MarkdownWrappedJSON(t *testing.T) {
	response := "```json\n[\"ac-2\", \"ia-5\"]\n```"

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"ac-2", "ia-5"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d control IDs, got %d", len(expected), len(result))
	}

	for i, id := range expected {
		if result[i] != id {
			t.Errorf("Expected control ID[%d] = %s, got %s", i, id, result[i])
		}
	}
}

func TestExtractControlIDs_PlainCodeBlock(t *testing.T) {
	response := "```\n[\"ac-2\", \"si-10\"]\n```"

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"ac-2", "si-10"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d control IDs, got %d", len(expected), len(result))
	}
}

func TestExtractControlIDs_JSONObjectWithControlsField(t *testing.T) {
	response := `{"controls": ["ac-2", "ia-5"]}`

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"ac-2", "ia-5"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d control IDs, got %d", len(expected), len(result))
	}
}

func TestExtractControlIDs_JSONObjectWithControlIDsField(t *testing.T) {
	response := `{"control_ids": ["ac-2", "si-10"]}`

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []string{"ac-2", "si-10"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d control IDs, got %d", len(expected), len(result))
	}
}

func TestExtractControlIDs_TextWithControlIDs(t *testing.T) {
	response := "Based on the risk assessment, the following controls are recommended: ac-2, ia-5, and si-10"

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should find all three controls
	expected := []string{"ac-2", "ia-5", "si-10"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d control IDs, got %d: %v", len(expected), len(result), result)
	}
}

func TestExtractControlIDs_EmptyResponse(t *testing.T) {
	response := ""

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty response, got %d IDs: %v", len(result), result)
	}
}

func TestExtractControlIDs_NoControlIDsFound(t *testing.T) {
	response := "This text has no control identifiers in it"

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected empty result when no controls found, got %d IDs: %v", len(result), result)
	}
}

func TestExtractControlIDs_DuplicateControlIDs(t *testing.T) {
	response := `["ac-2", "ia-5", "ac-2", "si-10", "ia-5"]`

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should deduplicate
	expected := []string{"ac-2", "ia-5", "si-10"}
	if len(result) != len(expected) {
		t.Errorf("Expected %d unique control IDs (deduped), got %d: %v", len(expected), len(result), result)
	}

	// Check for duplicates in result
	seen := make(map[string]bool)
	for _, id := range result {
		if seen[id] {
			t.Errorf("Found duplicate control ID in result: %s", id)
		}
		seen[id] = true
	}
}

func TestExtractControlIDs_MixedCase(t *testing.T) {
	response := `["AC-2", "ia-5", "SI-10"]`

	result, err := extractControlIDs(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should normalize to lowercase
	for _, id := range result {
		for _, c := range id {
			if c >= 'A' && c <= 'Z' {
				t.Errorf("Expected lowercase control IDs, found uppercase in: %s", id)
			}
		}
	}
}

func TestNormalizeControlIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "lowercase already",
			input:    []string{"ac-2", "ia-5"},
			expected: []string{"ac-2", "ia-5"},
		},
		{
			name:     "uppercase to lowercase",
			input:    []string{"AC-2", "IA-5"},
			expected: []string{"ac-2", "ia-5"},
		},
		{
			name:     "with whitespace",
			input:    []string{" ac-2 ", "ia-5\n"},
			expected: []string{"ac-2", "ia-5"},
		},
		{
			name:     "with duplicates",
			input:    []string{"ac-2", "AC-2", "ia-5"},
			expected: []string{"ac-2", "ia-5"},
		},
		{
			name:     "with empty strings",
			input:    []string{"ac-2", "", "ia-5", ""},
			expected: []string{"ac-2", "ia-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeControlIDs(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d IDs, got %d", len(tt.expected), len(result))
			}

			for i, id := range tt.expected {
				if result[i] != id {
					t.Errorf("Expected ID[%d] = %s, got %s", i, id, result[i])
				}
			}
		})
	}
}

func TestExtractJSONFromMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "json code block",
			input:    "```json\n[\"ac-2\"]\n```",
			expected: "[\"ac-2\"]",
		},
		{
			name:     "plain code block with JSON",
			input:    "```\n[\"ac-2\"]\n```",
			expected: "[\"ac-2\"]",
		},
		{
			name:     "no code block",
			input:    "[\"ac-2\"]",
			expected: "[\"ac-2\"]",
		},
		{
			name:     "code block with non-JSON",
			input:    "```\nsome text\n```",
			expected: "some text",
		},
		{
			name:     "multiple code blocks",
			input:    "Some text\n```json\n[\"ac-2\"]\n```\nMore text\n```\nother\n```",
			expected: "[\"ac-2\"]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromMarkdown(tt.input)

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractControlIDsFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single control",
			text:     "The system should implement ac-2",
			expected: []string{"ac-2"},
		},
		{
			name:     "multiple controls",
			text:     "Implement ac-2, ia-5, and si-10 controls",
			expected: []string{"ac-2", "ia-5", "si-10"},
		},
		{
			name:     "uppercase controls",
			text:     "Apply AC-2 and IA-5",
			expected: []string{"ac-2", "ia-5"},
		},
		{
			name:     "no controls",
			text:     "This text has no control identifiers",
			expected: []string{},
		},
		{
			name:     "duplicate controls",
			text:     "ac-2 is important, apply ac-2",
			expected: []string{"ac-2"},
		},
		{
			name:     "controls in parentheses",
			text:     "Access control (ac-2) and identification (ia-5)",
			expected: []string{"ac-2", "ia-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractControlIDsFromText(tt.text)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d control IDs, got %d: %v", len(tt.expected), len(result), result)
			}

			// Check each expected ID is in result
			for _, expectedID := range tt.expected {
				found := false
				for _, resultID := range result {
					if resultID == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find control ID %s in result %v", expectedID, result)
				}
			}
		})
	}
}
