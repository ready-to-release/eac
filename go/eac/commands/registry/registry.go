// Package registry handles command registration
package registry

import (
	"bufio"
	"os"
	"runtime"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/workspace"
)

// InitialWorkingDir stores the working directory when the program started.
var InitialWorkingDir string

func init() {
	var err error
	InitialWorkingDir, err = workspace.WorkingDir()
	if err != nil {
		InitialWorkingDir = "."
	}
}

// GetWorkspaceRoot returns the repository root directory.
func GetWorkspaceRoot() (string, error) {
	return workspace.Root()
}

// CommandFunc is the signature for all command functions.
type CommandFunc func() int

// FlagMetadata holds structured metadata for a command flag.
type FlagMetadata struct {
	Name         string   // e.g., "debug", "module"
	Shorthand    string   // e.g., "d", "m" (optional single letter)
	Type         string   // "bool", "string", "int", etc.
	DefaultValue string   // default value as string
	Usage        string   // educational, specific description
	Required     bool     // is this flag required?
	Completion   []string // static completion values (optional)
}

// SubcommandGroup represents a logical grouping of subcommands for help display.
type SubcommandGroup struct {
	Name        string   // "Configuration", "Repository Structure", etc.
	Subcommands []string // ["config", "modules", "dependencies"]
}

// CommandRegistration holds command metadata.
type CommandRegistration struct {
	Func          CommandFunc
	ActualCommand string         // "get files" - the actual command users type
	CanonicalName string         // "get-files" - internal moniker (kebab-case)
	Short         string         // One sentence description
	Long          string         // Detailed multi-paragraph description
	Flags         []FlagMetadata // Structured flag definitions
	Args          string         // Argument completion type: "modules", "files", etc.

	// For parent commands (e.g., get, show, work)
	IsParent    bool              // true for parent commands with subcommands
	Subcommands []SubcommandGroup // grouped subcommands for help display
	Examples    []string          // usage examples
}

// commandRegistry maps command names (space-separated, e.g., "get files") to registrations.
var commandRegistry = map[string]*CommandRegistration{}

// Register allows command files to register themselves by extracting metadata from source comments
// The function automatically parses the calling file to extract:
// - Command name from "// Command: <name>"
// - Description from "// Description: <text>"
// - Usage from "// Usage: <text>".
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

	// Build SubcommandGroup slice from Groups map, preserving order
	var subcommands []SubcommandGroup
	for _, groupName := range metadata.GroupOrder {
		if subs, ok := metadata.Groups[groupName]; ok {
			subcommands = append(subcommands, SubcommandGroup{
				Name:        groupName,
				Subcommands: subs,
			})
		}
	}

	// Store in registry (keyed by ActualCommand for dispatch)
	commandRegistry[metadata.CommandName] = &CommandRegistration{
		Func:          fn,
		ActualCommand: metadata.CommandName,
		CanonicalName: canonicalName,
		Short:         short,
		Long:          long,
		Flags:         flags,
		Args:          metadata.Args,
		IsParent:      metadata.IsParent,
		Subcommands:   subcommands,
		Examples:      metadata.Examples,
	}
}

// commandMetadata holds extracted comment data.
type commandMetadata struct {
	CommandName string
	Description string   // Parsed from "// Description:" (used as fallback for Short)
	Short       string   // Parsed from "// Short:"
	LongLines   []string // Multi-line long description from "// Long:"
	FlagDefs    []flagDefinition
	Args        string // Argument completion type: "modules", "files", etc.

	// Parent command metadata
	IsParent   bool                // Parsed from "// IsParent: true"
	Groups     map[string][]string // Parsed from "// Group.<name>: sub1, sub2"
	GroupOrder []string            // Preserves order of groups as they appear
	Examples   []string            // Parsed from "// Example: ..."
}

// flagDefinition holds parsed flag definition from comments.
type flagDefinition struct {
	Name       string
	Attributes map[string]string // type, default, shorthand, usage, required, completion
}

// translateCrossCompilePath converts a compile-time file path to a runtime path
// when running a cross-compiled binary. This handles the case where a binary
// compiled on Linux (with paths like /home/runner/work/eac/eac/...) is executed
// on Windows where those paths don't exist.
func translateCrossCompilePath(compilePath string) string {
	// If the file already exists, use it as-is
	if _, err := os.Stat(compilePath); err == nil {
		return compilePath
	}

	// Check for Linux CI path pattern: /home/runner/work/<repo>/<repo>/...
	// We need to extract the relative path after the repo root and resolve it locally
	linuxCIPrefix := "/home/runner/work/"
	if strings.HasPrefix(compilePath, linuxCIPrefix) {
		// Path format: /home/runner/work/eac/eac/go/eac/commands/...
		// We need to extract: go/eac/commands/...
		parts := strings.SplitN(compilePath[len(linuxCIPrefix):], "/", 3)
		if len(parts) >= 3 {
			// parts[0] = repo name, parts[1] = repo name again, parts[2] = relative path
			relativePath := parts[2]
			// Try to find the file relative to current working directory
			if _, err := os.Stat(relativePath); err == nil {
				return relativePath
			}
			// Also try with backslashes for Windows
			windowsPath := strings.ReplaceAll(relativePath, "/", "\\")
			if _, err := os.Stat(windowsPath); err == nil {
				return windowsPath
			}
		}
	}

	// Fallback: return original path (will fail gracefully in caller)
	return compilePath
}

// extractCommandMetadata parses a Go source file to extract command metadata from header comments.
func extractCommandMetadata(filePath string) commandMetadata {
	var metadata commandMetadata

	// Handle cross-compiled binaries (e.g., Linux binary running on Windows)
	filePath = translateCrossCompilePath(filePath)

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

		// Extract Args (completion type for positional arguments)
		if strings.HasPrefix(line, "// Args:") {
			metadata.Args = strings.TrimSpace(strings.TrimPrefix(line, "// Args:"))
		}

		// Extract IsParent
		if strings.HasPrefix(line, "// IsParent:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "// IsParent:"))
			metadata.IsParent = value == "true"
		}

		// Extract Group definitions (// Group.<name>: sub1, sub2, ...)
		if strings.HasPrefix(line, "// Group.") {
			if metadata.Groups == nil {
				metadata.Groups = make(map[string][]string)
			}
			groupDef := strings.TrimPrefix(line, "// Group.")
			colonIdx := strings.Index(groupDef, ":")
			if colonIdx != -1 {
				groupName := strings.TrimSpace(groupDef[:colonIdx])
				subsStr := strings.TrimSpace(groupDef[colonIdx+1:])
				var subs []string
				for _, s := range strings.Split(subsStr, ",") {
					s = strings.TrimSpace(s)
					if s != "" {
						subs = append(subs, s)
					}
				}
				metadata.Groups[groupName] = subs
				metadata.GroupOrder = append(metadata.GroupOrder, groupName)
			}
		}

		// Extract Examples
		if strings.HasPrefix(line, "// Example:") {
			example := strings.TrimSpace(strings.TrimPrefix(line, "// Example:"))
			if example != "" {
				metadata.Examples = append(metadata.Examples, example)
			}
		}
	}

	return metadata
}

// parseFlagDefinition parses a flag definition comment line
// Format: // Flag.<name>: type=string, default=value, shorthand=s, usage=description, required=true, completion=val1,val2.
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
// Handles special case where usage field may contain commas.
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

// isValidAttributeKey checks if a string is a valid attribute key.
func isValidAttributeKey(s string) bool {
	validKeys := []string{"type", "default", "shorthand", "required", "completion"}
	for _, key := range validKeys {
		if s == key {
			return true
		}
	}
	return false
}

// GetCommands returns a map of command names to their functions (for dispatch).
func GetCommands() map[string]CommandFunc {
	result := make(map[string]CommandFunc, len(commandRegistry))
	for name, reg := range commandRegistry {
		result[name] = reg.Func
	}
	return result
}

// GetCommandRegistry returns the command registry.
func GetCommandRegistry() map[string]*CommandRegistration {
	return commandRegistry
}

// GetCanonicalName returns the kebab-case canonical name for a command.
func GetCanonicalName(commandName string) string {
	return strings.ReplaceAll(commandName, " ", "-")
}

// GetCommand retrieves a command registration by its command name (space-separated).
func GetCommand(commandName string) *CommandRegistration {
	return commandRegistry[commandName]
}

// GetCommandByCanonical retrieves a command registration by its canonical name (kebab-case).
func GetCommandByCanonical(canonicalName string) *CommandRegistration {
	// Convert kebab-case to space-separated for lookup
	actualName := strings.ReplaceAll(canonicalName, "-", " ")
	return commandRegistry[actualName]
}
