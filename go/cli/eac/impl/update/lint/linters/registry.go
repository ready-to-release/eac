// Package linters provides lint handlers using the pluggable tool system.
//
// All linters are defined in tool-config.yml and resolved via tool.GlobalLintBridge().
// This package provides convenience functions for accessing lint handlers.
package linters

import (
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/tool"
)

// Type aliases - delegate to tool package types.
type (
	// Handler is the interface for lint handlers.
	Handler = tool.LintHandler

	// LintOptions contains flags for controlling the lint process.
	LintOptions = tool.LintOptions
)

// GetHandler returns the handler for a given name from tool-config.yml.
func GetHandler(name string) Handler {
	return tool.GlobalLintBridge().GetHandler(name)
}

// GetAllHandlers returns all registered handlers from tool-config.yml.
func GetAllHandlers() map[string]Handler {
	return tool.GlobalLintBridge().GetAllHandlers()
}

// GetHandlerForProvider returns the handler for a lint provider.
// Provider names map to handler names via the providerHandlers mapping.
func GetHandlerForProvider(providerName string) Handler {
	return tool.GlobalLintBridge().GetHandlerForProvider(providerName)
}

// GetProvidersForModule returns all applicable lint providers for a module's components.
// Uses the lint-providers configuration to determine which providers apply.
func GetProvidersForModule(module interface{ GetEnabledComponents() []string }, lintProviders *config.LintProvidersConfig, componentTypes *config.ComponentTypesConfig) []string {
	return tool.GlobalLintBridge().GetProvidersForModule(module, lintProviders, componentTypes)
}

// HasHandler checks if a handler exists by name.
func HasHandler(name string) bool {
	return tool.GlobalLintBridge().HasHandler(name)
}
