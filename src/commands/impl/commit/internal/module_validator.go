// Package commitmessage provides commit message validation against contract
package commitmessage

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/src/core/contracts"
)

// ModuleSectionValidator validates individual module sections (not full commit messages)
//
// Module sections have the format:
//
//	<module-name>
//	------------
//	<module>: <type>: <description>
//
//	<body text>
//
// This validator checks ONLY this format, not full commit message structure.
type ModuleSectionValidator struct {
	moduleName string
}

// NewModuleSectionValidator creates a validator for a specific module's section
func NewModuleSectionValidator(moduleName string) *ModuleSectionValidator {
	return &ModuleSectionValidator{
		moduleName: moduleName,
	}
}

// Validate validates a module section against the expected format
func (v *ModuleSectionValidator) Validate(output string, context map[string]interface{}) []contracts.ValidationError {
	var errors []contracts.ValidationError

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		errors = append(errors, contracts.ValidationError{
			Code:     "EMPTY_MODULE_SECTION",
			Message:  "Module section is empty",
			Severity: "error",
		})
		return errors
	}

	// Find the structure: module name, dashes, subject line
	// Skip any leading blank lines
	startIdx := 0
	for startIdx < len(lines) && strings.TrimSpace(lines[startIdx]) == "" {
		startIdx++
	}

	if startIdx >= len(lines) {
		errors = append(errors, contracts.ValidationError{
			Code:     "EMPTY_MODULE_SECTION",
			Message:  "Module section contains only blank lines",
			Severity: "error",
		})
		return errors
	}

	// Line 1: Module name
	moduleName := strings.TrimSpace(lines[startIdx])
	if !isModuleName(moduleName) {
		errors = append(errors, contracts.ValidationError{
			Code:     "INVALID_MODULE_NAME",
			Message:  fmt.Sprintf("First line should be module name, got: %s", moduleName),
			Line:     startIdx + 1,
			Severity: "error",
		})
	}

	// Line 2: Dashes
	if startIdx+1 >= len(lines) {
		errors = append(errors, contracts.ValidationError{
			Code:     "MISSING_DASHES",
			Message:  "Module section missing dashes separator line",
			Severity: "error",
		})
		return errors
	}

	dashesLine := strings.TrimSpace(lines[startIdx+1])
	if !isDashesLine(dashesLine) {
		errors = append(errors, contracts.ValidationError{
			Code:     "INVALID_DASHES",
			Message:  fmt.Sprintf("Second line should be dashes (--------), got: %s", dashesLine),
			Line:     startIdx + 2,
			Severity: "error",
		})
	}

	// Line 3: Subject line (<module>: <type>: <description>)
	if startIdx+2 >= len(lines) {
		errors = append(errors, contracts.ValidationError{
			Code:     "MISSING_SUBJECT_LINE",
			Message:  "Module section missing subject line",
			Severity: "error",
		})
		return errors
	}

	subjectLine := strings.TrimSpace(lines[startIdx+2])
	if subjectLine == "" {
		// Maybe there's a blank line before subject - check next line
		if startIdx+3 < len(lines) {
			subjectLine = strings.TrimSpace(lines[startIdx+3])
		}
	}

	if subjectLine != "" && !getModuleSubjectLineRegex().MatchString(subjectLine) {
		errors = append(errors, contracts.ValidationError{
			Code:     "INVALID_SUBJECT_FORMAT",
			Message:  fmt.Sprintf("Subject line must follow '<module>: <type>: <description>' format, got: %s", subjectLine),
			Line:     startIdx + 3,
			Severity: "error",
		})
	}

	// Check subject line length
	if len(subjectLine) > MaxSubjectLength {
		errors = append(errors, contracts.ValidationError{
			Code:     "SUBJECT_TOO_LONG",
			Message:  fmt.Sprintf("Subject line exceeds %d characters (%d chars)", MaxSubjectLength, len(subjectLine)),
			Severity: "error",
		})
	}

	// Check for trailing period
	if strings.HasSuffix(subjectLine, ".") && !strings.HasSuffix(subjectLine, "...") {
		errors = append(errors, contracts.ValidationError{
			Code:     "SUBJECT_TRAILING_PERIOD",
			Message:  "Subject line must not end with period",
			Severity: "error",
		})
	}

	// Check body text exists (after blank line following subject)
	hasBody := false
	for i := startIdx + 3; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed != "" && !getModuleSubjectLineRegex().MatchString(trimmed) {
			hasBody = true
			break
		}
	}

	if !hasBody {
		errors = append(errors, contracts.ValidationError{
			Code:     "MISSING_BODY",
			Message:  "Module section missing body text",
			Severity: "warning",
		})
	}

	// Check line lengths in body
	for i := startIdx + 3; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || isDashesLine(trimmed) || trimmed == "---" {
			continue
		}
		if len(trimmed) > MaxLineLength {
			preview := trimmed
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			errors = append(errors, contracts.ValidationError{
				Code:     "LINE_TOO_LONG",
				Message:  fmt.Sprintf("Line exceeds %d characters (%d chars): %s", MaxLineLength, len(trimmed), preview),
				Line:     i + 1,
				Severity: "warning",
			})
		}
	}

	return errors
}

// VerifyImplementation is a no-op for module validators (no contract to verify against)
func (v *ModuleSectionValidator) VerifyImplementation() []contracts.ValidationError {
	return nil
}
