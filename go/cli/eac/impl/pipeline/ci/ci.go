// Command: pipeline ci
// Short: CI orchestration and diagnostics
// Long: Commands for CI/CD orchestration including workflow dispatch,
// Long: monitoring, and diagnostic summary generation.
// Long:
// Long: Available commands:
// Long:   dispatch-and-wait   Dispatch a workflow and wait for completion
// Long:   get-run-id          Get CI run ID for a workflow and commit SHA
// Long:   summary-link        Generate diagnostic markdown for CI summaries
package ci

import (
	"os"

	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

func PipelineCI() int {
	if len(os.Args) < 4 {
		printCIUsage()
		return 1
	}

	subcommand := os.Args[3]
	switch subcommand {
	case "dispatch-and-wait", "get-run-id", "summary-link":
		// These are handled by their own registrations
		return 0
	case "--help", "-h":
		printCIUsage()
		return 0
	default:
		log.Errorf("Unknown subcommand: %s", subcommand)
		printCIUsage()
		return 1
	}
}

func printCIUsage() {
	log.Info("CI orchestration and diagnostics")
	log.Info("")
	log.Info("Usage: pipeline ci <command> [options]")
	log.Info("")
	log.Info("Commands:")
	log.Info("  dispatch-and-wait   Dispatch a workflow and wait for completion")
	log.Info("  get-run-id          Get CI run ID for a workflow and commit SHA")
	log.Info("  summary-link        Generate diagnostic markdown for CI summaries")
	log.Info("")
	log.Info("Use 'pipeline ci <command> --help' for more information.")
}
