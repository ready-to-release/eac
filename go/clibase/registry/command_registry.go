package registry

import (
	"fmt"
	"sort"
	"strings"
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

// Register adds a command to the registry. Returns ErrDuplicateCommand if the name is taken.
func (r *CommandRegistry) Register(cmd core.CommandPort) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		return &ErrDuplicateCommand{Name: name}
	}
	r.commands[name] = cmd
	return nil
}

// MustRegister registers a command, panicking on duplicate.
func (r *CommandRegistry) MustRegister(cmd core.CommandPort) {
	if err := r.Register(cmd); err != nil {
		panic(err)
	}
}

// RegisterAll registers multiple commands. Stops and returns the first error encountered.
func (r *CommandRegistry) RegisterAll(cmds ...core.CommandPort) error {
	for _, cmd := range cmds {
		if err := r.Register(cmd); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves a command by its space-separated name (e.g. "get modules").
func (r *CommandRegistry) Get(name string) (core.CommandPort, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, ok := r.commands[name]
	return cmd, ok
}

// GetByCanonical retrieves a command by its kebab-case canonical name (e.g. "get-modules").
func (r *CommandRegistry) GetByCanonical(canonicalName string) (core.CommandPort, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	actualName := strings.ReplaceAll(canonicalName, "-", " ")
	cmd, ok := r.commands[actualName]
	return cmd, ok
}

// All returns all registered commands.
func (r *CommandRegistry) All() []core.CommandPort {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]core.CommandPort, 0, len(r.commands))
	for _, cmd := range r.commands {
		result = append(result, cmd)
	}
	return result
}

// Names returns all registered command names, sorted alphabetically.
func (r *CommandRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Subcommands returns direct child commands of a parent.
// For parent "get", returns "get modules", "get files", but not "get files tree".
func (r *CommandRegistry) Subcommands(parentName string) []core.CommandPort {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prefix := parentName + " "
	var result []core.CommandPort
	for name, cmd := range r.commands {
		if strings.HasPrefix(name, prefix) {
			remainder := strings.TrimPrefix(name, prefix)
			if !strings.Contains(remainder, " ") {
				result = append(result, cmd)
			}
		}
	}
	return result
}
