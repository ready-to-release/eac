package cmdframework

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/clibase/render"
	"github.com/ready-to-release/eac/go/core/config"
)

const (
	maxErrorLength = 500
	maxLineLength  = 100
)

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
