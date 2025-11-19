// Command: test
// Description: Run tests for modules and test suites
// HasSideEffects: true
package test

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Test)
}

// Test command entry point
func Test() int {
	args := os.Args[2:] // Skip program name and "test"

	if len(args) == 0 {
		printTestUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printTestUsage()
		return 0
	case "module", "modules", "suite", "list-suites":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printTestUsage()
		return 1
	}
}

func printTestUsage() {
	fmt.Println("Run tests for modules and test suites")
	fmt.Println()
	fmt.Println("Usage: r2r test <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  module <name>             Test a specific module")
	fmt.Println("  modules                   Test multiple modules")
	fmt.Println("  suite <name>              Run a specific test suite")
	fmt.Println("  list-suites               List all available test suites")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Test specific module")
	fmt.Println("  r2r test module src-commands")
	fmt.Println()
	fmt.Println("  # Run specific test suite")
	fmt.Println("  r2r test suite integration")
	fmt.Println()
	fmt.Println("  # List all test suites")
	fmt.Println("  r2r test list-suites")
	fmt.Println()
	fmt.Println("Use 'r2r test <subcommand> --help' for more information about a command.")
}
