// Command: show
// Description: Display repository information in human-readable format
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
)

func init() {
	registry.Register(Show)
}

// Show command entry point
func Show() int {
	args := os.Args[2:] // Skip program name and "show"

	if len(args) == 0 {
		printShowUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printShowUsage()
		return 0
	case "config", "dependencies", "environments", "files", "modules", "moduletypes", "suite", "tests":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printShowUsage()
		return 1
	}
}

func printShowUsage() {
	fmt.Println("Display repository information in human-readable format")
	fmt.Println()
	fmt.Println("Usage: r2r show <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  config                    Show all EAC configuration summary")
	fmt.Println()
	fmt.Println("Repository Structure:")
	fmt.Println("  modules                   Show all module contracts")
	fmt.Println("  moduletypes               Show module types grouped by count")
	fmt.Println("  dependencies              Show module dependencies")
	fmt.Println()
	fmt.Println("Files and Changes:")
	fmt.Println("  files                     Show repository files with module ownership")
	fmt.Println("  files changed             Show modified files with module ownership")
	fmt.Println("  files staged              Show staged files with module ownership")
	fmt.Println()
	fmt.Println("Testing:")
	fmt.Println("  tests                     Show all tests in the repository")
	fmt.Println("  suite <name>              Show detailed test suite information")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  environments              Show all environment contracts")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Show all modules")
	fmt.Println("  r2r show modules")
	fmt.Println()
	fmt.Println("  # Show changed files")
	fmt.Println("  r2r show files changed")
	fmt.Println()
	fmt.Println("  # Show test suite details")
	fmt.Println("  r2r show suite integration")
	fmt.Println()
	fmt.Println("Use 'r2r show <subcommand> --help' for more information about a command.")
}
