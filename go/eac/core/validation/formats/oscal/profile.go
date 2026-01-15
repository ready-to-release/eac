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
	formatter := validation.NewErrorFormatter()

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
		// Enhanced error for invalid JSON
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrInvalidJSON,
			fmt.Sprintf("Invalid OSCAL profile JSON: %v", err),
			formatter.TruncateJSON(output, 10),
			"Valid JSON with proper syntax",
			`{
  "profile": {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": { ... },
    "imports": [ ... ]
  }
}`,
			"Check for common JSON issues:\n• Missing or extra commas\n• Unmatched quotes or brackets\n• Invalid escape sequences",
			0,
		))
		return errors
	}

	// Check if document contains a profile
	if oscalDoc.Profile == nil {
		// Enhanced error for missing profile
		actualFields := formatter.ExtractJSONFields(output)
		errors = append(errors, *formatter.FormatStructuredError(
			validation.ErrOSCALInvalidDocument,
			"Not an OSCAL profile document - missing 'profile' root field",
			`{
  "profile": {
    "uuid": "...",
    "metadata": { ... },
    "imports": [ ... ]
  }
}`,
			actualFields,
			0,
		))
		return errors
	}

	profile := oscalDoc.Profile

	// Validate required UUID field
	if profile.UUID == "" {
		errors = append(errors, *formatter.FormatMissingFieldError(
			validation.ErrOSCALMissingUUID,
			"uuid",
			"profile",
			[]string{"metadata", "imports"},
			`{
  "profile": {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": { ... },
    "imports": [ ... ]
  }
}`,
			0,
		))
	}

	// Validate metadata title
	if profile.Metadata.Title == "" {
		errors = append(errors, *formatter.FormatMissingFieldError(
			validation.ErrOSCALMissingTitle,
			"metadata.title",
			"profile",
			[]string{},
			`{
  "profile": {
    "uuid": "...",
    "metadata": {
      "title": "My Security Controls Profile",
      "last-modified": "2026-01-14T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.2"
    }
  }
}`,
			0,
		))
	}

	// Validate metadata last-modified
	if profile.Metadata.LastModified.IsZero() {
		errors = append(errors, *formatter.FormatMissingFieldError(
			validation.ErrOSCALMissingLastModified,
			"metadata.last-modified",
			"profile",
			[]string{},
			`{
  "profile": {
    "uuid": "...",
    "metadata": {
      "title": "...",
      "last-modified": "2026-01-14T00:00:00Z",
      "version": "1.0.0",
      "oscal-version": "1.1.2"
    }
  }
}`,
			0,
		))
	}

	// Validate imports
	if len(profile.Imports) == 0 {
		errors = append(errors, *formatter.FormatEnhancedError(
			validation.ErrOSCALMissingImports,
			"Missing required field: imports - profile must import at least one catalog",
			"",
			"Profile with imports array",
			`{
  "profile": {
    "uuid": "...",
    "metadata": { ... },
    "imports": [
      {
        "href": "https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json",
        "include-controls": [
          {
            "with-ids": ["ac-1", "ac-2", "au-1"]
          }
        ]
      }
    ]
  }
}`,
			"Add an 'imports' array with at least one import object containing:\n• href: URL to the catalog to import from\n• include-controls: array specifying which controls to include",
			0,
		))
		return errors
	}

	// Validate each import
	for i, imp := range profile.Imports {
		if imp.Href == "" {
			errors = append(errors, *formatter.FormatEnhancedError(
				validation.ErrOSCALMissingImportHref,
				fmt.Sprintf("Import[%d] missing required field: href", i),
				"",
				"Import with href to catalog",
				`{
  "imports": [
    {
      "href": "https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json",
      "include-controls": [
        {
          "with-ids": ["ac-1", "ac-2"]
        }
      ]
    }
  ]
}`,
				"Add an 'href' field with a URL pointing to the OSCAL catalog to import from.",
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
							errors = append(errors, *formatter.FormatEnhancedError(
								validation.ErrOSCALInvalidControlID,
								fmt.Sprintf("Import[%d].include-controls[%d].with-ids[%d]: Invalid control ID '%s'", i, j, k, id),
								fmt.Sprintf(`"with-ids": ["%s"]`, id),
								"Valid NIST 800-53 control ID format",
								`Valid examples:
  "ac-1"   - Access Control Policy
  "au-2"   - Audit Events
  "sc-28"  - Protection of Information at Rest
  "si-10"  - Information Input Validation

Format: <family>-<number>
  family: 2 lowercase letters (ac, au, cm, ia, etc.)
  number: 1 or more digits`,
								"Use lowercase NIST 800-53 control IDs in the format 'xx-n' (e.g., 'ac-1', 'au-2').",
								0,
							))
						}
					}
				}
			}
		}

		if !hasControls {
			errors = append(errors, *formatter.FormatEnhancedError(
				validation.ErrOSCALEmptyImport,
				fmt.Sprintf("Import[%d] has no control selections - must specify which controls to include", i),
				"",
				"Import with control selections",
				`{
  "href": "...",
  "include-controls": [
    {
      "with-ids": ["ac-1", "ac-2", "au-1", "au-2"]
    }
  ]
}`,
				"Add an 'include-controls' array with at least one entry containing 'with-ids' to specify which controls to import.",
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
