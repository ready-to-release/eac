package riskprofile

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	"github.com/ready-to-release/eac/go/cli/eac/internal/risk/oscal"
	coreai "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/logging"
)

// defaultProfilePrompt is the fallback prompt when .eac/templates/ai/risk-profile/risk-profile.md is not found.
const defaultProfilePrompt = `# Risk Assessment to NIST 800-53 Controls Mapper

You are a security controls analyst. Analyze the risk assessment document and identify NIST 800-53 controls.

Return a JSON array of control IDs (lowercase): ["ac-2", "ia-2", "si-10"]`

// generateProfile generates an OSCAL profile using AI.
func generateProfile(config *Config, assessmentContent string, catalog *oscalTypes.Catalog) (*oscalTypes.Profile, error) {
	// Check for mock response
	if mockAIResponse != "" {
		return parseProfileFromAI(mockAIResponse, config, catalog)
	}
	if mock, ok := coreai.GetMockResponse("risk-profile"); ok {
		return parseProfileFromAI(mock, config, catalog)
	}

	// Extract available control IDs from catalog
	availableControls, err := oscal.ExtractControlIDs(catalog)
	if err != nil {
		return nil, fmt.Errorf("failed to extract control IDs from catalog: %w", err)
	}

	log.Debugf("Extracted controls from catalog: count=%d", len(availableControls))

	// Build AI prompt with available controls
	prompt := buildProfilePrompt(config.WorkspaceRoot, assessmentContent, config.CatalogURL, availableControls)

	log.Debugf("AI prompt built: length=%d", len(prompt))

	// Call AI using two-phase generation
	ctx := context.Background()
	response, err := callAIWithRetry(ctx, prompt, config)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	log.Debugf("AI response received: length=%d", len(response))

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

// buildProfilePrompt constructs the AI prompt for profile generation using templates.
func buildProfilePrompt(workspaceRoot, assessmentContent, catalogURL string, availableControls []string) string {
	// Load prompt template with three-tier priority:
	// 1. Command flag (not applicable - internal function)
	// 2. Team override (.eac/templates/ai/risk-profile/risk-profile.md)
	// 3. System default (templates/ai/risk-profile/risk-profile.md)
	loader := coreai.NewContractLoader(workspaceRoot, coreai.TypeRiskProfile, "")
	promptTemplate, _, err := loader.LoadPrompt("", defaultProfilePrompt)
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
	renderedPrompt, err := coreai.BuildPromptWithTemplate(
		promptTemplate,
		nil, // No contract needed
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

// callAIWithRetry calls the AI provider using two-phase generation with retry.
func callAIWithRetry(ctx context.Context, prompt string, config *Config) (string, error) {
	// Show progress message (AI calls can take time)
	log.Info("  → Waiting for AI response...")

	// Create executor
	executor := ai.NewExecutor(config.WorkspaceRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match domain.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Load AI config for retry strategy
	aiConfig, err := coreai.LoadAIConfig(config.WorkspaceRoot)
	if err != nil {
		log.Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil
	}

	// Build retry configuration using factory (validator auto-loaded for FormatOSCALProfile)
	retryConfig, err := coreai.BuildRetryConfig(
		coreai.TypeRiskProfile,
		coreai.FormatOSCALProfile, // Generate and validate OSCAL profile structure
		executorAdapter,
		nil, // Let BuildRetryConfig auto-load OSCAL profile validator
		config.WorkspaceRoot,
		aiConfig,
		coreai.WithDebug(config.Debug),
		coreai.WithLogger(logging.C().Zap()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build retry config: %w", err)
	}

	// Generate with retry
	result, err := coreai.GenerateWithRetry(ctx, retryConfig, prompt)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Log provider information
	log.Debugf("AI call completed: provider=%s, response_length=%d, attempts=%d",
		result.ProviderName, len(result.Output), result.Attempts)
	if config.Debug {
		log.Infof("  → AI provider used: %s", result.ProviderName)
	}

	// Check validation errors
	if len(result.ValidationErrors) > 0 {
		log.Warnf("  ⚠ Generated output has %d validation issues", len(result.ValidationErrors))
		for _, verr := range result.ValidationErrors {
			log.Warnf("Validation error: code=%s, message=%s", verr.GetCode(), verr.Message)
		}
	}

	log.Infof("  → Received response (%d chars, %d attempts)", len(result.Output), result.Attempts)
	return result.Output, nil
}

// parseProfileFromAI parses AI response (full OSCAL profile JSON) into an OSCAL profile.
func parseProfileFromAI(response string, config *Config, catalog *oscalTypes.Catalog) (*oscalTypes.Profile, error) {
	// Parse the OSCAL profile JSON directly (AI generates full OSCAL structure)
	var oscalDoc oscalTypes.OscalModels
	if err := json.Unmarshal([]byte(response), &oscalDoc); err != nil {
		return nil, fmt.Errorf("failed to parse OSCAL profile JSON: %w", err)
	}

	if oscalDoc.Profile == nil {
		return nil, fmt.Errorf("AI response does not contain a valid OSCAL profile")
	}

	return oscalDoc.Profile, nil
}

// filterInvalidControls removes control IDs that don't exist in the catalog.
// This handles cases where AI returns valid NIST controls not in custom catalogs.
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
		log.Warnf("Controls filtered out (not in catalog): removed=%v, kept=%v", removedIDs, filteredIDs)
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
