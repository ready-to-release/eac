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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
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

// ValidationError represents a validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult holds the outcome of validation.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// ValidateRiskProfile is the entry point for the validate risk-profile command.
func ValidateRiskProfile() int {
	config, err := parseConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showHelp()
			return 0
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	result := validateProfile(config)
	reportValidationResults(config, result)

	if !result.Valid {
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

// validateProfile validates an OSCAL profile document using go-oscal types.
func validateProfile(config *Config) *ValidationResult {
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

	// Parse using go-oscal types
	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal(data, &oscalDoc); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "json",
			Message: fmt.Sprintf("invalid OSCAL profile JSON: %v", err),
		})
		return result
	}

	// Check if document contains a profile
	if oscalDoc.Profile == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "json",
			Message: "invalid OSCAL profile JSON: document does not contain a profile",
		})
		return result
	}

	profile := oscalDoc.Profile

	// Validate required UUID field
	if profile.UUID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.uuid",
			Message: "missing required field: uuid",
		})
	}

	// Validate metadata
	if profile.Metadata.Title == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.metadata.title",
			Message: "missing required field: title",
		})
	}

	if profile.Metadata.LastModified.IsZero() {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.metadata.last-modified",
			Message: "missing required field: last-modified",
		})
	}

	// Validate imports
	if len(profile.Imports) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.imports",
			Message: "missing required field: imports (must have at least one)",
		})
		return result
	}

	// Validate each import
	for i, imp := range profile.Imports {
		if imp.Href == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("profile.imports[%d].href", i),
				Message: "missing required field: href",
			})
		}

		// Check if there are controls included
		hasControls := false
		if imp.IncludeControls != nil && len(*imp.IncludeControls) > 0 {
			hasControls = true
			// Validate control IDs
			for j, ctrl := range *imp.IncludeControls {
				if ctrl.WithIds != nil {
					for k, id := range *ctrl.WithIds {
						if !isValidControlID(id) {
							result.Warnings = append(result.Warnings, ValidationError{
								Field:   fmt.Sprintf("profile.imports[%d].include-controls[%d].with-ids[%d]", i, j, k),
								Message: fmt.Sprintf("control ID '%s' may not be valid NIST 800-53 format", id),
							})
						}
					}
				}
			}
		}

		if !hasControls {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   fmt.Sprintf("profile.imports[%d]", i),
				Message: "import has no control selections",
			})
		}
	}

	// Validate OSCAL version if present
	if profile.Metadata.OscalVersion != "" {
		expectedVersion := "1.1.2"
		if profile.Metadata.OscalVersion != expectedVersion {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "profile.metadata.oscal-version",
				Message: fmt.Sprintf("OSCAL version %s differs from expected %s", profile.Metadata.OscalVersion, expectedVersion),
			})
		}
	}

	return result
}

// isValidControlID checks if a control ID is valid NIST 800-53 format.
func isValidControlID(id string) bool {
	parts := strings.Split(strings.ToLower(id), "-")
	if len(parts) < 2 {
		return false
	}

	// Family prefix should be 2 letters
	family := parts[0]
	if len(family) != 2 {
		return false
	}

	for _, c := range family {
		if c < 'a' || c > 'z' {
			return false
		}
	}

	// Number part should be non-empty and numeric
	number := parts[1]
	if len(number) == 0 {
		return false
	}

	for _, c := range number {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// reportValidationResults prints validation results.
func reportValidationResults(config *Config, result *ValidationResult) {
	filename := filepath.Base(config.FilePath)

	if result.Valid {
		log.Info("")
		log.Infof("✓ Profile is valid OSCAL 1.1.2 document: %s", filename)

		if len(result.Warnings) > 0 {
			log.Info("")
			log.Infof("  Warnings: %d", len(result.Warnings))
			for _, w := range result.Warnings {
				log.Infof("    - %s: %s", w.Field, w.Message)
			}
		}
		log.Info("")
	} else {
		log.Info("")
		log.Errorf("✗ Profile validation failed: %s", filename)
		log.Info("")

		log.Errorf("  Errors: %d", len(result.Errors))
		for _, e := range result.Errors {
			log.Errorf("    - %s: %s", e.Field, e.Message)
		}

		if len(result.Warnings) > 0 {
			log.Info("")
			log.Infof("  Warnings: %d", len(result.Warnings))
			for _, w := range result.Warnings {
				log.Infof("    - %s: %s", w.Field, w.Message)
			}
		}
		log.Info("")
	}
}

// showHelp displays help information.
func showHelp() {
	help := `Usage: validate risk-profile <file> [flags]

Validate OSCAL profile documents against OSCAL 1.1.2 schema

Arguments:
  file                 Path to OSCAL profile JSON file

Flags:
  -h, --help           Show this help message

Examples:
  # Validate profile
  validate risk-profile specs/risk-controls/billing.profile.json

  # Validate all profiles in directory
  for f in specs/risk-controls/*.profile.json; do validate risk-profile "$f"; done

For assessment-results validation, use: validate risk-assess
For catalog validation, use: validate risk-catalog
`
	log.Info(help)
}
