package eac

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/stretchr/testify/assert"
)

func TestFactory_FallbackToNative(t *testing.T) {
	registry := tool.NewRegistry()
	executor := tool.NewExecutor()

	adapter := New("/tmp/workspace", registry, executor)

	// Should return a nativeAdapter since "eac" is not registered
	_, ok := adapter.(*nativeAdapter)
	assert.True(t, ok, "expected *nativeAdapter when eac not in registry")
}
