package tui

import (
	"context"
	"testing"
)

func TestRegisterSelector(t *testing.T) {
	// Clear any existing registration first
	resetSelectorRegistry()

	// Before registration, HasSelector should be false
	if HasSelector() {
		t.Error("HasSelector() should be false before registration")
	}

	// Create a mock factory
	mockCalled := false
	mockFactory := func() SelectorConsole {
		mockCalled = true
		return &mockSelectorConsole{}
	}

	// Register it
	RegisterSelector(mockFactory)

	// HasSelector should now be true
	if !HasSelector() {
		t.Error("HasSelector() should be true after registration")
	}

	// NewSelector should return the mock
	selector := NewSelector()
	if selector == nil {
		t.Fatal("NewSelector() returned nil")
	}
	if !mockCalled {
		t.Error("Factory was not called")
	}
}

func TestNewSelectorPanicsWithoutRegistration(t *testing.T) {
	// Clear registration
	resetSelectorRegistry()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewSelector() should panic when no factory registered")
		}
	}()

	_ = NewSelector()
}

func TestHasSelectorReturnsFalseWhenNotRegistered(t *testing.T) {
	resetSelectorRegistry()

	if HasSelector() {
		t.Error("HasSelector() should return false when no factory registered")
	}
}

// mockSelectorConsole is a minimal mock for testing.
type mockSelectorConsole struct {
	commands []CommandOption
}

func (m *mockSelectorConsole) SetCommands(commands []CommandOption) {
	m.commands = commands
}

func (m *mockSelectorConsole) Run(_ context.Context) (string, string, bool) {
	if len(m.commands) > 0 {
		return m.commands[0].Name, "", false
	}
	return "", "", true
}

// resetSelectorRegistry clears the global factory for test isolation.
// This function is exported for testing but implementation is in selector.go.
func resetSelectorRegistry() {
	selectorMu.Lock()
	defer selectorMu.Unlock()
	globalSelectorFactory = nil
}
