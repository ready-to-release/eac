package gherkin

import (
	"fmt"

	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/validation"
)

// VerifyImplementation verifies that the validator implements all contract rules
//
// This is a self-check to ensure the validator stays in sync with the contract.
func (v *Validator) VerifyImplementation() []validation.ValidationError {
	var errors []validation.ValidationError

	// Check that contract is loaded
	if v.contract == nil {
		errors = append(errors, *validation.NewValidationError(validation.ErrNoContract, "Contract not loaded", 0))
		return errors
	}

	// Verify contract version matches
	if v.contract.Version != paths.DefaultsVersion {
		errors = append(errors, *validation.NewValidationError(validation.ErrContractVersionMismatch, fmt.Sprintf("Expected contract version %s, got %s", paths.DefaultsVersion, v.contract.Version), 0))
	}

	// Verify contract name
	if v.contract.Name != "Gherkin Specification Structure" {
		errors = append(errors, *validation.NewValidationError(validation.ErrContractNameMismatch, fmt.Sprintf("Expected contract name 'Gherkin Specification Structure', got '%s'", v.contract.Name), 0))
	}

	// Verify tag contract is loaded (warning only, not required)
	if v.tagsConfig == nil {
		errors = append(errors, *validation.NewValidationError(validation.ErrNoTagContract, "Tag contract not loaded - advanced tag validation disabled", 0))
	}

	return errors
}
