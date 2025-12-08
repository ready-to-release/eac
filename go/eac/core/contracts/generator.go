package contracts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Intent: Provide a high-level helper for AI-powered contract generation
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Single function call to generate with validation
//   - Automatic contract loading from standard paths
//   - Clear configuration struct
//
// Easy to change:
//   - Pluggable executors (OpenAI, Anthropic, etc.)
//   - Optional debug mode
//   - Configurable retry attempts
//
// Hard to break:
//   - Automatic anti-corruption filtering
//   - Validation with retry
//   - Detailed error reporting

// GeneratorConfig configures AI generation with contracts
type GeneratorConfig struct {
	// ContractName is the name of the contract (e.g., "commit-message", "specifications")
	ContractName string

	// ContractVersion is the version to use (e.g., "0.1.0")
	ContractVersion string

	// WorkspaceRoot is the root directory of the workspace
	WorkspaceRoot string

	// Executor performs AI generation
	Executor AIExecutor

	// MaxAttempts is the maximum number of generation attempts (default: 2)
	MaxAttempts int

	// Debug enables saving intermediate outputs
	Debug bool

	// DebugOutputDir is where debug files are saved (required if Debug is true)
	DebugOutputDir string

	// PromptBuilder creates retry prompts (optional, uses default if nil)
	PromptBuilder RetryPromptBuilder

	// ValidationContext provides additional context for validation
	ValidationContext map[string]interface{}
}

// GenerateWithContract generates AI output using a contract for validation
//
// This is a convenience function that:
//  1. Loads the AI config and anti-corruption rules
//  2. Creates a validator
//  3. Runs generation with retry on validation errors
//
// Returns:
//   - The generated output (cleaned)
//   - Validation errors (if any)
//   - error if loading or generation fails
func GenerateWithContract(ctx context.Context, cfg *GeneratorConfig, prompt string) (*RetryResult, error) {
	// Validate configuration
	if err := validateGeneratorConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid generator config: %w", err)
	}

	// Load AI config
	loader := NewAIConfigLoader(cfg.WorkspaceRoot)
	antiCorruptionRules, err := loader.GetAntiCorruptionRules(cfg.ContractName)
	if err != nil {
		return nil, fmt.Errorf("failed to load anti-corruption rules: %w", err)
	}

	// Create validator based on contract type
	var validator Validator
	switch cfg.ContractName {
	case "specs":
		validator = NewGherkinValidatorFromConfig(loader, cfg.ContractName, antiCorruptionRules)
	case "commit-message":
		// For commit messages, we might not have a specific validator
		// Just use anti-corruption filtering
		validator = nil
	default:
		// No specific validator, just use anti-corruption filtering
		validator = nil
	}

	// Configure retry
	retryConfig := &RetryConfig{
		Executor:          cfg.Executor,
		Validator:         validator,
		PromptBuilder:     cfg.PromptBuilder,
		AntiCorruption:    antiCorruptionRules,
		ContentMarker:     "", // Will be inferred from contract if needed
		MaxAttempts:       cfg.MaxAttempts,
		Debug:             cfg.Debug,
		DebugOutputDir:    cfg.DebugOutputDir,
		ValidationContext: cfg.ValidationContext,
	}

	// Generate with retry
	result, err := GenerateWithRetry(ctx, retryConfig, prompt)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	return result, nil
}

// validateGeneratorConfig checks that configuration is valid
func validateGeneratorConfig(cfg *GeneratorConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if cfg.ContractName == "" {
		return fmt.Errorf("contract name is required")
	}

	if cfg.ContractVersion == "" {
		return fmt.Errorf("contract version is required")
	}

	if cfg.WorkspaceRoot == "" {
		return fmt.Errorf("workspace root is required")
	}

	if cfg.Executor == nil {
		return fmt.Errorf("executor is required")
	}

	if cfg.Debug && cfg.DebugOutputDir == "" {
		return fmt.Errorf("debug mode requires DebugOutputDir to be set")
	}

	// Create debug directory if needed
	if cfg.Debug {
		if err := os.MkdirAll(cfg.DebugOutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create debug output directory: %w", err)
		}
	}

	return nil
}

// QuickGenerate is a convenience function for simple AI generation with contracts
//
// Example:
//   result, err := QuickGenerate(ctx, "specifications", "0.1.0", workspaceRoot, executor, prompt)
func QuickGenerate(ctx context.Context, contractName, version, workspaceRoot string, executor AIExecutor, prompt string) (*RetryResult, error) {
	cfg := &GeneratorConfig{
		ContractName:    contractName,
		ContractVersion: version,
		WorkspaceRoot:   workspaceRoot,
		Executor:        executor,
		MaxAttempts:     2,
		Debug:           false,
	}

	return GenerateWithContract(ctx, cfg, prompt)
}

// DebugGenerate is a convenience function for generation with debug output
//
// Debug files will be saved to: {workspaceRoot}/out/debug/{contractName}/{timestamp}/
func DebugGenerate(ctx context.Context, contractName, version, workspaceRoot string, executor AIExecutor, prompt string) (*RetryResult, error) {
	// Create debug directory with timestamp
	debugDir := filepath.Join(workspaceRoot, "out", "debug", contractName, fmt.Sprintf("%d", os.Getpid()))

	cfg := &GeneratorConfig{
		ContractName:    contractName,
		ContractVersion: version,
		WorkspaceRoot:   workspaceRoot,
		Executor:        executor,
		MaxAttempts:     2,
		Debug:           true,
		DebugOutputDir:  debugDir,
	}

	return GenerateWithContract(ctx, cfg, prompt)
}
