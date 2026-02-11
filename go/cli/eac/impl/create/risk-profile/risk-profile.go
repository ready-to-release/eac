package riskprofile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	eacConfig "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/repository"
)

type createRiskProfileCommand struct{}

var _ core.SimpleCommandPort = (*createRiskProfileCommand)(nil)

// Commands returns all command ports provided by this package.
func Commands() []core.CommandPort {
	return []core.CommandPort{
		&createRiskProfileCommand{},
	}
}

func (c *createRiskProfileCommand) Name() string { return "create risk-profile" }

func (c *createRiskProfileCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "create-risk-profile",
		Short:         "Create OSCAL profile from risk assessment using AI",
		Long:          "The create risk-profile command analyzes a risk assessment document and generates an OSCAL profile\nselecting appropriate controls from a custom catalog. The AI extracts risks\nfrom the assessment and maps them to controls for the entire solution.\n\nThe generated profile is saved to specs/.risk-controls/risk-profile.json for version control.\nUse --debug to inspect intermediate outputs and AI reasoning.\n\nExpected Output:\n- OSCAL profile JSON file selecting controls from catalog\n- AI reasoning for control selection in debug output",
		Args:          "file",
		Flags: []core.FlagSpec{
			{Name: "catalog", Type: "string", Usage: "Catalog URL for control selection and validation (default: NIST 800-53 Rev5)"},
			{Name: "output", Shorthand: "o", Type: "string", Usage: "Custom output path for the profile file"},
			{Name: "force", Shorthand: "f", Type: "bool", DefaultValue: "false", Usage: "Overwrite existing profile file"},
			{Name: "debug", Shorthand: "d", Type: "bool", DefaultValue: "false", Usage: "Save intermediate outputs to out/commands.log"},
		},
	}
}

func (c *createRiskProfileCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return CreateRiskProfile()
}

var log = logging.C()

// Config holds configuration for create risk-profile command.
type Config struct {
	AssessmentPath string
	CatalogURL     string
	OutputPath     string
	Force          bool
	Debug          bool
	WorkspaceRoot  string
}

// CreateRiskProfile is the entry point for the create risk-profile command.
func CreateRiskProfile() int {
	return createRiskProfile(defaultDeps())
}

func createRiskProfile(deps *Deps) int {
	config, err := parseConfig()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Configure logging system (component loggers + file logging)
	if err := logging.ConfigureLoggingSimple(config.WorkspaceRoot, "commands", nil, config.Debug); err != nil {
		log.Warnf("Failed to configure logging: %v", err)
	}
	defer logging.CloseLogging()

	log.Infof("Starting create risk-profile: assessment=%s, debug=%v", config.AssessmentPath, config.Debug)

	// Read assessment file
	log.Infof("Reading assessment file: %s", config.AssessmentPath)
	assessmentContent, err := os.ReadFile(config.AssessmentPath)
	if err != nil {
		log.Errorf("Error reading assessment file: %v", err)
		return 1
	}
	log.Infof("Assessment file loaded (%d bytes)", len(assessmentContent))
	log.Info("")

	// Determine output path using RiskConfig or fallback
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = getDefaultProfilePath(config.WorkspaceRoot)
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

	// Generate profile using core AI logic (handles retry and validation)
	log.Info("Calling AI to analyze risks and map controls...")
	profile, err := generateProfile(deps, config, string(assessmentContent), catalog)
	if err != nil {
		log.Errorf("Failed to generate profile: %v", err)
		return 1
	}
	log.Infof("  ✓ AI returned %d controls", len(oscal.GetProfileControlIDs(profile)))
	log.Info("")

	// Validate controls against catalog (domain-specific validation)
	log.Info("Validating controls against catalog...")
	if err := oscal.ValidateControlIDsAgainstCatalog(oscal.GetProfileControlIDs(profile), catalog); err != nil {
		log.Errorf("Catalog validation failed: %v", err)
		return 1
	}
	log.Info("  ✓ All controls validated against catalog")
	log.Info("")

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

	log.Infof("Profile created successfully: path=%s, controls=%d", outputPath, len(controlIDs))

	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	args := os.Args[3:] // Skip program name, "create", and "risk-profile"

	// Validate flags before parsing
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		return nil, err
	}

	// Get workspace root first (needed for config loading)
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace root: %w", err)
	}

	config := &Config{
		CatalogURL:    getDefaultCatalogURL(workspaceRoot),
		WorkspaceRoot: workspaceRoot,
	}

	// Parse debug flag using shared package
	config.Debug = flags.ParseDebugFlag(args)

	// Parse flags and arguments
	i := 0
	for i < len(args) {
		arg := args[i]

		switch {
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
			// Already handled by shared flags package
			i++

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

// getDefaultCatalogURL returns the catalog URL from RiskConfig or falls back to NIST 800-53.
func getDefaultCatalogURL(workspaceRoot string) string {
	// Try to load RiskConfig for catalog URL
	configRoot := paths.EACConfigPath(workspaceRoot)
	riskCfg, err := eacConfig.LoadRiskConfig(workspaceRoot, configRoot)
	if err == nil && riskCfg.GetCatalogURL() != "" {
		return riskCfg.GetCatalogURL()
	}
	// Fallback to NIST 800-53 Rev5
	return oscal.NIST80053Rev5CatalogURL
}

// getDefaultProfilePath returns the default output path for the risk profile.
func getDefaultProfilePath(workspaceRoot string) string {
	// Try to load EAC config for specs path
	cfg := eacConfig.LoadOrNil(workspaceRoot)
	if cfg == nil {
		return filepath.Join(workspaceRoot, "specs", ".risk-controls", "risk-profile.json")
	}
	return filepath.Join(workspaceRoot, cfg.Repository.Paths.SpecsRoot, ".risk-controls", "risk-profile.json")
}
