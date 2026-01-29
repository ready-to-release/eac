package defaulttui

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/tui"
)

func TestInit_RegistersDefault(t *testing.T) {
	// The init() function should have registered the default "*"
	bindings := tui.ListBindings()

	found := false
	for _, b := range bindings {
		if b == "*" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected '*' (default) to be registered")
	}
}

func TestInit_UnknownCommandUsesDefault(t *testing.T) {
	// Any unknown command should get the default TUI
	console := tui.NewForCommand("unknown-command", tui.Config{})
	if console == nil {
		t.Fatal("Expected console, got nil")
	}

	// Verify it's the default TUI
	if _, ok := console.(*Console); !ok {
		t.Error("Unknown command should use default TUI")
	}
}
