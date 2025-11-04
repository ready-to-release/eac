// Feature: cli_command_routing
// Unit tests for command routing and registration
package main

import (
	"strings"
	"testing"
)

func TestCommandsMapExists(t *testing.T) {
	if commands == nil {
		t.Fatal("commands map should not be nil")
	}
}

func TestCommandsRegistered(t *testing.T) {
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
	expectedSubs := []string{"modules", "files", "moduletypes"}
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

func TestGetSubcommandsForNestedCommand(t *testing.T) {
	// "show files" has subcommands: "changed", "staged"
	subcommands := getSubcommands("show files")

	expectedSubs := []string{"changed", "staged"}
	for _, expected := range expectedSubs {
		found := false
		for _, sub := range subcommands {
			if sub == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected 'show files' to have subcommand '%s'", expected)
		}
	}
}

func TestGetSubcommandsDeduplication(t *testing.T) {
	// Ensure no duplicate subcommands in results
	subcommands := getSubcommands("show")

	seen := make(map[string]bool)
	for _, sub := range subcommands {
		if seen[sub] {
			t.Errorf("duplicate subcommand found: '%s'", sub)
		}
		seen[sub] = true
	}
}

func TestRegisterFunction(t *testing.T) {
	// Test the Register function
	initialCount := len(commands)

	testFunc := func() int { return 0 }
	Register("test-command", testFunc)

	if len(commands) != initialCount+1 {
		t.Error("Register should add command to map")
	}

	if _, exists := commands["test-command"]; !exists {
		t.Error("registered command should exist in map")
	}

	// Cleanup
	delete(commands, "test-command")
}

func TestRegisterOverwrite(t *testing.T) {
	// Test that registering same command twice overwrites
	func1 := func() int { return 1 }
	func2 := func() int { return 2 }

	Register("test-overwrite", func1)
	Register("test-overwrite", func2)

	result := commands["test-overwrite"]()
	if result != 2 {
		t.Error("second registration should overwrite first")
	}

	// Cleanup
	delete(commands, "test-overwrite")
}

func TestCommandNamesAreLowercase(t *testing.T) {
	// Verify convention that command names use lowercase
	for cmdName := range commands {
		if strings.ToLower(cmdName) != cmdName {
			t.Errorf("command name should be lowercase: '%s'", cmdName)
		}
	}
}

func TestCommandNamesUseSpaces(t *testing.T) {
	// Verify multi-word commands use spaces, not hyphens (except root commands)
	for cmdName := range commands {
		if strings.Contains(cmdName, " ") {
			// Multi-word command
			if strings.Contains(cmdName, "_") {
				t.Errorf("multi-word command should use spaces, not underscores: '%s'", cmdName)
			}
		}
	}
}

func TestAllRegisteredCommandsAreCallable(t *testing.T) {
	// Verify all registered commands have callable functions
	for cmdName, cmdFunc := range commands {
		if cmdFunc == nil {
			t.Errorf("command '%s' has nil function", cmdName)
		}

		// Attempt to call (in test environment, may fail but shouldn't panic)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("command '%s' panicked: %v", cmdName, r)
				}
			}()

			// Call the command (it will likely fail in test environment, but shouldn't panic)
			_ = cmdFunc()
		}()
	}
}
