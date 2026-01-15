//go:build L1 && ov
// +build L1,ov

package generation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/ai/config"
	configpkg "github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockExecutor is a simple mock implementation of AIExecutor for testing
type mockExecutor struct {
	executeErr error
}

func (m *mockExecutor) Execute(ctx interface{}, prompt string, opts ...interface{}) (string, error) {
	if m.executeErr != nil {
		return "", m.executeErr
	}
	return "mock output", nil
}

// mockValidator is a simple mock implementation of Validator for testing
type mockValidator struct {
	errors []validation.ValidationError
}

func (m *mockValidator) Validate(output string, context map[string]interface{}) []validation.ValidationError {
	return m.errors
}

func (m *mockValidator) VerifyImplementation() []validation.ValidationError {
	return nil
}

// createTempRepoDir creates a temporary directory with minimal repository structure
func createTempRepoDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "retry-config-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	return tmpDir
}

// TestBuildRetryConfig_RequiredParameters tests that required parameters are validated
func TestBuildRetryConfig_RequiredParameters(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	tests := []struct {
		name         string
		typeName     string
		outputFormat StructuredFormat
		executor     validation.AIExecutor
		validator    validation.Validator
		templateRoot string
		wantErr      string
	}{
		{
			name:         "nil executor",
			typeName:     "test-type",
			outputFormat: FormatJSON,
			executor:     nil,
			validator:    validator,
			templateRoot: tmpDir,
			wantErr:      "executor is required",
		},
		{
			name:         "empty typeName",
			typeName:     "",
			outputFormat: FormatJSON,
			executor:     executor,
			validator:    validator,
			templateRoot: tmpDir,
			wantErr:      "typeName is required",
		},
		{
			name:         "empty outputFormat",
			typeName:     "test-type",
			outputFormat: "",
			executor:     executor,
			validator:    validator,
			templateRoot: tmpDir,
			wantErr:      "outputFormat is required",
		},
		{
			name:         "empty templateRoot",
			typeName:     "test-type",
			outputFormat: FormatJSON,
			executor:     executor,
			validator:    validator,
			templateRoot: "",
			wantErr:      "templateRoot is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := BuildRetryConfig(
				tt.typeName,
				tt.outputFormat,
				tt.executor,
				tt.validator,
				tt.templateRoot,
				nil, // aiConfig
			)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Nil(t, cfg)
		})
	}
}

// TestBuildRetryConfig_DefaultBehavior tests default behavior when no AI config is provided
func TestBuildRetryConfig_DefaultBehavior(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	cfg, err := BuildRetryConfig(
		"test-type",
		FormatJSON,
		executor,
		validator,
		tmpDir,
		nil, // no AI config
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify defaults
	assert.Equal(t, "test-type", cfg.TypeName)
	assert.Equal(t, FormatJSON, cfg.OutputFormat)
	assert.Equal(t, executor, cfg.Executor)
	assert.Equal(t, validator, cfg.Validator)
	assert.Equal(t, tmpDir, cfg.TemplateRoot)
	assert.Equal(t, defaultMaxAttempts, cfg.MaxAttempts, "should use default maxAttempts=2")
	assert.False(t, cfg.Debug, "debug should default to false")
	assert.NotNil(t, cfg.Logger, "logger should not be nil")
	assert.NotNil(t, cfg.Strategy, "strategy should not be nil")
	assert.Equal(t, config.StrategyStandard, cfg.Strategy.Name(), "should use StandardStrategy by default")
}

// TestBuildRetryConfig_WithNilValidator tests that nil validator is acceptable
func TestBuildRetryConfig_WithNilValidator(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}

	cfg, err := BuildRetryConfig(
		"test-type",
		FormatPlainText,
		executor,
		nil, // nil validator is OK - will be auto-loaded
		tmpDir,
		nil,
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Nil(t, cfg.Validator, "validator should remain nil (auto-loaded later)")
}

// TestBuildRetryConfig_AIConfigMaxAttempts tests loading maxAttempts from AI config
func TestBuildRetryConfig_AIConfigMaxAttempts(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	tests := []struct {
		name            string
		aiConfig        *config.AIConfig
		expectedAttempts int
		description     string
	}{
		{
			name: "AI config with maxAttempts=5",
			aiConfig: &config.AIConfig{
				Types: map[string]config.AITypeConfig{
					"test-type": {
						RetryStrategy: &config.RetryStrategyConfig{
							MaxAttempts: 5,
						},
					},
				},
			},
			expectedAttempts: 5,
			description:      "should use maxAttempts from AI config",
		},
		{
			name: "AI config with maxAttempts=0 (invalid)",
			aiConfig: &config.AIConfig{
				Types: map[string]config.AITypeConfig{
					"test-type": {
						RetryStrategy: &config.RetryStrategyConfig{
							MaxAttempts: 0,
						},
					},
				},
			},
			expectedAttempts: defaultMaxAttempts,
			description:      "should ignore invalid maxAttempts and use default",
		},
		{
			name: "AI config without retry strategy",
			aiConfig: &config.AIConfig{
				Types: map[string]config.AITypeConfig{
					"test-type": {
						RetryStrategy: nil,
					},
				},
			},
			expectedAttempts: defaultMaxAttempts,
			description:      "should use default when no retry strategy",
		},
		{
			name: "AI config without type entry",
			aiConfig: &config.AIConfig{
				Types: map[string]config.AITypeConfig{
					"other-type": {
						RetryStrategy: &config.RetryStrategyConfig{
							MaxAttempts: 5,
						},
					},
				},
			},
			expectedAttempts: defaultMaxAttempts,
			description:      "should use default when type not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := BuildRetryConfig(
				"test-type",
				FormatJSON,
				executor,
				validator,
				tmpDir,
				tt.aiConfig,
			)

			require.NoError(t, err, tt.description)
			require.NotNil(t, cfg)
			assert.Equal(t, tt.expectedAttempts, cfg.MaxAttempts, tt.description)
		})
	}
}

// TestBuildRetryConfig_AIConfigStrategy tests loading retry strategy from AI config
func TestBuildRetryConfig_AIConfigStrategy(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	tests := []struct {
		name            string
		strategyType    string
		focusCategories []string
		expectedName    string
		description     string
	}{
		{
			name:         "standard strategy",
			strategyType: config.StrategyStandard,
			expectedName: config.StrategyStandard,
			description:  "should create StandardStrategy",
		},
		{
			name:            "focused strategy",
			strategyType:    config.StrategyFocused,
			focusCategories: []string{"syntax", "validation"},
			expectedName:    config.StrategyFocused,
			description:     "should create FocusedStrategy",
		},
		{
			name:         "escalating strategy",
			strategyType: config.StrategyEscalating,
			expectedName: config.StrategyEscalating,
			description:  "should create EscalatingStrategy",
		},
		{
			name:            "escalating-focused strategy",
			strategyType:    config.StrategyEscalatingFocused,
			focusCategories: []string{"schema"},
			expectedName:    config.StrategyEscalating, // EscalatingStrategy wraps FocusedStrategy
			description:     "should create EscalatingStrategy with FocusedStrategy base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aiConfig := &config.AIConfig{
				Types: map[string]config.AITypeConfig{
					"test-type": {
						RetryStrategy: &config.RetryStrategyConfig{
							Type:            tt.strategyType,
							FocusCategories: tt.focusCategories,
							MaxAttempts:     3,
						},
					},
				},
			}

			cfg, err := BuildRetryConfig(
				"test-type",
				FormatJSON,
				executor,
				validator,
				tmpDir,
				aiConfig,
			)

			require.NoError(t, err, tt.description)
			require.NotNil(t, cfg)
			require.NotNil(t, cfg.Strategy)
			assert.Equal(t, tt.expectedName, cfg.Strategy.Name(), tt.description)
		})
	}
}

// TestBuildRetryConfig_AIConfigInvalidStrategy tests handling of invalid strategy type
func TestBuildRetryConfig_AIConfigInvalidStrategy(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	aiConfig := &config.AIConfig{
		Types: map[string]config.AITypeConfig{
			"test-type": {
				RetryStrategy: &config.RetryStrategyConfig{
					Type:        "invalid-strategy",
					MaxAttempts: 3,
				},
			},
		},
	}

	cfg, err := BuildRetryConfig(
		"test-type",
		FormatJSON,
		executor,
		validator,
		tmpDir,
		aiConfig,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create retry strategy")
	assert.Contains(t, err.Error(), "unknown retry strategy")
	assert.Nil(t, cfg)
}

// TestBuildRetryConfig_WithDebugOption tests the WithDebug option
func TestBuildRetryConfig_WithDebugOption(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	tests := []struct {
		name          string
		debug         bool
		expectedDebug bool
	}{
		{
			name:          "debug enabled",
			debug:         true,
			expectedDebug: true,
		},
		{
			name:          "debug disabled",
			debug:         false,
			expectedDebug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := BuildRetryConfig(
				"test-type",
				FormatJSON,
				executor,
				validator,
				tmpDir,
				nil,
				WithDebug(tt.debug),
			)

			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, tt.expectedDebug, cfg.Debug)
		})
	}
}

// TestBuildRetryConfig_WithLoggerOption tests the WithLogger option
func TestBuildRetryConfig_WithLoggerOption(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	t.Run("custom logger", func(t *testing.T) {
		customLogger := zap.NewExample()
		defer customLogger.Sync()

		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
			WithLogger(customLogger),
		)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, customLogger, cfg.Logger)
	})

	t.Run("nil logger should not panic", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
			WithLogger(nil),
		)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.NotNil(t, cfg.Logger, "should fall back to default no-op logger")
	})

	t.Run("no logger option uses default", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
		)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.NotNil(t, cfg.Logger, "should have default no-op logger")
	})
}

// TestBuildRetryConfig_WithTagsConfigOption tests the WithTagsConfig option
func TestBuildRetryConfig_WithTagsConfigOption(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	t.Run("with tags config", func(t *testing.T) {
		tagsConfig := &configpkg.TestingTagsConfig{
			// Add relevant fields from actual TestingTagsConfig
		}

		cfg, err := BuildRetryConfig(
			"test-type",
			FormatGherkin,
			executor,
			validator,
			tmpDir,
			nil,
			WithTagsConfig(tagsConfig),
		)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, tagsConfig, cfg.TagsConfig)
	})

	t.Run("nil tags config should not panic", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatGherkin,
			executor,
			validator,
			tmpDir,
			nil,
			WithTagsConfig(nil),
		)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Nil(t, cfg.TagsConfig)
	})

	t.Run("no tags config option", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
		)

		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Nil(t, cfg.TagsConfig)
	})
}

// TestBuildRetryConfig_WithDefaultMaxAttemptsOption tests the WithDefaultMaxAttempts option
func TestBuildRetryConfig_WithDefaultMaxAttemptsOption(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	tests := []struct {
		name             string
		maxAttempts      int
		expectedAttempts int
		description      string
	}{
		{
			name:             "valid max attempts (3)",
			maxAttempts:      3,
			expectedAttempts: 3,
			description:      "should use custom default maxAttempts",
		},
		{
			name:             "valid max attempts (10)",
			maxAttempts:      10,
			expectedAttempts: 10,
			description:      "should accept max value of 10",
		},
		{
			name:             "invalid max attempts (0)",
			maxAttempts:      0,
			expectedAttempts: defaultMaxAttempts,
			description:      "should ignore 0 and use default",
		},
		{
			name:             "invalid max attempts (-1)",
			maxAttempts:      -1,
			expectedAttempts: defaultMaxAttempts,
			description:      "should ignore negative and use default",
		},
		{
			name:             "invalid max attempts (11)",
			maxAttempts:      11,
			expectedAttempts: defaultMaxAttempts,
			description:      "should ignore >10 and use default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := BuildRetryConfig(
				"test-type",
				FormatJSON,
				executor,
				validator,
				tmpDir,
				nil, // no AI config
				WithDefaultMaxAttempts(tt.maxAttempts),
			)

			require.NoError(t, err, tt.description)
			require.NotNil(t, cfg)
			assert.Equal(t, tt.expectedAttempts, cfg.MaxAttempts, tt.description)
		})
	}
}

// TestBuildRetryConfig_AIConfigOverridesDefaultMaxAttempts tests that AI config takes precedence
func TestBuildRetryConfig_AIConfigOverridesDefaultMaxAttempts(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	aiConfig := &config.AIConfig{
		Types: map[string]config.AITypeConfig{
			"test-type": {
				RetryStrategy: &config.RetryStrategyConfig{
					MaxAttempts: 7,
				},
			},
		},
	}

	cfg, err := BuildRetryConfig(
		"test-type",
		FormatJSON,
		executor,
		validator,
		tmpDir,
		aiConfig,
		WithDefaultMaxAttempts(3), // This should be overridden by AI config
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 7, cfg.MaxAttempts, "AI config should override WithDefaultMaxAttempts")
}

// TestBuildRetryConfig_AllOptions tests using all options together
func TestBuildRetryConfig_AllOptions(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}
	customLogger := zap.NewExample()
	defer customLogger.Sync()
	tagsConfig := &configpkg.TestingTagsConfig{}

	aiConfig := &config.AIConfig{
		Types: map[string]config.AITypeConfig{
			"test-type": {
				RetryStrategy: &config.RetryStrategyConfig{
					Type:        config.StrategyFocused,
					MaxAttempts: 4,
				},
			},
		},
	}

	cfg, err := BuildRetryConfig(
		"test-type",
		FormatGherkin,
		executor,
		validator,
		tmpDir,
		aiConfig,
		WithDebug(true),
		WithLogger(customLogger),
		WithTagsConfig(tagsConfig),
		WithDefaultMaxAttempts(3), // Should be overridden by AI config
	)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify all fields
	assert.Equal(t, "test-type", cfg.TypeName)
	assert.Equal(t, FormatGherkin, cfg.OutputFormat)
	assert.Equal(t, executor, cfg.Executor)
	assert.Equal(t, validator, cfg.Validator)
	assert.Equal(t, tmpDir, cfg.TemplateRoot)
	assert.Equal(t, 4, cfg.MaxAttempts, "AI config should override default")
	assert.True(t, cfg.Debug)
	assert.Equal(t, customLogger, cfg.Logger)
	assert.Equal(t, tagsConfig, cfg.TagsConfig)
	assert.NotNil(t, cfg.Strategy)
	assert.Equal(t, config.StrategyFocused, cfg.Strategy.Name())
}

// TestBuildRetryConfig_OptionOrder tests that option order doesn't matter
func TestBuildRetryConfig_OptionOrder(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}
	customLogger := zap.NewExample()
	defer customLogger.Sync()
	tagsConfig := &configpkg.TestingTagsConfig{}

	// Build with options in different orders
	cfg1, err1 := BuildRetryConfig(
		"test-type",
		FormatJSON,
		executor,
		validator,
		tmpDir,
		nil,
		WithDebug(true),
		WithLogger(customLogger),
		WithTagsConfig(tagsConfig),
		WithDefaultMaxAttempts(5),
	)

	cfg2, err2 := BuildRetryConfig(
		"test-type",
		FormatJSON,
		executor,
		validator,
		tmpDir,
		nil,
		WithDefaultMaxAttempts(5),
		WithTagsConfig(tagsConfig),
		WithDebug(true),
		WithLogger(customLogger),
	)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, cfg1)
	require.NotNil(t, cfg2)

	// Both configs should be equivalent
	assert.Equal(t, cfg1.TypeName, cfg2.TypeName)
	assert.Equal(t, cfg1.OutputFormat, cfg2.OutputFormat)
	assert.Equal(t, cfg1.MaxAttempts, cfg2.MaxAttempts)
	assert.Equal(t, cfg1.Debug, cfg2.Debug)
	assert.Equal(t, cfg1.Logger, cfg2.Logger)
	assert.Equal(t, cfg1.TagsConfig, cfg2.TagsConfig)
}

// TestBuildRetryConfig_DifferentFormats tests factory works with all supported formats
func TestBuildRetryConfig_DifferentFormats(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	// Create necessary schema files for JSON format test
	schemaDir := filepath.Join(tmpDir, "contracts", "schemas", "ai")
	require.NoError(t, os.MkdirAll(schemaDir, 0755))

	formats := []struct {
		format StructuredFormat
		name   string
	}{
		{FormatJSON, "json"},
		{FormatGherkin, "gherkin"},
		{FormatOSCALCatalog, "oscal-catalog"},
		{FormatOSCALProfile, "oscal-profile"},
		{FormatStructurizr, "structurizr"},
		{FormatPlainText, "plaintext"},
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			cfg, err := BuildRetryConfig(
				"test-type",
				fmt.format,
				executor,
				validator,
				tmpDir,
				nil,
			)

			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, fmt.format, cfg.OutputFormat)
		})
	}
}

// TestBuildRetryConfig_RealWorldUseCases tests real-world command scenarios
func TestBuildRetryConfig_RealWorldUseCases(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	t.Run("specs command", func(t *testing.T) {
		tagsConfig := &configpkg.TestingTagsConfig{}
		customLogger := zap.NewExample()
		defer customLogger.Sync()

		cfg, err := BuildRetryConfig(
			"specs",
			FormatGherkin,
			executor,
			validator,
			tmpDir,
			nil,
			WithDebug(true),
			WithLogger(customLogger),
			WithTagsConfig(tagsConfig),
		)

		require.NoError(t, err)
		assert.Equal(t, "specs", cfg.TypeName)
		assert.Equal(t, FormatGherkin, cfg.OutputFormat)
		assert.Equal(t, defaultMaxAttempts, cfg.MaxAttempts)
		assert.True(t, cfg.Debug)
		assert.NotNil(t, cfg.TagsConfig)
	})

	t.Run("design command (create)", func(t *testing.T) {
		customLogger := zap.NewExample()
		defer customLogger.Sync()

		cfg, err := BuildRetryConfig(
			"design",
			FormatStructurizr,
			executor,
			validator,
			tmpDir,
			nil,
			WithDebug(false),
			WithLogger(customLogger),
			WithDefaultMaxAttempts(3),
		)

		require.NoError(t, err)
		assert.Equal(t, "design", cfg.TypeName)
		assert.Equal(t, FormatStructurizr, cfg.OutputFormat)
		assert.Equal(t, 3, cfg.MaxAttempts, "design commands use 3 attempts")
		assert.False(t, cfg.Debug)
	})

	t.Run("design command (update)", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"design",
			FormatStructurizr,
			executor,
			validator,
			tmpDir,
			nil,
			WithDebug(true),
			WithDefaultMaxAttempts(3),
			// Note: no logger option (should use default)
		)

		require.NoError(t, err)
		assert.Equal(t, "design", cfg.TypeName)
		assert.Equal(t, FormatStructurizr, cfg.OutputFormat)
		assert.Equal(t, 3, cfg.MaxAttempts)
		assert.True(t, cfg.Debug)
		assert.NotNil(t, cfg.Logger, "should have default logger")
	})

	t.Run("commit-message command", func(t *testing.T) {
		customLogger := zap.NewExample()
		defer customLogger.Sync()

		cfg, err := BuildRetryConfig(
			"commit-message",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
			WithDebug(false),
			WithLogger(customLogger),
		)

		require.NoError(t, err)
		assert.Equal(t, "commit-message", cfg.TypeName)
		assert.Equal(t, FormatJSON, cfg.OutputFormat)
		assert.Equal(t, defaultMaxAttempts, cfg.MaxAttempts)
	})
}

// TestBuildRetryConfig_EdgeCases tests edge cases and boundary conditions
func TestBuildRetryConfig_EdgeCases(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	t.Run("empty AI config", func(t *testing.T) {
		emptyAIConfig := &config.AIConfig{
			Types: map[string]config.AITypeConfig{},
		}

		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			emptyAIConfig,
		)

		require.NoError(t, err)
		assert.Equal(t, defaultMaxAttempts, cfg.MaxAttempts)
		assert.Equal(t, config.StrategyStandard, cfg.Strategy.Name())
	})

	t.Run("AI config with nil Types map", func(t *testing.T) {
		aiConfig := &config.AIConfig{
			Types: nil,
		}

		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			aiConfig,
		)

		require.NoError(t, err)
		assert.Equal(t, defaultMaxAttempts, cfg.MaxAttempts)
	})

	t.Run("multiple WithDebug options", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
			WithDebug(true),
			WithDebug(false), // Last one wins
		)

		require.NoError(t, err)
		assert.False(t, cfg.Debug, "last option should win")
	})

	t.Run("multiple WithDefaultMaxAttempts options", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
			WithDefaultMaxAttempts(3),
			WithDefaultMaxAttempts(5), // Last one wins
		)

		require.NoError(t, err)
		assert.Equal(t, 5, cfg.MaxAttempts, "last option should win")
	})

	t.Run("validator nil (auto-load scenario)", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatPlainText,
			executor,
			nil, // Will be auto-loaded by GenerateWithRetry
			tmpDir,
			nil,
		)

		require.NoError(t, err)
		assert.Nil(t, cfg.Validator, "validator can be nil for auto-loading")
	})
}

// TestBuildRetryConfig_ValidationFailures tests validation catches errors before GenerateWithRetry
func TestBuildRetryConfig_ValidationFailures(t *testing.T) {
	tmpDir := createTempRepoDir(t)
	executor := &mockExecutor{}
	validator := &mockValidator{}

	// The factory should create valid RetryConfig that will pass validate() check
	t.Run("valid config passes validation", func(t *testing.T) {
		cfg, err := BuildRetryConfig(
			"test-type",
			FormatJSON,
			executor,
			validator,
			tmpDir,
			nil,
		)

		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Call validate() directly to ensure it passes
		validateErr := cfg.validate()
		assert.NoError(t, validateErr, "factory should create valid config")
	})
}
