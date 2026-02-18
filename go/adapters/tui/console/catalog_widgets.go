package console

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/ready-to-release/eac/go/adapters/tui/console/render"
)

// RegisterAllWidgets populates the catalog with all known widgets.
// Called once during NewModel().
func RegisterAllWidgets(catalog *WidgetCatalog) {

	// === Status Bar Widgets (view_summary.go) ===

	catalog.Register(&Widget{
		ID:          "res-cpu",
		ElementName: "CPU",
		HelpText:    "Per-core CPU pressure (16 lamps); green <50%, yellow 50-80%, orange 80-95%, red >95%",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			lamps := render.RenderCPULamps(snap.CPUPercent, snap.AsciiMode)
			if len(snap.CPUPercent) == 0 {
				return Styles.Dim.Render("CPU:") + lamps
			}
			totalCores := len(snap.CPUPercent)
			activeCores := 0
			var sum float64
			for _, pct := range snap.CPUPercent {
				sum += pct
				if pct > 5.0 {
					activeCores++
				}
			}
			avgPct := sum / float64(totalCores)
			white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			color := render.GetPressureColor(activeCores, totalCores)
			return white.Render("CPU:") + lamps + " " + color.Render(fmt.Sprintf("%3.0f%%", avgPct))
		},
	})

	catalog.Register(&Widget{
		ID:          "res-slots",
		ElementName: "Slots",
		HelpText:    "Scheduler slots: H=host D=docker (used/capacity); wN = waiting jobs (yellow 1-5, red >5 = backpressure)",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

			// Host slots
			hostTarget := snap.HostPressureTarget
			if hostTarget == 0 {
				hostTarget = snap.HostRoof
			}
			hostColor := getPressureColor(snap.HostRunning, hostTarget)
			hostStr := white.Render("H:") + hostColor.Render(fmt.Sprintf("%d/%d", snap.HostRunning, hostTarget))
			if snap.HostWaiting > 0 {
				wStyle := yellow
				if snap.HostWaiting > 5 {
					wStyle = red
				}
				hostStr += wStyle.Render(fmt.Sprintf("w%d", snap.HostWaiting))
			}

			// Docker slots
			dockerRoof := snap.DockerRoof
			if dockerRoof == 0 {
				dockerRoof = 1
			}
			dockerPressureTarget := snap.DockerPressureTarget
			if dockerPressureTarget == 0 {
				dockerPressureTarget = dockerRoof
			}
			dockerColor := getPressureColor(snap.DockerRunning, dockerPressureTarget)
			dockerStr := white.Render("D:") + dockerColor.Render(fmt.Sprintf("%d/%d", snap.DockerRunning, dockerRoof))
			if snap.DockerWaiting > 0 {
				wStyle := yellow
				if snap.DockerWaiting > 5 {
					wStyle = red
				}
				dockerStr += wStyle.Render(fmt.Sprintf("w%d", snap.DockerWaiting))
			}

			return hostStr + " " + dockerStr
		},
	})

	catalog.Register(&Widget{
		ID:          "res-counters",
		ElementName: "Counters",
		HelpText:    "Progress: blue=cached (higher=good incremental), green=done, red=failed",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			blueBg := lipgloss.NewStyle().Background(lipgloss.Color("39")).Foreground(lipgloss.Color("0"))
			greenBg := lipgloss.NewStyle().Background(lipgloss.Color("40")).Foreground(lipgloss.Color("0"))
			redBg := lipgloss.NewStyle().Background(lipgloss.Color("196")).Foreground(lipgloss.Color("0"))
			return blueBg.Render(fmt.Sprintf("%3d", snap.Counts.Cached)) + " " +
				greenBg.Render(fmt.Sprintf("%3d", snap.Counts.Done)) + " " +
				redBg.Render(fmt.Sprintf("%3d", snap.Counts.Failed))
		},
	})

	catalog.Register(&Widget{
		ID:          "progress-count",
		ElementName: "Progress Count",
		HelpText:    "Finalized / total units of work (cached + done + failed = finalized)",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			white := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
			finalized := snap.Counts.Done + snap.Counts.Cached + snap.Counts.Failed
			totalStr := fmt.Sprintf("%d", snap.UoWTotal)
			totalWidth := len(totalStr)
			return white.Render(fmt.Sprintf("%*d/%s", totalWidth, finalized, totalStr))
		},
	})

	catalog.Register(&Widget{
		ID:          "status-text",
		ElementName: "Status",
		HelpText:    "Current execution status: waiting, running with phase name and elapsed time, or final result",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			if snap.SummaryData != nil {
				icon := "✓"
				label := "Complete"
				style := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				if !snap.SummaryData.Success {
					icon = "✗"
					label = "Failed"
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
				}
				result := style.Render(fmt.Sprintf("%s %s", icon, label)) +
					" " + Styles.Dim.Render(fmt.Sprintf("(%s)", formatElapsed(snap.SummaryData.TotalTime)))
				if snap.SummaryData.RunSummary != "" {
					result += " " + Styles.Dim.Render(snap.SummaryData.RunSummary)
				}
				return result
			}
			if snap.RunPhaseActive {
				runName := snap.RunPhaseName
				if runName == "" {
					runName = "Running"
				}
				return Styles.Running.Render("▶") + " " + Styles.Phase.Render(runName) +
					" " + Styles.Time.Render(formatElapsed(snap.Elapsed))
			}
			return Styles.Dim.Render("waiting...")
		},
	})

	// === Resources Pane Widgets (view_resources.go) ===

	catalog.Register(&Widget{
		ID:          "res-mem",
		ElementName: "Memory",
		HelpText:    "Host RAM usage: green <37% ok, yellow 37-68% normal, orange 68-87% pressure, red >87% constrained (risk of swapping)",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			if snap.MemPercent == 0 && snap.LastMetricsUpdate.IsZero() {
				return render.RenderEmptyLamps(render.LampVeryDark, snap.AsciiMode)
			}
			return render.RenderPressureLamps(render.PercentToLamps(snap.MemPercent), snap.AsciiMode)
		},
	})

	catalog.Register(&Widget{
		ID:          "res-dmem",
		ElementName: "Mem",
		HelpText:    "Docker memory pool: green <37%, yellow 37-68%, orange 68-87%, red >87% (containers may OOM)",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			return render.RenderPressureLamps(render.PercentToLamps(snap.DockerMemPercent), snap.AsciiMode)
		},
	})

	catalog.Register(&Widget{
		ID:          "res-host",
		ElementName: "Host",
		HelpText:    "Host scheduler pressure (running/target): green <37%, yellow 37-68%, orange 68-87%, red >87% queuing jobs",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			if snap.HostPressureTarget <= 0 {
				return render.RenderEmptyLamps(render.LampVeryDark, snap.AsciiMode)
			}
			usagePct := float64(snap.HostRunning) / float64(snap.HostPressureTarget) * 100.0
			return render.RenderPressureLamps(render.PercentToLamps(usagePct), snap.AsciiMode)
		},
	})

	catalog.Register(&Widget{
		ID:          "res-docker",
		ElementName: "Pressure",
		HelpText:    "Docker scheduler pressure: green <37%, yellow 37-68%, orange 68-87%, red >87% queuing containers",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			target := snap.DockerPressureTarget
			if target <= 0 {
				target = snap.DockerRoof
			}
			if target <= 0 {
				return render.RenderEmptyLamps(render.LampVeryDark, snap.AsciiMode)
			}
			usagePct := float64(snap.DockerRunning) / float64(target) * 100.0
			return render.RenderPressureLamps(render.PercentToLamps(usagePct), snap.AsciiMode)
		},
	})

	// === Tabs Panel Header (view_tabs.go) ===

	catalog.Register(&Widget{
		ID:          "res-progress",
		ElementName: "Progress",
		HelpText:    "Overall progress: green gradient bar showing finalized/total UoWs (dark green=start, bright green=done)",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			const progressLampCount = 12
			if snap.TotalWeight == 0 {
				return render.RenderEmptyLamps(render.LampVeryDark, snap.AsciiMode)
			}
			activeLamps := snap.FinalizedWeight * progressLampCount / snap.TotalWeight
			if snap.FinalizedWeight >= snap.TotalWeight {
				activeLamps = progressLampCount
			}
			return render.RenderProgressGradientLamps(activeLamps, progressLampCount, snap.AsciiMode)
		},
	})

	// === Controls (view_resources.go) ===

	catalog.Register(&Widget{
		ID:          "freeze-button",
		ElementName: "Freeze",
		HelpText:    "Click [...] to pause auto-exit (2min); shows countdown when active: green >60s, yellow 20-60s, red <20s",
		ZoneEnabled: true,
		Render: func(snap WidgetSnapshot) string {
			if snap.ExitCountdownSecs > 0 && snap.UserHasInteracted {
				mins := snap.ExitCountdownSecs / 60
				secs := snap.ExitCountdownSecs % 60
				var btnStyle lipgloss.Style
				if snap.ExitCountdownSecs > 60 {
					btnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34")).Bold(true)
				} else if snap.ExitCountdownSecs > 20 {
					btnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
				} else {
					btnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				}
				return btnStyle.Render(fmt.Sprintf("[%d:%02d]", mins, secs))
			}
			return Styles.Dim.Render("[...]")
		},
	})

	// === Template Widget: UoW Tab (view_tabs.go) ===

	tabColors := map[UoWStatus]TabBadgeColors{
		UoWPending:  {Bg: "238", Text: "250", BgActive: "245", NameBg: "236"},
		UoWRunning:  {Bg: "208", Text: "232", BgActive: "220", NameBg: "94"},
		UoWComplete: {Bg: "34", Text: "232", BgActive: "46", NameBg: "22"},
		UoWSkipped:  {Bg: "31", Text: "232", BgActive: "45", NameBg: "23"},
		UoWFailed:   {Bg: "160", Text: "255", BgActive: "196", NameBg: "52"},
	}

	catalog.RegisterTabWidget(&TabWidgetDef{
		ID:          "uow-tab",
		ElementName: "Component",
		HelpText:    "Unit of work: click to select, shows logs + status in right panel. Badge color = status, number = weight.",
		ColorMap:    tabColors,
		Render: func(instance TabInstance, sizing TabSizing) string {
			// Badge is fixed width (3 chars): " N " or "NN "
			var weightStr string
			if instance.Weight < 10 {
				weightStr = fmt.Sprintf(" %d ", instance.Weight)
			} else {
				weightStr = fmt.Sprintf("%2d ", instance.Weight)
			}

			// Select label text based on view mode
			label := tabLabelForViewMode(instance, sizing.ViewMode)

			// Marquee scrolling for hovered tab when label exceeds label width
			if instance.IsHovered && len(label) > sizing.LabelWidth {
				const startDelay = 10
				if sizing.MarqueePos > startDelay {
					effectiveScroll := (sizing.MarqueePos - startDelay) / 4
					scrollPos := effectiveScroll % (len(label) + 3)
					if scrollPos < len(label) {
						scrolled := label[scrollPos:]
						if len(scrolled) < sizing.LabelWidth {
							scrolled = scrolled + "   " + label
						}
						label = scrolled
					} else {
						gapPos := scrollPos - len(label)
						label = strings.Repeat(" ", 3-gapPos) + label
					}
				}
			}

			// Truncate to fit
			for lipgloss.Width(label) > sizing.LabelWidth && len(label) > 0 {
				label = label[:len(label)-1]
			}
			// Remove trailing colon
			if strings.HasSuffix(label, ":") {
				label = label[:len(label)-1]
			}
			// Pad to fixed width
			if lipgloss.Width(label) < sizing.LabelWidth {
				label = label + strings.Repeat(" ", sizing.LabelWidth-lipgloss.Width(label))
			}

			// Look up colors (all UoWStatus values are in tabColors)
			colors := tabColors[instance.Status]

			// Badge background: brighter when active
			badgeBg := colors.Bg
			if instance.IsActive {
				badgeBg = colors.BgActive
			}

			// Name style
			nameBgColor := colors.NameBg
			nameFgColor := lipgloss.Color("252")
			if instance.IsHovered {
				nameBgColor = colors.Bg
				nameFgColor = lipgloss.Color("232")
			}
			nameStyle := lipgloss.NewStyle().Foreground(nameFgColor).Background(nameBgColor)
			namePart := nameStyle.Render(label)

			// Badge style
			badgeStyle := lipgloss.NewStyle().Foreground(colors.Text).Background(badgeBg)
			if instance.Status == UoWRunning {
				badgeStyle = badgeStyle.Bold(true)
			}
			badgePart := badgeStyle.Render(weightStr)

			return namePart + badgePart
		},
	})
}

// tabLabelForViewMode selects the label text for a tab based on the view mode.
// Falls back to DisplayName when the requested field is empty.
func tabLabelForViewMode(inst TabInstance, mode TabViewMode) string {
	switch mode {
	case TabViewModule:
		if inst.Module != "" {
			return inst.Module
		}
	case TabViewType:
		if inst.ComponentType != "" {
			return inst.ComponentType
		}
	case TabViewTool:
		if inst.Tool != "" {
			return inst.Tool
		}
	case TabViewExec:
		if inst.Container {
			return "container"
		}
		return "host"
	case TabViewState:
		return formatTabState(inst)
	}
	return inst.DisplayName
}

// formatTabState returns compact state text for the State view mode.
func formatTabState(inst TabInstance) string {
	switch inst.Status {
	case UoWPending:
		return "pending"
	case UoWRunning:
		if !inst.StartTime.IsZero() {
			return "run " + formatElapsed(time.Since(inst.StartTime))
		}
		return "running"
	case UoWComplete:
		if !inst.StartTime.IsZero() && !inst.EndTime.IsZero() {
			return "done " + formatElapsed(inst.EndTime.Sub(inst.StartTime))
		}
		return "done"
	case UoWSkipped:
		return "cached"
	case UoWFailed:
		if inst.ExitCode > 0 {
			return fmt.Sprintf("fail:%d", inst.ExitCode)
		}
		return "failed"
	default:
		return "?"
	}
}

// formatElapsed returns a compact human-readable duration string.
func formatElapsed(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	secs := int(d.Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	secs = secs % 60
	if secs == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%dm%ds", mins, secs)
}
