package get

import (
	"testing"
)

func TestGetSpecs_Exists(t *testing.T) {
	// Verify command function exists (registration happens in init())
	// This is a compile-time check - if GetSpecs doesn't exist, this won't compile
	var _ func() int = GetSpecs
}
