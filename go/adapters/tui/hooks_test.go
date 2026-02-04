package tui

import (
	"context"
	"testing"
	"time"

	"github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces"
)

func TestTUIHooksImpl_SelectCommand_NoOptions_ReturnsOriginal(t *testing.T) {
	hooks := NewTUIHooks(nil) // Console can be nil for this test

	req := interfaces.CommandSelectionRequest{
		OriginalCommand: "test-command",
		Options:         nil, // No options
	}

	resp := hooks.SelectCommand(context.Background(), req)

	if resp.SelectedCommand != "test-command" {
		t.Errorf("expected SelectedCommand=%q, got %q", "test-command", resp.SelectedCommand)
	}
	if resp.Cancelled {
		t.Error("expected Cancelled=false")
	}
}

func TestTUIHooksImpl_SelectCommand_EmptyOptions_ReturnsOriginal(t *testing.T) {
	hooks := NewTUIHooks(nil)

	req := interfaces.CommandSelectionRequest{
		OriginalCommand: "work",
		Options:         []interfaces.CommandOption{}, // Empty slice
	}

	resp := hooks.SelectCommand(context.Background(), req)

	if resp.SelectedCommand != "work" {
		t.Errorf("expected SelectedCommand=%q, got %q", "work", resp.SelectedCommand)
	}
}

func TestTUIHooksImpl_ReceiveUoWs_StoresData(t *testing.T) {
	hooks := NewTUIHooks(nil)

	data := interfaces.UoWData{
		Modules: []interfaces.UoWModule{
			{
				Name: "test-module",
				Units: []interfaces.UoWUnit{
					{ID: "build:test:go:go", DisplayName: "test:go", Weight: 1},
				},
			},
		},
	}

	hooks.ReceiveUoWs(data)

	// Verify data is stored
	stored := hooks.GetUoWData()
	if len(stored.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(stored.Modules))
	}
	if stored.Modules[0].Name != "test-module" {
		t.Errorf("expected module name %q, got %q", "test-module", stored.Modules[0].Name)
	}
}

func TestTUIHooksImpl_BeforeStart_ReturnsNil(t *testing.T) {
	hooks := NewTUIHooks(nil)

	err := hooks.BeforeStart(context.Background())
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestTUIHooksImpl_HoldExit_DelegatesToController(t *testing.T) {
	hooks := NewTUIHooks(nil)

	// HoldExit should create a hold
	release := hooks.HoldExit()

	// WaitForRelease with very short timeout should fail (hold is active)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	result := hooks.WaitForRelease(ctx, 10*time.Millisecond)
	if result {
		t.Error("expected WaitForRelease to return false while hold is active")
	}

	// Release the hold
	release()

	// Now WaitForRelease should succeed
	result = hooks.WaitForRelease(context.Background(), 100*time.Millisecond)
	if !result {
		t.Error("expected WaitForRelease to return true after release")
	}
}

func TestTUIHooksImpl_WaitForRelease_NoHolds(t *testing.T) {
	hooks := NewTUIHooks(nil)

	// Should return immediately when no holds
	start := time.Now()
	result := hooks.WaitForRelease(context.Background(), 5*time.Second)
	elapsed := time.Since(start)

	if !result {
		t.Error("expected WaitForRelease to return true")
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("expected immediate return, took %v", elapsed)
	}
}

func TestTUIHooksImpl_ConcurrentReceiveUoWs(t *testing.T) {
	hooks := NewTUIHooks(nil)

	// Concurrent calls should not race
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func(idx int) {
			data := interfaces.UoWData{
				Modules: []interfaces.UoWModule{
					{Name: "module-" + string(rune('a'+idx))},
				},
			}
			hooks.ReceiveUoWs(data)
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have some data stored (doesn't matter which one wins)
	stored := hooks.GetUoWData()
	if len(stored.Modules) != 1 {
		t.Errorf("expected 1 module, got %d", len(stored.Modules))
	}
}

// Test interface compliance
func TestTUIHooksImpl_ImplementsTUIHooks(t *testing.T) {
	hooks := NewTUIHooks(nil)

	// Should implement TUIHooks interface
	var _ interfaces.TUIHooks = hooks
}

func TestTUIHooksImpl_SelectCommand_WithOptionsNoSelector_ReturnsOriginal(t *testing.T) {
	// Clear any registered selector
	resetSelectorRegistry()

	hooks := NewTUIHooks(nil)

	req := interfaces.CommandSelectionRequest{
		OriginalCommand: "work",
		Options: []interfaces.CommandOption{
			{Name: "create", Description: "Create a new work item"},
			{Name: "merge", Description: "Merge work items"},
		},
	}

	resp := hooks.SelectCommand(context.Background(), req)

	// When no selector is registered, should return original command
	if resp.SelectedCommand != "work" {
		t.Errorf("expected SelectedCommand=%q, got %q", "work", resp.SelectedCommand)
	}
	if resp.Cancelled {
		t.Error("expected Cancelled=false")
	}
}

func TestTUIHooksImpl_SelectCommand_WithOptionsAndMockSelector(t *testing.T) {
	// Clear and register a mock selector
	resetSelectorRegistry()

	// Register mock that selects first option
	RegisterSelector(func() SelectorConsole {
		return &mockSelectingConsole{selectIndex: 0}
	})
	defer resetSelectorRegistry()

	hooks := NewTUIHooks(nil)

	req := interfaces.CommandSelectionRequest{
		OriginalCommand: "work",
		Options: []interfaces.CommandOption{
			{Name: "create", Description: "Create a new work item"},
			{Name: "merge", Description: "Merge work items"},
		},
	}

	resp := hooks.SelectCommand(context.Background(), req)

	// Mock selector should return first option
	if resp.SelectedCommand != "create" {
		t.Errorf("expected SelectedCommand=%q, got %q", "create", resp.SelectedCommand)
	}
	if resp.Cancelled {
		t.Error("expected Cancelled=false")
	}
}

func TestTUIHooksImpl_SelectCommand_MockSelectorCancels(t *testing.T) {
	resetSelectorRegistry()

	// Register mock that cancels
	RegisterSelector(func() SelectorConsole {
		return &mockCancellingConsole{}
	})
	defer resetSelectorRegistry()

	hooks := NewTUIHooks(nil)

	req := interfaces.CommandSelectionRequest{
		OriginalCommand: "work",
		Options: []interfaces.CommandOption{
			{Name: "create", Description: "Create"},
		},
	}

	resp := hooks.SelectCommand(context.Background(), req)

	if resp.SelectedCommand != "" {
		t.Errorf("expected empty SelectedCommand on cancel, got %q", resp.SelectedCommand)
	}
	if !resp.Cancelled {
		t.Error("expected Cancelled=true")
	}
}

func TestTUIHooksImpl_SelectCommand_MockSelectorWithArgs(t *testing.T) {
	resetSelectorRegistry()

	// Register mock that returns args
	RegisterSelector(func() SelectorConsole {
		return &mockSelectingWithArgsConsole{selectIndex: 0, args: "--verbose"}
	})
	defer resetSelectorRegistry()

	hooks := NewTUIHooks(nil)

	req := interfaces.CommandSelectionRequest{
		OriginalCommand: "work",
		Options: []interfaces.CommandOption{
			{Name: "create", Description: "Create"},
		},
	}

	resp := hooks.SelectCommand(context.Background(), req)

	if resp.SelectedCommand != "create" {
		t.Errorf("expected SelectedCommand=%q, got %q", "create", resp.SelectedCommand)
	}
	if resp.Args != "--verbose" {
		t.Errorf("expected Args=%q, got %q", "--verbose", resp.Args)
	}
}

// Mock consoles for testing

type mockSelectingConsole struct {
	commands    []CommandOption
	selectIndex int
}

func (m *mockSelectingConsole) SetCommands(commands []CommandOption) {
	m.commands = commands
}

func (m *mockSelectingConsole) Run(_ context.Context) (string, string, bool) {
	if len(m.commands) > m.selectIndex {
		return m.commands[m.selectIndex].Name, "", false
	}
	return "", "", true
}

type mockCancellingConsole struct {
	commands []CommandOption
}

func (m *mockCancellingConsole) SetCommands(commands []CommandOption) {
	m.commands = commands
}

func (m *mockCancellingConsole) Run(_ context.Context) (string, string, bool) {
	return "", "", true // cancelled
}

type mockSelectingWithArgsConsole struct {
	commands    []CommandOption
	selectIndex int
	args        string
}

func (m *mockSelectingWithArgsConsole) SetCommands(commands []CommandOption) {
	m.commands = commands
}

func (m *mockSelectingWithArgsConsole) Run(_ context.Context) (string, string, bool) {
	if len(m.commands) > m.selectIndex {
		return m.commands[m.selectIndex].Name, m.args, false
	}
	return "", "", true
}
