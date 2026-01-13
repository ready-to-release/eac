package oscal

import (
	"encoding/json"
	"fmt"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// CatalogValidator validates OSCAL catalog documents
type CatalogValidator struct {
	expectedVersion string
}

// Validate validates an OSCAL catalog document
// The output parameter should contain the JSON content of the OSCAL catalog
// Context can optionally include:
//   - "file_path": path to the file being validated (for better error messages)
func (v *CatalogValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	var errors []validation.ValidationError

	// Check for empty output
	if output == "" {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrEmptyOutput,
			"OSCAL catalog content is empty",
			0,
		))
		return errors
	}

	// Parse JSON using go-oscal wrapper type
	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal([]byte(output), &oscalDoc); err != nil {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrInvalidJSON,
			fmt.Sprintf("invalid JSON: %v", err),
			0,
		))
		return errors
	}

	// Check if it's a catalog document
	if oscalDoc.Catalog == nil {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALInvalidDocument,
			"not an OSCAL catalog document",
			0,
		))
		return errors
	}

	catalog := oscalDoc.Catalog

	// Validate required UUID field
	if catalog.UUID == "" {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALMissingUUID,
			"missing required field: uuid",
			0,
		))
	}

	// Validate metadata title
	if catalog.Metadata.Title == "" {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALMissingTitle,
			"missing required field: title",
			0,
		))
	}

	// Validate metadata last-modified
	if catalog.Metadata.LastModified.IsZero() {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALMissingLastModified,
			"missing required field: last-modified",
			0,
		))
	}

	// Validate that catalog has at least one control or group
	hasControls := catalog.Controls != nil && len(*catalog.Controls) > 0
	hasGroups := catalog.Groups != nil && len(*catalog.Groups) > 0

	if !hasControls && !hasGroups {
		errors = append(errors, *validation.NewValidationError(
			validation.ErrOSCALEmptyCatalog,
			"catalog must have at least one control or group",
			0,
		))
	}

	return errors
}

// VerifyImplementation checks if the validator is properly configured
func (v *CatalogValidator) VerifyImplementation() []validation.ValidationError {
	// No external dependencies to verify for catalog validator
	return nil
}
