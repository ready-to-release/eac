// Package cells provides modular TUI cell components for the tui3 layout.
// Each cell represents a discrete visual element in the TUI grid.
package cells

// Cell is the interface for renderable TUI cell components.
type Cell interface {
	// Render returns the cell's view at the given dimensions.
	Render(width, height int) string
	// ZoneID returns the mouse zone identifier for this cell.
	// Returns "" if the cell has no interactive zone.
	ZoneID() string
}

// DataSetter is implemented by cells that accept external data updates.
type DataSetter interface {
	SetData(data interface{})
}
