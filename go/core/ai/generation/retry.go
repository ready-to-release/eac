package generation

import (
	"context"
	"fmt"

	configpkg "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/validation"
	"go.uber.org/zap"
)

const defaultMaxAttempts = 2

// RetryConfig configures retry behavior with output format specification
// Commands handle deterministic formatting of the generated output
//
// IMPORTANT: This is a CORE library and must NOT use component loggers (logging.C()).
// All logging must go through the Logger field, which defaults to zap.NewNop() if nil.
// This ensures the generation layer never writes to command-specific log files.
type RetryConfig struct {
	// TypeName is the AI type (e.g., "commit-message", "specs", "design")
	// Used to load schemas and validators
	TypeName string

	// Generation configuration (MANDATORY)
	OutputFormat StructuredFormat     // Output format (json, gherkin, oscal, structurizr, plaintext)
	Validator    validation.Validator // Validator (optional - will auto-load if not provided)

	// Executor performs AI generation
	Executor validation.AIExecutor

	// TemplateRoot is the repository root (for loading schemas)
	TemplateRoot string

	// MaxAttempts is the maximum number of generation attempts (default: 2)
	MaxAttempts int

	// Debug enables saving intermediate outputs for troubleshooting
	Debug bool

	// Strategy determines retry behavior (optional, defaults to StandardStrategy)
	Strategy RetryStrategy

	// Logger for structured logging (optional, defaults to zap.NewNop())
	// NEVER pass component logger (logging.C()) - use nil to disable logging
	Logger *zap.Logger

	// TagsConfig is the testing tags configuration for Gherkin validation (optional)
	// Required for Gherkin format validation, ignored for other formats
	TagsConfig *configpkg.TestingTagsConfig
}

// RetryResult holds the result of generation with retry.
type RetryResult struct {
	Output            string                       // Final formatted output
	ValidationErrors  []validation.ValidationError // Validation errors (empty if valid)
	Attempts          int                          // Total attempts made
	RetriedWithErrors bool                         // Whether a retry was triggered
	ProviderName      string                       // AI provider name (e.g., "claude-cli")
	GeneratedContent  string                       // Generated content (format depends on OutputFormat)
}

// GenerateWithRetry generates AI output using format-aware generation
//
// Flow:
//  1. Generate structured content in OutputFormat (with retry on validation failure)
//  2. Commands handle formatting output to final format (no AI)
//
// Returns RetryResult with output and metadata. Returns error only if AI execution fails.
func GenerateWithRetry(ctx context.Context, cfg *RetryConfig, prompt string) (*RetryResult, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid retry config: %w", err)
	}

	// Load validator if not provided
	validator := cfg.Validator
	if validator == nil {
		var err error
		validator, err = loadValidatorForFormat(cfg.TemplateRoot, cfg.TypeName, cfg.OutputFormat, cfg.TagsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load validator: %w", err)
		}
	}

	generator := cfg.buildGenerator(validator)
	logger := cfg.getLogger()

	logger.Info("Starting generation",
		zap.String("type", cfg.TypeName),
		zap.String("outputFormat", string(cfg.OutputFormat)),
		zap.String("strategy", generator.Strategy.Name()),
		zap.Int("maxAttempts", generator.MaxAttempts))

	genResult, genErr := generator.Generate(ctx, prompt)
	result := cfg.buildResult(genResult)

	if genErr != nil {
		return result, fmt.Errorf("generation failed: %w", genErr)
	}

	cfg.logResult(logger, result, genResult)
	return result, nil
}

// validate checks that all required config fields are set.
func (cfg *RetryConfig) validate() error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.Executor == nil {
		return fmt.Errorf("executor is required")
	}
	if cfg.OutputFormat == "" {
		return fmt.Errorf("OutputFormat is required - must be explicitly specified (json, gherkin, oscal, structurizr, plaintext)")
	}
	if cfg.TypeName == "" {
		return fmt.Errorf("type name is required")
	}
	if cfg.TemplateRoot == "" {
		return fmt.Errorf("template root is required")
	}
	return nil
}

// getMaxAttempts returns configured max attempts or default.
func (cfg *RetryConfig) getMaxAttempts() int {
	if cfg.MaxAttempts > 0 {
		return cfg.MaxAttempts
	}
	return defaultMaxAttempts
}

// getStrategy returns configured strategy or default StandardStrategy.
func (cfg *RetryConfig) getStrategy() RetryStrategy {
	if cfg.Strategy != nil {
		return cfg.Strategy
	}
	return &StandardStrategy{}
}

// getLogger returns configured logger or no-op logger
// NEVER returns component logger (logging.C()) to prevent writing to command log files.
func (cfg *RetryConfig) getLogger() *zap.Logger {
	if cfg.Logger != nil {
		return cfg.Logger
	}
	// Return no-op logger to discard all logs
	return zap.NewNop()
}

// getProviderName extracts provider name from executor if available.
func (cfg *RetryConfig) getProviderName() string {
	if provider, ok := cfg.Executor.(validation.AIExecutorWithProviderInfo); ok {
		return provider.GetProviderName()
	}
	return "unknown"
}

// buildGenerator creates a StructuredGenerator from config.
func (cfg *RetryConfig) buildGenerator(validator validation.Validator) *StructuredGenerator {
	return &StructuredGenerator{
		OutputFormat: cfg.OutputFormat,
		Executor:     cfg.Executor,
		Validator:    validator,
		MaxAttempts:  cfg.getMaxAttempts(),
		Strategy:     cfg.getStrategy(),
		Debug:        cfg.Debug,
		Logger:       cfg.getLogger(),
	}
}

// buildResult converts GenerationResult to RetryResult.
func (cfg *RetryConfig) buildResult(gr *GenerationResult) *RetryResult {
	return &RetryResult{
		Output:            gr.FormattedContent,
		ValidationErrors:  gr.Errors,
		Attempts:          gr.Attempts,
		RetriedWithErrors: gr.Attempts > 1,
		ProviderName:      cfg.getProviderName(),
		GeneratedContent:  gr.GeneratedContent,
	}
}

// logResult logs generation outcome.
func (cfg *RetryConfig) logResult(logger *zap.Logger, result *RetryResult, gr *GenerationResult) {
	if len(gr.Errors) == 0 {
		logger.Info("Generation successful",
			zap.Int("attempts", result.Attempts))
	} else {
		logger.Warn("Generation completed with validation errors",
			zap.Int("attempts", result.Attempts),
			zap.Int("errorCount", len(gr.Errors)))
	}
}
