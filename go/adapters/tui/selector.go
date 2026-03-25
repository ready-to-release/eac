package tui

import (
	"context"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/clibase/registry"
)

// CommandOption represents a selectable command in the Selector TUI.
// This is the new interface for subcommand selection, replacing the bloated
// InteractiveConsole pattern.
type CommandOption struct {
	Name        string   // Command name (e.g., "create", "merge")
	Description string   // Brief description
	Examples    []string // Optional usage examples
}

// --- Selector Factory Registry ---
// Follows the same pattern as the parallel TUI registry.
// Selector implementations register themselves via init().

var (
	globalSelectorFactory SelectorFactory
	selectorMu            sync.RWMutex
)

// SelectorConsole is a minimal TUI for subcommand selection.
// It shows options, user picks one, TUI exits, caller executes.
//
// Key design principles:
//   - NO command execution inside the TUI
//   - NO output viewport or subprocess spawning
//   - Just: show list → user picks → return selection → exit
//
// This separates the concerns: TUI handles presentation, caller handles execution.
type SelectorConsole interface {
	// SetCommands configures the available commands to select from.
	SetCommands(commands []CommandOption)

	// Run shows the selector UI and blocks until user makes a choice or cancels.
	// Returns:
	//   - selected: the command name the user chose (empty if cancelled)
	//   - args: any additional arguments the user entered
	//   - cancelled: true if user pressed Esc/q, false if they made a selection
	Run(ctx context.Context) (selected string, args string, cancelled bool)
}

// SelectorFactory creates SelectorConsole instances.
type SelectorFactory func() SelectorConsole

// SubcommandToOption converts a SubcommandInfo to a CommandOption.
// This provides backwards compatibility with existing command definitions.
func SubcommandToOption(sub SubcommandInfo) CommandOption {
	return CommandOption{
		Name:        sub.Name,
		Description: sub.Description,
		// SubcommandInfo doesn't have Examples, so leave empty
	}
}

// SubcommandsToOptions converts a slice of SubcommandInfo to CommandOptions.
func SubcommandsToOptions(subs []SubcommandInfo) []CommandOption {
	opts := make([]CommandOption, len(subs))
	for i, sub := range subs {
		opts[i] = SubcommandToOption(sub)
	}
	return opts
}

// SubcommandsFromRegistry creates CommandOptions from registry for a parent command.
// This replaces the hardcoded `var subcommands = []tui.SubcommandInfo{...}` pattern.
// Uses SubcommandEntries for correct labels when aliases are present.
func SubcommandsFromRegistry(parentName string) []CommandOption {
	entries := registry.Global().SubcommandEntries(parentName)
	opts := make([]CommandOption, len(entries))

	for i, entry := range entries {
		// Extract just the subcommand name (remove parent prefix)
		label := strings.TrimPrefix(entry.Key, parentName+" ")
		opts[i] = CommandOption{
			Name:        label,
			Description: entry.Cmd.Metadata().Short,
		}
	}

	return opts
}

// RegisterSelector sets the global selector factory.
// Called from init() in selector implementation packages.
// Thread-safe for concurrent access.
func RegisterSelector(factory SelectorFactory) {
	selectorMu.Lock()
	defer selectorMu.Unlock()
	globalSelectorFactory = factory
}

// NewSelector returns a SelectorConsole using the registered factory.
// Panics if no factory has been registered - call HasSelector() first
// if you need to check availability.
func NewSelector() SelectorConsole {
	selectorMu.RLock()
	defer selectorMu.RUnlock()
	if globalSelectorFactory == nil {
		panic("tui: no selector factory registered - import the selector package with blank import")
	}
	return globalSelectorFactory()
}

// HasSelector returns true if a selector factory has been registered.
// Use this to check before calling NewSelector() if you need graceful fallback.
func HasSelector() bool {
	selectorMu.RLock()
	defer selectorMu.RUnlock()
	return globalSelectorFactory != nil
}
