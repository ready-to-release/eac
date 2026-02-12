// Package tool provides handler adapters that bridge the new tool system
// with existing handler core.
package tool

// BaseHandlerAdapter provides common handler adapter functionality.
// Embed this in specific handler adapters to avoid repeating Name/Requirements/IsContainer/IsHostInstalled.
type BaseHandlerAdapter struct {
	Tool     *ToolDefinition
	Executor Executor
}

// Name returns the tool ID.
func (b *BaseHandlerAdapter) Name() string {
	return b.Tool.ID
}

// Requirements returns the tool's requirements.
func (b *BaseHandlerAdapter) Requirements() []string {
	return b.Tool.Requirements
}

// IsContainer returns true if this handler runs in a Docker container.
func (b *BaseHandlerAdapter) IsContainer() bool {
	return b.Tool.Type == ToolTypeContainer
}

// IsHostInstalled returns true if this handler runs using host-installed tools.
func (b *BaseHandlerAdapter) IsHostInstalled() bool {
	return b.Tool.Type == ToolTypeSystem
}
