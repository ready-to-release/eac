// Package schema provides JSON Schema validation for repository configuration files
package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/paths"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// SchemaType represents the type of schema to validate against.
type SchemaType string

const (
	SchemaComponentTypes     SchemaType = "component-types"
	SchemaEnvironments       SchemaType = "environments"
	SchemaTestingTags        SchemaType = "testing-tags"
	SchemaTestSuites         SchemaType = "test-suites"
	SchemaSystemDependencies SchemaType = "system-dependencies"
	SchemaRepository         SchemaType = "repository"
	SchemaEACConfig          SchemaType = "ai-provider"
	SchemaBooks              SchemaType = "books"
	SchemaSecurityTools      SchemaType = "security-tools"
	SchemaCommands           SchemaType = "commands"
)

// schemaFileNames maps schema types to their file names (without path).
var schemaFileNames = map[SchemaType]string{
	SchemaComponentTypes:     "component-types.schema.json",
	SchemaEnvironments:       "environments.schema.json",
	SchemaTestingTags:        "testing-tags.schema.json",
	SchemaTestSuites:         "test-suites.schema.json",
	SchemaSystemDependencies: "system-dependencies.schema.json",
	SchemaRepository:         "repository.schema.json",
	SchemaEACConfig:          "ai-provider.schema.json",
	SchemaBooks:              "books.schema.json",
	SchemaSecurityTools:      "security-tools.schema.json",
	SchemaCommands:           "commands.schema.json",
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

// NewValidator creates a new schema validator with all schemas loaded from the repository
// workspaceRoot should be the repository root directory.
func NewValidator(workspaceRoot string) (*Validator, error) {
	c := jsonschema.NewCompiler()

	v := &Validator{
		compiler:      c,
		schemas:       make(map[SchemaType]*jsonschema.Schema),
		workspaceRoot: workspaceRoot,
	}

	// Build the schema directory path: contracts/eac-core/<version>/
	// Use distribution root (schemas are part of tool distribution, not user workspace)
	// Note: Can't import repository package here to avoid cycles, so inline the check
	// See repository.GetDistRoot() for the canonical implementation
	schemaRoot := workspaceRoot
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
		schemaRoot = containerRoot
	}
	schemaDir := paths.ContractsVersionPath(schemaRoot, "eac-core", ContractVersion)

	// Load and compile all schemas from the repository
	for schemaType, fileName := range schemaFileNames {
		filePath := filepath.Join(schemaDir, fileName)

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema %s: %w", filePath, err)
		}

		// Parse the schema JSON
		var schemaDoc any
		if err := json.Unmarshal(data, &schemaDoc); err != nil {
			return nil, fmt.Errorf("failed to parse schema %s: %w", filePath, err)
		}

		// Add schema to compiler
		schemaURL := fmt.Sprintf("file:///%s", fileName)
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
func GetSchemaTypes() []SchemaType {
	return []SchemaType{
		SchemaComponentTypes,
		SchemaEnvironments,
		SchemaTestingTags,
		SchemaTestSuites,
		SchemaSystemDependencies,
		SchemaRepository,
		SchemaEACConfig,
		SchemaBooks,
		SchemaSecurityTools,
	}
}

// GetSchemaPath returns the path to the schema directory.
func (v *Validator) GetSchemaPath() string {
	return filepath.Join(v.workspaceRoot, "contracts", "eac-core", ContractVersion)
}
