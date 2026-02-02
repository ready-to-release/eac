// Package oscal provides validators for OSCAL (Open Security Controls Assessment Language) documents.
package oscal

import (
	"encoding/json"
	"fmt"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/core/validation"
)

// CatalogValidator validates OSCAL catalog documents.
type CatalogValidator struct {
	expectedVersion string
}

// Validate validates an OSCAL catalog document
// The output parameter should contain the JSON content of the OSCAL catalog
// Context can optionally include:
//   - "file_path": path to the file being validated (for better error messages)
func (v *CatalogValidator) Validate(output string, _ map[string]interface{}) []validation.ValidationError {
	var errors []validation.ValidationError
	formatter := validation.NewErrorFormatter()

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
		// Enhanced error for invalid JSON
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrInvalidJSON,
			fmt.Sprintf("Invalid JSON format: %v", err),
			formatter.TruncateJSON(output, 10),
			"Valid JSON with proper syntax",
			`{
  "catalog": {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": { ... },
    ...
  }
}`,
			"Check for common JSON issues:\n• Missing or extra commas\n• Unmatched quotes or brackets\n• Invalid escape sequences",
			0,
		))
		return errors
	}

	// Check if it's a catalog document
	if oscalDoc.Catalog == nil {
		// Enhanced error for missing catalog
		actualFields := formatter.ExtractJSONFields(output)
		errors = append(errors, *formatter.FormatStructuredError(
			validation.ErrOSCALInvalidDocument,
			"Not an OSCAL catalog document - missing 'catalog' root field",
			`{
  "catalog": {
    "uuid": "...",
    "metadata": { ... },
    "controls": [ ... ]
  }
}`,
			actualFields,
			0,
		))
		return errors
	}

	catalog := oscalDoc.Catalog

	// Validate required UUID field
	if catalog.UUID == "" {
		errors = append(errors, *formatter.FormatMissingFieldError(
			validation.ErrOSCALMissingUUID,
			"uuid",
			"catalog",
			[]string{"metadata", "controls"},
			`{
  "catalog": {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": { ... }
  }
}`,
			0,
		))
	}

	// Validate metadata title
	if catalog.Metadata.Title == "" {
		errors = append(errors, *formatter.FormatMissingFieldError(
			validation.ErrOSCALMissingTitle,
			"metadata.title",
			"catalog",
			[]string{},
			`{
  "catalog": {
    "uuid": "...",
    "metadata": {
      "title": "My Security Controls Catalog",
      "last-modified": "2026-01-14T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.3"
    }
  }
}`,
			0,
		))
	}

	// Validate metadata last-modified
	if catalog.Metadata.LastModified.IsZero() {
		errors = append(errors, *formatter.FormatMissingFieldError(
			validation.ErrOSCALMissingLastModified,
			"metadata.last-modified",
			"catalog",
			[]string{},
			`{
  "catalog": {
    "uuid": "...",
    "metadata": {
      "title": "...",
      "last-modified": "2026-01-14T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.3"
    }
  }
}`,
			0,
		))
	}

	// Validate that catalog has at least one control or group
	hasControls := catalog.Controls != nil && len(*catalog.Controls) > 0
	hasGroups := catalog.Groups != nil && len(*catalog.Groups) > 0

	if !hasControls && !hasGroups {
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrOSCALEmptyCatalog,
			"Catalog must have at least one control or group",
			"",
			"Catalog with controls or groups",
			`{
  "catalog": {
    "uuid": "...",
    "metadata": { ... },
    "controls": [
      {
        "id": "ac-1",
        "title": "Access Control Policy",
        "params": [],
        "parts": []
      }
    ]
  }
}

OR with groups:

{
  "catalog": {
    "uuid": "...",
    "metadata": { ... },
    "groups": [
      {
        "id": "access-control",
        "title": "Access Control",
        "controls": [ ... ]
      }
    ]
  }
}`,
			"Add at least one control using the 'controls' array, or organize controls into groups using the 'groups' array.",
			0,
		))
	}

	return errors
}

// VerifyImplementation checks if the validator is properly configured.
func (v *CatalogValidator) VerifyImplementation() []validation.ValidationError {
	// No external dependencies to verify for catalog validator
	return nil
}
