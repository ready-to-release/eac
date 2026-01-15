package json

import (
	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// Validator validates JSON against JSON Schema
type Validator struct {
	schemaPath string
}

// NewValidator creates a new JSON schema validator
func NewValidator(schemaPath string) (*Validator, error) {
	return &Validator{
		schemaPath: schemaPath,
	}, nil
}

// Validate validates JSON output against schema
func (v *Validator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	// Implementation will be moved from contracts.JSONSchemaValidator
	return nil
}

// VerifyImplementation checks if the validator is properly configured
func (v *Validator) VerifyImplementation() []validation.ValidationError {
	return nil
}
