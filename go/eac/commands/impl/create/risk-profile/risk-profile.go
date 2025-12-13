// Command: create risk-profile
// Short: Create OSCAL profile from risk assessment using AI
// Long: The create risk-profile command analyzes a risk assessment document and generates an OSCAL profile
// Long: selecting appropriate controls from a custom catalog. The AI extracts risks
// Long: from the assessment and maps them to controls for the entire solution.
// Long:
// Long: The generated profile is saved to specs/.risk-controls/risk-profile.json for version control.
// Long: Use --debug to inspect intermediate outputs and AI reasoning.
// Long:
// Long: Expected Output:
// Long: - OSCAL profile JSON file selecting controls from catalog
// Long: - AI reasoning for control selection in debug output
// Flag.catalog: type=string, usage=Catalog URL for control selection and validation (default: NIST 800-53 Rev5)
// Flag.output: type=string, shorthand=o, usage=Custom output path for the profile file
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite existing profile file
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs to out/logs/risk/
// Flag.max-retries: type=int, default=3, usage=Maximum AI generation retries on validation failure
// Args: file
package riskprofile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(CreateRiskProfile)
}

// MockAIResponse holds mock response for testing.
var mockAIResponse string

// SetMockAIResponse sets a mock AI response for testing.
func SetMockAIResponse(response string) {
	mockAIResponse = response
}

// ResetMockAIResponse clears the mock AI response.
func ResetMockAIResponse() {
	mockAIResponse = ""
}

// Config holds configuration for create risk-profile command.
type Config struct {
	AssessmentPath string
	CatalogURL     string
	OutputPath     string
	Force          bool
	Debug          bool
	MaxRetries     int
	WorkspaceRoot  string
	Logger         *logging.Logger
}

// CreateRiskProfile is the entry point for the create risk-profile command.
func CreateRiskProfile() int {
	config, err := parseConfig()
	if err != nil {
		if err.Error() == "help requested" {
			showHelp()
			return 0
		}
		log.Errorf("Error: %v", err)
		return 1
	}

	// Initialize logger
	var logger *logging.Logger
	if config.Debug {
		logger, err = logging.NewWithDebug("risk", config.WorkspaceRoot)
	} else {
		logger, err = logging.NewDefault("risk", config.WorkspaceRoot)
	}
	if err != nil {
		log.Errorf("Warning: Failed to initialize logger: %v", err)
	} else {
		config.Logger = logger
		defer logger.Sync()
	}

	if config.Logger != nil {
		config.Logger.Info("Starting create risk-profile",
			zap.String("assessment", config.AssessmentPath),
			zap.Bool("debug", config.Debug))
	}

	// Read assessment file
	log.Infof("Reading assessment file: %s", config.AssessmentPath)
	assessmentContent, err := os.ReadFile(config.AssessmentPath)
	if err != nil {
		log.Errorf("Error reading assessment file: %v", err)
		return 1
	}
	log.Infof("Assessment file loaded (%d bytes)", len(assessmentContent))
	log.Info("")

	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(config.WorkspaceRoot, "specs", ".risk-controls", "risk-profile.json")
	}

	// Check if file exists
	if !config.Force {
		if _, err := os.Stat(outputPath); err == nil {
			log.Errorf("Error: Profile already exists: %s", outputPath)
			log.Error("Use --force to overwrite")
			return 1
		}
	}

	// Load catalog once before retry loop (for control info and validation)
	log.Infof("Analyzing assessment and generating OSCAL profile...")
	log.Infof("Catalog: %s", config.CatalogURL)
	log.Infof("Loading catalog for control information...")

	catalog, err := oscal.LoadCatalog(config.CatalogURL)
	if err != nil {
		log.Errorf("Error loading catalog: %v", err)
		return 1
	}
	log.Infof("  ✓ Catalog loaded successfully")
	log.Info("")

	var profile *oscalTypes.Profile
	var lastErr error

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		if attempt > 1 {
			log.Infof("Retry %d/%d...", attempt, config.MaxRetries)
		}

		log.Infof("Step 1/%d: Calling AI to analyze risks and map controls...", 3)
		profile, err = generateProfile(config, string(assessmentContent), catalog)
		if err != nil {
			lastErr = err
			log.Warnf("  ✗ AI generation failed: %v", err)
			if config.Logger != nil {
				config.Logger.Warn("Profile generation failed",
					zap.Int("attempt", attempt),
					zap.Error(err))
			}
			continue
		}
		log.Infof("  ✓ AI returned %d controls", len(oscal.GetProfileControlIDs(profile)))

		log.Infof("Step 2/%d: Validating profile structure...", 3)
		if err := validateProfile(profile); err != nil {
			lastErr = err
			log.Warnf("  ✗ Validation failed: %v", err)
			if config.Logger != nil {
				config.Logger.Warn("Profile validation failed",
					zap.Int("attempt", attempt),
					zap.Error(err))
			}
			continue
		}
		log.Info("  ✓ Profile structure valid")

		log.Infof("Step 3/%d: Validating controls against catalog...", 3)
		if err := oscal.ValidateControlIDsAgainstCatalog(oscal.GetProfileControlIDs(profile), catalog); err != nil {
			lastErr = err
			log.Warnf("  ✗ Catalog validation failed: %v", err)
			if config.Logger != nil {
				config.Logger.Warn("Catalog validation failed",
					zap.Int("attempt", attempt),
					zap.Error(err))
			}
			continue
		}
		log.Info("  ✓ All controls validated against catalog")
		log.Info("")

		// Success - clear lastErr to indicate success
		lastErr = nil
		break
	}

	if profile == nil || lastErr != nil {
		log.Errorf("Failed to generate valid profile after %d attempts: %v", config.MaxRetries, lastErr)
		return 1
	}

	// Write profile
	log.Infof("Writing profile to: %s", outputPath)
	if err := oscal.WriteProfile(outputPath, profile); err != nil {
		log.Errorf("Error writing profile: %v", err)
		return 1
	}

	// Report success
	controlIDs := oscal.GetProfileControlIDs(profile)
	log.Info("")
	log.Infof("Created OSCAL profile: %s", outputPath)
	log.Infof("  Controls: %d (%s)", len(controlIDs), strings.Join(controlIDs, ", "))
	log.Infof("  Catalog: %s", oscal.GetProfileCatalogURL(profile))
	log.Info("")
	log.Info("Next steps:")
	log.Infof("  1. Add @control(%s) tags to your .feature files", controlIDs[0])
	log.Infof("  2. Run: create risk-assess --profile %s", outputPath)

	if config.Logger != nil {
		config.Logger.Info("Profile created successfully",
			zap.String("path", outputPath),
			zap.Int("controls", len(controlIDs)))
	}

	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	args := os.Args[3:] // Skip program name, "create", and "risk-profile"

	config := &Config{
		CatalogURL: oscal.NIST80053Rev5CatalogURL,
		MaxRetries: 3,
	}

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

		case arg == "--catalog":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--catalog requires a value")
			}
			config.CatalogURL = args[i+1]
			i += 2

		case arg == "--output" || arg == "-o":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output requires a value")
			}
			config.OutputPath = args[i+1]
			i += 2

		case arg == "--force":
			config.Force = true
			i++

		case arg == "--debug" || arg == "-d":
			config.Debug = true
			i++

		case arg == "--max-retries":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--max-retries requires a value")
			}
			fmt.Sscanf(args[i+1], "%d", &config.MaxRetries)
			i += 2

		case strings.HasPrefix(arg, "-"):
			return nil, fmt.Errorf("unknown flag: %s", arg)

		default:
			// Positional argument: assessment file
			if config.AssessmentPath == "" {
				config.AssessmentPath = arg
				// Make path absolute if relative
				if !filepath.IsAbs(config.AssessmentPath) {
					config.AssessmentPath = filepath.Join(config.WorkspaceRoot, config.AssessmentPath)
				}
			}
			i++
		}
	}

	// Validate required arguments
	if config.AssessmentPath == "" {
		return nil, fmt.Errorf("assessment file path required")
	}

	// Check file exists and is regular file
	fileInfo, err := os.Stat(config.AssessmentPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf(`assessment file not found: %s

Possible causes:
  - File path is incorrect
  - File is in a different directory
  - File hasn't been created yet

Try:
  - Check current directory: pwd
  - List assessment files: ls *.md
  - Use absolute path: /full/path/to/assessment.md`, config.AssessmentPath)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot access assessment file: %w", err)
	}

	// Ensure it's a file, not a directory
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", config.AssessmentPath)
	}

	// Validate file size (max 10MB to avoid overwhelming AI)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("assessment file too large: %d bytes (max: %d bytes / 10MB)", fileInfo.Size(), maxFileSize)
	}

	// Check file is not empty
	if fileInfo.Size() == 0 {
		return nil, fmt.Errorf("assessment file is empty: %s", config.AssessmentPath)
	}

	// Make catalog path absolute if it's a local file (not a URL)
	if !strings.HasPrefix(config.CatalogURL, "http://") && !strings.HasPrefix(config.CatalogURL, "https://") {
		if !filepath.IsAbs(config.CatalogURL) {
			config.CatalogURL = filepath.Join(config.WorkspaceRoot, config.CatalogURL)
		}
	}

	return config, nil
}

// validateProfile validates the generated profile structure.
func validateProfile(profile *oscalTypes.Profile) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}

	if profile.UUID == "" {
		return fmt.Errorf("profile missing UUID")
	}

	if profile.Metadata.Title == "" {
		return fmt.Errorf("profile missing title")
	}

	if len(profile.Imports) == 0 {
		return fmt.Errorf("profile missing imports")
	}

	controlIDs := oscal.GetProfileControlIDs(profile)
	if len(controlIDs) == 0 {
		return fmt.Errorf("profile has no controls")
	}

	// Note: Control ID format validation is handled by go-oscal library
	// and catalog validation. No need for custom format checks here.

	return nil
}

// validateProfileWithCatalog validates that all control IDs in the profile
// exist in the specified catalog.
func validateProfileWithCatalog(profile *oscalTypes.Profile, catalogURL string, logger *logging.Logger) error {
	if logger != nil {
		logger.Debug("Loading catalog for validation",
			zap.String("catalog", catalogURL))
	}

	// Load catalog
	catalog, err := oscal.LoadCatalog(catalogURL)
	if err != nil {
		return fmt.Errorf("failed to load catalog: %w", err)
	}

	if logger != nil {
		logger.Debug("Catalog loaded successfully",
			zap.String("uuid", catalog.UUID))
	}

	// Extract control IDs from profile
	profileControlIDs := oscal.GetProfileControlIDs(profile)

	// Validate control IDs against catalog
	if err := oscal.ValidateControlIDsAgainstCatalog(profileControlIDs, catalog); err != nil {
		return fmt.Errorf("control validation failed: %w", err)
	}

	if logger != nil {
		logger.Debug("All control IDs validated against catalog",
			zap.Int("count", len(profileControlIDs)))
	}

	return nil
}


// showHelp displays help information.
func showHelp() {
	help := `Usage: create risk-profile <assessment-file> [flags]

Create OSCAL profile from risk assessment using AI for the entire solution

Arguments:
  assessment-file    Path to the risk assessment document (markdown, text, etc.)

Optional Flags:
      --catalog <url>      Catalog URL for control selection and validation (default: NIST 800-53 Rev 5)
  -o, --output <path>      Custom output path for the profile file
      --force              Overwrite existing profile file
  -d, --debug              Save intermediate outputs to out/logs/risk/
      --max-retries <n>    Maximum AI generation retries (default: 3)
  -h, --help               Show this help message

Output:
  Profile is saved to: specs/.risk-controls/risk-profile.json

Examples:
  # Create profile from assessment document
  create risk-profile docs/security-assessment.md

  # Use local catalog for validation
  create risk-profile assessment.md --catalog catalogs/nist-800-53.json

  # Force overwrite existing profile
  create risk-profile assessment.md --force

  # Use debug mode to inspect AI reasoning
  create risk-profile assessment.md --debug
`
	log.Info(help)
}
