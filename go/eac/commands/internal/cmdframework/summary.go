package cmdframework

import (
	"fmt"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
)

const (
	maxErrorLength = 500
	maxLineLength  = 100
)

// SummaryGenerator is a function that generates custom summary data.
type SummaryGenerator func(ctx *ExecutionContext) *initsummary.Summary

// phaseSummary handles the summary phase:
// - Generate summary data
// - Display TUI summary or print to console
// - Return exit code.
func phaseSummary(ctx *ExecutionContext, customSummary SummaryGenerator) int {
	totalTime := time.Since(ctx.StartTime)
	exitCode := ctx.GetExitCode()

	if ctx.Config.UseTUI {
		// Generate TUI summary
		summaryData := generateTUISummary(ctx, totalTime)
		ctx.Orchestrator.SendSummary(summaryData)

		// Wait for TUI to finish
		ctx.Orchestrator.WaitTUI()
		ctx.Orchestrator.StopTUI()
	} else {
		// Print summary to console
		printConsoleSummary(ctx, totalTime)
	}

	return exitCode
}

// generateTUISummary creates TUI summary data from execution results.
func generateTUISummary(ctx *ExecutionContext, totalTime time.Duration) *tui.SummaryData {
	results := ctx.Results
	successCount := ctx.GetSuccessCount()
	failureCount := ctx.GetFailureCount()

	// Build run summary line
	runSummary := fmt.Sprintf("%d succeeded", successCount)
	if failureCount > 0 {
		runSummary += fmt.Sprintf(", %d failed", failureCount)
	}

	// Aggregate results by type
	type typeSummary struct {
		Packages int
		Passed   int
		Failed   int
		Warnings int
	}
	summaryByType := make(map[string]*typeSummary)

	type failedResult struct {
		moniker    string
		duration   time.Duration
		moduleType string
		errors     []string
		logPath    string
	}
	var failedResults []failedResult

	for _, result := range results {
		moduleType := ctx.ModuleTypes[result.Moniker]
		if moduleType == "" {
			moduleType = "unknown"
		}

		if _, ok := summaryByType[moduleType]; !ok {
			summaryByType[moduleType] = &typeSummary{}
		}
		summaryByType[moduleType].Packages++

		if result.ExitCode != 0 {
			summaryByType[moduleType].Failed++
			failedResults = append(failedResults, failedResult{
				moniker:    result.Moniker,
				duration:   result.Duration,
				moduleType: moduleType,
				errors:     result.Errors,
				logPath:    result.LogPath,
			})
		} else if len(result.Warnings) > 0 {
			summaryByType[moduleType].Warnings++
			failedResults = append(failedResults, failedResult{
				moniker:    result.Moniker,
				duration:   result.Duration,
				moduleType: moduleType,
				errors:     result.Warnings,
				logPath:    result.LogPath,
			})
		} else {
			summaryByType[moduleType].Passed++
		}
	}

	// Build details with summary table
	var details []string

	// Collect and sort types
	types := make([]string, 0, len(summaryByType))
	for t := range summaryByType {
		types = append(types, t)
	}
	// Sort types alphabetically
	for i := 0; i < len(types)-1; i++ {
		for j := i + 1; j < len(types); j++ {
			if types[i] > types[j] {
				types[i], types[j] = types[j], types[i]
			}
		}
	}

	// Add summary table
	details = append(details,
		fmt.Sprintf("%-12s %8s %6s %6s %8s", "Type", "Packages", "Passed", "Failed", "Warnings"),
		strings.Repeat("-", 44))
	for _, t := range types {
		s := summaryByType[t]
		details = append(details, fmt.Sprintf("%-12s %8d %6d %6d %8d", t, s.Packages, s.Passed, s.Failed, s.Warnings))
	}

	// Add failed/warning results with error details
	if len(failedResults) > 0 {
		details = append(details, "")
		for _, fr := range failedResults {
			status := "✗"
			details = append(details, fmt.Sprintf("%s %s (%s) - %s",
				status, output.PackageDisplayName(fr.moniker), fr.moduleType, formatDuration(fr.duration)))
			for _, err := range fr.errors {
				details = append(details, formatErrorLines("  Error: ", err)...)
			}
			if fr.logPath != "" {
				details = append(details, fmt.Sprintf("  Log: %s", fr.logPath))
			}
		}
	}

	// Create summary data matching TUI structure
	data := &tui.SummaryData{
		Success:    failureCount == 0,
		TotalTime:  totalTime,
		RunSummary: runSummary,
		Details:    details,
	}

	return data
}

// printConsoleSummary prints a summary to the console for non-TUI mode.
func printConsoleSummary(ctx *ExecutionContext, totalTime time.Duration) {
	results := ctx.Results
	successCount := ctx.GetSuccessCount()
	failureCount := ctx.GetFailureCount()

	log.Info("")
	log.Info("=== Summary ===")
	log.Infof("Total: %d modules", len(results))
	log.Infof("Succeeded: %d", successCount)
	log.Infof("Failed: %d", failureCount)
	log.Infof("Duration: %s", formatDuration(totalTime))

	if failureCount > 0 {
		log.Info("")
		log.Info("Failed modules:")
		for _, result := range results {
			if result.ExitCode != 0 {
				log.Infof("  - %s (exit code %d)", output.PackageDisplayName(result.Moniker), result.ExitCode)
				if result.LogPath != "" {
					log.Infof("    Log: %s", result.LogPath)
				}
				for _, err := range result.Errors {
					for _, line := range formatErrorLines("    Error: ", err) {
						log.Info(line)
					}
				}
			}
		}
	}
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// formatErrorLines formats an error message for display by:
// - Truncating to maxErrorLength characters
// - Breaking long lines at maxLineLength
// Returns a slice of formatted lines with the given prefix.
func formatErrorLines(prefix, errMsg string) []string {
	// Truncate if too long
	if len(errMsg) > maxErrorLength {
		errMsg = errMsg[:maxErrorLength-3] + "..."
	}

	// Split on existing newlines first
	rawLines := strings.Split(errMsg, "\n")
	var result []string

	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Break long lines
		for len(line) > maxLineLength {
			// Find a good break point (space) near maxLineLength
			breakPoint := maxLineLength
			for i := maxLineLength; i > maxLineLength-20 && i > 0; i-- {
				if line[i] == ' ' {
					breakPoint = i
					break
				}
			}
			result = append(result, fmt.Sprintf("%s%s", prefix, strings.TrimSpace(line[:breakPoint])))
			line = strings.TrimSpace(line[breakPoint:])
		}
		if line != "" {
			result = append(result, fmt.Sprintf("%s%s", prefix, line))
		}
	}

	return result
}
