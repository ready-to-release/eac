package tui

import (
	"context"
	"io"
	"time"
)

// Console is the primary interface for all TUI implementations.
// Both parallel task TUIs (build, test, lint, scan) and simple TUIs
// must implement this interface.
type Console interface {
	// Lifecycle
	Start(ctx context.Context) error
	StartAsync(ctx context.Context)
	Wait()
	Stop()

	// Output
	NewWriter(source string, logWriter io.Writer) io.Writer
	SendLine(line Line)
	WriteResult(text string)

	// Status Updates
	UpdateStatus(status Status)

	// Phase Management
	SetPhase(phase Phase)
	CompletePhase(phase Phase, success bool, summary string)
	WriteToPhase(phase Phase, text string)
	SetPhaseSummary(phase Phase, summary string)

	// Module/Task Tracking (for parallel TUIs)
	StartModule(moniker string, weight int)
	MarkModuleRunning(moniker string)
	MarkModuleComplete(moniker string, exitCode int)
	MarkModuleCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string)

	// Summary
	SendSummary(data *SummaryData)
	SetInitSummary(summary *InitSummary)
}

// ConsoleFactory creates Console instances with the given configuration.
// This allows TUI implementations to be instantiated lazily with runtime config.
type ConsoleFactory func(config Config) Console

// Note: InteractiveConsole interface has been removed.
// For subcommand selection, use the selector package directly:
//   selector.RunSelector(ctx, tui.SubcommandsToOptions(subcommands))
