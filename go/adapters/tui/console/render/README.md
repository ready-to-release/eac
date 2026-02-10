# render

Stateless rendering functions for the console TUI, providing icons, lamp displays, text formatting, layout metrics, and style definitions.

## Key Types

- **`LayoutMetrics`** -- Pre-calculated layout dimensions used by both rendering and mouse handling

## Key Functions

- `RenderPressureLamps` -- Renders a 16-lamp pressure display with green/yellow/orange/red zones
- `RenderCPULamps` -- Renders per-core CPU lamps normalized to 16 positions with per-lamp coloring
- `RenderProgressGradientLamps` -- Renders a progress bar with 16-shade green gradient
- `RenderCounterLamps` -- Renders a simple counter-based lamp display
- `RenderSegmentedLamps` -- Renders a 3-segment display (idle, active, unused)
- `RenderWeightDots` -- Renders weight-colored dots for active jobs
- `UoWStatusIcon` -- Returns status icons with ASCII fallback
- `PhaseIcon` -- Returns phase status icons with ASCII fallback
- `FormatElapsed` -- Formats duration for display
- `PadOrTruncate` -- Pads or truncates a string to exact width with Unicode ellipsis
- `StripMarkdownPipes` -- Converts markdown table lines to plain terminal text

## Patterns

- Stateless rendering: All functions take explicit data parameters, not Model references
- ASCII fallback: Icon and lamp functions accept `asciiMode` for terminal compatibility
- Pressure zones: Non-uniform color distribution (3/8 green, remainder yellow, 3/16 orange, 1/8 red)
- Style constants: Package-level lipgloss style variables for consistent theming

## Internal Structure

| File | Responsibility |
| --- | --- |
| icons.go | UoW status icons, phase icons, and line level icons with ASCII fallback |
| lamps.go | Pressure lamps, CPU lamps, counter lamps, segmented lamps, progress gradient, and weight dots |
| styles.go | Color constants, lamp zone styles, CPU styles, container/tool styles, badge styles, and gradient palette |
| text.go | Text formatting: elapsed time, pad/truncate, markdown stripping, tool name cleanup, weight digits |
| layout.go | `LayoutMetrics` struct for shared layout dimensions |
| doc.go | Package documentation |

## Dependencies

- None (leaf package; imports only lipgloss for styling)

## Role in System

The render sub-package extracts all stateless rendering logic from the console model, enabling isolated unit testing of visual output without bubbletea scaffolding. It is the single source of truth for color theming, icon sets, and lamp display calculations used throughout the parallel TUI.

## Code Health

### Tech Debt
- None

### Pain Points
- None identified

### Optimization Opportunities
- None identified
