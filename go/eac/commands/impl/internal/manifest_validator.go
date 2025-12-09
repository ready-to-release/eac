// Package internal provides build manifest validation against the contract schema
package internal

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed contracts/build-manifest.schema.json
var buildManifestSchema string

// ManifestValidator validates build manifests against the contract schema
type ManifestValidator struct {
	schema *jsonschema.Schema
}

// ManifestValidationError represents a manifest validation error
type ManifestValidationError struct {
	Moniker string
	Message string
	Details []string
}

func (e *ManifestValidationError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("manifest validation failed for %s: %s (%v)", e.Moniker, e.Message, e.Details)
	}
	return fmt.Sprintf("manifest validation failed for %s: %s", e.Moniker, e.Message)
}

// NewManifestValidator creates a new manifest validator with the embedded schema
func NewManifestValidator() (*ManifestValidator, error) {
	c := jsonschema.NewCompiler()

	// Parse the embedded schema
	var schemaDoc any
	if err := json.Unmarshal([]byte(buildManifestSchema), &schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to parse embedded schema: %w", err)
	}

	// Add schema to compiler
	schemaURL := "file:///build-manifest.schema.json"
	if err := c.AddResource(schemaURL, schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile the schema
	schema, err := c.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &ManifestValidator{schema: schema}, nil
}

// ValidateManifest validates a ModuleManifest against the contract schema
func (v *ManifestValidator) ValidateManifest(manifest *ModuleManifest) error {
	// Convert manifest to JSON for validation
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return v.ValidateJSON(jsonData, manifest.Moniker)
}

// ValidateJSON validates raw JSON data against the manifest schema
func (v *ManifestValidator) ValidateJSON(jsonData []byte, moniker string) error {
	// Parse JSON to generic interface
	var data any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate against schema
	if err := v.schema.Validate(data); err != nil {
		return &ManifestValidationError{
			Moniker: moniker,
			Message: err.Error(),
			Details: extractManifestValidationDetails(err),
		}
	}

	return nil
}

// extractManifestValidationDetails extracts detailed error messages from validation errors
func extractManifestValidationDetails(err error) []string {
	var details []string

	if validErr, ok := err.(*jsonschema.ValidationError); ok {
		details = append(details, validErr.Error())
		for _, cause := range validErr.Causes {
			details = append(details, extractManifestValidationDetails(cause)...)
		}
	}

	return details
}

// Global validator instance (lazy initialized)
var globalManifestValidator *ManifestValidator

// GetManifestValidator returns the global manifest validator instance
func GetManifestValidator() (*ManifestValidator, error) {
	if globalManifestValidator == nil {
		var err error
		globalManifestValidator, err = NewManifestValidator()
		if err != nil {
			return nil, err
		}
	}
	return globalManifestValidator, nil
}

// ValidateAndSave validates the manifest against the schema and saves it if valid
func (m *ModuleManifest) ValidateAndSave(moduleBuildDir string) error {
	validator, err := GetManifestValidator()
	if err != nil {
		return fmt.Errorf("failed to get manifest validator: %w", err)
	}

	if err := validator.ValidateManifest(m); err != nil {
		return err
	}

	return m.Save(moduleBuildDir)
}
