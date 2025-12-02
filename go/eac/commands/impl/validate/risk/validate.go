// Command: validate risk
// Short: Validate OSCAL profiles and assessment-results against schemas
// Long: The validate risk command validates OSCAL documents against the OSCAL 1.1.2 JSON schema.
// Long: It can validate both profile and assessment-results documents, auto-detecting the type
// Long: if not specified.
// Long:
// Long: Validation checks include:
// Long: - JSON schema compliance
// Long: - Required field presence
// Long: - UUID uniqueness
// Long: - Control ID format
// Long: - Evidence link existence (optional)
// Flag.type: type=string, completion=profile,assessment, usage=Document type: profile or assessment (auto-detected if not specified)
// Flag.check-evidence: type=bool, default=false, usage=Verify evidence file links exist
// Args: files
package risk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(ValidateRisk)
}

// Config holds configuration for validate risk command.
type Config struct {
	FilePath      string
	DocumentType  string // "profile" or "assessment"
	CheckEvidence bool
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

// ValidateRisk is the entry point for the validate risk command.
func ValidateRisk() int {
	config, err := parseConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showHelp()
			return 0
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	// Auto-detect document type if not specified
	if config.DocumentType == "" {
		detectedType, err := oscal.DetectOSCALDocumentType(config.FilePath)
		if err != nil {
			log.Errorf("Error detecting document type: %v", err)
			return 1
		}
		if detectedType == "" {
			log.Error("Error: Could not auto-detect document type")
			log.Error("Use --type to specify: profile or assessment")
			return 1
		}
		config.DocumentType = detectedType
		log.Infof("Auto-detected document type: %s", config.DocumentType)
	}

	// Validate based on type
	var result *ValidationResult
	switch config.DocumentType {
	case "profile":
		result = validateProfile(config)
	case "assessment", "assessment-results":
		result = validateAssessmentResults(config)
	default:
		log.Errorf("Error: Unknown document type: %s", config.DocumentType)
		log.Error("Valid types: profile, assessment")
		return 1
	}

	// Report results
	reportValidationResults(config, result)

	if !result.Valid {
		return 1
	}
	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	args := os.Args[3:] // Skip program name, "validate", and "risk"

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

		case arg == "--type":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--type requires a value")
			}
			config.DocumentType = args[i+1]
			i += 2

		case arg == "--check-evidence":
			config.CheckEvidence = true
			i++

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

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

	// Check file exists
	if _, err := os.Stat(config.FilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", config.FilePath)
	}

	return config, nil
}

// validateProfile validates an OSCAL profile document.
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

	// Parse JSON
	var profile oscal.Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "json",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		})
		return result
	}

	// Validate required fields
	if profile.Profile.UUID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.uuid",
			Message: "missing required field: uuid",
		})
	}

	if profile.Profile.Metadata.Title == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.metadata.title",
			Message: "missing required field: title",
		})
	}

	if profile.Profile.Metadata.LastModified == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.metadata.last-modified",
			Message: "missing required field: last-modified",
		})
	}

	if len(profile.Profile.Imports) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "profile.imports",
			Message: "missing required field: imports (must have at least one)",
		})
	}

	// Validate imports
	for i, imp := range profile.Profile.Imports {
		if imp.Href == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("profile.imports[%d].href", i),
				Message: "missing required field: href",
			})
		}

		// Validate control ID format
		for j, ctrl := range imp.IncludeControls {
			if !isValidControlID(ctrl.WithID) {
				result.Warnings = append(result.Warnings, ValidationError{
					Field:   fmt.Sprintf("profile.imports[%d].include-controls[%d].with-id", i, j),
					Message: fmt.Sprintf("control ID '%s' may not be valid NIST 800-53 format", ctrl.WithID),
				})
			}
		}
	}

	// Validate OSCAL version
	if profile.Profile.Metadata.OSCALVersion != "" && profile.Profile.Metadata.OSCALVersion != oscal.OSCALVersion {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "profile.metadata.oscal-version",
			Message: fmt.Sprintf("OSCAL version %s differs from expected %s", profile.Profile.Metadata.OSCALVersion, oscal.OSCALVersion),
		})
	}

	return result
}

// validateAssessmentResults validates an OSCAL assessment-results document.
func validateAssessmentResults(config *Config) *ValidationResult {
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

	// Parse JSON
	var ar oscal.AssessmentResults
	if err := json.Unmarshal(data, &ar); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "json",
			Message: fmt.Sprintf("invalid JSON: %v", err),
		})
		return result
	}

	// Validate required fields
	if ar.AssessmentResults.UUID == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "assessment-results.uuid",
			Message: "missing required field: uuid",
		})
	}

	if ar.AssessmentResults.Metadata.Title == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "assessment-results.metadata.title",
			Message: "missing required field: title",
		})
	}

	if ar.AssessmentResults.Metadata.LastModified == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "assessment-results.metadata.last-modified",
			Message: "missing required field: last-modified",
		})
	}

	if len(ar.AssessmentResults.Results) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "assessment-results.results",
			Message: "missing required field: results (must have at least one)",
		})
	}

	// Validate UUID uniqueness
	uuidMap := make(map[string]string)
	checkUUID := func(uuid, location string) {
		if uuid == "" {
			return
		}
		if existing, ok := uuidMap[uuid]; ok {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   location,
				Message: fmt.Sprintf("duplicate UUID '%s' (also found at %s)", uuid, existing),
			})
		}
		uuidMap[uuid] = location
	}

	checkUUID(ar.AssessmentResults.UUID, "assessment-results.uuid")

	// Validate each result
	for i, res := range ar.AssessmentResults.Results {
		checkUUID(res.UUID, fmt.Sprintf("assessment-results.results[%d].uuid", i))

		if len(res.Findings) == 0 {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   fmt.Sprintf("assessment-results.results[%d].findings", i),
				Message: "result has no findings",
			})
		}

		// Build observation UUID map for reference validation
		obsMap := make(map[string]bool)
		for j, obs := range res.Observations {
			checkUUID(obs.UUID, fmt.Sprintf("assessment-results.results[%d].observations[%d].uuid", i, j))
			obsMap[obs.UUID] = true

			// Check evidence links if requested
			if config.CheckEvidence {
				for k, evidence := range obs.RelevantEvidence {
					evidencePath := filepath.Join(config.WorkspaceRoot, evidence.Href)
					if _, err := os.Stat(evidencePath); os.IsNotExist(err) {
						result.Warnings = append(result.Warnings, ValidationError{
							Field:   fmt.Sprintf("assessment-results.results[%d].observations[%d].relevant-evidence[%d].href", i, j, k),
							Message: fmt.Sprintf("evidence file not found: %s", evidence.Href),
						})
					}
				}
			}
		}

		// Validate findings
		for j, finding := range res.Findings {
			checkUUID(finding.UUID, fmt.Sprintf("assessment-results.results[%d].findings[%d].uuid", i, j))

			// Validate target
			if finding.Target.TargetID == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("assessment-results.results[%d].findings[%d].target.target-id", i, j),
					Message: "missing required field: target-id",
				})
			}

			// Validate status
			if finding.Target.Status.State == "" {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("assessment-results.results[%d].findings[%d].target.status.state", i, j),
					Message: "missing required field: state",
				})
			} else if finding.Target.Status.State != oscal.StateSatisfied && finding.Target.Status.State != oscal.StateNotSatisfied {
				result.Warnings = append(result.Warnings, ValidationError{
					Field:   fmt.Sprintf("assessment-results.results[%d].findings[%d].target.status.state", i, j),
					Message: fmt.Sprintf("unexpected state value: %s (expected: satisfied or not-satisfied)", finding.Target.Status.State),
				})
			}

			// Validate related observation references
			for k, relObs := range finding.RelatedObservations {
				if !obsMap[relObs.ObservationUUID] {
					result.Valid = false
					result.Errors = append(result.Errors, ValidationError{
						Field:   fmt.Sprintf("assessment-results.results[%d].findings[%d].related-observations[%d].observation-uuid", i, j, k),
						Message: fmt.Sprintf("references non-existent observation: %s", relObs.ObservationUUID),
					})
				}
			}
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

	// Number part should be numeric
	for _, c := range parts[1] {
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
		log.Infof("✓ Validation passed: %s", filename)
		log.Infof("  Type: %s", config.DocumentType)

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
		log.Errorf("✗ Validation failed: %s", filename)
		log.Infof("  Type: %s", config.DocumentType)
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
	help := `Usage: validate risk <file> [flags]

Validate OSCAL profiles and assessment-results against schemas

Arguments:
  file                 Path to OSCAL document to validate

Flags:
      --type <type>        Document type: profile or assessment (auto-detected if not specified)
      --check-evidence     Verify evidence file links exist
  -h, --help               Show this help message

Examples:
  # Validate a profile (auto-detect type)
  validate risk specs/risk-controls/billing.profile.json

  # Validate assessment-results with explicit type
  validate risk out/risk/billing/assessment-results.json --type assessment

  # Validate and check evidence links exist
  validate risk out/risk/billing/assessment-results.json --check-evidence
`
	log.Info(help)
}
