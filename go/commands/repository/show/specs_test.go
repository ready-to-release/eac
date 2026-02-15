package show

import (
	"testing"
)

func TestShowSpecs_Exists(t *testing.T) {
	// Verify command function exists (registration happens in init())
	// This is a compile-time check - if ShowSpecs doesn't exist, this won't compile
	_ = ShowSpecs
}
