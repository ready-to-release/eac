// Command: build
// Description: Build modules in the repository
// HasSideEffects: true
package build

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Build)
}

// Build command entry point
func Build() int {
	args := os.Args[2:] // Skip program name and "build"

	if len(args) == 0 {
		printBuildUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printBuildUsage()
		return 0
	case "module", "modules":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printBuildUsage()
		return 1
	}
}

func printBuildUsage() {
	fmt.Println("Build modules in the repository")
	fmt.Println()
	fmt.Println("Usage: r2r build <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  module <name>             Build a specific module")
	fmt.Println("  modules                   Build multiple modules")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Build specific module")
	fmt.Println("  r2r build module src-commands")
	fmt.Println()
	fmt.Println("  # Build all modules")
	fmt.Println("  r2r build modules")
	fmt.Println()
	fmt.Println("Use 'r2r build <subcommand> --help' for more information about a command.")
}
