package parallel

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/tui"
)

func TestInit_RegistersCommands(t *testing.T) {
	// The init() function should have registered build, test, lint, scan
	bindings := tui.ListBindings()

	found := make(map[string]bool)
	for _, b := range bindings {
		found[b] = true
	}

	for _, cmd := range []string{"build", "test", "lint", "scan"} {
		if !found[cmd] {
			t.Errorf("Expected %q to be registered, but it was not", cmd)
		}
	}
}
