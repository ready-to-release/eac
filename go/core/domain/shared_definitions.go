// Package contracts provides shared type definitions that must match JSON schemas.
// The schema (shared-definitions.schema.json) is the source of truth for config validation.
// Go constants are hardcoded for performance but validated against schema in tests.
package domain

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Scanner category constants - must match shared-definitions.schema.json
var validScannerCategories = map[string]bool{
	"sbom":       true,
	"vuln":       true,
	"secrets":    true,
	"iac":        true,
	"compliance": true,
	"sast":       true,
	"zap":        true,
}

// Server type constants - must match shared-definitions.schema.json
var validServerTypes = map[string]bool{
	"static-site":      true,
	"mkdocs-live":      true,
	"structurizr-lite": true,
}

// Platform constants - must match shared-definitions.schema.json
var validPlatforms = map[string]bool{
	"linux":   true,
	"darwin":  true,
	"windows": true,
}

// Severity constants - must match shared-definitions.schema.json
var validSeverities = map[string]bool{
	"LOW":      true,
	"MEDIUM":   true,
	"HIGH":     true,
	"CRITICAL": true,
}

// copyMap returns a shallow copy of the given map, protecting the original from mutation.
func copyMap(m map[string]bool) map[string]bool {
	c := make(map[string]bool, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

// ValidScannerCategories returns a copy of the valid scanner categories.
func ValidScannerCategories() map[string]bool {
	return copyMap(validScannerCategories)
}

// IsValidScannerCategory returns true if the category is valid.
func IsValidScannerCategory(category string) bool {
	return validScannerCategories[category]
}

// ValidServerTypes returns a copy of the valid server types.
func ValidServerTypes() map[string]bool {
	return copyMap(validServerTypes)
}

// IsValidServerType returns true if the server type is valid.
func IsValidServerType(serverType string) bool {
	return validServerTypes[serverType]
}

// ValidPlatforms returns a copy of the valid platforms.
func ValidPlatforms() map[string]bool {
	return copyMap(validPlatforms)
}

// IsValidPlatform returns true if the platform is valid.
func IsValidPlatform(platform string) bool {
	return validPlatforms[platform]
}

// ValidSeverities returns a copy of the valid severity levels.
func ValidSeverities() map[string]bool {
	return copyMap(validSeverities)
}

// IsValidSeverity returns true if the severity is valid.
func IsValidSeverity(severity string) bool {
	return validSeverities[severity]
}

// schemaDefinitions represents the JSON schema $defs structure for validation.
type schemaDefinitions struct {
	Defs map[string]schemaEnum `json:"$defs"`
}

type schemaEnum struct {
	Enum []string `json:"enum"`
}

// ValidateGoMatchesSchema validates that Go constants match the JSON schema.
// Returns an error with details if there's a mismatch.
// This should be called in tests to ensure Go and schema stay in sync.
func ValidateGoMatchesSchema() error {
	data, err := core.FS.ReadFile(core.SchemaPath("shared-definitions.schema.json"))
	if err != nil {
		return fmt.Errorf("failed to read shared-definitions.schema.json: %w", err)
	}

	var schema schemaDefinitions
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Errorf("failed to parse shared-definitions.schema.json: %w", err)
	}

	var mismatches []string

	// Validate scanner categories
	if err := validateEnumMatch("scanner_category", schema.Defs, validScannerCategories); err != nil {
		mismatches = append(mismatches, err.Error())
	}

	// Validate server types
	if err := validateEnumMatch("server_type", schema.Defs, validServerTypes); err != nil {
		mismatches = append(mismatches, err.Error())
	}

	// Validate platforms
	if err := validateEnumMatch("platform", schema.Defs, validPlatforms); err != nil {
		mismatches = append(mismatches, err.Error())
	}

	// Validate severities
	if err := validateEnumMatch("severity", schema.Defs, validSeverities); err != nil {
		mismatches = append(mismatches, err.Error())
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("Go constants do not match shared-definitions.schema.json:\n%s\n\nFix: Update Go constants in shared_definitions.go to match the schema",
			strings.Join(mismatches, "\n"))
	}

	return nil
}

// validateEnumMatch checks if Go map matches schema enum.
func validateEnumMatch(defName string, defs map[string]schemaEnum, goMap map[string]bool) error {
	schemaDef, ok := defs[defName]
	if !ok {
		return fmt.Errorf("  %s: missing from schema", defName)
	}

	schemaSet := make(map[string]bool, len(schemaDef.Enum))
	for _, v := range schemaDef.Enum {
		schemaSet[v] = true
	}

	// Check for values in Go but not in schema
	var inGoOnly []string
	for k := range goMap {
		if !schemaSet[k] {
			inGoOnly = append(inGoOnly, k)
		}
	}

	// Check for values in schema but not in Go
	var inSchemaOnly []string
	for k := range schemaSet {
		if !goMap[k] {
			inSchemaOnly = append(inSchemaOnly, k)
		}
	}

	if len(inGoOnly) > 0 || len(inSchemaOnly) > 0 {
		sort.Strings(inGoOnly)
		sort.Strings(inSchemaOnly)
		var parts []string
		if len(inGoOnly) > 0 {
			parts = append(parts, fmt.Sprintf("in Go only: %v", inGoOnly))
		}
		if len(inSchemaOnly) > 0 {
			parts = append(parts, fmt.Sprintf("in schema only: %v", inSchemaOnly))
		}
		return fmt.Errorf("  %s: %s", defName, strings.Join(parts, "; "))
	}

	return nil
}
