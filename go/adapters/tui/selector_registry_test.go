package tui

import (
	"context"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/registry"
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

// testCommandPort is a minimal CommandPort for testing.
type testCommandPort struct {
	name  string
	short string
}

func (m *testCommandPort) Name() string { return m.name }
func (m *testCommandPort) Metadata() core.CommandMetadata {
	return core.CommandMetadata{Short: m.short}
}

// TestSubcommandsFromRegistry tests the SubcommandsFromRegistry function.
func TestSubcommandsFromRegistry(t *testing.T) {
	// Create a test registry with known commands
	testReg := registry.NewCommandRegistry()
	testReg.MustRegister(&testCommandPort{name: "test", short: "Test parent command"})
	testReg.MustRegister(&testCommandPort{name: "test alpha", short: "Alpha subcommand"})
	testReg.MustRegister(&testCommandPort{name: "test beta", short: "Beta subcommand"})
	testReg.MustRegister(&testCommandPort{name: "test gamma", short: "Gamma subcommand"})

	// Set as global so SubcommandsFromRegistry can find it
	oldGlobal := registry.Global()
	registry.SetGlobal(testReg)
	defer registry.SetGlobal(oldGlobal)

	t.Run("converts registry to CommandOptions", func(t *testing.T) {
		opts := SubcommandsFromRegistry("test")
		if len(opts) != 3 {
			t.Errorf("Expected 3 options, got %d", len(opts))
		}
	})

	t.Run("options have correct names", func(t *testing.T) {
		opts := SubcommandsFromRegistry("test")
		names := make(map[string]bool)
		for _, opt := range opts {
			names[opt.Name] = true
		}

		expected := []string{"alpha", "beta", "gamma"}
		for _, exp := range expected {
			if !names[exp] {
				t.Errorf("Missing expected option: %q", exp)
			}
		}
	})

	t.Run("options have descriptions", func(t *testing.T) {
		opts := SubcommandsFromRegistry("test")
		for _, opt := range opts {
			if opt.Description == "" {
				t.Errorf("Option %q has empty description", opt.Name)
			}
		}
	})

	t.Run("returns empty for nonexistent parent", func(t *testing.T) {
		opts := SubcommandsFromRegistry("nonexistent")
		if len(opts) != 0 {
			t.Errorf("Expected 0 options for nonexistent parent, got %d", len(opts))
		}
	})
}
