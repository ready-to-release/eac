// Command: show
// Description: Display repository information in human-readable format
package show

import (
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
	log.Info("Repository Structure:")
	log.Info("  modules                   Show all module contracts")
	log.Info("  moduletypes               Show module types grouped by count")
	log.Info("  dependencies              Show module dependencies")
	log.Info("")
	log.Info("Files and Changes:")
	log.Info("  files                     Show repository files with module ownership")
	log.Info("  files changed             Show modified files with module ownership")
	log.Info("  files staged              Show staged files with module ownership")
	log.Info("")
	log.Info("Testing:")
	log.Info("  tests                     Show all tests in the repository")
	log.Info("  suite <name>              Show detailed test suite information")
	log.Info("")
	log.Info("Environment:")
	log.Info("  environments              Show all environment contracts")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Show all modules")
	log.Info("  r2r show modules")
	log.Info("")
	log.Info("  # Show changed files")
	log.Info("  r2r show files changed")
	log.Info("")
	log.Info("  # Show test suite details")
	log.Info("  r2r show suite integration")
	log.Info("")
	log.Info("Use 'r2r show <subcommand> --help' for more information about a command.")
}
