package riskcatalog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/validation"
	"github.com/ready-to-release/eac/go/core/validation/formats/oscal"
)

var log = logging.C()

type validateRiskCatalogCommand struct{}

var _ core.SimpleCommandPort = (*validateRiskCatalogCommand)(nil)

func (c *validateRiskCatalogCommand) Name() string { return "validate risk-catalog" }

func (c *validateRiskCatalogCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-risk-catalog",
		Short:         "Validate OSCAL catalogs against OSCAL 1.1.3 schema",
		Long:          "The validate risk-catalog command validates OSCAL catalog documents against the official\nOSCAL 1.1.3 JSON schema from NIST.\n\nCatalogs define security control libraries (e.g., NIST 800-53) with structured control\ndefinitions, parameters, and supporting materials.\n\nValidation uses the official OSCAL JSON schema:\nhttps://github.com/usnistgov/OSCAL/releases/download/v1.1.3/oscal_catalog_schema.json\n\nExpected Output:\n  Displays OSCAL schema validation results for catalog document.\n  Shows missing required fields, schema violations, and structural errors.\n  Exit code 0 if valid OSCAL 1.1.3 catalog, 1 if validation errors.",
		Args:          "file",
	}
}

func (c *validateRiskCatalogCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateRiskCatalog()
}

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&validateRiskCatalogCommand{},
	}
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
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}
	config.WorkspaceRoot = workspaceRoot

	// Parse flags and arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
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
