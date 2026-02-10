package cmdframework

import (
	"context"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/orchestrator"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// UnitWorkProvider is a function that converts execution context to work units.
// This is provided by the build package to avoid import cycles.
// Returns a flat list of units - scheduling order is determined by DependsOn constraints.
type UnitWorkProvider func(ctx *ExecutionContext) []workunit.UnitSpec

// All commands use dependency-based parallel execution.
// Work units are scheduled based on their DependsOn constraints.

// UnitRegistry holds providers and workers for each command type.
// It provides thread-safe access to component execution functions.
type UnitRegistry struct {
	mu        sync.RWMutex
	providers map[core.ActionType]UnitWorkProvider
	workers   map[core.ActionType]UnitWorkerFunc
}

// NewUnitRegistry creates a new component registry.
func NewUnitRegistry() *UnitRegistry {
	return &UnitRegistry{
		providers: make(map[core.ActionType]UnitWorkProvider),
		workers:   make(map[core.ActionType]UnitWorkerFunc),
	}
}

// RegisterProvider registers a work provider for a command type.
func (r *UnitRegistry) RegisterProvider(cmdType core.ActionType, provider UnitWorkProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[cmdType] = provider
}

// RegisterWorker registers a worker function for a command type.
func (r *UnitRegistry) RegisterWorker(cmdType core.ActionType, worker UnitWorkerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workers[cmdType] = worker
}

// GetProvider returns the registered provider for a command type.
func (r *UnitRegistry) GetProvider(cmdType core.ActionType) UnitWorkProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[cmdType]
}

// GetWorker returns the registered worker for a command type.
func (r *UnitRegistry) GetWorker(cmdType core.ActionType) UnitWorkerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.workers[cmdType]
}

// HasComponents returns true if both provider and worker are registered.
func (r *UnitRegistry) HasComponents(cmdType core.ActionType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[cmdType] != nil && r.workers[cmdType] != nil
}

// registry is the global component registry.
var registry = NewUnitRegistry()

// RegisterUnitProvider registers a work provider for a command type.
func RegisterUnitProvider(cmdType core.ActionType, provider UnitWorkProvider) {
	registry.RegisterProvider(cmdType, provider)
}

// RegisterUnitWorker registers a worker function for a command type.
func RegisterUnitWorker(cmdType core.ActionType, worker UnitWorkerFunc) {
	registry.RegisterWorker(cmdType, worker)
}

// GetUnitProvider returns the registered provider for a command type.
func GetUnitProvider(cmdType core.ActionType) UnitWorkProvider {
	return registry.GetProvider(cmdType)
}

// GetUnitWorker returns the registered worker for a command type.
func GetUnitWorker(cmdType core.ActionType) UnitWorkerFunc {
	return registry.GetWorker(cmdType)
}

// HasUnitExecution returns true if component-level execution is available for a command type.
func HasUnitExecution(cmdType core.ActionType) bool {
	return registry.HasComponents(cmdType)
}


// phaseExecute handles the execution phase using component-level execution.
// The worker parameter is retained for API compatibility but is not used;
// component execution uses workers registered via RegisterUnitWorker.
func phaseExecute(ctx *ExecutionContext, _ CommandWorkerFunc) error {
	// Early return if nothing to execute
	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// All command types use component-level execution
	cmdType := ctx.Config.Type
	if !HasUnitExecution(cmdType) {
		return fmt.Errorf("no component execution registered for command type: %s", cmdType)
	}

	return phaseExecuteComponentsUnified(ctx, cmdType)
}

// phaseExecuteComponentsUnified handles component-level execution for all command types.
// It uses the registered provider and worker for the given command type.
// Scheduling order is determined by DependsOn constraints in the work units.
func phaseExecuteComponentsUnified(ctx *ExecutionContext, cmdType core.ActionType) error {
	provider := GetUnitProvider(cmdType)
	worker := GetUnitWorker(cmdType)

	if provider == nil || worker == nil {
		return fmt.Errorf("component provider or worker not registered for %s", cmdType)
	}

	// Get component work items from provider
	allWork := provider(ctx)
	if len(allWork) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// Inject inter-module dependencies into UnitSpec.DependsOn.
	// Module-level depends_on (from repository.yml) is not encoded in UnitSpecs
	// by work providers — this step wires them so the scheduler can enforce
	// inter-module ordering and propagate failures across module boundaries.
	allWork = injectModuleDependencies(allWork, ctx.ModuleRegistry)

	// Sort work by display order for consistent TUI tab ordering.
	// Falls back to alphabetical when display order is unavailable.
	var displayOrder *config.DisplayOrder
	if ctx.EACConfig != nil {
		displayOrder = ctx.EACConfig.Repository.DisplayOrder
	}
	sortWorkByDisplayOrder(allWork, displayOrder)

	// Create incremental summary builder with component counts per module
	componentCounts := computeComponentCounts(allWork)
	ctx.SummaryBuilder = NewSummaryBuilder(cmdType, componentCounts)
	ctx.SummaryBuilder.SetStartTime(ctx.StartTime)
	if ctx.EACConfig != nil && ctx.EACConfig.Repository.DisplayOrder != nil {
		ctx.SummaryBuilder.SetDisplayOrder(ctx.EACConfig.Repository.DisplayOrder)
	}
	ctx.Orchestrator.SetSummaryBuilder(ctx.SummaryBuilder)

	// Set completion callback to send summary immediately when all components finish.
	// This eliminates the delay between TUI showing "rendering summary" and receiving data.
	// The TUI exits immediately on receiving summary; framework waits via WaitTUI().
	if ctx.Config.UseTUI {
		ctx.SummaryBuilder.SetOnComplete(func(sb *SummaryBuilder) {
			callbackStart := time.Now()
			log.Debugf("OnComplete callback: starting")
			totalTime := time.Since(ctx.StartTime)
			summaryData := sb.Finalize(totalTime)
			ctx.Orchestrator.SendSummary(summaryData)
			sb.MarkSummarySent()
			log.Debugf("OnComplete callback: completed in %v", time.Since(callbackStart))
		})
	}

	// Wrap worker to match orchestrator signature, forwarding cancellation context
	orchWorker := func(goCtx context.Context, spec core.UnitSpec, logWriter io.Writer) int {
		return worker(goCtx, ctx, spec, logWriter)
	}

	// Execute components with dependency-based parallel scheduling
	log.Debugf("Executing %d %s components with dependency-based scheduling",
		len(allWork), cmdType)

	execStart := time.Now()
	results, err := ctx.Orchestrator.RunUnitsParallel(allWork, orchWorker)
	log.Debugf("phaseExecuteComponentsUnified: execution returned in %v", time.Since(execStart))

	if err != nil {
		return fmt.Errorf("%s component execution failed: %w", cmdType, err)
	}

	ctx.Results = results

	// Populate component-level results for detailed reporting
	// Use SummaryBuilder's pre-aggregated results when available
	if ctx.SummaryBuilder != nil {
		ctx.ModuleResultSets = ctx.SummaryBuilder.GetResultSets()
		// Flatten for ComponentResults
		var allResults []orchestrator.UnitResult
		for _, rs := range ctx.ModuleResultSets {
			allResults = append(allResults, rs.Units...)
		}
		ctx.UnitResults = allResults
	} else {
		ctx.UnitResults = ctx.Orchestrator.GetLastUnitResults()
		ctx.ModuleResultSets = orchestrator.AggregateToModuleResultSets(ctx.UnitResults)
	}

	return nil
}

// sortWorkByDisplayOrder sorts work units by display order for consistent TUI tab ordering.
// Uses module ordering from DisplayOrder.Modules, then component ordering from DisplayOrder.Components.
func sortWorkByDisplayOrder(work []workunit.UnitSpec, displayOrder *config.DisplayOrder) {
	// Build module rank map (empty if no display order)
	moduleRank := make(map[string]int)
	if displayOrder != nil {
		for i, m := range displayOrder.Modules {
			moduleRank[m] = i
		}
	}

	// Build component rank maps per module (empty if no display order)
	compRank := make(map[string]map[string]int)
	if displayOrder != nil {
		for mod, comps := range displayOrder.Components {
			rank := make(map[string]int, len(comps))
			for i, c := range comps {
				rank[c] = i
			}
			compRank[mod] = rank
		}
	}

	sort.SliceStable(work, func(i, j int) bool {
		mi, mj := work[i].ID.Module, work[j].ID.Module
		if mi != mj {
			ri, oki := moduleRank[mi]
			rj, okj := moduleRank[mj]
			if oki && okj {
				return ri < rj
			}
			if oki != okj {
				return oki
			}
			return mi < mj
		}

		// Same module — sort by component order, fallback alphabetical
		ci, cj := work[i].ID.ComponentName, work[j].ID.ComponentName
		if ci != cj {
			if cr, ok := compRank[mi]; ok {
				ri, oki := cr[ci]
				rj, okj := cr[cj]
				if oki && okj {
					return ri < rj
				}
				if oki != okj {
					return oki
				}
			}
			return ci < cj
		}

		// Same component — sort by tool name for stability
		return work[i].ID.Tool < work[j].ID.Tool
	})
}

// computeComponentCounts computes the number of components per module.
func computeComponentCounts(units []workunit.UnitSpec) map[string]int {
	counts := make(map[string]int)
	for _, work := range units {
		counts[work.ID.Module]++
	}
	return counts
}

// injectModuleDependencies adds inter-module dependency edges to UnitSpecs.
// For each UoW in a module that depends on another module, DependsOn entries
// are added based on how the dependency is declared:
//
//   - component_deps (narrowed): Only matching UoWs from the dep module are added.
//     Matching uses ParsedComponentDep.MatchesUnitID for progressive field matching.
//   - depends_on only (legacy all-to-all): ALL UoWs in the dep module are added.
//
// Narrowing is per source-module → dep-module pair. If ANY component in module A
// has component_deps referencing module B, then ALL components in module A use
// narrowed matching for B. Components without their own component_deps for B
// simply get no deps from B.
func injectModuleDependencies(work []workunit.UnitSpec, registry *modules.Registry) []workunit.UnitSpec {
	if registry == nil || len(work) == 0 {
		return work
	}

	// Build module -> UnitIDs index (which UoWs exist for each module)
	moduleUnitIDs := make(map[string][]workunit.UnitID)
	for _, w := range work {
		moduleUnitIDs[w.ID.Module] = append(moduleUnitIDs[w.ID.Module], w.ID)
	}

	// Pre-build component_deps index:
	// narrowDeps[srcModule][depModule] = true means srcModule uses narrowed matching for depModule
	// parsedDeps[srcModule][componentName] = []ParsedComponentDep for that component
	narrowDeps, parsedDeps := buildComponentDepsIndex(work, registry)

	// For each UoW, add DependsOn entries for dependency modules
	for i := range work {
		srcModule := work[i].ID.Module
		module, exists := registry.Get(srcModule)
		if !exists {
			continue
		}

		for _, depModule := range module.GetDependencies() {
			depUnitIDs, hasDeps := moduleUnitIDs[depModule]
			if !hasDeps {
				continue // Dependency module not in current execution set
			}

			if narrowDeps[srcModule][depModule] {
				// Narrowed: only add UoWs matching this component's parsed deps
				compName := work[i].ID.ComponentName
				for _, parsed := range parsedDeps[srcModule][compName] {
					if parsed.Module != depModule {
						continue
					}
					for _, depUID := range depUnitIDs {
						if parsed.MatchesUnitID(depUID) {
							work[i].DependsOn = append(work[i].DependsOn, depUID)
						}
					}
				}
			} else {
				// Legacy all-to-all
				work[i].DependsOn = append(work[i].DependsOn, depUnitIDs...)
			}
		}
	}

	return work
}

// buildComponentDepsIndex pre-parses component_deps from all modules in the work set.
// Returns:
//   - narrowDeps: map[srcModule][depModule] = true if narrowed matching applies
//   - parsedDeps: map[srcModule][componentName] = parsed deps for that component
func buildComponentDepsIndex(work []workunit.UnitSpec, registry *modules.Registry) (
	map[string]map[string]bool,
	map[string]map[string][]config.ParsedComponentDep,
) {
	narrowDeps := make(map[string]map[string]bool)
	parsedDeps := make(map[string]map[string][]config.ParsedComponentDep)

	// Collect unique source modules from work items
	seenModules := make(map[string]bool)
	for _, w := range work {
		seenModules[w.ID.Module] = true
	}

	for srcModule := range seenModules {
		module, exists := registry.Get(srcModule)
		if !exists {
			continue
		}

		for compName, entry := range module.Components {
			if entry == nil || len(entry.ComponentDeps) == 0 {
				continue
			}

			// Lazy-init maps for this source module
			if narrowDeps[srcModule] == nil {
				narrowDeps[srcModule] = make(map[string]bool)
			}
			if parsedDeps[srcModule] == nil {
				parsedDeps[srcModule] = make(map[string][]config.ParsedComponentDep)
			}

			for _, dep := range entry.ComponentDeps {
				parsed, err := config.ParseComponentDep(dep)
				if err != nil {
					continue // Validation errors caught at config load time
				}
				narrowDeps[srcModule][parsed.Module] = true
				parsedDeps[srcModule][compName] = append(parsedDeps[srcModule][compName], parsed)
			}
		}
	}

	return narrowDeps, parsedDeps
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
// Only counts actual failures (ExitCode > 0), not skipped (-1).
func (ctx *ExecutionContext) GetFailureCount() int {
	count := 0
	for _, r := range ctx.Results {
		if r.ExitCode > 0 {
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
// Only returns actual failures (ExitCode > 0), not skipped (-1).
func (ctx *ExecutionContext) GetFailedMonikers() []string {
	var failed []string
	for _, r := range ctx.Results {
		if r.ExitCode > 0 {
			failed = append(failed, r.Moniker)
		}
	}
	return failed
}
