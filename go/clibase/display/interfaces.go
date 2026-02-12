// Package display defines TUI display interfaces and data types for clibase.
//
// This package decouples clibase from adapters/tui by providing interfaces
// and data types that clibase uses. The adapters/tui package implements
// these interfaces, and concrete implementations are injected at runtime.
//
// Dependency direction: adapters/tui -> clibase/display (not the reverse).
package display

import (
	"context"
	"io"
	"time"
)

// ---------------------------------------------------------------------------
// Sub-interfaces: focused concerns extracted from Console (TD-072)
// ---------------------------------------------------------------------------

// ConsoleLifecycle manages the TUI console start/stop lifecycle.
type ConsoleLifecycle interface {
	Start(ctx context.Context) error
	StartAsync(ctx context.Context)
	Wait()
	Stop()
}

// ConsoleOutput handles writing output lines, results, status updates,
// and provides writer factories and status refresh configuration.
type ConsoleOutput interface {
	NewWriter(source string, logWriter io.Writer) io.Writer
	SendLine(line Line)
	WriteResult(text string)
	UpdateStatus(status Status)
	StatusRefreshInterval() time.Duration
}

// ConsolePhaseManagement controls phase transitions and phase-scoped output.
type ConsolePhaseManagement interface {
	SetPhase(phase Phase)
	CompletePhase(phase Phase, success bool, summary string)
	WriteToPhase(phase Phase, text string)
	SetPhaseSummary(phase Phase, summary string)
}

// ConsoleUoWTracking tracks unit-of-work lifecycle events for parallel TUIs.
type ConsoleUoWTracking interface {
	StartUoW(moniker, displayName string, weight int)
	MarkUoWRunning(moniker string)
	MarkUoWComplete(moniker string, exitCode int)
	MarkUoWCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string)
}

// ConsolePlanning handles early configuration, planned work streaming,
// enrichment updates, and summary delivery.
type ConsolePlanning interface {
	SendConfigReady(commandName, executionContext, parallelismMode string,
		effectiveWorkers, weightedCapacity int, outputDir string)
	SendPlannedWork(items []PlannedWorkItem)
	EnrichUoW(enrichment UoWEnrichment)
	SignalAllWorkDone()
	SetInitSummary(summary *InitSummary)
	SendSummary(data *SummaryData)
}

// ---------------------------------------------------------------------------
// Console: composed interface (backward compatible)
// ---------------------------------------------------------------------------

// Console is the interface for TUI console implementations.
// This mirrors the contract Console interface but is defined in clibase
// to avoid the layering violation of clibase importing adapters.
//
// Console is composed of focused sub-interfaces. Consumers that only need
// a subset of functionality should accept the narrower sub-interface instead.
type Console interface {
	ConsoleLifecycle
	ConsoleOutput
	ConsolePhaseManagement
	ConsoleUoWTracking
	ConsolePlanning
}

// ConsoleFactory creates Console instances with the given configuration.
type ConsoleFactory func(config Config) Console
