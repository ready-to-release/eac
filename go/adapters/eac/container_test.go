package eac

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/stretchr/testify/assert"
)

func TestContainerAdapter_PlaceholderResolution(t *testing.T) {
	td := &tool.ToolDefinition{
		ID:        "eac",
		Type:      tool.ToolTypeContainer,
		LocalPath: "containers/ext-eac",
	}
	adapter := NewContainer("/workspace", nil, td)
	ca := adapter.(*containerAdapter)

	assert.Equal(t, "/workspace", ca.workspaceRoot)
	assert.Equal(t, td, ca.toolDef)
}

func TestContainerAdapter_NilConfig(t *testing.T) {
	// Verify the adapter handles nil config without panicking in the config-handling path.
	// We can't actually Execute because we don't have a real executor,
	// but we verify the adapter was constructed correctly.
	td := &tool.ToolDefinition{
		ID:        "eac",
		Type:      tool.ToolTypeContainer,
		LocalPath: "containers/ext-eac",
	}
	adapter := NewContainer("/workspace", nil, td)
	assert.NotNil(t, adapter)
}
