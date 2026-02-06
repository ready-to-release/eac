package console

import (
	"time"
)

// WidgetSnapshot contains all data a widget might need, extracted from Model.
// Fields are flat and cheap to construct (no pointers to mutable state).
// Not every widget uses every field -- each widget reads only what it needs.
type WidgetSnapshot struct {
	// System metrics (from cached values)
	CPUPercent       []float64
	MemPercent       float64
	DockerMemPercent float64

	// Capacity / scheduler
	HostRunning          int
	HostPressureTarget   int
	HostRoof             int
	HostWaiting          int
	DockerRunning        int
	DockerPressureTarget int
	DockerRoof           int
	DockerWaiting        int

	// Progress
	Counts   DerivedCounts
	UoWTotal int // From initSummary.UoWCount or len(uowOrder)

	// Timing
	Elapsed time.Duration

	// Completion state
	SummaryData    *SummaryData
	RunPhaseActive bool
	RunPhaseName   string

	// Freeze/exit
	ExitCountdownSecs int
	UserHasInteracted bool

	// Display
	AsciiMode bool
	Width     int // Terminal width

	// Metrics state
	LastMetricsUpdate time.Time // For "no data yet" detection in mem widgets

	// Progress (weighted) -- pre-computed from uowStates
	FinalizedWeight int
	TotalWeight     int
}

// Widget is a registered displayable element in the TUI.
// Each widget knows how to render itself given a data snapshot.
// Widgets are stateless -- all state comes from the snapshot.
type Widget struct {
	// Identity
	ID          string // Unique identifier (e.g., "res-cpu", "freeze-button")
	ElementName string // Human-readable name for Selected pane (e.g., "CPU", "Freeze")
	HelpText    string // Help description shown on hover

	// Rendering
	Render func(snap WidgetSnapshot) string // Returns styled string (without zone.Mark wrapping)

	// Zone integration
	ZoneEnabled bool // Whether to wrap output in zone.Mark(ID, ...)
}

// TabInstance holds per-UoW data for the template tab widget.
// One instance per UoW tab in the grid. Cheap value type.
type TabInstance struct {
	Moniker     string    // Full moniker (used as zone ID): "build:eac-cli:go:go"
	DisplayName string    // Short display name for tab label
	Status      UoWStatus // Current execution status
	Weight      int       // Execution weight (shown in badge)
	IsActive    bool      // Currently selected tab
	IsHovered   bool      // Mouse cursor over this tab

	// Structured identity (from UoWState)
	Module        string
	Component     string
	Tool          string
	ComponentType string
	Container     bool

	// Timing (for State view mode)
	StartTime time.Time
	EndTime   time.Time
	ExitCode  int
}

// TabSizing holds dynamically computed dimensions for tab rendering.
// Recomputed when terminal width or tab width changes.
type TabSizing struct {
	TabWidth   int         // Total width per tab (name + badge). Dynamic, NOT hardcoded 15.
	LabelWidth int         // Name area = TabWidth - BadgeWidth
	BadgeWidth int         // Weight badge (always 3)
	TabColumns int         // Number of columns in the grid (2-6)
	MarqueePos int         // Current marquee scroll position for hovered tab animation
	AsciiMode  bool        // ASCII fallback mode
	ViewMode   TabViewMode // Which label text to show
}

// ComputeTabSizing calculates tab dimensions from panel width and tab width.
// panelWidth is the total panel width including borders.
// tabWidth is the desired width per tab (label + badge).
func ComputeTabSizing(panelWidth, tabWidth int, marqueePos int, ascii bool) TabSizing {
	const badgeWidth = 3
	if tabWidth < badgeWidth+4 { // Minimum: 4 chars for name + 3 for badge
		tabWidth = badgeWidth + 4
	}
	available := panelWidth - tabBorder
	cols := (available + tabGap) / (tabWidth + tabGap)
	if cols < 1 {
		cols = 1
	}
	return TabSizing{
		TabWidth:   tabWidth,
		LabelWidth: tabWidth - badgeWidth,
		BadgeWidth: badgeWidth,
		TabColumns: cols,
		MarqueePos: marqueePos,
		AsciiMode:  ascii,
	}
}

// buildWidgetSnapshot creates a WidgetSnapshot from the current Model state.
// Called once per View() frame -- O(1) for most fields, O(n) for DeriveCounts + weight sums.
func (m Model) buildWidgetSnapshot() WidgetSnapshot {
	counts := m.DeriveCounts()

	// Get capacity info
	var pressureCap int
	for _, lock := range m.locks {
		if lock.Name == "component-scheduler" {
			pressureCap = lock.Capacity
			break
		}
	}
	capInfo := m.GetCapacityInfo()
	if capInfo.Roof == 0 {
		capInfo.Roof = pressureCap
	}
	if capInfo.PressureTarget == 0 {
		capInfo.PressureTarget = pressureCap
	}

	// Determine host waiting from locks
	var hostWaiting int
	for _, lock := range m.locks {
		if lock.Name == "component-scheduler" {
			hostWaiting = lock.Waiting
			break
		}
	}

	// UoW total from initSummary or fallback
	uowTotal := len(m.uowOrder)
	if m.initSummary != nil && m.initSummary.UoWCount > 0 {
		uowTotal = m.initSummary.UoWCount
	}

	// Pre-compute weighted progress
	var totalWeight, finalizedWeight int
	for _, state := range m.uowStates {
		w := state.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
		switch state.Status {
		case UoWComplete, UoWSkipped, UoWFailed:
			finalizedWeight += w
		}
	}

	return WidgetSnapshot{
		CPUPercent:           m.cachedCPUPercent,
		MemPercent:           m.cachedMemPercent,
		DockerMemPercent:     m.cachedDockerMemPercent,
		HostRunning:          counts.Running,
		HostPressureTarget:   capInfo.PressureTarget,
		HostRoof:             capInfo.Roof,
		HostWaiting:          hostWaiting,
		DockerRunning:        m.dockerRunning,
		DockerPressureTarget: m.dockerPressureTarget,
		DockerRoof:           m.dockerRoof,
		DockerWaiting:        m.dockerWaiting,
		Counts:               counts,
		UoWTotal:             uowTotal,
		Elapsed:              time.Since(m.startTime),
		SummaryData:          m.summaryData,
		RunPhaseActive:       m.panes[PhaseRun].Status == PhaseActive,
		RunPhaseName:         m.runPhaseName,
		ExitCountdownSecs:    m.exitCountdownSecs,
		UserHasInteracted:    m.userHasInteracted,
		AsciiMode:            m.asciiMode,
		Width:                m.width,
		LastMetricsUpdate:    m.lastMetricsUpdate,
		FinalizedWeight:      finalizedWeight,
		TotalWeight:          totalWeight,
	}
}