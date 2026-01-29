package defaulttui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ready-to-release/eac/go/eac/commands/tui"
)

func TestModel_Init(t *testing.T) {
	m := NewModel(tui.Config{}, nil)
	cmd := m.Init()
	// Init returns a batch command for text input blink, tick, and docker fetch
	if cmd == nil {
		t.Error("Init should return batch command")
	}
}

func TestModel_Update_Navigation(t *testing.T) {
	commands := []tui.SubcommandInfo{
		{Name: "lint", Description: "Lint"},
		{Name: "docs", Description: "Docs"},
		{Name: "test", Description: "Test"},
	}
	m := NewModel(tui.Config{}, commands)

	// Initial cursor position
	if m.cursor != 0 {
		t.Errorf("Initial cursor should be 0, got %d", m.cursor)
	}

	// Move down with 'j'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("After 'j', cursor should be 1, got %d", m.cursor)
	}

	// Move down with arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	if m.cursor != 2 {
		t.Errorf("After down arrow, cursor should be 2, got %d", m.cursor)
	}

	// Move down at end (should stay)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(Model)
	if m.cursor != 2 {
		t.Errorf("At end, cursor should stay at 2, got %d", m.cursor)
	}

	// Move up with 'k'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = newModel.(Model)
	if m.cursor != 1 {
		t.Errorf("After 'k', cursor should be 1, got %d", m.cursor)
	}

	// Move up with arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("After up arrow, cursor should be 0, got %d", m.cursor)
	}

	// Move up at start (should stay)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = newModel.(Model)
	if m.cursor != 0 {
		t.Errorf("At start, cursor should stay at 0, got %d", m.cursor)
	}
}

func TestModel_Update_Selection(t *testing.T) {
	commands := []tui.SubcommandInfo{
		{Name: "lint", Description: "Lint"},
		{Name: "docs", Description: "Docs"},
	}
	m := NewModel(tui.Config{CommandName: "update"}, commands)

	// Initial selection
	if m.selected != -1 {
		t.Errorf("Initial selected should be -1, got %d", m.selected)
	}

	// Move to second item and select
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = newModel.(Model)
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	m = newModel.(Model)
	if m.selected != 1 {
		t.Errorf("After enter, selected should be 1, got %d", m.selected)
	}

	// Should be running (not quitting) after enter
	if !m.running {
		t.Error("Enter should set running to true")
	}

	// Should return a command to run
	if cmd == nil {
		t.Error("Enter should return command to run")
	}
}

func TestModel_Update_Quit(t *testing.T) {
	m := NewModel(tui.Config{}, nil)

	// Test 'q' to quit
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = newModel.(Model)
	if !m.quitting {
		t.Error("'q' should set quitting to true")
	}
	if cmd == nil {
		t.Error("'q' should return quit command")
	}

	// Test 'esc' to quit
	m = NewModel(tui.Config{}, nil)
	newModel, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(Model)
	if !m.quitting {
		t.Error("'esc' should set quitting to true")
	}
	if cmd == nil {
		t.Error("'esc' should return quit command")
	}
}

func TestModel_View_WithCommands(t *testing.T) {
	commands := []tui.SubcommandInfo{
		{Name: "lint", Description: "Update lint results"},
		{Name: "docs", Description: "Update documentation"},
	}
	m := NewModel(tui.Config{CommandName: "update"}, commands)

	view := m.View()

	// Check title
	if !strings.Contains(view, "update") {
		t.Error("View should contain command name 'update'")
	}

	// Check subcommands appear in the tree
	if !strings.Contains(view, "lint") {
		t.Error("View should contain 'lint'")
	}
	if !strings.Contains(view, "docs") {
		t.Error("View should contain 'docs'")
	}

	// Check footer
	if !strings.Contains(view, "Enter: run") && !strings.Contains(view, "quit") {
		t.Error("View should contain navigation instructions")
	}
}

func TestModel_View_NoCommands(t *testing.T) {
	m := NewModel(tui.Config{}, nil)

	view := m.View()

	if !strings.Contains(view, "No subcommands available") {
		t.Error("View should indicate no subcommands when empty")
	}
}

func TestModel_View_Quitting(t *testing.T) {
	m := NewModel(tui.Config{}, nil)
	m.quitting = true

	view := m.View()

	if view != "" {
		t.Errorf("Quitting view should be empty, got %q", view)
	}
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := NewModel(tui.Config{}, nil)

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newModel.(Model)

	if m.width != 100 {
		t.Errorf("Width should be 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Errorf("Height should be 50, got %d", m.height)
	}
}

func TestModel_Update_FocusCycling(t *testing.T) {
	m := NewModel(tui.Config{}, []tui.SubcommandInfo{{Name: "test"}})

	// Initial focus is on tree (0)
	if m.focusedField != 0 {
		t.Errorf("Initial focus should be 0, got %d", m.focusedField)
	}

	// Tab to text input (1)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)
	if m.focusedField != 1 {
		t.Errorf("After tab, focus should be 1, got %d", m.focusedField)
	}

	// Tab to execute button (2)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)
	if m.focusedField != 2 {
		t.Errorf("After second tab, focus should be 2, got %d", m.focusedField)
	}

	// Tab back to tree (0)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)
	if m.focusedField != 0 {
		t.Errorf("After third tab, focus should cycle to 0, got %d", m.focusedField)
	}

	// Shift+tab to execute button (2)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = newModel.(Model)
	if m.focusedField != 2 {
		t.Errorf("After shift+tab, focus should be 2, got %d", m.focusedField)
	}
}

func TestModel_View_SplitPane(t *testing.T) {
	commands := []tui.SubcommandInfo{
		{Name: "build", Description: "Build modules"},
	}
	m := NewModel(tui.Config{CommandName: "test"}, commands)
	m.width = 80
	m.height = 24

	view := m.View()

	// Check for pane headers
	if !strings.Contains(view, "Commands") {
		t.Error("View should contain 'Commands' pane header")
	}
	if !strings.Contains(view, "Output") {
		t.Error("View should contain 'Output' pane header")
	}

	// Check for resource row
	if !strings.Contains(view, "CPU:") {
		t.Error("View should contain CPU metrics")
	}
	if !strings.Contains(view, "Mem:") {
		t.Error("View should contain Memory metrics")
	}
}

func TestModel_GetSelectedSubcommand(t *testing.T) {
	commands := []tui.SubcommandInfo{
		{Name: "lint"},
		{Name: "docs"},
	}
	m := NewModel(tui.Config{}, commands)

	// No selection
	if got := m.GetSelectedSubcommand(); got != "" {
		t.Errorf("No selection should return empty, got %q", got)
	}

	// With selection
	m.selected = 1
	if got := m.GetSelectedSubcommand(); got != "docs" {
		t.Errorf("Selected should be 'docs', got %q", got)
	}
}

func TestModel_DockerContainersUpdate(t *testing.T) {
	m := NewModel(tui.Config{}, nil)

	// Simulate receiving docker containers
	containers := []string{"nginx", "postgres", "redis"}
	newModel, _ := m.Update(dockerContainersMsg(containers))
	m = newModel.(Model)

	if len(m.dockerContainers) != 3 {
		t.Errorf("Should have 3 containers, got %d", len(m.dockerContainers))
	}
	if m.dockerContainers[0] != "nginx" {
		t.Errorf("First container should be nginx, got %s", m.dockerContainers[0])
	}
}
