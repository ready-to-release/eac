// Package shared provides common types and utilities for TUI components.
package shared

import tea "github.com/charmbracelet/bubbletea"

// Component is the interface for renderable TUI components.
type Component interface {
	// Render returns the component's view at the given dimensions.
	Render(width, height int) string
}

// Interactive extends Component with input handling.
type Interactive interface {
	Component
	// Update processes messages and returns updated state and command.
	Update(msg tea.Msg) (Interactive, tea.Cmd)
}

// Focusable extends Interactive with focus management.
type Focusable interface {
	Interactive
	// Focus sets whether the component has keyboard focus.
	Focus(focused bool)
	// IsFocused returns true if the component has keyboard focus.
	IsFocused() bool
}
