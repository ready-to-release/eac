package testing

import (
	"fmt"
	"strings"
)

// ValidTags defines all known tags from contracts/testing/0.1.0/tags.yml
var ValidTags = map[string]bool{
	// System dependencies
	"@dep:docker": true,
	"@dep:git":    true,
	"@dep:go":     true,
	"@dep:claude": true,
	"@dep:az-cli": true,

	// Taxonomy levels
	"@L0": true,
	"@L1": true,
	"@L2": true,
	"@L3": true,
	"@L4": true,

	// Verification types
	"@ov":  true,
	"@iv":  true,
	"@pv":  true,
	"@piv": true,
	"@ppv": true,

	// Safety tags
	"@requires_isolation": true,

	// Execution control tags
	"@ignore": true,
	"@Manual": true,

	// GxP regulatory tags
	"@gxp":             true,
	"@critical-aspect": true,
}

// LevelTags are taxonomy level tags
var LevelTags = []string{"@L0", "@L1", "@L2", "@L3", "@L4"}

// VerificationTags are verification type tags
var VerificationTags = []string{"@ov", "@iv", "@pv", "@piv", "@ppv"}

// ValidTestTypes are the allowed test types
var ValidTestTypes = []string{"gotest", "godog"}

// ValidateTags checks if tags are valid and don't conflict
func ValidateTags(tags []string) []string {
	errors := []string{}

	// Check for invalid tags
	for _, tag := range tags {
		if !IsValidTag(tag) {
			errors = append(errors, fmt.Sprintf("tag %s is not defined in contracts/testing/0.1.0/tags.yml", tag))
		}
	}

	// Check for multiple level tags
	levelCount := 0
	for _, tag := range tags {
		if contains(LevelTags, tag) {
			levelCount++
		}
	}
	if levelCount > 1 {
		errors = append(errors, "test has multiple level tags (only one of @L0-@L4 allowed)")
	}

	// Check for multiple verification tags
	verificationCount := 0
	for _, tag := range tags {
		if contains(VerificationTags, tag) {
			verificationCount++
		}
	}
	if verificationCount > 1 {
		errors = append(errors, "test has multiple verification tags (only one of @ov/@iv/@pv/@piv/@ppv allowed)")
	}

	return errors
}

// ValidateTestReference validates a complete test reference
func ValidateTestReference(test TestReference) []string {
	errors := []string{}

	// Validate test type
	if !contains(ValidTestTypes, test.Type) {
		errors = append(errors, fmt.Sprintf("invalid test type: %s (must be gotest or godog)", test.Type))
	}

	// Validate tags
	tagErrors := ValidateTags(test.Tags)
	errors = append(errors, tagErrors...)

	return errors
}

// IsValidTag checks if a tag is valid according to contracts
func IsValidTag(tag string) bool {
	// Check known tags
	if ValidTags[tag] {
		return true
	}

	// Check @risk:* pattern
	if strings.HasPrefix(tag, "@risk:") {
		return true
	}

	// Check @risk-control:* pattern
	if strings.HasPrefix(tag, "@risk-control:") {
		return validateRiskControlTag(tag)
	}

	return false
}

// validateRiskControlTag validates @risk-control:<name>-<id> format
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

// GetKnownTags returns all known tags
func GetKnownTags() []string {
	tags := []string{}
	for tag := range ValidTags {
		tags = append(tags, tag)
	}
	return tags
}

// ValidateGxPRequirements validates GxP-specific requirements
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

	// @critical-aspect must be used with @gxp
	if test.IsCriticalAspect && !test.IsGxP {
		errors = append(errors, "@critical-aspect must be used with @gxp tag")
	}

	return errors
}

// ValidateRiskControls validates risk control tags
func ValidateRiskControls(test TestReference) []string {
	errors := []string{}

	for _, rc := range test.RiskControls {
		if !validateRiskControlTag(rc) {
			errors = append(errors, fmt.Sprintf("Invalid risk control tag format: %s", rc))
		}
	}

	return errors
}

// ParseRiskControlTag parses a risk control tag into components
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
