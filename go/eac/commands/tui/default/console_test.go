package defaulttui

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/tui"
)

func TestConsole_ImplementsInterface(t *testing.T) {
	var _ tui.Console = (*Console)(nil)
}

func TestConsole_Factory(t *testing.T) {
	factory := Factory()
	config := tui.Config{
		Height:      40,
		CommandName: "update",
	}

	console := factory(config)
	if console == nil {
		t.Fatal("Factory returned nil console")
	}

	if _, ok := console.(*Console); !ok {
		t.Error("Factory did not return a *Console")
	}
}

func TestConsole_SetSubcommands(t *testing.T) {
	c := New(tui.Config{})

	commands := []tui.SubcommandInfo{
		{Name: "lint", Description: "Update lint results"},
		{Name: "docs", Description: "Update documentation"},
	}

	c.SetSubcommands(commands)

	// Verify subcommands were set
	if len(c.subcommands) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(c.subcommands))
	}
}

func TestConsole_GetSelectedCommand_NoSelection(t *testing.T) {
	c := New(tui.Config{})

	cmd, params := c.GetSelectedCommand()
	if cmd != "" {
		t.Errorf("Expected empty command, got %q", cmd)
	}
	if params != nil {
		t.Errorf("Expected nil params, got %v", params)
	}
}

func TestConsole_GetSelectedCommand_WithSelection(t *testing.T) {
	c := New(tui.Config{})

	commands := []tui.SubcommandInfo{
		{Name: "lint", Description: "Update lint results"},
		{Name: "docs", Description: "Update documentation"},
	}
	c.SetSubcommands(commands)
	c.selected = 1 // Select "docs"

	cmd, _ := c.GetSelectedCommand()
	if cmd != "docs" {
		t.Errorf("Expected 'docs', got %q", cmd)
	}
}

// Test that no-op methods don't panic
func TestConsole_NoOpMethods(t *testing.T) {
	c := New(tui.Config{})

	// These should all be no-ops for the default TUI
	c.SendLine(tui.Line{Text: "test"})
	c.WriteResult("test")
	c.UpdateStatus(tui.Status{})
	c.SetPhase(tui.PhaseInit)
	c.CompletePhase(tui.PhaseInit, true, "done")
	c.WriteToPhase(tui.PhaseInit, "test")
	c.SetPhaseSummary(tui.PhaseInit, "summary")
	c.StartModule("test", 1)
	c.MarkModuleRunning("test")
	c.MarkModuleComplete("test", 0)
	c.SendSummary(&tui.SummaryData{})
	c.SetInitSummary(&tui.InitSummary{})
}
