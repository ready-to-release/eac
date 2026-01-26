package cmdframework

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/initsummary"
	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
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

	// Debug: Log results and exit code to help diagnose mismatch
	log.Debugf("phaseSummary: %d results, exit code: %d", len(ctx.Results), exitCode)
	for i, r := range ctx.Results {
		if r.ExitCode != 0 {
			log.Debugf("  Result[%d]: moniker=%s, exitCode=%d, errors=%v", i, r.Moniker, r.ExitCode, r.Errors)
		}
	}

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
	// Use component-level results if available, otherwise fall back to module-level
	if len(ctx.ComponentResultSets) > 0 {
		return generateComponentTUISummary(ctx, totalTime)
	}
	return generateModuleTUISummary(ctx, totalTime)
}

// generateComponentTUISummary creates TUI summary showing component-level results.
// Format: Name (module:component), Status, Duration
func generateComponentTUISummary(ctx *ExecutionContext, totalTime time.Duration) *tui.SummaryData {
	resultSets := ctx.ComponentResultSets

	// Count successes and failures at component level
	var successCount, failureCount int
	for _, rs := range resultSets {
		for _, comp := range rs.Components {
			if comp.ExitCode == 0 {
				successCount++
			} else {
				failureCount++
			}
		}
	}

	// Build run summary line
	runSummary := fmt.Sprintf("%d succeeded", successCount)
	if failureCount > 0 {
		runSummary += fmt.Sprintf(", %d failed", failureCount)
	}

	// Build details with summary table - one row per component
	var details []string
	tb := render.NewTableBuilder().
		WithHeaders("Name", "Status", "Duration")

	// ComponentResultSets are already sorted by module, components sorted within
	for _, rs := range resultSets {
		for _, comp := range rs.GetSortedComponents() {
			// Format: "module:component" (truncated to fit table width)
			name := output.TruncateComponentName(rs.Module, comp.Component, output.NameWidth)
			statusIcon := "✓"
			if comp.ExitCode != 0 {
				statusIcon = "✗"
			} else if len(comp.Warnings) > 0 {
				statusIcon = "⚠"
			}
			duration := formatDuration(comp.Duration)
			tb.AddRow(name, statusIcon, duration)
		}
	}

	// Split table into individual lines for TUI rendering
	tableStr := tb.Build()
	for _, line := range strings.Split(tableStr, "\n") {
		if line != "" {
			details = append(details, line)
		}
	}

	// Add failed/warning results with error details
	if hasComponentFailures(resultSets) {
		details = append(details, "")
		for _, rs := range resultSets {
			for _, comp := range rs.GetSortedComponents() {
				if comp.ExitCode == 0 && len(comp.Warnings) == 0 {
					continue
				}
				statusIcon := "✗"
				if comp.ExitCode == 0 && len(comp.Warnings) > 0 {
					statusIcon = "⚠"
				}
				name := output.TruncateComponentName(rs.Module, comp.Component, output.NameWidth)
				details = append(details, fmt.Sprintf("%s %s - %s",
					statusIcon, name, formatDuration(comp.Duration)))
				for _, err := range comp.Errors {
					details = append(details, formatErrorLines("  Error: ", err)...)
				}
				for _, warn := range comp.Warnings {
					details = append(details, formatErrorLines("  Warning: ", warn)...)
				}
				if comp.LogPath != "" {
					details = append(details, fmt.Sprintf("  Log: %s", comp.LogPath))
				}
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

// generateModuleTUISummary creates TUI summary showing module-level results (legacy).
func generateModuleTUISummary(ctx *ExecutionContext, totalTime time.Duration) *tui.SummaryData {
	results := ctx.Results
	successCount := ctx.GetSuccessCount()
	failureCount := ctx.GetFailureCount()

	// Build run summary line
	runSummary := fmt.Sprintf("%d succeeded", successCount)
	if failureCount > 0 {
		runSummary += fmt.Sprintf(", %d failed", failureCount)
	}

	// Build per-module results
	var moduleResults []moduleResult

	for _, result := range results {
		components := ctx.ModuleTypes[result.Moniker]
		if components == "" {
			components = "unknown"
		}

		mr := moduleResult{
			moniker:    result.Moniker,
			components: components,
			duration:   result.Duration,
			logPath:    result.LogPath,
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
	sort.Slice(moduleResults, func(i, j int) bool {
		return moduleResults[i].moniker < moduleResults[j].moniker
	})

	// Build details with summary table - one row per module
	var details []string
	tb := render.NewTableBuilder().
		WithHeaders("Module", "Components", "Status")
	for _, mr := range moduleResults {
		statusIcon := "✓"
		if mr.status == "failed" {
			statusIcon = "✗"
		} else if mr.status == "warning" {
			statusIcon = "⚠"
		}
		tb.AddRow(mr.moniker, mr.components, statusIcon)
	}
	// Split table into individual lines for TUI rendering
	tableStr := tb.Build()
	for _, line := range strings.Split(tableStr, "\n") {
		if line != "" {
			details = append(details, line)
		}
	}

	// Add failed/warning results with error details
	if hasModuleFailures(moduleResults) {
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
	// Use component-level results if available
	if len(ctx.ComponentResultSets) > 0 {
		printComponentConsoleSummary(ctx, totalTime)
		return
	}
	printModuleConsoleSummary(ctx, totalTime)
}

// printComponentConsoleSummary prints component-level summary to console.
func printComponentConsoleSummary(ctx *ExecutionContext, totalTime time.Duration) {
	resultSets := ctx.ComponentResultSets

	// Count at component level
	var totalComponents, successCount, failureCount int
	for _, rs := range resultSets {
		for _, comp := range rs.Components {
			totalComponents++
			if comp.ExitCode == 0 {
				successCount++
			} else {
				failureCount++
			}
		}
	}

	log.Info("")
	log.Info("=== Summary ===")
	log.Infof("Total: %d components", totalComponents)
	log.Infof("Succeeded: %d", successCount)
	log.Infof("Failed: %d", failureCount)
	log.Infof("Duration: %s", formatDuration(totalTime))

	if failureCount > 0 {
		log.Info("")
		log.Info("Failed components:")
		for _, rs := range resultSets {
			for _, comp := range rs.GetSortedComponents() {
				if comp.ExitCode == 0 {
					continue
				}
				name := output.TruncateComponentName(rs.Module, comp.Component, output.NameWidth)
				log.Infof("  - %s (exit code %d)", name, comp.ExitCode)
				if comp.LogPath != "" {
					log.Infof("    Log: %s", comp.LogPath)
				}
				for _, err := range comp.Errors {
					for _, line := range formatErrorLines("    Error: ", err) {
						log.Info(line)
					}
				}
			}
		}
	}
}

// printModuleConsoleSummary prints module-level summary to console (legacy).
func printModuleConsoleSummary(ctx *ExecutionContext, totalTime time.Duration) {
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

// moduleResult holds per-module result data for summary display.
type moduleResult struct {
	moniker    string
	components string
	status     string // "passed", "failed", "warning"
	duration   time.Duration
	errors     []string
	logPath    string
}

// hasComponentFailures checks if any component has failures or warnings.
func hasComponentFailures(resultSets []orchestrator.ComponentResultSet) bool {
	for _, rs := range resultSets {
		for _, comp := range rs.Components {
			if comp.ExitCode != 0 || len(comp.Warnings) > 0 {
				return true
			}
		}
	}
	return false
}

// hasModuleFailures checks if any module result is not passed.
func hasModuleFailures(moduleResults []moduleResult) bool {
	for _, mr := range moduleResults {
		if mr.status != "passed" {
			return true
		}
	}
	return false
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
