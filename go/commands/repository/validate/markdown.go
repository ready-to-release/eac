package validate

import (
	"context"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/markdown"
	"github.com/ready-to-release/eac/go/core/repository"
)

type validateMarkdownCommand struct{}

var _ core.SimpleCommandPort = (*validateMarkdownCommand)(nil)

func (c *validateMarkdownCommand) Name() string { return "validate markdown" }

func (c *validateMarkdownCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-markdown",
		Short:         "Validate markdown file syntax",
		Long: "Validates markdown files for proper syntax, heading hierarchy, and code block formatting.",
		Notes: "Expected Output:\n  Displays markdown syntax validation results for all markdown files in repository.\n  Shows errors for invalid syntax, heading issues, and malformed code blocks.\n  Exit code 0 if all valid, 1 if validation errors found.",
	}
}

func (c *validateMarkdownCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateMarkdown()
}

// ValidateMarkdown validates all markdown files in the repository.
func ValidateMarkdown() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "validate", and "markdown"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printMarkdownUsage()
		return 0
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	// Create validator with default options
	opts := markdown.DefaultValidatorOptions()
	validator := markdown.NewValidator(opts, os.Stdout)

	// Validate all markdown files
	results, err := validator.ValidateDirectory(repoRoot)
	if err != nil {
		log.Errorf("Error: failed to validate directory: %v", err)
		return 1
	}

	// Print results
	return validator.PrintResults(results, repoRoot)
}

func printMarkdownUsage() {
	log.Info("Validate markdown file syntax")
	log.Info("")
	log.Info("Usage: eac validate markdown")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Valid markdown syntax")
	log.Info("  - Proper heading hierarchy")
	log.Info("  - Valid code blocks (JSON, YAML)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate all markdown files")
	log.Info("  eac validate markdown")
}
