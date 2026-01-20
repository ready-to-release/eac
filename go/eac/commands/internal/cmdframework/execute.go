package cmdframework

import (
	"fmt"
	"io"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
)

// ComponentWorkProvider is a function that converts execution context to component work items.
// This is provided by the build package to avoid import cycles.
type ComponentWorkProvider func(ctx *ExecutionContext) [][]orchestrator.ComponentWork

// componentWorkProvider is set by the build package to provide component flattening.
var componentWorkProvider ComponentWorkProvider

// SetComponentWorkProvider registers the function to flatten modules to components.
// Called by the build package during init.
func SetComponentWorkProvider(provider ComponentWorkProvider) {
	componentWorkProvider = provider
}

// componentWorkerFunc is the registered component worker (set by build package).
var componentWorkerFunc ComponentWorkerFunc

// SetComponentWorker registers the component worker function.
// Called by the build package during init.
func SetComponentWorker(worker ComponentWorkerFunc) {
	componentWorkerFunc = worker
}

// testComponentWorkProvider is set by the test package to provide test component flattening.
var testComponentWorkProvider ComponentWorkProvider

// SetTestComponentWorkProvider registers the function to flatten tests to components.
// Called by the test package during init.
func SetTestComponentWorkProvider(provider ComponentWorkProvider) {
	testComponentWorkProvider = provider
}

// testComponentWorkerFunc is the registered test component worker (set by test package).
var testComponentWorkerFunc ComponentWorkerFunc

// SetTestComponentWorker registers the test component worker function.
// Called by the test package during init.
func SetTestComponentWorker(worker ComponentWorkerFunc) {
	testComponentWorkerFunc = worker
}

// scanComponentWorkProvider is set by the scan package to provide scan component flattening.
var scanComponentWorkProvider ComponentWorkProvider

// SetScanComponentWorkProvider registers the function to flatten modules to scan components.
// Called by the scan package during init.
func SetScanComponentWorkProvider(provider ComponentWorkProvider) {
	scanComponentWorkProvider = provider
}

// scanComponentWorkerFunc is the registered scan component worker (set by scan package).
var scanComponentWorkerFunc ComponentWorkerFunc

// SetScanComponentWorker registers the scan component worker function.
// Called by the scan package during init.
func SetScanComponentWorker(worker ComponentWorkerFunc) {
	scanComponentWorkerFunc = worker
}

// phaseExecute handles the execution phase:
// - Set up worker function
// - Run orchestrator (layered or parallel)
// - Collect results.
func phaseExecute(ctx *ExecutionContext, worker WorkerFunc) error {
	if worker == nil {
		return fmt.Errorf("worker function is required")
	}

	// Early return if nothing to execute
	monikers := ctx.GetExecutionMonikers()
	if len(monikers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// Check if this is a build command with component-level execution enabled
	// Component-level execution uses weighted parallelism and intra-module dependencies
	if ctx.Config.Type == CommandTypeBuild && componentWorkProvider != nil && componentWorkerFunc != nil {
		return phaseExecuteComponents(ctx)
	}

	// Check if this is a test command with component-level execution enabled
	if ctx.Config.Type == CommandTypeTest && testComponentWorkProvider != nil && testComponentWorkerFunc != nil {
		return phaseExecuteTestComponents(ctx)
	}

	// Check if this is a scan command with component-level execution enabled
	if ctx.Config.Type == CommandTypeScan && scanComponentWorkProvider != nil && scanComponentWorkerFunc != nil {
		return phaseExecuteScanComponents(ctx)
	}

	// Fall back to module-level execution (legacy)
	return phaseExecuteModules(ctx, worker)
}

// phaseExecuteComponents runs component-level parallel execution for builds.
// This is the new execution path that provides weighted parallelism.
func phaseExecuteComponents(ctx *ExecutionContext) error {
	// Get component work items from the provider
	componentLayers := componentWorkProvider(ctx)
	if len(componentLayers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// Wrap the component worker to match orchestrator signature
	orchCompWorker := func(module, component string, logWriter io.Writer) int {
		return componentWorkerFunc(ctx, module, component, logWriter)
	}

	var results []orchestrator.WorkResult
	var err error

	if ctx.Config.Layered {
		// Layered execution: inter-module dependency order preserved
		log.Debugf("Executing %d component layers with weighted parallelism", len(componentLayers))
		results, err = ctx.Orchestrator.RunComponentsLayered(componentLayers, orchCompWorker)
	} else {
		// Parallel execution: all components at once (with intra-module deps)
		allWork := flattenComponentLayers(componentLayers)
		log.Debugf("Executing %d components in parallel with weighted scheduling", len(allWork))
		results, err = ctx.Orchestrator.RunComponentsParallel(allWork, orchCompWorker)
	}

	if err != nil {
		return fmt.Errorf("component execution failed: %w", err)
	}

	ctx.Results = results
	return nil
}

// phaseExecuteTestComponents runs component-level parallel execution for tests.
// This uses weighted parallelism and separates parallel/sequential tests into layers.
func phaseExecuteTestComponents(ctx *ExecutionContext) error {
	// Get test component work items from the provider
	componentLayers := testComponentWorkProvider(ctx)
	if len(componentLayers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// Wrap the test component worker to match orchestrator signature
	orchCompWorker := func(module, component string, logWriter io.Writer) int {
		return testComponentWorkerFunc(ctx, module, component, logWriter)
	}

	var results []orchestrator.WorkResult
	var err error

	// Test execution always uses layered mode (parallel tests first, sequential second)
	log.Debugf("Executing %d test component layers with weighted parallelism", len(componentLayers))
	results, err = ctx.Orchestrator.RunComponentsLayered(componentLayers, orchCompWorker)

	if err != nil {
		return fmt.Errorf("test component execution failed: %w", err)
	}

	ctx.Results = results
	return nil
}

// phaseExecuteScanComponents runs component-level parallel execution for scans.
// This uses weighted parallelism for scanning different components in parallel.
func phaseExecuteScanComponents(ctx *ExecutionContext) error {
	// Get scan component work items from the provider
	componentLayers := scanComponentWorkProvider(ctx)
	if len(componentLayers) == 0 {
		ctx.Results = []orchestrator.WorkResult{}
		return nil
	}

	// Wrap the scan component worker to match orchestrator signature
	orchCompWorker := func(module, component string, logWriter io.Writer) int {
		return scanComponentWorkerFunc(ctx, module, component, logWriter)
	}

	// Flatten all layers - scans run in parallel (no dependency order needed)
	allWork := flattenComponentLayers(componentLayers)
	log.Debugf("Executing %d scan components in parallel with weighted scheduling", len(allWork))
	results, err := ctx.Orchestrator.RunComponentsParallel(allWork, orchCompWorker)

	if err != nil {
		return fmt.Errorf("scan component execution failed: %w", err)
	}

	ctx.Results = results
	return nil
}

// phaseExecuteModules runs module-level execution (legacy).
func phaseExecuteModules(ctx *ExecutionContext, worker WorkerFunc) error {
	// Wrap the user's worker to match orchestrator signature
	orchWorker := func(moniker string, logWriter io.Writer) int {
		return worker(ctx, moniker, logWriter)
	}

	// Set worker on orchestrator
	ctx.Orchestrator.SetWorker(orchWorker)

	// Execute based on mode
	var results []orchestrator.WorkResult
	var err error

	if ctx.Config.Layered {
		// Layered execution (build): respect dependency order
		layers := ctx.GetLayers()
		log.Debugf("Executing %d layers with %d total modules",
			len(layers), len(ctx.GetExecutionMonikers()))
		results, err = ctx.Orchestrator.RunLayered(layers)
	} else {
		// Parallel execution (test/scan): all at once
		monikers := ctx.GetExecutionMonikers()
		log.Debugf("Executing %d modules in parallel", len(monikers))
		results, err = ctx.Orchestrator.Run(monikers)
	}

	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	ctx.Results = results
	return nil
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
func (ctx *ExecutionContext) GetFailureCount() int {
	count := 0
	for _, r := range ctx.Results {
		if r.ExitCode != 0 {
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
func (ctx *ExecutionContext) GetFailedMonikers() []string {
	var failed []string
	for _, r := range ctx.Results {
		if r.ExitCode != 0 {
			failed = append(failed, r.Moniker)
		}
	}
	return failed
}
