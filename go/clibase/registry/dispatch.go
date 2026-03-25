package registry

import (
	"slices"
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

// All returns all registered commands, excluding alias entries.
// Each command appears exactly once, keyed by its primary Name().
func (r *CommandRegistry) All() []core.CommandPort {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]bool)
	var result []core.CommandPort
	for key, cmd := range r.commands {
		if key == cmd.Name() && !seen[cmd.Name()] {
			seen[cmd.Name()] = true
			result = append(result, cmd)
		}
	}
	return result
}

// Names returns all registered primary command names, sorted alphabetically.
// Alias keys are excluded.
func (r *CommandRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for key, cmd := range r.commands {
		if key == cmd.Name() {
			names = append(names, key)
		}
	}
	sort.Strings(names)
	return names
}

// Subcommands returns direct child commands of a parent.
// For parent "get", returns "get modules", "get files", but not "get files tree".
// Each command appears at most once per parent (deduplication by command pointer).
func (r *CommandRegistry) Subcommands(parentName string) []core.CommandPort {
	entries := r.SubcommandEntries(parentName)
	result := make([]core.CommandPort, len(entries))
	for i, e := range entries {
		result[i] = e.Cmd
	}
	return result
}

// SubcommandEntries returns (key, command) pairs for direct children of parentName.
// Key is the map key used to look up the command, which may be an alias.
// Use Key (not Cmd.Name()) for display labels when listing children of a parent.
// Each command appears at most once: if both a primary key and an alias key match
// the same parent prefix, only one entry is returned.
func (r *CommandRegistry) SubcommandEntries(parentName string) []core.SubcommandEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prefix := parentName + " "
	// Collect one entry per command pointer. If multiple keys match for the same
	// command under this parent, the map naturally deduplicates (last write wins,
	// but in practice each command has at most one alias per parent).
	seen := make(map[string]core.SubcommandEntry) // keyed by cmd.Name()
	for key, cmd := range r.commands {
		if strings.HasPrefix(key, prefix) {
			remainder := strings.TrimPrefix(key, prefix)
			if !strings.Contains(remainder, " ") {
				cmdName := cmd.Name()
				if _, exists := seen[cmdName]; !exists {
					seen[cmdName] = core.SubcommandEntry{Key: key, Cmd: cmd}
				}
			}
		}
	}
	result := make([]core.SubcommandEntry, 0, len(seen))
	for _, entry := range seen {
		result = append(result, entry)
	}
	slices.SortFunc(result, func(a, b core.SubcommandEntry) int {
		return strings.Compare(a.Key, b.Key)
	})
	return result
}
