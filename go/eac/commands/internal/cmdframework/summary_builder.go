package cmdframework

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
	"github.com/ready-to-release/eac/go/eac/commands/internal/output"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
)

// SummaryBuilder incrementally builds summary data as components complete.
// It runs in parallel with execution, computing aggregates on-the-fly.
type SummaryBuilder struct {
	mu sync.Mutex

	// Incremental state (updated per result)
	moduleCaches map[string]*incrementalModuleCache // module -> pre-computed data
	successCount int                                // Modules with all components succeeded
	failureCount int                                // Modules with any component failed
	skippedCount int                                // Modules with all components skipped (cached)

	// Tracking for status derivation
	moduleCompCount map[string]int // module -> expected component count
	moduleCompDone  map[string]int // module -> components completed so far

	// Global completion tracking
	totalModules     int  // Total number of modules expected
	completedModules int  // Number of modules that have completed
	allComplete      bool // True when all modules have completed

	// Completion callback
	onComplete func(*SummaryBuilder) // Called when all modules have completed

	// Command context
	commandType CommandType
}

// incrementalModuleCache holds pre-computed data for a module, updated incrementally.
type incrementalModuleCache struct {
	components     []orchestrator.ComponentResult // All component results for this module
	moduleDuration time.Duration                  // Max duration across components
	errorCount     int
	warnCount      int
	testsTotal     int
	testsPassed    int
	testsFailed    int

	// Status tracking
	hasFailure bool // Any component failed (exit > 0)
	allSkipped bool // All components so far are skipped (exit == -1)
}

// NewSummaryBuilder creates a new incremental summary builder.
// expectedModules is the list of modules that will have results.
// componentCounts maps module -> expected number of components.
func NewSummaryBuilder(cmdType CommandType, componentCounts map[string]int) *SummaryBuilder {
	sb := &SummaryBuilder{
		moduleCaches:    make(map[string]*incrementalModuleCache),
		moduleCompCount: make(map[string]int),
		moduleCompDone:  make(map[string]int),
		totalModules:    len(componentCounts),
		commandType:     cmdType,
	}

	// Initialize module caches with expected component counts
	for module, count := range componentCounts {
		sb.moduleCaches[module] = &incrementalModuleCache{
			components: make([]orchestrator.ComponentResult, 0, count),
			allSkipped: true, // Assume skipped until we see a non-skipped component
		}
		sb.moduleCompCount[module] = count
	}

	return sb
}

// SetOnComplete sets a callback that will be called when all modules have completed.
// The callback receives the builder so it can call Finalize().
// This allows immediate summary send without waiting for execution phase to fully unwind.
func (sb *SummaryBuilder) SetOnComplete(callback func(*SummaryBuilder)) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.onComplete = callback
}

// AddResult adds a component result to the builder.
// This is called by the scheduler as each component completes.
// Thread-safe: can be called from multiple goroutines.
func (sb *SummaryBuilder) AddResult(result orchestrator.ComponentResult) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Get or create module cache
	cache, exists := sb.moduleCaches[result.Module]
	if !exists {
		cache = &incrementalModuleCache{
			components: make([]orchestrator.ComponentResult, 0, 4),
			allSkipped: true,
		}
		sb.moduleCaches[result.Module] = cache
	}

	// Update cache incrementally
	cache.components = append(cache.components, result)

	// Track max duration
	if result.Duration > cache.moduleDuration {
		cache.moduleDuration = result.Duration
	}

	// Aggregate counts
	cache.errorCount += len(result.Errors)
	cache.warnCount += len(result.Warnings)
	cache.testsTotal += result.TestsTotal
	cache.testsPassed += result.TestsPassed
	cache.testsFailed += result.TestsFailed

	// Track status
	if result.ExitCode > 0 {
		cache.hasFailure = true
	}
	if result.ExitCode != -1 {
		cache.allSkipped = false
	}

	// Track module completion
	sb.moduleCompDone[result.Module]++

	// Check if module is now complete and update counts
	if sb.moduleCompDone[result.Module] >= sb.moduleCompCount[result.Module] {
		// Module is complete, update running totals
		if cache.hasFailure {
			sb.failureCount++
		} else if cache.allSkipped && len(cache.components) > 0 {
			sb.skippedCount++
		} else {
			sb.successCount++
		}

		// Track global completion
		sb.completedModules++
		if sb.completedModules >= sb.totalModules {
			sb.allComplete = true

			// Call completion callback if set
			if sb.onComplete != nil {
				sb.onComplete(sb)
			}
		}
	}
}

// Finalize computes the final summary data.
// Called once after all components have completed.
// Returns the summary data ready for TUI display.
func (sb *SummaryBuilder) Finalize(totalTime time.Duration) *tui.SummaryData {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Get sorted module names for consistent output
	modules := make([]string, 0, len(sb.moduleCaches))
	for module := range sb.moduleCaches {
		modules = append(modules, module)
	}
	sort.Strings(modules)

	// Build run summary line
	runSummary := sb.buildRunSummary()

	// Build table based on command type
	var details []string
	var tb *render.TableBuilder

	switch sb.commandType {
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

	// Build rows for each module
	for _, module := range modules {
		cache := sb.moduleCaches[module]

		// Sort components within module (small sort, already accumulated)
		sort.Slice(cache.components, func(i, j int) bool {
			return cache.components[i].Component < cache.components[j].Component
		})

		// Derive status
		status := sb.deriveModuleStatus(cache)

		// Build component names string
		components := sb.buildComponentString(cache.components)

		// Status icon
		statusIcon := " ✓"
		if status == orchestrator.ModuleStatusFailed {
			statusIcon = " ✗"
		} else if cache.warnCount > 0 {
			statusIcon = " ⚠"
		}

		moduleName := output.PackageDisplayName(module)
		duration := formatDuration(cache.moduleDuration)

		// Add row based on command type
		switch sb.commandType {
		case CommandTypeTest:
			testTypes := sb.extractUniqueTestTypes(cache.components)
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
	details = sb.appendFailureDetails(details, modules)

	// Create summary data
	data := &tui.SummaryData{
		Success:    sb.failureCount == 0,
		TotalTime:  totalTime,
		RunSummary: runSummary,
		Details:    details,
	}

	return data
}

// buildRunSummary creates the run summary line.
func (sb *SummaryBuilder) buildRunSummary() string {
	switch {
	case sb.skippedCount > 0 && sb.successCount > 0 && sb.failureCount > 0:
		return fmt.Sprintf("%d cached, %d built, %d failed", sb.skippedCount, sb.successCount, sb.failureCount)
	case sb.skippedCount > 0 && sb.successCount > 0:
		return fmt.Sprintf("%d cached, %d built", sb.skippedCount, sb.successCount)
	case sb.skippedCount > 0 && sb.failureCount > 0:
		return fmt.Sprintf("%d cached, %d failed", sb.skippedCount, sb.failureCount)
	case sb.successCount > 0 && sb.failureCount > 0:
		return fmt.Sprintf("%d built, %d failed", sb.successCount, sb.failureCount)
	case sb.skippedCount > 0:
		return fmt.Sprintf("%d cached", sb.skippedCount)
	case sb.successCount > 0:
		return fmt.Sprintf("%d built", sb.successCount)
	case sb.failureCount > 0:
		return fmt.Sprintf("%d failed", sb.failureCount)
	default:
		return "0 modules"
	}
}

// deriveModuleStatus computes module status from cached data.
func (sb *SummaryBuilder) deriveModuleStatus(cache *incrementalModuleCache) orchestrator.ModuleStatus {
	if cache.hasFailure {
		return orchestrator.ModuleStatusFailed
	}
	if cache.allSkipped && len(cache.components) > 0 {
		return orchestrator.ModuleStatusSkipped
	}
	return orchestrator.ModuleStatusSuccess
}

// buildComponentString creates the component names string for display.
func (sb *SummaryBuilder) buildComponentString(components []orchestrator.ComponentResult) string {
	if len(components) == 0 {
		return ""
	}

	if len(components) <= 3 {
		// Fast path: few components, direct concatenation
		compNames := make([]string, len(components))
		for i, comp := range components {
			compNames[i] = comp.Component
		}
		result := strings.Join(compNames, ", ")
		if len(result) > 60 {
			return result[:57] + "..."
		}
		return result
	}

	// Use strings.Builder for many components
	var sb2 strings.Builder
	for i, comp := range components {
		if i > 0 {
			sb2.WriteString(", ")
		}
		if sb2.Len()+len(comp.Component) > 57 {
			sb2.WriteString("...")
			break
		}
		sb2.WriteString(comp.Component)
	}

	result := sb2.String()
	if len(result) > 60 {
		return result[:57] + "..."
	}
	return result
}

// extractUniqueTestTypes extracts unique test types from component names.
func (sb *SummaryBuilder) extractUniqueTestTypes(components []orchestrator.ComponentResult) string {
	if len(components) == 0 {
		return "-"
	}

	seen := make(map[string]struct{}, 4)
	types := make([]string, 0, 4)

	for _, comp := range components {
		testType := comp.Component
		if colonIdx := strings.LastIndex(comp.Component, ":"); colonIdx >= 0 {
			testType = comp.Component[colonIdx+1:]
		}

		if _, exists := seen[testType]; !exists {
			seen[testType] = struct{}{}
			types = append(types, testType)
		}
	}

	sort.Strings(types)
	return strings.Join(types, ", ")
}

// appendFailureDetails adds failed/warning module details to the details slice.
func (sb *SummaryBuilder) appendFailureDetails(details []string, modules []string) []string {
	// Count failures/warnings
	hasFailures := false
	totalFailures := 0
	for _, module := range modules {
		cache := sb.moduleCaches[module]
		if cache.hasFailure || cache.warnCount > 0 {
			hasFailures = true
			totalFailures++
		}
	}

	if !hasFailures {
		return details
	}

	details = append(details, "")
	failCount := 0
	const maxFailures = 5

	for _, module := range modules {
		if failCount >= maxFailures {
			break
		}
		cache := sb.moduleCaches[module]

		if !cache.hasFailure && cache.warnCount == 0 {
			continue
		}
		failCount++

		// Module header with status
		statusIcon := "✗"
		if !cache.hasFailure {
			statusIcon = "⚠"
		}
		moduleName := output.PackageDisplayName(module)
		details = append(details, fmt.Sprintf("%s %s", statusIcon, moduleName))

		// Show first error/warning from failed components
		for _, comp := range cache.components {
			if comp.ExitCode == 0 && len(comp.Warnings) == 0 {
				continue
			}

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

			if comp.LogPath != "" {
				details = append(details, fmt.Sprintf("    Log: %s", comp.LogPath))
			}
			break // Only show first failed component per module
		}
	}

	if failCount >= maxFailures && totalFailures > maxFailures {
		remaining := totalFailures - maxFailures
		details = append(details, fmt.Sprintf("  ... and %d more failures", remaining))
	}

	return details
}

// GetResultSets returns the component results grouped by module.
// This can be used to populate ComponentResultSets without re-aggregating.
func (sb *SummaryBuilder) GetResultSets() []orchestrator.ComponentResultSet {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Get sorted module names
	modules := make([]string, 0, len(sb.moduleCaches))
	for module := range sb.moduleCaches {
		modules = append(modules, module)
	}
	sort.Strings(modules)

	// Build result sets
	resultSets := make([]orchestrator.ComponentResultSet, 0, len(modules))
	for _, module := range modules {
		cache := sb.moduleCaches[module]

		// Sort components
		sortedComponents := make([]orchestrator.ComponentResult, len(cache.components))
		copy(sortedComponents, cache.components)
		sort.Slice(sortedComponents, func(i, j int) bool {
			return sortedComponents[i].Component < sortedComponents[j].Component
		})

		rs := orchestrator.ComponentResultSet{
			Module:     module,
			Components: sortedComponents,
			Duration:   cache.moduleDuration,
		}
		rs.Status = rs.DeriveStatus()

		resultSets = append(resultSets, rs)
	}

	return resultSets
}
