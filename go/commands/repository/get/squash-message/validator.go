package squashmessage

import (
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/core/domain"
)

// SquashMessageValidator validates squash message output.
type SquashMessageValidator struct{}

// NewSquashMessageValidator creates a new validator.
func NewSquashMessageValidator() *SquashMessageValidator {
	return &SquashMessageValidator{}
}

// Validate validates the squash message output (formatted text).
func (v *SquashMessageValidator) Validate(output string, context map[string]interface{}) []domain.ValidationError {
	var errors []domain.ValidationError

	// Basic checks for formatted commit message
	lines := strings.Split(output, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrEmptyOutput,
			"Commit message is empty",
			0,
		))
		return errors
	}

	// Check header format (type(scope): subject)
	header := strings.TrimSpace(lines[0])
	if !strings.Contains(header, ":") {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrInvalidPattern,
			"Header must follow format: type(scope): subject",
			1,
		))
	}

	// Check header length
	if len(header) > 72 {
		errors = append(errors, *domain.NewValidationError(
			domain.ErrLineTooLong,
			fmt.Sprintf("Header exceeds 72 characters (%d)", len(header)),
			1,
		))
	}

	return errors
}

// VerifyImplementation checks validator implementation.
func (v *SquashMessageValidator) VerifyImplementation() []domain.ValidationError {
	return nil
}
