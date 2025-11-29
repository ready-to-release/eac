// Command: specs
// Description: Specification management using AI and Gherkin
// HasSideEffects: false
package specs

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Specs)
}

// Specs command entry point
func Specs() int {
	args := os.Args[2:] // Skip program name and "specs"

	if len(args) == 0 {
		printSpecsUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printSpecsUsage()
		return 0
	case "create", "validate", "unused-steps":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printSpecsUsage()
		return 1
	}
}

func printSpecsUsage() {
	fmt.Println("Specification management using AI and Gherkin")
	fmt.Println()
	fmt.Println("Usage: r2r specs <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  create                    Create specifications using AI from natural language")
	fmt.Println("  validate                  Validate Gherkin specifications against contracts")
	fmt.Println("  unused-steps              Find step definitions not used by any feature file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Create specification from description")
	fmt.Println("  r2r specs create \"As a user I want to login with email and password\"")
	fmt.Println()
	fmt.Println("  # Create with specific module")
	fmt.Println("  r2r specs create \"Add authentication\" --module=src-auth")
	fmt.Println()
	fmt.Println("  # Validate specifications")
	fmt.Println("  r2r specs validate specs/src-commands/work/create/specification.feature")
	fmt.Println()
	fmt.Println("Use 'r2r specs <subcommand> --help' for more information about a command.")
}
