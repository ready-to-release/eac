// Command: get
package get

import (
	"fmt"
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
	case "artifacts", "build-deps", "build-times", "changed-modules", "changed-modules-ci", "changed-modules-local", "ci-dispatch", "config", "dependencies", "documented-commands", "environments", "execution-order", "files", "modules", "release-bundle", "suite", "test-results", "tests", "test-timings", "valid-commands":
		// Handled by separate registrations in respective files
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand: %s\n", args[0])
		printGetUsage()
		return 1
	}
}

func printGetUsage() {
	fmt.Println("Retrieve repository data in structured format")
	fmt.Println("")
	fmt.Println("Usage: r2r get <subcommand> [args...]")
	fmt.Println("")
	fmt.Println("Configuration:")
	fmt.Println("  config                    Get all EAC configuration (6 configs with defaults)")
	fmt.Println("")
	fmt.Println("Repository Structure:")
	fmt.Println("  modules                   Get all module contracts")
	fmt.Println("  dependencies              Get module dependency graph")
	fmt.Println("  execution-order           Get build order for modules based on dependencies")
	fmt.Println("")
	fmt.Println("Files and Changes:")
	fmt.Println("  files                     Get repository files with module mappings")
	fmt.Println("  changed-modules           Get modules affected by changed files")
	fmt.Println("  changed-modules-ci        Get modules requiring rebuild since last successful CI")
	fmt.Println("  changed-modules-local     Get modules requiring rebuild based on local build state")
	fmt.Println("")
	fmt.Println("CI/CD:")
	fmt.Println("  ci-dispatch               Filter modules for CI dispatch (skip those with valid CI)")
	fmt.Println("")
	fmt.Println("Build:")
	fmt.Println("  build-times               Get build timing data from build logs")
	fmt.Println("  build-deps                Get build dependencies for a module")
	fmt.Println("  artifacts                 Get resolved artifacts for a module")
	fmt.Println("")
	fmt.Println("Testing:")
	fmt.Println("  test-results              Get test execution results from test manifests")
	fmt.Println("  tests                     Get all tests in structured format")
	fmt.Println("  suite <name>              Get test suite information")
	fmt.Println("  test-timings              Get test timing data from test logs")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  valid-commands            Get all valid commands in structured format")
	fmt.Println("  documented-commands       Get EAC commands documented in markdown files")
	fmt.Println("")
	fmt.Println("Release:")
	fmt.Println("  release-bundle            Get release bundle configuration with module details")
	fmt.Println("")
	fmt.Println("Environment:")
	fmt.Println("  environments              Get all environment contracts")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  # Get all modules as JSON")
	fmt.Println("  r2r get modules")
	fmt.Println("")
	fmt.Println("  # Get dependency graph")
	fmt.Println("  r2r get dependencies")
	fmt.Println("")
	fmt.Println("  # Get affected modules")
	fmt.Println("  r2r get changed-modules")
	fmt.Println("")
	fmt.Println("  # Get modules requiring rebuild in CI (includes cache invalidation)")
	fmt.Println("  r2r get changed-modules-ci --as-json")
	fmt.Println("")
	fmt.Println("Use 'r2r get <subcommand> --help' for more information about a command.")
}
