package orchestrator

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/display"
)

// SetConsole sets an external TUI console for the orchestrator.
// This allows the cmdframework to inject a console created via the TUI registry.
// If called with a non-nil console, it replaces any console created in New().
// The console should implement the display.Console interface.
func (o *Orchestrator) SetConsole(console display.Console) {
	if console != nil {
		o.tuiConsole = console
		if o.tuiCtx == nil {
			o.tuiCtx, o.tuiCancel = context.WithCancel(context.Background())
		}
	}
}

// WaitTUI waits for the TUI to exit naturally (e.g., user presses a key).
// Use this after SendSummary() to wait for user to review and exit.
func (o *Orchestrator) WaitTUI() {
	if o.tuiConsole != nil {
		o.tuiConsole.Wait()
	}
	// Restore stdout after TUI exits
	o.orchestratorOut = os.Stdout
	o.logger = log.New(o.orchestratorOut, "", 0)
}

// StopTUI stops the TUI console and restores stdout output.
// Must be called before PrintSummary when TUI is enabled.
func (o *Orchestrator) StopTUI() {
	if o.tuiConsole != nil {
		o.tuiConsole.Stop()
		o.tuiConsole = nil
	}
	if o.tuiCancel != nil {
		o.tuiCancel()
		o.tuiCancel = nil
	}
	// Restore stdout for subsequent output (like PrintSummary)
	o.orchestratorOut = os.Stdout
	o.logger = log.New(o.orchestratorOut, "", 0)
}

// IsTUIEnabled returns whether TUI mode is enabled in config.
func (o *Orchestrator) IsTUIEnabled() bool {
	return o.config.TUI
}

// IsTUIStarted returns whether StartTUI has been called.
// When TUI is enabled but not started, WriteInit should output to console.
func (o *Orchestrator) IsTUIStarted() bool {
	o.tuiMu.Lock()
	defer o.tuiMu.Unlock()
	return o.tuiStarted
}

// StartTUI starts only the TUI console without running any jobs.
// Use this to enable phase output before calling Run or RunLayered.
// Returns quickly after TUI is initialized.
// Call Init() first if you need logging infrastructure.
func (o *Orchestrator) StartTUI() {
	if o.tuiConsole == nil {
		// Debug: TUI console not set - this means the type assertion failed or
		// useRegistryTUI was false. Check cmdframework init.go debug logs.
		return
	}
	o.tuiConsole.StartAsync(o.tuiCtx)
	// Set initial phase to Init
	o.SetPhase(display.PhaseInit)
	// Mark TUI as started so WriteInit knows to skip console output
	o.tuiMu.Lock()
	o.tuiStarted = true
	o.tuiMu.Unlock()
}

// phaseWriter implements io.Writer and forwards all writes to a specific phase.
type phaseWriter struct {
	orch  *Orchestrator
	phase display.Phase
}

// Write implements io.Writer by emitting OutputLineEvent for the phase.
func (w *phaseWriter) Write(p []byte) (n int, err error) {
	// Convert bytes to string, trim trailing newline (observers add their own)
	text := string(p)
	text = strings.TrimSuffix(text, "\n")
	if text != "" {
		w.orch.WriteToPhase(w.phase, text)
	}
	return len(p), nil
}

// GetTUIWriter returns an io.Writer that sends output to the specified phase.
// Returns nil if TUI mode is not enabled.
func (o *Orchestrator) GetTUIWriter(phase display.Phase) io.Writer {
	if !o.config.TUI {
		return nil
	}
	return &phaseWriter{orch: o, phase: phase}
}

// tuiMarkRunning adds a module to the running list and updates TUI status.
func (o *Orchestrator) tuiMarkRunning(moniker string) {
	o.tuiMu.Lock()
	o.tuiRunning = append(o.tuiRunning, moniker)
	running := make([]string, len(o.tuiRunning))
	copy(running, o.tuiRunning)
	completed := o.tuiCompleted
	total := o.tuiTotal
	o.tuiMu.Unlock()

	// Emit events to observers
	o.emit(core.UnitStartedEvent{
		Time: time.Now(),
		ID:   moniker,
	})
	o.emit(core.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   running,
		Completed: completed,
		Total:     total,
	})
}

// tuiMarkCompleted removes a module from running, increments completed count, and reports exit code.
func (o *Orchestrator) tuiMarkCompleted(moniker string, exitCode int, duration time.Duration) {
	o.tuiMu.Lock()
	// Remove from running list
	for i, m := range o.tuiRunning {
		if m == moniker {
			o.tuiRunning = append(o.tuiRunning[:i], o.tuiRunning[i+1:]...)
			break
		}
	}
	o.tuiCompleted++
	running := make([]string, len(o.tuiRunning))
	copy(running, o.tuiRunning)
	completed := o.tuiCompleted
	total := o.tuiTotal
	o.tuiMu.Unlock()

	// Emit events to observers
	o.emit(core.UnitCompletedEvent{
		Time:     time.Now(),
		ID:       moniker,
		ExitCode: exitCode,
		Duration: duration,
	})
	o.emit(core.ProgressUpdateEvent{
		Time:      time.Now(),
		Running:   running,
		Completed: completed,
		Total:     total,
	})
}

// AddObserver registers an observer to receive execution events.
func (o *Orchestrator) AddObserver(observer core.ExecutionObserver) {
	o.observersMu.Lock()
	defer o.observersMu.Unlock()
	o.observers = append(o.observers, observer)
}

// RemoveObserver unregisters an observer.
func (o *Orchestrator) RemoveObserver(observer core.ExecutionObserver) {
	o.observersMu.Lock()
	defer o.observersMu.Unlock()
	for i, obs := range o.observers {
		if obs == observer {
			o.observers = append(o.observers[:i], o.observers[i+1:]...)
			break
		}
	}
}

// SetWriterFactory sets the factory for creating output writers.
// If not set, output goes only to log files.
func (o *Orchestrator) SetWriterFactory(factory core.WriterFactory) {
	o.writerFactory = factory
}

// emit sends an event to all registered observers.
func (o *Orchestrator) emit(event core.ExecutionEvent) {
	o.observersMu.RLock()
	observers := make([]core.ExecutionObserver, len(o.observers))
	copy(observers, o.observers)
	o.observersMu.RUnlock()

	for _, obs := range observers {
		obs.OnEvent(event)
	}
}

// SendSummary sends summary data to all registered observers.
// This emits a SummaryReadyEvent that observers can handle appropriately.
func (o *Orchestrator) SendSummary(data *display.SummaryData) {
	if data == nil {
		return
	}
	o.emit(core.SummaryReadyEvent{
		Time:      time.Now(),
		Success:   data.Success,
		TotalTime: data.TotalTime,
		Details:   data.Details,
		NextSteps: data.NextSteps,
	})
}

// SendConfigReady sends early configuration metadata to all registered observers.
// This enables progressive TUI display before full init summary.
func (o *Orchestrator) SendConfigReady(commandName, executionContext, parallelismMode string,
	effectiveWorkers, weightedCapacity int, outputDir string) {
	o.emit(core.ConfigReadyEvent{
		Time:             time.Now(),
		CommandName:      commandName,
		ExecutionContext: executionContext,
		ParallelismMode:  parallelismMode,
		EffectiveWorkers: effectiveWorkers,
		WeightedCapacity: weightedCapacity,
		OutputDir:        outputDir,
	})
}

// SetInitSummary sends structured init summary data to all registered observers.
// This emits an InitSummaryEvent that observers can handle appropriately.
func (o *Orchestrator) SetInitSummary(data *display.InitSummary) {
	if data == nil {
		return
	}

	// Convert TUI InitSummary to observer event
	modules := make([]core.ModuleInfo, len(data.ExecutionTree))
	for i, mod := range data.ExecutionTree {
		units := make([]core.UnitInfo, len(mod.UoWs))
		for j, uow := range mod.UoWs {
			units[j] = core.UnitInfo{
				ID:            uow.ID,
				DisplayName:   uow.DisplayName,
				Weight:        uow.Weight,
				Tags:          uow.Tags,
				Module:        uow.Module,
				Component:     uow.Component,
				Tool:          uow.Tool,
				ComponentType: uow.ComponentType,
				Container:     uow.Container,
			}
		}
		modules[i] = core.ModuleInfo{
			Name:  mod.Name,
			Units: units,
		}
	}

	// Convert PlannedTools to interface type
	plannedTools := make([]core.PlannedToolInfo, len(data.PlannedTools))
	for i, tool := range data.PlannedTools {
		plannedTools[i] = core.PlannedToolInfo{
			Name:        tool.Name,
			IsContainer: tool.IsContainer,
		}
	}

	o.emit(core.InitSummaryEvent{
		Time:             time.Now(),
		Command:          data.Command,
		ExecutionContext: data.ExecutionContext,
		RequestedModules: data.RequestedModules,
		ResolvedModules:  data.CalculatedModules,
		TotalUnits:       data.UoWCount,
		Modules:          modules,
		Parallelism: core.ParallelismInfo{
			Mode:             data.ParallelismMode,
			EffectiveWorkers: data.EffectiveWorkers,
			TurboBoost:       data.TurboBoost,
			WeightedCapacity: data.WeightedCapacity,
		},
		Flags: core.FlagsInfo{
			TidyFirst:    data.Flags.TidyFirst,
			ForceRebuild: data.Flags.ForceRebuild,
			DryRun:       data.Flags.DryRun,
			UseTUI:       data.Flags.UseTUI,
			SkipDeps:     data.Flags.SkipDeps,
			SkipDepm:     data.Flags.SkipDepm,
		},
		PlannedTools: plannedTools,
	})

	o.initSummaryEmitted = true
}

// SendPlannedWork sends predicted work items to all registered observers.
// This enables the TUI to show grey skeleton tabs before tool resolution.
func (o *Orchestrator) SendPlannedWork(items []display.PlannedWorkItem) {
	eventItems := make([]core.PlannedWorkItemInfo, len(items))
	for i, item := range items {
		eventItems[i] = core.PlannedWorkItemInfo{
			ID:            item.ID,
			DisplayName:   item.DisplayName,
			Weight:        item.Weight,
			Module:        item.Module,
			Component:     item.Component,
			ComponentType: item.ComponentType,
		}
	}
	o.emit(core.PlannedWorkEvent{
		Time:  time.Now(),
		Items: eventItems,
	})
}

// EnrichUoW sends incremental enrichment data for a planned work item.
func (o *Orchestrator) EnrichUoW(enrichment display.UoWEnrichment) {
	o.emit(core.UoWEnrichmentEvent{
		Time:        time.Now(),
		ID:          enrichment.ID,
		Tool:        enrichment.Tool,
		Container:   enrichment.Container,
		Weight:      enrichment.Weight,
		CacheStatus: int(enrichment.CacheStatus),
		CacheTime:   enrichment.CacheTime,
		DependsOn:   enrichment.DependsOn,
	})
}

// SignalAllWorkDone signals that all work including AfterExecute hooks is complete.
// The TUI must not exit before receiving this signal.
func (o *Orchestrator) SignalAllWorkDone() {
	o.emit(core.AllWorkDoneEvent{
		Time: time.Now(),
	})
}

// SendInitLine sends a line to the Init phase buffer.
func (o *Orchestrator) SendInitLine(text string) {
	o.WriteToPhase(display.PhaseInit, text)
}

// SendEndLine sends a line to the results buffer (appears below Run pane).
// This emits an OutputLineEvent to all registered observers.
func (o *Orchestrator) SendEndLine(text string) {
	o.emit(core.OutputLineEvent{
		Time:   time.Now(),
		Source: "results",
		Text:   text,
		Level:  core.OutputLevelInfo,
	})
}
