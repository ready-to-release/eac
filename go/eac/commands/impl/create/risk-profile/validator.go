package riskprofile

import (
	"encoding/json"
	"fmt"

	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

// RiskProfileValidator validates risk profile JSON output
type RiskProfileValidator struct{}

// NewRiskProfileValidator creates a new risk profile validator
func NewRiskProfileValidator() *RiskProfileValidator {
	return &RiskProfileValidator{}
}

// Validate validates the risk profile JSON output
func (v *RiskProfileValidator) Validate(output string, context map[string]interface{}) []contracts.ValidationError {
	var errors []contracts.ValidationError

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return []contracts.ValidationError{
			*contracts.NewValidationError(contracts.ErrInvalidJSON, fmt.Sprintf("Invalid JSON: %v", err), 0),
		}
	}

	// Check for required controls field
	controls, ok := data["controls"]
	if !ok {
		return []contracts.ValidationError{
			*contracts.NewValidationError(contracts.ErrMissingRequiredElement, "Missing required field: controls", 0),
		}
	}

	// Validate controls is an array
	controlsArray, ok := controls.([]interface{})
	if !ok {
		return []contracts.ValidationError{
			*contracts.NewValidationError(contracts.ErrInvalidPattern, "Field 'controls' must be an array", 0),
		}
	}

	// Validate at least one control
	if len(controlsArray) == 0 {
		errors = append(errors, *contracts.NewValidationError(
			contracts.ErrMissingRequiredElement, "At least one control ID is required", 0,
		))
	}

	// Validate control ID format (lowercase with hyphen: ac-2, ia-5, etc.)
	for i, ctrl := range controlsArray {
		ctrlStr, ok := ctrl.(string)
		if !ok {
			errors = append(errors, *contracts.NewValidationError(
				contracts.ErrInvalidPattern, fmt.Sprintf("Control at index %d is not a string", i), 0,
			))
			continue
		}

		// Basic format check: should be lowercase letters, hyphen, numbers
		if len(ctrlStr) < 4 || (len(ctrlStr) > 2 && ctrlStr[2] != '-') {
			errors = append(errors, *contracts.NewValidationError(
				contracts.ErrInvalidPattern, fmt.Sprintf("Control '%s' has invalid format (expected: xx-#)", ctrlStr), 0,
			))
		}
	}

	return errors
}

// VerifyImplementation checks that all contract rules are implemented
func (v *RiskProfileValidator) VerifyImplementation() []contracts.ValidationError {
	// No specific verification needed for risk profile validator
	return nil
}
