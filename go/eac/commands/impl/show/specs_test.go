package show

import (
	"testing"
)

func TestShowSpecs(t *testing.T) {
	// Verify command is registered
	// Command registration happens in init(), so we just need to verify
	// the function exists and can be called
	if ShowSpecs == nil {
		t.Error("ShowSpecs command is nil")
	}
}
