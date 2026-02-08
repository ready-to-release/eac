package tui

import (
	"io"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// TUIObserver implements ExecutionObserver and WriterFactory,
// translating generic execution events to TUI operations.
type TUIObserver struct {
	console Console
}

// NewTUIObserver creates a new TUI observer wrapping a Console.
func NewTUIObserver(console Console) *TUIObserver {
	return &TUIObserver{console: console}
}

// OnEvent handles execution events and translates them to TUI operations.
func (o *TUIObserver) OnEvent(event core.ExecutionEvent) {
	switch e := event.(type) {
	case core.PhaseStartedEvent:
		o.onPhaseStarted(e)
	case core.PhaseCompletedEvent:
		o.onPhaseCompleted(e)
	case core.UnitQueuedEvent:
		o.onUnitQueued(e)
	case core.UnitStartedEvent:
		o.onUnitStarted(e)
	case core.UnitCompletedEvent:
		o.onUnitCompleted(e)
	case core.ProgressUpdateEvent:
		o.onProgressUpdate(e)
	case core.ResourceStatusEvent:
		o.onResourceStatus(e)
	case core.ToolStatusEvent:
		o.onToolStatus(e)
	case core.OutputLineEvent:
		o.onOutputLine(e)
	case core.SummaryReadyEvent:
		o.onSummaryReady(e)
	case core.ConfigReadyEvent:
		o.onConfigReady(e)
	case core.InitSummaryEvent:
		o.onInitSummary(e)
	case core.PlannedWorkEvent:
		o.onPlannedWork(e)
	case core.UoWEnrichmentEvent:
		o.onUoWEnrichment(e)
	case core.AllWorkDoneEvent:
		o.onAllWorkDone()
	}
}

func (o *TUIObserver) onPhaseStarted(e core.PhaseStartedEvent) {
	phase := Phase(e.Phase)
	o.console.SetPhase(phase)
}

func (o *TUIObserver) onPhaseCompleted(e core.PhaseCompletedEvent) {
	phase := Phase(e.Phase)
	o.console.CompletePhase(phase, e.Success, e.Summary)
}

func (o *TUIObserver) onUnitQueued(e core.UnitQueuedEvent) {
	o.console.StartUoW(e.ID, e.DisplayName, e.Weight)
}

func (o *TUIObserver) onUnitStarted(e core.UnitStartedEvent) {
	o.console.MarkUoWRunning(e.ID)
}

func (o *TUIObserver) onUnitCompleted(e core.UnitCompletedEvent) {
	if e.ExitCode < 0 && !e.CacheTime.IsZero() {
		o.console.MarkUoWCompleteWithCacheInfo(e.ID, e.ExitCode, e.CacheTime, e.LogPath)
	} else {
		o.console.MarkUoWComplete(e.ID, e.ExitCode)
	}
}

func (o *TUIObserver) onProgressUpdate(e core.ProgressUpdateEvent) {
	o.console.UpdateStatus(Status{
		Phase:          "Run",
		Running:        e.Running,
		Completed:      e.Completed,
		Total:          e.Total,
		Roof:           e.Roof,
		PressureTarget: e.PressureTarget,
	})
}

func (o *TUIObserver) onResourceStatus(e core.ResourceStatusEvent) {
	locks := make([]LockStatus, len(e.Resources))

	// Extract docker memory metrics from docker-scheduler resource
	var dockerMemPercent float64
	var dockerAvailable bool

	for i, r := range e.Resources {
		locks[i] = LockStatus{
			Name:     r.Name,
			Type:     r.Type,
			Capacity: r.Capacity,
			Used:     r.Used,
			Waiting:  r.Waiting,
		}

		// Check for docker-scheduler to extract docker memory metrics
		// Mark as available when resource exists (even with 0 capacity during init)
		// Only calculate percentage when capacity > 0 to avoid division by zero
		if r.Name == "docker-scheduler" {
			dockerAvailable = true
			if r.Capacity > 0 {
				dockerMemPercent = float64(r.Used) / float64(r.Capacity) * 100
			}
		}
	}

	o.console.UpdateStatus(Status{
		Locks:            locks,
		DockerMemPercent: dockerMemPercent,
		DockerAvailable:  dockerAvailable,
	})
}

func (o *TUIObserver) onToolStatus(e core.ToolStatusEvent) {
	o.console.UpdateStatus(Status{
		ActiveContainerTools:  e.ActiveContainerTools,
		UsedContainerTools:    e.UsedContainerTools,
		ActiveSystemTools:     e.ActiveSystem,
		UsedSystemTools:       e.UsedSystem,
		RunningContainerCount: e.ContainerInstancesRunning,
		TotalContainerCount:   e.ContainerInstancesTotal,
		RunningSystemCount:    e.SystemInvocationsRunning,
		TotalSystemCount:      e.SystemInvocationsTotal,
	})
}

func (o *TUIObserver) onOutputLine(e core.OutputLineEvent) {
	level := LevelInfo
	switch e.Level {
	case core.OutputLevelWarn:
		level = LevelWarn
	case core.OutputLevelError:
		level = LevelError
	}
	o.console.SendLine(Line{
		Text:      e.Text,
		Source:    e.Source,
		Level:     level,
		Timestamp: e.Time,
	})
}

func (o *TUIObserver) onSummaryReady(e core.SummaryReadyEvent) {
	o.console.SendSummary(&SummaryData{
		Success:   e.Success,
		TotalTime: e.TotalTime,
		Details:   e.Details,
		NextSteps: e.NextSteps,
	})
}

func (o *TUIObserver) onConfigReady(e core.ConfigReadyEvent) {
	o.console.SendConfigReady(e.CommandName, e.ExecutionContext, e.ParallelismMode,
		e.EffectiveWorkers, e.WeightedCapacity, e.OutputDir)
}

func (o *TUIObserver) onInitSummary(e core.InitSummaryEvent) {
	// Convert to TUI InitSummary format
	modules := make([]ExecutionModule, len(e.Modules))
	for i, mod := range e.Modules {
		uows := make([]UoWEntry, len(mod.Units))
		for k, unit := range mod.Units {
			uows[k] = UoWEntry{
				ID:            unit.ID,
				DisplayName:   unit.DisplayName,
				Weight:        unit.Weight,
				Tags:          unit.Tags,
				Module:        unit.Module,
				Component:     unit.Component,
				Tool:          unit.Tool,
				ComponentType: unit.ComponentType,
				Container:     unit.Container,
			}
		}
		modules[i] = ExecutionModule{
			Name: mod.Name,
			UoWs: uows,
		}
	}

	// Convert PlannedTools
	plannedTools := make([]PlannedTool, len(e.PlannedTools))
	for i, tool := range e.PlannedTools {
		plannedTools[i] = PlannedTool{
			Name:        tool.Name,
			IsContainer: tool.IsContainer,
		}
	}

	o.console.SetInitSummary(&InitSummary{
		Command:           e.Command,
		ExecutionContext:  e.ExecutionContext,
		RequestedModules:  e.RequestedModules,
		CalculatedModules: e.ResolvedModules,
		UoWCount:          e.TotalUnits,
		ExecutionTree:     modules,
		ParallelismMode:   e.Parallelism.Mode,
		EffectiveWorkers:  e.Parallelism.EffectiveWorkers,
		TurboBoost:        e.Parallelism.TurboBoost,
		WeightedCapacity:  e.Parallelism.WeightedCapacity,
		Flags: InitSummaryFlags{
			TidyFirst:    e.Flags.TidyFirst,
			ForceRebuild: e.Flags.ForceRebuild,
			DryRun:       e.Flags.DryRun,
			UseTUI:       e.Flags.UseTUI,
			SkipDeps:     e.Flags.SkipDeps,
			SkipDepm:     e.Flags.SkipDepm,
		},
		PlannedTools: plannedTools,
	})
}

func (o *TUIObserver) onPlannedWork(e core.PlannedWorkEvent) {
	items := make([]PlannedWorkItem, len(e.Items))
	for i, item := range e.Items {
		items[i] = PlannedWorkItem{
			ID:            item.ID,
			DisplayName:   item.DisplayName,
			Weight:        item.Weight,
			Module:        item.Module,
			Component:     item.Component,
			ComponentType: item.ComponentType,
		}
	}
	o.console.SendPlannedWork(items)
}

func (o *TUIObserver) onUoWEnrichment(e core.UoWEnrichmentEvent) {
	o.console.EnrichUoW(UoWEnrichment{
		ID:          e.ID,
		Tool:        e.Tool,
		Container:   e.Container,
		Weight:      e.Weight,
		CacheStatus: CacheHit(e.CacheStatus),
		CacheTime:   e.CacheTime,
		DependsOn:   e.DependsOn,
	})
}

func (o *TUIObserver) onAllWorkDone() {
	o.console.SignalAllWorkDone()
}

// NewWriter implements WriterFactory for TUI output interception.
func (o *TUIObserver) NewWriter(unitID string, logWriter io.Writer) io.WriteCloser {
	// The console.NewWriter returns io.Writer, we need to wrap it with Close capability
	w := o.console.NewWriter(unitID, logWriter)
	if wc, ok := w.(io.WriteCloser); ok {
		return wc
	}
	return &nopCloserWriter{w}
}

// nopCloserWriter wraps an io.Writer to add a no-op Close method.
type nopCloserWriter struct {
	io.Writer
}

func (n *nopCloserWriter) Close() error { return nil }

// Console returns the underlying TUI console.
func (o *TUIObserver) Console() Console {
	return o.console
}
