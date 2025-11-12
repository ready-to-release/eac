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

	return false
}

// GetKnownTags returns all known tags
func GetKnownTags() []string {
	tags := []string{}
	for tag := range ValidTags {
		tags = append(tags, tag)
	}
	return tags
}
