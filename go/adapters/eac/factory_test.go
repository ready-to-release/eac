package eac

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/stretchr/testify/assert"
)

func TestFactory_FallbackToNative(t *testing.T) {
	// Clear any existing global registry so "eac" is not found
	tool.SetGlobalRegistry(tool.NewRegistry())
	defer tool.SetGlobalRegistry(tool.NewRegistry())

	adapter := New("/tmp/workspace")

	// Should return a nativeAdapter since "eac" is not registered
	_, ok := adapter.(*nativeAdapter)
	assert.True(t, ok, "expected *nativeAdapter when eac not in registry")
}
