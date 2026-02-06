package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderInitAnimatedStatus renders the animated "Initializing..." dots line
// for the logs panel during the init phase.
func (m Model) renderInitAnimatedStatus(innerWidth int) string {
	elapsed := time.Since(m.startTime)
	dotCount := int(elapsed.Seconds()*2) % 4
	var dots string
	if m.asciiMode {
		dots = strings.Repeat(".", dotCount+1)
	} else {
		dots = strings.Repeat("·", dotCount+1)
	}

	statusText := m.getWaitingForLocksMessage()
	if statusText == "" {
		statusText = "Initializing"
	}

	icon := "▶"
	if m.asciiMode {
		icon = ">"
	}
	animLine := Styles.Running.Render(icon) + " " + Styles.Phase.Render(statusText) + Styles.Dim.Render(dots)
	animPad := innerWidth - lipgloss.Width(animLine)
	if animPad < 0 {
		animPad = 0
	}
	return animLine + strings.Repeat(" ", animPad)
}

// getWaitingForLocksMessage returns a message if any locks are being waited on.
// Returns empty string if no locks are waiting.
func (m Model) getWaitingForLocksMessage() string {
	var waitingLocks []string
	for _, lock := range m.locks {
		if lock.Waiting > 0 {
			// Extract just the identifier from "type:identifier" format
			name := lock.Name
			if idx := strings.Index(name, ":"); idx >= 0 {
				name = name[idx+1:]
			}
			waitingLocks = append(waitingLocks, name)
		}
	}

	if len(waitingLocks) == 0 {
		return ""
	}

	if len(waitingLocks) == 1 {
		return fmt.Sprintf("Waiting for lock: %s", waitingLocks[0])
	}
	return fmt.Sprintf("Waiting for locks: %s", strings.Join(waitingLocks, ", "))
}
