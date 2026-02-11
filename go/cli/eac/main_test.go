//go:build L1 && ov
// +build L1,ov

// Feature: cli_command_routing
// Unit tests for command routing and registration
package main

import (
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/clibase/registry"
)

func TestCommandRegistryExists(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry should not be nil")
	}
}

func TestCommandsRegistered(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}
	expectedCommands := []string{
		"show modules",
		"show files",
		"show files-changed",
		"show files-staged",
	}

	for _, cmdName := range expectedCommands {
		if _, ok := reg.Get(cmdName); !ok {
			t.Errorf("expected command '%s' to be registered", cmdName)
		}
	}
}

func TestGetSubcommands(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}

	// Test that "show" parent returns subcommands
	subs := reg.Subcommands("show")

	if len(subs) == 0 {
		t.Error("expected 'show' to have subcommands")
	}

	// Check for expected subcommands
	expectedSubs := []string{"files", "modules", "dependencies"}
	for _, expected := range expectedSubs {
		found := false
		for _, sub := range subs {
			name := strings.TrimPrefix(sub.Name(), "show ")
			if name == expected {
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
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}

	// Leaf commands should have no subcommands
	subs := reg.Subcommands("show modules")

	if len(subs) != 0 {
		t.Errorf("expected 'show modules' to have no subcommands, got %d", len(subs))
	}
}

func TestGetSubcommandsSorted(t *testing.T) {
	reg := registry.Global()
	if reg == nil {
		t.Fatal("global registry is nil")
	}

	subs := reg.Subcommands("show")

	// Verify alphabetical sorting
	for i := 1; i < len(subs); i++ {
		if subs[i-1].Name() > subs[i].Name() {
			t.Errorf("subcommands not sorted: '%s' should come before '%s'",
				subs[i-1].Name(), subs[i].Name())
		}
	}
}
