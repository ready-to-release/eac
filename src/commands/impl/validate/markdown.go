// Command: validate markdown
// Description: Validate markdown file syntax
// HasSideEffects: false
package validate

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"github.com/ready-to-release/eac/src/core/markdown"
	"github.com/ready-to-release/eac/src/core/repository"
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
		fmt.Fprintf(os.Stderr, "Error: failed to get repository root: %v\n", err)
		return 1
	}

	// Create validator with default options
	opts := markdown.DefaultValidatorOptions()
	validator := markdown.NewValidator(opts, os.Stdout)

	// Validate all markdown files
	results, err := validator.ValidateDirectory(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to validate directory: %v\n", err)
		return 1
	}

	// Print results
	return validator.PrintResults(results, repoRoot)
}

func printMarkdownUsage() {
	fmt.Println("Validate markdown file syntax")
	fmt.Println()
	fmt.Println("Usage: r2r validate markdown")
	fmt.Println()
	fmt.Println("Checks:")
	fmt.Println("  - Valid markdown syntax")
	fmt.Println("  - Proper heading hierarchy")
	fmt.Println("  - Valid code blocks (JSON, YAML)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Validate all markdown files")
	fmt.Println("  r2r validate markdown")
}
