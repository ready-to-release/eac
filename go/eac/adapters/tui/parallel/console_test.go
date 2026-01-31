package parallel

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/adapters/tui"
)

func TestConsole_ImplementsInterface(t *testing.T) {
	// Compile-time check is in console.go, but this test verifies at runtime
	var _ tui.Console = (*Console)(nil)
}

func TestConsole_Factory(t *testing.T) {
	factory := Factory()
	config := tui.Config{
		Height:       40,
		BufferSize:   1000,
		RunPhaseName: "building",
	}

	console := factory(config)
	if console == nil {
		t.Fatal("Factory returned nil console")
	}

	// Verify it's a parallel.Console
	if _, ok := console.(*Console); !ok {
		t.Error("Factory did not return a *Console")
	}
}

func TestConsole_Inner(t *testing.T) {
	c := New(tui.Config{})
	if c.Inner() == nil {
		t.Error("Inner() returned nil")
	}
}
