// Command: validate
// Description: Validate repository contracts and dependencies
// HasSideEffects: false
package validate

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register(Validate)
}

// Validate command entry point
func Validate() int {
	args := os.Args[2:] // Skip program name and "validate"

	if len(args) == 0 {
		printValidateUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printValidateUsage()
		return 0
	case "contracts":
		// Handled by separate registrations in respective files
		return 0
	case "dependencies":
		// Handled by separate registrations in respective files
		return 0
	case "test-tags":
		// Handled by separate registrations in respective files
		return 0
	case "module-hierarchy":
		// Handled by separate registrations in respective files
		return 0
	case "module-files":
		// Handled by separate registrations in respective files
		return 0
	case "markdown":
		// Handled by separate registrations in respective files
		return 0
	case "go-tidy":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printValidateUsage()
		return 1
	}
}

func printValidateUsage() {
	fmt.Println("Validate repository contracts and dependencies")
	fmt.Println()
	fmt.Println("Usage: r2r validate <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  contracts                 Validate repository contracts against JSON schemas")
	fmt.Println("  dependencies              Validate module dependency contracts")
	fmt.Println("  test-tags                 Validate that all test tags are defined in the tag contract")
	fmt.Println("  module-hierarchy          Validate module dependency graph structure")
	fmt.Println("  module-files              Validate module file ownership")
	fmt.Println("  markdown                  Validate markdown file syntax")
	fmt.Println("  go-tidy                   Validate Go module dependencies are tidy")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Validate all contracts against schemas")
	fmt.Println("  r2r validate contracts")
	fmt.Println()
	fmt.Println("  # Validate all dependencies")
	fmt.Println("  r2r validate dependencies")
	fmt.Println()
	fmt.Println("  # Validate test tags")
	fmt.Println("  r2r validate test-tags")
	fmt.Println()
	fmt.Println("  # Validate module hierarchy")
	fmt.Println("  r2r validate module-hierarchy")
	fmt.Println()
	fmt.Println("  # Validate module file ownership")
	fmt.Println("  r2r validate module-files")
	fmt.Println()
	fmt.Println("  # Validate markdown files")
	fmt.Println("  r2r validate markdown")
	fmt.Println()
	fmt.Println("  # Validate Go module tidiness")
	fmt.Println("  r2r validate go-tidy")
	fmt.Println()
	fmt.Println("Use 'r2r validate <subcommand> --help' for more information about a command.")
}
