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
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	"github.com/ready-to-release/eac/go/eac/core/validation"
	"github.com/ready-to-release/eac/go/eac/core/validation/formats/oscal"
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

	// Validate the catalog using core validator
	errors := validateCatalog(config)

	// Report results
	reportValidationResults(config, errors)

	if len(errors) > 0 {
		return 1
	}
	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	args := os.Args[3:] // Skip program name, "validate", and "risk-catalog"

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

// validateCatalog validates an OSCAL catalog using the core validator.
func validateCatalog(config *Config) []validation.ValidationError {
	// Read file
	data, err := os.ReadFile(config.FilePath)
	if err != nil {
		return []validation.ValidationError{
			*validation.NewValidationError(
				validation.ErrOSCALFileRead,
				fmt.Sprintf("failed to read file: %v", err),
				0,
			),
		}
	}

	// Create OSCAL catalog validator
	validator, err := oscal.NewValidator(oscal.TypeCatalog)
	if err != nil {
		return []validation.ValidationError{
			*validation.NewValidationError(
				validation.ErrSetupError,
				fmt.Sprintf("failed to create OSCAL validator: %v", err),
				0,
			),
		}
	}

	// Validate the catalog content
	context := map[string]interface{}{
		"file_path": config.FilePath,
	}
	return validator.Validate(string(data), context)
}

// reportValidationResults prints validation results.
func reportValidationResults(config *Config, errors []validation.ValidationError) {
	filename := filepath.Base(config.FilePath)

	if len(errors) == 0 {
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

		// Separate errors and warnings
		var errorList []validation.ValidationError
		var warningList []validation.ValidationError

		for _, e := range errors {
			if e.Code.Severity == validation.SeverityError {
				errorList = append(errorList, e)
			} else {
				warningList = append(warningList, e)
			}
		}

		if len(errorList) > 0 {
			log.Errorf("Errors: %d", len(errorList))
			for _, e := range errorList {
				log.Errorf("  - %s: %s", e.Code.Code, e.Message)
			}
		}

		if len(warningList) > 0 {
			log.Info("")
			log.Infof("Warnings: %d", len(warningList))
			for _, w := range warningList {
				log.Infof("  - %s: %s", w.Code.Code, w.Message)
			}
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
