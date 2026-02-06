package cmdframework

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/adapters/tui"
)

const (
	maxErrorLength = 500
	maxLineLength  = 100
)

// SummaryGenerator is a function that generates custom summary data.
type SummaryGenerator func(ctx *ExecutionContext) *initsummary.Summary

// phaseSummary handles the summary phase:
// - Generate summary data (using incremental builder if available)
// - Display TUI summary or print to console
// - Return exit code.
func phaseSummary(ctx *ExecutionContext, customSummary SummaryGenerator) int {
	summaryStart := time.Now()
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
		// Check if summary was already sent via completion callback (immediate send path)
		alreadySent := ctx.SummaryBuilder != nil && ctx.SummaryBuilder.WasSummarySent()

		if !alreadySent {
			// Send summary using incremental builder if available, otherwise generate from scratch
			var summaryData *tui.SummaryData
			if ctx.SummaryBuilder != nil {
				summaryData = ctx.SummaryBuilder.Finalize(totalTime)
			} else {
				summaryData = generateTUISummary(ctx, totalTime)
			}
			log.Debugf("phaseSummary: summary generation took %v", time.Since(summaryStart))
			ctx.Orchestrator.SendSummary(summaryData)
		} else {
			log.Debugf("phaseSummary: summary already sent via callback")
		}

		// Wait for TUI to exit - handles user timer if active, else exits immediately
		log.Debugf("phaseSummary: calling WaitTUI")
		waitStart := time.Now()
		ctx.Orchestrator.WaitTUI()
		log.Debugf("phaseSummary: WaitTUI returned in %v", time.Since(waitStart))

		// TUI Hook 4: Wait for exit hold release (user interacting with output)
		if ctx.TUIHooks != nil && !ctx.Config.SkipTUIDelay {
			holdTimeout := 2 * time.Minute
			log.Debugf("phaseSummary: waiting for exit hold release (timeout=%v)", holdTimeout)
			if released := ctx.TUIHooks.WaitForRelease(context.Background(), holdTimeout); !released {
				log.Debugf("phaseSummary: exit hold timed out after %v", holdTimeout)
			}
		}

		ctx.Orchestrator.StopTUI()
	} else {
		// Print summary to console
		printConsoleSummary(ctx, totalTime)
	}

	log.Debugf("phaseSummary: total phase took %v", time.Since(summaryStart))
	return exitCode
}

// generateTUISummary creates TUI summary data from execution results.
func generateTUISummary(ctx *ExecutionContext, totalTime time.Duration) *tui.SummaryData {
	return generateComponentTUISummary(ctx, totalTime)
}

// moduleCache holds precomputed data for a module to avoid repeated calculations.
type moduleCache struct {
	status         orchestrator.ModuleStatus
	sortedComps    []orchestrator.UnitResult
	moduleDuration time.Duration
	errorCount     int
	warnCount      int
	testsTotal     int
	testsPassed    int
	testsFailed    int
}

// generateComponentTUISummary creates TUI summary showing module-level aggregated results.
// Table format varies by command type (build, test, lint, scan).
func generateComponentTUISummary(ctx *ExecutionContext, totalTime time.Duration) *tui.SummaryData {
	resultSets := ctx.ModuleResultSets

	// Pre-compute all module data in a single pass
	caches := make([]moduleCache, len(resultSets))
	var successCount, skippedCount, failureCount int

	for i := range resultSets {
		rs := &resultSets[i]
		cache := &caches[i]

		// Derive status once
		cache.status = rs.DeriveStatus()
		switch cache.status {
		case orchestrator.ModuleStatusFailed:
			failureCount++
		case orchestrator.ModuleStatusSuccess:
			successCount++
		case orchestrator.ModuleStatusSkipped:
			skippedCount++
		}

		// Sort components once
		cache.sortedComps = rs.GetSortedUnits()

		// Aggregate stats in single pass
		for _, comp := range rs.Units {
			if comp.Duration > cache.moduleDuration {
				cache.moduleDuration = comp.Duration
			}
			cache.errorCount += len(comp.Errors)
			cache.warnCount += len(comp.Warnings)
			cache.testsTotal += comp.TestsTotal
			cache.testsPassed += comp.TestsPassed
			cache.testsFailed += comp.TestsFailed
		}
	}

	// Build run summary line with command-appropriate verbs
	successVerb := "built"
	switch ctx.Config.Type {
	case CommandTypeTest:
		successVerb = "tested"
	case CommandTypeLint:
		successVerb = "linted"
	case CommandTypeScan:
		successVerb = "scanned"
	}

	var runSummary string
	switch {
	case skippedCount > 0 && successCount > 0 && failureCount > 0:
		runSummary = fmt.Sprintf("%d cached, %d %s, %d failed", skippedCount, successCount, successVerb, failureCount)
	case skippedCount > 0 && successCount > 0:
		runSummary = fmt.Sprintf("%d cached, %d %s", skippedCount, successCount, successVerb)
	case skippedCount > 0 && failureCount > 0:
		runSummary = fmt.Sprintf("%d cached, %d failed", skippedCount, failureCount)
	case successCount > 0 && failureCount > 0:
		runSummary = fmt.Sprintf("%d %s, %d failed", successCount, successVerb, failureCount)
	case skippedCount > 0:
		runSummary = fmt.Sprintf("%d cached", skippedCount)
	case successCount > 0:
		runSummary = fmt.Sprintf("%d %s", successCount, successVerb)
	case failureCount > 0:
		runSummary = fmt.Sprintf("%d failed", failureCount)
	default:
		runSummary = "0 modules"
	}

	// Create sorted indices by module name (avoid copying entire structs)
	sortedIndices := make([]int, len(resultSets))
	for i := range sortedIndices {
		sortedIndices[i] = i
	}
	sort.Slice(sortedIndices, func(i, j int) bool {
		return resultSets[sortedIndices[i]].Module < resultSets[sortedIndices[j]].Module
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

	for _, idx := range sortedIndices {
		rs := &resultSets[idx]
		cache := &caches[idx]

		// Build component names string from cached sorted components
		var components string
		if len(cache.sortedComps) <= 3 {
			// Fast path: few components, direct concatenation
			compNames := make([]string, len(cache.sortedComps))
			for i, comp := range cache.sortedComps {
				compNames[i] = comp.Component
			}
			components = strings.Join(compNames, ", ")
		} else {
			// Use strings.Builder for many components
			var sb strings.Builder
			for i, comp := range cache.sortedComps {
				if i > 0 {
					sb.WriteString(", ")
				}
				if sb.Len()+len(comp.Component) > 57 {
					sb.WriteString("...")
					break
				}
				sb.WriteString(comp.Component)
			}
			components = sb.String()
		}
		// Truncate if too long (max 60 chars)
		if len(components) > 60 {
			components = components[:57] + "..."
		}

		// Use cached status
		statusIcon := " ✓"
		if cache.status == orchestrator.ModuleStatusFailed {
			statusIcon = " ✗"
		} else if cache.warnCount > 0 {
			statusIcon = " ⚠"
		}

		moduleName := output.PackageDisplayName(rs.Module)
		duration := formatDuration(cache.moduleDuration)

		// Add row based on command type
		switch ctx.Config.Type {
		case CommandTypeTest:
			// Extract unique test types from component names (format: "subpath:testType")
			testTypes := extractUniqueTestTypes(cache.sortedComps)

			// Format test count: "passed/total" if failures, else "total"
			var testCount string
			if cache.testsTotal > 0 {
				if cache.testsFailed > 0 {
					testCount = fmt.Sprintf("%d/%d", cache.testsPassed, cache.testsTotal)
				} else {
					testCount = fmt.Sprintf("%d", cache.testsTotal)
				}
			} else {
				testCount = "-"
			}
			tb.AddRow(moduleName, testTypes, testCount, duration, statusIcon)
		case CommandTypeLint, CommandTypeScan:
			tb.AddRow(moduleName, components, cache.errorCount, cache.warnCount, duration, statusIcon)
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
	// Check if there are any failures/warnings using cached data
	hasFailures := false
	totalFailures := 0
	for i := range caches {
		if caches[i].status == orchestrator.ModuleStatusFailed || caches[i].warnCount > 0 {
			hasFailures = true
			totalFailures++
		}
	}

	if hasFailures {
		details = append(details, "")
		failCount := 0
		const maxFailures = 5
		for _, idx := range sortedIndices {
			if failCount >= maxFailures {
				break
			}
			rs := &resultSets[idx]
			cache := &caches[idx]

			if cache.status != orchestrator.ModuleStatusFailed && cache.warnCount == 0 {
				continue
			}
			failCount++

			// Module header with status
			statusIcon := "✗"
			if cache.status != orchestrator.ModuleStatusFailed {
				statusIcon = "⚠"
			}
			moduleName := output.PackageDisplayName(rs.Module)
			details = append(details, fmt.Sprintf("%s %s", statusIcon, moduleName))

			// Show first error/warning from failed components (use cached sorted)
			for _, comp := range cache.sortedComps {
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
				break // Only show first failed component per module
			}

			// List ALL log paths from every component in this module
			for _, comp := range cache.sortedComps {
				if comp.LogPath != "" {
					details = append(details, fmt.Sprintf("    Log: %s", comp.LogPath))
				}
			}
		}
		if failCount >= maxFailures && totalFailures > maxFailures {
			remaining := totalFailures - maxFailures
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

// printConsoleSummary prints a summary to the console for non-TUI mode.
func printConsoleSummary(ctx *ExecutionContext, totalTime time.Duration) {
	printComponentConsoleSummary(ctx, totalTime)
}

// printComponentConsoleSummary prints component-level summary to console.
// For test commands, displays a module-level table with test types.
func printComponentConsoleSummary(ctx *ExecutionContext, totalTime time.Duration) {
	resultSets := ctx.ModuleResultSets

	// Pre-compute all module data in a single pass (reuse moduleCache type)
	caches := make([]moduleCache, len(resultSets))
	var moduleSuccessCount, moduleFailureCount int

	for i := range resultSets {
		rs := &resultSets[i]
		cache := &caches[i]

		cache.status = rs.DeriveStatus()
		if cache.status == orchestrator.ModuleStatusFailed {
			moduleFailureCount++
		} else {
			moduleSuccessCount++
		}

		cache.sortedComps = rs.GetSortedUnits()

		for _, comp := range rs.Units {
			if comp.Duration > cache.moduleDuration {
				cache.moduleDuration = comp.Duration
			}
			cache.testsTotal += comp.TestsTotal
			cache.testsPassed += comp.TestsPassed
			cache.testsFailed += comp.TestsFailed
		}
	}

	// Create sorted indices by module name
	sortedIndices := make([]int, len(resultSets))
	for i := range sortedIndices {
		sortedIndices[i] = i
	}
	sort.Slice(sortedIndices, func(i, j int) bool {
		return resultSets[sortedIndices[i]].Module < resultSets[sortedIndices[j]].Module
	})

	// Build module-level table for test commands
	if ctx.Config.Type == CommandTypeTest && len(resultSets) > 0 {
		tb := render.NewTableBuilder().
			WithHeaders("Module", "Test Types", "#Test", "Time", "Stat")

		for _, idx := range sortedIndices {
			rs := &resultSets[idx]
			cache := &caches[idx]

			statusIcon := " ✓"
			if cache.status == orchestrator.ModuleStatusFailed {
				statusIcon = " ✗"
			}

			moduleName := output.PackageDisplayName(rs.Module)
			duration := formatDuration(cache.moduleDuration)
			testTypes := extractUniqueTestTypes(cache.sortedComps)

			// Format test count
			var testCount string
			if cache.testsTotal > 0 {
				if cache.testsFailed > 0 {
					testCount = fmt.Sprintf("%d/%d", cache.testsPassed, cache.testsTotal)
				} else {
					testCount = fmt.Sprintf("%d", cache.testsTotal)
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
		log.Infof("Total: %d modules", len(resultSets))
		log.Infof("Duration: %s", formatDuration(totalTime))
	}

	// Show failed modules with details and all log paths
	if moduleFailureCount > 0 {
		log.Info("")
		log.Info("Failed modules:")
		for _, idx := range sortedIndices {
			rs := &resultSets[idx]
			cache := &caches[idx]

			if cache.status != orchestrator.ModuleStatusFailed {
				continue
			}

			moduleName := output.PackageDisplayName(rs.Module)
			log.Infof("  ✗ %s", moduleName)

			// Show first error from failed components
			for _, comp := range cache.sortedComps {
				if comp.ExitCode == 0 {
					continue
				}
				for _, err := range comp.Errors {
					for _, line := range formatErrorLines("    Error: ", err) {
						log.Info(line)
					}
				}
				break // Only show first failed component's errors
			}

			// List ALL log paths from every component in this module
			for _, comp := range cache.sortedComps {
				if comp.LogPath != "" {
					log.Infof("    Log: %s", comp.LogPath)
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

// extractUniqueTestTypes extracts unique test types from component results.
// Uses the Handler field which contains the test type (e.g., "gotest", "godog").
// Returns a comma-separated string of unique test types.
func extractUniqueTestTypes(components []orchestrator.UnitResult) string {
	if len(components) == 0 {
		return "-"
	}

	// Pre-size map based on typical test type count (usually 1-3)
	seen := make(map[string]struct{}, 4)
	types := make([]string, 0, 4)

	for _, comp := range components {
		// Use Handler field which contains the test type
		testType := comp.Handler
		if testType == "" {
			// Fallback: extract test type from component name
			// Component format: "name/testType" (e.g., "config/gotest", "docs-drawio-cache/godog")
			testType = comp.Component
			if slashIdx := strings.LastIndex(comp.Component, "/"); slashIdx >= 0 {
				testType = comp.Component[slashIdx+1:]
			}
		}

		if testType != "" {
			if _, exists := seen[testType]; !exists {
				seen[testType] = struct{}{}
				types = append(types, testType)
			}
		}
	}

	if len(types) == 0 {
		return "-"
	}

	// Sort for consistent output
	sort.Strings(types)
	return strings.Join(types, ", ")
}
