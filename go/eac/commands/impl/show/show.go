// Command: show
// Description: Display repository information in human-readable format
package show

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
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
	case "books", "build-summary", "build-times", "config", "dependencies", "environments", "files", "files-changed", "files-staged", "modules", "moduletypes", "suite", "test-summary", "tests", "test-timings", "valid-commands", "workspaces":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("unknown subcommand: %s\n", args[0])
		printShowUsage()
		return 1
	}
}

func printShowUsage() {
	log.Info("Display repository information in human-readable format")
	log.Info("")
	log.Info("Usage: r2r show <subcommand> [args...]")
	log.Info("")
	log.Info("Configuration:")
	log.Info("  config                    Show all EAC configuration summary")
	log.Info("")
	log.Info("Documentation:")
	log.Info("  books                     Show all configured books")
	log.Info("")
	log.Info("Repository Structure:")
	log.Info("  modules                   Show all module contracts")
	log.Info("  moduletypes               Show module types grouped by count")
	log.Info("  dependencies              Show module dependencies")
	log.Info("")
	log.Info("Files and Changes:")
	log.Info("  files                     Show repository files with module ownership")
	log.Info("  files-changed             Show modified files with module ownership")
	log.Info("  files-staged              Show staged files with module ownership")
	log.Info("")
	log.Info("Build:")
	log.Info("  build-summary <module>    Generate pretty build summary for GitHub Actions")
	log.Info("  build-times               Show build timing analysis from recent builds")
	log.Info("")
	log.Info("Testing:")
	log.Info("  test-summary <module> <suite>  Generate pretty test summary for GitHub Actions")
	log.Info("  tests                     Show all tests in the repository")
	log.Info("  suite <name>              Show detailed test suite information")
	log.Info("  test-timings              Show test timing analysis from recent test runs")
	log.Info("")
	log.Info("Environment:")
	log.Info("  environments              Show all environment contracts")
	log.Info("")
	log.Info("Commands:")
	log.Info("  valid-commands            Show all valid commands in a table")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Show all modules")
	log.Info("  r2r show modules")
	log.Info("")
	log.Info("  # Show changed files")
	log.Info("  r2r show files-changed")
	log.Info("")
	log.Info("  # Show test suite details")
	log.Info("  r2r show suite integration")
	log.Info("")
	log.Info("Use 'r2r show <subcommand> --help' for more information about a command.")
}
