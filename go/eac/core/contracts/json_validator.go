package contracts

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// JSONSchemaValidator validates JSON against a schema
type JSONSchemaValidator struct {
	SchemaPath string
	Schema     *gojsonschema.Schema
}

// NewJSONSchemaValidator creates a JSON schema validator
func NewJSONSchemaValidator(schemaPath string) (*JSONSchemaValidator, error) {
	// Convert file path to proper file:// URL
	fileURL := pathToFileURL(schemaPath)
	schemaLoader := gojsonschema.NewReferenceLoader(fileURL)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return &JSONSchemaValidator{
		SchemaPath: schemaPath,
		Schema:     schema,
	}, nil
}

// Validate validates JSON content against schema (implements Validator interface)
func (v *JSONSchemaValidator) Validate(output string, context map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Parse JSON
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		return []ValidationError{
			*NewValidationError(ErrInvalidJSON, fmt.Sprintf("Invalid JSON: %v", err), 0),
		}
	}

	// Validate against schema
	documentLoader := gojsonschema.NewGoLoader(jsonData)
	result, err := v.Schema.Validate(documentLoader)
	if err != nil {
		return []ValidationError{
			*NewValidationError(ErrJSONSchemaViolation, fmt.Sprintf("Schema validation failed: %v", err), 0),
		}
	}

	// Convert schema errors to ValidationErrors
	if !result.Valid() {
		for _, schemaErr := range result.Errors() {
			errors = append(errors, *NewValidationError(
				ErrJSONSchemaViolation,
				fmt.Sprintf("%s: %s", schemaErr.Field(), schemaErr.Description()),
				0,
			))
		}
	}

	return errors
}

// VerifyImplementation verifies the validator is ready (implements Validator interface)
func (v *JSONSchemaValidator) VerifyImplementation() []ValidationError {
	// No special verification needed for JSON schema validator
	return []ValidationError{}
}

// pathToFileURL converts a file path to a proper file:// URL
// This handles Windows paths correctly by converting backslashes to forward slashes
// and ensuring the proper URL format
func pathToFileURL(path string) string {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't get absolute path, use as-is
		absPath = path
	}

	// Convert backslashes to forward slashes for URL
	urlPath := filepath.ToSlash(absPath)

	// On Windows, paths start with drive letter (e.g., C:/)
	// Need to ensure proper file:// URL format: file:///C:/path
	if runtime.GOOS == "windows" {
		// Remove leading slash if present (ToSlash might add it)
		urlPath = strings.TrimPrefix(urlPath, "/")
		return "file:///" + urlPath
	}

	// On Unix-like systems, path already starts with /
	return "file://" + urlPath
}
