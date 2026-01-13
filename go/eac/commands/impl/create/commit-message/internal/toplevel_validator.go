// Package commitmessage provides commit message validation against contract
package commitmessage

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

// TopLevelValidator validates the top-level commit message (header, auditor-summary, body)
//
// Top-level output has the format:
//
//	<type>(<scope>): <summary>
//
//	Auditor-Summary: <one sentence>
//
//	<body text>
//
//	Changes: N files, +X insertions, -Y deletions
//
// This validator checks ONLY this format, not module sections.
type TopLevelValidator struct {
	affectedModules []string
}

// NewTopLevelValidator creates a validator for top-level commit output
func NewTopLevelValidator(affectedModules []string) *TopLevelValidator {
	return &TopLevelValidator{
		affectedModules: affectedModules,
	}
}

// Validate validates a top-level commit message against the expected format
func (v *TopLevelValidator) Validate(output string, context map[string]interface{}) []contracts.ValidationError {
	var errors []contracts.ValidationError

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"EMPTY_MESSAGE",
			"Top-level commit message is empty",
			0,
			"error",
		))
		return errors
	}

	// Rule 1: First line must be conventional commit header
	firstLine := strings.TrimSpace(lines[0])
	if !getConventionalCommitRegex().MatchString(firstLine) {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"INVALID_HEADER_FORMAT",
			"Header must follow format: <type>(<scope>): <summary> (e.g., feat(cli): add new command)",
			1,
			"error",
		))
	}

	// Rule 2: Header max length
	if len(firstLine) > MaxHeaderLength {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"HEADER_TOO_LONG",
			fmt.Sprintf("Header exceeds %d characters (%d chars)", MaxHeaderLength, len(firstLine)),
			1,
			"error",
		))
	}

	// Rule 3: No trailing period (except ellipsis)
	if strings.HasSuffix(firstLine, ".") && !strings.HasSuffix(firstLine, "...") {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"HEADER_TRAILING_PERIOD",
			"Header must not end with period",
			1,
			"error",
		))
	}

	// Rule 4: Check for Auditor-Summary field
	hasAuditorSummary := false
	for i := 1; i < len(lines) && i < 10; i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "Auditor-Summary:") {
			hasAuditorSummary = true
			break
		}
	}
	if !hasAuditorSummary {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"MISSING_AUDITOR_SUMMARY",
			"Missing Auditor-Summary field after header",
			0,
			"error",
		))
	}

	// Rule 5: Check for body text after Auditor-Summary
	hasBody := false
	afterAuditorSummary := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Auditor-Summary:") {
			afterAuditorSummary = true
			continue
		}

		// Check for body text (not empty, not Changes line, not module-related)
		if afterAuditorSummary && trimmed != "" &&
			!strings.HasPrefix(trimmed, "Changes:") &&
			!isModuleName(trimmed) &&
			!isDashesLine(trimmed) {
			hasBody = true
			break
		}
	}
	if !hasBody {
		errors = append(errors, *contracts.NewLegacyValidationError(
			"MISSING_BODY",
			"Missing body text after Auditor-Summary",
			0,
			"error",
		))
	}

	// Rule 6: Check that output does NOT contain module sections (those go separately)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for any markdown headers (## or ###) - these shouldn't be in top-level
		if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "### ") {
			errors = append(errors, *contracts.NewLegacyValidationError(
			"UNEXPECTED_MODULE_SECTION",
			fmt.Sprintf("Top-level output should not contain markdown headers (found: %s)", trimmed),
			0,
			"error",
		))
		}

		// Check for bullet points starting with module-like patterns (- src-, - ext-, etc.)
		if strings.HasPrefix(trimmed, "- ") && len(trimmed) > 2 {
			rest := trimmed[2:]
			// Check if it looks like a file path or module reference
			if strings.Contains(rest, "/") || strings.HasPrefix(rest, "New ") || strings.HasPrefix(rest, "Updated ") {
				errors = append(errors, *contracts.NewLegacyValidationError(
			"UNEXPECTED_FILE_LIST",
			fmt.Sprintf("Top-level output should not contain file/change lists (found: %s)", trimmed),
			0,
			"error",
		))
			}
		}

		// Check for module name + dashes pattern
		if i < len(lines)-1 && isModuleName(trimmed) {
			nextLine := strings.TrimSpace(lines[i+1])
			if isDashesLine(nextLine) && len(nextLine) > 3 {
				errors = append(errors, *contracts.NewLegacyValidationError(
			"UNEXPECTED_MODULE_SECTION",
			fmt.Sprintf("Top-level output should not contain module sections (found: %s)", trimmed),
			0,
			"error",
		))
			}
		}
	}

	// Rule 7: Check line lengths in body
	inBody := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Auditor-Summary:") {
			inBody = true
			continue
		}

		if !inBody {
			continue
		}

		// Skip special lines
		if trimmed == "" ||
			strings.HasPrefix(trimmed, "Changes:") ||
			strings.HasPrefix(trimmed, "|") ||
			strings.HasPrefix(trimmed, "```") {
			continue
		}

		if len(trimmed) > MaxLineLength {
			preview := trimmed
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			errors = append(errors, *contracts.NewLegacyValidationError(
			"LINE_TOO_LONG",
			fmt.Sprintf("Line exceeds %d characters (%d chars): %s", MaxLineLength, len(trimmed), preview),
			0,
			"warning",
		))
		}
	}

	return errors
}

// VerifyImplementation is a no-op for top-level validators
func (v *TopLevelValidator) VerifyImplementation() []contracts.ValidationError {
	return nil
}
