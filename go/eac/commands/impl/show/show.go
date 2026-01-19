// Command: show
package show

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Show)
}

// Show command entry point.
func Show() int {
	args := os.Args[2:] // Skip program name and "show"

	if len(args) == 0 {
		printShowUsage()
		return 1
	}

	// Check for subcommands
	switch args[0] {
	case "books", "build-summary", "build-times", "config", "dependencies", "environments", "files", "files-changed", "files-staged", "modules", "suite", "test-results", "test-summary", "tests", "test-timings", "valid-commands", "workspaces":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
		printShowUsage()
		return 1
	}
}

func printShowUsage() {
	fmt.Println("Display repository information in human-readable format")
	fmt.Println("")
	fmt.Println("Usage: r2r show <subcommand> [args...]")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  config                    Show all EAC configuration summary")
	fmt.Println("")
	fmt.Println("Documentation:")
	fmt.Println("  books                     Show all configured books")
	fmt.Println("")
	fmt.Println("Repository Structure:")
	fmt.Println("  modules                   Show all module contracts")
	fmt.Println("  dependencies              Show module dependencies")
	fmt.Println("")
	fmt.Println("Files and Changes:")
	fmt.Println("  files                     Show repository files with module ownership")
	fmt.Println("  files-changed             Show modified files with module ownership")
	fmt.Println("  files-staged              Show staged files with module ownership")
	fmt.Println("")
	fmt.Println("Build:")
	fmt.Println("  build-summary <module>    Generate pretty build summary for GitHub Actions")
	fmt.Println("  build-times               Show build timing analysis from recent builds")
	fmt.Println("")
	fmt.Println("Testing:")
	fmt.Println("  test-results              Display test execution results in human-readable format")
	fmt.Println("  test-summary <module> <suite>  Generate pretty test summary for GitHub Actions")
	fmt.Println("  tests                     Show all tests in the repository")
	fmt.Println("  suite <name>              Show detailed test suite information")
	fmt.Println("  test-timings              Show test timing analysis from recent test runs")
	fmt.Println("")
	fmt.Println("Environment:")
	fmt.Println("  environments              Show all environment contracts")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  valid-commands            Show all valid commands in a table")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Show all modules")
	fmt.Println("  r2r show modules")
	fmt.Println("")
	fmt.Println("  # Show changed files")
	fmt.Println("  r2r show files-changed")
	fmt.Println("")
	fmt.Println("  # Show test suite details")
	fmt.Println("  r2r show suite integration")
	fmt.Println("")
	fmt.Println("Use 'r2r show <subcommand> --help' for more information about a command.")
}
