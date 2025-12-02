// Command: risk
// Short: Risk assessment and management commands
// Long: The risk command provides tools for creating, assessing, and reporting on security risks
// Long: using the OSCAL (Open Security Controls Assessment Language) format.
// Long:
// Long: Available subcommands:
// Long:   assess    - Update OSCAL assessment-results with test and security evidence
// Long:
// Long: Related commands:
// Long:   create risk     - Create OSCAL profile from risk assessment using AI
// Long:   validate risk   - Validate OSCAL profiles and assessment-results
// Long:   show risk-report - Display aggregated risk assessment report
package risk

import (
	"os"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(Risk)
}

// Risk is the entry point for the risk command.
func Risk() int {
	args := os.Args[2:] // Skip program name and "risk"

	if len(args) == 0 {
		printUsage()
		return 1
	}

	// Check for help flag
	if args[0] == "--help" || args[0] == "-h" {
		printUsage()
		return 0
	}

	// Subcommands are handled by separate registrations
	// This function handles the bare "risk" command
	log.Errorf("Error: unknown subcommand: %s", args[0])
	printUsage()
	return 1
}

func printUsage() {
	log.Info("Risk assessment and management commands")
	log.Info("")
	log.Info("Usage: risk <subcommand> [args...]")
	log.Info("")
	log.Info("Subcommands:")
	log.Info("  assess          Update OSCAL assessment-results with test and security evidence")
	log.Info("")
	log.Info("Related commands:")
	log.Info("  create risk     Create OSCAL profile from risk assessment using AI")
	log.Info("  validate risk   Validate OSCAL profiles and assessment-results")
	log.Info("  show risk-report Display aggregated risk assessment report")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Create profile from assessment document")
	log.Info("  create risk .docs/assessments/security-assessment.md --module billing")
	log.Info("")
	log.Info("  # Assess a module")
	log.Info("  risk assess billing-service")
	log.Info("")
	log.Info("  # Show aggregated report")
	log.Info("  show risk-report")
	log.Info("")
	log.Info("  # Validate OSCAL documents")
	log.Info("  validate risk specs/risk-controls/billing.profile.json")
	log.Info("")
	log.Info("Use '<command> --help' for more information about a specific command.")
}
