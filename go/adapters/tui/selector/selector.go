// Package selector provides a minimal TUI for subcommand selection.
// It shows a list of commands, user picks one, and the TUI exits.
// NO command execution happens inside this TUI - that's the caller's job.
package selector

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ready-to-release/eac/go/adapters/tui"
)

// Model is the Bubbletea model for the selector TUI.
type Model struct {
	commands []tui.CommandOption

	// Text input for additional arguments
	textInput textinput.Model

	// Navigation
	cursor int

	// Result
	selected  string // Set when user confirms selection
	cancelled bool   // Set when user cancels

	// Display
	width  int
	height int

	// Focus: 0 = command list, 1 = args input
	focus int
}

// NewModel creates a new selector model.
func NewModel(commands []tui.CommandOption) Model {
	ti := textinput.New()
	ti.Placeholder = "additional arguments..."
	ti.CharLimit = 256
	ti.Width = 40

	return Model{
		commands:  commands,
		textInput: ti,
		width:     80,
		height:    24,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 20
		if m.textInput.Width < 20 {
			m.textInput.Width = 20
		}
	}

	// Update text input if focused
	if m.focus == 1 {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKey processes keyboard input.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.cancelled = true
		return m, tea.Quit

	case tea.KeyEsc:
		m.cancelled = true
		return m, tea.Quit

	case tea.KeyEnter:
		if len(m.commands) == 0 {
			m.cancelled = true
			return m, tea.Quit
		}
		m.selected = m.commands[m.cursor].Name
		return m, tea.Quit

	case tea.KeyTab:
		// Toggle focus between command list and args input
		m.focus = (m.focus + 1) % 2
		if m.focus == 1 {
			return m, m.textInput.Focus()
		}
		m.textInput.Blur()
		return m, nil

	case tea.KeyUp:
		if m.focus == 0 && m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyDown:
		if m.focus == 0 && m.cursor < len(m.commands)-1 {
			m.cursor++
		}
		return m, nil

	case tea.KeyRunes:
		// Handle vim-style navigation when not in text input
		if m.focus == 0 {
			switch string(msg.Runes) {
			case "j":
				if m.cursor < len(m.commands)-1 {
					m.cursor++
				}
				return m, nil
			case "k":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			case "q":
				m.cancelled = true
				return m, tea.Quit
			}
		}
	}

	// If focus is on text input, let it handle the key
	if m.focus == 1 {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.cancelled || m.selected != "" {
		return ""
	}

	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	dimStyle := lipgloss.NewStyle().Faint(true)
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("86"))

	// Title
	b.WriteString(titleStyle.Render("Select a command"))
	b.WriteString("\n\n")

	// Command list
	for i, cmd := range m.commands {
		cursor := "  "
		nameStyle := lipgloss.NewStyle()

		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
			nameStyle = selectedStyle
		}

		b.WriteString(cursor)
		b.WriteString(nameStyle.Render(cmd.Name))

		if cmd.Description != "" {
			b.WriteString(" ")
			b.WriteString(dimStyle.Render("- " + cmd.Description))
		}
		b.WriteString("\n")
	}

	if len(m.commands) == 0 {
		b.WriteString(dimStyle.Render("  No commands available"))
		b.WriteString("\n")
	}

	// Args input
	b.WriteString("\n")
	argsLabel := "Args: "
	if m.focus == 1 {
		argsLabel = selectedStyle.Render("Args: ")
	}
	b.WriteString(argsLabel)
	b.WriteString(m.textInput.View())
	b.WriteString("\n")

	// Help
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("↑↓/jk: navigate • Tab: focus args • Enter: select • Esc/q: cancel"))

	return b.String()
}

// Selected returns the selected command name (empty if cancelled).
func (m Model) Selected() string {
	return m.selected
}

// Args returns the entered arguments.
func (m Model) Args() string {
	return m.textInput.Value()
}

// Cancelled returns true if the user cancelled the selection.
func (m Model) Cancelled() bool {
	return m.cancelled
}

// Console implements tui.SelectorConsole.
type Console struct {
	commands []tui.CommandOption

	// Result from Run
	selected  string
	args      string
	cancelled bool
}

// New creates a new Selector Console.
func New() *Console {
	return &Console{}
}

// SetCommands implements tui.SelectorConsole.
func (c *Console) SetCommands(commands []tui.CommandOption) {
	c.commands = commands
}

// Run implements tui.SelectorConsole.
// Shows the selector UI and blocks until user makes a choice or cancels.
func (c *Console) Run(ctx context.Context) (selected string, args string, cancelled bool) {
	model := NewModel(c.commands)

	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)

	// Run with context cancellation support
	go func() {
		<-ctx.Done()
		p.Quit()
	}()

	finalModel, err := p.Run()
	if err != nil {
		// On error, treat as cancelled
		return "", "", true
	}

	m, ok := finalModel.(Model)
	if !ok {
		return "", "", true
	}
	c.selected = m.Selected()
	c.args = m.Args()
	c.cancelled = m.Cancelled()

	return c.selected, c.args, c.cancelled
}

// Factory returns a tui.SelectorFactory for creating Selector instances.
func Factory() tui.SelectorFactory {
	return func() tui.SelectorConsole {
		return New()
	}
}

// Compile-time check that Console implements tui.SelectorConsole.
var _ tui.SelectorConsole = (*Console)(nil)

// RunSelector is a convenience function to run the selector directly.
// Returns (selected command, args, cancelled).
func RunSelector(ctx context.Context, commands []tui.CommandOption) (string, string, bool) {
	c := New()
	c.SetCommands(commands)
	return c.Run(ctx)
}

// RunSelectorWithSubcommands is a convenience function using SubcommandInfo.
func RunSelectorWithSubcommands(ctx context.Context, subs []tui.SubcommandInfo) (string, string, bool) {
	return RunSelector(ctx, tui.SubcommandsToOptions(subs))
}

// ShouldUseSelector returns true if the selector should be shown.
// This checks if we're in an interactive terminal and not in CI.
func ShouldUseSelector() bool {
	return tui.ShouldUseTUI("", false, false)
}

// PrintSelection prints a formatted message about the selection.
// Useful for debugging or verbose mode.
func PrintSelection(selected, args string, cancelled bool) string {
	if cancelled {
		return "Selection cancelled"
	}
	if args != "" {
		return fmt.Sprintf("Selected: %s %s", selected, args)
	}
	return fmt.Sprintf("Selected: %s", selected)
}
