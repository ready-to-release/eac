package tui

import (
	"context"
	"sync"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// TUIHooksImpl implements core.TUIHooks for interactive TUI mode.
// It bridges between the command framework and TUI components.
type TUIHooksImpl struct {
	console  Console
	exitHold *ExitHoldController
	uowData  core.UoWData
	mu       sync.RWMutex
}

// NewTUIHooks creates a new TUIHooksImpl.
// The console parameter can be nil for testing or non-interactive use.
func NewTUIHooks(console Console) *TUIHooksImpl {
	return &TUIHooksImpl{
		console:  console,
		exitHold: NewExitHoldController(),
	}
}

// SelectCommand implements core.CommandSelectionHook.
// If no options are provided, returns the original command unchanged.
// If a selector is registered and interactive terminal, shows selection UI.
// Otherwise, returns the original command.
func (h *TUIHooksImpl) SelectCommand(ctx context.Context, req core.CommandSelectionRequest) core.CommandSelectionResponse {
	// If no options, return original command unchanged
	if len(req.Options) == 0 {
		return core.CommandSelectionResponse{
			SelectedCommand: req.OriginalCommand,
		}
	}

	// Check if selector is registered
	if !HasSelector() {
		return core.CommandSelectionResponse{
			SelectedCommand: req.OriginalCommand,
		}
	}

	// Create and run selector
	sel := NewSelector()
	sel.SetCommands(interfaceOptionsToTUI(req.Options))
	selected, args, cancelled := sel.Run(ctx)

	return core.CommandSelectionResponse{
		SelectedCommand: selected,
		Args:            args,
		Cancelled:       cancelled,
	}
}

// interfaceOptionsToTUI converts contract CommandOptions to TUI CommandOptions.
func interfaceOptionsToTUI(opts []core.CommandOption) []CommandOption {
	result := make([]CommandOption, len(opts))
	for i, opt := range opts {
		result[i] = CommandOption{
			Name:        opt.Name,
			Description: opt.Description,
			// Examples not in core.CommandOption, leave nil
		}
	}
	return result
}

// ReceiveUoWs implements core.UoWDataHook.
// Stores the UoW data for the TUI to use for visualization.
// Note: Pre-population of tabs happens via SetInitSummary -> InitSummaryMsg handler,
// which processes the ExecutionTree. This hook stores data for potential future use.
func (h *TUIHooksImpl) ReceiveUoWs(data core.UoWData) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.uowData = data
}

// GetUoWData returns the stored UoW data.
// Thread-safe for concurrent access.
func (h *TUIHooksImpl) GetUoWData() core.UoWData {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.uowData
}

// BeforeStart implements core.PreTUIStartHook.
// Called just before the TUI starts; returns nil to proceed.
func (h *TUIHooksImpl) BeforeStart(_ context.Context) error {
	// No special setup needed currently
	return nil
}

// HoldExit implements core.ExitHoldController.
// Signals that exit should be delayed (user is interacting).
func (h *TUIHooksImpl) HoldExit() func() {
	return h.exitHold.HoldExit()
}

// WaitForRelease implements core.ExitHoldController.
// Blocks until all holds are released or timeout.
func (h *TUIHooksImpl) WaitForRelease(ctx context.Context, timeout time.Duration) bool {
	return h.exitHold.WaitForRelease(ctx, timeout)
}

// GetExitHoldController returns the underlying exit hold controller.
// Useful for passing to TUI components that need to manage holds.
func (h *TUIHooksImpl) GetExitHoldController() *ExitHoldController {
	return h.exitHold
}

// Compile-time interface check
var _ core.TUIHooks = (*TUIHooksImpl)(nil)
