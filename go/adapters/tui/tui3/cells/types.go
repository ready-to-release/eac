package cells

// DependencyInfo holds information about a unit's dependency.
type DependencyInfo struct {
	Moniker string // Dependency moniker (e.g., "contracts")
	Status  string // Status (e.g., "ready", "building", "pending")
}

// ArtifactInfo holds information about a build artifact.
type ArtifactInfo struct {
	Name string // Artifact name (e.g., "eac.exe")
	Path string // Full path (optional)
	Size int64  // Size in bytes
}

// UnitStatus represents execution state of a unit.
type UnitStatus int

const (
	UnitPending UnitStatus = iota
	UnitQueued
	UnitRunning
	UnitComplete
	UnitSkipped
	UnitFailed
)

// Icon returns an ASCII icon for the status.
func (s UnitStatus) Icon() string {
	switch s {
	case UnitPending:
		return "o"
	case UnitQueued:
		return "*"
	case UnitRunning:
		return ">"
	case UnitComplete:
		return "V"
	case UnitSkipped:
		return "="
	case UnitFailed:
		return "X"
	default:
		return "?"
	}
}

// Colors returns ANSI 256-color codes for the status.
// Returns (border, text, bg) color codes.
func (s UnitStatus) Colors() (border, text, bg string) {
	switch s {
	case UnitPending, UnitQueued:
		return "238", "245", "234" // Gray
	case UnitRunning:
		return "214", "214", "94" // Orange/Yellow
	case UnitComplete:
		return "40", "40", "22" // Green
	case UnitSkipped:
		return "75", "75", "23" // Cyan/Blue
	case UnitFailed:
		return "196", "196", "52" // Red
	default:
		return "238", "245", "234" // Gray (default)
	}
}

// BadgeColors returns colors for the icon badge area.
// Returns (fg, bg) for normal state and (fgActive, bgActive) for selected/active state.
func (s UnitStatus) BadgeColors() (fg, bg, fgActive, bgActive string) {
	switch s {
	case UnitPending, UnitQueued:
		return "250", "238", "255", "245" // Gray -> lighter gray
	case UnitRunning:
		return "232", "208", "232", "220" // Dark on orange -> dark on bright yellow
	case UnitComplete:
		return "232", "34", "232", "46" // Dark on green -> dark on bright green
	case UnitSkipped:
		return "232", "31", "232", "45" // Dark on cyan -> dark on bright cyan
	case UnitFailed:
		return "255", "160", "255", "196" // White on red -> white on bright red
	default:
		return "250", "238", "255", "245" // Gray
	}
}

// NameBgColor returns a subtle background color for the tab name area.
// This is a darker/lighter version of the badge color for visual continuity.
func (s UnitStatus) NameBgColor() string {
	switch s {
	case UnitPending, UnitQueued:
		return "236" // Dark gray
	case UnitRunning:
		return "94" // Dark orange/brown
	case UnitComplete:
		return "22" // Dark green
	case UnitSkipped:
		return "23" // Dark cyan
	case UnitFailed:
		return "52" // Dark red
	default:
		return "236" // Dark gray
	}
}

// SelectorUnit represents a unit in the selector grid.
type SelectorUnit struct {
	Moniker     string     // Globally unique ID (Longname) for matching
	DisplayName string     // Short name for tab display
	Status      UnitStatus
	Weight      int        // Number of components/work items
}

// LineLevel represents the severity level of an output line.
type LineLevel int

const (
	LineLevelInfo LineLevel = iota
	LineLevelWarn
	LineLevelError
)

// OutputLine represents a single output line with metadata.
type OutputLine struct {
	Text   string
	Source string // Module moniker or "system"
	Level  LineLevel
}
