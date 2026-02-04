package cells

import "fmt"

// DepsCell displays the dependencies for the selected unit.
type DepsCell struct {
	deps []DependencyInfo
}

// NewDepsCell creates a new deps cell.
func NewDepsCell() *DepsCell {
	return &DepsCell{}
}

// SetDeps sets the dependency information.
func (c *DepsCell) SetDeps(deps []DependencyInfo) {
	c.deps = deps
}

// Render returns the dependencies display.
func (c *DepsCell) Render(width, height int) string {
	if len(c.deps) == 0 {
		return "deps: none"
	}

	prefix := "deps: "
	var parts []string

	for _, dep := range c.deps {
		part := fmt.Sprintf("%s (%s)", dep.Moniker, dep.Status)
		parts = append(parts, part)
	}

	// Build result with truncation
	result := prefix
	for i, part := range parts {
		candidate := result
		if i > 0 {
			candidate += ", "
		}
		candidate += part

		// Check if adding this part would exceed width (leave room for " ...")
		if len(candidate) > width-4 && i > 0 {
			return result + " ..."
		}
		result = candidate
	}

	return result
}

// ZoneID returns empty string (not an interactive zone).
func (c *DepsCell) ZoneID() string {
	return ""
}
