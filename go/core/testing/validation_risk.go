package testing

import (
	"fmt"
	"strings"
)

// validateRiskControlTag validates @risk-control:<name>-<id> format.
func validateRiskControlTag(tag string) bool {
	// Format: @risk-control:<name>-<id>
	// Example: @risk-control:auth-mfa-01
	// GxP Format: @risk-control:gxp-<name>

	parts := strings.TrimPrefix(tag, "@risk-control:")
	if len(parts) == 0 {
		return false
	}

	// GxP format: @risk-control:gxp-<name>
	if strings.HasPrefix(parts, "gxp-") {
		return len(parts) > 4 // At least "gxp-x"
	}

	// Standard format: must have dash and at least 2-digit numeric ID
	dashIndex := strings.LastIndex(parts, "-")
	if dashIndex == -1 {
		return false
	}

	controlName := parts[:dashIndex]
	scenarioID := parts[dashIndex+1:]

	// Scenario ID must be at least 2 characters and all digits
	if len(scenarioID) < 2 {
		return false
	}

	for _, ch := range scenarioID {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return len(controlName) > 0
}

// ValidateGxPRequirements validates GxP-specific requirements.
func ValidateGxPRequirements(test TestReference) []string {
	errors := []string{}

	// GxP requirements must have risk control
	if test.IsGxP {
		hasGxPRiskControl := false
		for _, rc := range test.RiskControls {
			if strings.HasPrefix(rc, "@risk-control:gxp-") {
				hasGxPRiskControl = true
				break
			}
		}

		if !hasGxPRiskControl {
			errors = append(errors, "GxP requirement must have @risk-control:gxp-<name> tag")
		}
	}

	// @gmp-critical-aspect must be used with @gxp
	if test.IsCriticalAspect && !test.IsGxP {
		errors = append(errors, "@gmp-critical-aspect must be used with @gxp tag")
	}

	return errors
}

// ValidateRiskControls validates risk control tags.
func ValidateRiskControls(test TestReference) []string {
	errors := []string{}

	for _, rc := range test.RiskControls {
		if !validateRiskControlTag(rc) {
			errors = append(errors, fmt.Sprintf("Invalid risk control tag format: %s", rc))
		}
	}

	return errors
}

// ParseRiskControlTag parses a risk control tag into components.
func ParseRiskControlTag(tag string) (*RiskControlRef, error) {
	if !strings.HasPrefix(tag, "@risk-control:") {
		return nil, fmt.Errorf("not a risk control tag: %s", tag)
	}

	parts := strings.TrimPrefix(tag, "@risk-control:")

	ref := &RiskControlRef{
		FullTag: tag,
	}

	// GxP format: @risk-control:gxp-<name>
	if strings.HasPrefix(parts, "gxp-") {
		ref.ControlName = parts
		ref.ScenarioID = ""
		ref.IsGxP = true
		return ref, nil
	}

	// Standard format: @risk-control:<name>-<id>
	dashIndex := strings.LastIndex(parts, "-")
	if dashIndex == -1 {
		return nil, fmt.Errorf("invalid risk control tag format: %s", tag)
	}

	ref.ControlName = parts[:dashIndex]
	ref.ScenarioID = parts[dashIndex+1:]
	ref.IsGxP = false

	return ref, nil
}
