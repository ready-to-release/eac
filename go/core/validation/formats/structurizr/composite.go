package structurizr

import (
	"github.com/ready-to-release/eac/go/core/validation"
)

// CompositeValidator runs quick validation followed by full validation
// Used for create command to provide fast feedback with comprehensive validation.
type CompositeValidator struct {
	quickValidator        *QuickValidator
	fullValidator         *DockerValidator
	skipFullOnQuickErrors bool // Skip Docker validation if quick validation fails
}

// NewCompositeValidator creates a validator that combines quick and full validation.
func NewCompositeValidator(module, templateRoot string, skipFullOnQuickErrors bool) (*CompositeValidator, error) {
	// Use default quick validator
	quickValidator := NewQuickValidator()

	fullValidator, err := NewDockerValidator(module, templateRoot)
	if err != nil {
		return nil, err
	}

	return &CompositeValidator{
		quickValidator:        quickValidator,
		fullValidator:         fullValidator,
		skipFullOnQuickErrors: skipFullOnQuickErrors,
	}, nil
}

// Validate runs quick validation first, then full validation if needed.
func (v *CompositeValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	var allErrors []validation.ValidationError

	// Phase 1: Quick validation (< 100ms)
	quickErrors := v.quickValidator.Validate(output, context)

	if len(quickErrors) > 0 {
		// Found quick errors
		allErrors = append(allErrors, quickErrors...)

		// If configured to skip full validation on quick errors, return now
		// This saves 6-7 seconds of Docker startup when we know there are basic syntax errors
		if v.skipFullOnQuickErrors {
			// Add a helpful note
			allErrors = append(allErrors, *validation.NewValidationError(
				validation.ErrQuickValidationFailed,
				"Skipping full validation due to syntax errors. Fix the above errors and retry.",
				0,
			))
			return allErrors
		}
	}

	// Phase 2: Full Docker validation (6-7 seconds)
	// Only reached if:
	// - No quick errors found, OR
	// - skipFullOnQuickErrors is false
	fullErrors := v.fullValidator.Validate(output, context)
	allErrors = append(allErrors, fullErrors...)

	return allErrors
}

// VerifyImplementation verifies both validators are ready.
func (v *CompositeValidator) VerifyImplementation() []validation.ValidationError {
	quickErrors := v.quickValidator.VerifyImplementation()
	fullErrors := v.fullValidator.VerifyImplementation()

	errors := make([]validation.ValidationError, 0, len(quickErrors)+len(fullErrors))
	errors = append(errors, quickErrors...)
	errors = append(errors, fullErrors...)

	return errors
}

// IsDockerRunning checks if Docker is available for full validation.
func (v *CompositeValidator) IsDockerRunning() bool {
	return v.fullValidator.IsDockerRunning()
}

// Cleanup cleans up resources from the full validator.
func (v *CompositeValidator) Cleanup() {
	v.fullValidator.Cleanup()
}
