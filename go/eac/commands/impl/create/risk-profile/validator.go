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
		errors = append(errors, contracts.ValidationError{
			Code:       &contracts.ErrInvalidJSON,
			LegacyCode: contracts.ErrInvalidJSON.Code,
			Message:    fmt.Sprintf("Invalid JSON: %v", err),
			Line:       0,
			Severity:   string(contracts.ErrInvalidJSON.Severity),
		})
		return errors
	}

	// Check for required controls field
	controls, ok := data["controls"]
	if !ok {
		errors = append(errors, contracts.ValidationError{
			Code:       &contracts.ErrMissingRequiredElement,
			LegacyCode: contracts.ErrMissingRequiredElement.Code,
			Message:    "Missing required field: controls",
			Line:       0,
			Severity:   string(contracts.ErrMissingRequiredElement.Severity),
		})
		return errors
	}

	// Validate controls is an array
	controlsArray, ok := controls.([]interface{})
	if !ok {
		errors = append(errors, contracts.ValidationError{
			Code:       &contracts.ErrInvalidPattern,
			LegacyCode: contracts.ErrInvalidPattern.Code,
			Message:    "Field 'controls' must be an array",
			Line:       0,
			Severity:   string(contracts.ErrInvalidPattern.Severity),
		})
		return errors
	}

	// Validate at least one control
	if len(controlsArray) == 0 {
		errors = append(errors, contracts.ValidationError{
			Code:       &contracts.ErrMissingRequiredElement,
			LegacyCode: contracts.ErrMissingRequiredElement.Code,
			Message:    "At least one control ID is required",
			Line:       0,
			Severity:   string(contracts.ErrMissingRequiredElement.Severity),
		})
	}

	// Validate control ID format (lowercase with hyphen: ac-2, ia-5, etc.)
	for i, ctrl := range controlsArray {
		ctrlStr, ok := ctrl.(string)
		if !ok {
			errors = append(errors, contracts.ValidationError{
				Code:       &contracts.ErrInvalidPattern,
				LegacyCode: contracts.ErrInvalidPattern.Code,
				Message:    fmt.Sprintf("Control at index %d is not a string", i),
				Line:       0,
				Severity:   string(contracts.ErrInvalidPattern.Severity),
			})
			continue
		}

		// Basic format check: should be lowercase letters, hyphen, numbers
		if len(ctrlStr) < 4 || (len(ctrlStr) > 2 && ctrlStr[2] != '-') {
			errors = append(errors, contracts.ValidationError{
				Code:       &contracts.ErrInvalidPattern,
				LegacyCode: contracts.ErrInvalidPattern.Code,
				Message:    fmt.Sprintf("Control '%s' has invalid format (expected: xx-#)", ctrlStr),
				Line:       0,
				Severity:   string(contracts.ErrInvalidPattern.Severity),
			})
		}
	}

	// Check for reasoning field (recommended but not required)
	if _, ok := data["reasoning"]; !ok {
		// Not an error, just log it
		// Could add a warning here if we implement warning system
	}

	return errors
}

// VerifyImplementation checks that all contract rules are implemented
func (v *RiskProfileValidator) VerifyImplementation() []contracts.ValidationError {
	// No specific verification needed for risk profile validator
	return nil
}
