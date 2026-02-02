package preflight

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/adapters/tui/console/components/shared"
)

// Phase represents the execution phase.
type Phase int

const (
	PhaseInit Phase = iota
	PhaseRun
	PhaseSummary
)

// String returns the phase name.
func (p Phase) String() string {
	switch p {
	case PhaseInit:
		return "Init"
	case PhaseRun:
		return "Run"
	case PhaseSummary:
		return "Summary"
	default:
		return "Unknown"
	}
}

// Icon returns an ASCII-safe icon for the phase.
func (p Phase) Icon() string {
	switch p {
	case PhaseInit:
		return ">"
	case PhaseRun:
		return ">"
	case PhaseSummary:
		return "V"
	default:
		return "?"
	}
}

// Header displays the pre-flight header with phase and timer.
type Header struct {
	phase     Phase
	phaseName string
	startTime time.Time
	paused    bool
}

// NewHeader creates a new header component.
func NewHeader() *Header {
	return &Header{
		phase:     PhaseInit,
		startTime: time.Now(),
	}
}

// SetPhase updates the current phase.
func (h *Header) SetPhase(phase Phase, name string) {
	h.phase = phase
	h.phaseName = name
	if h.startTime.IsZero() {
		h.startTime = time.Now()
	}
}

// SetPaused updates the paused state.
func (h *Header) SetPaused(paused bool) {
	h.paused = paused
}

// SetStartTime updates the start time for the timer.
func (h *Header) SetStartTime(t time.Time) {
	h.startTime = t
}

// Render renders the header.
func (h *Header) Render(width, height int) string {
	// Phase icon and name
	icon := h.phase.Icon()
	name := h.phaseName
	if name == "" {
		name = h.phase.String()
	}

	phaseStyle := lipgloss.NewStyle().Bold(true).Foreground(shared.ColorCyan)
	phaseStr := phaseStyle.Render(icon + " " + name)

	// Timer
	elapsed := time.Since(h.startTime)
	timerStr := shared.FormatDuration(elapsed)
	timerStyle := lipgloss.NewStyle().Foreground(shared.ColorGrey)
	timer := timerStyle.Render(timerStr)

	// Paused indicator
	pausedStr := ""
	if h.paused {
		pausedStyle := lipgloss.NewStyle().Bold(true).Foreground(shared.ColorYellow)
		pausedStr = " " + pausedStyle.Render("[PAUSED]")
	}

	// Build the header line
	left := phaseStr + pausedStr
	right := timer

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := width - leftWidth - rightWidth
	if spacing < 1 {
		spacing = 1
	}

	line := left + strings.Repeat(" ", spacing) + right

	// Add border
	border := shared.BorderStyle.Render(strings.Repeat("-", width))
	return line + "\n" + border
}

// RenderWithProgress renders the header with a progress indicator.
func (h *Header) RenderWithProgress(width, height int, completed, total int) string {
	// Phase icon and name
	icon := h.phase.Icon()
	name := h.phaseName
	if name == "" {
		name = h.phase.String()
	}

	phaseStyle := lipgloss.NewStyle().Bold(true).Foreground(shared.ColorCyan)
	phaseStr := phaseStyle.Render(icon + " " + name)

	// Progress
	progressStyle := lipgloss.NewStyle().Foreground(shared.ColorGreen)
	progressStr := progressStyle.Render(fmt.Sprintf("%d/%d", completed, total))

	// Timer
	elapsed := time.Since(h.startTime)
	timerStr := shared.FormatDuration(elapsed)
	timerStyle := lipgloss.NewStyle().Foreground(shared.ColorGrey)
	timer := timerStyle.Render(timerStr)

	// Paused indicator
	pausedStr := ""
	if h.paused {
		pausedStyle := lipgloss.NewStyle().Bold(true).Foreground(shared.ColorYellow)
		pausedStr = " " + pausedStyle.Render("[PAUSED]")
	}

	// Build the header line
	left := phaseStr + " " + progressStr + pausedStr
	right := timer

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := width - leftWidth - rightWidth
	if spacing < 1 {
		spacing = 1
	}

	line := left + strings.Repeat(" ", spacing) + right

	// Add border
	border := shared.BorderStyle.Render(strings.Repeat("-", width))
	return line + "\n" + border
}
