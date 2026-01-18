package cmdframework

import (
	"fmt"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
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

	// Build per-module results
	type moduleResult struct {
		moniker     string
		pkgTypes    string
		status      string // "passed", "failed", "warning"
		duration    time.Duration
		errors      []string
		logPath     string
	}
	var moduleResults []moduleResult

	for _, result := range results {
		pkgTypes := ctx.ModuleTypes[result.Moniker]
		if pkgTypes == "" {
			pkgTypes = "unknown"
		}

		mr := moduleResult{
			moniker:  result.Moniker,
			pkgTypes: pkgTypes,
			duration: result.Duration,
			logPath:  result.LogPath,
		}

		if result.ExitCode != 0 {
			mr.status = "failed"
			mr.errors = result.Errors
		} else if len(result.Warnings) > 0 {
			mr.status = "warning"
			mr.errors = result.Warnings
		} else {
			mr.status = "passed"
		}
		moduleResults = append(moduleResults, mr)
	}

	// Sort by moniker
	for i := 0; i < len(moduleResults)-1; i++ {
		for j := i + 1; j < len(moduleResults); j++ {
			if moduleResults[i].moniker > moduleResults[j].moniker {
				moduleResults[i], moduleResults[j] = moduleResults[j], moduleResults[i]
			}
		}
	}

	// Build details with summary table - one row per module
	var details []string
	tb := render.NewTableBuilder().
		WithHeaders("Module", "Packages", "Status")
	for _, mr := range moduleResults {
		statusIcon := "✓"
		if mr.status == "failed" {
			statusIcon = "✗"
		} else if mr.status == "warning" {
			statusIcon = "⚠"
		}
		tb.AddRow(mr.moniker, mr.pkgTypes, statusIcon)
	}
	details = append(details, tb.Build())

	// Add failed/warning results with error details
	hasErrors := false
	for _, mr := range moduleResults {
		if mr.status != "passed" {
			hasErrors = true
			break
		}
	}
	if hasErrors {
		details = append(details, "")
		for _, mr := range moduleResults {
			if mr.status == "passed" {
				continue
			}
			statusIcon := "✗"
			if mr.status == "warning" {
				statusIcon = "⚠"
			}
			details = append(details, fmt.Sprintf("%s %s - %s",
				statusIcon, output.PackageDisplayName(mr.moniker), formatDuration(mr.duration)))
			for _, err := range mr.errors {
				details = append(details, formatErrorLines("  Error: ", err)...)
			}
			if mr.logPath != "" {
				details = append(details, fmt.Sprintf("  Log: %s", mr.logPath))
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
