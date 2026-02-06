package render

import "github.com/charmbracelet/lipgloss"

// --- Color constants (256-color palette) ---

var (
	ColorWhite     = lipgloss.Color("255")
	ColorLightGray = lipgloss.Color("252")
	ColorGray      = lipgloss.Color("245")
	ColorDim       = lipgloss.Color("240")
	ColorDarkGray  = lipgloss.Color("238")
	ColorVeryDark  = lipgloss.Color("236")
	ColorDarkBg    = lipgloss.Color("234")
	ColorBlack     = lipgloss.Color("0")

	ColorGreen     = lipgloss.Color("71")
	ColorBrightGreen = lipgloss.Color("40")
	ColorDarkGreen = lipgloss.Color("22")
	ColorGreenBold = lipgloss.Color("34")
	ColorMintGreen = lipgloss.Color("42")

	ColorYellow     = lipgloss.Color("226")
	ColorDimYellow  = lipgloss.Color("58")
	ColorGold       = lipgloss.Color("214")

	ColorOrange    = lipgloss.Color("208")
	ColorDimOrange = lipgloss.Color("130")

	ColorRed       = lipgloss.Color("196")
	ColorDarkRed   = lipgloss.Color("52")

	ColorCyan       = lipgloss.Color("39")
	ColorLightCyan  = lipgloss.Color("75")
	ColorBrightCyan = lipgloss.Color("51")

	ColorBlue    = lipgloss.Color("33")
	ColorDeepBlue = lipgloss.Color("24")

	ColorPurple = lipgloss.Color("141")
)

// --- Lamp zone styles (filled/empty for pressure displays) ---

var (
	LampGreenFilled  = lipgloss.NewStyle().Foreground(ColorGreen)
	LampGreenEmpty   = lipgloss.NewStyle().Foreground(ColorDarkGreen)
	LampYellowFilled = lipgloss.NewStyle().Foreground(ColorYellow)
	LampYellowEmpty  = lipgloss.NewStyle().Foreground(ColorDimYellow)
	LampOrangeFilled = lipgloss.NewStyle().Foreground(ColorOrange)
	LampOrangeEmpty  = lipgloss.NewStyle().Foreground(ColorDimOrange)
	LampRedFilled    = lipgloss.NewStyle().Foreground(ColorRed)
	LampRedEmpty     = lipgloss.NewStyle().Foreground(ColorDarkRed)
	LampDim          = lipgloss.NewStyle().Foreground(ColorDim)
	LampVeryDark     = lipgloss.NewStyle().Foreground(ColorVeryDark)
)

// --- CPU-specific lamp styles ---

var (
	CPUGreen  = lipgloss.NewStyle().Foreground(ColorGreen)
	CPUYellow = lipgloss.NewStyle().Foreground(ColorYellow)
	CPUOrange = lipgloss.NewStyle().Foreground(ColorOrange)
	CPURed    = lipgloss.NewStyle().Foreground(ColorRed)
)

// --- Container/tool indicator styles ---

var (
	ContainerRunning = lipgloss.NewStyle().Foreground(ColorBrightCyan)
	ContainerEmpty   = lipgloss.NewStyle().Foreground(ColorVeryDark)
	ContainerActive  = lipgloss.NewStyle().Foreground(ColorCyan)
	ContainerIdle    = lipgloss.NewStyle().Foreground(ColorLightCyan)

	ToolActive = lipgloss.NewStyle().Foreground(ColorOrange)
	ToolIdle   = lipgloss.NewStyle().Foreground(ColorDimOrange)
	ToolEmpty  = lipgloss.NewStyle().Foreground(ColorVeryDark)
)

// --- Weight-based active job styles ---

var (
	WeightLow    = lipgloss.NewStyle().Foreground(ColorGreen)    // weight 1-2
	WeightMedium = lipgloss.NewStyle().Foreground(ColorYellow)   // weight 3-4
	WeightHigh   = lipgloss.NewStyle().Foreground(ColorOrange)   // weight 5-6
	WeightMax    = lipgloss.NewStyle().Foreground(ColorRed)      // weight 7+
)

// --- Status indicator styles (for selected UoW pane) ---

var (
	StatusLabelStyle   = lipgloss.NewStyle().Foreground(ColorWhite)
	StatusActiveStyle  = lipgloss.NewStyle().Foreground(ColorGold)
	StatusPendingStyle = lipgloss.NewStyle().Foreground(ColorGray)
	StatusRunningStyle = lipgloss.NewStyle().Foreground(ColorBlue)
	StatusCompleteStyle = lipgloss.NewStyle().Foreground(ColorMintGreen)
	StatusCachedStyle  = lipgloss.NewStyle().Foreground(ColorPurple)
	StatusFailedStyle  = lipgloss.NewStyle().Foreground(ColorRed)
)

// --- Resource pane header styles ---

var (
	ResourceHeaderGreen  = lipgloss.NewStyle().Foreground(ColorGreenBold).Bold(true)
	ResourceHeaderYellow = lipgloss.NewStyle().Foreground(ColorGold).Bold(true)
	ResourceHeaderRed    = lipgloss.NewStyle().Foreground(ColorRed).Bold(true)
	ResourceValueWhite   = lipgloss.NewStyle().Foreground(ColorWhite)
	ResourceValueYellow  = lipgloss.NewStyle().Foreground(ColorGold)
	ResourceValueRed     = lipgloss.NewStyle().Foreground(ColorRed)
	ResourceValueDim     = lipgloss.NewStyle().Foreground(ColorVeryDark)
)

// --- Badge styles (for tab grid summary indicators) ---

var (
	BadgeCyan  = lipgloss.NewStyle().Background(ColorCyan).Foreground(ColorBlack)
	BadgeGreen = lipgloss.NewStyle().Background(ColorBrightGreen).Foreground(ColorBlack)
	BadgeRed   = lipgloss.NewStyle().Background(ColorRed).Foreground(ColorBlack)
)

// --- Help pane styles ---

var (
	HelpTitleStyle = lipgloss.NewStyle().Foreground(ColorWhite).Bold(true)
	HelpBodyStyle  = lipgloss.NewStyle().Foreground(ColorLightGray)
)

// --- Progress gradient lamp styles (dark green to bright green) ---

// ProgressGradientColors defines the 16-shade green gradient from dark to bright.
// Each index maps to a lamp position (0=leftmost/darkest, 15=rightmost/brightest).
var ProgressGradientColors = [16]lipgloss.Color{
	lipgloss.Color("22"),  // 0: darkest green
	lipgloss.Color("22"),  // 1: dark forest
	lipgloss.Color("28"),  // 2: deep green
	lipgloss.Color("28"),  // 3: forest green
	lipgloss.Color("34"),  // 4: medium-dark green
	lipgloss.Color("34"),  // 5: medium green
	lipgloss.Color("70"),  // 6: green-yellow hint
	lipgloss.Color("71"),  // 7: sage green
	lipgloss.Color("76"),  // 8: lime-green
	lipgloss.Color("77"),  // 9: light lime
	lipgloss.Color("40"),  // 10: bright green
	lipgloss.Color("41"),  // 11: vivid green
	lipgloss.Color("42"),  // 12: mint green
	lipgloss.Color("48"),  // 13: bright mint
	lipgloss.Color("46"),  // 14: pure green
	lipgloss.Color("83"),  // 15: brightest green
}

// ProgressGradientFilled are per-position filled lamp styles for the progress bar.
var ProgressGradientFilled [16]lipgloss.Style

// ProgressGradientEmpty is the uniform style for unfilled progress lamps.
var ProgressGradientEmpty = lipgloss.NewStyle().Foreground(ColorVeryDark)

func init() {
	for i, c := range ProgressGradientColors {
		ProgressGradientFilled[i] = lipgloss.NewStyle().Foreground(c)
	}
}

// --- Misc styles ---

var (
	ReverseStyle     = lipgloss.NewStyle().Reverse(true)
	InitCompleteStyle = lipgloss.NewStyle().Foreground(ColorGreen)
	LogsTitleStyle    = lipgloss.NewStyle().Foreground(ColorLightCyan)
	CachedTitleStyle  = lipgloss.NewStyle().Foreground(ColorLightCyan)
	TreeGreenStyle    = lipgloss.NewStyle().Foreground(ColorGreen)
)
