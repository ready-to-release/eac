// Package registry handles command registration
package registry

import (
	"bufio"
	"os"
	"runtime"
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
	Func           CommandFunc
	ActualCommand  string // "get files" - the actual command users type
	CanonicalName  string // "get-files" - internal moniker (kebab-case)
	Description    string // Command description from file header
	Usage          string // Command usage from file header
	HasSideEffects bool   // Whether command modifies repository files
}

// commands maps command names to their implementation functions
var commands = map[string]CommandFunc{}

// commandRegistry maps canonical kebab-case names to registrations
var commandRegistry = map[string]*CommandRegistration{}

// Register allows command files to register themselves by extracting metadata from source comments
// The function automatically parses the calling file to extract:
// - Command name from "// Command: <name>"
// - Description from "// Description: <text>"
// - Usage from "// Usage: <text>"
func Register(fn CommandFunc) {
	// Get the caller's file location
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("registry.Register: could not determine caller")
	}

	// Extract command metadata from file comments
	metadata := extractCommandMetadata(file)
	if metadata.CommandName == "" {
		panic("registry.Register: no '// Command:' found in " + file)
	}

	// Validate HasSideEffects is declared
	if metadata.HasSideEffectsStr == "" {
		panic("registry.Register: no '// HasSideEffects:' declaration found in " + file +
			"\nPlease add '// HasSideEffects: true' or '// HasSideEffects: false' to the command file header.")
	}

	// Parse and validate HasSideEffects value
	var hasSideEffects bool
	switch metadata.HasSideEffectsStr {
	case "true":
		hasSideEffects = true
	case "false":
		hasSideEffects = false
	default:
		panic("registry.Register: invalid HasSideEffects value '" + metadata.HasSideEffectsStr +
			"' in " + file + "\nMust be 'true' or 'false'")
	}

	// Store in original commands map for backward compatibility
	commands[metadata.CommandName] = fn

	// Derive canonical kebab-case name
	canonicalName := strings.ReplaceAll(metadata.CommandName, " ", "-")

	// Store in registry with both forms
	commandRegistry[canonicalName] = &CommandRegistration{
		Func:           fn,
		ActualCommand:  metadata.CommandName,
		CanonicalName:  canonicalName,
		Description:    metadata.Description,
		Usage:          metadata.Usage,
		HasSideEffects: hasSideEffects,
	}
}

// commandMetadata holds extracted comment data
type commandMetadata struct {
	CommandName       string
	Description       string
	Usage             string
	HasSideEffectsStr string // Parsed from "// HasSideEffects:" comment
}

// extractCommandMetadata parses a Go source file to extract command metadata from header comments
func extractCommandMetadata(filePath string) commandMetadata {
	var metadata commandMetadata

	file, err := os.Open(filePath)
	if err != nil {
		return metadata
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Stop at package declaration
		if strings.HasPrefix(line, "package ") {
			break
		}

		// Extract Command name
		if strings.HasPrefix(line, "// Command:") {
			metadata.CommandName = strings.TrimSpace(strings.TrimPrefix(line, "// Command:"))
		}

		// Extract Description
		if strings.HasPrefix(line, "// Description:") {
			metadata.Description = strings.TrimSpace(strings.TrimPrefix(line, "// Description:"))
		}

		// Extract Usage
		if strings.HasPrefix(line, "// Usage:") {
			metadata.Usage = strings.TrimSpace(strings.TrimPrefix(line, "// Usage:"))
		}

		// Extract HasSideEffects
		if strings.HasPrefix(line, "// HasSideEffects:") {
			metadata.HasSideEffectsStr = strings.TrimSpace(strings.TrimPrefix(line, "// HasSideEffects:"))
		}
	}

	return metadata
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
