package cmdframework

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
)

// ComponentWorkProvider is a function that converts execution context to component work items.
// This is provided by the build package to avoid import cycles.
type ComponentWorkProvider func(ctx *ExecutionContext) [][]orchestrator.ComponentWork

// ExecutionMode defines how components should be executed.
type ExecutionMode int

const (
	// ExecutionModeConfigured uses ctx.Config.Layered to decide execution mode.
	ExecutionModeConfigured ExecutionMode = iota
	// ExecutionModeLayered always uses layered execution (respects inter-layer dependencies).
	ExecutionModeLayered
	// ExecutionModeParallel always uses parallel execution (no dependency order).
	ExecutionModeParallel
)

// executionModeConfig holds per-command-type execution settings.
var executionModeConfig = map[CommandType]ExecutionMode{
	CommandTypeBuild: ExecutionModeConfigured, // Respects ctx.Config.Layered
	CommandTypeTest:  ExecutionModeLayered,    // Always layered (parallel first, sequential second)
	CommandTypeScan:  ExecutionModeParallel,   // No dependency order needed
	CommandTypeLint:  ExecutionModeParallel,   // No dependency order needed
}

// ComponentRegistry holds providers and workers for each command type.
// It provides thread-safe access to component execution functions.
type ComponentRegistry struct {
	mu        sync.RWMutex
	providers map[CommandType]ComponentWorkProvider
	workers   map[CommandType]ComponentWorkerFunc
}

// NewComponentRegistry creates a new component registry.
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		providers: make(map[CommandType]ComponentWorkProvider),
		workers:   make(map[CommandType]ComponentWorkerFunc),
	}
}

// RegisterProvider registers a work provider for a command type.
func (r *ComponentRegistry) RegisterProvider(cmdType CommandType, provider ComponentWorkProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[cmdType] = provider
}

// RegisterWorker registers a worker function for a command type.
func (r *ComponentRegistry) RegisterWorker(cmdType CommandType, worker ComponentWorkerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workers[cmdType] = worker
}

// GetProvider returns the registered provider for a command type.
func (r *ComponentRegistry) GetProvider(cmdType CommandType) ComponentWorkProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[cmdType]
}

// GetWorker returns the registered worker for a command type.
func (r *ComponentRegistry) GetWorker(cmdType CommandType) ComponentWorkerFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.workers[cmdType]
}

// HasComponents returns true if both provider and worker are registered.
func (r *ComponentRegistry) HasComponents(cmdType CommandType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[cmdType] != nil && r.workers[cmdType] != nil
}

// registry is the global component registry.
var registry = NewComponentRegistry()

// RegisterComponentProvider registers a work provider for a command type.
func RegisterComponentProvider(cmdType CommandType, provider ComponentWorkProvider) {
	registry.RegisterProvider(cmdType, provider)
}

// RegisterComponentWorker registers a worker function for a command type.
func RegisterComponentWorker(cmdType CommandType, worker ComponentWorkerFunc) {
	registry.RegisterWorker(cmdType, worker)
}

// GetComponentProvider returns the registered provider for a command type.
func GetComponentProvider(cmdType CommandType) ComponentWorkProvider {
	return registry.GetProvider(cmdType)
}

// GetComponentWorker returns the registered worker for a command type.
func GetComponentWorker(cmdType CommandType) ComponentWorkerFunc {
	return registry.GetWorker(cmdType)
}

// HasComponentExecution returns true if component-level execution is available for a command type.
func HasComponentExecution(cmdType CommandType) bool {
	return registry.HasComponents(cmdType)
}


// phaseExecute handles the execution phase using component-level execution.
// The worker parameter is retained for API compatibility but is not used;
// component execution uses workers registered via RegisterComponentWorker.
func phaseExecute(ctx *ExecutionContext, _ WorkerFunc) error {
	// Early return if nothing to execute
	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// All command types use component-level execution
	cmdType := ctx.Config.Type
	if !HasComponentExecution(cmdType) {
		return fmt.Errorf("no component execution registered for command type: %s", cmdType)
	}

	return phaseExecuteComponentsUnified(ctx, cmdType)
}

// phaseExecuteComponentsUnified handles component-level execution for all command types.
// It uses the registered provider and worker for the given command type, and respects
// the execution mode configuration (layered vs parallel).
func phaseExecuteComponentsUnified(ctx *ExecutionContext, cmdType CommandType) error {
	provider := GetComponentProvider(cmdType)
	worker := GetComponentWorker(cmdType)

	if provider == nil || worker == nil {
		return fmt.Errorf("component provider or worker not registered for %s", cmdType)
	}

	// Get component work items from provider
	componentLayers := provider(ctx)
	if len(componentLayers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// Create incremental summary builder with component counts per module
	componentCounts := computeComponentCounts(componentLayers)
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

	// Wrap worker to match orchestrator signature
	orchWorker := func(module, component string, logWriter io.Writer) int {
		return worker(ctx, module, component, logWriter)
	}

	// Determine execution mode from config
	mode := executionModeConfig[cmdType]
	useLayered := mode == ExecutionModeLayered ||
		(mode == ExecutionModeConfigured && ctx.Config.Layered)

	log.Debugf("phaseExecuteComponentsUnified: cmdType=%s, mode=%d, useLayered=%v",
		cmdType, mode, useLayered)

	// Execute components
	var results []orchestrator.WorkResult
	var err error

	execStart := time.Now()
	if useLayered {
		log.Debugf("Executing %d %s component layers with weighted parallelism",
			len(componentLayers), cmdType)
		results, err = ctx.Orchestrator.RunComponentsLayered(componentLayers, orchWorker)
	} else {
		allWork := flattenComponentLayers(componentLayers)
		log.Debugf("Executing %d %s components in parallel with weighted scheduling",
			len(allWork), cmdType)
		results, err = ctx.Orchestrator.RunComponentsParallel(allWork, orchWorker)
	}
	log.Debugf("phaseExecuteComponentsUnified: execution returned in %v", time.Since(execStart))

	if err != nil {
		return fmt.Errorf("%s component execution failed: %w", cmdType, err)
	}

	ctx.Results = results

	// Populate component-level results for detailed reporting
	// Use SummaryBuilder's pre-aggregated results when available
	if ctx.SummaryBuilder != nil {
		ctx.ComponentResultSets = ctx.SummaryBuilder.GetResultSets()
		// Flatten for ComponentResults
		var allResults []orchestrator.ComponentResult
		for _, rs := range ctx.ComponentResultSets {
			allResults = append(allResults, rs.Components...)
		}
		ctx.ComponentResults = allResults
	} else {
		ctx.ComponentResults = ctx.Orchestrator.GetLastComponentResults()
		ctx.ComponentResultSets = orchestrator.AggregateToComponentResultSets(ctx.ComponentResults)
	}

	return nil
}

// computeComponentCounts computes the number of components per module from work layers.
func computeComponentCounts(layers [][]orchestrator.ComponentWork) map[string]int {
	counts := make(map[string]int)
	for _, layer := range layers {
		for _, work := range layer {
			counts[work.Module]++
		}
	}
	return counts
}

// flattenComponentLayers flattens component work layers to a single slice.
func flattenComponentLayers(layers [][]orchestrator.ComponentWork) []orchestrator.ComponentWork {
	var all []orchestrator.ComponentWork
	for _, layer := range layers {
		all = append(all, layer...)
	}
	return all
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
