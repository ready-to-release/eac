package get

import (
	"testing"
)

func TestGetSpecs(t *testing.T) {
	// Verify command is registered
	// Command registration happens in init(), so we just need to verify
	// the function exists and can be called
	if GetSpecs == nil {
		t.Error("GetSpecs command is nil")
	}
}
