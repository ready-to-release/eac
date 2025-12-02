// Command: validate markdown
// Description: Validate markdown file syntax
package validate

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/markdown"
	"github.com/ready-to-release/eac/go/eac/core/repository"
)

func init() {
	registry.Register(ValidateMarkdown)
}

// ValidateMarkdown validates all markdown files in the repository
func ValidateMarkdown() int {
	args := os.Args[2:] // Skip program name and "validate"

	// Check if this is being called as a subcommand
	if len(args) > 0 && args[0] == "markdown" {
		args = args[1:] // Skip the subcommand name
	}

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
	log.Info("Usage: r2r validate markdown")
	log.Info("")
	log.Info("Checks:")
	log.Info("  - Valid markdown syntax")
	log.Info("  - Proper heading hierarchy")
	log.Info("  - Valid code blocks (JSON, YAML)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate all markdown files")
	log.Info("  r2r validate markdown")
}
