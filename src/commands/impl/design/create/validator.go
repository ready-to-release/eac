package create

import (
	"fmt"
	"os"
	"path/filepath"

	design "github.com/ready-to-release/eac/src/commands/impl/design/internal"
	"github.com/ready-to-release/eac/src/core/contracts"
)

// StructurizrDSLValidator adapts design.StructurizrValidator to contracts.Validator
type StructurizrDSLValidator struct {
	validator    design.StructurizrValidator
	module       string
	tempDir      string
	templateRoot string
}

// NewStructurizrDSLValidator creates a new validator for Structurizr DSL
func NewStructurizrDSLValidator(module, templateRoot string) (*StructurizrDSLValidator, error) {
	validator, err := design.NewValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to create Structurizr validator: %w", err)
	}

	tempDir, err := os.MkdirTemp("", "design-validation-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return &StructurizrDSLValidator{
		validator:    validator,
		module:       module,
		tempDir:      tempDir,
		templateRoot: templateRoot,
	}, nil
}

// Validate validates the Structurizr DSL output
func (v *StructurizrDSLValidator) Validate(output string, context map[string]interface{}) []contracts.ValidationError {
	var errors []contracts.ValidationError

	// Create module design directory in temp
	moduleDesignDir := filepath.Join(v.tempDir, "specs", v.module, ".design")
	if err := os.MkdirAll(moduleDesignDir, 0755); err != nil {
		errors = append(errors, contracts.ValidationError{
			Code:     "setup_error",
			Message:  fmt.Sprintf("Failed to create temp design directory: %v", err),
			Severity: "error",
		})
		return errors
	}

	// Write DSL to temp workspace file
	tempWorkspace := filepath.Join(moduleDesignDir, "workspace.dsl")
	if err := os.WriteFile(tempWorkspace, []byte(output), 0644); err != nil {
		errors = append(errors, contracts.ValidationError{
			Code:     "setup_error",
			Message:  fmt.Sprintf("Failed to write temp workspace: %v", err),
			Severity: "error",
		})
		return errors
	}

	// Run Structurizr validation using absolute path
	// The validator will use the workspace path relative to the temp directory
	result, err := v.validateWithTempWorkspace(tempWorkspace)
	if err != nil {
		errors = append(errors, contracts.ValidationError{
			Code:     "validation_error",
			Message:  fmt.Sprintf("Validation execution failed: %v", err),
			Severity: "error",
		})
		return errors
	}

	// Convert to contracts.ValidationError format
	for _, msg := range result.Errors {
		errors = append(errors, contracts.ValidationError{
			Code:     "dsl_validation",
			Message:  msg.Message,
			Severity: "error",
			Line:     msg.Line,
		})
	}

	return errors
}

// VerifyImplementation verifies that the validator is properly configured
func (v *StructurizrDSLValidator) VerifyImplementation() []contracts.ValidationError {
	// No special requirements for this validator
	return []contracts.ValidationError{}
}

// IsDockerRunning checks if Docker daemon is available
func (v *StructurizrDSLValidator) IsDockerRunning() bool {
	return v.validator.IsDockerRunning()
}

// validateWithTempWorkspace validates a workspace file at an absolute path
func (v *StructurizrDSLValidator) validateWithTempWorkspace(workspacePath string) (*design.ValidationResult, error) {
	// Get the internal validator implementation to call executeDockerValidation directly
	implValidator, ok := v.validator.(*design.StructurizrValidatorImpl)
	if !ok {
		return nil, fmt.Errorf("validator is not a StructurizrValidatorImpl")
	}

	// Call the Docker validation directly with absolute path
	return implValidator.ValidateWorkspacePath(workspacePath)
}

// Cleanup removes temporary files
func (v *StructurizrDSLValidator) Cleanup() {
	os.RemoveAll(v.tempDir)
}
