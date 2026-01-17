package oscal

import (
	"fmt"

	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// DocumentType represents the type of OSCAL document.
type DocumentType string

const (
	TypeCatalog DocumentType = "catalog"
	TypeProfile DocumentType = "profile"
)

// Validator is a generic OSCAL validator that delegates to specific validators.
type Validator struct {
	docType   DocumentType
	validator validation.Validator
}

// NewValidator creates a validator for the specified OSCAL document type.
func NewValidator(docType DocumentType) (*Validator, error) {
	v := &Validator{docType: docType}

	switch docType {
	case TypeCatalog:
		v.validator = &CatalogValidator{expectedVersion: "1.1.3"}
	case TypeProfile:
		v.validator = &ProfileValidator{expectedVersion: "1.1.2"}
	default:
		return nil, fmt.Errorf("unsupported OSCAL document type: %s", docType)
	}

	return v, nil
}

// Validate validates OSCAL output by delegating to the appropriate validator.
func (v *Validator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	if v.validator == nil {
		return []validation.ValidationError{
			*validation.NewValidationError(
				validation.ErrSetupError,
				fmt.Sprintf("OSCAL validator not initialized for document type: %s", v.docType),
				0,
			),
		}
	}
	return v.validator.Validate(output, context)
}

// VerifyImplementation checks if the validator is properly configured.
func (v *Validator) VerifyImplementation() []validation.ValidationError {
	if v.validator == nil {
		return []validation.ValidationError{
			*validation.NewValidationError(
				validation.ErrSetupError,
				"OSCAL validator not initialized",
				0,
			),
		}
	}
	return v.validator.VerifyImplementation()
}
