// Package commitmessage provides commit message validation against contract
package commitmessage

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/domain"
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

// NewModuleSectionValidator creates a validator for a specific module's section.
func NewModuleSectionValidator(moduleName string) *ModuleSectionValidator {
	return &ModuleSectionValidator{
		moduleName: moduleName,
	}
}

// Validate validates a module section against the expected format.
func (v *ModuleSectionValidator) Validate(output string, context map[string]interface{}) []domain.ValidationError {
	var errors []domain.ValidationError

	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrEmptyModuleSection,
			"Module section is empty",
			0,
		))
		return errors
	}

	// Find the structure: module name, dashes, subject line
	// Skip any leading blank lines
	startIdx := 0
	for startIdx < len(lines) && strings.TrimSpace(lines[startIdx]) == "" {
		startIdx++
	}

	if startIdx >= len(lines) {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrEmptyModuleSection,
			"Module section contains only blank lines",
			0,
		))
		return errors
	}

	// Line 1: Module name
	moduleName := strings.TrimSpace(lines[startIdx])
	if !isModuleName(moduleName) {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrInvalidModuleName,
			fmt.Sprintf("First line should be module name, got: %s", moduleName),
			startIdx+1,
		))
	}

	// Line 2: Dashes
	if startIdx+1 >= len(lines) {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrMissingDashes,
			"Module section missing dashes separator line",
			0,
		))
		return errors
	}

	dashesLine := strings.TrimSpace(lines[startIdx+1])
	if !isDashesLine(dashesLine) {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrInvalidDashes,
			fmt.Sprintf("Second line should be dashes (--------), got: %s", dashesLine),
			startIdx+2,
		))
	}

	// Line 3: Subject line (<module>: <type>: <description>)
	if startIdx+2 >= len(lines) {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrMissingSubjectLine,
			"Module section missing subject line",
			0,
		))
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
		errors = append(errors, *domain.NewValidationError(
			domain.ErrInvalidSubjectFormat,
			fmt.Sprintf("Subject line must follow '<module>: <type>: <description>' format, got: %s", subjectLine),
			startIdx+3,
		))
	}

	// Check subject line length
	if len(subjectLine) > MaxSubjectLength {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrSubjectTooLong,
			fmt.Sprintf("Subject line exceeds %d characters (%d chars)", MaxSubjectLength, len(subjectLine)),
			0,
		))
	}

	// Check for trailing period
	if strings.HasSuffix(subjectLine, ".") && !strings.HasSuffix(subjectLine, "...") {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrSubjectTrailingPeriod,
			"Subject line must not end with period",
			0,
		))
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
		errors = append(errors, *domain.NewValidationError(
			domain.ErrMissingBody,
			"Module section missing body text",
			0,
		))
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
			errors = append(errors, *domain.NewValidationError(
				domain.ErrLineTooLong,
				fmt.Sprintf("Line exceeds %d characters (%d chars): %s", MaxLineLength, len(trimmed), preview),
				i+1,
			))
		}
	}

	return errors
}

// VerifyImplementation is a no-op for module validators (no contract to verify against).
func (v *ModuleSectionValidator) VerifyImplementation() []domain.ValidationError {
	return nil
}
