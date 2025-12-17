// Package flags provides shared flag parsing utilities for commands.
package flags

import (
	"fmt"
	"strings"
)

// FlagDefinition defines metadata for a command flag.
// Used to declare expected flags and enable validation.
type FlagDefinition struct {
	Name      string   // Primary flag name (e.g., "--ai-provider")
	Shorthand string   // Optional short form (e.g., "-a")
	Aliases   []string // Common typos/alternatives (e.g., ["--api-provider"])
	ValueType string   // "string", "bool", "int" (for documentation)
	Required  bool     // Whether flag is required
	HasValue  bool     // Whether flag expects a value
}

// ValidateFlags checks for unknown flags and provides helpful error messages.
// Returns an error if unknown flags are found, suggesting valid alternatives.
//
// Example:
//
//	defs := []flags.FlagDefinition{
//	    {Name: "--module", Shorthand: "-m", HasValue: true},
//	    {Name: "--debug", Shorthand: "-d", HasValue: false},
//	}
//	if err := flags.ValidateFlags(os.Args[2:], defs); err != nil {
//	    log.Errorf("%v", err)
//	    return 1
//	}
func ValidateFlags(args []string, known []FlagDefinition) error {
	// Build map of valid flags (NOT including aliases - aliases should error)
	validFlags := make(map[string]bool)
	for _, def := range known {
		validFlags[def.Name] = true
		if def.Shorthand != "" {
			validFlags[def.Shorthand] = true
		}
	}

	// Check each argument
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip non-flags
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		// Extract flag name (handle --flag=value format)
		flagName := arg
		if idx := strings.Index(arg, "="); idx != -1 {
			flagName = arg[:idx]
		}

		// Check if flag is valid
		if !validFlags[flagName] {
			// Unknown flag - build helpful error message
			return buildUnknownFlagError(flagName, known)
		}
	}

	return nil
}

// buildUnknownFlagError creates a detailed error message for unknown flags.
func buildUnknownFlagError(unknownFlag string, known []FlagDefinition) error {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("Unknown flag: %s\n\n", unknownFlag))
	msg.WriteString("Valid flags:\n")

	for _, def := range known {
		if def.Shorthand != "" {
			msg.WriteString(fmt.Sprintf("  %s, %-15s ", def.Name, def.Shorthand))
		} else {
			msg.WriteString(fmt.Sprintf("  %-20s ", def.Name))
		}

		if def.HasValue {
			msg.WriteString(fmt.Sprintf("(requires value, type: %s)", def.ValueType))
		} else {
			msg.WriteString("(boolean flag)")
		}
		msg.WriteString("\n")
	}

	// Suggest similar flags
	if suggestions := suggestSimilarFlags(unknownFlag, known); len(suggestions) > 0 {
		msg.WriteString("\nDid you mean:\n")
		for _, suggestion := range suggestions {
			msg.WriteString(fmt.Sprintf("  %s\n", suggestion))
		}
	}

	return fmt.Errorf("%s", msg.String())
}

// suggestSimilarFlags finds flags that are similar to the unknown flag.
// Uses simple string matching to detect common typos.
func suggestSimilarFlags(unknownFlag string, known []FlagDefinition) []string {
	var suggestions []string

	for _, def := range known {
		// Check if unknown flag matches an alias (common typo)
		for _, alias := range def.Aliases {
			if alias == unknownFlag {
				suggestions = append(suggestions, fmt.Sprintf("%s (instead of %s)", def.Name, alias))
				continue
			}
		}

		// Check for similar prefix (e.g., --api vs --ai)
		if strings.HasPrefix(unknownFlag, "--") && strings.HasPrefix(def.Name, "--") {
			unknown := strings.TrimPrefix(unknownFlag, "--")
			valid := strings.TrimPrefix(def.Name, "--")

			// Similar length and common prefix
			if len(unknown) >= 2 && len(valid) >= 2 {
				// Check if either starts with the other's prefix
				if strings.HasPrefix(valid, unknown) || strings.HasPrefix(unknown, valid) {
					// Avoid duplicates
					found := false
					for _, s := range suggestions {
						if s == def.Name || strings.Contains(s, def.Name) {
							found = true
							break
						}
					}
					if !found {
						suggestions = append(suggestions, def.Name)
					}
				}
			}
		}
	}

	return suggestions
}

// ParseDebugFlag checks if the debug flag is present in the arguments.
// It accepts both --debug and -d shorthand.
//
// Example:
//
//	args := os.Args[2:] // Skip program name and command
//	debug := flags.ParseDebugFlag(args)
func ParseDebugFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--debug" || arg == "-d" {
			return true
		}
	}
	return false
}

// HasFlag checks if a specific flag is present in the arguments.
// It accepts both the full flag name and optional shorthand.
//
// Example:
//
//	hasAll := flags.HasFlag(args, "--all", "-a")
func HasFlag(args []string, flag string, shorthand string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
		if shorthand != "" && arg == shorthand {
			return true
		}
	}
	return false
}

// GetFlagValue retrieves the value for a flag in --flag=value format.
// Returns empty string if not found.
//
// Example:
//
//	target := flags.GetFlagValue(args, "--target")
func GetFlagValue(args []string, prefix string) string {
	for _, arg := range args {
		if len(arg) > len(prefix) && arg[:len(prefix)] == prefix && arg[len(prefix)] == '=' {
			return arg[len(prefix)+1:]
		}
	}
	return ""
}

// GetPositionalArgs returns non-flag arguments.
// Arguments starting with -- or - are considered flags.
//
// Example:
//
//	positional := flags.GetPositionalArgs(args)
func GetPositionalArgs(args []string) []string {
	var positional []string
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}
		// Skip flags
		if arg[0] == '-' {
			continue
		}
		positional = append(positional, arg)
	}
	return positional
}
