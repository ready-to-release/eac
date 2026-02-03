// Package internal provides build manifest validation against the contract schema
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// ManifestValidator validates build manifests against the contract schema.
type ManifestValidator struct {
	schema *jsonschema.Schema
}

// ManifestValidationError represents a manifest validation error.
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

// Contract version for build manifest schema.
const manifestContractVersion = "0.1.0"

// NewManifestValidator creates a new manifest validator that loads schema from domain.
func NewManifestValidator(workspaceRoot string) (*ManifestValidator, error) {
	c := jsonschema.NewCompiler()

	// Use distribution root (schemas are part of tool distribution, not user workspace)
	// In containers, R2R_CONTAINER_ROOT points to the tool distribution
	schemaRoot := workspaceRoot
	if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
		schemaRoot = containerRoot
	}

	// Build path to schema: contracts/core/<version>/build-manifest.schema.json
	schemaPath := filepath.Join(
		paths.ContractsVersionPath(schemaRoot, "core", manifestContractVersion),
		"build-manifest.schema.json",
	)

	// Read schema from contracts
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read build manifest schema from %s: %w", schemaPath, err)
	}

	// Parse the schema
	var schemaDoc any
	if err := json.Unmarshal(data, &schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to parse build manifest schema: %w", err)
	}

	// Add schema to compiler
	schemaURL := "file:///build-manifest.schema.json"
	if err := c.AddResource(schemaURL, schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile the schema
	schema, err := c.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile build manifest schema: %w", err)
	}

	return &ManifestValidator{schema: schema}, nil
}

// ValidateManifest validates a ModuleManifest against the contract schema.
func (v *ManifestValidator) ValidateManifest(manifest *ModuleManifest) error {
	// Convert manifest to JSON for validation
	jsonData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return v.ValidateJSON(jsonData, manifest.Moniker)
}

// ValidateJSON validates raw JSON data against the manifest schema.
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

// extractManifestValidationDetails extracts detailed error messages from validation errors.
func extractManifestValidationDetails(err error) []string {
	var details []string

	validErr := &jsonschema.ValidationError{}
	if errors.As(err, &validErr) {
		details = append(details, validErr.Error())
		for _, cause := range validErr.Causes {
			details = append(details, extractManifestValidationDetails(cause)...)
		}
	}

	return details
}

// Global validator instance (lazy initialized).
var (
	globalManifestValidator     *ManifestValidator
	globalManifestValidatorRoot string
)

// GetManifestValidator returns the global manifest validator instance.
// Uses repository.GetRepositoryRoot() to find workspace root.
func GetManifestValidator() (*ManifestValidator, error) {
	return GetManifestValidatorWithRoot("")
}

// GetManifestValidatorWithRoot returns the global manifest validator instance with explicit workspace root.
// If workspaceRoot is empty, it will be detected automatically.
func GetManifestValidatorWithRoot(workspaceRoot string) (*ManifestValidator, error) {
	// If no root provided, try to detect it
	if workspaceRoot == "" {
		// Use distribution root for schema loading
		if containerRoot := os.Getenv("R2R_CONTAINER_ROOT"); containerRoot != "" {
			workspaceRoot = containerRoot
		} else {
			// Try to find workspace root by looking for .r2r directory
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			workspaceRoot = findWorkspaceRoot(cwd)
			if workspaceRoot == "" {
				return nil, fmt.Errorf("could not find workspace root (no .r2r directory found)")
			}
		}
	}

	// Check if we need to reinitialize (different root)
	if globalManifestValidator != nil && globalManifestValidatorRoot == workspaceRoot {
		return globalManifestValidator, nil
	}

	var err error
	globalManifestValidator, err = NewManifestValidator(workspaceRoot)
	if err != nil {
		return nil, err
	}
	globalManifestValidatorRoot = workspaceRoot

	return globalManifestValidator, nil
}

// findWorkspaceRoot walks up the directory tree looking for .r2r directory.
func findWorkspaceRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(paths.R2RPath(dir)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // Reached root
		}
		dir = parent
	}
}
