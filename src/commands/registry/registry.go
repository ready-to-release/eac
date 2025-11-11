// Package registry handles command registration
package registry

import (
	"os"
	"strings"
)

// InitialWorkingDir stores the working directory when the program started
var InitialWorkingDir string

// WorkspaceRoot stores the repository root (cached after first call)
var WorkspaceRoot string

func init() {
	// Initialize working directory
	InitialWorkingDir = os.Getenv("CLI_ORIGINAL_PWD")
	if InitialWorkingDir == "" {
		var err error
		InitialWorkingDir, err = os.Getwd()
		if err != nil {
			InitialWorkingDir = "."
		}
	}
}

// GetWorkspaceRoot returns the repository root directory, using cached value if available
func GetWorkspaceRoot() (string, error) {
	if WorkspaceRoot != "" {
		return WorkspaceRoot, nil
	}

	// Import repository package at runtime to avoid circular dependency
	// We'll use a simple implementation here instead
	return findRepositoryRoot()
}

// findRepositoryRoot finds the git repository root by walking up directories
func findRepositoryRoot() (string, error) {
	startPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	currentPath := startPath
	for {
		gitPath := currentPath + string(os.PathSeparator) + ".git"
		if _, err := os.Stat(gitPath); err == nil {
			WorkspaceRoot = currentPath
			return currentPath, nil
		}

		parentPath := currentPath[:strings.LastIndex(currentPath, string(os.PathSeparator))]
		if parentPath == currentPath || parentPath == "" {
			return "", os.ErrNotExist
		}
		currentPath = parentPath
	}
}

// CommandFunc is the signature for all command functions
type CommandFunc func() int

// CommandRegistration holds command metadata
type CommandRegistration struct {
	Func          CommandFunc
	DisplayName   string // "get files" (with spaces)
	CanonicalName string // "get-files" (kebab-case)
}

// commands maps command names to their implementation functions
var commands = map[string]CommandFunc{}

// commandRegistry maps canonical kebab-case names to registrations
var commandRegistry = map[string]*CommandRegistration{}

// Register allows command files to register themselves
// commandName should be in display format with spaces (e.g., "get files")
func Register(commandName string, fn CommandFunc) {
	// Store in original commands map for backward compatibility
	commands[commandName] = fn

	// Derive canonical kebab-case name
	canonicalName := strings.ReplaceAll(commandName, " ", "-")

	// Store in registry with both forms
	commandRegistry[canonicalName] = &CommandRegistration{
		Func:          fn,
		DisplayName:   commandName,
		CanonicalName: canonicalName,
	}
}

// GetCommands returns the commands map
func GetCommands() map[string]CommandFunc {
	return commands
}

// GetCommandRegistry returns the command registry
func GetCommandRegistry() map[string]*CommandRegistration {
	return commandRegistry
}

// GetCanonicalName returns the kebab-case canonical name for a command
func GetCanonicalName(commandName string) string {
	return strings.ReplaceAll(commandName, " ", "-")
}

// GetCommandByCanonical retrieves a command registration by its canonical name
func GetCommandByCanonical(canonicalName string) *CommandRegistration {
	return commandRegistry[canonicalName]
}
