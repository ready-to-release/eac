package riskprofile

import (
	"encoding/json"
	"strings"
)

// extractControlIDs extracts control IDs from AI response.
// Supports multiple formats: JSON array, JSON object with controls field, or text with control IDs.
func extractControlIDs(response string) ([]string, error) {
	// Try to find JSON array in response
	response = extractJSONFromMarkdown(response)

	// Try parsing as JSON array
	var controlIDs []string
	if err := json.Unmarshal([]byte(response), &controlIDs); err == nil {
		return normalizeControlIDs(controlIDs), nil
	}

	// Try parsing as JSON object with controls field
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(response), &obj); err == nil {
		// Look for common field names
		for _, field := range []string{"controls", "control_ids", "controlIds", "ids"} {
			if ids, ok := obj[field].([]interface{}); ok {
				var result []string
				for _, id := range ids {
					if s, ok := id.(string); ok {
						result = append(result, s)
					}
				}
				return normalizeControlIDs(result), nil
			}
		}
	}

	// Fallback: extract control IDs using pattern matching
	return extractControlIDsFromText(response), nil
}

// extractJSONFromMarkdown extracts JSON from markdown code blocks.
// Handles both ```json and ``` blocks.
func extractJSONFromMarkdown(text string) string {
	// Look for ```json ... ``` blocks
	jsonStart := strings.Index(text, "```json")
	if jsonStart != -1 {
		jsonStart += 7
		jsonEnd := strings.Index(text[jsonStart:], "```")
		if jsonEnd != -1 {
			return strings.TrimSpace(text[jsonStart : jsonStart+jsonEnd])
		}
	}

	// Look for ``` ... ``` blocks
	codeStart := strings.Index(text, "```")
	if codeStart != -1 {
		codeStart += 3
		codeEnd := strings.Index(text[codeStart:], "```")
		if codeEnd != -1 {
			return strings.TrimSpace(text[codeStart : codeStart+codeEnd])
		}
	}

	return text
}

// extractControlIDsFromText extracts control IDs using pattern matching.
// Looks for NIST 800-53 style control IDs (e.g., ac-2, ia-5, si-10).
func extractControlIDsFromText(text string) []string {
	// Common NIST 800-53 control ID patterns
	patterns := []string{
		"ac-", "at-", "au-", "ca-", "cm-", "cp-", "ia-", "ir-",
		"ma-", "mp-", "pe-", "pl-", "pm-", "ps-", "pt-", "ra-",
		"sa-", "sc-", "si-", "sr-",
	}

	var controlIDs []string
	seen := make(map[string]bool)

	text = strings.ToLower(text)

	for _, prefix := range patterns {
		idx := 0
		for {
			pos := strings.Index(text[idx:], prefix)
			if pos == -1 {
				break
			}
			idx += pos

			// Extract the full control ID (e.g., "ac-2", "si-10")
			end := idx + len(prefix)
			for end < len(text) && (text[end] >= '0' && text[end] <= '9') {
				end++
			}

			if end > idx+len(prefix) {
				controlID := text[idx:end]
				if !seen[controlID] {
					seen[controlID] = true
					controlIDs = append(controlIDs, controlID)
				}
			}

			idx++
		}
	}

	return controlIDs
}

// normalizeControlIDs normalizes control IDs to lowercase and removes duplicates.
func normalizeControlIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]bool)

	for _, id := range ids {
		normalized := strings.ToLower(strings.TrimSpace(id))
		if normalized != "" && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}
