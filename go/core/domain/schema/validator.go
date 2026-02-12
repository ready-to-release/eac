// Package schema provides JSON Schema validation for repository configuration files
package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// SchemaType represents the type of schema to validate against.
type SchemaType string

// Schema type constants for different configuration file types.
const (
	SchemaEnvironments SchemaType = "environments"
	SchemaTestingTags    SchemaType = "testing-tags"
	SchemaTestSuites     SchemaType = "test-suites"
	SchemaRepository     SchemaType = "repository"
	SchemaAIProvider     SchemaType = "ai-provider"
	SchemaBooks          SchemaType = "books"
	SchemaCommands       SchemaType = "commands"
	SchemaToolConfig     SchemaType = "tool-config"
	SchemaBlueprints     SchemaType = "blueprints"
)

// schemaFileNames maps schema types to their file names (without path).
var schemaFileNames = map[SchemaType]string{
	SchemaEnvironments: "environments.schema.json",
	SchemaTestingTags:    "testing-tags.schema.json",
	SchemaTestSuites:     "test-suites.schema.json",
	SchemaRepository:     "repository.schema.json",
	SchemaAIProvider:     "ai-provider.schema.json",
	SchemaBooks:          "books.schema.json",
	SchemaCommands:       "commands.schema.json",
	SchemaToolConfig:     "tool-config.schema.json",
	SchemaBlueprints:     "blueprints.schema.json",
}

// ContractVersion is the schema contract version.
const ContractVersion = "0.1.0"

// Validator provides JSON Schema validation for repository configs.
type Validator struct {
	compiler      *jsonschema.Compiler
	schemas       map[SchemaType]*jsonschema.Schema
	workspaceRoot string
}

// ValidationError represents a schema validation error.
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

// NewValidator creates a new schema validator with schemas loaded from the embedded contract FS.
// workspaceRoot is retained for backward compatibility (used by GetSchemaPath) but schemas
// are compiled once per process via FactoryValidator and shared across all instances.
func NewValidator(workspaceRoot string) (*Validator, error) {
	v, err := FactoryValidator()
	if err != nil {
		return nil, err
	}
	// Return a thin wrapper that adds workspaceRoot for GetSchemaPath()
	// while sharing the compiled schemas from the factory singleton.
	return &Validator{
		compiler:      v.compiler,
		schemas:       v.schemas,
		workspaceRoot: workspaceRoot,
	}, nil
}

// ValidateYAML validates YAML data against the specified schema.
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

// ValidateJSON validates JSON data against the specified schema.
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
// YAML uses map[string]any but sometimes map[any]any, JSON needs map[string]any.
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

// extractValidationDetails extracts detailed error messages from validation errors.
func extractValidationDetails(err error) []string {
	var details []string

	// The jsonschema library provides detailed error info
	validErr := &jsonschema.ValidationError{}
	if errors.As(err, &validErr) {
		details = append(details, validErr.Error())
		for _, cause := range validErr.Causes {
			details = append(details, extractValidationDetails(cause)...)
		}
	}

	return details
}

// GetSchemaTypes returns all available schema types.
// The list is auto-generated from the schemaFileNames map keys to avoid
// hand-maintained lists drifting out of sync.
func GetSchemaTypes() []SchemaType {
	result := make([]SchemaType, 0, len(schemaFileNames))
	for k := range schemaFileNames {
		result = append(result, k)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// GetSchemaPath returns the path to the schema directory.
func (v *Validator) GetSchemaPath() string {
	return filepath.Join(v.workspaceRoot, "contracts", "core", ContractVersion)
}
