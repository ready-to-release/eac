package validator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// Embed the r2r-cli-config v0.1.0 schema at compile time// The schema is copied from contracts/cli/0.1.0/config.json by build-cli.ps1//
//go:embed config/schema.json
var embeddedSchema string

// EmbeddedValidator validates configurations using the embedded JSON schema
type EmbeddedValidator struct {
	schema *gojsonschema.Schema
}

// NewEmbeddedValidator creates a validator using the embedded schema
func NewEmbeddedValidator() (*EmbeddedValidator, error) {
	// Load the embedded schema
	schemaLoader := gojsonschema.NewStringLoader(embeddedSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to compile embedded schema: %w", err)
	}

	return &EmbeddedValidator{
		schema: schema,
	}, nil
}

// ValidateJSON validates a JSON document against the embedded schema
func (v *EmbeddedValidator) ValidateJSON(jsonData []byte) (*ValidationResult, error) {
	// Create document loader from JSON bytes
	documentLoader := gojsonschema.NewBytesLoader(jsonData)
	
	// Validate against schema
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Convert to our ValidationResult format
	valResult := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			valResult.Errors = append(valResult.Errors, ValidationError{
				Field:    err.Field(),
				Rule:     err.Type(),
				Message:  err.Description(),
				Value:    err.Value(),
				Expected: formatExpected(err),
			})
		}
	}

	return valResult, nil
}

// ValidateInterface validates a Go struct/map against the embedded schema
func (v *EmbeddedValidator) ValidateInterface(config interface{}) (*ValidationResult, error) {
	// Convert the config to JSON
	// We need to handle the struct differently than a map
	
	// If it's already a map (from Viper), use it directly
	if configMap, ok := config.(map[string]interface{}); ok {
		jsonData, err := json.Marshal(configMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config map to JSON: %w", err)
		}
		return v.ValidateJSON(jsonData)
	}
	
	// Otherwise, we need to convert the struct to a map with proper field names
	// For now, just marshal directly (this will use PascalCase for struct fields)
	jsonData, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	return v.ValidateJSON(jsonData)
}

// formatExpected formats the expected value from a validation error
func formatExpected(err gojsonschema.ResultError) string {
	details := err.Details()
	
	// Try to extract useful expected information
	if enum, ok := details["enum"]; ok {
		if enumValues, ok := enum.([]interface{}); ok {
			values := make([]string, len(enumValues))
			for i, v := range enumValues {
				values[i] = fmt.Sprintf("%v", v)
			}
			return fmt.Sprintf("one of [%s]", strings.Join(values, ", "))
		}
	}
	
	if pattern, ok := details["pattern"]; ok {
		return fmt.Sprintf("match pattern %v", pattern)
	}
	
	if minimum, ok := details["minimum"]; ok {
		return fmt.Sprintf(">= %v", minimum)
	}
	
	if maximum, ok := details["maximum"]; ok {
		return fmt.Sprintf("<= %v", maximum)
	}
	
	if format, ok := details["format"]; ok {
		return fmt.Sprintf("format: %v", format)
	}
	
	if expectedType, ok := details["expected"]; ok {
		return fmt.Sprintf("type: %v", expectedType)
	}
	
	return err.Type()
}

// GetEmbeddedSchemaVersion returns the version of the embedded schema
func GetEmbeddedSchemaVersion() string {
	// Parse the embedded schema to extract version
	var schemaMap map[string]interface{}
	if err := json.Unmarshal([]byte(embeddedSchema), &schemaMap); err != nil {
		return "unknown"
	}
	
	// Try to extract version from $id or custom version field
	if id, ok := schemaMap["$id"].(string); ok {
		// Extract version from ID path
		if strings.Contains(id, "/v1.0/") {
			return "v1.0"
		}
	}
	
	return "v1.0" // Default version
}

// GetEmbeddedSchema returns the raw embedded schema for inspection
func GetEmbeddedSchema() string {
	return embeddedSchema
}