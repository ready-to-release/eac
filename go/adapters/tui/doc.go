// Package tui provides a text user interface abstraction for CLI commands.
//
// This package defines the core interfaces and types for building interactive
// terminal user interfaces. It supports multiple TUI implementations that can
// be bound to specific commands via a registry pattern.
//
// # Architecture
//
// The TUI system is built around three core concepts:
//
// 1. Console Interface: The primary interface that all TUI implementations must
// satisfy. It provides methods for lifecycle management, output handling,
// status updates, phase management, and module tracking.
//
// 2. Registry: A command-to-TUI binding system that maps command paths to
// specific TUI implementations. Commands can be matched exactly ("build"),
// by prefix ("work *"), or fall back to a default TUI ("*").
//
// 3. Implementations: Two built-in TUI implementations are provided:
//   - parallel: Rich visualization for build, test, lint, and scan commands
//     with multi-pane layouts, progress tracking, and output streaming.
//   - basic: Simple interactive command picker for general commands,
//     providing subcommand navigation and execution.
//
// # Subpackages
//
//   - console: Core bubbletea model and rendering logic for the parallel TUI.
//   - stream: Output stream handling and filtering utilities.
//   - parallel: Parallel task TUI implementation and registration.
//   - basic: Default interactive TUI implementation and registration.
//
// # Usage
//
// TUI implementations register themselves during init():
//
//	func init() {
//	    tui.Register("build", parallel.Factory())
//	    tui.Register("*", basic.Factory())
//	}
//
// Commands can then create TUIs through the registry:
//
//	console := tui.NewForCommand("build", tui.Config{
//	    RunPhaseName: "building",
//	})
//	console.StartAsync(ctx)
//	defer console.Stop()
//
// # Type Aliases
//
// For compatibility with the internal console package, this package re-exports
// key types:
//
//   - Line, Level, LevelInfo, LevelWarn, LevelError
//   - Status, LockStatus
//   - Phase, PhaseInit, PhaseRun, PhaseSummary
//   - PhaseStatus, PhasePending, PhaseActive, PhaseComplete, PhaseFailed
//   - SummaryData, InitSummary, InitSummaryFlags
//   - ExecutionLayer, ExecutionModule, PlannedTool
//   - SubcommandInfo
package tui
