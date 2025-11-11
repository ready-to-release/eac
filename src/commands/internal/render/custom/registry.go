package custom

import (
	"fmt"
	"strings"
)

// CustomRenderer is a function that takes YAML bytes and returns a formatted string
type CustomRenderer func(yamlBytes []byte) (string, error)

// RendererRegistration holds a renderer and its command filters
type RendererRegistration struct {
	Renderer CustomRenderer
	Commands []string // Empty means "*" (all commands)
}

// registry holds all registered custom renderers with their command filters
var registry = make(map[string]*RendererRegistration)

// Register adds a custom renderer to the registry
// commands can be:
//   - []string{} or []string{"*"} - matches all commands
//   - []string{"get-modules"} - only for get modules command
//   - []string{"get-modules", "get-files"} - for multiple commands
//
// This is typically called from init() functions in custom renderer files
func Register(name string, renderer CustomRenderer, commands []string) {
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("custom renderer %q already registered", name))
	}

	// Normalize empty array to "*" wildcard
	if len(commands) == 0 {
		commands = []string{"*"}
	}

	registry[name] = &RendererRegistration{
		Renderer: renderer,
		Commands: commands,
	}
}

// Get retrieves a custom renderer by name, checking if it supports the given command
// commandName format: kebab-case (e.g., "get-modules")
func Get(name string, commandName string) (CustomRenderer, error) {
	reg, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("custom renderer %q not found", name)
	}

	// Check if renderer supports this command (expects kebab-case)
	if !reg.SupportsCommand(commandName) {
		return nil, fmt.Errorf("custom renderer %q does not support command %q (supports: %s)",
			name, commandName, strings.Join(reg.Commands, ", "))
	}

	return reg.Renderer, nil
}

// List returns custom renderer names that support the given command
// commandName format: kebab-case (e.g., "get-modules")
// If commandName is empty, returns all renderers
func List(commandName string) []string {
	names := make([]string, 0, len(registry))

	for name, reg := range registry {
		// If no command filter specified, include all renderers
		if commandName == "" || reg.SupportsCommand(commandName) {
			names = append(names, name)
		}
	}

	return names
}

// SupportsCommand checks if this renderer supports the given command
// commandName should be in kebab-case (e.g., "get-modules")
func (r *RendererRegistration) SupportsCommand(commandName string) bool {
	// Check for wildcard
	for _, cmd := range r.Commands {
		if cmd == "*" {
			return true
		}
		if cmd == commandName {
			return true
		}
	}
	return false
}
