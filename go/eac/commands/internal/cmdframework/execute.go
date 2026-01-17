package cmdframework

import (
	"fmt"
	"io"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
)

// phaseExecute handles the execution phase:
// - Set up worker function
// - Run orchestrator (layered or parallel)
// - Collect results.
func phaseExecute(ctx *ExecutionContext, worker WorkerFunc) error {
	if worker == nil {
		return fmt.Errorf("worker function is required")
	}

	// Early return if nothing to execute
	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// Wrap the user's worker to match orchestrator signature
	orchWorker := func(moniker string, logWriter io.Writer) int {
		return worker(ctx, moniker, logWriter)
	}

	// Set worker on orchestrator
	ctx.Orchestrator.SetWorker(orchWorker)

	// Execute based on mode
	var results []orchestrator.WorkResult
	var err error

	if ctx.Config.Layered {
		// Layered execution (build): respect dependency order
		layers := ctx.GetLayers()
		log.Debugf("Executing %d layers with %d total modules",
			len(layers), len(ctx.GetExecutionMonikers()))
		results, err = ctx.Orchestrator.RunLayered(layers)
	} else {
		// Parallel execution (test/scan): all at once
		monikers := ctx.GetExecutionMonikers()
		log.Debugf("Executing %d modules in parallel", len(monikers))
		results, err = ctx.Orchestrator.Run(monikers)
	}

	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	ctx.Results = results
	return nil
}

// GetExitCode returns the overall exit code from results (0 if all succeeded).
func (ctx *ExecutionContext) GetExitCode() int {
	return orchestrator.GetExitCode(ctx.Results)
}

// GetSuccessCount returns the number of successful results.
func (ctx *ExecutionContext) GetSuccessCount() int {
	count := 0
	for _, r := range ctx.Results {
		if r.ExitCode == 0 {
			count++
		}
	}
	return count
}

// GetFailureCount returns the number of failed results.
func (ctx *ExecutionContext) GetFailureCount() int {
	count := 0
	for _, r := range ctx.Results {
		if r.ExitCode != 0 {
			count++
		}
	}
	return count
}

// GetResultByMoniker finds a result by moniker.
func (ctx *ExecutionContext) GetResultByMoniker(moniker string) *orchestrator.WorkResult {
	for i := range ctx.Results {
		if ctx.Results[i].Moniker == moniker {
			return &ctx.Results[i]
		}
	}
	return nil
}

// GetFailedMonikers returns the monikers of failed modules.
func (ctx *ExecutionContext) GetFailedMonikers() []string {
	var failed []string
	for _, r := range ctx.Results {
		if r.ExitCode != 0 {
			failed = append(failed, r.Moniker)
		}
	}
	return failed
}
