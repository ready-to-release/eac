package tui

import (
	"context"
	"io"
	"time"

	"github.com/ready-to-release/eac/go/clibase/display"
)

// consoleAdapter wraps a tui.Console (contract Console) to implement display.Console.
// It converts between display package types and contract types.
type consoleAdapter struct {
	inner Console
}

// Lifecycle methods - pass through directly.

func (a *consoleAdapter) Start(ctx context.Context) error {
	return a.inner.Start(ctx)
}

func (a *consoleAdapter) StartAsync(ctx context.Context) {
	a.inner.StartAsync(ctx)
}

func (a *consoleAdapter) Wait() {
	a.inner.Wait()
}

func (a *consoleAdapter) Stop() {
	a.inner.Stop()
}

// Output methods.

func (a *consoleAdapter) NewWriter(source string, logWriter io.Writer) io.Writer {
	return a.inner.NewWriter(source, logWriter)
}

func (a *consoleAdapter) SendLine(line display.Line) {
	a.inner.SendLine(Line{
		Text:      line.Text,
		Source:    line.Source,
		Level:     Level(line.Level),
		Timestamp: line.Timestamp,
	})
}

func (a *consoleAdapter) WriteResult(text string) {
	a.inner.WriteResult(text)
}

// Status Updates.

func (a *consoleAdapter) UpdateStatus(status display.Status) {
	locks := make([]LockStatus, len(status.Locks))
	for i, l := range status.Locks {
		locks[i] = LockStatus{
			Name:     l.Name,
			Type:     l.Type,
			Capacity: l.Capacity,
			Used:     l.Used,
			Waiting:  l.Waiting,
		}
	}
	a.inner.UpdateStatus(Status{
		Phase:                 status.Phase,
		Running:               status.Running,
		Completed:             status.Completed,
		Total:                 status.Total,
		Roof:                  status.Roof,
		PressureTarget:        status.PressureTarget,
		Locks:                 locks,
		ActiveContainerTools:  status.ActiveContainerTools,
		UsedContainerTools:    status.UsedContainerTools,
		ActiveSystemTools:     status.ActiveSystemTools,
		UsedSystemTools:       status.UsedSystemTools,
		DockerMemPercent:      status.DockerMemPercent,
		DockerAvailable:       status.DockerAvailable,
		RunningContainerCount: status.RunningContainerCount,
		TotalContainerCount:   status.TotalContainerCount,
		RunningSystemCount:    status.RunningSystemCount,
		TotalSystemCount:      status.TotalSystemCount,
	})
}

// Configuration.

func (a *consoleAdapter) StatusRefreshInterval() time.Duration {
	return a.inner.StatusRefreshInterval()
}

// Phase Management.

func (a *consoleAdapter) SetPhase(phase display.Phase) {
	a.inner.SetPhase(Phase(phase))
}

func (a *consoleAdapter) CompletePhase(phase display.Phase, success bool, summary string) {
	a.inner.CompletePhase(Phase(phase), success, summary)
}

func (a *consoleAdapter) WriteToPhase(phase display.Phase, text string) {
	a.inner.WriteToPhase(Phase(phase), text)
}

func (a *consoleAdapter) SetPhaseSummary(phase display.Phase, summary string) {
	a.inner.SetPhaseSummary(Phase(phase), summary)
}

// UoW/Task Tracking.

func (a *consoleAdapter) StartUoW(moniker, displayName string, weight int) {
	a.inner.StartUoW(moniker, displayName, weight)
}

func (a *consoleAdapter) MarkUoWRunning(moniker string) {
	a.inner.MarkUoWRunning(moniker)
}

func (a *consoleAdapter) MarkUoWComplete(moniker string, exitCode int) {
	a.inner.MarkUoWComplete(moniker, exitCode)
}

func (a *consoleAdapter) MarkUoWCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
	a.inner.MarkUoWCompleteWithCacheInfo(moniker, exitCode, cacheTime, logPath)
}

// Early Configuration.

func (a *consoleAdapter) SendConfigReady(commandName, executionContext, parallelismMode string,
	effectiveWorkers, weightedCapacity int, outputDir string) {
	a.inner.SendConfigReady(commandName, executionContext, parallelismMode,
		effectiveWorkers, weightedCapacity, outputDir)
}

// Summary.

func (a *consoleAdapter) SendSummary(data *display.SummaryData) {
	if data == nil {
		return
	}
	a.inner.SendSummary(&SummaryData{
		Success:     data.Success,
		TotalTime:   data.TotalTime,
		InitSummary: data.InitSummary,
		RunSummary:  data.RunSummary,
		Details:     data.Details,
		NextSteps:   data.NextSteps,
	})
}

func (a *consoleAdapter) SetInitSummary(summary *display.InitSummary) {
	if summary == nil {
		return
	}

	// Convert execution tree
	modules := make([]ExecutionModule, len(summary.ExecutionTree))
	for i, mod := range summary.ExecutionTree {
		uows := make([]UoWEntry, len(mod.UoWs))
		for k, uow := range mod.UoWs {
			uows[k] = UoWEntry{
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
		modules[i] = ExecutionModule{
			Name: mod.Name,
			UoWs: uows,
		}
	}

	// Convert planned tools
	plannedTools := make([]PlannedTool, len(summary.PlannedTools))
	for i, tool := range summary.PlannedTools {
		plannedTools[i] = PlannedTool{
			Name:        tool.Name,
			IsContainer: tool.IsContainer,
		}
	}

	a.inner.SetInitSummary(&InitSummary{
		Command:             summary.Command,
		ExecutionContext:    summary.ExecutionContext,
		RequestedModules:    summary.RequestedModules,
		CalculatedModules:   summary.CalculatedModules,
		AddedDepm:           summary.AddedDepm,
		UoWCount:            summary.UoWCount,
		ExecutionTree:       modules,
		ParallelismMode:     summary.ParallelismMode,
		EffectiveWorkers:    summary.EffectiveWorkers,
		TurboBoost:          summary.TurboBoost,
		WeightedCapacity:    summary.WeightedCapacity,
		Flags: InitSummaryFlags{
			TidyFirst:    summary.Flags.TidyFirst,
			ForceRebuild: summary.Flags.ForceRebuild,
			DryRun:       summary.Flags.DryRun,
			UseTUI:       summary.Flags.UseTUI,
			SkipDeps:     summary.Flags.SkipDeps,
			SkipDepm:     summary.Flags.SkipDepm,
		},
		DepsVerified:        summary.DepsVerified,
		DepsSkipped:         summary.DepsSkipped,
		DepsAvailable:       summary.DepsAvailable,
		DepsMissing:         summary.DepsMissing,
		DepmVerified:        summary.DepmVerified,
		DepmSkipped:         summary.DepmSkipped,
		DepmResolved:        summary.DepmResolved,
		DepmExisting:        summary.DepmExisting,
		DepmTotal:           summary.DepmTotal,
		DepmMissing:         summary.DepmMissing,
		IncrementalEnabled:  summary.IncrementalEnabled,
		IncrementalChanged:  summary.IncrementalChanged,
		IncrementalUpToDate: summary.IncrementalUpToDate,
		IncrementalFresh:    summary.IncrementalFresh,
		TestSuiteName:       summary.TestSuiteName,
		TestSelected:        summary.TestSelected,
		TestDiscovered:      summary.TestDiscovered,
		TestOSFiltered:      summary.TestOSFiltered,
		OutputDir:           summary.OutputDir,
		PlannedTools:        plannedTools,
	})
}

// Fast Init + Streaming Updates + Hold-Until-Done.

func (a *consoleAdapter) SendPlannedWork(items []display.PlannedWorkItem) {
	contractItems := make([]PlannedWorkItem, len(items))
	for i, item := range items {
		contractItems[i] = PlannedWorkItem{
			ID:            item.ID,
			DisplayName:   item.DisplayName,
			Weight:        item.Weight,
			Module:        item.Module,
			Component:     item.Component,
			ComponentType: item.ComponentType,
		}
	}
	a.inner.SendPlannedWork(contractItems)
}

func (a *consoleAdapter) EnrichUoW(enrichment display.UoWEnrichment) {
	a.inner.EnrichUoW(UoWEnrichment{
		ID:          enrichment.ID,
		Tool:        enrichment.Tool,
		Container:   enrichment.Container,
		Weight:      enrichment.Weight,
		CacheStatus: CacheHit(enrichment.CacheStatus),
		CacheTime:   enrichment.CacheTime,
		DependsOn:   enrichment.DependsOn,
	})
}

func (a *consoleAdapter) SignalAllWorkDone() {
	a.inner.SignalAllWorkDone()
}

// Compile-time interface check.
var _ display.Console = (*consoleAdapter)(nil)
