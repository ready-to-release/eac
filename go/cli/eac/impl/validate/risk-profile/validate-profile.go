// Command: validate risk-profile
// Short: Validate OSCAL profile documents against OSCAL 1.1.2 schema
// Long: The validate risk-profile command validates OSCAL profile documents using go-oscal types.
// Long:
// Long: Validation checks include:
// Long: - JSON parsing with go-oscal types
// Long: - Required field presence (UUID, metadata, imports)
// Long: - Control ID format validation
// Long: - Import href validation
// Long:
// Long: Expected Output:
// Long:   Displays OSCAL profile validation results including required field checks,
// Long:   control ID format validation, and import href validation. Shows errors and warnings.
// Long:   Exit code 0 if valid OSCAL 1.1.2 profile, 1 if validation errors.
// Args: file
package riskprofile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/validation"
	"github.com/ready-to-release/eac/go/core/validation/formats/oscal"
)

var log = logging.C()

func init() {
	registry.Register(ValidateRiskProfile)
}

// Config holds configuration for validate risk-profile command.
type Config struct {
	FilePath      string
	WorkspaceRoot string
}

// ValidateRiskProfile is the entry point for the validate risk-profile command.
func ValidateRiskProfile() int {
	config, err := parseConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	errors := validateProfile(config)
	reportValidationResults(config, errors)

	// Check if there are any actual errors (not just warnings)
	hasErrors := false
	for _, e := range errors {
		if e.Code.Severity == validation.SeverityError {
			hasErrors = true
			break
		}
	}

	if hasErrors {
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

	args := os.Args[3:] // Skip program name, "validate", and "risk-profile"

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

// validateProfile validates an OSCAL profile document using the core validator.
func validateProfile(config *Config) []validation.ValidationError {
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

	// Create OSCAL profile validator
	validator, err := oscal.NewValidator(oscal.TypeProfile)
	if err != nil {
		return []validation.ValidationError{
			*validation.NewValidationError(
				validation.ErrSetupError,
				fmt.Sprintf("failed to create OSCAL validator: %v", err),
				0,
			),
		}
	}

	// Validate the profile content
	context := map[string]interface{}{
		"file_path": config.FilePath,
	}
	return validator.Validate(string(data), context)
}

// reportValidationResults prints validation results.
func reportValidationResults(config *Config, errors []validation.ValidationError) {
	filename := filepath.Base(config.FilePath)

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

	if len(errorList) == 0 {
		log.Info("")
		log.Infof("✓ Profile is valid OSCAL 1.1.2 document: %s", filename)

		if len(warningList) > 0 {
			log.Info("")
			log.Infof("  Warnings: %d", len(warningList))
			for _, w := range warningList {
				log.Infof("    - %s: %s", w.Code.Code, w.Message)
			}
		}
		log.Info("")
	} else {
		log.Info("")
		log.Errorf("✗ Profile validation failed: %s", filename)
		log.Info("")

		log.Errorf("  Errors: %d", len(errorList))
		for _, e := range errorList {
			log.Errorf("    - %s: %s", e.Code.Code, e.Message)
		}

		if len(warningList) > 0 {
			log.Info("")
			log.Infof("  Warnings: %d", len(warningList))
			for _, w := range warningList {
				log.Infof("    - %s: %s", w.Code.Code, w.Message)
			}
		}
		log.Info("")
	}
}
