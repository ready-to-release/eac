// Package parallel provides a TUI implementation for parallel task execution.
// It wraps the existing ParallelConsole to implement the tui.Console interface.
package parallel

import (
	"context"
	"io"
	"time"

	"github.com/ready-to-release/eac/go/adapters/tui"
)

// Console wraps the existing ParallelConsole to implement tui.Console.
// This provides parallel task visualization for build, test, lint, and scan commands.
type Console struct {
	inner *tui.ParallelConsole
}

// New creates a new parallel Console with the given configuration.
func New(config tui.Config) *Console {
	return &Console{
		inner: tui.NewParallelConsole(config),
	}
}

// Factory returns a ConsoleFactory for the parallel TUI.
func Factory() tui.ConsoleFactory {
	return func(config tui.Config) tui.Console {
		return New(config)
	}
}

// Inner returns the underlying ParallelConsole.
// This is provided for backward compatibility during migration.
func (c *Console) Inner() *tui.ParallelConsole {
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
func (c *Console) SendLine(line tui.Line) {
	// tui.Line is an alias for tuicontract.Line, so pass through directly
	c.inner.SendLine(line)
}

// WriteResult writes a line to the results buffer.
func (c *Console) WriteResult(text string) {
	c.inner.WriteResult(text)
}

// UpdateStatus sends a status update to the console.
func (c *Console) UpdateStatus(status tui.Status) {
	// tui.Status is an alias for tuicontract.Status, so pass through directly
	c.inner.UpdateStatus(status)
}

// SetPhase switches to a new phase.
func (c *Console) SetPhase(phase tui.Phase) {
	c.inner.SetPhase(phase)
}

// CompletePhase marks a phase as complete with a summary.
func (c *Console) CompletePhase(phase tui.Phase, success bool, summary string) {
	c.inner.CompletePhase(phase, success, summary)
}

// WriteToPhase writes a line to a specific phase's buffer.
func (c *Console) WriteToPhase(phase tui.Phase, text string) {
	c.inner.WriteToPhase(phase, text)
}

// SetPhaseSummary sets the summary text for a collapsed phase.
func (c *Console) SetPhaseSummary(phase tui.Phase, summary string) {
	c.inner.SetPhaseSummary(phase, summary)
}

// StartUoW notifies the TUI that a module has started.
// moniker is the globally unique ID for matching; displayName is for tab labels.
func (c *Console) StartUoW(moniker, displayName string, weight int) {
	c.inner.StartUoW(moniker, displayName, weight)
}

// MarkUoWRunning notifies the TUI that a module has acquired its execution slot.
func (c *Console) MarkUoWRunning(moniker string) {
	c.inner.MarkUoWRunning(moniker)
}

// MarkUoWComplete notifies the TUI that a module has finished.
func (c *Console) MarkUoWComplete(moniker string, exitCode int) {
	c.inner.MarkUoWComplete(moniker, exitCode)
}

// MarkUoWCompleteWithCacheInfo marks a module as complete with optional cache info.
func (c *Console) MarkUoWCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
	c.inner.MarkUoWCompleteWithCacheInfo(moniker, exitCode, cacheTime, logPath)
}

// SendSummary sends summary data and activates the Summary pane.
func (c *Console) SendSummary(data *tui.SummaryData) {
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
func (c *Console) SetInitSummary(summary *tui.InitSummary) {
	if summary == nil {
		return
	}
	c.inner.SetInitSummary(summary)
}

// Compile-time check that Console implements tui.Console.
var _ tui.Console = (*Console)(nil)
