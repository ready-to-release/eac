package contract

import (
	"context"
	"fmt"
)

// AIExecutor interface abstracts AI execution for contract-based generation
type AIExecutor interface {
	Execute(ctx context.Context, prompt string, opts ...interface{}) (string, error)
}

// GenerateWithContract orchestrates contract-based AI generation with validation
//
// This is the main entry point for generating output with contracts.
// It follows this workflow:
//   1. Build prompt from contract and template
//   2. Execute AI generation
//   3. Apply anti-corruption filter
//   4. Validate output against contract
//   5. Return validated output or errors
//
// Parameters:
//   - executor: AI executor instance
//   - loader: Contract loader with loaded contracts
//   - promptTemplate: Template string with {{.Field}} placeholders
//   - customData: Custom data for template (e.g., tags, taxonomy)
//   - contentMarker: Content start marker for anti-corruption (e.g., "Feature:")
//   - validator: Contract-specific validator implementation
//   - validationContext: Context data for validation
//
// Returns:
//   - Cleaned output string
//   - Validation errors (empty if valid)
//   - Error if generation or validation fails critically
type GenerateConfig struct {
	Executor          AIExecutor
	Loader            *SpecContractLoader
	PromptTemplate    string
	CustomData        map[string]string
	ContentMarker     string
	Validator         SpecValidator
	ValidationContext *ValidationContext
}

// GenerateWithContract generates and validates output using contracts
func GenerateWithContract(ctx context.Context, cfg *GenerateConfig) (string, []ValidationError, error) {
	// Load contracts
	contract, err := cfg.Loader.LoadContract()
	if err != nil {
		return "", nil, fmt.Errorf("failed to load contract: %w", err)
	}

	antiCorruption, err := cfg.Loader.LoadAntiCorruptionRules()
	if err != nil {
		return "", nil, fmt.Errorf("failed to load anti-corruption rules: %w", err)
	}

	// Build prompt with template
	prompt, err := BuildPromptWithTemplate(
		cfg.PromptTemplate,
		contract,
		antiCorruption,
		cfg.CustomData,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Execute AI generation
	rawOutput, err := cfg.Executor.Execute(ctx, prompt)
	if err != nil {
		return "", nil, fmt.Errorf("AI execution failed: %w", err)
	}

	// Apply anti-corruption filter
	cleanedOutput := ApplyWithFallback(rawOutput, antiCorruption, cfg.ContentMarker)

	// Validate output against contract
	var validationErrors []ValidationError
	if cfg.Validator != nil {
		validationErrors = cfg.Validator.Validate(cleanedOutput, cfg.ValidationContext.ToMap())
	}

	return cleanedOutput, validationErrors, nil
}

// MustValidate checks validation errors and returns error if any critical errors exist
func MustValidate(validationErrors []ValidationError) error {
	var criticalErrors []ValidationError
	for _, verr := range validationErrors {
		if verr.Severity == "error" {
			criticalErrors = append(criticalErrors, verr)
		}
	}

	if len(criticalErrors) > 0 {
		return fmt.Errorf("contract validation failed with %d error(s)", len(criticalErrors))
	}

	return nil
}

// FormatValidationErrors formats validation errors for display
func FormatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var result string
	errorCount := 0
	warningCount := 0

	for _, verr := range errors {
		if verr.Severity == "error" {
			errorCount++
		} else {
			warningCount++
		}
	}

	if errorCount > 0 {
		result += fmt.Sprintf("❌ Found %d contract violation(s):\n\n", errorCount)
	}
	if warningCount > 0 {
		result += fmt.Sprintf("⚠️  Found %d warning(s):\n\n", warningCount)
	}

	for _, verr := range errors {
		icon := "❌"
		if verr.Severity == "warning" {
			icon = "⚠️ "
		}
		result += fmt.Sprintf("%s %s\n", icon, verr.Error())
	}

	return result
}
