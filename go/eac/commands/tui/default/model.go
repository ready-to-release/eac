package defaulttui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/ready-to-release/eac/go/eac/commands/tui"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/net/context"
)

// tickMsg is sent periodically for time and metrics updates.
type tickMsg time.Time

// commandOutputMsg is sent when command output is ready.
type commandOutputMsg struct {
	output string
	err    error
}

// dockerContainersMsg is sent when docker container list is ready.
type dockerContainersMsg []string

// Model is the Bubbletea model for the default TUI.
type Model struct {
	config      tui.Config
	subcommands []tui.SubcommandInfo

	// Text input for command editing
	textInput textinput.Model

	// Output viewport
	outputViewport viewport.Model
	outputContent  string

	// Navigation
	cursor       int
	selected     int  // -1 if nothing selected yet
	focusedField int  // 0 = subcommand tree, 1 = text input, 2 = execute button

	// Display
	width  int
	height int

	// System metrics (cached, updated periodically)
	cachedCPUPercent   []float64
	cachedMemPercent   float64
	dockerContainers   []string
	lastMetricsUpdate  time.Time
	lastDockerUpdate   time.Time

	// State
	quitting   bool
	executing  bool // True when user pressed execute
	running    bool // True when command is currently running
	startTime  time.Time
	asciiMode  bool
}

// NewModel creates a new default TUI model.
func NewModel(config tui.Config, subcommands []tui.SubcommandInfo) Model {
	ti := textinput.New()
	ti.Placeholder = "command arguments..."
	ti.CharLimit = 256
	ti.Width = 40

	vp := viewport.New(40, 10)
	vp.SetContent("Output will appear here...")
	vp.MouseWheelEnabled = true

	return Model{
		config:      config,
		subcommands: subcommands,
		textInput:   ti,
		outputViewport: vp,
		selected:    -1,
		width:       80,
		height:      config.Height,
		startTime:   time.Now(),
		asciiMode:   config.ASCIIMode,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.tickCmd(),
		m.fetchDockerContainers(),
	)
}

// tickCmd returns a command that sends a tick message.
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchDockerContainers returns a command that lists running docker containers.
func (m Model) fetchDockerContainers() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return dockerContainersMsg{}
		}
		defer cli.Close()

		containers, err := cli.ContainerList(ctx, container.ListOptions{})
		if err != nil {
			return dockerContainersMsg{}
		}

		var names []string
		for _, c := range containers {
			if len(c.Names) > 0 {
				name := strings.TrimPrefix(c.Names[0], "/")
				names = append(names, name)
			}
		}
		return dockerContainersMsg(names)
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't process keys while running
		if m.running {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			// Cycle focus: tree -> text input -> execute button
			m.focusedField = (m.focusedField + 1) % 3
			if m.focusedField == 1 {
				m.textInput.Focus()
			} else {
				m.textInput.Blur()
			}

		case "shift+tab":
			// Reverse cycle
			m.focusedField = (m.focusedField + 2) % 3
			if m.focusedField == 1 {
				m.textInput.Focus()
			} else {
				m.textInput.Blur()
			}

		case "up", "k":
			if m.focusedField == 0 && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.focusedField == 0 && m.cursor < len(m.subcommands)-1 {
				m.cursor++
			}

		case "enter":
			if m.focusedField == 2 || m.focusedField == 0 {
				// Execute selected command and show output
				if len(m.subcommands) > 0 {
					m.selected = m.cursor
					m.running = true
					m.outputContent = "Running..."

					// Build the command to run
					cmdName := m.config.CommandName + " " + m.subcommands[m.cursor].Name
					args := m.textInput.Value()

					return m, runCommand(cmdName, args)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width/2 - 20

		// Update viewport size (right pane)
		vpWidth := msg.Width/2 - 4
		vpHeight := msg.Height - 12 // Account for resources, command row, and borders
		if vpHeight < 3 {
			vpHeight = 3
		}
		m.outputViewport.Width = vpWidth
		m.outputViewport.Height = vpHeight

	case tickMsg:
		// Update system metrics
		m.updateMetrics()
		cmds = append(cmds, m.tickCmd())

		// Refresh docker containers every 5 seconds
		if time.Since(m.lastDockerUpdate) >= 5*time.Second {
			cmds = append(cmds, m.fetchDockerContainers())
		}

	case dockerContainersMsg:
		m.dockerContainers = msg
		m.lastDockerUpdate = time.Now()

	case commandOutputMsg:
		m.running = false
		if msg.err != nil {
			m.outputContent = fmt.Sprintf("Error: %v\n\n%s", msg.err, msg.output)
		} else {
			m.outputContent = msg.output
		}
		m.outputViewport.SetContent(m.outputContent)
	}

	// Update text input if focused
	if m.focusedField == 1 {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport for scrolling
	var cmd tea.Cmd
	m.outputViewport, cmd = m.outputViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateMetrics updates cached CPU and memory metrics.
func (m *Model) updateMetrics() {
	// Only update every second
	if time.Since(m.lastMetricsUpdate) < time.Second {
		return
	}
	m.lastMetricsUpdate = time.Now()

	// Get CPU percentages per core (non-blocking with 0 interval uses cached values)
	if percents, err := cpu.Percent(0, true); err == nil {
		m.cachedCPUPercent = percents
	}

	// Get memory usage
	if memInfo, err := mem.VirtualMemory(); err == nil {
		m.cachedMemPercent = memInfo.UsedPercent
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	dimStyle := lipgloss.NewStyle().Faint(true)
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("255"))
	buttonFocusedStyle := buttonStyle.
		Background(lipgloss.Color("86")).
		Bold(true)

	// ═══════════════════════════════════════════════════════════════════════════
	// ROW 1: Title
	// ═══════════════════════════════════════════════════════════════════════════
	title := m.config.CommandName
	if title == "" {
		title = "EAC Command"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════════════════
	// ROW 2: Resources (Timer | CPU | Memory | Docker)
	// ═══════════════════════════════════════════════════════════════════════════
	b.WriteString(m.renderResourcesRow(white, dimStyle))
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════════════════
	// ROW 3: Blank row (spacing)
	// ═══════════════════════════════════════════════════════════════════════════
	b.WriteString("\n")

	// ═══════════════════════════════════════════════════════════════════════════
	// ROW 4: Command Editor with Execute Button
	// ═══════════════════════════════════════════════════════════════════════════
	cmdLabel := "Command: "
	if m.focusedField == 1 {
		cmdLabel = selectedStyle.Render("Command: ")
	} else {
		cmdLabel = white.Render("Command: ")
	}

	// Show selected subcommand in the text input placeholder
	if len(m.subcommands) > 0 && m.cursor < len(m.subcommands) {
		m.textInput.Placeholder = m.subcommands[m.cursor].Name + " [args...]"
	}

	// Execute button
	var buttonText string
	if m.focusedField == 2 {
		buttonText = buttonFocusedStyle.Render(" Execute ")
	} else {
		buttonText = buttonStyle.Render(" Execute ")
	}

	commandRow := cmdLabel + m.textInput.View() + "  " + buttonText
	b.WriteString(commandRow)
	b.WriteString("\n\n")

	// ═══════════════════════════════════════════════════════════════════════════
	// ROWS 5+: Split Pane (Command Tree | Output)
	// ═══════════════════════════════════════════════════════════════════════════
	leftWidth := m.width/2 - 2
	rightWidth := m.width/2 - 2
	paneHeight := m.height - 10
	if paneHeight < 5 {
		paneHeight = 5
	}

	// Left pane: Command Tree
	leftPane := m.renderCommandTree(leftWidth, paneHeight, selectedStyle, dimStyle, borderStyle)

	// Right pane: Output
	rightPane := m.renderOutputPane(rightWidth, paneHeight, dimStyle, borderStyle)

	// Join panes side by side
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane))
	b.WriteString("\n")

	// ═══════════════════════════════════════════════════════════════════════════
	// Footer: Navigation hints
	// ═══════════════════════════════════════════════════════════════════════════
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Tab: focus  ↑↓/jk: navigate  Enter: run  Scroll: mouse wheel  q/Esc: quit"))

	return b.String()
}

// renderResourcesRow renders the resources row with CPU, Memory, and Docker info.
func (m Model) renderResourcesRow(white, dimStyle lipgloss.Style) string {
	sep := dimStyle.Render(" │ ")

	// Timer
	elapsed := time.Since(m.startTime)
	timerStr := dimStyle.Render(fmt.Sprintf("%02d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60))

	// CPU dots
	cpuStr := white.Render("CPU: ") + m.renderCPUDots()

	// Memory dots
	memStr := white.Render("Mem: ") + m.renderMemDots()

	// Docker containers
	var dockerStr string
	if len(m.dockerContainers) > 0 {
		dockerStr = sep + white.Render("Docker: ") + m.renderDockerDots()
	}

	return timerStr + sep + cpuStr + sep + memStr + dockerStr
}

// renderCPUDots returns colored dots representing per-core CPU usage.
func (m Model) renderCPUDots() string {
	perCore := m.cachedCPUPercent
	if len(perCore) == 0 {
		return "----"
	}

	var dots strings.Builder
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	for _, pct := range perCore {
		switch {
		case pct < 5:
			dots.WriteString(dim.Render(emptyChar))
		case pct < 50:
			dots.WriteString(green.Render(filledChar))
		case pct < 80:
			dots.WriteString(yellow.Render(filledChar))
		case pct < 95:
			dots.WriteString(orange.Render(filledChar))
		default:
			dots.WriteString(red.Render(filledChar))
		}
	}
	return dots.String()
}

// renderMemDots returns colored dots representing memory usage.
func (m Model) renderMemDots() string {
	pct := m.cachedMemPercent
	if pct == 0 && m.lastMetricsUpdate.IsZero() {
		return "----"
	}

	numDots := 10
	usedDots := int(pct / 10)
	if usedDots > numDots {
		usedDots = numDots
	}

	var dots strings.Builder
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("71"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	orange := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	filledChar, emptyChar := "●", "○"
	if m.asciiMode {
		filledChar, emptyChar = "*", "o"
	}

	for i := 0; i < numDots; i++ {
		if i < usedDots {
			switch {
			case i < 3:
				dots.WriteString(green.Render(filledChar))
			case i < 6:
				dots.WriteString(yellow.Render(filledChar))
			case i < 9:
				dots.WriteString(orange.Render(filledChar))
			default:
				dots.WriteString(red.Render(filledChar))
			}
		} else {
			dots.WriteString(dim.Render(emptyChar))
		}
	}

	dots.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf(" %.0f%%", pct)))
	return dots.String()
}

// renderDockerDots returns colored dots for running Docker containers.
func (m Model) renderDockerDots() string {
	if len(m.dockerContainers) == 0 {
		return ""
	}

	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
	filledChar := "●"
	if m.asciiMode {
		filledChar = "*"
	}

	var dots strings.Builder
	maxDots := 8
	for i, name := range m.dockerContainers {
		if i >= maxDots {
			dots.WriteString(fmt.Sprintf("+%d", len(m.dockerContainers)-maxDots))
			break
		}
		dots.WriteString(cyan.Render(filledChar))
		_ = name // Could add tooltip later
	}

	return dots.String() + lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf(" %d", len(m.dockerContainers)))
}

// renderCommandTree renders the left pane with the command tree.
func (m Model) renderCommandTree(width, height int, selectedStyle, dimStyle, borderStyle lipgloss.Style) string {
	var b strings.Builder

	// Header
	header := "┌─ Commands "
	headerLine := header + borderStyle.Render(strings.Repeat("─", width-lipgloss.Width(header)-1)) + "┐"
	b.WriteString(headerLine)
	b.WriteString("\n")

	// Content
	contentHeight := height - 2 // Account for header and footer
	if contentHeight < 1 {
		contentHeight = 1
	}

	if len(m.subcommands) == 0 {
		line := "│ " + dimStyle.Render("No subcommands available")
		padding := width - lipgloss.Width(line) - 1
		if padding < 0 {
			padding = 0
		}
		b.WriteString(line + strings.Repeat(" ", padding) + "│\n")
	} else {
		// Calculate visible range for scrolling
		startIdx := 0
		if m.cursor >= contentHeight {
			startIdx = m.cursor - contentHeight + 1
		}
		endIdx := startIdx + contentHeight
		if endIdx > len(m.subcommands) {
			endIdx = len(m.subcommands)
		}

		for i := startIdx; i < endIdx; i++ {
			cmd := m.subcommands[i]
			prefix := "  "
			if m.cursor == i && m.focusedField == 0 {
				prefix = "> "
			}

			nameStyle := lipgloss.NewStyle()
			if m.cursor == i {
				nameStyle = selectedStyle
			}

			line := "│ " + prefix + nameStyle.Render(cmd.Name)
			if cmd.Description != "" {
				descWidth := width - lipgloss.Width(line) - 5
				if descWidth > 0 {
					desc := cmd.Description
					if len(desc) > descWidth {
						desc = desc[:descWidth-3] + "..."
					}
					line += " " + dimStyle.Render(desc)
				}
			}

			padding := width - lipgloss.Width(line) - 1
			if padding < 0 {
				padding = 0
			}
			b.WriteString(line + strings.Repeat(" ", padding) + "│\n")
		}

		// Fill remaining height with empty lines
		for i := endIdx - startIdx; i < contentHeight; i++ {
			b.WriteString("│" + strings.Repeat(" ", width-2) + "│\n")
		}
	}

	// Footer
	footer := "└" + borderStyle.Render(strings.Repeat("─", width-2)) + "┘"
	b.WriteString(footer)

	return b.String()
}

// renderOutputPane renders the right pane with command output.
func (m Model) renderOutputPane(width, height int, dimStyle, borderStyle lipgloss.Style) string {
	var b strings.Builder

	// Header
	header := "┌─ Output "
	if m.running {
		header = "┌─ Output (running...) "
	}
	headerLine := header + borderStyle.Render(strings.Repeat("─", width-lipgloss.Width(header)-1)) + "┐"
	b.WriteString(headerLine)
	b.WriteString("\n")

	// Content height for viewport
	contentHeight := height - 2
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Get viewport content (supports scrolling)
	vpContent := m.outputViewport.View()
	vpLines := strings.Split(vpContent, "\n")

	// Render each line with border
	for i := 0; i < contentHeight; i++ {
		var line string
		if i < len(vpLines) {
			line = vpLines[i]
			// Truncate if too long
			if lipgloss.Width(line) > width-4 {
				runes := []rune(line)
				if len(runes) > width-7 {
					line = string(runes[:width-7]) + "..."
				}
			}
		}
		paddedLine := "│ " + line
		padding := width - lipgloss.Width(paddedLine) - 1
		if padding < 0 {
			padding = 0
		}
		b.WriteString(paddedLine + strings.Repeat(" ", padding) + "│\n")
	}

	// Footer with scroll indicator
	scrollInfo := ""
	totalLines := strings.Count(m.outputContent, "\n") + 1
	if totalLines > contentHeight {
		percent := int(m.outputViewport.ScrollPercent() * 100)
		scrollInfo = fmt.Sprintf(" [%d%%] ", percent)
	}
	footerPadding := width - 2 - len(scrollInfo)
	if footerPadding < 0 {
		footerPadding = 0
	}
	footer := "└" + borderStyle.Render(strings.Repeat("─", footerPadding)) + dimStyle.Render(scrollInfo) + "┘"
	b.WriteString(footer)

	return b.String()
}

// runCommand executes a command and returns its output.
func runCommand(cmdName string, args string) tea.Cmd {
	return func() tea.Msg {
		fullCmd := cmdName
		if args != "" {
			fullCmd = cmdName + " " + args
		}

		cmd := exec.Command("sh", "-c", fullCmd)
		output, err := cmd.CombinedOutput()
		return commandOutputMsg{
			output: string(output),
			err:    err,
		}
	}
}

// GetSelectedSubcommand returns the selected subcommand name.
func (m Model) GetSelectedSubcommand() string {
	if m.selected >= 0 && m.selected < len(m.subcommands) {
		return m.subcommands[m.selected].Name
	}
	return ""
}

// GetArguments returns the entered arguments.
func (m Model) GetArguments() string {
	return m.textInput.Value()
}

// IsExecuting returns true if user pressed execute.
func (m Model) IsExecuting() bool {
	return m.executing
}
