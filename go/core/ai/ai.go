// Package ai provides consolidated access to AI generation functionality
// This package re-exports from the organized subpackages for convenience
package ai

import (
	"context"

	ai "github.com/ready-to-release/eac/contracts/ai-provider/0.1.0"
	"github.com/ready-to-release/eac/go/core/ai/config"
	"github.com/ready-to-release/eac/go/core/ai/generation"
	"github.com/ready-to-release/eac/go/core/ai/mock"
	"github.com/ready-to-release/eac/go/core/ai/templates"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/validation"
)

// ============================================================================
// Config types (from ai/config)
// ============================================================================

type (
	AIConfig            = config.AIConfig
	AIDefaults          = config.AIDefaults
	AITypeConfig        = config.AITypeConfig
	RetryStrategyConfig = config.RetryStrategyConfig
	AIConfigLoader      = config.AIConfigLoader
	ContractLoader      = config.ContractLoader
)

// Retry strategy constants (from ai-config.schema.json).
const (
	StrategyStandard          = config.StrategyStandard
	StrategyFocused           = config.StrategyFocused
	StrategyEscalating        = config.StrategyEscalating
	StrategyEscalatingFocused = config.StrategyEscalatingFocused
)

func NewAIConfigLoader(workspaceRoot string) *AIConfigLoader {
	return config.NewAIConfigLoader(workspaceRoot)
}

func NewContractLoader(workspaceRoot, contractPath, version string) *ContractLoader {
	return config.NewContractLoader(workspaceRoot, contractPath, version)
}

func LoadAIConfig(workspaceRoot string) (*AIConfig, error) {
	return config.LoadAIConfig(workspaceRoot)
}

func ExtractStringList(data map[string]interface{}, key string) []string {
	return config.ExtractStringList(data, key)
}

func ExtractString(data map[string]interface{}, key string) string {
	return config.ExtractString(data, key)
}

func ExtractInt(data map[string]interface{}, key string) int {
	return config.ExtractInt(data, key)
}

func ExtractBool(data map[string]interface{}, key string) bool {
	return config.ExtractBool(data, key)
}

func ExtractMap(data map[string]interface{}, key string) map[string]interface{} {
	return config.ExtractMap(data, key)
}

// ============================================================================
// Templates (from ai/templates)
// ============================================================================

type PromptData = templates.PromptData

// BuildPromptWithTemplate builds an AI prompt using Go text/template.
// See templates.BuildPromptWithTemplate for full documentation.
func BuildPromptWithTemplate(
	promptTemplate string,
	contract *domain.Contract,
	customData map[string]string,
) (string, error) {
	return templates.BuildPromptWithTemplate(promptTemplate, contract, customData)
}

// ============================================================================
// Generation types (from ai/generation)
// ============================================================================

type StructuredFormat = generation.StructuredFormat

const (
	FormatJSON         = generation.FormatJSON
	FormatGherkin      = generation.FormatGherkin
	FormatOSCALCatalog = generation.FormatOSCALCatalog
	FormatOSCALProfile = generation.FormatOSCALProfile
	FormatStructurizr  = generation.FormatStructurizr
	FormatPlainText    = generation.FormatPlainText
)

type (
	StructuredGenerator = generation.StructuredGenerator
	GenerationResult    = generation.GenerationResult
	RetryConfig         = generation.RetryConfig
	RetryResult         = generation.RetryResult
	RetryStrategy       = generation.RetryStrategy
	StandardStrategy    = generation.StandardStrategy
	FocusedStrategy     = generation.FocusedStrategy
	EscalatingStrategy  = generation.EscalatingStrategy
)

const (
	TypeCommitMessage = generation.TypeCommitMessage
	TypeSquashMessage = generation.TypeSquashMessage
	TypeDesign        = generation.TypeDesign
	TypeSpecs         = generation.TypeSpecs
	TypeRiskProfile   = generation.TypeRiskProfile
	TypeRiskAssess    = generation.TypeRiskAssess
	TypeUpdateDesign  = generation.TypeUpdateDesign
)

const (
	JSONPhaseInstruction        = generation.JSONPhaseInstruction
	GherkinPhaseInstruction     = generation.GherkinPhaseInstruction
	OSCALPhaseInstruction       = generation.OSCALPhaseInstruction
	StructurizrPhaseInstruction = generation.StructurizrPhaseInstruction
	PlainTextPhaseInstruction   = generation.PlainTextPhaseInstruction
)

func GenerateWithRetry(
	ctx context.Context,
	config *RetryConfig,
	fullPrompt string,
) (*RetryResult, error) {
	return generation.GenerateWithRetry(ctx, config, fullPrompt)
}

func GetRetryStrategy(strategyName string, focusCategories []string) (RetryStrategy, error) {
	return generation.GetRetryStrategy(strategyName, focusCategories)
}

func GetPhaseInstruction(format StructuredFormat) (string, error) {
	return generation.GetPhaseInstruction(format)
}

// BuildRetryConfig creates a RetryConfig from command-level inputs using the factory pattern.
// See generation.BuildRetryConfig for detailed documentation.
func BuildRetryConfig(
	typeName string,
	outputFormat StructuredFormat,
	executor ai.GenerationExecutor,
	validator validation.Validator,
	templateRoot string,
	aiConfig *AIConfig,
	options ...RetryConfigOption,
) (*RetryConfig, error) {
	return generation.BuildRetryConfig(typeName, outputFormat, executor, validator, templateRoot, aiConfig, options...)
}

// RetryConfigOption configures optional RetryConfig fields using functional options pattern.
type RetryConfigOption = generation.RetryConfigOption

// WithDebug enables debug mode for intermediate output logging.
func WithDebug(debug bool) RetryConfigOption {
	return generation.WithDebug(debug)
}

// WithLogger sets a custom logger (defaults to zap.NewNop() if not provided).
func WithLogger(logger interface{}) RetryConfigOption {
	// Accept interface{} to allow passing *zap.Logger without importing zap in this file
	return generation.WithLogger(logger)
}

// WithTagsConfig sets the testing tags config (required for Gherkin validation).
func WithTagsConfig(tags interface{}) RetryConfigOption {
	// Accept interface{} to avoid circular imports
	return generation.WithTagsConfig(tags)
}

// WithDefaultMaxAttempts sets the default max attempts (used if not in aiConfig).
func WithDefaultMaxAttempts(max int) RetryConfigOption {
	return generation.WithDefaultMaxAttempts(max)
}

// ============================================================================
// Mock types (from ai/mock)
// ============================================================================

const (
	EnvMockAIDir    = mock.EnvMockAIDir
	EnvMockAIPrefix = mock.EnvMockAIPrefix
)

func GetMockResponse(command string) (string, bool) {
	return mock.GetMockResponse(command)
}

func GetMockResponseWithSubcommand(command, subcommand string) (string, bool) {
	return mock.GetMockResponseWithSubcommand(command, subcommand)
}

func IsMockEnabled() bool {
	return mock.IsMockEnabled()
}

type (
	MockAIExecutor = mock.AIExecutor
	MockValidator  = mock.Validator
)

// ============================================================================
// Validation types (from validation package - for convenience)
// ============================================================================

type (
	GenerationExecutor             = ai.GenerationExecutor
	GenerationExecutorWithProvider = ai.GenerationExecutorWithProviderInfo
	AIExecutor                     = validation.AIExecutor // Deprecated: use GenerationExecutor
	Validator                      = validation.Validator
	ValidationError                = validation.ValidationError
)
