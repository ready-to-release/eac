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

// generateComponentTUISummary creates TUI summary showing module-level aggregated results.
// Table format varies by command type (build, test, lint, scan).
func generateComponentTUISummary(ctx *ExecutionContext, totalTime time.Duration) *tui.SummaryData {
	resultSets := ctx.ComponentResultSets

	// Count successes, skipped (cached), and failures at module level
	var successCount, skippedCount, failureCount int
	for _, rs := range resultSets {
		status := rs.DeriveStatus()
		switch status {
		case orchestrator.ModuleStatusFailed:
			failureCount++
		case orchestrator.ModuleStatusSuccess:
			successCount++
		case orchestrator.ModuleStatusSkipped:
			skippedCount++
		}
	}

	// Build run summary line
	var runParts []string
	if skippedCount > 0 {
		runParts = append(runParts, fmt.Sprintf("%d cached", skippedCount))
	}
	if successCount > 0 {
		runParts = append(runParts, fmt.Sprintf("%d built", successCount))
	}
	if failureCount > 0 {
		runParts = append(runParts, fmt.Sprintf("%d failed", failureCount))
	}
	runSummary := strings.Join(runParts, ", ")
	if runSummary == "" {
		runSummary = "0 modules"
	}

	// Sort resultSets by module name
	sortedSets := make([]orchestrator.ComponentResultSet, len(resultSets))
	copy(sortedSets, resultSets)
	sort.Slice(sortedSets, func(i, j int) bool {
		return sortedSets[i].Module < sortedSets[j].Module
	})

	// Build table based on command type
	var details []string
	var tb *render.TableBuilder

	switch ctx.Config.Type {
	case CommandTypeTest:
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Test Types", "#Test", "Time", "Stat")
	case CommandTypeLint:
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Components", "#Err", "#Warn", "Time", "Stat")
	case CommandTypeScan:
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Components", "#Err", "#Warn", "Time", "Stat")
	default: // CommandTypeBuild
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Components", "Time", "Stat")
	}

	for _, rs := range sortedSets {
		// Get sorted component names
		sortedComps := rs.GetSortedComponents()
		compNames := make([]string, len(sortedComps))
		for i, comp := range sortedComps {
			compNames[i] = comp.Component
		}
		components := strings.Join(compNames, ", ")
		// Truncate if too long (max 60 chars)
		if len(components) > 60 {
			components = components[:57] + "..."
		}

		// Aggregate component stats
		// Use max duration (wall-clock approximation for parallel execution)
		var moduleDuration time.Duration
		var errorCount, warnCount int
		var testsTotal, testsPassed, testsFailed int
		for _, comp := range rs.Components {
			// Components run in parallel, so use max duration for wall-clock time
			if comp.Duration > moduleDuration {
				moduleDuration = comp.Duration
			}
			errorCount += len(comp.Errors)
			warnCount += len(comp.Warnings)
			testsTotal += comp.TestsTotal
			testsPassed += comp.TestsPassed
			testsFailed += comp.TestsFailed
		}

		// Derive status
		statusIcon := " ✓"
		status := rs.DeriveStatus()
		if status == orchestrator.ModuleStatusFailed {
			statusIcon = " ✗"
		} else if warnCount > 0 {
			statusIcon = " ⚠"
		}

		moduleName := output.PackageDisplayName(rs.Module)
		duration := formatDuration(moduleDuration)

		// Add row based on command type
		switch ctx.Config.Type {
		case CommandTypeTest:
			// Extract unique test types from component names (format: "subpath:testType")
			testTypes := extractUniqueTestTypes(sortedComps)

			// Format test count: "passed/total" if failures, else "total"
			var testCount string
			if testsTotal > 0 {
				if testsFailed > 0 {
					testCount = fmt.Sprintf("%d/%d", testsPassed, testsTotal)
				} else {
					testCount = fmt.Sprintf("%d", testsTotal)
				}
			} else {
				testCount = "-"
			}
			tb.AddRow(moduleName, testTypes, testCount, duration, statusIcon)
		case CommandTypeLint, CommandTypeScan:
			tb.AddRow(moduleName, components, errorCount, warnCount, duration, statusIcon)
		default: // CommandTypeBuild
			tb.AddRow(moduleName, components, duration, statusIcon)
		}
	}

	// Split table into individual lines for TUI rendering
	tableStr := tb.Build()
	for _, line := range strings.Split(tableStr, "\n") {
		if line != "" {
			details = append(details, line)
		}
	}

	// Add failed/warning results with error details (top 5 failures)
	if hasComponentFailures(resultSets) {
		details = append(details, "")
		failCount := 0
		const maxFailures = 5
		for _, rs := range sortedSets {
			if failCount >= maxFailures {
				break
			}
			status := rs.DeriveStatus()
			if status != orchestrator.ModuleStatusFailed {
				// Check for warnings
				hasWarnings := false
				for _, comp := range rs.Components {
					if len(comp.Warnings) > 0 {
						hasWarnings = true
						break
					}
				}
				if !hasWarnings {
					continue
				}
			}
			failCount++

			// Module header with status
			statusIcon := "✗"
			if status != orchestrator.ModuleStatusFailed {
				statusIcon = "⚠"
			}
			moduleName := output.PackageDisplayName(rs.Module)
			details = append(details, fmt.Sprintf("%s %s", statusIcon, moduleName))

			// Show first error/warning from failed components
			for _, comp := range rs.GetSortedComponents() {
				if comp.ExitCode == 0 && len(comp.Warnings) == 0 {
					continue
				}
				// Show component name and first error
				if len(comp.Errors) > 0 {
					errMsg := comp.Errors[0]
					if len(errMsg) > 80 {
						errMsg = errMsg[:77] + "..."
					}
					details = append(details, fmt.Sprintf("    %s: %s", comp.Component, errMsg))
				} else if len(comp.Warnings) > 0 {
					warnMsg := comp.Warnings[0]
					if len(warnMsg) > 80 {
						warnMsg = warnMsg[:77] + "..."
					}
					details = append(details, fmt.Sprintf("    %s: %s", comp.Component, warnMsg))
				}
				// Show log path
				if comp.LogPath != "" {
					details = append(details, fmt.Sprintf("    Log: %s", comp.LogPath))
				}
				break // Only show first failed component per module
			}
		}
		if failCount >= maxFailures && countFailures(sortedSets) > maxFailures {
			remaining := countFailures(sortedSets) - maxFailures
			details = append(details, fmt.Sprintf("  ... and %d more failures", remaining))
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
// For test commands, displays a module-level table with test types.
func printComponentConsoleSummary(ctx *ExecutionContext, totalTime time.Duration) {
	resultSets := ctx.ComponentResultSets

	// Sort by module name for consistent output
	sortedSets := make([]orchestrator.ComponentResultSet, len(resultSets))
	copy(sortedSets, resultSets)
	sort.Slice(sortedSets, func(i, j int) bool {
		return sortedSets[i].Module < sortedSets[j].Module
	})

	// Count at module level
	var moduleSuccessCount, moduleFailureCount int
	for _, rs := range sortedSets {
		status := rs.DeriveStatus()
		if status == orchestrator.ModuleStatusFailed {
			moduleFailureCount++
		} else {
			moduleSuccessCount++
		}
	}

	// Build module-level table for test commands
	if ctx.Config.Type == CommandTypeTest && len(sortedSets) > 0 {
		tb := render.NewTableBuilder().
			WithHeaders("Module", "Test Types", "#Test", "Time", "Stat")

		for _, rs := range sortedSets {
			sortedComps := rs.GetSortedComponents()

			// Use max duration (wall-clock approximation for parallel execution)
			var moduleDuration time.Duration
			var testsTotal, testsPassed, testsFailed int
			for _, comp := range rs.Components {
				if comp.Duration > moduleDuration {
					moduleDuration = comp.Duration // Max, not sum
				}
				testsTotal += comp.TestsTotal
				testsPassed += comp.TestsPassed
				testsFailed += comp.TestsFailed
			}

			// Derive status
			statusIcon := " ✓"
			status := rs.DeriveStatus()
			if status == orchestrator.ModuleStatusFailed {
				statusIcon = " ✗"
			}

			moduleName := output.PackageDisplayName(rs.Module)
			duration := formatDuration(moduleDuration)

			// Extract unique test types from component names
			testTypes := extractUniqueTestTypes(sortedComps)

			// Format test count
			var testCount string
			if testsTotal > 0 {
				if testsFailed > 0 {
					testCount = fmt.Sprintf("%d/%d", testsPassed, testsTotal)
				} else {
					testCount = fmt.Sprintf("%d", testsTotal)
				}
			} else {
				testCount = "-"
			}

			tb.AddRow(moduleName, testTypes, testCount, duration, statusIcon)
		}

		log.Info("")
		for _, line := range strings.Split(tb.Build(), "\n") {
			if line != "" {
				log.Info(line)
			}
		}
	}

	// Print summary totals
	log.Info("")
	if moduleSuccessCount > 0 || moduleFailureCount > 0 {
		summaryLine := fmt.Sprintf("✓ PASSED (%s)", formatDuration(totalTime))
		if moduleFailureCount > 0 {
			summaryLine = fmt.Sprintf("✗ FAILED (%s)", formatDuration(totalTime))
		}
		log.Info(summaryLine)
		log.Infof("  %d modules succeeded", moduleSuccessCount)
		if moduleFailureCount > 0 {
			log.Infof("  %d modules failed", moduleFailureCount)
		}
	} else {
		log.Info("=== Summary ===")
		log.Infof("Total: %d modules", len(sortedSets))
		log.Infof("Duration: %s", formatDuration(totalTime))
	}

	// Show failed components with details
	if moduleFailureCount > 0 {
		log.Info("")
		log.Info("Failed components:")
		for _, rs := range sortedSets {
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
			if result.ExitCode > 0 { // Only actual failures, not skipped (-1)
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

// countFailures counts the number of failed or warned modules.
func countFailures(resultSets []orchestrator.ComponentResultSet) int {
	count := 0
	for _, rs := range resultSets {
		status := rs.DeriveStatus()
		if status == orchestrator.ModuleStatusFailed {
			count++
			continue
		}
		// Check for warnings
		for _, comp := range rs.Components {
			if len(comp.Warnings) > 0 {
				count++
				break
			}
		}
	}
	return count
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

// extractUniqueTestTypes extracts unique test types from component names.
// Component names are in format "subpath:testType" (e.g., "config:gotest", "cli:godog").
// Returns a comma-separated string of unique test types (e.g., "gotest, godog").
func extractUniqueTestTypes(components []orchestrator.ComponentResult) string {
	seen := make(map[string]bool)
	var types []string

	for _, comp := range components {
		// Extract test type from component name (format: "subpath:testType")
		testType := comp.Component
		if colonIdx := strings.LastIndex(comp.Component, ":"); colonIdx >= 0 {
			testType = comp.Component[colonIdx+1:]
		}

		log.Debugf("extractUniqueTestTypes: component=%s -> testType=%s", comp.Component, testType)

		if !seen[testType] {
			seen[testType] = true
			types = append(types, testType)
		}
	}

	// Sort for consistent output
	sort.Strings(types)
	result := strings.Join(types, ", ")
	log.Debugf("extractUniqueTestTypes: found types=%v", types)
	return result
}
