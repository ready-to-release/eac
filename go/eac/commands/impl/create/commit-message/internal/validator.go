package commitmessage

import (
	"github.com/ready-to-release/eac/go/eac/core/contracts"
)

// CommitMessageValidator implements contracts.Validator for commit message validation.
type CommitMessageValidator struct {
	contract        *CommitMessageContract
	affectedModules []string
	workspaceRoot   string
}

// NewCommitMessageValidator creates a new commit message validator
//
// Parameters:
//   - contractData: The loaded commit message contract
//   - affectedModules: List of modules with staged changes
//   - workspaceRoot: Workspace root for config loading
func NewCommitMessageValidator(
	contractData *CommitMessageContract,
	affectedModules []string,
	workspaceRoot string,
) *CommitMessageValidator {
	return &CommitMessageValidator{
		contract:        contractData,
		affectedModules: affectedModules,
		workspaceRoot:   workspaceRoot,
	}
}

// Validate validates a commit message against the contract
//
// The context parameter can contain:
//   - "affectedModules": []string - list of modules with changes (overrides constructor value)
//
// Returns validation errors (empty slice if valid).
func (v *CommitMessageValidator) Validate(output string, context map[string]interface{}) []contracts.ValidationError {
	// Use affected modules from context if provided, otherwise use constructor value
	modules := v.affectedModules
	if contextModules, ok := context["affectedModules"].([]string); ok {
		modules = contextModules
	}

	// Use existing validation function
	return VerifyCommitMessageContract(output, modules)
}

// VerifyImplementation verifies that the validator implements all contract rules.
func (v *CommitMessageValidator) VerifyImplementation() []contracts.ValidationError {
	// Verify contract can be loaded from unified config
	_, err := LoadContractFromConfig(v.workspaceRoot)
	if err != nil {
		return []contracts.ValidationError{*contracts.NewLegacyValidationError(
			"CONTRACT_LOAD_ERROR",
			err.Error(),
			0,
			"error",
		)}
	}
	return nil
}
