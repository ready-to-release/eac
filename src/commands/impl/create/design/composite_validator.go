// Intent: Combine quick and full validation for optimal performance
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear two-phase validation: quick then full
//   - Explicit decision logic
//   - Aggregates errors from both validators
//
// Easy to change:
//   - Validators are pluggable
//   - Can easily add more validation phases
//   - Configuration controls behavior
//
// Hard to break:
//   - Always runs quick validation
//   - Only runs expensive validation when needed
//   - Collects all errors for better feedback

package design

import (
	"github.com/ready-to-release/eac/src/core/contracts"
)

// CompositeValidator runs quick validation followed by full validation
type CompositeValidator struct {
	quickValidator *QuickDSLValidator
	fullValidator  *StructurizrDSLValidator
	skipFullOnQuickErrors bool // Skip Docker validation if quick validation fails
}

// NewCompositeValidator creates a validator that combines quick and full validation
func NewCompositeValidator(module, templateRoot string, skipFullOnQuickErrors bool) (*CompositeValidator, error) {
	// Load design contract if available
	loader := contracts.NewContractLoader(templateRoot, "ai/design", "0.1.0")
	contract, err := loader.LoadContract()

	var quickValidator *QuickDSLValidator
	if err == nil && contract != nil {
		// Use contract-aware validator
		quickValidator = NewQuickDSLValidatorWithContract(contract)
	} else {
		// Fallback to default validator
		quickValidator = NewQuickDSLValidator()
	}

	fullValidator, err := NewStructurizrDSLValidator(module, templateRoot)
	if err != nil {
		return nil, err
	}

	return &CompositeValidator{
		quickValidator: quickValidator,
		fullValidator:  fullValidator,
		skipFullOnQuickErrors: skipFullOnQuickErrors,
	}, nil
}

// Validate runs quick validation first, then full validation if needed
func (v *CompositeValidator) Validate(output string, context map[string]interface{}) []contracts.ValidationError {
	var allErrors []contracts.ValidationError

	// Phase 1: Quick validation (< 100ms)
	quickErrors := v.quickValidator.Validate(output, context)

	if len(quickErrors) > 0 {
		// Found quick errors
		allErrors = append(allErrors, quickErrors...)

		// If configured to skip full validation on quick errors, return now
		// This saves 6-7 seconds of Docker startup when we know there are basic syntax errors
		if v.skipFullOnQuickErrors {
			// Add a helpful note
			allErrors = append(allErrors, contracts.ValidationError{
				Code:     "quick_validation_failed",
				Message:  "Skipping full validation due to syntax errors. Fix the above errors and retry.",
				Severity: "warning",
			})
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

// VerifyImplementation verifies both validators are ready
func (v *CompositeValidator) VerifyImplementation() []contracts.ValidationError {
	var errors []contracts.ValidationError

	quickErrors := v.quickValidator.VerifyImplementation()
	errors = append(errors, quickErrors...)

	fullErrors := v.fullValidator.VerifyImplementation()
	errors = append(errors, fullErrors...)

	return errors
}

// IsDockerRunning checks if Docker is available for full validation
func (v *CompositeValidator) IsDockerRunning() bool {
	return v.fullValidator.IsDockerRunning()
}

// Cleanup cleans up resources from the full validator
func (v *CompositeValidator) Cleanup() {
	v.fullValidator.Cleanup()
}
