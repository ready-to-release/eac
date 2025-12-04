// Command: validate
// Description: Validate repository contracts and dependencies
package validate

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(Validate)
}

// Validate command entry point
func Validate() int {
	args := os.Args[2:] // Skip program name and "validate"

	if len(args) == 0 {
		printValidateUsage()
		return 1
	}

	// Check for help flag
	switch args[0] {
	case "--help", "-h":
		printValidateUsage()
		return 0
	case "books":
		// Handled by separate registrations in respective files
		return 0
	case "contracts":
		// Handled by separate registrations in respective files
		return 0
	case "dependencies":
		// Handled by separate registrations in respective files
		return 0
	case "test-tags":
		// Handled by separate registrations in respective files
		return 0
	case "module-hierarchy":
		// Handled by separate registrations in respective files
		return 0
	case "module-files":
		// Handled by separate registrations in respective files
		return 0
	case "markdown":
		// Handled by separate registrations in respective files
		return 0
	case "go-tidy":
		// Handled by separate registrations in respective files
		return 0
	case "risk":
		// Handled by separate registrations in respective files
		return 0
	case "release-version":
		// Handled by separate registrations in respective files
		return 0
	default:
		log.Errorf("Error: unknown subcommand: %s\n", args[0])
		printValidateUsage()
		return 1
	}
}

func printValidateUsage() {
	log.Info("Validate repository contracts and dependencies")
	log.Info("")
	log.Info("Usage: r2r validate <subcommand> [args...]")
	log.Info("")
	log.Info("Subcommands:")
	log.Info("  books                     Validate books.yml configuration")
	log.Info("  contracts                 Validate repository contracts against JSON schemas")
	log.Info("  dependencies              Validate module dependency contracts")
	log.Info("  go-tidy                   Validate Go module dependencies are tidy")
	log.Info("  markdown                  Validate markdown file syntax")
	log.Info("  module-files              Validate module file ownership")
	log.Info("  module-hierarchy          Validate module dependency graph structure")
	log.Info("  release-version           Validate release version format (semver)")
	log.Info("  risk                      Validate OSCAL profiles and assessment-results")
	log.Info("  test-tags                 Validate that all test tags are defined in the tag contract")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate all contracts against schemas")
	log.Info("  r2r validate contracts")
	log.Info("")
	log.Info("  # Validate all dependencies")
	log.Info("  r2r validate dependencies")
	log.Info("")
	log.Info("  # Validate test tags")
	log.Info("  r2r validate test-tags")
	log.Info("")
	log.Info("  # Validate module hierarchy")
	log.Info("  r2r validate module-hierarchy")
	log.Info("")
	log.Info("  # Validate module file ownership")
	log.Info("  r2r validate module-files")
	log.Info("")
	log.Info("  # Validate markdown files")
	log.Info("  r2r validate markdown")
	log.Info("")
	log.Info("  # Validate Go module tidiness")
	log.Info("  r2r validate go-tidy")
	log.Info("")
	log.Info("  # Validate OSCAL risk documents")
	log.Info("  r2r validate risk specs/risk-controls/billing.profile.json")
	log.Info("")
	log.Info("  # Validate release version format")
	log.Info("  r2r validate release-version 1.2.3")
	log.Info("")
	log.Info("Use 'r2r validate <subcommand> --help' for more information about a command.")
}
