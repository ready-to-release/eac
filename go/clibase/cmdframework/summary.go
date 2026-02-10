package cmdframework

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/initsummary"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/clibase/output"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
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
			var summaryData *display.SummaryData
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
func generateTUISummary(ctx *ExecutionContext, totalTime time.Duration) *display.SummaryData {
	return generateComponentTUISummary(ctx, totalTime)
}

// handlerStats holds per-handler aggregate statistics for test summary.
type handlerStats struct {
	testCount   int
	passed      int
	failed      int
	maxWait     time.Duration // max queue time before execution started
	minStartedAt time.Time   // earliest UoW start (for cycle span)
	maxEndedAt   time.Time   // latest UoW end (for cycle span)
	hasFail     bool
}

// accumulateHandlerStats updates the handler stats map with a single UnitResult.
// cmdStartTime is the command start time for wait-time calculation.
func accumulateHandlerStats(byHandler map[string]*handlerStats, comp orchestrator.UnitResult, cmdStartTime time.Time) {
	handler := comp.Handler
	if handler == "" {
		handler = comp.Component
		if slashIdx := strings.LastIndex(comp.Component, "/"); slashIdx >= 0 {
			handler = comp.Component[slashIdx+1:]
		}
	}
	hs, ok := byHandler[handler]
	if !ok {
		hs = &handlerStats{}
		byHandler[handler] = hs
	}
	hs.testCount += comp.TestsTotal
	hs.passed += comp.TestsPassed
	hs.failed += comp.TestsFailed
	if comp.ExitCode > 0 {
		hs.hasFail = true
	}
	if !comp.StartedAt.IsZero() {
		endedAt := comp.StartedAt.Add(comp.Duration)
		if hs.minStartedAt.IsZero() || comp.StartedAt.Before(hs.minStartedAt) {
			hs.minStartedAt = comp.StartedAt
		}
		if endedAt.After(hs.maxEndedAt) {
			hs.maxEndedAt = endedAt
		}
		if !cmdStartTime.IsZero() {
			waitTime := comp.StartedAt.Sub(cmdStartTime)
			if waitTime > hs.maxWait {
				hs.maxWait = waitTime
			}
		}
	}
}

// addTestTypeRows adds per-handler-type rows to the table builder.
// The module name appears on the first row only; subsequent rows leave it blank.
func addTestTypeRows(tb *render.TableBuilder, moduleName string, byHandler map[string]*handlerStats) {
	handlerNames := make([]string, 0, len(byHandler))
	for h := range byHandler {
		handlerNames = append(handlerNames, h)
	}
	sort.Strings(handlerNames)

	for i, handler := range handlerNames {
		hs := byHandler[handler]
		var testCount string
		if hs.testCount > 0 {
			if hs.failed > 0 {
				testCount = fmt.Sprintf("%d/%d", hs.passed, hs.testCount)
			} else {
				testCount = fmt.Sprintf("%d", hs.testCount)
			}
		} else if !hs.minStartedAt.IsZero() {
			testCount = "0"
		} else {
			testCount = "-"
		}
		wait := formatDuration(hs.maxWait)
		cycle := formatDuration(hs.maxEndedAt.Sub(hs.minStartedAt))
		typeStatus := " ✓"
		if hs.hasFail {
			typeStatus = " ✗"
		}
		rowModule := ""
		if i == 0 {
			rowModule = moduleName
		}
		tb.AddRow(rowModule, handler, testCount, wait, cycle, typeStatus)
	}
}

// addBuildComponentRows adds per-component rows to the table builder.
// The module name appears on the first row only; subsequent rows leave it blank.
// cmdStartTime is used to compute wait time.
func addBuildComponentRows(tb *render.TableBuilder, moduleName string, comps []orchestrator.UnitResult, cmdStartTime time.Time) {
	for i, comp := range comps {
		statusIcon := " ✓"
		if comp.ExitCode > 0 {
			statusIcon = " ✗"
		} else if len(comp.Warnings) > 0 {
			statusIcon = " ⚠"
		}
		var wait, cycle string
		if !comp.StartedAt.IsZero() && !cmdStartTime.IsZero() {
			wait = formatDuration(comp.StartedAt.Sub(cmdStartTime))
		} else {
			wait = "-"
		}
		cycle = formatDuration(comp.Duration)
		rowModule := ""
		if i == 0 {
			rowModule = moduleName
		}
		tb.AddRow(rowModule, comp.Component, wait, cycle, statusIcon)
	}
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
	byHandler      map[string]*handlerStats // per-handler aggregates (for test summary)
}

// generateComponentTUISummary creates TUI summary showing module-level aggregated results.
// Table format varies by command type (build, test, lint, scan).
func generateComponentTUISummary(ctx *ExecutionContext, totalTime time.Duration) *display.SummaryData {
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
		cache.byHandler = make(map[string]*handlerStats, 4)
		for _, comp := range rs.Units {
			if comp.Duration > cache.moduleDuration {
				cache.moduleDuration = comp.Duration
			}
			cache.errorCount += len(comp.Errors)
			cache.warnCount += len(comp.Warnings)
			cache.testsTotal += comp.TestsTotal
			cache.testsPassed += comp.TestsPassed
			cache.testsFailed += comp.TestsFailed
			accumulateHandlerStats(cache.byHandler, comp, ctx.StartTime)
		}
	}

	// Build run summary line with command-appropriate verbs
	successVerb := "built"
	if desc, ok := core.GetActionDescriptor(ctx.Config.Type); ok {
		successVerb = desc.PastVerb
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

	// Sort modules by display order
	var displayOrder *config.DisplayOrder
	if ctx.EACConfig != nil {
		displayOrder = ctx.EACConfig.Repository.DisplayOrder
	}
	sortedIndices := make([]int, len(resultSets))
	for i := range sortedIndices {
		sortedIndices[i] = i
	}
	moduleNames := make([]string, len(resultSets))
	for i := range resultSets {
		moduleNames[i] = resultSets[i].Module
	}
	sortModulesByDisplayOrder(moduleNames, displayOrder)
	// Build name->index map for reordering
	nameToIdx := make(map[string]int, len(resultSets))
	for i := range resultSets {
		nameToIdx[resultSets[i].Module] = i
	}
	for i, name := range moduleNames {
		sortedIndices[i] = nameToIdx[name]
	}

	// Build table based on command type
	var details []string
	var tb *render.TableBuilder

	switch ctx.Config.Type {
	case core.ActionTest:
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Type", "#Test", "Wait", "Cycle", "Stat")
	case core.ActionLint:
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Components", "#Err", "#Warn", "Time", "Stat")
	case core.ActionScan:
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Components", "#Err", "#Warn", "Time", "Stat")
	default: // core.ActionBuild
		tb = render.NewTableBuilder().
			WithHeaders("Module", "Component", "Wait", "Cycle", "Stat")
	}

	for _, idx := range sortedIndices {
		rs := &resultSets[idx]
		cache := &caches[idx]

		moduleName := output.PackageDisplayName(rs.Module)

		// Add row based on command type
		switch ctx.Config.Type {
		case core.ActionTest:
			addTestTypeRows(tb, moduleName, cache.byHandler)
		case core.ActionLint, core.ActionScan:
			// Build component names string from cached sorted components
			var components string
			if len(cache.sortedComps) <= 3 {
				compNames := make([]string, len(cache.sortedComps))
				for i, comp := range cache.sortedComps {
					compNames[i] = comp.Component
				}
				components = strings.Join(compNames, ", ")
			} else {
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
			if len(components) > 60 {
				components = components[:57] + "..."
			}
			statusIcon := " ✓"
			if cache.status == orchestrator.ModuleStatusFailed {
				statusIcon = " ✗"
			} else if cache.warnCount > 0 {
				statusIcon = " ⚠"
			}
			duration := formatDuration(cache.moduleDuration)
			tb.AddRow(moduleName, components, cache.errorCount, cache.warnCount, duration, statusIcon)
		default: // core.ActionBuild
			addBuildComponentRows(tb, moduleName, cache.sortedComps, ctx.StartTime)
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

		// Order: root failures (ran and failed) before dependency failures
		failedIndices := make([]int, 0, totalFailures)
		for _, idx := range sortedIndices {
			cache := &caches[idx]
			if cache.status == orchestrator.ModuleStatusFailed || cache.warnCount > 0 {
				failedIndices = append(failedIndices, idx)
			}
		}
		sort.SliceStable(failedIndices, func(i, j int) bool {
			return hasRootFailure(caches[failedIndices[i]].sortedComps) &&
				!hasRootFailure(caches[failedIndices[j]].sortedComps)
		})

		const maxFailures = 5
		for failCount, idx := range failedIndices {
			if failCount >= maxFailures {
				break
			}
			rs := &resultSets[idx]
			cache := &caches[idx]

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

			// List log paths from failed components only
			for _, comp := range cache.sortedComps {
				if comp.LogPath != "" && comp.ExitCode > 0 {
					details = append(details, fmt.Sprintf("    Log: %s", comp.LogPath))
				}
			}
		}
		if len(failedIndices) > maxFailures {
			remaining := len(failedIndices) - maxFailures
			details = append(details, fmt.Sprintf("  ... and %d more failures", remaining))
		}
	}

	// Create summary data matching TUI structure
	data := &display.SummaryData{
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

		cache.byHandler = make(map[string]*handlerStats, 4)
		for _, comp := range rs.Units {
			if comp.Duration > cache.moduleDuration {
				cache.moduleDuration = comp.Duration
			}
			cache.testsTotal += comp.TestsTotal
			cache.testsPassed += comp.TestsPassed
			cache.testsFailed += comp.TestsFailed
			accumulateHandlerStats(cache.byHandler, comp, ctx.StartTime)
		}
	}

	// Sort modules by display order
	var displayOrder *config.DisplayOrder
	if ctx.EACConfig != nil {
		displayOrder = ctx.EACConfig.Repository.DisplayOrder
	}
	sortedIndices := make([]int, len(resultSets))
	for i := range sortedIndices {
		sortedIndices[i] = i
	}
	moduleNames := make([]string, len(resultSets))
	for i := range resultSets {
		moduleNames[i] = resultSets[i].Module
	}
	sortModulesByDisplayOrder(moduleNames, displayOrder)
	nameToIdx := make(map[string]int, len(resultSets))
	for i := range resultSets {
		nameToIdx[resultSets[i].Module] = i
	}
	for i, name := range moduleNames {
		sortedIndices[i] = nameToIdx[name]
	}

	// Build module-level table for test commands
	if ctx.Config.Type == core.ActionTest && len(resultSets) > 0 {
		tb := render.NewTableBuilder().
			WithHeaders("Module", "Type", "#Test", "Wait", "Cycle", "Stat")

		for _, idx := range sortedIndices {
			rs := &resultSets[idx]
			cache := &caches[idx]
			moduleName := output.PackageDisplayName(rs.Module)
			addTestTypeRows(tb, moduleName, cache.byHandler)
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

	// Show failed modules with details and log paths from failed components
	if moduleFailureCount > 0 {
		log.Info("")
		log.Info("Failed modules:")

		// Order: root failures (ran and failed) before dependency failures
		failedIndices := make([]int, 0, moduleFailureCount)
		for _, idx := range sortedIndices {
			if caches[idx].status == orchestrator.ModuleStatusFailed {
				failedIndices = append(failedIndices, idx)
			}
		}
		sort.SliceStable(failedIndices, func(i, j int) bool {
			return hasRootFailure(caches[failedIndices[i]].sortedComps) &&
				!hasRootFailure(caches[failedIndices[j]].sortedComps)
		})

		for _, idx := range failedIndices {
			rs := &resultSets[idx]
			cache := &caches[idx]
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

			// List log paths from failed components only
			for _, comp := range cache.sortedComps {
				if comp.LogPath != "" && comp.ExitCode > 0 {
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

// sortModulesByDisplayOrder sorts module names using DisplayOrder.
// Falls back to alphabetical if displayOrder is nil or empty.
func sortModulesByDisplayOrder(modules []string, displayOrder *config.DisplayOrder) {
	if displayOrder == nil || len(displayOrder.Modules) == 0 {
		sort.Strings(modules)
		return
	}
	rank := make(map[string]int, len(displayOrder.Modules))
	for i, m := range displayOrder.Modules {
		rank[m] = i
	}
	sort.SliceStable(modules, func(i, j int) bool {
		ri, oki := rank[modules[i]]
		rj, okj := rank[modules[j]]
		if oki && okj {
			return ri < rj
		}
		if oki != okj {
			return oki // known modules before unknown
		}
		return modules[i] < modules[j]
	})
}

// hasRootFailure returns true if the module has at least one component that
// actually ran and failed (has a log path), as opposed to only having
// dependency-skipped failures (no log path).
func hasRootFailure(comps []orchestrator.UnitResult) bool {
	for _, comp := range comps {
		if comp.ExitCode > 0 && comp.LogPath != "" {
			return true
		}
	}
	return false
}

