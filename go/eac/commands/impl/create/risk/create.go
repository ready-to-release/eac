// Command: create risk
// Short: Create OSCAL profile from risk assessment using AI
// Long: The create risk command analyzes a risk assessment document and generates an OSCAL profile
// Long: selecting appropriate controls from NIST 800-53 or a custom catalog. The AI extracts risks
// Long: from the assessment and maps them to security controls.
// Long:
// Long: The generated profile is saved to specs/risk-controls/<module>.profile.json for version control.
// Long: Use --debug to inspect intermediate outputs and AI reasoning.
// Flag.module: type=string, shorthand=m, usage=Target module for the profile. If not provided, uses 'common' as the module name
// Flag.catalog: type=string, default=https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json, usage=Catalog URL for control selection (NIST 800-53 Rev 5 by default)
// Flag.output: type=string, shorthand=o, usage=Custom output path for the profile file
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite existing profile file
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs to out/logs/risk/
// Flag.max-retries: type=int, default=3, usage=Maximum AI generation retries on validation failure
// Args: files
package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ready-to-release/eac/go/eac/ai"
	"github.com/ready-to-release/eac/go/eac/ai/providers"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(CreateRisk)
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

// Config holds configuration for create risk command.
type Config struct {
	AssessmentPath string
	Module         string
	CatalogURL     string
	OutputPath     string
	Force          bool
	Debug          bool
	MaxRetries     int
	WorkspaceRoot  string
	Logger         *logging.Logger
}

// CreateRisk is the entry point for the create risk command.
func CreateRisk() int {
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
		config.Logger.Info("Starting create risk",
			zap.String("assessment", config.AssessmentPath),
			zap.String("module", config.Module),
			zap.Bool("debug", config.Debug))
	}

	// Read assessment file
	assessmentContent, err := os.ReadFile(config.AssessmentPath)
	if err != nil {
		log.Errorf("Error reading assessment file: %v", err)
		return 1
	}

	// Determine output path
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = oscal.GetProfilePath(config.WorkspaceRoot, config.Module)
	}

	// Check if file exists
	if !config.Force {
		if _, err := os.Stat(outputPath); err == nil {
			log.Errorf("Error: Profile already exists: %s", outputPath)
			log.Error("Use --force to overwrite")
			return 1
		}
	}

	// Generate profile using AI with retry logic
	log.Infof("Analyzing assessment and generating OSCAL profile...")

	var profile *oscal.Profile
	var lastErr error

	for attempt := 1; attempt <= config.MaxRetries; attempt++ {
		if attempt > 1 {
			log.Infof("Retry %d/%d...", attempt, config.MaxRetries)
		}

		profile, err = generateProfile(config, string(assessmentContent))
		if err != nil {
			lastErr = err
			if config.Logger != nil {
				config.Logger.Warn("Profile generation failed",
					zap.Int("attempt", attempt),
					zap.Error(err))
			}
			continue
		}

		// Validate generated profile
		if err := validateProfile(profile); err != nil {
			lastErr = err
			if config.Logger != nil {
				config.Logger.Warn("Profile validation failed",
					zap.Int("attempt", attempt),
					zap.Error(err))
			}
			continue
		}

		// Success
		break
	}

	if profile == nil {
		log.Errorf("Failed to generate valid profile after %d attempts: %v", config.MaxRetries, lastErr)
		return 1
	}

	// Write profile
	if err := oscal.WriteProfile(outputPath, profile); err != nil {
		log.Errorf("Error writing profile: %v", err)
		return 1
	}

	// Report success
	controlIDs := profile.GetControlIDs()
	log.Info("")
	log.Infof("Created OSCAL profile: %s", outputPath)
	log.Infof("  Module: %s", config.Module)
	log.Infof("  Controls: %d (%s)", len(controlIDs), strings.Join(controlIDs, ", "))
	log.Infof("  Catalog: %s", profile.GetCatalogURL())
	log.Info("")
	log.Info("Next steps:")
	log.Infof("  1. Add @control(%s) tags to your .feature files", controlIDs[0])
	log.Infof("  2. Run: create risk-assess %s", config.Module)

	if config.Logger != nil {
		config.Logger.Info("Profile created successfully",
			zap.String("path", outputPath),
			zap.Int("controls", len(controlIDs)))
	}

	return 0
}

// parseConfig parses command line configuration.
func parseConfig() (*Config, error) {
	args := os.Args[3:] // Skip program name, "create", and "risk"

	config := &Config{
		Module:     "common",
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

		case arg == "--module" || arg == "-m":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--module requires a value")
			}
			config.Module = args[i+1]
			i += 2

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

		case arg == "--force" || arg == "-f":
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

	// Check file exists
	if _, err := os.Stat(config.AssessmentPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("assessment file not found: %s", config.AssessmentPath)
	}

	return config, nil
}

// generateProfile generates an OSCAL profile using AI.
func generateProfile(config *Config, assessmentContent string) (*oscal.Profile, error) {
	// Check for mock response
	if mockAIResponse != "" {
		return parseProfileFromAI(mockAIResponse, config)
	}
	if mock, ok := aimock.GetMockResponse("risk-create"); ok {
		return parseProfileFromAI(mock, config)
	}

	// Build AI prompt
	prompt := buildProfilePrompt(config.WorkspaceRoot, assessmentContent, config.CatalogURL)

	if config.Logger != nil {
		config.Logger.Debug("AI prompt built",
			zap.Int("prompt_length", len(prompt)))
	}

	// Call AI
	ctx := context.Background()
	response, err := callAI(ctx, prompt, config)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	if config.Logger != nil {
		config.Logger.Debug("AI response received",
			zap.Int("response_length", len(response)))
	}

	return parseProfileFromAI(response, config)
}

// buildProfilePrompt constructs the AI prompt for profile generation.
func buildProfilePrompt(workspaceRoot, assessmentContent, catalogURL string) string {
	// Load prompt from .r2r/eac/ai/risk-create/profile.md
	loader := contracts.NewContractLoader(workspaceRoot, "ai/risk-create", "")
	systemPrompt, _, err := loader.LoadPrompt("profile.md", defaultProfilePrompt)
	if err != nil {
		systemPrompt = defaultProfilePrompt
	}

	var builder strings.Builder
	builder.WriteString(systemPrompt)
	builder.WriteString("\n\n## Assessment Document\n\n")
	builder.WriteString(assessmentContent)
	builder.WriteString("\n\n## Catalog URL\n\n")
	builder.WriteString(catalogURL)
	builder.WriteString("\n\n## Instructions\n\n")
	builder.WriteString("Analyze the assessment document and identify the NIST 800-53 controls that address the identified risks. ")
	builder.WriteString("Return a JSON array of control IDs (lowercase, e.g., 'ac-2', 'si-10').")
	return builder.String()
}

// defaultProfilePrompt is the fallback prompt when .r2r/eac/ai/risk-create/profile.md is not found.
const defaultProfilePrompt = `# Risk Assessment to NIST 800-53 Controls Mapper

You are a security controls analyst. Analyze the risk assessment document and identify NIST 800-53 controls.

Return a JSON array of control IDs (lowercase): ["ac-2", "ia-2", "si-10"]`

// callAI calls the AI provider.
func callAI(ctx context.Context, prompt string, config *Config) (string, error) {
	// Create executor
	executor := ai.NewExecutor(config.WorkspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Configure options
	opts := []ai.Option{}
	if config.Debug {
		opts = append(opts, ai.WithDebug(true))
	}

	// Execute
	response, err := executor.Execute(ctx, prompt, opts...)
	if err != nil {
		return "", fmt.Errorf("AI execution failed: %w", err)
	}

	return response, nil
}

// parseProfileFromAI parses AI response into an OSCAL profile.
func parseProfileFromAI(response string, config *Config) (*oscal.Profile, error) {
	// Extract control IDs from AI response
	controlIDs, err := extractControlIDs(response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract control IDs: %w", err)
	}

	if len(controlIDs) == 0 {
		return nil, fmt.Errorf("AI did not identify any controls")
	}

	// Generate UUID
	profileUUID := uuid.New().String()

	// Create profile
	title := fmt.Sprintf("Risk Profile for %s", config.Module)
	profile := oscal.NewProfile(profileUUID, title, config.CatalogURL, controlIDs)

	return profile, nil
}

// extractControlIDs extracts control IDs from AI response.
func extractControlIDs(response string) ([]string, error) {
	// Try to find JSON array in response
	response = extractJSONFromMarkdown(response)

	// Try parsing as JSON array
	var controlIDs []string
	if err := json.Unmarshal([]byte(response), &controlIDs); err == nil {
		return normalizeControlIDs(controlIDs), nil
	}

	// Try parsing as JSON object with controls field
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(response), &obj); err == nil {
		// Look for common field names
		for _, field := range []string{"controls", "control_ids", "controlIds", "ids"} {
			if ids, ok := obj[field].([]interface{}); ok {
				var result []string
				for _, id := range ids {
					if s, ok := id.(string); ok {
						result = append(result, s)
					}
				}
				return normalizeControlIDs(result), nil
			}
		}
	}

	// Fallback: extract control IDs using pattern matching
	return extractControlIDsFromText(response), nil
}

// extractJSONFromMarkdown extracts JSON from markdown code blocks.
func extractJSONFromMarkdown(text string) string {
	// Look for ```json ... ``` blocks
	jsonStart := strings.Index(text, "```json")
	if jsonStart != -1 {
		jsonStart += 7
		jsonEnd := strings.Index(text[jsonStart:], "```")
		if jsonEnd != -1 {
			return strings.TrimSpace(text[jsonStart : jsonStart+jsonEnd])
		}
	}

	// Look for ``` ... ``` blocks
	codeStart := strings.Index(text, "```")
	if codeStart != -1 {
		codeStart += 3
		codeEnd := strings.Index(text[codeStart:], "```")
		if codeEnd != -1 {
			extracted := strings.TrimSpace(text[codeStart : codeStart+codeEnd])
			if strings.HasPrefix(extracted, "[") || strings.HasPrefix(extracted, "{") {
				return extracted
			}
		}
	}

	return text
}

// extractControlIDsFromText extracts control IDs using pattern matching.
func extractControlIDsFromText(text string) []string {
	// Common NIST 800-53 control ID patterns
	patterns := []string{
		"ac-", "at-", "au-", "ca-", "cm-", "cp-", "ia-", "ir-",
		"ma-", "mp-", "pe-", "pl-", "pm-", "ps-", "pt-", "ra-",
		"sa-", "sc-", "si-", "sr-",
	}

	var controlIDs []string
	seen := make(map[string]bool)

	text = strings.ToLower(text)

	for _, prefix := range patterns {
		idx := 0
		for {
			pos := strings.Index(text[idx:], prefix)
			if pos == -1 {
				break
			}
			idx += pos

			// Extract the full control ID (e.g., "ac-2", "si-10")
			end := idx + len(prefix)
			for end < len(text) && (text[end] >= '0' && text[end] <= '9') {
				end++
			}

			if end > idx+len(prefix) {
				controlID := text[idx:end]
				if !seen[controlID] {
					seen[controlID] = true
					controlIDs = append(controlIDs, controlID)
				}
			}

			idx++
		}
	}

	return controlIDs
}

// normalizeControlIDs normalizes control IDs to lowercase.
func normalizeControlIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := make(map[string]bool)

	for _, id := range ids {
		normalized := strings.ToLower(strings.TrimSpace(id))
		if normalized != "" && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}

// validateProfile validates the generated profile.
func validateProfile(profile *oscal.Profile) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}

	if profile.Profile.UUID == "" {
		return fmt.Errorf("profile missing UUID")
	}

	if profile.Profile.Metadata.Title == "" {
		return fmt.Errorf("profile missing title")
	}

	if len(profile.Profile.Imports) == 0 {
		return fmt.Errorf("profile missing imports")
	}

	controlIDs := profile.GetControlIDs()
	if len(controlIDs) == 0 {
		return fmt.Errorf("profile has no controls")
	}

	// Validate control ID format
	for _, id := range controlIDs {
		if !isValidControlID(id) {
			return fmt.Errorf("invalid control ID format: %s", id)
		}
	}

	return nil
}

// isValidControlID checks if a control ID is valid NIST 800-53 format.
func isValidControlID(id string) bool {
	parts := strings.Split(id, "-")
	if len(parts) != 2 {
		return false
	}

	// Family prefix should be 2 letters
	if len(parts[0]) != 2 {
		return false
	}

	// Number part should be numeric
	for _, c := range parts[1] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// showHelp displays help information.
func showHelp() {
	help := `Usage: create risk <assessment-file> [flags]

Create OSCAL profile from risk assessment using AI

Arguments:
  assessment-file    Path to the risk assessment document (markdown, text, etc.)

Flags:
  -m, --module <name>      Target module for the profile (default: common)
      --catalog <url>      Catalog URL for control selection (default: NIST 800-53 Rev 5)
  -o, --output <path>      Custom output path for the profile file
  -f, --force              Overwrite existing profile file
  -d, --debug              Save intermediate outputs to out/logs/risk/
      --max-retries <n>    Maximum AI generation retries (default: 3)
  -h, --help               Show this help message

Output:
  specs/risk-controls/<module>.profile.json

Examples:
  # Create profile from assessment document
  create risk .docs/assessments/security-assessment.md

  # Create profile for specific module
  create risk assessment.md --module billing-service

  # Force overwrite existing profile
  create risk assessment.md --force

  # Use custom catalog
  create risk assessment.md --catalog https://example.com/catalog.json
`
	log.Info(help)
}
