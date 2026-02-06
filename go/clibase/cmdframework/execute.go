package cmdframework

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/clibase/orchestrator"
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
	providers map[CommandType]UnitWorkProvider
	workers   map[CommandType]UnitWorkerFunc
}

// NewUnitRegistry creates a new component registry.
func NewUnitRegistry() *UnitRegistry {
	return &UnitRegistry{
		providers: make(map[CommandType]UnitWorkProvider),
		workers:   make(map[CommandType]UnitWorkerFunc),
	}
}

// RegisterProvider registers a work provider for a command type.
func (r *UnitRegistry) RegisterProvider(cmdType CommandType, provider UnitWorkProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[cmdType] = provider
}

// RegisterWorker registers a worker function for a command type.
func (r *UnitRegistry) RegisterWorker(cmdType CommandType, worker UnitWorkerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workers[cmdType] = worker
}

// GetProvider returns the registered provider for a command type.
func (r *UnitRegistry) GetProvider(cmdType CommandType) UnitWorkProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[cmdType]
}

// GetWorker returns the registered worker for a command type.
func (r *UnitRegistry) GetWorker(cmdType CommandType) UnitWorkerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.workers[cmdType]
}

// HasComponents returns true if both provider and worker are registered.
func (r *UnitRegistry) HasComponents(cmdType CommandType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[cmdType] != nil && r.workers[cmdType] != nil
}

// registry is the global component registry.
var registry = NewUnitRegistry()

// RegisterUnitProvider registers a work provider for a command type.
func RegisterUnitProvider(cmdType CommandType, provider UnitWorkProvider) {
	registry.RegisterProvider(cmdType, provider)
}

// RegisterUnitWorker registers a worker function for a command type.
func RegisterUnitWorker(cmdType CommandType, worker UnitWorkerFunc) {
	registry.RegisterWorker(cmdType, worker)
}

// GetUnitProvider returns the registered provider for a command type.
func GetUnitProvider(cmdType CommandType) UnitWorkProvider {
	return registry.GetProvider(cmdType)
}

// GetUnitWorker returns the registered worker for a command type.
func GetUnitWorker(cmdType CommandType) UnitWorkerFunc {
	return registry.GetWorker(cmdType)
}

// HasUnitExecution returns true if component-level execution is available for a command type.
func HasUnitExecution(cmdType CommandType) bool {
	return registry.HasComponents(cmdType)
}


// phaseExecute handles the execution phase using component-level execution.
// The worker parameter is retained for API compatibility but is not used;
// component execution uses workers registered via RegisterUnitWorker.
func phaseExecute(ctx *ExecutionContext, _ WorkerFunc) error {
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
func phaseExecuteComponentsUnified(ctx *ExecutionContext, cmdType CommandType) error {
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

	// Create incremental summary builder with component counts per module
	componentCounts := computeComponentCounts(allWork)
	ctx.SummaryBuilder = NewSummaryBuilder(cmdType, componentCounts)
	ctx.SummaryBuilder.SetStartTime(ctx.StartTime)
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
	orchWorker := func(goCtx context.Context, module, component string, logWriter io.Writer) int {
		return worker(goCtx, ctx, module, component, logWriter)
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
// are added for ALL UoWs in the dependency module. This ensures:
//  1. UoWs wait until all dependency module UoWs complete
//  2. If any dependency module UoW fails, dependent UoWs see it via HasFailedDependency
func injectModuleDependencies(work []workunit.UnitSpec, registry *modules.Registry) []workunit.UnitSpec {
	if registry == nil || len(work) == 0 {
		return work
	}

	// Build module -> UnitIDs index (which UoWs exist for each module)
	moduleUnitIDs := make(map[string][]workunit.UnitID)
	for _, w := range work {
		moduleUnitIDs[w.ID.Module] = append(moduleUnitIDs[w.ID.Module], w.ID)
	}

	// For each UoW, add DependsOn entries for all UoWs in dependency modules
	for i := range work {
		module, exists := registry.Get(work[i].ID.Module)
		if !exists {
			continue
		}

		for _, depModule := range module.GetDependencies() {
			depUnitIDs, hasDeps := moduleUnitIDs[depModule]
			if !hasDeps {
				continue // Dependency module not in current execution set
			}
			work[i].DependsOn = append(work[i].DependsOn, depUnitIDs...)
		}
	}

	return work
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
