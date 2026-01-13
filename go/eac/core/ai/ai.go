// Package ai provides consolidated access to AI generation functionality
// This package re-exports from the organized subpackages for convenience
package ai

import (
	"context"

	"github.com/ready-to-release/eac/go/eac/core/ai/config"
	"github.com/ready-to-release/eac/go/eac/core/ai/generation"
	"github.com/ready-to-release/eac/go/eac/core/ai/mock"
	"github.com/ready-to-release/eac/go/eac/core/ai/templates"
	"github.com/ready-to-release/eac/go/eac/core/validation"
)

// ============================================================================
// Config types (from ai/config)
// ============================================================================

type AIConfig = config.AIConfig
type AIDefaults = config.AIDefaults
type AITypeConfig = config.AITypeConfig
type RetryStrategyConfig = config.RetryStrategyConfig
type AIConfigLoader = config.AIConfigLoader
type ContractLoader = config.ContractLoader

// Retry strategy constants (from ai-config.schema.json)
const (
	StrategyStandard          = config.StrategyStandard
	StrategyFocused           = config.StrategyFocused
	StrategyEscalating        = config.StrategyEscalating
	StrategyEscalatingFocused = config.StrategyEscalatingFocused
)

func NewAIConfigLoader(workspaceRoot string) *AIConfigLoader {
	return config.NewAIConfigLoader(workspaceRoot)
}

func NewContractLoader(workspaceRoot string, contractPath string, version string) *ContractLoader {
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

var BuildPromptWithTemplate = templates.BuildPromptWithTemplate

// ============================================================================
// Generation types (from ai/generation)
// ============================================================================

type StructuredFormat = generation.StructuredFormat

const (
	FormatJSON          = generation.FormatJSON
	FormatGherkin       = generation.FormatGherkin
	FormatOSCALCatalog  = generation.FormatOSCALCatalog
	FormatOSCALProfile  = generation.FormatOSCALProfile
	FormatStructurizr   = generation.FormatStructurizr
	FormatPlainText     = generation.FormatPlainText
)

type StructuredGenerator = generation.StructuredGenerator
type GenerationResult = generation.GenerationResult
type RetryConfig = generation.RetryConfig
type RetryResult = generation.RetryResult
type RetryStrategy = generation.RetryStrategy
type StandardStrategy = generation.StandardStrategy
type FocusedStrategy = generation.FocusedStrategy
type EscalatingStrategy = generation.EscalatingStrategy

const (
	TypeCommitMessage = generation.TypeCommitMessage
	TypeSquashMessage = generation.TypeSquashMessage
	TypeDesign        = generation.TypeDesign
	TypeSpecs         = generation.TypeSpecs
	TypeRiskProfile   = generation.TypeRiskProfile
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

type MockAIExecutor = mock.MockAIExecutor
type MockValidator = mock.MockValidator

// ============================================================================
// Validation types (from validation package - for convenience)
// ============================================================================

type AIExecutor = validation.AIExecutor
type Validator = validation.Validator
type ValidationError = validation.ValidationError
