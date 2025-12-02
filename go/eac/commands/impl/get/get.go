// Command: get
// Description: Retrieve repository data in structured format
package get

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

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
	case "changed", "config", "dependencies", "environments", "execution", "files", "modules", "suite", "tests":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("Error: unknown subcommand: %s\n", args[0])
		printGetUsage()
		return 1
	}
}

func printGetUsage() {
	log.Info("Retrieve repository data in structured format")
	log.Info("")
	log.Info("Usage: r2r get <subcommand> [args...]")
	log.Info("")
	log.Info("Configuration:")
	log.Info("  config                    Get all EAC configuration (6 configs with defaults)")
	log.Info("")
	log.Info("Repository Structure:")
	log.Info("  modules                   Get all module contracts")
	log.Info("  dependencies              Get module dependency graph")
	log.Info("  execution order           Get build order for modules based on dependencies")
	log.Info("")
	log.Info("Files and Changes:")
	log.Info("  files                     Get repository files with module mappings")
	log.Info("  changed modules           Get modules affected by changed files")
	log.Info("  changed modules ci        Get modules requiring rebuild since last successful CI")
	log.Info("")
	log.Info("Testing:")
	log.Info("  tests                     Get all tests in structured format")
	log.Info("  suite <name>              Get test suite information")
	log.Info("")
	log.Info("Environment:")
	log.Info("  environments              Get all environment contracts")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Get all modules as JSON")
	log.Info("  r2r get modules")
	log.Info("")
	log.Info("  # Get dependency graph")
	log.Info("  r2r get dependencies")
	log.Info("")
	log.Info("  # Get affected modules")
	log.Info("  r2r get changed modules")
	log.Info("")
	log.Info("  # Get modules requiring rebuild in CI (includes cache invalidation)")
	log.Info("  r2r get changed modules ci --as-json")
	log.Info("")
	log.Info("Use 'r2r get <subcommand> --help' for more information about a command.")
}
