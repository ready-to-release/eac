package eac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNativeAdapter_ResolvedBinaryPath(t *testing.T) {
	adapter := NewNative("/usr/local/bin/eac", "/workspace", nil)

	// Type assert to access non-interface method
	native, ok := adapter.(*nativeAdapter)
	if !ok {
		t.Fatal("expected *nativeAdapter")
	}
	assert.Equal(t, "/usr/local/bin/eac", native.ResolvedBinaryPath())
}

func TestNativeAdapter_ToolDefFields(t *testing.T) {
	adapter := NewNative("/path/to/eac", "/ws", nil)
	native := adapter.(*nativeAdapter)

	assert.Equal(t, "eac", native.toolDef.ID)
	assert.Equal(t, "system", string(native.toolDef.Type))
	assert.Equal(t, "/path/to/eac", native.toolDef.Binary)
}
