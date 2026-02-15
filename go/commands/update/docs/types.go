package docs

import (
	"context"

	"github.com/ready-to-release/eac/go/core/tool"
)

// ToolBridge abstracts external tool execution for testability.
// In production, this is backed by tool.GlobalHandlerToolBridge().
type ToolBridge interface {
	ExecuteTool(ctx context.Context, toolName string, tc *tool.ToolContext) (int, error)
}

// defaultToolBridge wraps the global tool bridge for production use.
type defaultToolBridge struct{}

func (d *defaultToolBridge) ExecuteTool(ctx context.Context, toolName string, tc *tool.ToolContext) (int, error) {
	bridge := tool.GlobalHandlerToolBridge()
	return bridge.ExecuteTool(ctx, toolName, tc)
}

// UpdateOptions holds parsed flags for the update docs command.
type UpdateOptions struct {
	// Areas specifies which documentation areas to update.
	// An empty set means all areas.
	Areas *AreaSet

	// DryRun shows what would be changed without actually changing.
	DryRun bool

	// Force re-renders/re-optimizes all assets even if cached.
	Force bool

	// Verbose shows detailed progress for each file.
	Verbose bool

	// PruneCache identifies and removes orphaned cache files (mermaid/drawio).
	PruneCache bool

	// Bridge provides external tool execution. Defaults to the global tool bridge.
	// Tests can inject a mock to avoid Docker/CLI dependencies.
	Bridge ToolBridge
}

// GetBridge returns the configured ToolBridge, defaulting to the global tool bridge.
func (o UpdateOptions) GetBridge() ToolBridge {
	if o.Bridge != nil {
		return o.Bridge
	}
	return &defaultToolBridge{}
}
