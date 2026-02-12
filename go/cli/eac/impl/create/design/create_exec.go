package design

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	design "github.com/ready-to-release/eac/go/cli/eac/impl/design"
	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	coreai "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/validation/formats/structurizr"
)

// validateModuleExists checks if the source code exists for the specified module.
func validateModuleExists(config *DesignConfig, out *design.Output) error {
	out.Progressf("🔍 Validating source code for module '%s'...", config.Module)

	// Check if source directory exists
	if _, err := os.Stat(config.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source code not found for module '%s'\n\nExpected at: %s\n\nUsage: design create <module>\nExample: design create clie (analyzes code in go/cli/clie/)",
			config.Module, config.SourcePath)
	}

	out.Progressf("✅ Source code found at: %s", config.SourcePath)
	return nil
}

// checkDockerAvailability checks if Docker is running before starting expensive operations.
func checkDockerAvailability(config *DesignConfig, out *design.Output) error {
	out.Progress("🐳 Checking Docker availability...")

	// Create a temporary validator to check Docker
	validator, err := structurizr.NewCompositeValidator("", "", true)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Check if Docker is running
	if !validator.IsDockerRunning() {
		return fmt.Errorf("Docker is not running\n\nStructurizr validation requires Docker to be running.\nPlease start Docker and try again.")
	}

	out.Progress("✅ Docker is available")
	return nil
}

// generateAndValidate generates AI output with retry and validates with Structurizr CLI.
func generateAndValidate(config *DesignConfig, prompt string, out *design.Output) (string, error) {
	// Create executor
	executor := ai.NewExecutor(config.TemplateRoot)
	providers.RegisterBuiltIn(executor)

	// Wrap executor to match contract.AIExecutor interface
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Create composite validator (quick + full validation)
	// Skip expensive Docker validation if quick validation finds errors or if --skip-validation
	validator, err := structurizr.NewCompositeValidator(config.Module, config.TemplateRoot, !config.SkipValidation)
	if err != nil {
		return "", fmt.Errorf("failed to create validator: %w", err)
	}
	defer validator.Cleanup()

	// Load AI config to get retry strategy
	var aiConfig *coreai.AIConfig
	aiConfig, err = coreai.LoadAIConfig(config.TemplateRoot)
	if err != nil {
		log.Warnf("Could not load AI config, using default retry strategy: %v", err)
		aiConfig = nil
	}

	// Build retry configuration using factory
	retryConfig, err := coreai.BuildRetryConfig(
		coreai.TypeDesign,
		coreai.FormatStructurizr,
		executorAdapter,
		validator,
		config.TemplateRoot,
		aiConfig,
		coreai.WithDebug(config.Debug),
		coreai.WithLogger(logging.C().Zap()),
		coreai.WithDefaultMaxAttempts(3),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build retry config: %w", err)
	}

	// Generate with retry and validation
	out.Progress("🤖 Generating architecture design with AI...")

	result, err := coreai.GenerateWithRetry(
		context.Background(),
		retryConfig,
		prompt,
	)
	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return result.Output, nil
}

// writeOutputAndReportSuccess writes the workspace file and reports success.
func writeOutputAndReportSuccess(config *DesignConfig, outputPath, content string, out *design.Output) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Report success
	out.Progress("\n✅ Architecture design created")
	out.Progressf("   File: %s", outputPath)
	if !config.SkipValidation {
		out.Progress("   Valid: ✅ Passed Structurizr validation\n")
	} else {
		out.Progress("   Valid: ⚠️  Validation skipped\n")
	}
	out.Progress("ℹ️  Next steps:")
	out.Progress("   1. Review the generated workspace")
	out.Progress("   2. Refine containers and relationships as needed")
	out.Progressf("   3. View in browser: design serve %s", config.Module)
	out.Progressf("   4. Validate anytime: design validate %s", config.Module)

	return nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
