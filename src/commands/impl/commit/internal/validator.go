package commitmessage

import (
	"github.com/ready-to-release/eac/src/core/ai/contract"
)

// CommitMessageValidator implements contract.SpecValidator for commit message validation
type CommitMessageValidator struct {
	contract        *CommitMessageContract
	antiCorruption  *contract.AntiCorruptionRules
	affectedModules []string
	contractPath    string
}

// NewCommitMessageValidator creates a new commit message validator
//
// Parameters:
//   - contractData: The loaded commit message contract (structure.yml)
//   - antiCorruption: Anti-corruption rules
//   - affectedModules: List of modules with staged changes
//   - contractPath: Path to contract file for implementation verification
func NewCommitMessageValidator(
	contractData *CommitMessageContract,
	antiCorruption *contract.AntiCorruptionRules,
	affectedModules []string,
	contractPath string,
) *CommitMessageValidator {
	return &CommitMessageValidator{
		contract:        contractData,
		antiCorruption:  antiCorruption,
		affectedModules: affectedModules,
		contractPath:    contractPath,
	}
}

// Validate validates a commit message against the contract
//
// The context parameter can contain:
//   - "affectedModules": []string - list of modules with changes (overrides constructor value)
//
// Returns validation errors (empty slice if valid)
func (v *CommitMessageValidator) Validate(output string, context map[string]interface{}) []contract.ValidationError {
	// Use affected modules from context if provided, otherwise use constructor value
	modules := v.affectedModules
	if contextModules, ok := context["affectedModules"].([]string); ok {
		modules = contextModules
	}

	// Use existing validation function
	return VerifyCommitMessageContract(output, modules)
}

// VerifyImplementation verifies that the validator implements all contract rules
func (v *CommitMessageValidator) VerifyImplementation() []contract.ValidationError {
	// Use existing contract verification function
	return VerifyContractImplementation(v.contractPath)
}
