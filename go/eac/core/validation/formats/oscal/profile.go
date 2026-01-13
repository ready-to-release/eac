package oscal

import (
	"encoding/json"
	"fmt"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// ProfileValidator validates OSCAL profile documents
type ProfileValidator struct {
	expectedVersion string
}

// Validate validates an OSCAL profile document
// The output parameter should contain the JSON content of the OSCAL profile
// Context can optionally include:
//   - "file_path": path to the file being validated (for better error messages)
func (v *ProfileValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	var errors []validation.ValidationError

	// Check for empty output
	if output == "" {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrEmptyOutput,
			"OSCAL profile content is empty",
			0,
		))
		return errors
	}

	// Parse JSON using go-oscal types
	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal([]byte(output), &oscalDoc); err != nil {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrInvalidJSON,
			fmt.Sprintf("invalid OSCAL profile JSON: %v", err),
			0,
		))
		return errors
	}

	// Check if document contains a profile
	if oscalDoc.Profile == nil {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALInvalidDocument,
			"document does not contain a profile",
			0,
		))
		return errors
	}

	profile := oscalDoc.Profile

	// Validate required UUID field
	if profile.UUID == "" {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALMissingUUID,
			"missing required field: uuid",
			0,
		))
	}

	// Validate metadata title
	if profile.Metadata.Title == "" {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALMissingTitle,
			"missing required field: metadata.title",
			0,
		))
	}

	// Validate metadata last-modified
	if profile.Metadata.LastModified.IsZero() {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALMissingLastModified,
			"missing required field: metadata.last-modified",
			0,
		))
	}

	// Validate imports
	if len(profile.Imports) == 0 {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALMissingImports,
			"missing required field: imports (must have at least one)",
			0,
		))
		return errors
	}

	// Validate each import
	for i, imp := range profile.Imports {
		if imp.Href == "" {
			errors = append(errors, *validation.NewValidationError(
				validation.ErrOSCALMissingImportHref,
				fmt.Sprintf("import[%d] missing required field: href", i),
				0,
			))
		}

		// Check if there are controls included
		hasControls := false
		if imp.IncludeControls != nil && len(*imp.IncludeControls) > 0 {
			hasControls = true
			// Validate control IDs
			for j, ctrl := range *imp.IncludeControls {
				if ctrl.WithIds != nil {
					for k, id := range *ctrl.WithIds {
						if !IsValidControlID(id) {
							errors = append(errors, *validation.NewValidationError(
								validation.ErrOSCALInvalidControlID,
								fmt.Sprintf("import[%d].include-controls[%d].with-ids[%d]: control ID '%s' may not be valid NIST 800-53 format", i, j, k, id),
								0,
							))
						}
					}
				}
			}
		}

		if !hasControls {
			errors = append(errors, *validation.NewValidationError(
				validation.ErrOSCALEmptyImport,
				fmt.Sprintf("import[%d] has no control selections", i),
				0,
			))
		}
	}

	// Validate OSCAL version if present (warning only)
	if profile.Metadata.OscalVersion != "" {
		if profile.Metadata.OscalVersion != v.expectedVersion {
			errors = append(errors, *validation.NewValidationError(
				validation.ErrOSCALVersionMismatch,
				fmt.Sprintf("OSCAL version %s differs from expected %s", profile.Metadata.OscalVersion, v.expectedVersion),
				0,
			))
		}
	}

	return errors
}

// VerifyImplementation checks if the validator is properly configured
func (v *ProfileValidator) VerifyImplementation() []validation.ValidationError {
	// No external dependencies to verify for profile validator
	return nil
}

// IsValidControlID checks if a control ID is valid NIST 800-53 format.
func IsValidControlID(id string) bool {
	parts := strings.Split(strings.ToLower(id), "-")
	if len(parts) < 2 {
		return false
	}

	// Family prefix should be 2 letters
	family := parts[0]
	if len(family) != 2 {
		return false
	}

	for _, c := range family {
		if c < 'a' || c > 'z' {
			return false
		}
	}

	// Number part should be non-empty and numeric
	number := parts[1]
	if len(number) == 0 {
		return false
	}

	for _, c := range number {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
