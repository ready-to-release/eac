package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StableLampsCount is the fixed number of lamps for all resource displays.
const StableLampsCount = 16

// --- Pressure zone calculation ---

// PressureLampZones calculates non-uniform color zone boundaries.
// Distribution: 3/8 green, remainder yellow, 3/16 orange, 1/8 red.
// For 16 lamps: 6 green, 5 yellow, 3 orange, 2 red.
func PressureLampZones(totalLamps int) (greenEnd, yellowEnd, orangeEnd int) {
	greenCount := totalLamps * 3 / 8
	redCount := totalLamps / 8
	orangeCount := totalLamps * 3 / 16
	yellowCount := totalLamps - greenCount - redCount - orangeCount

	greenEnd = greenCount
	yellowEnd = greenEnd + yellowCount
	orangeEnd = yellowEnd + orangeCount
	return
}

// GetPressureColor returns a lipgloss style for the given pressure level.
func GetPressureColor(running, capacity int) lipgloss.Style {
	if capacity <= 0 {
		return lipgloss.NewStyle().Foreground(ColorWhite)
	}
	activeDots := (running * StableLampsCount) / capacity
	if activeDots > StableLampsCount {
		activeDots = StableLampsCount
	}
	greenEnd, yellowEnd, orangeEnd := PressureLampZones(StableLampsCount)
	pos := activeDots
	if pos > 0 {
		pos--
	}
	switch {
	case pos < greenEnd:
		return lipgloss.NewStyle().Foreground(ColorBrightGreen)
	case pos < yellowEnd:
		return lipgloss.NewStyle().Foreground(ColorGold)
	case pos < orangeEnd:
		return lipgloss.NewStyle().Foreground(ColorOrange)
	default:
		return lipgloss.NewStyle().Foreground(ColorRed)
	}
}

// --- Generic lamp rendering ---

// LampChars returns the filled and empty characters based on ascii mode.
func LampChars(asciiMode bool) (filled, empty string) {
	if asciiMode {
		return "*", "o"
	}
	return "●", "○"
}

// PressureZoneStyles returns the standard 4-zone filled/empty style pairs.
func PressureZoneStyles() [4][2]lipgloss.Style {
	return [4][2]lipgloss.Style{
		{LampGreenFilled, LampGreenEmpty},   // green zone
		{LampYellowFilled, LampYellowEmpty}, // yellow zone
		{LampOrangeFilled, LampOrangeEmpty}, // orange zone
		{LampRedFilled, LampRedEmpty},       // red zone
	}
}

// RenderPressureLamps renders a 16-lamp pressure display with standard green/yellow/orange/red zones.
// activeDots is the number of filled lamps (0 to totalLamps).
func RenderPressureLamps(activeDots int, asciiMode bool) string {
	const totalDots = StableLampsCount
	if activeDots > totalDots {
		activeDots = totalDots
	}
	if activeDots < 0 {
		activeDots = 0
	}

	filledChar, emptyChar := LampChars(asciiMode)
	greenEnd, yellowEnd, orangeEnd := PressureLampZones(totalDots)
	styles := PressureZoneStyles()

	var dots strings.Builder
	for i := 0; i < totalDots; i++ {
		isFilled := i < activeDots
		char := emptyChar
		if isFilled {
			char = filledChar
		}

		var zoneIdx int
		switch {
		case i < greenEnd:
			zoneIdx = 0
		case i < yellowEnd:
			zoneIdx = 1
		case i < orangeEnd:
			zoneIdx = 2
		default:
			zoneIdx = 3
		}

		if isFilled {
			dots.WriteString(styles[zoneIdx][0].Render(char))
		} else {
			dots.WriteString(styles[zoneIdx][1].Render(char))
		}
	}
	return dots.String()
}

// PercentToLamps converts a percentage (0-100) to active lamp count out of 16.
func PercentToLamps(pct float64) int {
	active := int(pct / 100.0 * float64(StableLampsCount))
	if active > StableLampsCount {
		active = StableLampsCount
	}
	if active < 0 {
		active = 0
	}
	return active
}

// RenderCounterLamps renders a simple counter-based lamp display.
// filled lamps use filledStyle, empty lamps use emptyStyle.
func RenderCounterLamps(count int, filledStyle, emptyStyle lipgloss.Style, asciiMode bool) string {
	const totalDots = StableLampsCount
	if count > totalDots {
		count = totalDots
	}
	filledChar, emptyChar := LampChars(asciiMode)

	var dots strings.Builder
	for i := 0; i < totalDots; i++ {
		if i < count {
			dots.WriteString(filledStyle.Render(filledChar))
		} else {
			dots.WriteString(emptyStyle.Render(emptyChar))
		}
	}
	return dots.String()
}

// RenderSegmentedLamps renders a 3-segment display: idle (left), active (middle), unused (right).
func RenderSegmentedLamps(idleCount, activeCount int, idleStyle, activeStyle, emptyStyle lipgloss.Style, asciiMode bool) string {
	const fixedLamps = StableLampsCount
	filledChar, emptyChar := LampChars(asciiMode)

	// Clamp to max
	total := idleCount + activeCount
	if total > fixedLamps {
		if idleCount > fixedLamps {
			idleCount = fixedLamps
			activeCount = 0
		} else {
			activeCount = fixedLamps - idleCount
		}
	}

	var dots strings.Builder
	lampCount := 0

	for i := 0; i < idleCount && lampCount < fixedLamps; i++ {
		dots.WriteString(idleStyle.Render(emptyChar))
		lampCount++
	}
	for i := 0; i < activeCount && lampCount < fixedLamps; i++ {
		dots.WriteString(activeStyle.Render(filledChar))
		lampCount++
	}
	for lampCount < fixedLamps {
		dots.WriteString(emptyStyle.Render(emptyChar))
		lampCount++
	}
	return dots.String()
}

// RenderEmptyLamps renders all-empty placeholder lamps with dim styling.
func RenderEmptyLamps(style lipgloss.Style, asciiMode bool) string {
	_, emptyChar := LampChars(asciiMode)
	return style.Render(strings.Repeat(emptyChar, StableLampsCount))
}

// RenderCPULamps renders per-core CPU lamps normalized to 16 positions.
// Each lamp is independently colored based on its core's usage percentage.
func RenderCPULamps(perCore []float64, asciiMode bool) string {
	if len(perCore) == 0 {
		return RenderEmptyLamps(LampDim, asciiMode)
	}

	// Normalize cores to StableLampsCount lamps
	normalized := make([]float64, StableLampsCount)
	coreCount := len(perCore)

	if coreCount <= StableLampsCount {
		lampsPerCore := float64(StableLampsCount) / float64(coreCount)
		for i, pct := range perCore {
			startLamp := int(float64(i) * lampsPerCore)
			endLamp := int(float64(i+1) * lampsPerCore)
			if endLamp > StableLampsCount {
				endLamp = StableLampsCount
			}
			for j := startLamp; j < endLamp; j++ {
				normalized[j] = pct
			}
		}
	} else {
		coresPerLamp := float64(coreCount) / float64(StableLampsCount)
		for i := 0; i < StableLampsCount; i++ {
			startCore := int(float64(i) * coresPerLamp)
			endCore := int(float64(i+1) * coresPerLamp)
			if endCore > coreCount {
				endCore = coreCount
			}
			sum := 0.0
			count := 0
			for j := startCore; j < endCore; j++ {
				sum += perCore[j]
				count++
			}
			if count > 0 {
				normalized[i] = sum / float64(count)
			}
		}
	}

	filledChar, emptyChar := LampChars(asciiMode)

	var dots strings.Builder
	for _, pct := range normalized {
		switch {
		case pct < 5:
			dots.WriteString(LampGreenEmpty.Render(emptyChar))
		case pct < 50:
			dots.WriteString(CPUGreen.Render(filledChar))
		case pct < 80:
			dots.WriteString(CPUYellow.Render(filledChar))
		case pct < 95:
			dots.WriteString(CPUOrange.Render(filledChar))
		default:
			dots.WriteString(CPURed.Render(filledChar))
		}
	}
	return dots.String()
}

// RenderProgressGradientLamps renders a progress bar with per-lamp green gradient.
// Each filled lamp has a unique color sampled from the 16-shade ProgressGradientColors palette.
// lampCount specifies the number of lamps (typically 12 for progress column, 16 for standard).
// activeLamps is the number of filled lamps (0 to lampCount).
// Unfilled lamps use a uniform very-dark style.
func RenderProgressGradientLamps(activeLamps, lampCount int, asciiMode bool) string {
	if lampCount <= 0 {
		lampCount = StableLampsCount
	}
	if activeLamps > lampCount {
		activeLamps = lampCount
	}
	if activeLamps < 0 {
		activeLamps = 0
	}

	filledChar, emptyChar := LampChars(asciiMode)

	var dots strings.Builder
	for i := 0; i < lampCount; i++ {
		if i < activeLamps {
			// Sample gradient color: map lamp index to 0..15 palette evenly
			colorIdx := i * (len(ProgressGradientColors) - 1) / (lampCount - 1)
			dots.WriteString(ProgressGradientFilled[colorIdx].Render(filledChar))
		} else {
			dots.WriteString(ProgressGradientEmpty.Render(emptyChar))
		}
	}
	return dots.String()
}

// RenderWeightDots renders one filled dot per job, colored by weight.
func RenderWeightDots(weights []int, asciiMode bool) string {
	if len(weights) == 0 {
		return ""
	}

	filledChar, _ := LampChars(asciiMode)

	var dots strings.Builder
	for _, weight := range weights {
		switch {
		case weight <= 2:
			dots.WriteString(WeightLow.Render(filledChar))
		case weight <= 4:
			dots.WriteString(WeightMedium.Render(filledChar))
		case weight <= 6:
			dots.WriteString(WeightHigh.Render(filledChar))
		default:
			dots.WriteString(WeightMax.Render(filledChar))
		}
	}
	return dots.String()
}
