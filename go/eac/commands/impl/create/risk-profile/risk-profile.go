// Command: create risk-profile
// Short: Create OSCAL profile from risk assessment using AI
// Long: The create risk-profile command analyzes a risk assessment document and generates an OSCAL profile
// Long: selecting appropriate controls from a custom catalog. The AI extracts risks
// Long: from the assessment and maps them to controls for the entire solution.
// Long:
// Long: The generated profile is saved to specs/.risk-controls/risk-profile.json for version control.
// Long: Use --debug to inspect intermediate outputs and AI reasoning.
// Flag.catalog: type=string, usage=Catalog URL for control selection and validation (default: NIST 800-53 Rev5)
// Flag.output: type=string, shorthand=o, usage=Custom output path for the profile file
// Flag.force: type=bool, shorthand=f, default=false, usage=Overwrite existing profile file
// Flag.debug: type=bool, shorthand=d, default=false, usage=Save intermediate outputs to out/logs/risk/
// Flag.max-retries: type=int, default=3, usage=Maximum AI generation retries on validation failure
// Args: file
package riskprofile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/ai"
	"github.com/ready-to-release/eac/go/eac/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
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

	// Check file exists
	if _, err := os.Stat(config.AssessmentPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("assessment file not found: %s", config.AssessmentPath)
	}

	// Make catalog path absolute if it's a local file (not a URL)
	if !strings.HasPrefix(config.CatalogURL, "http://") && !strings.HasPrefix(config.CatalogURL, "https://") {
		if !filepath.IsAbs(config.CatalogURL) {
			config.CatalogURL = filepath.Join(config.WorkspaceRoot, config.CatalogURL)
		}
	}

	return config, nil
}

// generateProfile generates an OSCAL profile using AI.
func generateProfile(config *Config, assessmentContent string, catalog *oscalTypes.Catalog) (*oscalTypes.Profile, error) {
	// Check for mock response
	if mockAIResponse != "" {
		return parseProfileFromAI(mockAIResponse, config, catalog)
	}
	if mock, ok := aimock.GetMockResponse("risk-profile"); ok {
		return parseProfileFromAI(mock, config, catalog)
	}

	// Extract available control IDs from catalog
	availableControls, err := oscal.ExtractControlIDs(catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to extract control IDs from catalog: %w", err)
	}

	if config.Logger != nil {
		config.Logger.Debug("Extracted controls from catalog",
			zap.Int("control_count", len(availableControls)))
	}

	// Build AI prompt with available controls
	prompt := buildProfilePrompt(config.WorkspaceRoot, assessmentContent, config.CatalogURL, availableControls)

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

	profile, err := parseProfileFromAI(response, config, catalog)
	if err != nil {
		return nil, err
	}

	// Filter out any control IDs that don't exist in the catalog
	// This handles cases where AI returns valid NIST controls not in custom catalogs
	profile, err = filterInvalidControls(profile, catalog, config)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

// filterInvalidControls removes control IDs that don't exist in the catalog
func filterInvalidControls(profile *oscalTypes.Profile, catalog *oscalTypes.Catalog, config *Config) (*oscalTypes.Profile, error) {
	originalIDs := oscal.GetProfileControlIDs(profile)
	if len(originalIDs) == 0 {
		return profile, nil
	}

	// Extract valid control IDs from catalog
	validControlIDs, err := oscal.ExtractControlIDs(catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to extract valid control IDs: %w", err)
	}

	// Build lookup map for O(1) validation
	validMap := make(map[string]bool)
	for _, id := range validControlIDs {
		validMap[strings.ToLower(id)] = true
	}

	// Filter controls, keeping only valid ones
	var filteredIDs []string
	var removedIDs []string
	for _, id := range originalIDs {
		normalizedID := strings.ToLower(id)
		if validMap[normalizedID] {
			filteredIDs = append(filteredIDs, id)
		} else {
			removedIDs = append(removedIDs, id)
		}
	}

	// Warn about removed controls
	if len(removedIDs) > 0 {
		log.Warnf("  ⚠ Filtered out %d control(s) not in catalog: %s", len(removedIDs), strings.Join(removedIDs, ", "))
		if config.Logger != nil {
			config.Logger.Warn("Controls filtered out (not in catalog)",
				zap.Strings("removed", removedIDs),
				zap.Strings("kept", filteredIDs))
		}
	}

	// If no valid controls remain, return error
	if len(filteredIDs) == 0 {
		return nil, fmt.Errorf("AI returned controls but none exist in catalog (removed: %s)", strings.Join(removedIDs, ", "))
	}

	// Rebuild profile with filtered controls
	title := "Solution Risk Profile"
	controlInfo, err := oscal.GetControlInfo(catalog, filteredIDs)
	if err != nil {
		log.Warnf("Could not fetch control info: %v, creating profile without back-matter", err)
		oscalDoc, err := oscal.NewProfileDocument(title, oscal.GetProfileCatalogURL(profile), filteredIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create filtered profile: %w", err)
		}
		return oscalDoc.Profile, nil
	}

	oscalDoc, err := oscal.NewProfileDocumentWithInfo(title, oscal.GetProfileCatalogURL(profile), controlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create filtered profile: %w", err)
	}

	return oscalDoc.Profile, nil
}

// buildProfilePrompt constructs the AI prompt for profile generation using templates.
func buildProfilePrompt(workspaceRoot, assessmentContent, catalogURL string, availableControls []string) string {
	// Load prompt template from .r2r/eac/ai/risk-profile/profile.md
	loader := contracts.NewContractLoader(workspaceRoot, "ai/risk-profile", "")
	promptTemplate, _, err := loader.LoadPrompt("profile.md", defaultProfilePrompt)
	if err != nil {
		promptTemplate = defaultProfilePrompt
	}

	// Prepare custom data for template
	customData := map[string]string{
		"AvailableControls": strings.Join(availableControls, ", "),
		"ControlCount":      fmt.Sprintf("%d", len(availableControls)),
		"CatalogURL":        catalogURL,
	}

	// Build prompt with template replacements
	renderedPrompt, err := contracts.BuildPromptWithTemplate(
		promptTemplate,
		nil, // No contract needed
		nil, // No anti-corruption rules needed
		customData,
	)
	if err != nil {
		// Fallback to default prompt if template rendering fails
		log.Warnf("Failed to render prompt template, using default: %v", err)
		renderedPrompt = defaultProfilePrompt
	}

	// Append assessment document
	var builder strings.Builder
	builder.WriteString(renderedPrompt)
	builder.WriteString("\n\n## Assessment Document\n\n")
	builder.WriteString(assessmentContent)

	return builder.String()
}

// defaultProfilePrompt is the fallback prompt when .r2r/eac/ai/risk-profile/profile.md is not found.
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

	// Show progress message (AI calls can take time)
	log.Info("  → Waiting for AI response...")

	// Execute
	response, err := executor.Execute(ctx, prompt, opts...)

	// Log provider information after execution
	provider := executor.GetLastUsedProvider()
	if provider != nil {
		if config.Logger != nil {
			config.Logger.Debug("AI call completed",
				zap.String("provider", provider.Name()),
				zap.Int("response_length", len(response)),
				zap.Error(err))
		}
		if config.Debug {
			log.Infof("  → AI provider used: %s", provider.Name())
		}
	}

	if err != nil {
		if config.Logger != nil {
			config.Logger.Error("AI execution failed",
				zap.Error(err),
				zap.String("provider", provider.Name()))
		}
		return "", fmt.Errorf("AI execution failed: %w", err)
	}

	// Check if response is empty or very short
	if len(response) == 0 {
		if config.Logger != nil {
			config.Logger.Warn("AI returned empty response",
				zap.String("provider", provider.Name()),
				zap.Int("prompt_length", len(prompt)))
		}
		log.Warn("  ⚠ AI returned empty response (this may be an API issue, retrying...)")
		return "", fmt.Errorf("AI provider returned empty response")
	}

	if len(response) < 10 {
		if config.Logger != nil {
			config.Logger.Warn("AI returned very short response",
				zap.String("provider", provider.Name()),
				zap.Int("response_length", len(response)),
				zap.String("response_preview", response))
		}
	}

	log.Infof("  → Received response (%d chars)", len(response))
	return response, nil
}

// parseProfileFromAI parses AI response into an OSCAL profile.
func parseProfileFromAI(response string, config *Config, catalog *oscalTypes.Catalog) (*oscalTypes.Profile, error) {
	// Extract control IDs from AI response
	controlIDs, err := extractControlIDs(response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract control IDs: %w", err)
	}

	if len(controlIDs) == 0 {
		// Provide more helpful error with response preview
		responsePreview := response
		if len(responsePreview) > 200 {
			responsePreview = responsePreview[:200] + "..."
		}
		return nil, fmt.Errorf("AI did not identify any controls (response: %s)", responsePreview)
	}

	// Fetch control information from catalog for back-matter
	controlInfo, err := oscal.GetControlInfo(catalog, controlIDs)
	if err != nil {
		log.Warnf("Could not fetch control info: %v, creating profile without back-matter", err)
		// Fallback to basic profile without control info
		title := "Solution Risk Profile"
		oscalDoc, err := oscal.NewProfileDocument(title, config.CatalogURL, controlIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
		return oscalDoc.Profile, nil
	}

	// Create profile with control information in back-matter
	title := "Solution Risk Profile"
	oscalDoc, err := oscal.NewProfileDocumentWithInfo(title, config.CatalogURL, controlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return oscalDoc.Profile, nil
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
