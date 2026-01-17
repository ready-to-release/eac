package commitmessage

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CommitJSON matches contracts/eac-core/0.1.0/commit-message.schema.json.
type CommitJSON struct {
	Type     string         `json:"type"`
	Scope    string         `json:"scope"`
	Subject  string         `json:"subject"`
	Body     string         `json:"body,omitempty"`
	Breaking bool           `json:"breaking,omitempty"`
	Modules  []ModuleChange `json:"modules,omitempty"`
}

// ModuleChange represents module-specific changes in multi-module commits.
type ModuleChange struct {
	Name    string `json:"name"`
	Changes string `json:"changes"`
}

// ModuleSectionJSON matches contracts/eac-core/0.1.0/commit-message-module.schema.json.
type ModuleSectionJSON struct {
	Module      string `json:"module"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

// FormatCommitMessage converts JSON schema output to conventional commit format
//
// Expected format:
//
//	type(scope): subject
//
//	Auditor-Summary: One-sentence summary of the change.
//
//	Body paragraph explaining the change in detail.
//
//	module-name
//	-----------
//	module-name: type: module subject
//
//	Module-specific body text.
func FormatCommitMessage(jsonOutput string) (string, error) {
	var commit CommitJSON
	if err := json.Unmarshal([]byte(jsonOutput), &commit); err != nil {
		return "", fmt.Errorf("failed to parse commit JSON: %w", err)
	}

	var result strings.Builder

	// 1. Header: type(scope): subject
	result.WriteString(fmt.Sprintf("%s(%s): %s\n\n", commit.Type, commit.Scope, commit.Subject))

	// 2. Auditor-Summary: (required field - use first sentence of body or subject)
	result.WriteString("Auditor-Summary: ")
	auditorSummary := extractAuditorSummary(commit.Body, commit.Subject)
	result.WriteString(auditorSummary)
	result.WriteString("\n\n")

	// 3. Body text (if provided)
	if commit.Body != "" {
		result.WriteString(commit.Body)
		result.WriteString("\n\n")
	}

	// 4. Breaking change notice (if applicable)
	if commit.Breaking {
		result.WriteString("BREAKING CHANGE: ")
		result.WriteString(commit.Subject)
		result.WriteString("\n\n")
	}

	// 5. Module sections (if multi-module)
	if len(commit.Modules) > 0 {
		for i, mod := range commit.Modules {
			// Module header
			result.WriteString(mod.Name)
			result.WriteString("\n")
			result.WriteString(strings.Repeat("-", len(mod.Name)))
			result.WriteString("\n")

			// Module subject line: module-name: type: subject
			moduleSubject := extractModuleSubject(mod.Changes)
			result.WriteString(fmt.Sprintf("%s: %s: %s\n\n", mod.Name, commit.Type, moduleSubject))

			// Module body
			result.WriteString(mod.Changes)
			result.WriteString("\n")

			// Add extra newline between modules (but not after last one)
			if i < len(commit.Modules)-1 {
				result.WriteString("\n")
			}
		}
	}

	return strings.TrimRight(result.String(), "\n") + "\n", nil
}

// extractAuditorSummary extracts a one-sentence summary for Auditor-Summary field
// Uses the first sentence of the body if available, otherwise falls back to subject.
func extractAuditorSummary(body, subject string) string {
	if body == "" {
		return subject + "."
	}

	// Find first sentence (terminated by period, question mark, or exclamation)
	body = strings.TrimSpace(body)

	// Find first sentence ending
	for _, delim := range []string{". ", ".\n", "? ", "?\n", "! ", "!\n"} {
		if idx := strings.Index(body, delim); idx != -1 {
			sentence := strings.TrimSpace(body[:idx+1])
			return sentence
		}
	}

	// If no sentence delimiter found, check if body itself ends with punctuation
	if strings.HasSuffix(body, ".") || strings.HasSuffix(body, "?") || strings.HasSuffix(body, "!") {
		// Get first line/paragraph
		lines := strings.Split(body, "\n")
		return strings.TrimSpace(lines[0])
	}

	// Fallback: use first line and add period
	lines := strings.Split(body, "\n")
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return subject + "."
	}
	if !strings.HasSuffix(firstLine, ".") {
		firstLine += "."
	}
	return firstLine
}

// extractModuleSubject extracts a concise subject line from module changes
// Takes the first sentence or first line as the subject.
func extractModuleSubject(changes string) string {
	changes = strings.TrimSpace(changes)

	// Find first sentence
	for _, delim := range []string{". ", ".\n"} {
		if idx := strings.Index(changes, delim); idx != -1 {
			return strings.TrimSpace(changes[:idx])
		}
	}

	// Get first line
	lines := strings.Split(changes, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Remove trailing period if present
	return strings.TrimSuffix(firstLine, ".")
}

// FormatModuleSection converts module section JSON to formatted text
//
// Expected format:
//
//	module-name
//	-----------
//	module-name: type: description
//
//	Body text explaining changes.
func FormatModuleSection(jsonOutput string) (string, error) {
	var module ModuleSectionJSON
	if err := json.Unmarshal([]byte(jsonOutput), &module); err != nil {
		return "", fmt.Errorf("failed to parse module section JSON: %w", err)
	}

	var result strings.Builder

	// 1. Module name header
	result.WriteString(module.Module)
	result.WriteString("\n")

	// 2. Separator line (dashes matching module name length)
	result.WriteString(strings.Repeat("-", len(module.Module)))
	result.WriteString("\n")

	// 3. Module subject line: module-name: type: description
	result.WriteString(fmt.Sprintf("%s: %s: %s\n\n", module.Module, module.Type, module.Description))

	// 4. Body text
	result.WriteString(module.Body)

	return result.String(), nil
}
