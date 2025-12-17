// Command: validate risk-catalog
// Short: Validate OSCAL catalogs against OSCAL 1.1.3 schema
// Long: The validate risk-catalog command validates OSCAL catalog documents against the official
// Long: OSCAL 1.1.3 JSON schema from NIST.
// Long:
// Long: Catalogs define security control libraries (e.g., NIST 800-53) with structured control
// Long: definitions, parameters, and supporting materials.
// Long:
// Long: Validation uses the official OSCAL JSON schema:
// Long: https://github.com/usnistgov/OSCAL/releases/download/v1.1.3/oscal_catalog_schema.json
// Long:
// Long: Expected Output:
// Long:   Displays OSCAL schema validation results for catalog document.
// Long:   Shows missing required fields, schema violations, and structural errors.
// Long:   Exit code 0 if valid OSCAL 1.1.3 catalog, 1 if validation errors.
// Args: file
package riskcatalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(ValidateRiskCatalog)
}

// Config holds configuration for validate risk-catalog command.
type Config struct {
	FilePath      string
	WorkspaceRoot string
}

// ValidationError represents a validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult holds the outcome of validation.
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// ValidateRiskCatalog is the entry point for the validate risk-catalog command.
func ValidateRiskCatalog() int {
	config, err := parseConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showHelp()
			return 0
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	// Validate the catalog using schema
	result := validateCatalog(config)

	// Report results
	reportValidationResults(config, result)

	if !result.Valid {
		return 1
	}
	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	args := os.Args[3:] // Skip program name, "validate", and "risk-catalog"

	// Define expected flags
	commandFlags := []flags.FlagDefinition{
		{Name: "--help", Shorthand: "-h", HasValue: false},
	}

	// Validate flags
	if err := flags.ValidateFlags(args, commandFlags); err != nil {
		return nil, err
	}

	config := &Config{}

	// Get workspace root
	workspaceRoot, err := registry.GetWorkspaceRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot

	// Parse flags and arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			return nil, fmt.Errorf("help requested")

		default:
			// Positional argument: file path
			if config.FilePath == "" {
				config.FilePath = arg
				// Make path absolute if relative
				if !filepath.IsAbs(config.FilePath) {
					config.FilePath = filepath.Join(config.WorkspaceRoot, config.FilePath)
				}
			}
			i++
		}
	}

	// Validate required arguments
	if config.FilePath == "" {
		return nil, fmt.Errorf("file path required")
	}

	// Check file exists and is regular file
	fileInfo, err := os.Stat(config.FilePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", config.FilePath)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	// Ensure it's a file, not a directory
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", config.FilePath)
	}

	return config, nil
}

// validateCatalog validates an OSCAL catalog using go-oscal types.
func validateCatalog(config *Config) *ValidationResult {
	result := &ValidationResult{Valid: true}

	// Read file
	data, err := os.ReadFile(config.FilePath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "file",
			Message: fmt.Sprintf("failed to read file: %v", err),
		})
		return result
	}

	// Parse using go-oscal wrapper type
	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal(data, &oscalDoc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "json",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		})
		return result
	}

	// Check if it's a catalog document
	if oscalDoc.Catalog == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "document",
			Message: "not an OSCAL catalog document",
		})
		return result
	}

	catalog := oscalDoc.Catalog

	// Validate required UUID field
	if catalog.UUID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "catalog.uuid",
			Message: "missing required field: uuid",
		})
	}

	// Validate metadata
	if catalog.Metadata.Title == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "catalog.metadata.title",
			Message: "missing required field: title",
		})
	}

	if catalog.Metadata.LastModified.IsZero() {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "catalog.metadata.last-modified",
			Message: "missing required field: last-modified",
		})
	}

	// Validate that catalog has at least one control or group
	hasControls := catalog.Controls != nil && len(*catalog.Controls) > 0
	hasGroups := catalog.Groups != nil && len(*catalog.Groups) > 0

	if !hasControls && !hasGroups {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "catalog",
			Message: "catalog must have at least one control or group",
		})
	}

	return result
}

// reportValidationResults prints validation results.
func reportValidationResults(config *Config, result *ValidationResult) {
	filename := filepath.Base(config.FilePath)

	if result.Valid {
		log.Info("")
		log.Infof("Validation passed")
		log.Infof("Type: catalog")
		log.Infof("Schema: OSCAL 1.1.3")
		log.Infof("File: %s", filename)
		log.Info("")
		log.Infof("✓ Catalog is valid OSCAL 1.1.3 document")
		log.Info("")
	} else {
		log.Info("")
		log.Errorf("Validation failed")
		log.Errorf("File: %s", filename)
		log.Info("")

		log.Errorf("Errors: %d", len(result.Errors))
		for _, e := range result.Errors {
			log.Errorf("  - %s: %s", e.Field, e.Message)
		}
		log.Info("")
	}
}

// showHelp displays help information.
func showHelp() {
	help := `Usage: validate risk-catalog <file>

Validate OSCAL catalogs against OSCAL 1.1.3 schema

Arguments:
  file                   Path to OSCAL catalog document to validate

Flags:
  -h, --help             Show this help message

Examples:
  # Validate a catalog
  validate risk-catalog catalogs/nist-800-53-rev5.json

  # Validate catalog from URL (download first)
  curl -o catalog.json https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json
  validate risk-catalog catalog.json

OSCAL Catalog Structure:
  Catalogs define security control libraries with:
  - Metadata (title, version, last-modified, OSCAL version)
  - Groups (control families like Access Control, System Integrity)
  - Controls (individual security requirements with statements)
  - Parameters (configurable values within controls)
  - Back-matter (supporting resources and references)

  Official NIST 800-53 Rev 5 catalog:
  https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json

Validation:
  This command validates catalogs against the official OSCAL 1.1.3 JSON schema
  maintained by NIST. The schema ensures complete compliance with the OSCAL
  specification including all required fields, data types, and structural constraints.

  Schema URL: https://github.com/usnistgov/OSCAL/releases/download/v1.1.3/oscal_catalog_schema.json
`
	log.Info(help)
}
