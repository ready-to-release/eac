package interfaces

import (
	"context"
	"time"
)

// TUI Hook Interfaces
//
// These interfaces allow TUIs to interact with command execution in controlled ways
// while maintaining the observer pattern. Hooks fire at specific points during
// the cmdframework execution sequence.
//
// Hook Sequence:
//   1. CommandSelectionHook - before module resolution (user picks command)
//   2. UoWDataHook - after resolution (TUI receives execution plan)
//   3. PreTUIStartHook - before TUI boots (final configuration)
//   4. ExitHoldController - after execution (TUI can delay exit)

// ExitHoldController provides controlled exit timing.
// TUIs can signal "hold exit" when the user is interacting with output.
// Commands wait until holds are released or timeout.
type ExitHoldController interface {
	// HoldExit signals that exit should be delayed.
	// Returns a release function that must be called when done.
	// The release function is safe to call multiple times.
	HoldExit() (release func())

	// WaitForRelease blocks until all holds are released OR timeout expires.
	// Returns true if released cleanly, false on timeout or context cancellation.
	// If no holds are active, returns immediately with true.
	WaitForRelease(ctx context.Context, timeout time.Duration) bool
}

// CommandSelectionHook allows TUI to modify command before module resolution.
// Fires during init phase, before AfterInit hook.
type CommandSelectionHook interface {
	// SelectCommand presents options and returns the selected command.
	// If TUI is not interactive, returns original command unchanged.
	// Blocks until user selection or context cancelled.
	SelectCommand(ctx context.Context, req CommandSelectionRequest) CommandSelectionResponse
}

// CommandSelectionRequest contains data for command selection.
type CommandSelectionRequest struct {
	OriginalCommand string          // e.g., "work" with no subcommand
	Options         []CommandOption // Available subcommands
}

// CommandOption describes a selectable command.
type CommandOption struct {
	Name        string
	Description string
	Aliases     []string
}

// CommandSelectionResponse contains the result of command selection.
type CommandSelectionResponse struct {
	SelectedCommand string // The command to execute
	Args            string // Additional arguments
	Cancelled       bool   // User cancelled selection
}

// UoWDataHook receives resolved execution plan for TUI visualization.
// Fires after AfterResolve hook, before verify phase.
type UoWDataHook interface {
	// ReceiveUoWs is called with the resolved execution plan.
	// TUI uses this to pre-create tabs/display before execution.
	// Non-blocking - TUI stores data for later use.
	ReceiveUoWs(data UoWData)
}

// UoWData contains the resolved execution plan.
type UoWData struct {
	Modules []UoWModule // Modules with their units of work
}

// UoWModule describes a module with its units of work.
type UoWModule struct {
	Name  string
	Units []UoWUnit
}

// UoWUnit describes a unit of work.
type UoWUnit struct {
	ID          string   // Longname (context:module:component:tool)
	DisplayName string   // Shortname (module:component)
	Weight      int
	DependsOn   []string // Unit IDs this depends on
}

// PreTUIStartHook allows final configuration before TUI displays.
// Fires after verify phase, before StartTUI().
type PreTUIStartHook interface {
	// BeforeStart is called just before TUI boots.
	// Returns error to abort TUI start.
	BeforeStart(ctx context.Context) error
}

// TUIHooks combines all TUI interaction hooks.
// TUI adapters implement this; core calls hooks at appropriate points.
type TUIHooks interface {
	CommandSelectionHook
	UoWDataHook
	PreTUIStartHook
	ExitHoldController
}

// --- Null Implementations ---
// Used for console-only mode - never blocks, never modifies.

// NullExitHoldController never holds. Used for console-only mode.
type NullExitHoldController struct{}

// HoldExit returns a no-op release function.
func (NullExitHoldController) HoldExit() func() { return func() {} }

// WaitForRelease returns immediately with true.
func (NullExitHoldController) WaitForRelease(context.Context, time.Duration) bool { return true }

// NullCommandSelectionHook returns original command unchanged.
type NullCommandSelectionHook struct{}

// SelectCommand returns the original command without modification.
func (NullCommandSelectionHook) SelectCommand(_ context.Context, req CommandSelectionRequest) CommandSelectionResponse {
	return CommandSelectionResponse{SelectedCommand: req.OriginalCommand}
}

// NullUoWDataHook discards data (console-only mode).
type NullUoWDataHook struct{}

// ReceiveUoWs does nothing.
func (NullUoWDataHook) ReceiveUoWs(UoWData) {}

// NullPreTUIStartHook does nothing before start.
type NullPreTUIStartHook struct{}

// BeforeStart returns nil (no error).
func (NullPreTUIStartHook) BeforeStart(context.Context) error { return nil }

// NullTUIHooks provides default no-op implementations for all hooks.
// Used for console-only mode - never blocks, never modifies.
type NullTUIHooks struct {
	NullCommandSelectionHook
	NullUoWDataHook
	NullPreTUIStartHook
	NullExitHoldController
}

// Compile-time interface checks
var (
	_ ExitHoldController    = NullExitHoldController{}
	_ CommandSelectionHook  = NullCommandSelectionHook{}
	_ UoWDataHook           = NullUoWDataHook{}
	_ PreTUIStartHook       = NullPreTUIStartHook{}
	_ TUIHooks              = NullTUIHooks{}
)
