package generation

import (
	"context"
	"fmt"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/ai/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// RetryStrategy determines retry behavior based on validation results.
type RetryStrategy interface {
	// ShouldRetry decides if retry should happen based on validation result
	ShouldRetry(ctx context.Context, errors []validation.ValidationError, attempt, maxAttempts int) bool

	// BuildRetryPrompt constructs retry prompt based on strategy and errors
	BuildRetryPrompt(ctx context.Context, originalPrompt, lastOutput string, errors []validation.ValidationError, attempt int) string

	// Name returns the strategy name for logging
	Name() string
}

// hasRetriableErrors checks if errors contain any retriable (non-critical, non-warning) errors.
func hasRetriableErrors(errors []validation.ValidationError) bool {
	for _, err := range errors {
		if !err.IsCritical() && !err.IsWarning() {
			return true
		}
	}
	return false
}

// countCriticalErrors returns the number of non-warning errors.
func countCriticalErrors(errors []validation.ValidationError) int {
	count := 0
	for _, err := range errors {
		if !err.IsWarning() {
			count++
		}
	}
	return count
}

// filterErrorsByCategories returns errors matching any of the given categories.
func filterErrorsByCategories(errors []validation.ValidationError, categories []validation.ErrorCategory) []validation.ValidationError {
	if len(categories) == 0 {
		return errors
	}

	categorySet := make(map[validation.ErrorCategory]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}

	var filtered []validation.ValidationError
	for _, err := range errors {
		if categorySet[err.GetCategory()] {
			filtered = append(filtered, err)
		}
	}

	if len(filtered) == 0 {
		return errors // Fallback to all errors
	}
	return filtered
}

// groupErrorsByCategory organizes errors by their category.
func groupErrorsByCategory(errors []validation.ValidationError) map[validation.ErrorCategory][]validation.ValidationError {
	grouped := make(map[validation.ErrorCategory][]validation.ValidationError)
	for _, err := range errors {
		cat := err.GetCategory()
		grouped[cat] = append(grouped[cat], err)
	}
	return grouped
}

// formatCategories converts category list to comma-separated string.
func formatCategories(categories []validation.ErrorCategory) string {
	if len(categories) == 0 {
		return "all"
	}
	parts := make([]string, len(categories))
	for i, cat := range categories {
		parts[i] = string(cat)
	}
	return strings.Join(parts, ", ")
}

// Retry prompt section headers.
const (
	headerStandardCorrection = "INSTRUCTIONS FOR CORRECTION"
	headerFocusedCorrection  = "CORRECTION INSTRUCTIONS"
)

// Retry prompt instructions.
const (
	standardInstructions = `Please regenerate the output, carefully correcting ALL of the above issues.

IMPORTANT:
- Address each validation error specifically
- Follow the exact format requirements shown in the contract
- Return ONLY the corrected output
- Do NOT include explanations, apologies, or meta-commentary
- Do NOT wrap output in markdown code fences`

	focusedInstructions = `Pay special attention to the error categories listed above.
Return ONLY the corrected output without explanations.`
)

// buildRetryPrompt constructs a retry prompt with consistent structure.
func buildRetryPrompt(originalPrompt, header, summary, errorsText, sectionTitle, instructions string) string {
	var sb strings.Builder

	// Original prompt
	sb.WriteString(originalPrompt)
	sb.WriteString("\n\n")

	// Error notification section
	sb.WriteString(RetryPromptSeparator)
	sb.WriteString("\n")
	sb.WriteString(header)
	sb.WriteString("\n")
	sb.WriteString(RetryPromptSeparator)
	sb.WriteString("\n\n")

	// Error summary and details
	sb.WriteString(summary)
	sb.WriteString("\n\n")
	sb.WriteString(errorsText)
	sb.WriteString("\n\n")

	// Correction instructions
	sb.WriteString(RetryPromptSeparator)
	sb.WriteString("\n")
	sb.WriteString(sectionTitle)
	sb.WriteString("\n")
	sb.WriteString(RetryPromptSeparator)
	sb.WriteString("\n\n")
	sb.WriteString(instructions)
	sb.WriteString("\n\nGenerate the corrected output now:\n")

	return sb.String()
}

// StandardStrategy retries on any retriable error.
type StandardStrategy struct{}

func (s *StandardStrategy) ShouldRetry(ctx context.Context, errors []validation.ValidationError, attempt, maxAttempts int) bool {
	return attempt < maxAttempts && hasRetriableErrors(errors)
}

func (s *StandardStrategy) BuildRetryPrompt(ctx context.Context, originalPrompt, lastOutput string, errors []validation.ValidationError, attempt int) string {
	header := RetryPromptWarning + " FROM PREVIOUS ATTEMPT"
	summary := fmt.Sprintf("The previous generation had %d validation error(s) (%d critical):",
		len(errors), countCriticalErrors(errors))

	return buildRetryPrompt(
		originalPrompt,
		header,
		summary,
		contracts.FormatValidationErrors(errors),
		headerStandardCorrection,
		standardInstructions,
	)
}

func (s *StandardStrategy) Name() string {
	return config.StrategyStandard
}

// FocusedStrategy only retries if errors in the specified categories are present.
type FocusedStrategy struct {
	FocusCategories []validation.ErrorCategory
}

func (s *FocusedStrategy) ShouldRetry(ctx context.Context, errors []validation.ValidationError, attempt, maxAttempts int) bool {
	if attempt >= maxAttempts {
		return false
	}

	for _, err := range errors {
		if err.IsCritical() || err.IsWarning() {
			continue
		}
		if s.matchesFocus(err.GetCategory()) {
			return true
		}
	}
	return false
}

func (s *FocusedStrategy) matchesFocus(cat validation.ErrorCategory) bool {
	if len(s.FocusCategories) == 0 {
		return true
	}
	for _, focus := range s.FocusCategories {
		if cat == focus {
			return true
		}
	}
	return false
}

func (s *FocusedStrategy) BuildRetryPrompt(ctx context.Context, originalPrompt, lastOutput string, errors []validation.ValidationError, attempt int) string {
	focusedErrors := filterErrorsByCategories(errors, s.FocusCategories)
	header := RetryPromptFocus + " CORRECTION REQUIRED"
	summary := fmt.Sprintf("Focus on fixing these specific issues (%d error(s) in categories: %s):",
		len(focusedErrors), formatCategories(s.FocusCategories))

	return buildRetryPrompt(
		originalPrompt,
		header,
		summary,
		contracts.FormatValidationErrors(focusedErrors),
		headerFocusedCorrection,
		focusedInstructions,
	)
}

func (s *FocusedStrategy) Name() string {
	return config.StrategyFocused
}

// EscalatingStrategy wraps another strategy and adds urgency on later attempts.
type EscalatingStrategy struct {
	BaseStrategy RetryStrategy
}

func (s *EscalatingStrategy) ShouldRetry(ctx context.Context, errors []validation.ValidationError, attempt, maxAttempts int) bool {
	return s.BaseStrategy.ShouldRetry(ctx, errors, attempt, maxAttempts)
}

func (s *EscalatingStrategy) BuildRetryPrompt(ctx context.Context, originalPrompt, lastOutput string, errors []validation.ValidationError, attempt int) string {
	prompt := s.BaseStrategy.BuildRetryPrompt(ctx, originalPrompt, lastOutput, errors, attempt)

	if attempt >= 2 {
		prompt += s.buildEscalationNotice(attempt)
	}

	if attempt >= 3 {
		prompt += s.buildFinalAttemptSection(errors)
	}

	return prompt
}

func (s *EscalatingStrategy) buildEscalationNotice(attempt int) string {
	return fmt.Sprintf("\n\nRETRY ATTEMPT %d of %d - INCREASED ATTENTION REQUIRED\n"+
		"\nThis is a retry. Pay close attention to the validation rules above.\n",
		attempt, 3)
}

func (s *EscalatingStrategy) buildFinalAttemptSection(errors []validation.ValidationError) string {
	var sb strings.Builder
	sb.WriteString("\n\nCRITICAL: FINAL RETRY ATTEMPT\n")
	sb.WriteString("\nMultiple retries have failed. This is the last attempt.\n")
	sb.WriteString("Carefully review EACH requirement and ensure EXACT compliance.\n")

	grouped := groupErrorsByCategory(errors)
	if len(grouped) > 0 {
		sb.WriteString("\n" + RetryPromptSeparator + "\n")
		sb.WriteString("ERRORS BY CATEGORY:\n")
		sb.WriteString(RetryPromptSeparator + "\n")

		for cat, errs := range grouped {
			sb.WriteString(fmt.Sprintf("\n%s (%d error(s)):\n", strings.ToUpper(string(cat)), len(errs)))
			for _, err := range errs {
				if err.Line > 0 {
					sb.WriteString(fmt.Sprintf("  - Line %d: [%s] %s\n", err.Line, err.GetCode(), err.Message))
				} else {
					sb.WriteString(fmt.Sprintf("  - [%s] %s\n", err.GetCode(), err.Message))
				}
			}
		}
		sb.WriteString("\n" + RetryPromptSeparator + "\n")
	}

	return sb.String()
}

func (s *EscalatingStrategy) Name() string {
	return config.StrategyEscalating
}

// GetRetryStrategy creates a retry strategy from configuration.
func GetRetryStrategy(strategyName string, focusCategories []string) (RetryStrategy, error) {
	categories := toErrorCategories(focusCategories)

	switch strategyName {
	case config.StrategyStandard, "":
		return &StandardStrategy{}, nil

	case config.StrategyFocused:
		return &FocusedStrategy{FocusCategories: categories}, nil

	case config.StrategyEscalating:
		return &EscalatingStrategy{BaseStrategy: &StandardStrategy{}}, nil

	case config.StrategyEscalatingFocused:
		return &EscalatingStrategy{BaseStrategy: &FocusedStrategy{FocusCategories: categories}}, nil

	default:
		return nil, fmt.Errorf("unknown retry strategy: %s (valid: %s, %s, %s, %s)",
			strategyName, config.StrategyStandard, config.StrategyFocused, config.StrategyEscalating, config.StrategyEscalatingFocused)
	}
}

// toErrorCategories converts string slice to ErrorCategory slice.
func toErrorCategories(categories []string) []validation.ErrorCategory {
	result := make([]validation.ErrorCategory, len(categories))
	for i, cat := range categories {
		result[i] = validation.ErrorCategory(cat)
	}
	return result
}
