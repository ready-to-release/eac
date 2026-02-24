package generation

import (
	"fmt"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/core/ai/config"
	configpkg "github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/validation"
	"go.uber.org/zap"
)

// ============================================================================
// Factory Function for RetryConfig Construction
// ============================================================================

// BuildRetryConfig creates a RetryConfig from command-level inputs.
// This is the centralized factory for all retry configuration across commands.
//
// Required Parameters:
//   - typeName: AI type (e.g., TypeSpecs, TypeDesign, TypeCommitMessage)
//   - outputFormat: Output format (e.g., FormatGherkin, FormatStructurizr, FormatJSON)
//   - executor: AI executor for generation (must not be nil)
//   - validator: Validator for output (can be nil, will auto-load based on format)
//   - templateRoot: Repository root for schema/contract loading (must not be empty)
//   - aiConfig: Optional AI config for strategy (can be nil, uses defaults)
//
// Optional Parameters (via functional options):
//   - WithDebug(bool): Enable debug mode for intermediate output logging
//   - WithLogger(*zap.Logger): Set custom logger (defaults to zap.NewNop())
//   - WithTagsConfig(*configpkg.TestingTagsConfig): Set testing tags (for Gherkin)
//   - WithDefaultMaxAttempts(int): Set default max attempts (used if not in aiConfig)
//
// Example usage:
//
//	retryConfig, err := BuildRetryConfig(
//	    TypeSpecs,
//	    FormatGherkin,
//	    executor,
//	    validator,
//	    "/path/to/repo",
//	    aiConfig,
//	    WithDebug(true),
//	    WithLogger(logger),
//	    WithTagsConfig(tags),
//	)
//
// Returns configured RetryConfig ready for GenerateWithRetry, or error if validation fails.
func BuildRetryConfig(
	typeName string,
	outputFormat StructuredFormat,
	executor ai.GenerationExecutor,
	validator validation.Validator,
	templateRoot string,
	aiConfig *config.AIConfig,
	options ...RetryConfigOption,
) (*RetryConfig, error) {
	// Validate required parameters
	if executor == nil {
		return nil, fmt.Errorf("executor is required")
	}
	if typeName == "" {
		return nil, fmt.Errorf("typeName is required")
	}
	if outputFormat == "" {
		return nil, fmt.Errorf("outputFormat is required")
	}
	if templateRoot == "" {
		return nil, fmt.Errorf("templateRoot is required")
	}

	// Apply defaults
	cfg := &retryConfigBuilder{
		defaultMaxAttempts: 2,
		logger:             zap.NewNop(),
	}

	// Apply options
	for _, opt := range options {
		opt(cfg)
	}

	// Load strategy from aiConfig if available
	maxAttempts := cfg.defaultMaxAttempts
	var strategy RetryStrategy

	if aiConfig != nil {
		if typeConfig, ok := aiConfig.Types[typeName]; ok {
			if typeConfig.RetryStrategy != nil && typeConfig.RetryStrategy.MaxAttempts > 0 {
				maxAttempts = typeConfig.RetryStrategy.MaxAttempts
			}

			if typeConfig.RetryStrategy != nil {
				var focusCategories []string
				if typeConfig.RetryStrategy.FocusCategories != nil {
					focusCategories = typeConfig.RetryStrategy.FocusCategories
				}

				var err error
				strategy, err = GetRetryStrategy(typeConfig.RetryStrategy.Type, focusCategories)
				if err != nil {
					return nil, fmt.Errorf("failed to create retry strategy: %w", err)
				}
			}
		}
	}

	// Default to StandardStrategy if not configured
	if strategy == nil {
		strategy = &StandardStrategy{}
	}

	return &RetryConfig{
		TypeName:     typeName,
		OutputFormat: outputFormat,
		Validator:    validator,
		Executor:     executor,
		TemplateRoot: templateRoot,
		MaxAttempts:  maxAttempts,
		Debug:        cfg.debug,
		Strategy:     strategy,
		Logger:       cfg.logger,
		TagsConfig:   cfg.tagsConfig,
	}, nil
}

// RetryConfigOption configures optional RetryConfig fields using functional options pattern.
type RetryConfigOption func(*retryConfigBuilder)

// retryConfigBuilder holds optional configuration during RetryConfig construction.
type retryConfigBuilder struct {
	defaultMaxAttempts int
	debug              bool
	logger             *zap.Logger
	tagsConfig         *configpkg.TestingTagsConfig
}

// WithDebug enables debug mode for intermediate output logging.
func WithDebug(debug bool) RetryConfigOption {
	return func(cfg *retryConfigBuilder) {
		cfg.debug = debug
	}
}

// WithLogger sets a custom logger (defaults to zap.NewNop() if not provided).
// Pass nil to explicitly disable logging.
func WithLogger(logger interface{}) RetryConfigOption {
	return func(cfg *retryConfigBuilder) {
		if logger != nil {
			if zapLogger, ok := logger.(*zap.Logger); ok {
				cfg.logger = zapLogger
			}
		}
	}
}

// WithTagsConfig sets the testing tags config (required for Gherkin validation).
// This is only needed for commands that generate Gherkin specifications.
func WithTagsConfig(tags interface{}) RetryConfigOption {
	return func(cfg *retryConfigBuilder) {
		if tags != nil {
			if tagsConfig, ok := tags.(*configpkg.TestingTagsConfig); ok {
				cfg.tagsConfig = tagsConfig
			}
		}
	}
}

// WithDefaultMaxAttempts sets the default max attempts (used if not in aiConfig).
// Design commands typically use 3, other commands use 2.
// Value must be between 1 and 10 (inclusive), otherwise ignored.
func WithDefaultMaxAttempts(max int) RetryConfigOption {
	return func(cfg *retryConfigBuilder) {
		if max > 0 && max <= 10 {
			cfg.defaultMaxAttempts = max
		}
	}
}
