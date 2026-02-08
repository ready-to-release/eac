package tui

import (
	"github.com/ready-to-release/eac/go/clibase/display"
)

func init() {
	display.SetTUIBootstrap(&display.TUIBootstrap{
		NewForCommand: newForCommandAdapter,
		NewObserver:   newObserverAdapter,
		NewHooks:      newHooksAdapter,
		UnwrapConsole: unwrapConsoleAdapter,
	})
}

// newForCommandAdapter wraps the global registry's NewForCommand.
// It converts display.Config to tui Config (which is the contract Config)
// and returns a display.Console (which the contract Console satisfies).
func newForCommandAdapter(commandPath string, cfg display.Config) display.Console {
	tuiCfg := Config{
		Height:       cfg.Height,
		BufferSize:   cfg.BufferSize,
		ASCIIMode:    cfg.ASCIIMode,
		TUI3Demo:     cfg.TUI3Demo,
		SkipTUIDelay: cfg.SkipTUIDelay,
		RunPhaseName: cfg.RunPhaseName,
		CommandName:  cfg.CommandName,
		ActionType:   cfg.ActionType,
	}
	console := NewForCommand(commandPath, tuiCfg)
	return &consoleAdapter{inner: console}
}

// newObserverAdapter creates a TUIObserver from a display.Console.
// Returns (ExecutionObserver, WriterFactory) as interface{}.
func newObserverAdapter(console display.Console) (observer interface{}, writerFactory interface{}) {
	// Unwrap to get the underlying tui.Console
	inner := unwrapToTUIConsole(console)
	if inner == nil {
		return nil, nil
	}
	obs := NewTUIObserver(inner)
	return obs, obs
}

// newHooksAdapter creates TUIHooks from a display.Console.
// Returns interfaces.TUIHooks as interface{}.
func newHooksAdapter(console display.Console) interface{} {
	inner := unwrapToTUIConsole(console)
	if inner == nil {
		return nil
	}
	return NewTUIHooks(inner)
}

// unwrapConsoleAdapter extracts the inner console from the adapter.
// This is used by the orchestrator when it needs the actual console implementation.
func unwrapConsoleAdapter(console display.Console) display.Console {
	// If it's our adapter, return it as-is (the adapter IS the usable console)
	return console
}

// unwrapToTUIConsole extracts the underlying tui.Console from a display.Console.
func unwrapToTUIConsole(console display.Console) Console {
	if ca, ok := console.(*consoleAdapter); ok {
		return ca.inner
	}
	return nil
}
