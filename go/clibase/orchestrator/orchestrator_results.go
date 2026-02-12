package orchestrator

import (
	"fmt"
	"sort"
	"time"

	"github.com/ready-to-release/eac/go/clibase/output"
)

// PrintSummary prints a summary of all results to the orchestrator output.
func (o *Orchestrator) PrintSummary(results []WorkResult) {
	totalFailed := 0
	totalWarnings := 0
	var totalDuration time.Duration
	failedModules := []string{}

	for _, result := range results {
		totalDuration += result.Duration
		if result.ExitCode != 0 {
			totalFailed++
			failedModules = append(failedModules, result.Moniker)
		}
		if len(result.Warnings) > 0 {
			totalWarnings += len(result.Warnings)
		}
	}

	nl := LineEnding
	fmt.Fprintf(o.orchestratorOut, "%s%s%s", nl, output.SectionHeader(capitalize(o.config.ActionVerb())+" Summary"), nl)
	fmt.Fprintf(o.orchestratorOut, "%s%s", output.SummaryCount("Modules", len(results), len(results)-totalFailed, totalFailed), nl)

	if totalWarnings > 0 {
		fmt.Fprintf(o.orchestratorOut, "Warnings: %d%s", totalWarnings, nl)
	}

	// Show failed modules with errors
	if len(failedModules) > 0 {
		fmt.Fprintf(o.orchestratorOut, "%s❌ Failed:%s", nl, nl)
		for _, result := range results {
			if result.ExitCode != 0 {
				fmt.Fprintf(o.orchestratorOut, "  %s%s", result.Moniker, nl)
				if len(result.Errors) > 0 {
					for _, errMsg := range result.Errors {
						if len(errMsg) > 120 {
							errMsg = errMsg[:117] + "..."
						}
						fmt.Fprintf(o.orchestratorOut, "    %s%s", errMsg, nl)
					}
				}
			}
		}
	}

	// Timing summary (only shown with --timings flag)
	if o.config.ShowTimings {
		fmt.Fprintf(o.orchestratorOut, "%s%s%s", nl, output.SectionHeader("Timing Summary"), nl)

		// Sort results by duration (longest first)
		sortedResults := make([]WorkResult, len(results))
		copy(sortedResults, results)
		sort.Slice(sortedResults, func(i, j int) bool {
			return sortedResults[i].Duration > sortedResults[j].Duration
		})

		for _, result := range sortedResults {
			fmt.Fprintf(o.orchestratorOut, "%s%s", output.TimingLine(result.Duration, result.Moniker), nl)
		}
		fmt.Fprintf(o.orchestratorOut, "%s%s", output.TimingTotal(totalDuration), nl)
	}

	// Output location
	fmt.Fprintf(o.orchestratorOut, "%sOutput: %s%s", nl, o.config.OutputBaseDir, nl)
}

// GetExitCode returns the appropriate exit code based on results.
// Returns 1 if any module failed (ExitCode > 0), 0 otherwise.
// ExitCode -1 means skipped/cached and is treated as success.
func GetExitCode(results []WorkResult) int {
	for _, result := range results {
		if result.ExitCode > 0 {
			return 1
		}
	}
	return 0
}

// GetLastUnitResults returns the component-level results from the last
// RunUnitsParallel call.
// Returns nil if no component execution has occurred.
func (o *Orchestrator) GetLastUnitResults() []UnitResult {
	return o.lastUnitResults
}
