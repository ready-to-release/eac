// Feature: cli_command_routing
// Unit tests for command routing and registration
package main

import (
	"testing"

	"github.com/ready-to-release/eac/src/commands/registry"
)

func TestCommandsMapExists(t *testing.T) {
	commands := registry.GetCommands()
	if commands == nil {
		t.Fatal("commands map should not be nil")
	}
}

func TestCommandsRegistered(t *testing.T) {
	commands := registry.GetCommands()
	expectedCommands := []string{
		"list commands",
		"describe commands",
		"show modules",
		"show files",
		"show files changed",
		"show files staged",
		"show moduletypes",
		"commit-ai",
	}

	for _, cmdName := range expectedCommands {
		if _, exists := commands[cmdName]; !exists {
			t.Errorf("expected command '%s' to be registered", cmdName)
		}
	}
}

func TestGetSubcommands(t *testing.T) {
	// Test that "show" parent returns subcommands
	subcommands := getSubcommands("show")

	if len(subcommands) == 0 {
		t.Error("expected 'show' to have subcommands")
	}

	// Check for expected subcommands
	expectedSubs := []string{"files", "modules", "moduletypes", "dependencies"}
	for _, expected := range expectedSubs {
		found := false
		for _, sub := range subcommands {
			if sub == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected 'show' to have subcommand '%s'", expected)
		}
	}
}

func TestGetSubcommandsReturnsEmpty(t *testing.T) {
	// Leaf commands should have no subcommands
	subcommands := getSubcommands("commit-ai")

	if len(subcommands) != 0 {
		t.Errorf("expected 'commit-ai' to have no subcommands, got %d", len(subcommands))
	}
}

func TestGetSubcommandsSorted(t *testing.T) {
	subcommands := getSubcommands("show")

	// Verify alphabetical sorting
	for i := 1; i < len(subcommands); i++ {
		if subcommands[i-1] > subcommands[i] {
			t.Errorf("subcommands not sorted: '%s' should come before '%s'",
				subcommands[i-1], subcommands[i])
		}
	}
}
