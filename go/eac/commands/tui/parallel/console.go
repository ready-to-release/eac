// Package parallel provides a TUI implementation for parallel task execution.
// It wraps the existing internal/tui Console to implement the tui.Console interface.
package parallel

import (
	"context"
	"io"
	"time"

	"github.com/ready-to-release/eac/go/eac/commands/internal/tui"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console"
	rootTUI "github.com/ready-to-release/eac/go/eac/commands/tui"
)

// Console wraps the existing internal/tui.Console to implement tui.Console.
// This provides parallel task visualization for build, test, lint, and scan commands.
type Console struct {
	inner *tui.Console
}

// New creates a new parallel Console with the given configuration.
func New(config rootTUI.Config) *Console {
	innerConfig := tui.Config{
		Height:       config.Height,
		BufferSize:   config.BufferSize,
		RunPhaseName: config.RunPhaseName,
		ASCIIMode:    config.ASCIIMode,
		SkipTUIDelay: config.SkipTUIDelay,
	}

	return &Console{
		inner: tui.New(innerConfig),
	}
}

// Factory returns a ConsoleFactory for the parallel TUI.
func Factory() rootTUI.ConsoleFactory {
	return func(config rootTUI.Config) rootTUI.Console {
		return New(config)
	}
}

// Inner returns the underlying internal/tui.Console.
// This is provided for backward compatibility during migration.
func (c *Console) Inner() *tui.Console {
	return c.inner
}

// Start starts the TUI program. This blocks until the program exits.
func (c *Console) Start(ctx context.Context) error {
	return c.inner.Start(ctx)
}

// StartAsync starts the TUI program in a goroutine.
func (c *Console) StartAsync(ctx context.Context) {
	c.inner.StartAsync(ctx)
}

// Wait waits for the TUI program to fully complete.
func (c *Console) Wait() {
	c.inner.Wait()
}

// Stop stops the TUI program.
func (c *Console) Stop() {
	c.inner.Stop()
}

// NewWriter creates an io.Writer for a module's output.
func (c *Console) NewWriter(source string, logWriter io.Writer) io.Writer {
	return c.inner.NewWriter(source, logWriter)
}

// SendLine directly sends a line to the console.
func (c *Console) SendLine(line rootTUI.Line) {
	c.inner.SendLine(console.Line{
		Text:      line.Text,
		Source:    line.Source,
		Level:     console.Level(line.Level),
		Timestamp: line.Timestamp,
	})
}

// WriteResult writes a line to the results buffer.
func (c *Console) WriteResult(text string) {
	c.inner.WriteResult(text)
}

// UpdateStatus sends a status update to the console.
func (c *Console) UpdateStatus(status rootTUI.Status) {
	// Convert lock statuses
	locks := make([]console.LockStatus, len(status.Locks))
	for i, lock := range status.Locks {
		locks[i] = console.LockStatus{
			Name:     lock.Name,
			Type:     lock.Type,
			Capacity: lock.Capacity,
			Used:     lock.Used,
			Waiting:  lock.Waiting,
		}
	}

	c.inner.UpdateStatus(console.Status{
		Phase:             status.Phase,
		Running:           status.Running,
		Completed:         status.Completed,
		Total:             status.Total,
		Layer:             status.Layer,
		TotalLayers:       status.TotalLayers,
		Locks:             locks,
		ActiveContainers:  status.ActiveContainers,
		UsedContainers:    status.UsedContainers,
		ActiveSystemTools: status.ActiveSystemTools,
		UsedSystemTools:   status.UsedSystemTools,
	})
}

// SetPhase switches to a new phase.
func (c *Console) SetPhase(phase rootTUI.Phase) {
	c.inner.SetPhase(tui.Phase(phase))
}

// CompletePhase marks a phase as complete with a summary.
func (c *Console) CompletePhase(phase rootTUI.Phase, success bool, summary string) {
	c.inner.CompletePhase(tui.Phase(phase), success, summary)
}

// WriteToPhase writes a line to a specific phase's buffer.
func (c *Console) WriteToPhase(phase rootTUI.Phase, text string) {
	c.inner.WriteToPhase(tui.Phase(phase), text)
}

// SetPhaseSummary sets the summary text for a collapsed phase.
func (c *Console) SetPhaseSummary(phase rootTUI.Phase, summary string) {
	c.inner.SetPhaseSummary(tui.Phase(phase), summary)
}

// StartModule notifies the TUI that a module has started.
func (c *Console) StartModule(moniker string, weight int) {
	c.inner.StartModule(moniker, weight)
}

// MarkModuleRunning notifies the TUI that a module has acquired its execution slot.
func (c *Console) MarkModuleRunning(moniker string) {
	c.inner.MarkModuleRunning(moniker)
}

// MarkModuleComplete notifies the TUI that a module has finished.
func (c *Console) MarkModuleComplete(moniker string, exitCode int) {
	c.inner.MarkModuleComplete(moniker, exitCode)
}

// MarkModuleCompleteWithCacheInfo marks a module as complete with optional cache info.
func (c *Console) MarkModuleCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
	c.inner.MarkModuleCompleteWithCacheInfo(moniker, exitCode, cacheTime, logPath)
}

// SendSummary sends summary data and activates the Summary pane.
func (c *Console) SendSummary(data *rootTUI.SummaryData) {
	if data == nil {
		return
	}
	c.inner.SendSummary(&tui.SummaryData{
		Success:     data.Success,
		TotalTime:   data.TotalTime,
		InitSummary: data.InitSummary,
		RunSummary:  data.RunSummary,
		Details:     data.Details,
		NextSteps:   data.NextSteps,
	})
}

// SetInitSummary sends init summary data for structured display.
func (c *Console) SetInitSummary(summary *rootTUI.InitSummary) {
	if summary == nil {
		return
	}

	// Convert execution layers
	layers := make([]tui.ExecutionLayer, len(summary.ExecutionTree))
	for i, layer := range summary.ExecutionTree {
		modules := make([]tui.ExecutionModule, len(layer.Modules))
		for j, mod := range layer.Modules {
			modules[j] = tui.ExecutionModule{
				Name:       mod.Name,
				Components: mod.Components,
			}
		}
		layers[i] = tui.ExecutionLayer{Modules: modules}
	}

	c.inner.SetInitSummary(&tui.InitSummary{
		Command:               summary.Command,
		ExecutionContext:      summary.ExecutionContext,
		RequestedModules:      summary.RequestedModules,
		CalculatedModules:     summary.CalculatedModules,
		AddedDepm:             summary.AddedDepm,
		ComponentCount:        summary.ComponentCount,
		ExecutionTree:         layers,
		LayerCount:            summary.LayerCount,
		LayerSizes:            summary.LayerSizes,
		ComponentsPerModLayer: summary.ComponentsPerModLayer,
		FlatExecution:         summary.FlatExecution,
		ParallelismMode:       summary.ParallelismMode,
		EffectiveWorkers:      summary.EffectiveWorkers,
		TurboBoost:            summary.TurboBoost,
		WeightedCapacity:      summary.WeightedCapacity,
		Flags: tui.InitSummaryFlags{
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
	})
}

// Compile-time check that Console implements tui.Console.
var _ rootTUI.Console = (*Console)(nil)
