package selector

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ready-to-release/eac/go/adapters/tui"
)

func TestTextInputReceivesFocusCommand(t *testing.T) {
	model := NewModel([]tui.CommandOption{
		{Name: "test", Description: "Test command"},
	})

	// Simulate Tab to focus text input
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	m := updated.(Model)

	// Focus should be on text input
	if m.focus != 1 {
		t.Errorf("expected focus=1 (text input), got %d", m.focus)
	}

	// CRITICAL: Focus() must return a non-nil cmd for text input to work
	if cmd == nil {
		t.Error("Focus() must return a tea.Cmd for text input to receive keyboard events")
	}
}

func TestTextInputBlurReturnsNil(t *testing.T) {
	model := NewModel([]tui.CommandOption{
		{Name: "test", Description: "Test command"},
	})

	// Tab to focus
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	m := updated.(Model)

	// Tab again to blur
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)

	// Focus should be back on command list
	if m.focus != 0 {
		t.Errorf("expected focus=0 (command list), got %d", m.focus)
	}

	// Blur doesn't need a command
	if cmd != nil {
		t.Log("Blur returned a command (acceptable but not required)")
	}
}

func TestNavigationInCommandList(t *testing.T) {
	model := NewModel([]tui.CommandOption{
		{Name: "first", Description: "First"},
		{Name: "second", Description: "Second"},
		{Name: "third", Description: "Third"},
	})

	// Initial cursor at 0
	if model.cursor != 0 {
		t.Errorf("expected initial cursor=0, got %d", model.cursor)
	}

	// Down arrow
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after down, got %d", m.cursor)
	}

	// j key (vim style)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = updated.(Model)
	if m.cursor != 2 {
		t.Errorf("expected cursor=2 after j, got %d", m.cursor)
	}

	// Up arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1 after up, got %d", m.cursor)
	}

	// k key (vim style)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor=0 after k, got %d", m.cursor)
	}
}

func TestSelectionAndCancellation(t *testing.T) {
	model := NewModel([]tui.CommandOption{
		{Name: "test-cmd", Description: "Test"},
	})

	// Enter selects
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)

	if m.Selected() != "test-cmd" {
		t.Errorf("expected selected='test-cmd', got %q", m.Selected())
	}
	if cmd == nil {
		t.Error("Enter should return tea.Quit")
	}

	// Test cancellation with Esc
	model = NewModel([]tui.CommandOption{
		{Name: "test-cmd", Description: "Test"},
	})
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if !m.Cancelled() {
		t.Error("expected Cancelled()=true after Esc")
	}
	if cmd == nil {
		t.Error("Esc should return tea.Quit")
	}

	// Test cancellation with q
	model = NewModel([]tui.CommandOption{
		{Name: "test-cmd", Description: "Test"},
	})
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	m = updated.(Model)

	if !m.Cancelled() {
		t.Error("expected Cancelled()=true after q")
	}
}
