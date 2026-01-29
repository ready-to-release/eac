package tui

import (
	"os"

	"golang.org/x/term"
)

// IsInteractive returns true if running in an interactive terminal.
// Returns false in CI environments.
func IsInteractive() bool {
	// Check for CI environment variables
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_HOME",
		"TF_BUILD",
		"BUILDKITE",
		"CIRCLECI",
		"TRAVIS",
	}

	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return false
		}
	}

	// Check if stdout is a terminal
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// ShouldUseTUI determines if TUI should activate for a command.
// Parameters:
//   - commandPath: the command being run (e.g., "build", "work create")
//   - hasSubcommand: true if a subcommand was provided (e.g., "update lint" not "update")
//   - isHelpRequest: true if --help was requested
func ShouldUseTUI(commandPath string, hasSubcommand bool, isHelpRequest bool) bool {
	// Never use TUI for --help
	if isHelpRequest {
		return false
	}

	// Never use TUI in CI
	if !IsInteractive() {
		return false
	}

	// Check if this command has a specific TUI binding
	binding := resolveBinding(commandPath)

	// For default TUI: only activate if no subcommand provided
	// (user ran bare "update" not "update lint")
	if binding != nil && binding.IsDefault && hasSubcommand {
		return false
	}

	return true
}

// Binding represents a resolved command-to-TUI binding.
type Binding struct {
	Pattern   string
	IsDefault bool // true for "*" binding
}

// resolveBinding returns the binding that would be used for a command.
func resolveBinding(commandPath string) *Binding {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	// Check exact match
	if _, ok := globalRegistry.exact[commandPath]; ok {
		return &Binding{Pattern: commandPath, IsDefault: false}
	}

	// Check prefix match
	for prefix := range globalRegistry.prefix {
		if len(commandPath) > len(prefix) && commandPath[:len(prefix)] == prefix {
			return &Binding{Pattern: prefix + " *", IsDefault: false}
		}
	}

	// Default
	if globalRegistry.defaultTUI != nil {
		return &Binding{Pattern: "*", IsDefault: true}
	}

	return nil
}
