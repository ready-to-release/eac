package ci

import (
	"context"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/logging"
)

type pipelineCICommand struct{}

var _ core.SimpleCommandPort = (*pipelineCICommand)(nil)

func (c *pipelineCICommand) Name() string { return "pipeline ci" }

func (c *pipelineCICommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "pipeline-ci",
		Short:         "CI orchestration and diagnostics",
		Long:          "Commands for CI/CD orchestration including workflow dispatch,\nmonitoring, and diagnostic summary generation.\n\nAvailable commands:\n  dispatch-and-wait   Dispatch a workflow and wait for completion\n  get-run-id          Get CI run ID for a workflow and commit SHA\n  schedule            Schedule and dispatch CI workflows with concurrency limits\n  summary-link        Generate diagnostic markdown for CI summaries",
	}
}

func (c *pipelineCICommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return PipelineCI()
}

var log = logging.C()

func PipelineCI() int {
	if len(os.Args) < 4 {
		printCIUsage()
		return 1
	}

	subcommand := os.Args[3]
	switch subcommand {
	case "dispatch-and-wait", "get-run-id", "schedule", "summary-link":
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
	log.Info("  schedule            Schedule and dispatch CI with concurrency limits")
	log.Info("  summary-link        Generate diagnostic markdown for CI summaries")
	log.Info("")
	log.Info("Use 'pipeline ci <command> --help' for more information.")
}
