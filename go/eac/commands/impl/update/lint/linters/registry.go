// registry.go - Self-registration system for lint handlers
package linters

import (
	"io"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/logging"
)

// LintOptions contains flags for controlling the lint process.
type LintOptions struct {
	Fix       bool     // Auto-fix issues where possible
	Config    string   // Override config file path
	Files     []string // Specific files to lint (for file-based linting)
	LintInput string   // How files are passed: "packages" or "files"
}

// Handler is the interface for lint handlers.
// Each handler is responsible for linting files of a specific package type.
type Handler interface {
	// Name returns the handler identifier (e.g., "go", "typescript")
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
// Call this from init() in your builder file.
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

// GetHandlerByLinter returns the handler for a specific linter tool name.
// This is the primary dispatch mechanism for package-based linting.
func GetHandlerByLinter(linterName string) Handler {
	mu.RLock()
	defer mu.RUnlock()

	// Map linter names to handler names
	// Some linters have specific handler implementations
	linterToHandler := map[string]string{
		"golangci-lint":     "go",
		"markdownlint-cli2": "markdown",
		"eslint":            "typescript", // or javascript
	}

	handlerName, ok := linterToHandler[linterName]
	if !ok {
		// Try direct mapping (handler name == linter name)
		handlerName = linterName
	}

	return handlers[handlerName]
}
