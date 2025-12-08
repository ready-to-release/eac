package riskprofile

import (
	"context"
	"fmt"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/eac/ai"
	"github.com/ready-to-release/eac/go/eac/ai/providers"
	"github.com/ready-to-release/eac/go/eac/commands/internal/risk/oscal"
	aimock "github.com/ready-to-release/eac/go/eac/core/ai"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"go.uber.org/zap"
)

// defaultProfilePrompt is the fallback prompt when .r2r/eac/ai/risk-profile/profile.md is not found.
const defaultProfilePrompt = `# Risk Assessment to NIST 800-53 Controls Mapper

You are a security controls analyst. Analyze the risk assessment document and identify NIST 800-53 controls.

Return a JSON array of control IDs (lowercase): ["ac-2", "ia-2", "si-10"]`

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

// buildProfilePrompt constructs the AI prompt for profile generation using templates.
func buildProfilePrompt(workspaceRoot, assessmentContent, catalogURL string, availableControls []string) string {
	// Load prompt template from .r2r/eac/ai/prompts/risk/profile.md
	loader := contracts.NewContractLoader(workspaceRoot, "ai/risk-profile", "")
	promptTemplate, _, err := loader.LoadPrompt("risk/profile.md", defaultProfilePrompt)
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
