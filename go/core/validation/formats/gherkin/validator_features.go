package gherkin

import (
	"fmt"
	"regexp"

	"github.com/ready-to-release/eac/go/core/validation"
)

// validateFeatureNaming checks if feature name follows the naming convention
//
// Expected format: <module>_<feature-name>
// Pattern is read from contract.feature_naming_pattern.
func (v *Validator) validateFeatureNaming(featureName string, lineNum int) validation.ValidationError {
	// Get naming pattern from contract
	pattern := `^[a-z][a-z0-9-]*_[a-z][a-z0-9-]*$` // Default fallback
	if v.contract != nil && v.contract.RawData != nil {
		if patternVal, ok := v.contract.RawData["feature_naming_pattern"].(string); ok && patternVal != "" {
			pattern = patternVal
		}
	}

	matched, err := regexp.MatchString(pattern, featureName)
	if err != nil || !matched {
		formatter := validation.NewErrorFormatter()
		return *formatter.FormatEnhancedError(
			validation.ErrInvalidFeatureNaming,
			fmt.Sprintf("Feature name '%s' does not follow naming convention", featureName),
			fmt.Sprintf("Feature: %s", featureName),
			"Format: <module>_<feature-name>",
			`Valid examples:
  Feature: eac-cli_risk-assessment
  Feature: core_validation-framework
  Feature: repository_dependency-analysis

Invalid examples:
  Feature: Risk Assessment          ✗ (missing module prefix)
  Feature: eac_commands_risk        ✗ (use hyphen, not underscore within names)
  Feature: EAC-Commands_Risk        ✗ (must be lowercase)`,
			"Use lowercase letters, numbers, and hyphens. Format: <module>_<feature-name>",
			lineNum,
		)
	}

	return validation.ValidationError{} // No error (empty struct)
}
