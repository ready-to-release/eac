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
	InitialWorkingDir = os.Getenv("R2R_PWD")
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

// FlagMetadata holds structured metadata for a command flag
type FlagMetadata struct {
	Name         string   // e.g., "debug", "module"
	Shorthand    string   // e.g., "d", "m" (optional single letter)
	Type         string   // "bool", "string", "int", etc.
	DefaultValue string   // default value as string
	Usage        string   // educational, specific description
	Required     bool     // is this flag required?
	Completion   []string // static completion values (optional)
}

// CommandRegistration holds command metadata
type CommandRegistration struct {
	Func           CommandFunc
	ActualCommand  string         // "get files" - the actual command users type
	CanonicalName  string         // "get-files" - internal moniker (kebab-case)
	Short          string         // One sentence description
	Long           string         // Detailed multi-paragraph description
	Flags          []FlagMetadata // Structured flag definitions
	Args           string         // Argument completion type: "modules", "files", etc.
	HasSideEffects bool           // Whether command modifies repository files
}

// commandRegistry maps command names (space-separated, e.g., "get files") to registrations
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

	// Derive canonical kebab-case name for internal use
	canonicalName := strings.ReplaceAll(metadata.CommandName, " ", "-")

	// Build Long description from lines
	long := strings.Join(metadata.LongLines, "\n")

	// Parse flag definitions
	flags := make([]FlagMetadata, 0, len(metadata.FlagDefs))
	for _, flagDef := range metadata.FlagDefs {
		flag := FlagMetadata{
			Name:         flagDef.Name,
			Shorthand:    flagDef.Attributes["shorthand"],
			Type:         flagDef.Attributes["type"],
			DefaultValue: flagDef.Attributes["default"],
			Usage:        flagDef.Attributes["usage"],
			Required:     flagDef.Attributes["required"] == "true",
		}

		// Parse completion values (comma-separated)
		if comp := flagDef.Attributes["completion"]; comp != "" {
			flag.Completion = strings.Split(comp, ",")
		}

		flags = append(flags, flag)
	}

	// Use Short if available, fall back to Description during transition
	short := metadata.Short
	if short == "" {
		short = metadata.Description
	}

	// Store in registry (keyed by ActualCommand for dispatch)
	commandRegistry[metadata.CommandName] = &CommandRegistration{
		Func:           fn,
		ActualCommand:  metadata.CommandName,
		CanonicalName:  canonicalName,
		Short:          short,
		Long:           long,
		Flags:          flags,
		Args:           metadata.Args,
		HasSideEffects: hasSideEffects,
	}
}

// commandMetadata holds extracted comment data
type commandMetadata struct {
	CommandName       string
	Description       string   // Parsed from "// Description:" (used as fallback for Short)
	Short             string   // Parsed from "// Short:"
	LongLines         []string // Multi-line long description from "// Long:"
	FlagDefs          []flagDefinition
	Args              string // Argument completion type: "modules", "files", etc.
	HasSideEffectsStr string   // Parsed from "// HasSideEffects:" comment
}

// flagDefinition holds parsed flag definition from comments
type flagDefinition struct {
	Name         string
	Attributes   map[string]string // type, default, shorthand, usage, required, completion
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
		// Handle both LF and CRLF line endings
		line := strings.TrimSpace(strings.ReplaceAll(scanner.Text(), "\r", ""))

		// Stop at package declaration
		if strings.HasPrefix(line, "package ") {
			break
		}

		// Extract Command name
		if strings.HasPrefix(line, "// Command:") {
			metadata.CommandName = strings.TrimSpace(strings.TrimPrefix(line, "// Command:"))
		}

		// Extract Description (fallback for Short)
		if strings.HasPrefix(line, "// Description:") {
			metadata.Description = strings.TrimSpace(strings.TrimPrefix(line, "// Description:"))
		}

		// Extract Short description
		if strings.HasPrefix(line, "// Short:") {
			metadata.Short = strings.TrimSpace(strings.TrimPrefix(line, "// Short:"))
		}

		// Extract Long description (multi-line support)
		if strings.HasPrefix(line, "// Long:") {
			longLine := strings.TrimSpace(strings.TrimPrefix(line, "// Long:"))
			metadata.LongLines = append(metadata.LongLines, longLine)
		}

		// Extract Flag definitions
		if strings.HasPrefix(line, "// Flag.") {
			flagDef := parseFlagDefinition(line)
			if flagDef.Name != "" {
				metadata.FlagDefs = append(metadata.FlagDefs, flagDef)
			}
		}

		// Extract HasSideEffects
		if strings.HasPrefix(line, "// HasSideEffects:") {
			metadata.HasSideEffectsStr = strings.TrimSpace(strings.TrimPrefix(line, "// HasSideEffects:"))
		}

		// Extract Args (completion type for positional arguments)
		if strings.HasPrefix(line, "// Args:") {
			metadata.Args = strings.TrimSpace(strings.TrimPrefix(line, "// Args:"))
		}
	}

	return metadata
}

// parseFlagDefinition parses a flag definition comment line
// Format: // Flag.<name>: type=string, default=value, shorthand=s, usage=description, required=true, completion=val1,val2
func parseFlagDefinition(line string) flagDefinition {
	def := flagDefinition{
		Attributes: make(map[string]string),
	}

	// Extract flag name: "// Flag.<name>:"
	if !strings.HasPrefix(line, "// Flag.") {
		return def
	}

	line = strings.TrimPrefix(line, "// Flag.")
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return def
	}

	def.Name = strings.TrimSpace(line[:colonIdx])
	attributesStr := strings.TrimSpace(line[colonIdx+1:])

	// Parse attributes (key=value pairs separated by commas)
	// Handle usage field specially as it may contain commas
	parts := parseAttributes(attributesStr)
	for key, value := range parts {
		def.Attributes[key] = value
	}

	return def
}

// parseAttributes parses key=value pairs from attribute string
// Handles special case where usage field may contain commas
func parseAttributes(attrStr string) map[string]string {
	result := make(map[string]string)

	// Find usage= first, as it may contain commas
	usageIdx := strings.Index(attrStr, "usage=")
	if usageIdx != -1 {
		// Extract everything after "usage=" as the usage value
		usageStart := usageIdx + len("usage=")
		usageValue := attrStr[usageStart:]

		// Find the end of usage (next ", " followed by a key= pattern or end of string)
		usageEnd := len(usageValue)
		for i := 0; i < len(usageValue)-1; i++ {
			if usageValue[i] == ',' && usageValue[i+1] == ' ' {
				// Check if what follows looks like a key=value pattern
				remaining := strings.TrimSpace(usageValue[i+1:])
				if idx := strings.Index(remaining, "="); idx > 0 {
					// Check if text before = is a valid key (no spaces or special chars)
					potentialKey := remaining[:idx]
					if isValidAttributeKey(potentialKey) {
						usageEnd = i
						break
					}
				}
			}
		}

		result["usage"] = strings.TrimSpace(usageValue[:usageEnd])

		// Remove usage from original string
		beforeUsage := strings.TrimSpace(attrStr[:usageIdx])
		afterUsage := ""
		if usageStart+usageEnd < len(attrStr) {
			afterUsage = strings.TrimSpace(attrStr[usageStart+usageEnd:])
			if strings.HasPrefix(afterUsage, ",") {
				afterUsage = strings.TrimSpace(afterUsage[1:])
			}
		}

		// Parse remaining attributes
		remaining := beforeUsage
		if afterUsage != "" {
			if remaining != "" {
				remaining += ", " + afterUsage
			} else {
				remaining = afterUsage
			}
		}
		attrStr = remaining
	}

	// Parse remaining attributes normally
	if attrStr != "" {
		pairs := strings.Split(attrStr, ",")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				result[key] = value
			}
		}
	}

	return result
}

// isValidAttributeKey checks if a string is a valid attribute key
func isValidAttributeKey(s string) bool {
	validKeys := []string{"type", "default", "shorthand", "required", "completion"}
	for _, key := range validKeys {
		if s == key {
			return true
		}
	}
	return false
}

// GetCommands returns a map of command names to their functions (for dispatch)
func GetCommands() map[string]CommandFunc {
	result := make(map[string]CommandFunc, len(commandRegistry))
	for name, reg := range commandRegistry {
		result[name] = reg.Func
	}
	return result
}

// GetCommandRegistry returns the command registry
func GetCommandRegistry() map[string]*CommandRegistration {
	return commandRegistry
}

// GetCanonicalName returns the kebab-case canonical name for a command
func GetCanonicalName(commandName string) string {
	return strings.ReplaceAll(commandName, " ", "-")
}

// GetCommand retrieves a command registration by its command name (space-separated)
func GetCommand(commandName string) *CommandRegistration {
	return commandRegistry[commandName]
}

// GetCommandByCanonical retrieves a command registration by its canonical name (kebab-case)
func GetCommandByCanonical(canonicalName string) *CommandRegistration {
	// Convert kebab-case to space-separated for lookup
	actualName := strings.ReplaceAll(canonicalName, "-", " ")
	return commandRegistry[actualName]
}
