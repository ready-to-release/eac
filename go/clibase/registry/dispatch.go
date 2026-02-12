package registry

import (
	"sort"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

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
