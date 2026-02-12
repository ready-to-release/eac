package orchestrator

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
	"github.com/ready-to-release/eac/go/clibase/locktracker"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// Orchestrator manages parallel execution of work items.
type Orchestrator struct {
	config          Config
	worker          WorkerFunc
	display         *displayManager
	orchestratorOut io.Writer
	logger          *log.Logger // goroutine-safe logger

	// Lock tracking registry for visualization
	registry *locktracker.Registry

	// Current component scheduler (set during RunUnitsParallel)
	currentScheduler *UnitScheduler

	// TUI console for real-time output display
	tuiConsole display.Console
	tuiCtx     context.Context
	tuiCancel  context.CancelFunc

	// TUI status tracking (protected by tuiMu)
	tuiMu        sync.Mutex
	tuiRunning   []string
	tuiCompleted int
	tuiTotal     int
	tuiStarted   bool // True once StartTUI has been called

	// Component results from last execution (for detailed reporting)
	lastUnitResults []UnitResult

	// Pending cache times to apply when scheduler is created
	pendingCacheTimes map[string]time.Time

	// Pending cache detection config to apply when scheduler is created
	pendingCacheVerifier      CacheVerifier
	pendingCachedModules      map[string]bool

	// Summary builder for incremental summary computation
	summaryBuilder SummaryBuilder

	// initSummaryEmitted tracks whether SetInitSummary was called.
	// Used to tell UnitScheduler to skip redundant UnitQueuedEvent emissions.
	initSummaryEmitted bool

	// Observer pattern fields
	observers     []core.ExecutionObserver
	observersMu   sync.RWMutex
	writerFactory core.WriterFactory
}

// New creates a new Orchestrator with the given configuration and worker function.
func New(config *Config, worker WorkerFunc) *Orchestrator {
	// MaxConcurrency=0 means dynamic (calculated from CPU×RAM).
	// Only use MaxConcurrency as ceiling if explicitly set by user.

	// Set default TUI height
	if config.TUIHeight <= 0 {
		config.TUIHeight = display.DefaultHeight
	}

	// Use global lock tracking registry (shared with component scheduler and other components)
	registry := locktracker.Get()

	o := &Orchestrator{
		config:   *config,
		worker:   worker,
		registry: registry,
	}

	// Note: TUI console is no longer created here.
	// Use AddObserver() to register a TUIObserver for TUI display.
	// The tuiConsole field is kept for backward compatibility with SetConsole().

	return o
}

// Init initializes the orchestrator's output infrastructure.
// This must be called before StartTUI or using phase methods.
// It's automatically called by Run/RunLayered if not called explicitly.
func (o *Orchestrator) Init() error {
	if o.orchestratorOut != nil {
		return nil // Already initialized
	}

	// Configure output writer based on TUI mode
	// When TUI is enabled, discard orchestrator-level output (TUI handles display)
	// When TUI is disabled, write to stdout
	var writer io.Writer
	if o.config.TUI {
		writer = io.Discard // TUI mode - TUI handles console display
	} else {
		writer = os.Stdout
	}

	o.logger = log.New(writer, "", 0)
	o.orchestratorOut = writer

	return nil
}

// Close releases resources held by the orchestrator.
func (o *Orchestrator) Close() {
	// Stop TUI if not already stopped
	o.StopTUI()

	// Note: registry is global, don't close it here
	// Component scheduler manages its own semaphore lifecycle
}

// RunUnitsParallel executes all components in parallel with weighted scheduling.
// Components respect intra-module dependencies (DependsOn).
//
// Returns WorkResult aggregated at module level for compatibility with existing code.
func (o *Orchestrator) RunUnitsParallel(work []workunit.UnitSpec, worker UnitWorkerFunc) ([]WorkResult, error) {
	if len(work) == 0 {
		return []WorkResult{}, nil
	}

	// Initialize if not already done
	if err := o.Init(); err != nil {
		return nil, err
	}

	// Create component scheduler with dynamic capacity management
	scheduler := NewUnitScheduler(&o.config, o.tuiConsole, o.registry, o.emit, o.writerFactory)
	o.currentScheduler = scheduler

	// Apply pending cache times if any
	if o.pendingCacheTimes != nil {
		scheduler.SetCacheTimes(o.pendingCacheTimes)
		o.pendingCacheTimes = nil
	}

	// Apply pending cache detection if any
	if o.pendingCacheVerifier != nil {
		scheduler.SetCacheDetection(o.pendingCacheVerifier, o.pendingCachedModules)
		o.pendingCacheVerifier = nil
		o.pendingCachedModules = nil
	}

	// Apply summary builder if set
	if o.summaryBuilder != nil {
		scheduler.SetSummaryBuilder(o.summaryBuilder)
	}

	// Skip redundant UnitQueuedEvent if InitSummary already registered all tabs
	if o.initSummaryEmitted {
		scheduler.SetInitSummaryEmitted()
	}

	defer func() {
		scheduler.Close()
		o.currentScheduler = nil
	}()

	// Initialize tracking state
	o.tuiMu.Lock()
	o.tuiTotal = len(work)
	o.tuiCompleted = 0
	o.tuiRunning = nil
	o.tuiMu.Unlock()

	// Start TUI console if enabled (lifecycle management)
	if o.tuiConsole != nil {
		o.tuiConsole.StartAsync(o.tuiCtx)
	}

	// Emit initial progress event to all observers
	o.emit(core.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   nil,
		Completed: 0,
		Total:     len(work),
	})

	// Print header
	uniqueModules := countUniqueModules(work)
	fmt.Fprintf(o.orchestratorOut, "%s %d components across %d modules in parallel%s%s",
		capitalize(o.config.ActionVerb()), len(work), uniqueModules, LineEnding, LineEnding)

	// Create and start display manager (only when TUI is not enabled)
	if !o.config.TUI {
		o.display = newDisplayManager(o.logger, o.config.ActionVerb(), len(work), o.config.StatusUpdateInterval, false, o.registry)
		o.display.start()
	}

	// Set phase to Run
	o.SetPhase(display.PhaseRun)

	// Set indices
	for i := range work {
		work[i].Index = i
	}

	// Initialize scheduler
	scheduler.InitializeWork(work)

	// Execute all components
	results := scheduler.RunUnits(work, worker)

	// Stop display manager
	if o.display != nil {
		o.display.stop()
		o.display.flushCompletedLines()
	}

	// Store component results for detailed reporting
	o.lastUnitResults = results

	// Aggregate component results to module results
	return AggregateToWorkResults(results, work), nil
}

// GetRegistry returns the lock tracking registry.
// Useful for external components that need lock visualization.
func (o *Orchestrator) GetRegistry() *locktracker.Registry {
	return o.registry
}

// GetOrchestratorOut returns the writer for orchestrator-level output.
// When TUI is enabled, this writes to the log file only.
// When TUI is not enabled, this writes to both stdout and log file.
func (o *Orchestrator) GetOrchestratorOut() io.Writer {
	return o.orchestratorOut
}

// SetWorker sets the worker function for the orchestrator.
// Must be called before Run or RunLayered.
func (o *Orchestrator) SetWorker(worker WorkerFunc) {
	o.worker = worker
}

// SetUnitExtras stores additional data for a component result.
// This is called by workers to pass test counts or other data that will be
// merged into the UnitResult when processing completes.
// Does nothing if no component scheduler is active.
func (o *Orchestrator) SetUnitExtras(module, component string, extras UnitExtras) {
	if o.currentScheduler != nil {
		o.currentScheduler.SetUnitExtras(module, component, extras)
	}
}

// countUniqueModules counts unique modules in work units.
func countUniqueModules(work []workunit.UnitSpec) int {
	seen := make(map[string]bool)
	for _, w := range work {
		seen[w.ID.Module] = true
	}
	return len(seen)
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	// Simple ASCII capitalization
	first := s[0]
	if first >= 'a' && first <= 'z' {
		return string(first-32) + s[1:]
	}
	return s
}

// sanitizePathForFS converts a moniker to a filesystem-safe path
// Replaces : with _ (Windows doesn't allow : in paths)
// Normalizes path separators to forward slashes.
func sanitizePathForFS(path string) string {
	safe := strings.ReplaceAll(path, ":", "_")
	safe = strings.ReplaceAll(safe, "\\", "/")
	return safe
}

// formatMonikerList formats a list of monikers for display.
func formatMonikerList(monikers []string) string {
	if len(monikers) <= 5 {
		return fmt.Sprintf("%v", monikers)
	}
	return fmt.Sprintf("%v... (%d total)", monikers[:5], len(monikers))
}
