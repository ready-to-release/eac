package contracts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

// Intent: Provide generic retry framework for AI generation with validation feedback
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear configuration struct shows all options
//   - Simple flow: Generate → Validate → Retry if needed
//   - Interface-based design (SpecValidator, RetryPromptBuilder)
//
// Easy to change:
//   - Pluggable validators per command
//   - Pluggable prompt builders for custom retry messages
//   - Configurable max attempts
//   - Debug mode for troubleshooting
//
// Hard to break:
//   - Validation happens on every attempt
//   - Detailed error reporting with line numbers
//   - Type-safe interfaces
//   - Tests with mock executors and validators

// RetryPromptBuilder builds retry prompts with error feedback
type RetryPromptBuilder interface {
	// BuildRetryPrompt creates a new prompt that includes:
	// - The original prompt
	// - The invalid output from previous attempt
	// - Detailed validation errors
	BuildRetryPrompt(originalPrompt string, invalidOutput string, errors []ValidationError) string
}

// OutputFixer is a function that fixes/cleans output before validation
// This runs after anti-corruption and before validation
type OutputFixer func(output string) string

// RetryConfig configures the retry behavior
type RetryConfig struct {
	// Executor performs AI generation
	Executor AIExecutor

	// Validator checks output against contract
	Validator Validator

	// PromptBuilder creates retry prompts (optional, uses default if nil)
	PromptBuilder RetryPromptBuilder

	// AntiCorruption rules for cleaning AI output
	AntiCorruption *AntiCorruptionRules

	// Fixer applies automatic fixes before validation (optional)
	// This runs after anti-corruption filter and before validation
	// Use for things like line wrapping, trailing period removal, etc.
	Fixer OutputFixer

	// ContentMarker indicates where actual content starts (e.g., "Feature:")
	// Used by anti-corruption filter
	ContentMarker string

	// MaxAttempts is the maximum number of generation attempts (default: 2)
	MaxAttempts int

	// Debug enables saving intermediate outputs for troubleshooting
	Debug bool

	// DebugOutputDir is where debug files are saved (required if Debug is true)
	DebugOutputDir string

	// ValidationContext provides additional context for validation
	ValidationContext map[string]interface{}
}

// RetryResult holds the result of generation with retry
type RetryResult struct {
	// Output is the cleaned AI output (may still have validation errors)
	Output string

	// ValidationErrors contains any validation errors (empty if valid)
	ValidationErrors []ValidationError

	// Attempts is the number of attempts made
	Attempts int

	// RetriedWithErrors indicates whether a retry was triggered
	RetriedWithErrors bool

	// ProviderName is the name of the AI provider used (e.g., "claude-cli", "openai")
	ProviderName string
}

// GenerateWithRetry generates AI output with validation and automatic retry
//
// Flow:
//  1. Attempt 1: Generate → Clean → Validate
//  2. If validation errors exist and attempts < maxAttempts:
//     - Build retry prompt with error feedback
//     - Attempt 2: Generate → Clean → Validate
//  3. Return final output with any remaining validation errors
//
// Returns:
//   - RetryResult with output, validation errors, and attempt metadata
//   - error if AI execution fails (not validation failure)
func GenerateWithRetry(ctx context.Context, cfg *RetryConfig, prompt string) (*RetryResult, error) {
	// Validate configuration
	if err := validateRetryConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid retry config: %w", err)
	}

	// Default max attempts
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 2
	}

	// Use default prompt builder if none provided
	promptBuilder := cfg.PromptBuilder
	if promptBuilder == nil {
		promptBuilder = &DefaultRetryPromptBuilder{}
	}

	var lastOutput string
	var lastValidationErrors []ValidationError
	var retriedWithErrors bool

	// Attempt loop
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Log attempt
		if attempt == 1 {
			log.Infof("Generating output (attempt %d/%d)...", attempt, maxAttempts)
		} else {
			log.Infof("Retrying with error feedback (attempt %d/%d)...", attempt, maxAttempts)
		}

		// Use retry prompt on subsequent attempts
		currentPrompt := prompt
		if attempt > 1 {
			currentPrompt = promptBuilder.BuildRetryPrompt(prompt, lastOutput, lastValidationErrors)
			retriedWithErrors = true
		}

		// Execute AI generation
		rawOutput, err := cfg.Executor.Execute(ctx, currentPrompt)
		if err != nil {
			return nil, fmt.Errorf("AI execution failed on attempt %d: %w", attempt, err)
		}

		// Debug: Save raw output
		if cfg.Debug {
			debugPath := filepath.Join(cfg.DebugOutputDir, fmt.Sprintf("attempt-%d-raw.txt", attempt))
			if err := os.WriteFile(debugPath, []byte(rawOutput), 0644); err != nil {
				log.Warnf("DEBUG: Failed to save raw output: %v", err)
			} else {
				log.Infof("DEBUG: Saved raw output to %s", debugPath)
			}
		}

		// Apply anti-corruption filter
		cleanedOutput := rawOutput
		if cfg.AntiCorruption != nil {
			cleanedOutput = ApplyWithFallback(rawOutput, cfg.AntiCorruption, cfg.ContentMarker)
		}

		// Apply fixer (line wrapping, formatting fixes, etc.)
		if cfg.Fixer != nil {
			cleanedOutput = cfg.Fixer(cleanedOutput)
		}

		// Debug: Save cleaned output
		if cfg.Debug {
			debugPath := filepath.Join(cfg.DebugOutputDir, fmt.Sprintf("attempt-%d-cleaned.txt", attempt))
			if err := os.WriteFile(debugPath, []byte(cleanedOutput), 0644); err != nil {
				log.Warnf("DEBUG: Failed to save cleaned output: %v", err)
			} else {
				log.Infof("DEBUG: Saved cleaned output to %s", debugPath)
			}
		}

		// Validate output
		var validationErrors []ValidationError
		if cfg.Validator != nil {
			validationContext := cfg.ValidationContext
			if validationContext == nil {
				validationContext = make(map[string]interface{})
			}
			validationErrors = cfg.Validator.Validate(cleanedOutput, validationContext)
		}

		// Debug: Save validation results
		if cfg.Debug && len(validationErrors) > 0 {
			debugPath := filepath.Join(cfg.DebugOutputDir, fmt.Sprintf("attempt-%d-errors.txt", attempt))
			errorText := FormatValidationErrors(validationErrors)
			if err := os.WriteFile(debugPath, []byte(errorText), 0644); err != nil {
				log.Warnf("DEBUG: Failed to save validation errors: %v", err)
			} else {
				log.Infof("DEBUG: Saved validation errors to %s", debugPath)
			}
		}

		// Store results for potential retry
		lastOutput = cleanedOutput
		lastValidationErrors = validationErrors

		// Capture provider name if available (after first execution)
		providerName := getProviderName(cfg.Executor)

		// Count critical errors (severity == "error")
		criticalErrors := countCriticalErrors(validationErrors)

		// If no errors (only warnings or nothing), we're done
		if criticalErrors == 0 {
			if len(validationErrors) > 0 {
				log.Infof("Output validated (with %d warning(s))", len(validationErrors))
			} else {
				log.Info("Output validated successfully")
			}
			return &RetryResult{
				Output:            cleanedOutput,
				ValidationErrors:  validationErrors, // Return warnings if any
				Attempts:          attempt,
				RetriedWithErrors: retriedWithErrors,
				ProviderName:      providerName,
			}, nil
		}

		// If this is the last attempt, return with errors
		if attempt >= maxAttempts {
			log.Warnf("Structure validation failed (%d issue(s) after %d attempt(s))", criticalErrors, attempt)
			return &RetryResult{
				Output:            cleanedOutput,
				ValidationErrors:  validationErrors,
				Attempts:          attempt,
				RetriedWithErrors: retriedWithErrors,
				ProviderName:      providerName,
			}, nil
		}

		// Otherwise, show errors and continue to retry
		log.Warn("Structure validation failed, retrying...")
	}

	// Should not reach here, but return last result
	return &RetryResult{
		Output:            lastOutput,
		ValidationErrors:  lastValidationErrors,
		Attempts:          maxAttempts,
		RetriedWithErrors: retriedWithErrors,
		ProviderName:      getProviderName(cfg.Executor),
	}, nil
}

// validateRetryConfig checks that configuration is valid
func validateRetryConfig(cfg *RetryConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.Executor == nil {
		return fmt.Errorf("executor is required")
	}

	// Validator is optional (can generate without validation)
	// PromptBuilder is optional (has default)
	// AntiCorruption is optional
	// ValidationContext is optional

	if cfg.Debug && cfg.DebugOutputDir == "" {
		return fmt.Errorf("debug mode requires DebugOutputDir to be set")
	}

	return nil
}

// DefaultRetryPromptBuilder provides a sensible default for retry prompts
type DefaultRetryPromptBuilder struct{}

// BuildRetryPrompt creates a retry prompt with error feedback
func (b *DefaultRetryPromptBuilder) BuildRetryPrompt(originalPrompt string, invalidOutput string, errors []ValidationError) string {
	errorCount := len(errors)
	criticalCount := countCriticalErrors(errors)

	return fmt.Sprintf(`%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  VALIDATION ERRORS FROM PREVIOUS ATTEMPT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The previous generation had %d validation error(s) (%d critical):

%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📝 INSTRUCTIONS FOR CORRECTION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Please regenerate the output, carefully correcting ALL of the above issues.

IMPORTANT:
- Address each validation error specifically
- Follow the exact format requirements shown in the contract
- Return ONLY the corrected output
- Do NOT include explanations, apologies, or meta-commentary
- Do NOT wrap output in markdown code fences

Generate the corrected output now:
`,
		originalPrompt,
		errorCount,
		criticalCount,
		FormatValidationErrors(errors),
	)
}

// countCriticalErrors is an internal alias to CountCriticalErrors (now in format.go)
func countCriticalErrors(errors []ValidationError) int {
	return CountCriticalErrors(errors)
}

// getProviderName extracts the provider name from an executor if it implements AIExecutorWithProviderInfo.
// Returns empty string if the executor doesn't support provider info retrieval.
func getProviderName(executor AIExecutor) string {
	if withInfo, ok := executor.(AIExecutorWithProviderInfo); ok {
		return withInfo.GetProviderName()
	}
	return ""
}
