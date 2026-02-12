package registry

import (
	"fmt"
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// ErrDuplicateCommand is returned when registering a command with a name that already exists.
type ErrDuplicateCommand struct {
	Name string
}

func (e *ErrDuplicateCommand) Error() string {
	return fmt.Sprintf("duplicate command definition: %q", e.Name)
}

// global is set by main() so that command implementations can access the registry.
var global core.CommandRegistryPort

// SetGlobal sets the global command registry. Called once from main().
func SetGlobal(reg core.CommandRegistryPort) {
	global = reg
}

// Global returns the global command registry.
// Returns nil if SetGlobal has not been called.
func Global() core.CommandRegistryPort {
	return global
}

// CommandRegistry is a thread-safe registry of CommandPort implementations.
// It implements core.CommandRegistryPort for read access.
type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]core.CommandPort
}

var _ core.CommandRegistryPort = (*CommandRegistry)(nil)

// NewCommandRegistry creates an empty CommandRegistry.
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]core.CommandPort),
	}
}
