// Package oscal provides OSCAL document loading and validation.
package oscal

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed schemas/oscal_catalog_schema_v1.1.3.json
var catalogSchemaJSON []byte

//go:embed schemas/oscal_profile_schema_v1.1.3.json
var profileSchemaJSON []byte

//go:embed schemas/oscal_assessment-results_schema_v1.1.3.json
var assessmentResultsSchemaJSON []byte

const (
	// Schema identifiers for OSCAL 1.1.3
	OSCALCatalogSchemaID           = "http://csrc.nist.gov/ns/oscal/1.1.3/oscal-catalog-schema.json"
	OSCALProfileSchemaID           = "http://csrc.nist.gov/ns/oscal/1.1.3/oscal-profile-schema.json"
	OSCALAssessmentResultsSchemaID = "http://csrc.nist.gov/ns/oscal/1.1.3/oscal-assessment-results-schema.json"
)

// SchemaValidator validates OSCAL documents against official JSON schemas.
type SchemaValidator struct {
	catalogSchema           *jsonschema.Schema
	profileSchema           *jsonschema.Schema
	assessmentResultsSchema *jsonschema.Schema
}

// NewSchemaValidator creates a new schema validator.
// Schemas are loaded lazily on first use to avoid unnecessary network requests.
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// ValidateCatalog validates an OSCAL catalog document against the official schema.
func (v *SchemaValidator) ValidateCatalog(filePath string) error {
	if v.catalogSchema == nil {
		// Parse embedded schema
		var schemaDoc any
		if err := json.Unmarshal(catalogSchemaJSON, &schemaDoc); err != nil {
			return fmt.Errorf("failed to parse catalog schema: %w", err)
		}

		// Compile the schema
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(OSCALCatalogSchemaID, schemaDoc); err != nil {
			return fmt.Errorf("failed to add catalog schema resource: %w", err)
		}

		schema, err := compiler.Compile(OSCALCatalogSchemaID)
		if err != nil {
			return fmt.Errorf("failed to compile catalog schema: %w", err)
		}
		v.catalogSchema = schema
	}

	return v.validateFile(filePath, v.catalogSchema)
}

// ValidateProfile validates an OSCAL profile document against the official schema.
func (v *SchemaValidator) ValidateProfile(filePath string) error {
	if v.profileSchema == nil {
		// Parse embedded schema
		var schemaDoc any
		if err := json.Unmarshal(profileSchemaJSON, &schemaDoc); err != nil {
			return fmt.Errorf("failed to parse profile schema: %w", err)
		}

		// Compile the schema
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(OSCALProfileSchemaID, schemaDoc); err != nil {
			return fmt.Errorf("failed to add profile schema resource: %w", err)
		}

		schema, err := compiler.Compile(OSCALProfileSchemaID)
		if err != nil {
			return fmt.Errorf("failed to compile profile schema: %w", err)
		}
		v.profileSchema = schema
	}

	return v.validateFile(filePath, v.profileSchema)
}

// ValidateAssessmentResults validates an OSCAL assessment-results document against the official schema.
func (v *SchemaValidator) ValidateAssessmentResults(filePath string) error {
	if v.assessmentResultsSchema == nil {
		// Parse embedded schema
		var schemaDoc any
		if err := json.Unmarshal(assessmentResultsSchemaJSON, &schemaDoc); err != nil {
			return fmt.Errorf("failed to parse assessment-results schema: %w", err)
		}

		// Compile the schema
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(OSCALAssessmentResultsSchemaID, schemaDoc); err != nil {
			return fmt.Errorf("failed to add assessment-results schema resource: %w", err)
		}

		schema, err := compiler.Compile(OSCALAssessmentResultsSchemaID)
		if err != nil {
			return fmt.Errorf("failed to compile assessment-results schema: %w", err)
		}
		v.assessmentResultsSchema = schema
	}

	return v.validateFile(filePath, v.assessmentResultsSchema)
}

// validateFile validates a JSON file against a compiled schema.
func (v *SchemaValidator) validateFile(filePath string, schema *jsonschema.Schema) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var document interface{}
	if err := json.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if err := schema.Validate(document); err != nil {
		return err
	}

	return nil
}
