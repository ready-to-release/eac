// Command: get
// Description: Retrieve repository data in structured format
// HasSideEffects: false
package get

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

func init() {
	registry.Register(Get)
}

// Get command entry point
func Get() int {
	args := os.Args[2:] // Skip program name and "get"

	if len(args) == 0 {
		printGetUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printGetUsage()
		return 0
	case "changed", "dependencies", "environments", "execution", "files", "modules", "suite", "tests":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n\n", args[0])
		printGetUsage()
		return 1
	}
}

func printGetUsage() {
	fmt.Println("Retrieve repository data in structured format")
	fmt.Println()
	fmt.Println("Usage: r2r get <subcommand> [args...]")
	fmt.Println()
	fmt.Println("Repository Structure:")
	fmt.Println("  modules                   Get all module contracts")
	fmt.Println("  dependencies              Get module dependency graph")
	fmt.Println("  execution order           Get build order for modules based on dependencies")
	fmt.Println()
	fmt.Println("Files and Changes:")
	fmt.Println("  files                     Get repository files with module mappings")
	fmt.Println("  changed modules           Get modules affected by changed files")
	fmt.Println("  changed modules ci        Get modules requiring rebuild since last successful CI")
	fmt.Println()
	fmt.Println("Testing:")
	fmt.Println("  tests                     Get all tests in structured format")
	fmt.Println("  suite <name>              Get test suite information")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  environments              Get all environment contracts")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Get all modules as JSON")
	fmt.Println("  r2r get modules")
	fmt.Println()
	fmt.Println("  # Get dependency graph")
	fmt.Println("  r2r get dependencies")
	fmt.Println()
	fmt.Println("  # Get affected modules")
	fmt.Println("  r2r get changed modules")
	fmt.Println()
	fmt.Println("  # Get modules requiring rebuild in CI (includes cache invalidation)")
	fmt.Println("  r2r get changed modules ci --as-json")
	fmt.Println()
	fmt.Println("Use 'r2r get <subcommand> --help' for more information about a command.")
}
