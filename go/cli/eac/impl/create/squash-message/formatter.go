package squashmessage

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SquashJSON matches contracts/core/0.1.0/squash-message.schema.json.
type SquashJSON struct {
	Type                string       `json:"type"`
	Scope               string       `json:"scope"`
	Subject             string       `json:"subject"`
	Body                string       `json:"body"`
	Breaking            bool         `json:"breaking,omitempty"`
	BreakingDescription string       `json:"breaking_description,omitempty"`
	Changes             []ChangeItem `json:"changes,omitempty"`
	Modules             []string     `json:"modules,omitempty"`
	PRNumber            int          `json:"pr_number,omitempty"`
	AuditorSummary      string       `json:"auditor_summary,omitempty"`
}

// ChangeItem represents an individual change in the PR.
type ChangeItem struct {
	Type        string `json:"type"`
	Scope       string `json:"scope,omitempty"`
	Description string `json:"description"`
}

// FormatSquashMessage converts JSON schema output to squash commit format
//
// Expected format:
//
//	type(scope): subject
//
//	Auditor-Summary: High-level summary for auditors.
//
//	Body text with detailed explanation of all changes.
//
//	Changes: 2 files, +10 insertions, -0 deletions
//
//	BREAKING CHANGE: Description of breaking changes (if applicable)
func FormatSquashMessage(jsonOutput string) (string, error) {
	var squash SquashJSON
	if err := json.Unmarshal([]byte(jsonOutput), &squash); err != nil {
		return "", fmt.Errorf("failed to parse squash message JSON: %w", err)
	}

	var result strings.Builder

	// 1. Header: type(scope): subject
	result.WriteString(fmt.Sprintf("%s(%s): %s\n\n", squash.Type, squash.Scope, squash.Subject))

	// 2. Auditor-Summary (required field)
	result.WriteString("Auditor-Summary: ")
	if squash.AuditorSummary != "" {
		result.WriteString(squash.AuditorSummary)
	} else {
		// Fallback: extract from body
		auditorSummary := extractFirstSentence(squash.Body)
		result.WriteString(auditorSummary)
	}
	result.WriteString("\n\n")

	// 3. Body text
	result.WriteString(squash.Body)
	result.WriteString("\n\n")

	// 4. Changes list (if individual changes provided)
	if len(squash.Changes) > 0 {
		result.WriteString("Changes:\n")
		for _, change := range squash.Changes {
			if change.Scope != "" {
				result.WriteString(fmt.Sprintf("- %s(%s): %s\n", change.Type, change.Scope, change.Description))
			} else {
				result.WriteString(fmt.Sprintf("- %s: %s\n", change.Type, change.Description))
			}
		}
		result.WriteString("\n")
	}

	// 5. Breaking change notice (if applicable)
	if squash.Breaking {
		result.WriteString("BREAKING CHANGE: ")
		if squash.BreakingDescription != "" {
			result.WriteString(squash.BreakingDescription)
		} else {
			result.WriteString(squash.Subject)
		}
		result.WriteString("\n\n")
	}

	// 6. Modules affected (if multi-module)
	if len(squash.Modules) > 0 {
		result.WriteString("Modules: ")
		result.WriteString(strings.Join(squash.Modules, ", "))
		result.WriteString("\n\n")
	}

	// 7. PR reference (if provided)
	if squash.PRNumber > 0 {
		result.WriteString(fmt.Sprintf("PR: #%d\n\n", squash.PRNumber))
	}

	return strings.TrimRight(result.String(), "\n") + "\n", nil
}

// extractFirstSentence extracts the first sentence from text for auditor summary.
func extractFirstSentence(text string) string {
	text = strings.TrimSpace(text)

	// Find first sentence ending
	for _, delim := range []string{". ", ".\n", "? ", "?\n", "! ", "!\n"} {
		if idx := strings.Index(text, delim); idx != -1 {
			sentence := strings.TrimSpace(text[:idx+1])
			return sentence
		}
	}

	// If no sentence delimiter found, check if text ends with punctuation
	if strings.HasSuffix(text, ".") || strings.HasSuffix(text, "?") || strings.HasSuffix(text, "!") {
		lines := strings.Split(text, "\n")
		return strings.TrimSpace(lines[0])
	}

	// Fallback: use first line and add period
	lines := strings.Split(text, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return "Summary of changes."
	}
	if !strings.HasSuffix(firstLine, ".") {
		firstLine += "."
	}
	return firstLine
}
