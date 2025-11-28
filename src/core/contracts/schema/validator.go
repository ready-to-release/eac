// Package schema provides JSON Schema validation for repository configuration files
package schema

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed schemas/*.json
var schemasFS embed.FS

// SchemaType represents the type of schema to validate against
type SchemaType string

const (
	SchemaModules         SchemaType = "modules"
	SchemaEnvironments    SchemaType = "environments"
	SchemaTestingTags     SchemaType = "testing-tags"
	SchemaTestingTaxonomy SchemaType = "testing-taxonomy"
)

// schemaFiles maps schema types to their embedded file paths
var schemaFiles = map[SchemaType]string{
	SchemaModules:         "schemas/modules.schema.json",
	SchemaEnvironments:    "schemas/environments.schema.json",
	SchemaTestingTags:     "schemas/testing-tags.schema.json",
	SchemaTestingTaxonomy: "schemas/testing-taxonomy.schema.json",
}

// Validator provides JSON Schema validation for repository configs
type Validator struct {
	compiler *jsonschema.Compiler
	schemas  map[SchemaType]*jsonschema.Schema
}

// ValidationError represents a schema validation error
type ValidationError struct {
	SchemaType SchemaType
	Path       string
	Message    string
	Details    []string
}

func (e *ValidationError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("schema validation failed for %s at %s: %s (%v)", e.SchemaType, e.Path, e.Message, e.Details)
	}
	return fmt.Sprintf("schema validation failed for %s at %s: %s", e.SchemaType, e.Path, e.Message)
}

// NewValidator creates a new schema validator with all schemas pre-compiled
func NewValidator() (*Validator, error) {
	c := jsonschema.NewCompiler()

	v := &Validator{
		compiler: c,
		schemas:  make(map[SchemaType]*jsonschema.Schema),
	}

	// Load and compile all schemas
	for schemaType, filePath := range schemaFiles {
		data, err := schemasFS.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded schema %s: %w", filePath, err)
		}

		// Parse the schema JSON
		var schemaDoc any
		if err := json.Unmarshal(data, &schemaDoc); err != nil {
			return nil, fmt.Errorf("failed to parse schema %s: %w", filePath, err)
		}

		// Add schema to compiler
		schemaURL := fmt.Sprintf("file:///%s", filePath)
		if err := c.AddResource(schemaURL, schemaDoc); err != nil {
			return nil, fmt.Errorf("failed to add schema resource %s: %w", filePath, err)
		}

		// Compile the schema
		schema, err := c.Compile(schemaURL)
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema %s: %w", filePath, err)
		}

		v.schemas[schemaType] = schema
	}

	return v, nil
}

// ValidateYAML validates YAML data against the specified schema
func (v *Validator) ValidateYAML(schemaType SchemaType, yamlData []byte) error {
	schema, ok := v.schemas[schemaType]
	if !ok {
		return fmt.Errorf("unknown schema type: %s", schemaType)
	}

	// Parse YAML to generic interface
	var data any
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert YAML maps to JSON-compatible maps (string keys)
	data = convertYAMLToJSON(data)

	// Validate against schema
	if err := schema.Validate(data); err != nil {
		return &ValidationError{
			SchemaType: schemaType,
			Path:       "",
			Message:    err.Error(),
			Details:    extractValidationDetails(err),
		}
	}

	return nil
}

// ValidateJSON validates JSON data against the specified schema
func (v *Validator) ValidateJSON(schemaType SchemaType, jsonData []byte) error {
	schema, ok := v.schemas[schemaType]
	if !ok {
		return fmt.Errorf("unknown schema type: %s", schemaType)
	}

	// Parse JSON to generic interface
	var data any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate against schema
	if err := schema.Validate(data); err != nil {
		return &ValidationError{
			SchemaType: schemaType,
			Path:       "",
			Message:    err.Error(),
			Details:    extractValidationDetails(err),
		}
	}

	return nil
}

// convertYAMLToJSON converts YAML-parsed data to JSON-compatible format
// YAML uses map[string]any but sometimes map[any]any, JSON needs map[string]any
func convertYAMLToJSON(data any) any {
	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any)
		for key, value := range v {
			result[key] = convertYAMLToJSON(value)
		}
		return result
	case map[any]any:
		result := make(map[string]any)
		for key, value := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertYAMLToJSON(value)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, value := range v {
			result[i] = convertYAMLToJSON(value)
		}
		return result
	default:
		return v
	}
}

// extractValidationDetails extracts detailed error messages from validation errors
func extractValidationDetails(err error) []string {
	var details []string

	// The jsonschema library provides detailed error info
	if validErr, ok := err.(*jsonschema.ValidationError); ok {
		details = append(details, validErr.Error())
		for _, cause := range validErr.Causes {
			details = append(details, extractValidationDetails(cause)...)
		}
	}

	return details
}

// GetSchemaTypes returns all available schema types
func GetSchemaTypes() []SchemaType {
	return []SchemaType{
		SchemaModules,
		SchemaEnvironments,
		SchemaTestingTags,
		SchemaTestingTaxonomy,
	}
}
