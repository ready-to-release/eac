// registry.go - Self-registration system for lint handlers
package linters

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// LintOptions contains flags for controlling the lint process.
type LintOptions struct {
	Fix       bool     // Auto-fix issues where possible
	Config    string   // Override config file path
	Files     []string // Specific files to lint (for file-based linting)
	InputMode string   // How files are passed: "packages", "files", or "directory"
}

// Handler is the interface for lint handlers.
// Each handler is responsible for linting files of a specific type.
type Handler interface {
	// Name returns the handler identifier (e.g., "go", "markdown")
	Name() string

	// Lint executes linting for a module.
	// Returns exit code (0 = success/no issues, non-zero = issues found or error).
	Lint(moduleRoot, workspaceRoot, outputDir string, logWriter io.Writer, opts LintOptions) int

	// Requirements returns system dependencies required by this handler.
	// Used for early validation (e.g., ["golangci-lint"]).
	Requirements() []string

	// ValidateModule checks if a module's configuration is valid for linting.
	// Returns nil if valid, or an error describing the problem.
	ValidateModule(moduleRoot, workspaceRoot string) error
}

var (
	mu       sync.RWMutex
	handlers = make(map[string]Handler)
	log      = logging.C()
)

// RegisterHandler registers a handler for linting.
// Call this from init() in your linter file.
func RegisterHandler(h Handler) {
	mu.Lock()
	defer mu.Unlock()
	handlers[h.Name()] = h
}

// GetHandler returns the handler for a given name, or nil if not found.
func GetHandler(name string) Handler {
	mu.RLock()
	defer mu.RUnlock()
	return handlers[name]
}

// GetAllHandlers returns all registered handlers.
func GetAllHandlers() map[string]Handler {
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]Handler, len(handlers))
	for k, v := range handlers {
		result[k] = v
	}
	return result
}

// GetHandlerForProvider returns the handler for a lint provider.
// Provider names map to handler names via the providerHandlers mapping.
func GetHandlerForProvider(providerName string) Handler {
	mu.RLock()
	defer mu.RUnlock()

	// Provider-to-handler mapping
	// Provider names (from lint-providers.yml) map to handler names (from Go code)
	providerHandlers := map[string]string{
		"golangci-lint":     "go",
		"markdownlint-cli2": "markdown",
		"eslint":            "typescript",
	}

	handlerName, ok := providerHandlers[providerName]
	if !ok {
		// Direct mapping: provider name == handler name
		handlerName = providerName
	}

	return handlers[handlerName]
}

// GetProvidersForModule returns all applicable lint providers for a module's components.
// Uses the lint-providers configuration to determine which providers apply.
func GetProvidersForModule(module interface{ GetEnabledComponents() []string }, lintProviders *config.LintProvidersConfig, componentTypes *config.ComponentTypesConfig) []string {
	if module == nil || lintProviders == nil {
		return nil
	}

	providerSet := make(map[string]bool)
	for _, compName := range module.GetEnabledComponents() {
		// Get component type (may differ from name)
		compType := compName
		if componentTypes != nil {
			// For modules with explicit type mapping, use that
			// Otherwise component name is the type
		}

		// Find providers that apply to this component type
		for name, provider := range lintProviders.LintProviders {
			for _, applies := range provider.AppliesTo {
				if applies == compType {
					providerSet[name] = true
					break
				}
			}
		}
	}

	result := make([]string, 0, len(providerSet))
	for name := range providerSet {
		result = append(result, name)
	}
	return result
}
