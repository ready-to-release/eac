package tui

import (
	"io"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
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
func (o *TUIObserver) OnEvent(event interfaces.ExecutionEvent) {
	switch e := event.(type) {
	case interfaces.PhaseStartedEvent:
		o.onPhaseStarted(e)
	case interfaces.PhaseCompletedEvent:
		o.onPhaseCompleted(e)
	case interfaces.UnitQueuedEvent:
		o.onUnitQueued(e)
	case interfaces.UnitStartedEvent:
		o.onUnitStarted(e)
	case interfaces.UnitCompletedEvent:
		o.onUnitCompleted(e)
	case interfaces.ProgressUpdateEvent:
		o.onProgressUpdate(e)
	case interfaces.ResourceStatusEvent:
		o.onResourceStatus(e)
	case interfaces.ToolStatusEvent:
		o.onToolStatus(e)
	case interfaces.OutputLineEvent:
		o.onOutputLine(e)
	case interfaces.SummaryReadyEvent:
		o.onSummaryReady(e)
	case interfaces.InitSummaryEvent:
		o.onInitSummary(e)
	}
}

func (o *TUIObserver) onPhaseStarted(e interfaces.PhaseStartedEvent) {
	phase := Phase(e.Phase)
	o.console.SetPhase(phase)
}

func (o *TUIObserver) onPhaseCompleted(e interfaces.PhaseCompletedEvent) {
	phase := Phase(e.Phase)
	o.console.CompletePhase(phase, e.Success, e.Summary)
}

func (o *TUIObserver) onUnitQueued(e interfaces.UnitQueuedEvent) {
	o.console.StartUoW(e.ID, e.DisplayName, e.Weight)
}

func (o *TUIObserver) onUnitStarted(e interfaces.UnitStartedEvent) {
	o.console.MarkUoWRunning(e.ID)
}

func (o *TUIObserver) onUnitCompleted(e interfaces.UnitCompletedEvent) {
	if e.ExitCode < 0 && !e.CacheTime.IsZero() {
		o.console.MarkUoWCompleteWithCacheInfo(e.ID, e.ExitCode, e.CacheTime, e.LogPath)
	} else {
		o.console.MarkUoWComplete(e.ID, e.ExitCode)
	}
}

func (o *TUIObserver) onProgressUpdate(e interfaces.ProgressUpdateEvent) {
	o.console.UpdateStatus(Status{
		Phase:       "Run",
		Running:     e.Running,
		Completed:   e.Completed,
		Total:       e.Total,
		Layer:       e.CurrentLayer,
		TotalLayers: e.TotalLayers,
	})
}

func (o *TUIObserver) onResourceStatus(e interfaces.ResourceStatusEvent) {
	locks := make([]LockStatus, len(e.Resources))
	for i, r := range e.Resources {
		locks[i] = LockStatus{
			Name:     r.Name,
			Type:     r.Type,
			Capacity: r.Capacity,
			Used:     r.Used,
			Waiting:  r.Waiting,
		}
	}
	o.console.UpdateStatus(Status{Locks: locks})
}

func (o *TUIObserver) onToolStatus(e interfaces.ToolStatusEvent) {
	o.console.UpdateStatus(Status{
		ActiveContainers:  e.ActiveContainers,
		UsedContainers:    e.UsedContainers,
		ActiveSystemTools: e.ActiveSystem,
		UsedSystemTools:   e.UsedSystem,
	})
}

func (o *TUIObserver) onOutputLine(e interfaces.OutputLineEvent) {
	level := LevelInfo
	switch e.Level {
	case interfaces.OutputLevelWarn:
		level = LevelWarn
	case interfaces.OutputLevelError:
		level = LevelError
	}
	o.console.SendLine(Line{
		Text:      e.Text,
		Source:    e.Source,
		Level:     level,
		Timestamp: e.Time,
	})
}

func (o *TUIObserver) onSummaryReady(e interfaces.SummaryReadyEvent) {
	o.console.SendSummary(&SummaryData{
		Success:   e.Success,
		TotalTime: e.TotalTime,
		Details:   e.Details,
		NextSteps: e.NextSteps,
	})
}

func (o *TUIObserver) onInitSummary(e interfaces.InitSummaryEvent) {
	// Convert to TUI InitSummary format
	layers := make([]ExecutionLayer, len(e.Layers))
	for i, layer := range e.Layers {
		modules := make([]ExecutionModule, len(layer.Modules))
		for j, mod := range layer.Modules {
			uows := make([]UoWEntry, len(mod.Units))
			for k, unit := range mod.Units {
				uows[k] = UoWEntry{
					ID:          unit.ID,
					DisplayName: unit.DisplayName,
					Weight:      unit.Weight,
				}
			}
			modules[j] = ExecutionModule{
				Name: mod.Name,
				UoWs: uows,
			}
		}
		layers[i] = ExecutionLayer{Modules: modules}
	}

	o.console.SetInitSummary(&InitSummary{
		Command:           e.Command,
		ExecutionContext:  e.ExecutionContext,
		RequestedModules:  e.RequestedModules,
		CalculatedModules: e.ResolvedModules,
		UoWCount:          e.TotalUnits,
		ExecutionTree:     layers,
		LayerCount:        len(layers),
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
	})
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
